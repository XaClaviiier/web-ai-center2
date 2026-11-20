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
	panggil int
}

func baruMemStore() *memStore {
	return &memStore{data: map[string]Session{}, ttl: map[string]time.Duration{}}
}

func (m *memStore) Simpan(_ context.Context, id string, s Session, ttl time.Duration) error {
	m.panggil++
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

func managerUji(st Store, now *time.Time) *Manager {
	m := NewManager(st, time.Hour, 8*time.Hour)
	m.Sekarang = func() time.Time { return *now }
	return m
}

var dev = Identitas{Sub: "dev-admin", Nama: "Admin Dev", Role: "pegawai"}

func TestBuatMembuatIDDanCSRFAcakDanSah(t *testing.T) {
	now := time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)
	st := baruMemStore()
	m := managerUji(st, &now)

	id1, s1, err := m.Buat(context.Background(), dev)
	if err != nil {
		t.Fatal(err)
	}
	id2, s2, err := m.Buat(context.Background(), dev)
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
	id, _, _ := m.Buat(context.Background(), dev)

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
	id, _, _ := m.Buat(context.Background(), dev)

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
	id, _, _ := m.Buat(context.Background(), dev)
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
