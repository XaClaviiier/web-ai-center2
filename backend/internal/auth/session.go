// Package auth mengelola login admin lewat Keycloak (OIDC Authorization Code + PKCE) dan session server: ID acak,
// token CSRF, masa berlaku idle/mutlak, serta penyegaran token Keycloak di sisi server (security.md §1-§3).
// Token Keycloak tidak pernah dikirim ke browser; refresh token disimpan terenkripsi di Redis.
package auth

import (
	"cmp"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

// ErrTidakAda berarti session tidak ditemukan, sudah kedaluwarsa, atau ID-nya tidak sah.
var ErrTidakAda = errors.New("session tidak ada atau kedaluwarsa")

// ErrSegarDitolak berarti Keycloak menolak penyegaran token (invalid_grant: session SSO berakhir, pengguna dinonaktifkan)
// atau token hasil penyegaran tidak lolos verifikasi/role dicabut. Session lokal dihapus; pengguna harus login ulang.
var ErrSegarDitolak = errors.New("penyegaran token ditolak penyedia identitas")

// ErrSegarSibuk berarti penyegaran token yang sedang dijalankan request lain tidak selesai dalam batas tunggu.
var ErrSegarSibuk = errors.New("penyegaran token session belum selesai")

const (
	panjangBytes = 32 // 256 bit (security.md §2)
	panjangID    = 43 // base64url tanpa padding dari 32 byte
	alfabet      = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

	// kelonggaranAkses: access token dianggap habis sedikit lebih awal agar tidak kedaluwarsa di tengah jalan.
	kelonggaranAkses = 30 * time.Second
	// masaKunciSegar membatasi kunci penyegaran bila pemegangnya mati; lebih panjang dari batas panggilan ke Keycloak.
	masaKunciSegar = 15 * time.Second
	// batasPanggilSegar membatasi satu panggilan penyegaran ke Keycloak.
	batasPanggilSegar  = 10 * time.Second
	tungguSegarDefault = 5 * time.Second
	jedaTungguSegar    = 50 * time.Millisecond
)

// Identitas adalah siapa yang login. Sub dipakai untuk audit (created_by/updated_by); Role hanya label tampilan.
type Identitas struct {
	Sub      string
	Nama     string
	Role     string
	Username string // snapshot username untuk log aktivitas (claim preferred_username); kosong = pakai Sub
}

// Token adalah token Keycloak milik session; hanya disimpan di server.
type Token struct {
	Refresh    string    // refresh token; disimpan terenkripsi (Sandi)
	AksesHabis time.Time // masa berlaku access token terakhir; lewat ini session disegarkan
}

// Session adalah data yang disimpan di Redis pada key session:{id}.
type Session struct {
	Identitas
	Token
	CSRFToken  string
	DibuatPada time.Time
}

// Store menyimpan session. Implementasi nyata memakai Redis; tes memakai versi palsu.
type Store interface {
	Simpan(ctx context.Context, id string, s Session, ttl time.Duration) error
	// Ambil mengembalikan ErrTidakAda bila key tidak ada.
	Ambil(ctx context.Context, id string) (Session, error)
	// Perpanjang mengatur ulang TTL; ErrTidakAda bila key sudah hilang.
	Perpanjang(ctx context.Context, id string, ttl time.Duration) error
	// GantiBilaAda menimpa session HANYA bila key masih ada (logout bersamaan menang); ErrTidakAda bila sudah hilang.
	GantiBilaAda(ctx context.Context, id string, s Session, ttl time.Duration) error
	Hapus(ctx context.Context, id string) error
	// KunciSegar mencoba mengambil kunci penyegaran token session (SET NX); false bila dipegang pihak lain.
	KunciSegar(ctx context.Context, id, pemilik string, ttl time.Duration) (bool, error)
	// LepasKunciSegar melepas kunci hanya bila masih milik pemilik.
	LepasKunciSegar(ctx context.Context, id, pemilik string) error
}

// Penyegar menukar refresh token session dengan token baru dan memverifikasinya ulang (diimplementasikan *OIDC).
// Mengembalikan galat yang membungkus ErrSegarDitolak bila session harus diakhiri; galat lain = penyedia tidak terjangkau.
type Penyegar interface {
	Segarkan(ctx context.Context, s Session) (Identitas, Token, error)
}

// Manager menerapkan kebijakan masa berlaku: idle diperpanjang tiap request, batas mutlak tidak pernah dilewati
// (juga tidak oleh penyegaran token), dan access token Keycloak disegarkan di server saat habis.
type Manager struct {
	Store    Store
	IdleTTL  time.Duration
	AbsTTL   time.Duration
	Sekarang func() time.Time // default time.Now; diganti di tes
	// Penyegar menyegarkan token Keycloak; nil hanya untuk session uji tanpa token.
	Penyegar Penyegar
	// TungguSegar: lama request lain menunggu penyegaran yang sedang berjalan sebelum menyerah (503).
	TungguSegar time.Duration
}

// NewManager membuat Manager dengan jam sistem.
func NewManager(st Store, idle, mutlak time.Duration) *Manager {
	return &Manager{Store: st, IdleTTL: idle, AbsTTL: mutlak, Sekarang: time.Now, TungguSegar: tungguSegarDefault}
}

// Buat membuat session baru (ID dan token CSRF baru pada setiap login).
func (m *Manager) Buat(ctx context.Context, ident Identitas, tok Token) (string, Session, error) {
	id, err := TokenAcak()
	if err != nil {
		return "", Session{}, err
	}
	csrf, err := TokenAcak()
	if err != nil {
		return "", Session{}, err
	}
	s := Session{Identitas: ident, Token: tok, CSRFToken: csrf, DibuatPada: m.Sekarang().UTC()}
	if err := m.Store.Simpan(ctx, id, s, min(m.IdleTTL, m.AbsTTL)); err != nil {
		return "", Session{}, fmt.Errorf("menyimpan session: %w", err)
	}
	return id, s, nil
}

// Validasi memeriksa session, menyegarkan token Keycloak bila access token habis, lalu memperpanjang masa idle.
// ErrTidakAda bila tidak sah, melewati batas mutlak, atau penyegaran ditolak (session ikut dihapus). Galat lain berarti
// penyimpanan atau penyedia identitas bermasalah (gagal tertutup: session tidak diperpanjang).
func (m *Manager) Validasi(ctx context.Context, id string) (Session, error) {
	s, sisa, err := m.baca(ctx, id)
	if err != nil {
		return Session{}, err
	}
	if m.perluSegar(s) {
		return m.segarkan(ctx, id)
	}
	if err := m.Store.Perpanjang(ctx, id, min(m.IdleTTL, sisa)); err != nil {
		return Session{}, err
	}
	return s, nil
}

// Baca memeriksa session tanpa menyegarkan token dan tanpa memperpanjang idle. Dipakai logout agar penyedia
// identitas yang tidak terjangkau tidak menghalangi logout lokal.
func (m *Manager) Baca(ctx context.Context, id string) (Session, error) {
	s, _, err := m.baca(ctx, id)
	return s, err
}

// Hapus menghapus session (logout). ID tidak sah diabaikan.
func (m *Manager) Hapus(ctx context.Context, id string) error {
	if !IDSah(id) {
		return nil
	}
	return m.Store.Hapus(ctx, id)
}

// baca mengambil session dan sisa batas mutlaknya; session yang melewati batas mutlak dihapus.
func (m *Manager) baca(ctx context.Context, id string) (Session, time.Duration, error) {
	if !IDSah(id) {
		return Session{}, 0, ErrTidakAda
	}
	s, err := m.Store.Ambil(ctx, id)
	if err != nil {
		return Session{}, 0, err
	}
	sisa := m.AbsTTL - m.Sekarang().Sub(s.DibuatPada)
	// Dengan penyedia identitas aktif, session tanpa masa berlaku token tidak mungkin berasal dari login sah.
	if sisa <= 0 || (m.Penyegar != nil && s.AksesHabis.IsZero()) {
		_ = m.Store.Hapus(ctx, id)
		return Session{}, 0, ErrTidakAda
	}
	return s, sisa, nil
}

func (m *Manager) perluSegar(s Session) bool {
	return m.Penyegar != nil && !m.Sekarang().Add(kelonggaranAkses).Before(s.AksesHabis)
}

// segarkan menyegarkan token session. Hanya satu request per session yang memanggil Keycloak (kunci Redis SET NX);
// request lain menunggu hasilnya. Penulisan memakai GantiBilaAda sehingga logout bersamaan tidak dibatalkan.
func (m *Manager) segarkan(ctx context.Context, id string) (Session, error) {
	pemilik, err := TokenAcak()
	if err != nil {
		return Session{}, err
	}
	dapat, err := m.Store.KunciSegar(ctx, id, pemilik, masaKunciSegar)
	if err != nil {
		return Session{}, err
	}
	if !dapat {
		return m.tungguSegar(ctx, id)
	}
	defer func() { _ = m.Store.LepasKunciSegar(context.WithoutCancel(ctx), id, pemilik) }()

	// Baca ulang di bawah kunci: pemegang sebelumnya mungkin baru selesai menyegarkan.
	s, sisa, err := m.baca(ctx, id)
	if err != nil {
		return Session{}, err
	}
	if !m.perluSegar(s) {
		if err := m.Store.Perpanjang(ctx, id, min(m.IdleTTL, sisa)); err != nil {
			return Session{}, err
		}
		return s, nil
	}

	panggilCtx, batal := context.WithTimeout(ctx, batasPanggilSegar)
	ident, tok, err := m.Penyegar.Segarkan(panggilCtx, s)
	batal()
	if errors.Is(err, ErrSegarDitolak) {
		if err := m.Store.Hapus(ctx, id); err != nil {
			return Session{}, err
		}
		return Session{}, ErrTidakAda
	}
	if err != nil {
		return Session{}, err
	}

	s.Identitas = Identitas{Sub: s.Sub, Nama: cmp.Or(ident.Nama, s.Nama), Role: ident.Role, Username: cmp.Or(ident.Username, s.Username)}
	s.Refresh = cmp.Or(tok.Refresh, s.Refresh)
	s.AksesHabis = tok.AksesHabis
	sisa = m.AbsTTL - m.Sekarang().Sub(s.DibuatPada)
	if sisa <= 0 {
		_ = m.Store.Hapus(ctx, id)
		return Session{}, ErrTidakAda
	}
	if err := m.Store.GantiBilaAda(ctx, id, s, min(m.IdleTTL, sisa)); err != nil {
		return Session{}, err
	}
	return s, nil
}

// tungguSegar menunggu request lain menyelesaikan penyegaran. Session yang hilang selama menunggu (ditolak atau logout)
// menghasilkan ErrTidakAda; tidak ada penyegaran kedua.
func (m *Manager) tungguSegar(ctx context.Context, id string) (Session, error) {
	batas := time.Now().Add(m.TungguSegar)
	for {
		t := time.NewTimer(jedaTungguSegar)
		select {
		case <-ctx.Done():
			t.Stop()
			return Session{}, ctx.Err()
		case <-t.C:
		}
		s, sisa, err := m.baca(ctx, id)
		if err != nil {
			return Session{}, err
		}
		if !m.perluSegar(s) {
			if err := m.Store.Perpanjang(ctx, id, min(m.IdleTTL, sisa)); err != nil {
				return Session{}, err
			}
			return s, nil
		}
		if time.Now().After(batas) {
			return Session{}, ErrSegarSibuk
		}
	}
}

// IDSah memastikan ID berbentuk token yang kita buat, sebelum dipakai sebagai bagian key Redis.
func IDSah(id string) bool {
	if len(id) != panjangID {
		return false
	}
	for i := 0; i < len(id); i++ {
		if !adaDiAlfabet(id[i]) {
			return false
		}
	}
	return true
}

func adaDiAlfabet(c byte) bool {
	for i := 0; i < len(alfabet); i++ {
		if alfabet[i] == c {
			return true
		}
	}
	return false
}

// TokenAcak membuat token acak 256 bit (base64url tanpa padding): ID session, CSRF, state, nonce, pengikat login.
func TokenAcak() (string, error) {
	b := make([]byte, panjangBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("membuat token acak: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
