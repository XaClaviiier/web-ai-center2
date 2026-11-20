package auth

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

const idUji = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ" // 43 karakter sah

func sandiUji(t *testing.T) *Sandi {
	t.Helper()
	s, err := BuatSandi([]byte(strings.Repeat("s", PanjangKunciSandi)))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// Tes integrasi: Redis Docker lokal, DB 15 (dikosongkan oleh testutil.RedisTes).
func TestRedisStoreBulatPenuh(t *testing.T) {
	c := testutil.RedisTes(t)
	st := RedisStore{C: c}
	ctx := context.Background()
	id := idUji
	if !IDSah(id) {
		t.Fatal("ID uji tidak sah")
	}

	want := Session{Identitas: dev, CSRFToken: "csrf-token", DibuatPada: time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)}
	if err := st.Simpan(ctx, id, want, time.Hour); err != nil {
		t.Fatal(err)
	}
	got, err := st.Ambil(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Identitas != want.Identitas || got.CSRFToken != want.CSRFToken || !got.DibuatPada.Equal(want.DibuatPada) {
		t.Fatalf("data tidak kembali utuh: %+v", got)
	}

	ttl := c.TTL(ctx, "session:"+id).Val()
	if ttl <= 0 || ttl > time.Hour {
		t.Errorf("TTL di Redis salah: %v", ttl)
	}
	if err := st.Perpanjang(ctx, id, 10*time.Minute); err != nil {
		t.Fatal(err)
	}
	if ttl := c.TTL(ctx, "session:"+id).Val(); ttl <= 9*time.Minute || ttl > 10*time.Minute {
		t.Errorf("TTL setelah Perpanjang salah: %v", ttl)
	}

	if err := st.Hapus(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Ambil(ctx, id); !errors.Is(err, ErrTidakAda) {
		t.Errorf("setelah Hapus harus ErrTidakAda, dapat %v", err)
	}
	if err := st.Perpanjang(ctx, id, time.Minute); !errors.Is(err, ErrTidakAda) {
		t.Errorf("Perpanjang key hilang harus ErrTidakAda, dapat %v", err)
	}
}

func TestRedisStoreIsiRusakDianggapTidakAda(t *testing.T) {
	c := testutil.RedisTes(t)
	ctx := context.Background()
	if err := c.Set(ctx, "session:"+idUji, "bukan json", time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := (RedisStore{C: c}).Ambil(ctx, idUji); !errors.Is(err, ErrTidakAda) {
		t.Fatalf("isi rusak harus ErrTidakAda, dapat %v", err)
	}
}

// Refresh token disimpan terenkripsi: nilai mentah di Redis tidak memuatnya, dan hanya kunci yang sama yang bisa membuka.
func TestRedisStoreRefreshTokenTerenkripsi(t *testing.T) {
	c := testutil.RedisTes(t)
	ctx := context.Background()
	const rt = "refresh-token-rahasia-123"
	habis := time.Date(2026, 11, 9, 8, 5, 0, 0, time.UTC)
	s := Session{Identitas: dev, Token: Token{Refresh: rt, AksesHabis: habis}, CSRFToken: "c", DibuatPada: time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)}

	if err := (RedisStore{C: c}).Simpan(ctx, idUji, s, time.Minute); err == nil {
		t.Fatal("tanpa Sandi, refresh token tidak boleh disimpan")
	}
	st := RedisStore{C: c, Sandi: sandiUji(t)}
	if err := st.Simpan(ctx, idUji, s, time.Minute); err != nil {
		t.Fatal(err)
	}
	mentah := c.Get(ctx, "session:"+idUji).Val()
	if strings.Contains(mentah, rt) || !strings.Contains(mentah, `"rt":`) {
		t.Fatalf("nilai Redis harus memuat ciphertext, bukan refresh token: %s", mentah)
	}
	got, err := st.Ambil(ctx, idUji)
	if err != nil || got.Refresh != rt || !got.AksesHabis.Equal(habis) {
		t.Fatalf("Ambil: %+v %v", got, err)
	}
	lain, _ := BuatSandi([]byte(strings.Repeat("z", PanjangKunciSandi)))
	if _, err := (RedisStore{C: c, Sandi: lain}).Ambil(ctx, idUji); !errors.Is(err, ErrTidakAda) {
		t.Errorf("kunci lain: harus ErrTidakAda (login ulang), dapat %v", err)
	}
	// Ciphertext terikat ke ID session: disalin ke key lain tidak bisa dibuka.
	idLain := strings.Repeat("Z", 43)
	c.Set(ctx, "session:"+idLain, mentah, time.Minute)
	if _, err := st.Ambil(ctx, idLain); !errors.Is(err, ErrTidakAda) {
		t.Errorf("ciphertext yang dipindah ke session lain harus ditolak, dapat %v", err)
	}
}

func TestRedisStoreGantiBilaAdaTidakMenghidupkanKembali(t *testing.T) {
	c := testutil.RedisTes(t)
	ctx := context.Background()
	st := RedisStore{C: c}
	s := Session{Identitas: dev, CSRFToken: "c", DibuatPada: time.Now().UTC()}
	if err := st.GantiBilaAda(ctx, idUji, s, time.Minute); !errors.Is(err, ErrTidakAda) {
		t.Fatalf("key tidak ada: harus ErrTidakAda, dapat %v", err)
	}
	if n := c.Exists(ctx, "session:"+idUji).Val(); n != 0 {
		t.Fatal("GantiBilaAda tidak boleh membuat key")
	}
	_ = st.Simpan(ctx, idUji, s, time.Minute)
	s.Nama = "Baru"
	if err := st.GantiBilaAda(ctx, idUji, s, 30*time.Second); err != nil {
		t.Fatal(err)
	}
	if got, _ := st.Ambil(ctx, idUji); got.Nama != "Baru" {
		t.Errorf("isi tidak diganti: %+v", got)
	}
	if ttl := c.TTL(ctx, "session:"+idUji).Val(); ttl <= 0 || ttl > 30*time.Second {
		t.Errorf("TTL harus diganti: %v", ttl)
	}
}

func TestManagerDenganRedisNyata(t *testing.T) {
	c := testutil.RedisTes(t)
	ctx := context.Background()
	m := NewManager(RedisStore{C: c}, time.Hour, 8*time.Hour)

	id, s, err := m.Buat(ctx, dev, Token{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.Validasi(ctx, id)
	if err != nil || got.CSRFToken != s.CSRFToken {
		t.Fatalf("Validasi: %+v %v", got, err)
	}
	if err := m.Hapus(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Validasi(ctx, id); !errors.Is(err, ErrTidakAda) {
		t.Fatalf("setelah logout harus ErrTidakAda, dapat %v", err)
	}
}

// penyegarLambat menghitung panggilan dan bisa ditahan sampai tes melepasnya. Aman untuk banyak goroutine.
type penyegarLambat struct {
	panggil atomic.Int32
	masuk   chan struct{} // ditutup saat panggilan pertama masuk
	lepas   chan struct{} // panggilan menunggu sampai ditutup
	sekali  sync.Once
	hasil   Token
}

func (p *penyegarLambat) Segarkan(ctx context.Context, s Session) (Identitas, Token, error) {
	p.panggil.Add(1)
	p.sekali.Do(func() { close(p.masuk) })
	select {
	case <-p.lepas:
	case <-ctx.Done():
		return Identitas{}, Token{}, ctx.Err()
	}
	return Identitas{Sub: s.Sub, Nama: "Disegarkan", Role: RolePegawai, Username: "admin.dev"}, p.hasil, nil
}

func managerRedisSegar(t *testing.T) (*Manager, *penyegarLambat, string) {
	t.Helper()
	c := testutil.RedisTes(t)
	m := NewManager(RedisStore{C: c, Sandi: sandiUji(t)}, time.Hour, 8*time.Hour)
	p := &penyegarLambat{masuk: make(chan struct{}), lepas: make(chan struct{}), hasil: Token{Refresh: "rt-2", AksesHabis: time.Now().Add(5 * time.Minute)}}
	m.Penyegar = p
	// Access token sudah habis sejak awal: request pertama langsung menyegarkan.
	id, _, err := m.Buat(context.Background(), dev, Token{Refresh: "rt-1", AksesHabis: time.Now().Add(-time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	return m, p, id
}

// Banyak request bersamaan pada session yang sama: Keycloak hanya dipanggil sekali, semua mendapat session baru.
func TestPenyegaranBersamaanHanyaSekali(t *testing.T) {
	m, p, id := managerRedisSegar(t)
	const n = 8
	var wg sync.WaitGroup
	hasil := make([]Session, n)
	galat := make([]error, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hasil[i], galat[i] = m.Validasi(context.Background(), id)
		}()
	}
	<-p.masuk
	time.Sleep(200 * time.Millisecond) // beri waktu request lain sampai ke kunci
	close(p.lepas)
	wg.Wait()
	if got := p.panggil.Load(); got != 1 {
		t.Fatalf("Keycloak harus dipanggil sekali, dapat %d", got)
	}
	for i := range n {
		if galat[i] != nil || hasil[i].Nama != "Disegarkan" || hasil[i].Refresh != "rt-2" {
			t.Errorf("request %d: %+v %v", i, hasil[i], galat[i])
		}
	}
}

// Logout yang terjadi selama penyegaran berjalan menang: session tidak dihidupkan kembali.
func TestLogoutSelamaPenyegaranTidakMenghidupkanKembali(t *testing.T) {
	m, p, id := managerRedisSegar(t)
	selesai := make(chan error, 1)
	go func() {
		_, err := m.Validasi(context.Background(), id)
		selesai <- err
	}()
	<-p.masuk
	if err := m.Hapus(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	close(p.lepas)
	if err := <-selesai; !errors.Is(err, ErrTidakAda) {
		t.Fatalf("penyegaran setelah logout harus ErrTidakAda, dapat %v", err)
	}
	if _, err := m.Store.Ambil(context.Background(), id); !errors.Is(err, ErrTidakAda) {
		t.Fatalf("session tidak boleh hidup kembali, dapat %v", err)
	}
	if _, err := m.Validasi(context.Background(), id); !errors.Is(err, ErrTidakAda) {
		t.Errorf("request berikutnya harus 401: %v", err)
	}
}

func TestRedisTransaksiSekaliPakai(t *testing.T) {
	c := testutil.RedisTes(t)
	ctx := context.Background()
	r := RedisTransaksi{C: c}
	tr := TransaksiLogin{Verifier: "v", Nonce: "n", HashIkatan: Hash("ikatan")}
	if err := r.Simpan(ctx, "state-rahasia", tr, MasaTransaksiLogin); err != nil {
		t.Fatal(err)
	}
	keys := c.Keys(ctx, "*").Val()
	if len(keys) != 1 || keys[0] != "oidc:login:"+Hash("state-rahasia") {
		t.Fatalf("key transaksi harus oidc:login:{sha256(state)}: %v", keys)
	}
	if ttl := c.TTL(ctx, keys[0]).Val(); ttl <= 0 || ttl > MasaTransaksiLogin {
		t.Errorf("TTL transaksi: %v", ttl)
	}
	if strings.Contains(c.Get(ctx, keys[0]).Val(), "state-rahasia") {
		t.Error("state mentah tidak boleh disimpan")
	}
	got, err := r.Ambil(ctx, "state-rahasia")
	if err != nil || got != tr {
		t.Fatalf("Ambil: %+v %v", got, err)
	}
	if _, err := r.Ambil(ctx, "state-rahasia"); !errors.Is(err, ErrTransaksiTidakAda) {
		t.Errorf("state kedua kali harus ditolak, dapat %v", err)
	}
	if _, err := r.Ambil(ctx, "state-lain"); !errors.Is(err, ErrTransaksiTidakAda) {
		t.Errorf("state tak dikenal: %v", err)
	}
}
