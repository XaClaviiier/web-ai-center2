// Package auth mengelola session login admin: ID acak, token CSRF, dan masa berlaku (security.md §2-§3).
// Identitas pengguna (Identitas) dibuat oleh pemanggil: login development sekarang, callback Keycloak nanti;
// sisanya (session, cookie, CSRF) sama untuk keduanya.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

// ErrTidakAda berarti session tidak ditemukan, sudah kedaluwarsa, atau ID-nya tidak sah.
var ErrTidakAda = errors.New("session tidak ada atau kedaluwarsa")

const (
	panjangBytes = 32 // 256 bit (security.md §2)
	panjangID    = 43 // base64url tanpa padding dari 32 byte
	alfabet      = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
)

// Identitas adalah siapa yang login. Sub dipakai untuk audit (created_by/updated_by); Role hanya label tampilan.
type Identitas struct {
	Sub  string
	Nama string
	Role string
}

// Session adalah data yang disimpan di Redis pada key session:{id}.
type Session struct {
	Identitas
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
	Hapus(ctx context.Context, id string) error
}

// Manager menerapkan kebijakan masa berlaku: idle diperpanjang tiap request, batas mutlak tidak pernah dilewati.
type Manager struct {
	Store    Store
	IdleTTL  time.Duration
	AbsTTL   time.Duration
	Sekarang func() time.Time // default time.Now; diganti di tes
}

// NewManager membuat Manager dengan jam sistem.
func NewManager(st Store, idle, mutlak time.Duration) *Manager {
	return &Manager{Store: st, IdleTTL: idle, AbsTTL: mutlak, Sekarang: time.Now}
}

// Buat membuat session baru (ID dan token CSRF baru pada setiap login).
func (m *Manager) Buat(ctx context.Context, ident Identitas) (string, Session, error) {
	id, err := tokenAcak()
	if err != nil {
		return "", Session{}, err
	}
	csrf, err := tokenAcak()
	if err != nil {
		return "", Session{}, err
	}
	s := Session{Identitas: ident, CSRFToken: csrf, DibuatPada: m.Sekarang().UTC()}
	if err := m.Store.Simpan(ctx, id, s, min(m.IdleTTL, m.AbsTTL)); err != nil {
		return "", Session{}, fmt.Errorf("menyimpan session: %w", err)
	}
	return id, s, nil
}

// Validasi memeriksa session dan memperpanjang masa idle-nya. ErrTidakAda bila tidak sah atau melewati batas mutlak
// (session yang melewati batas mutlak ikut dihapus). Galat lain berarti penyimpanan bermasalah.
func (m *Manager) Validasi(ctx context.Context, id string) (Session, error) {
	if !IDSah(id) {
		return Session{}, ErrTidakAda
	}
	s, err := m.Store.Ambil(ctx, id)
	if err != nil {
		return Session{}, err
	}
	sisa := m.AbsTTL - m.Sekarang().Sub(s.DibuatPada)
	if sisa <= 0 {
		_ = m.Store.Hapus(ctx, id)
		return Session{}, ErrTidakAda
	}
	if err := m.Store.Perpanjang(ctx, id, min(m.IdleTTL, sisa)); err != nil {
		return Session{}, err
	}
	return s, nil
}

// Hapus menghapus session (logout). ID tidak sah diabaikan.
func (m *Manager) Hapus(ctx context.Context, id string) error {
	if !IDSah(id) {
		return nil
	}
	return m.Store.Hapus(ctx, id)
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

func tokenAcak() (string, error) {
	b := make([]byte, panjangBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("membuat token acak: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
