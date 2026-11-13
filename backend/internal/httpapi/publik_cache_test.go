package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/cache"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/ratelimit"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

// kontenHitung membungkus kontenPalsu dan menghitung berapa kali database "dibaca".
type kontenHitung struct {
	kontenPalsu
	panggilan *atomic.Int64
	tunda     time.Duration
}

func (k kontenHitung) ListBerita(ctx context.Context, p konten.ParamDaftar) ([]konten.BeritaRingkas, int64, error) {
	k.panggilan.Add(1)
	time.Sleep(k.tunda)
	return k.kontenPalsu.ListBerita(ctx, p)
}

func (k kontenHitung) BeritaBySlug(ctx context.Context, slug string) (konten.BeritaDetail, error) {
	k.panggilan.Add(1)
	time.Sleep(k.tunda)
	return k.kontenPalsu.BeritaBySlug(ctx, slug)
}

func routerCache(t *testing.T, rdb redis.Cmdable, k konten.BacaPublik, batas int) http.Handler {
	t.Helper()
	return NewRouter(Deps{
		Log: slog.New(slog.NewTextHandler(io.Discard, nil)), DB: pingerPalsu{}, Konten: k,
		Cache:         &cache.Cache{C: rdb, TTL: time.Minute},
		PublikLimiter: ratelimit.Redis{C: rdb, Jendela: time.Minute},
		BatasPublik:   batas,
	})
}

// kirimJSON seperti kirim tetapi dengan Content-Type JSON (diwajibkan route admin yang membaca body).
func kirimJSON(h http.Handler, metode, path, body string, o opsi) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metode, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Origin", o.origin)
	req.Header.Set("X-CSRF-Token", o.csrf)
	req.AddCookie(&http.Cookie{Name: "session", Value: o.cookie})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func bacaHitungan(t *testing.T, rdb *redis.Client, ip string) int {
	t.Helper()
	n, err := rdb.Get(context.Background(), "ratelimit:public-miss:"+ip).Int()
	if err == redis.Nil {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func kontenUji(n *atomic.Int64) kontenHitung {
	return kontenHitung{
		panggilan: n,
		kontenPalsu: kontenPalsu{
			berita:       []konten.BeritaRingkas{{ID: 1, Slug: "satu", Judul: "Satu", Ringkasan: "r", GambarKey: "berita/a.png", TanggalPublikasi: "2026-01-01"}},
			total:        1,
			beritaDetail: konten.BeritaDetail{BeritaRingkas: konten.BeritaRingkas{ID: 1, Slug: "satu", Judul: "Satu"}, Isi: "isi"},
		},
	}
}

func TestCacheHitTidakMenyentuhDatabaseMaupunLimiter(t *testing.T) {
	rdb := testutil.RedisTes(t)
	var n atomic.Int64
	h := routerCache(t, rdb, kontenUji(&n), 300)

	rec1 := panggil(h, http.MethodGet, "/api/berita")
	if rec1.Code != 200 || n.Load() != 1 || bacaHitungan(t, rdb, "192.0.2.1") != 1 {
		t.Fatalf("miss pertama: status %d, dbCalls %d, hitungan %d", rec1.Code, n.Load(), bacaHitungan(t, rdb, "192.0.2.1"))
	}
	for i := 0; i < 5; i++ {
		rec := panggil(h, http.MethodGet, "/api/berita")
		if rec.Code != 200 || rec.Body.String() != rec1.Body.String() {
			t.Fatalf("hit ke-%d: status %d, body berbeda", i, rec.Code)
		}
	}
	if n.Load() != 1 || bacaHitungan(t, rdb, "192.0.2.1") != 1 {
		t.Errorf("cache hit tidak boleh membaca database atau menambah hitungan: dbCalls %d, hitungan %d", n.Load(), bacaHitungan(t, rdb, "192.0.2.1"))
	}
}

func TestETagDan304TidakMenambahHitungan(t *testing.T) {
	rdb := testutil.RedisTes(t)
	var n atomic.Int64
	h := routerCache(t, rdb, kontenUji(&n), 300)

	rec := panggil(h, http.MethodGet, "/api/berita/satu")
	etag := rec.Header().Get("ETag")
	if rec.Code != 200 || !strings.HasPrefix(etag, `"`) || !strings.HasSuffix(etag, `"`) {
		t.Fatalf("status %d, ETag %q", rec.Code, etag)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, no-cache" {
		t.Errorf("Cache-Control = %q", cc)
	}
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/berita/satu", nil)
		req.Header.Set("If-None-Match", etag)
		r2 := httptest.NewRecorder()
		h.ServeHTTP(r2, req)
		if r2.Code != http.StatusNotModified || r2.Body.Len() != 0 || r2.Header().Get("ETag") != etag {
			t.Fatalf("304 salah: status %d, body %q, etag %q", r2.Code, r2.Body, r2.Header().Get("ETag"))
		}
	}
	if bacaHitungan(t, rdb, "192.0.2.1") != 1 || n.Load() != 1 {
		t.Errorf("304 menambah hitungan/DB: hitungan %d, dbCalls %d", bacaHitungan(t, rdb, "192.0.2.1"), n.Load())
	}
	// ETag lain tidak cocok -> 200 penuh.
	req := httptest.NewRequest(http.MethodGet, "/api/berita/satu", nil)
	req.Header.Set("If-None-Match", `"lain"`)
	r3 := httptest.NewRecorder()
	h.ServeHTTP(r3, req)
	if r3.Code != 200 {
		t.Errorf("ETag tak cocok harus 200, dapat %d", r3.Code)
	}
}

func TestTidakDiCache_PencarianNotFoundDanHalamanKosong(t *testing.T) {
	rdb := testutil.RedisTes(t)
	var n atomic.Int64
	k := kontenUji(&n)
	h := routerCache(t, rdb, k, 300)

	for _, path := range []string{"/api/berita?q=ab", "/api/berita/tidak-ada"} {
		before := n.Load()
		panggil(h, http.MethodGet, path)
		panggil(h, http.MethodGet, path)
		if n.Load()-before != 2 {
			t.Errorf("%s harus selalu ke database (2 panggilan), dapat %d", path, n.Load()-before)
		}
	}
	if got := bacaHitungan(t, rdb, "192.0.2.1"); got != 4 {
		t.Errorf("empat miss harus terhitung, dapat %d", got)
	}

	// Halaman kosong di luar rentang tidak disimpan; halaman 1 yang kosong disimpan.
	k.berita, k.total = nil, 0
	h = routerCache(t, rdb, k, 300)
	before := n.Load()
	panggil(h, http.MethodGet, "/api/berita?page=9")
	panggil(h, http.MethodGet, "/api/berita?page=9")
	if n.Load()-before != 2 {
		t.Errorf("halaman kosong >1 tidak boleh di-cache, dbCalls %d", n.Load()-before)
	}
	before = n.Load()
	panggil(h, http.MethodGet, "/api/berita?page=1")
	panggil(h, http.MethodGet, "/api/berita?page=1")
	if n.Load()-before != 1 {
		t.Errorf("halaman 1 kosong harus di-cache, dbCalls %d", n.Load()-before)
	}
}

func TestLimiterPublikHanyaSaatMissDan429(t *testing.T) {
	rdb := testutil.RedisTes(t)
	var n atomic.Int64
	h := routerCache(t, rdb, kontenUji(&n), 3)

	// Isi cache sekali (miss 1), lalu habiskan sisa jatah dengan miss berbeda (q tidak di-cache).
	if rec := panggil(h, http.MethodGet, "/api/berita"); rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	panggil(h, http.MethodGet, "/api/berita?q=aa")
	panggil(h, http.MethodGet, "/api/berita?q=bb")
	rec := panggil(h, http.MethodGet, "/api/berita?q=cc")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("miss ke-4 harus 429, dapat %d", rec.Code)
	}
	if ra, err := strconv.Atoi(rec.Header().Get("Retry-After")); err != nil || ra < 1 || ra > 60 {
		t.Errorf("Retry-After = %q", rec.Header().Get("Retry-After"))
	}
	// Yang sudah ada di cache tetap dilayani walau jatah habis.
	for i := 0; i < 5; i++ {
		if rec := panggil(h, http.MethodGet, "/api/berita"); rec.Code != 200 {
			t.Fatalf("cache hit harus tetap 200 saat jatah habis, dapat %d", rec.Code)
		}
	}
}

func TestLimiterPerIPTerpisah(t *testing.T) {
	rdb := testutil.RedisTes(t)
	var n atomic.Int64
	h := routerCache(t, rdb, kontenUji(&n), 1)
	dari := func(ip, path string) int {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.RemoteAddr = ip + ":4000"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	if dari("203.0.113.1", "/api/berita?q=aa") != 200 || dari("203.0.113.1", "/api/berita?q=bb") != 429 {
		t.Fatal("IP pertama harus dibatasi setelah 1 miss")
	}
	if dari("203.0.113.2", "/api/berita?q=aa") != 200 {
		t.Error("IP lain tidak boleh ikut terbatasi")
	}
}

func TestAntiStampede(t *testing.T) {
	rdb := testutil.RedisTes(t)
	var n atomic.Int64
	k := kontenUji(&n)
	k.tunda = 150 * time.Millisecond
	h := routerCache(t, rdb, k, 1000)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rec := panggil(h, http.MethodGet, "/api/berita"); rec.Code != 200 {
				t.Errorf("status %d", rec.Code)
			}
		}()
	}
	wg.Wait()
	if n.Load() != 1 {
		t.Errorf("20 permintaan bersamaan pada key dingin harus 1 pembacaan database, dapat %d", n.Load())
	}
}

func TestRedisMatiTidakMematikanSitusPublik(t *testing.T) {
	mati := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 200 * time.Millisecond, MaxRetries: -1})
	t.Cleanup(func() { _ = mati.Close() })
	var n atomic.Int64
	h := routerCache(t, mati, kontenUji(&n), 300)
	for i := 0; i < 2; i++ {
		rec := panggil(h, http.MethodGet, "/api/berita")
		if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"slug":"satu"`) {
			t.Fatalf("Redis mati: status %d, body %s", rec.Code, rec.Body)
		}
	}
	if n.Load() != 2 {
		t.Errorf("tanpa cache setiap request harus ke database, dapat %d", n.Load())
	}
}

func TestBatalCacheAdminMenaikkanVersiSesuaiAturan(t *testing.T) {
	rdb := testutil.RedisTes(t)
	svc := &layananPalsu{hasil: admin.Berita{ID: 1}}
	e := siapkanAuth(t, nil, func(d *Deps) {
		d.Admin = DepsAdmin{Berita: svc}
		d.Cache = &cache.Cache{C: rdb, TTL: time.Minute}
	})
	id, csrf := loginBerhasil(t, e)
	o := opsi{origin: originSah, csrf: csrf, cookie: id}
	ver := func() int64 {
		v, err := rdb.Get(context.Background(), "cache:ver:berita").Int64()
		if err == redis.Nil {
			return 0
		}
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	json := `{"version":1}`
	ubah := `{"judul":"x","isi":"y","tanggalPublikasi":"01-01-2026","gambar":"berita/a.png","version":1}`

	// Simpan/hapus/pulihkan draft tidak mengubah apa pun di publik.
	kirimJSON(e.h, "PUT", "/api/admin/berita/1", ubah, o)
	kirimJSON(e.h, "DELETE", "/api/admin/berita/1", "", o)
	kirimJSON(e.h, "POST", "/api/admin/berita/1/restore", "", o)
	kirimJSON(e.h, "GET", "/api/admin/berita/1", "", o)
	if ver() != 0 {
		t.Fatalf("aksi draft tidak boleh membatalkan cache, versi %d", ver())
	}
	// Post dan Retract membatalkan.
	if rec := kirimJSON(e.h, "POST", "/api/admin/berita/1/post", json, o); rec.Code != 200 || ver() != 1 {
		t.Fatalf("Post: status %d, versi %d", rec.Code, ver())
	}
	if rec := kirimJSON(e.h, "POST", "/api/admin/berita/1/retract", json, o); rec.Code != 200 || ver() != 2 {
		t.Fatalf("Retract: status %d, versi %d", rec.Code, ver())
	}
	// Post yang gagal tidak membatalkan.
	svc.err = admin.ErrVersiBentrok
	if rec := kirimJSON(e.h, "POST", "/api/admin/berita/1/post", json, o); rec.Code == 200 || ver() != 2 {
		t.Fatalf("Post gagal: status %d, versi %d", rec.Code, ver())
	}
}

// Kriteria selesai backend (requirement.md §8): berita yang di-Retract langsung 404 di publik, dengan store nyata.
func TestE2ERetractLangsung404DiPublik(t *testing.T) {
	url := testutil.DatabaseTes(t)
	rdb := testutil.RedisTes(t)
	ctx, batal := context.WithTimeout(context.Background(), 30*time.Second)
	defer batal()
	st, err := store.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(st.Close)
	testutil.BersihkanKonten(t, st.Pool)
	if _, err := st.Pool.Exec(ctx, `INSERT INTO berita (slug, judul, isi, tanggal_publikasi, gambar, status, created_by, updated_by)
		VALUES ('akan-ditarik', 'Akan Ditarik', 'Isi', '2026-06-01', 'berita/a.png', 'posted', 'u', 'u')`); err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	e := siapkanAuthDenganMedia(t, nil, nil, func(d *Deps) {
		d.DB, d.Konten = st, st
		d.Admin = DepsAdmin{Berita: admin.BeritaLayanan{Repo: st.AdminBerita(), Log: log}}
		d.Cache = &cache.Cache{C: rdb, TTL: time.Minute}
		d.PublikLimiter = ratelimit.Redis{C: rdb, Jendela: time.Minute}
		d.BatasPublik = 300
	})
	id, csrf := loginBerhasil(t, e)

	for _, path := range []string{"/api/berita/akan-ditarik", "/api/berita"} {
		panggil(e.h, http.MethodGet, path)
		if rec := panggil(e.h, http.MethodGet, path); rec.Code != 200 || !strings.Contains(rec.Body.String(), "akan-ditarik") {
			t.Fatalf("%s sebelum retract: %d %s", path, rec.Code, rec.Body)
		}
	}
	rec := kirimJSON(e.h, "POST", "/api/admin/berita/1/retract", `{"version":1}`, opsi{origin: originSah, csrf: csrf, cookie: id})
	if rec.Code != 200 {
		t.Fatalf("retract: %d %s", rec.Code, rec.Body)
	}
	if rec := panggil(e.h, http.MethodGet, "/api/berita/akan-ditarik"); rec.Code != http.StatusNotFound {
		t.Errorf("detail setelah retract harus 404, dapat %d", rec.Code)
	}
	if rec := panggil(e.h, http.MethodGet, "/api/berita"); rec.Code != 200 || strings.Contains(rec.Body.String(), "akan-ditarik") {
		t.Errorf("daftar setelah retract masih memuat berita: %d %s", rec.Code, rec.Body)
	}
}
