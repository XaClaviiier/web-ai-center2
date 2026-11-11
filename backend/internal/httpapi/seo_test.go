package httpapi

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/cache"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/ratelimit"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/seo"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

const (
	kerangkaUji = `<!doctype html><html lang="id"><head><meta charset="utf-8"><!--seo:head--></head><body><div id="app"></div><!--seo:konten--><!--seo:data--></body></html>`
	baseUji     = "https://situs.test"
)

type sitemapPalsu struct{ berita, dok []konten.EntriSitemap }

func (s sitemapPalsu) SitemapBerita(context.Context, int32) ([]konten.EntriSitemap, error) {
	return s.berita, nil
}
func (s sitemapPalsu) SitemapDokumentasi(context.Context, int32) ([]konten.EntriSitemap, error) {
	return s.dok, nil
}

func depsSEOUji(t *testing.T) *DepsSEO {
	t.Helper()
	k, err := seo.BuatKerangka(kerangkaUji)
	if err != nil {
		t.Fatal(err)
	}
	return &DepsSEO{
		BaseURL: baseUji, NamaSitus: "AI Center", KerangkaBerita: k, KerangkaDokumentasi: k,
		Sitemap: sitemapPalsu{
			berita: []konten.EntriSitemap{{Slug: "satu", Tanggal: "2026-11-01"}},
			dok:    []konten.EntriSitemap{{Slug: "kegiatan-a", Tanggal: "2026-10-01"}},
		},
	}
}

func kontenSEO(n *atomic.Int64) kontenHitung {
	k := kontenUji(n)
	k.beritaDetail = konten.BeritaDetail{
		BeritaRingkas: konten.BeritaRingkas{ID: 1, Slug: "satu", Judul: `Satu "judul" </title><script>x()</script>`, GambarKey: "berita/a.png", TanggalPublikasi: "2026-11-01"},
		Isi:           "Paragraf pertama <b>tebal</b>.\n\nParagraf kedua & ketiga.",
	}
	k.dokumDetail = konten.DokumentasiDetail{
		ID: 2, Slug: "kegiatan-a", Judul: "Kegiatan A", TanggalPublikasi: "2026-10-01",
		Gambar: []konten.GambarDokumentasi{{Key: "dokumentasi/utama.png", Alt: "Foto utama"}, {Key: "dokumentasi/kedua.png"}},
	}
	return k
}

func routerSEO(t *testing.T, k konten.BacaPublik, s *DepsSEO) http.Handler {
	t.Helper()
	return NewRouter(Deps{Log: slog.New(slog.NewTextHandler(io.Discard, nil)), DB: pingerPalsu{}, Konten: k, SEO: s})
}

func TestHTMLBeritaMemuatMetaKontenDanDataAwal(t *testing.T) {
	var n atomic.Int64
	rec := panggil(routerSEO(t, kontenSEO(&n), depsSEOUji(t)), http.MethodGet, "/berita/satu")
	if rec.Code != 200 || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("status %d, ct %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	html := rec.Body.String()
	for _, mau := range []string{
		`<link rel="canonical" href="` + baseUji + `/berita/satu">`,
		`property="og:image" content="` + baseUji + `/media/berita/a.png"`,
		`property="og:url" content="` + baseUji + `/berita/satu"`,
		`name="twitter:card" content="summary_large_image"`,
		`<script type="application/ld+json">`, `"@type":"NewsArticle"`,
		`<article id="konten-awal">`, `<p>Paragraf kedua &amp; ketiga.</p>`,
		`<script id="data-awal" type="application/json">`,
		`<div id="app"></div>`,
	} {
		if !strings.Contains(html, mau) {
			t.Errorf("HTML tidak memuat %q", mau)
		}
	}
	if strings.Contains(html, "<!--seo:") {
		t.Error("penanda kerangka harus sudah diganti")
	}
	if strings.Contains(html, "<script>x()") || strings.Contains(html, "</title><script>") || strings.Contains(html, "<b>tebal") {
		t.Errorf("judul/isi tidak ter-escape:\n%s", html)
	}
	// Data awal sama dengan respons API publik.
	m := regexp.MustCompile(`(?s)<script id="data-awal" type="application/json">(.*?)</script>`).FindStringSubmatch(html)
	if m == nil {
		t.Fatal("data awal tidak ditemukan")
	}
	api := panggil(routerKonten(kontenSEO(&n).kontenPalsu), http.MethodGet, "/api/berita/satu")
	var a, b any
	if err := json.Unmarshal([]byte(m[1]), &a); err != nil {
		t.Fatalf("data awal bukan JSON: %v", err)
	}
	_ = json.Unmarshal(api.Body.Bytes(), &b)
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Errorf("data awal harus sama dengan API:\n html: %s\n  api: %s", ja, jb)
	}
}

func TestHTMLDokumentasi(t *testing.T) {
	var n atomic.Int64
	rec := panggil(routerSEO(t, kontenSEO(&n), depsSEOUji(t)), http.MethodGet, "/dokumentasi/kegiatan-a")
	html := rec.Body.String()
	if rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
	for _, mau := range []string{
		`content="Dokumentasi kegiatan Kegiatan A — 2026-10-01"`,
		`property="og:image" content="` + baseUji + `/media/dokumentasi/utama.png"`, // headline = urutan 1
		`"@type":"ImageGallery"`, `alt="Foto utama"`, `alt="Kegiatan A"`, // alt default = judul
		`<title>Kegiatan A | AI Center</title>`,
	} {
		if !strings.Contains(html, mau) {
			t.Errorf("HTML tidak memuat %q", mau)
		}
	}
}

func TestHTML404UntukSlugTidakAdaDanTidakSah(t *testing.T) {
	var n atomic.Int64
	h := routerSEO(t, kontenSEO(&n), depsSEOUji(t))
	for _, path := range []string{"/berita/tidak-ada", "/dokumentasi/tidak-ada", "/berita/Slug_Aneh", "/berita/" + strings.Repeat("a", 300)} {
		rec := panggil(h, http.MethodGet, path)
		if rec.Code != 404 || rec.Header().Get("X-Robots-Tag") != "noindex" || !strings.Contains(rec.Body.String(), `content="noindex"`) {
			t.Errorf("%s: status %d, X-Robots-Tag %q, body noindex=%v", path, rec.Code, rec.Header().Get("X-Robots-Tag"), strings.Contains(rec.Body.String(), "noindex"))
		}
		if !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/html") {
			t.Errorf("%s: 404 harus HTML, dapat %q", path, rec.Header().Get("Content-Type"))
		}
	}
	// 404.html kustom dipakai bila tersedia.
	s := depsSEOUji(t)
	s.Halaman404 = seo.Halaman404([]byte("<html><head></head><body>Halaman hilang kustom</body></html>"))
	rec := panggil(routerSEO(t, kontenSEO(&n), s), http.MethodGet, "/berita/tidak-ada")
	if rec.Code != 404 || !strings.Contains(rec.Body.String(), "Halaman hilang kustom") || !strings.Contains(rec.Body.String(), "noindex") {
		t.Errorf("404.html kustom: %d %s", rec.Code, rec.Body)
	}
}

func TestHTMLTanpaKerangkaMembalas503(t *testing.T) {
	var n atomic.Int64
	s := depsSEOUji(t)
	s.KerangkaBerita = nil
	h := routerSEO(t, kontenSEO(&n), s)
	rec := panggil(h, http.MethodGet, "/berita/satu")
	if rec.Code != http.StatusServiceUnavailable || kodeGalat(t, rec) != "kerangka_tidak_ada" {
		t.Errorf("mau 503 kerangka_tidak_ada: %d %s", rec.Code, rec.Body)
	}
	if rec := panggil(h, http.MethodGet, "/dokumentasi/kegiatan-a"); rec.Code != 200 {
		t.Errorf("kerangka dokumentasi ada, harus 200: %d", rec.Code)
	}
}

func TestHTMLGalatDatabaseMenjadi500TanpaBocor(t *testing.T) {
	var n atomic.Int64
	k := kontenSEO(&n)
	k.err = context.DeadlineExceeded
	rec := panggil(routerSEO(t, k, depsSEOUji(t)), http.MethodGet, "/berita/satu")
	if rec.Code != 500 || strings.Contains(rec.Body.String(), "deadline") {
		t.Errorf("mau 500 generik: %d %s", rec.Code, rec.Body)
	}
}

func TestHTMLCacheETagDanLimiter(t *testing.T) {
	rdb := testutil.RedisTes(t)
	var n atomic.Int64
	h := NewRouter(Deps{
		Log: slog.New(slog.NewTextHandler(io.Discard, nil)), DB: pingerPalsu{}, Konten: kontenSEO(&n), SEO: depsSEOUji(t),
		Cache:         &cache.Cache{C: rdb, TTL: time.Minute},
		PublikLimiter: ratelimit.Redis{C: rdb, Jendela: time.Minute},
		BatasPublik:   300,
	})
	rec := panggil(h, http.MethodGet, "/berita/satu")
	etag := rec.Header().Get("ETag")
	if rec.Code != 200 || etag == "" || rec.Header().Get("Cache-Control") != "public, no-cache" {
		t.Fatalf("status %d, etag %q, cc %q", rec.Code, etag, rec.Header().Get("Cache-Control"))
	}
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/berita/satu", nil)
		req.Header.Set("If-None-Match", etag)
		r := httptest.NewRecorder()
		h.ServeHTTP(r, req)
		if r.Code != http.StatusNotModified || r.Body.Len() != 0 {
			t.Fatalf("mau 304 kosong: %d %q", r.Code, r.Body)
		}
	}
	if n.Load() != 1 || bacaHitungan(t, rdb, "192.0.2.1") != 1 {
		t.Errorf("hit/304 tidak boleh membaca database atau menambah hitungan: db %d, hitungan %d", n.Load(), bacaHitungan(t, rdb, "192.0.2.1"))
	}
	// 404 tidak di-cache dan dihitung sebagai miss.
	before := n.Load()
	panggil(h, http.MethodGet, "/berita/tidak-ada")
	panggil(h, http.MethodGet, "/berita/tidak-ada")
	if n.Load()-before != 2 || bacaHitungan(t, rdb, "192.0.2.1") != 3 {
		t.Errorf("404 harus ke database dan terhitung: db %d, hitungan %d", n.Load()-before, bacaHitungan(t, rdb, "192.0.2.1"))
	}
	// Tanpa key HTML bentrok dengan JSON API (sub-key berbeda).
	api := panggil(h, http.MethodGet, "/api/berita/satu")
	if !strings.HasPrefix(api.Header().Get("Content-Type"), "application/json") || strings.Contains(api.Body.String(), "<html") {
		t.Errorf("API harus tetap JSON: %q", api.Header().Get("Content-Type"))
	}
}

func TestSitemapDanRobots(t *testing.T) {
	var n atomic.Int64
	h := routerSEO(t, kontenSEO(&n), depsSEOUji(t))
	rec := panggil(h, http.MethodGet, "/sitemap.xml")
	if rec.Code != 200 || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/xml") {
		t.Fatalf("sitemap: %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	var set struct {
		URL []struct {
			Loc     string `xml:"loc"`
			Lastmod string `xml:"lastmod"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(rec.Body.Bytes(), &set); err != nil {
		t.Fatal(err)
	}
	locs := map[string]string{}
	for _, u := range set.URL {
		locs[u.Loc] = u.Lastmod
	}
	for _, mau := range []string{"/", "/berita", "/dokumentasi", "/tim", "/tentang-kami", "/layanan/pelatihan", "/layanan/inference-solutions"} {
		if _, ok := locs[baseUji+mau]; !ok {
			t.Errorf("sitemap tidak memuat halaman tetap %s", mau)
		}
	}
	if locs[baseUji+"/berita/satu"] != "2026-11-01" || locs[baseUji+"/dokumentasi/kegiatan-a"] != "2026-10-01" || len(set.URL) != 9 {
		t.Errorf("entri konten salah: %v", locs)
	}

	rob := panggil(h, http.MethodGet, "/robots.txt")
	if rob.Code != 200 || !strings.HasPrefix(rob.Header().Get("Content-Type"), "text/plain") ||
		!strings.Contains(rob.Body.String(), "Sitemap: "+baseUji+"/sitemap.xml") || !strings.Contains(rob.Body.String(), "Disallow: /admin") {
		t.Errorf("robots: %d %s", rob.Code, rob.Body)
	}
}

func TestRouteSEOTidakAdaBilaDepsNil(t *testing.T) {
	var n atomic.Int64
	h := routerKonten(kontenSEO(&n).kontenPalsu)
	for _, p := range []string{"/berita/satu", "/sitemap.xml", "/robots.txt"} {
		if rec := panggil(h, http.MethodGet, p); rec.Code != 404 {
			t.Errorf("%s tanpa SEO harus 404, dapat %d", p, rec.Code)
		}
	}
}

// Kriteria selesai (requirement.md bagian 8): HTML awal tanpa JavaScript berisi title, OG, dan JSON-LD; berita yang
// di-Retract langsung 404 (juga di HTML) dan hilang dari sitemap. Store, Redis, dan layanan admin nyata.
func TestE2ESEOHTMLSitemapDanRetract(t *testing.T) {
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
		VALUES ('akan-ditarik', 'Akan Ditarik', E'Paragraf satu.\n\nParagraf dua.', '2026-06-01', 'berita/a.png', 'posted', 'u', 'u'),
		       ('draft-rahasia', 'Draft Rahasia', 'Isi', '2026-06-02', 'berita/b.png', 'draft', 'u', 'u')`); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Pool.Exec(ctx, `INSERT INTO dokumentasi (slug, judul, tanggal_publikasi, status, created_by, updated_by)
		VALUES ('kegiatan-e2e', 'Kegiatan E2E', '2026-05-01', 'posted', 'u', 'u')`); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Pool.Exec(ctx, `INSERT INTO dokumentasi_gambar (dokumentasi_id, key, urutan) SELECT id, 'dokumentasi/e2e.png', 1 FROM dokumentasi WHERE slug = 'kegiatan-e2e'`); err != nil {
		t.Fatal(err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := depsSEOUji(t)
	s.Sitemap = st
	e := siapkanAuthDenganMedia(t, nil, nil, func(d *Deps) {
		d.DB, d.Konten = st, st
		d.Admin = DepsAdmin{Berita: admin.BeritaLayanan{Repo: st.AdminBerita(), Log: log}}
		d.Cache = &cache.Cache{C: rdb, TTL: time.Minute}
		d.PublikLimiter = ratelimit.Redis{C: rdb, Jendela: time.Minute}
		d.BatasPublik = 300
		d.SEO = s
	})
	id, csrf := loginBerhasil(t, e)

	rec := panggil(e.h, http.MethodGet, "/berita/akan-ditarik")
	html := rec.Body.String()
	for _, mau := range []string{"<title>Akan Ditarik | AI Center</title>", `property="og:image" content="` + baseUji + `/media/berita/a.png"`, `"@type":"NewsArticle"`, "<p>Paragraf dua.</p>"} {
		if rec.Code != 200 || !strings.Contains(html, mau) {
			t.Fatalf("HTML awal berita (%d) tidak memuat %q:\n%s", rec.Code, mau, html)
		}
	}
	if rec := panggil(e.h, http.MethodGet, "/dokumentasi/kegiatan-e2e"); rec.Code != 200 || !strings.Contains(rec.Body.String(), baseUji+"/media/dokumentasi/e2e.png") {
		t.Fatalf("HTML dokumentasi: %d %s", rec.Code, rec.Body)
	}
	if rec := panggil(e.h, http.MethodGet, "/berita/draft-rahasia"); rec.Code != 404 {
		t.Errorf("draft tidak boleh terbit di HTML: %d", rec.Code)
	}
	sm := func() string {
		r := panggil(e.h, http.MethodGet, "/sitemap.xml")
		if r.Code != 200 {
			t.Fatalf("sitemap %d", r.Code)
		}
		return r.Body.String()
	}
	peta := sm()
	if !strings.Contains(peta, "/berita/akan-ditarik") || !strings.Contains(peta, "/dokumentasi/kegiatan-e2e") || strings.Contains(peta, "draft-rahasia") {
		t.Fatalf("isi sitemap salah:\n%s", peta)
	}

	if rec := kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/berita/1/retract", `{"version":1}`); rec.Code != 200 {
		t.Fatalf("retract: %d %s", rec.Code, rec.Body)
	}
	if rec := panggil(e.h, http.MethodGet, "/berita/akan-ditarik"); rec.Code != http.StatusNotFound {
		t.Errorf("HTML setelah retract harus 404, dapat %d", rec.Code)
	}
	if peta := sm(); strings.Contains(peta, "akan-ditarik") || !strings.Contains(peta, "kegiatan-e2e") {
		t.Errorf("sitemap harus langsung berubah setelah retract:\n%s", peta)
	}
}
