package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// memStore adalah Store palsu di memori; mencatat TTL terakhir dan jumlah panggilan.
type memStore struct {
	data    map[string]Session
	ttl     map[string]time.Duration
	kunci   map[string]string
	panggil int
}

func baruMemStore() *memStore {
	return &memStore{data: map[string]Session{}, ttl: map[string]time.Duration{}, kunci: map[string]string{}}
}

func (m *memStore) Simpan(_ context.Context, id string, s Session, ttl time.Duration) error {
	m.panggil++
	m.data[id], m.ttl[id] = s, ttl
	return nil
}

func (m *memStore) GantiBilaAda(_ context.Context, id string, s Session, ttl time.Duration) error {
	m.panggil++
	if _, ok := m.data[id]; !ok {
		return ErrTidakAda
	}
	m.data[id], m.ttl[id] = s, ttl
	return nil
}

func (m *memStore) Ambil(_ context.Context, id string) (Session, error) {
	m.panggil++
	s, ok := m.data[id]
	if !ok {
		return Session{}, ErrTidakAda
	}
	return s, nil
}

func (m *memStore) Perpanjang(_ context.Context, id string, ttl time.Duration) error {
	m.panggil++
	if _, ok := m.data[id]; !ok {
		return ErrTidakAda
	}
	m.ttl[id] = ttl
	return nil
}

func (m *memStore) Hapus(_ context.Context, id string) error {
	m.panggil++
	delete(m.data, id)
	delete(m.ttl, id)
	return nil
}

func (m *memStore) KunciSegar(_ context.Context, id, pemilik string, _ time.Duration) (bool, error) {
	if _, ada := m.kunci[id]; ada {
		return false, nil
	}
	m.kunci[id] = pemilik
	return true, nil
}

func (m *memStore) LepasKunciSegar(_ context.Context, id, pemilik string) error {
	if m.kunci[id] == pemilik {
		delete(m.kunci, id)
	}
	return nil
}

func managerUji(st Store, now *time.Time) *Manager {
	m := NewManager(st, time.Hour, 8*time.Hour)
	m.Sekarang = func() time.Time { return *now }
	return m
}

var dev = Identitas{Sub: "dev-admin", Nama: "Admin Dev", Role: "pegawai"}

// penyegarPalsu mengembalikan hasil yang diatur tes.
type penyegarPalsu struct {
	ident   Identitas
	tok     Token
	err     error
	panggil int
	terima  Session
}

func (p *penyegarPalsu) Segarkan(_ context.Context, s Session) (Identitas, Token, error) {
	p.panggil++
	p.terima = s
	return p.ident, p.tok, p.err
}

func TestBuatMembuatIDDanCSRFAcakDanSah(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	st := baruMemStore()
	m := managerUji(st, &now)

	id1, s1, err := m.Buat(context.Background(), dev, Token{})
	if err != nil {
		t.Fatal(err)
	}
	id2, s2, err := m.Buat(context.Background(), dev, Token{})
	if err != nil {
		t.Fatal(err)
	}
	if !IDSah(id1) || !IDSah(s1.CSRFToken) || len(id1) != 43 {
		t.Fatalf("ID/CSRF tidak sah: %q %q", id1, s1.CSRFToken)
	}
	if id1 == id2 || s1.CSRFToken == s2.CSRFToken || id1 == s1.CSRFToken {
		t.Fatal("ID dan token CSRF harus acak dan berbeda pada setiap login")
	}
	if st.ttl[id1] != time.Hour {
		t.Errorf("TTL awal harus idle (1 jam), dapat %v", st.ttl[id1])
	}
	if !s1.DibuatPada.Equal(now) || s1.Identitas != dev {
		t.Errorf("isi session salah: %+v", s1)
	}
}

func TestValidasiMemperpanjangIdleTapiTidakMelewatiBatasMutlak(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	st := baruMemStore()
	m := managerUji(st, &now)
	id, _, _ := m.Buat(context.Background(), dev, Token{})

	// 30 menit kemudian: sisa mutlak masih banyak, TTL kembali ke idle penuh.
	now = now.Add(30 * time.Minute)
	if _, err := m.Validasi(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if st.ttl[id] != time.Hour {
		t.Errorf("TTL harus idle penuh, dapat %v", st.ttl[id])
	}

	// 7 jam 50 menit sejak dibuat: sisa mutlak 10 menit < idle, TTL dipotong ke sisa.
	now = time.Date(2026, 11, 9, 15, 50, 0, 0, time.UTC)
	if _, err := m.Validasi(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if st.ttl[id] != 10*time.Minute {
		t.Errorf("TTL harus dipotong ke sisa batas mutlak (10 menit), dapat %v", st.ttl[id])
	}
}

func TestValidasiMelewatiBatasMutlakMenghapusSession(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	st := baruMemStore()
	m := managerUji(st, &now)
	id, _, _ := m.Buat(context.Background(), dev, Token{})

	now = now.Add(8 * time.Hour) // tepat di batas mutlak
	if _, err := m.Validasi(context.Background(), id); !errors.Is(err, ErrTidakAda) {
		t.Fatalf("diharapkan ErrTidakAda, dapat %v", err)
	}
	if _, ada := st.data[id]; ada {
		t.Error("session yang melewati batas mutlak harus dihapus dari store")
	}
}

func TestValidasiIDTidakSahTidakMenyentuhStore(t *testing.T) {
	now := time.Now()
	st := baruMemStore()
	m := managerUji(st, &now)
	for _, id := range []string{"", "pendek", strings.Repeat("a", 44), strings.Repeat("a", 42) + "!", strings.Repeat("a", 42) + "/", "../" + strings.Repeat("a", 40)} {
		if _, err := m.Validasi(context.Background(), id); !errors.Is(err, ErrTidakAda) {
			t.Errorf("%q: diharapkan ErrTidakAda, dapat %v", id, err)
		}
		if err := m.Hapus(context.Background(), id); err != nil {
			t.Errorf("Hapus %q harus diabaikan, dapat %v", id, err)
		}
	}
	if st.panggil != 0 {
		t.Errorf("ID tidak sah tidak boleh sampai ke store (panggilan: %d)", st.panggil)
	}
}

func TestValidasiSessionTakAdaDiStore(t *testing.T) {
	now := time.Now()
	m := managerUji(baruMemStore(), &now)
	if _, err := m.Validasi(context.Background(), strings.Repeat("a", 43)); !errors.Is(err, ErrTidakAda) {
		t.Fatalf("diharapkan ErrTidakAda, dapat %v", err)
	}
}

func TestHapus(t *testing.T) {
	now := time.Now()
	st := baruMemStore()
	m := managerUji(st, &now)
	id, _, _ := m.Buat(context.Background(), dev, Token{})
	if err := m.Hapus(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Validasi(context.Background(), id); !errors.Is(err, ErrTidakAda) {
		t.Errorf("setelah Hapus harus ErrTidakAda, dapat %v", err)
	}
}

func TestIDSah(t *testing.T) {
	if !IDSah(strings.Repeat("A", 43)) || !IDSah(strings.Repeat("a-_9", 10)+"abc") {
		t.Error("ID dengan alfabet base64url dan panjang 43 harus sah")
	}
}

// — penyegaran token ————————————————————————————————————————————————

func siapkanSegar(t *testing.T, now *time.Time, aksesHabis time.Time) (*Manager, *memStore, *penyegarPalsu, string) {
	t.Helper()
	st := baruMemStore()
	m := managerUji(st, now)
	p := &penyegarPalsu{}
	m.Penyegar = p
	id, _, err := m.Buat(context.Background(), dev, Token{Refresh: "rt-1", AksesHabis: aksesHabis})
	if err != nil {
		t.Fatal(err)
	}
	return m, st, p, id
}

func TestValidasiTidakMenyegarkanSebelumAksesHabis(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	m, _, p, id := siapkanSegar(t, &now, now.Add(5*time.Minute))
	now = now.Add(4 * time.Minute) // sisa 60 detik > kelonggaran 30 detik
	if _, err := m.Validasi(context.Background(), id); err != nil || p.panggil != 0 {
		t.Fatalf("tidak boleh menyegarkan: %v panggil=%d", err, p.panggil)
	}
}

func TestValidasiMenyegarkanSaatAksesHabis(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	m, st, p, id := siapkanSegar(t, &now, now.Add(5*time.Minute))
	now = now.Add(4*time.Minute + 45*time.Second) // dalam kelonggaran 30 detik
	habisBaru := now.Add(5 * time.Minute)
	p.ident = Identitas{Sub: "dev-admin", Nama: "Admin Baru", Role: RoleDirektur, Username: "admin.baru"}
	p.tok = Token{Refresh: "rt-2", AksesHabis: habisBaru}

	s, err := m.Validasi(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if p.panggil != 1 || p.terima.Refresh != "rt-1" {
		t.Fatalf("penyegar harus dipanggil sekali dengan refresh token lama: %d %q", p.panggil, p.terima.Refresh)
	}
	if s.Nama != "Admin Baru" || s.Role != RoleDirektur || s.Username != "admin.baru" || s.Refresh != "rt-2" || !s.AksesHabis.Equal(habisBaru) {
		t.Errorf("session hasil salah: %+v", s)
	}
	if got := st.data[id]; got.Refresh != "rt-2" || !got.AksesHabis.Equal(habisBaru) || got.Role != RoleDirektur || got.CSRFToken != s.CSRFToken {
		t.Errorf("store tidak diperbarui: %+v", got)
	}
	if st.ttl[id] != time.Hour {
		t.Errorf("TTL setelah penyegaran harus idle penuh, dapat %v", st.ttl[id])
	}
	if len(st.kunci) != 0 {
		t.Error("kunci penyegaran harus dilepas")
	}
	// Refresh token kosong dari penyedia: yang lama dipertahankan.
	now = habisBaru
	p.tok = Token{AksesHabis: now.Add(5 * time.Minute)}
	if s, err := m.Validasi(context.Background(), id); err != nil || s.Refresh != "rt-2" {
		t.Errorf("refresh token lama harus dipertahankan: %q %v", s.Refresh, err)
	}
}

func TestPenyegaranTidakMelewatiBatasMutlak(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	m, st, p, id := siapkanSegar(t, &now, now.Add(5*time.Minute))
	now = now.Add(7*time.Hour + 55*time.Minute)
	p.ident = dev
	p.tok = Token{Refresh: "rt-2", AksesHabis: now.Add(5 * time.Minute)}
	if _, err := m.Validasi(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if st.ttl[id] != 5*time.Minute {
		t.Errorf("TTL harus dipotong ke sisa mutlak (5 menit), dapat %v", st.ttl[id])
	}
	now = now.Add(5 * time.Minute)
	if _, err := m.Validasi(context.Background(), id); !errors.Is(err, ErrTidakAda) || p.panggil != 1 {
		t.Errorf("batas mutlak tetap berlaku walau token baru: %v panggil=%d", err, p.panggil)
	}
}

func TestPenyegaranDitolakMenghapusSession(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	m, st, p, id := siapkanSegar(t, &now, now.Add(time.Minute))
	now = now.Add(time.Minute)
	p.err = errors.Join(ErrSegarDitolak, errors.New("invalid_grant"))
	if _, err := m.Validasi(context.Background(), id); !errors.Is(err, ErrTidakAda) {
		t.Fatalf("diharapkan ErrTidakAda, dapat %v", err)
	}
	if _, ada := st.data[id]; ada {
		t.Error("session harus dihapus saat penyegaran ditolak")
	}
}

func TestPenyegaranGagalJaringanTidakMemperpanjang(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	m, st, p, id := siapkanSegar(t, &now, now.Add(time.Minute))
	st.ttl[id] = 42 * time.Second // penanda: tidak boleh diubah
	now = now.Add(time.Minute)
	p.err = errors.New("dial tcp: connection refused")
	_, err := m.Validasi(context.Background(), id)
	if err == nil || errors.Is(err, ErrTidakAda) {
		t.Fatalf("galat jaringan harus diteruskan (503), dapat %v", err)
	}
	if _, ada := st.data[id]; !ada || st.ttl[id] != 42*time.Second || st.data[id].Refresh != "rt-1" {
		t.Errorf("session harus utuh dan tidak diperpanjang: %+v ttl=%v", st.data[id], st.ttl[id])
	}
	if len(st.kunci) != 0 {
		t.Error("kunci penyegaran harus dilepas walau gagal")
	}
}

func TestSessionTanpaTokenDitolakBilaPenyegarAktif(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	st := baruMemStore()
	m := managerUji(st, &now)
	m.Penyegar = &penyegarPalsu{}
	id, _, _ := m.Buat(context.Background(), dev, Token{})
	if _, err := m.Validasi(context.Background(), id); !errors.Is(err, ErrTidakAda) {
		t.Fatalf("session tanpa masa berlaku token harus ditolak: %v", err)
	}
}

func TestBacaTidakMenyegarkanDanTidakMemperpanjang(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	m, st, p, id := siapkanSegar(t, &now, now.Add(time.Minute))
	st.ttl[id] = 42 * time.Second
	now = now.Add(time.Hour - time.Minute)
	p.err = errors.New("tidak boleh dipanggil")
	s, err := m.Baca(context.Background(), id)
	if err != nil || s.Refresh != "rt-1" || p.panggil != 0 || st.ttl[id] != 42*time.Second {
		t.Errorf("Baca: %+v %v panggil=%d ttl=%v", s, err, p.panggil, st.ttl[id])
	}
	now = now.Add(8 * time.Hour)
	if _, err := m.Baca(context.Background(), id); !errors.Is(err, ErrTidakAda) {
		t.Errorf("Baca tetap menegakkan batas mutlak: %v", err)
	}
}

func TestSandiBulatPenuhDanAAD(t *testing.T) {
	kunci := []byte(strings.Repeat("k", PanjangKunciSandi))
	s, err := BuatSandi(kunci)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := s.Segel([]byte("rahasia"), []byte("session:a"))
	b, _ := s.Segel([]byte("rahasia"), []byte("session:a"))
	if string(a) == string(b) || strings.Contains(string(a), "rahasia") {
		t.Error("setiap penyegelan harus memakai nonce baru dan tidak memuat teks asli")
	}
	if teks, err := s.Buka(a, []byte("session:a")); err != nil || string(teks) != "rahasia" {
		t.Fatalf("Buka: %q %v", teks, err)
	}
	if _, err := s.Buka(a, []byte("session:b")); err == nil {
		t.Error("ciphertext tidak boleh terbuka untuk session lain")
	}
	lain, _ := BuatSandi([]byte(strings.Repeat("x", PanjangKunciSandi)))
	if _, err := lain.Buka(a, []byte("session:a")); err == nil {
		t.Error("kunci lain tidak boleh membuka")
	}
	if _, err := BuatSandi([]byte("pendek")); err == nil || strings.Contains(err.Error(), "pendek") {
		t.Errorf("kunci bukan 32 byte harus ditolak tanpa menyebut isinya: %v", err)
	}
}
