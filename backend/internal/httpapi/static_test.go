package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rutaUji memetakan URL bersih ke isi berkas fixture yang dibuat siapkanWeb.
var rutaUji = map[string]string{
	"/":                            "index.html",
	"/berita":                      "berita/index.html",
	"/dokumentasi":                 "dokumentasi/index.html",
	"/tim":                         "tim/index.html",
	"/tentang-kami":                "tentang-kami/index.html",
	"/layanan/pelatihan":           "layanan/pelatihan/index.html",
	"/layanan/inference-solutions": "layanan/inference-solutions/index.html",
	"/layanan/inference-solutions/jejak-alumni":            "layanan/inference-solutions/jejak-alumni/index.html",
	"/layanan/inference-solutions/perencanaan-implan-gigi": "layanan/inference-solutions/perencanaan-implan-gigi/index.html",
	"/layanan/inference-solutions/document-intelligence":   "layanan/inference-solutions/document-intelligence/index.html",
}

// siapkanWeb membuat folder web palsu. File "rahasia.txt" ada di LUAR dir web untuk menguji traversal.
func siapkanWeb(t *testing.T) (dir string) {
	t.Helper()
	induk := t.TempDir()
	dir = filepath.Join(induk, "web")
	tulis := func(rel, isi string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(isi), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, berkas := range rutaUji {
		tulis(berkas, "<html>halaman "+berkas+"</html>")
	}
	tulis("admin/index.html", "<html>admin</html>")
	tulis("admin/masuk.html", "<html>masuk</html>")
	tulis("404.html", `<html><head></head><body>kustom 404</body></html>`)
	tulis("assets/css/a.css", "body{}")
	tulis("assets/js/a.js", "var a=1;")
	tulis("assets/fonts/f.woff2", "woff2-data")
	tulis("assets/img/g.svg", "<svg/>")
	tulis("_astro/app.AbC123.js", "var b=1;")
	tulis("admin/assets/index-XyZ.js", "var c=1;")
	tulis("admin/masuk.css", "body{}")
	tulis("admin/masuk.js", "var m=1;")
	if err := os.WriteFile(filepath.Join(induk, "rahasia.txt"), []byte("RAHASIA"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func routerStatis(t *testing.T, devLogin bool, seoDeps *DepsSEO) http.Handler {
	t.Helper()
	dir := siapkanWeb(t)
	return NewRouter(Deps{
		Log: slog.New(slog.NewTextHandler(io.Discard, nil)), DB: pingerPalsu{}, Konten: kontenPalsu{}, SEO: seoDeps,
		Static: &DepsStatic{Dir: dir, DevLoginAktif: devLogin, Halaman404: []byte(`<html><head></head><body>kustom 404</body></html>`), NamaSitus: "AI Center"},
	})
}

func TestStaticRuteBersih(t *testing.T) {
	h := routerStatis(t, false, nil)
	for rute, berkas := range rutaUji {
		rec := panggil(h, http.MethodGet, rute)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status %d", rute, rec.Code)
			continue
		}
		if !strings.Contains(rec.Body.String(), "halaman "+berkas) {
			t.Errorf("%s: isi salah: %q", rute, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); ct != tipeHTML {
			t.Errorf("%s: Content-Type %q", rute, ct)
		}
		if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
			t.Errorf("%s: Cache-Control %q", rute, cc)
		}
		if rec.Header().Get("ETag") == "" {
			t.Errorf("%s: tanpa ETag", rute)
		}
	}
}

func TestStaticETagDanHEAD(t *testing.T) {
	h := routerStatis(t, false, nil)
	rec := panggil(h, http.MethodGet, "/tim")
	etag := rec.Header().Get("ETag")
	req := httptest.NewRequest(http.MethodGet, "/tim", nil)
	req.Header.Set("If-None-Match", etag)
	cond := httptest.NewRecorder()
	h.ServeHTTP(cond, req)
	if cond.Code != http.StatusNotModified {
		t.Errorf("If-None-Match: status %d, mau 304", cond.Code)
	}
	head := panggil(h, http.MethodHead, "/tim")
	if head.Code != http.StatusOK || head.Body.Len() != 0 || head.Header().Get("ETag") != etag {
		t.Errorf("HEAD: %d body=%d etag=%q", head.Code, head.Body.Len(), head.Header().Get("ETag"))
	}
}

func TestStaticPathTakDikenalMenjadi404HTML(t *testing.T) {
	h := routerStatis(t, false, nil)
	// Termasuk nama berkas mentah: hanya URL bersih yang dilayani.
	for _, p := range []string{"/acak", "/index.html", "/berita.html", "/layanan", "/layanan/pelatihan/", "/tim/", "/admin/lain", "/berita/", "/media"} {
		rec := panggil(h, http.MethodGet, p)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: status %d, mau 404", p, rec.Code)
			continue
		}
		if ct := rec.Header().Get("Content-Type"); ct != tipeHTML {
			t.Errorf("%s: Content-Type %q, mau HTML", p, ct)
		}
		if rec.Header().Get("X-Robots-Tag") != "noindex" {
			t.Errorf("%s: tanpa X-Robots-Tag noindex", p)
		}
		if !strings.Contains(rec.Body.String(), "kustom 404") {
			t.Errorf("%s: bukan 404.html: %q", p, rec.Body.String())
		}
	}
}

func TestStatic404DiBawahAPITetapJSON(t *testing.T) {
	h := routerStatis(t, false, nil)
	for _, p := range []string{"/api/acak", "/api/admin/tidak-ada"} {
		rec := panggil(h, http.MethodGet, p)
		if rec.Code != http.StatusNotFound && rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: status %d", p, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("%s: Content-Type %q, mau JSON", p, ct)
		}
	}
	if rec := panggil(h, http.MethodGet, "/api/acak"); rec.Code != http.StatusNotFound {
		t.Errorf("/api/acak: %d", rec.Code)
	}
}

func TestStaticAset(t *testing.T) {
	h := routerStatis(t, false, nil)
	kasus := map[string]string{
		"/assets/css/a.css":     "text/css",
		"/assets/js/a.js":       "text/javascript",
		"/assets/fonts/f.woff2": "font/woff2",
		"/assets/img/g.svg":     "image/svg+xml",
	}
	for p, tipe := range kasus {
		rec := panggil(h, http.MethodGet, p)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status %d", p, rec.Code)
			continue
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, tipe) {
			t.Errorf("%s: Content-Type %q, mau %q", p, ct, tipe)
		}
		if rec.Header().Get("ETag") == "" || rec.Header().Get("Cache-Control") != "no-cache" {
			t.Errorf("%s: header cache salah: %v", p, rec.Header())
		}
	}
}

func TestStaticAsetTanpaDaftarDirektoriDanTraversal(t *testing.T) {
	h := routerStatis(t, false, nil)
	for _, p := range []string{
		"/assets", "/assets/", "/assets/css", "/assets/css/", // direktori
		"/assets/tidak-ada.css",
		"/assets/../rahasia.txt", "/assets/../admin/index.html", "/assets/css/../../rahasia.txt",
		"/assets/%2e%2e/rahasia.txt", "/assets/%2e%2e%2frahasia.txt", "/assets/..%2f..%2frahasia.txt",
		"/assets/css/..\\..\\rahasia.txt", "/assets/%5c..%5c..%5crahasia.txt",
	} {
		rec := panggil(h, http.MethodGet, p)
		if rec.Code == http.StatusOK {
			t.Errorf("%s: dilayani (status 200), seharusnya tidak", p)
		}
		if strings.Contains(rec.Body.String(), "RAHASIA") || strings.Contains(rec.Body.String(), "admin") && p != "/assets/css/" {
			t.Errorf("%s: membocorkan isi: %q", p, rec.Body.String())
		}
	}
}

func TestStaticAdminTidakDiCache(t *testing.T) {
	h := routerStatis(t, false, nil)
	for _, p := range []string{"/admin", "/admin/"} {
		rec := panggil(h, http.MethodGet, p)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "admin") {
			t.Errorf("%s: %d %q", p, rec.Code, rec.Body.String())
		}
		if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
			t.Errorf("%s: Cache-Control %q, mau no-store (RP-05)", p, cc)
		}
	}
}

func TestStaticMasukHanyaDiDevelopment(t *testing.T) {
	dev := panggil(routerStatis(t, true, nil), http.MethodGet, "/admin/masuk")
	if dev.Code != http.StatusOK || !strings.Contains(dev.Body.String(), "masuk") || dev.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("development: %d cc=%q body=%q", dev.Code, dev.Header().Get("Cache-Control"), dev.Body.String())
	}
	prod := panggil(routerStatis(t, false, nil), http.MethodGet, "/admin/masuk")
	if prod.Code != http.StatusNotFound || !strings.Contains(prod.Body.String(), "kustom 404") {
		t.Errorf("production: %d %q (mau 404 HTML)", prod.Code, prod.Body.String())
	}
	if strings.Contains(prod.Body.String(), "<html>masuk") {
		t.Error("production membocorkan halaman masuk")
	}
}

// /berita (daftar, statis) dan /berita/{slug} (SEO) hidup berdampingan.
func TestStaticBerdampinganDenganRuteSEO(t *testing.T) {
	h := routerStatis(t, false, &DepsSEO{BaseURL: "https://situs.test", NamaSitus: "AI Center"})
	if rec := panggil(h, http.MethodGet, "/berita"); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "berita/index.html") {
		t.Errorf("/berita: %d %q", rec.Code, rec.Body.String())
	}
	// Tanpa kerangka, detail dilayani handler SEO (503 JSON), bukan handler statis.
	rec := panggil(h, http.MethodGet, "/berita/slug-apa-saja")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("/berita/slug: %d, mau 503 dari handler SEO", rec.Code)
	}
	if rec := panggil(h, http.MethodGet, "/robots.txt"); rec.Code != http.StatusOK {
		t.Errorf("/robots.txt: %d", rec.Code)
	}
}

func TestTanpaStaticNotFoundTetapJSON(t *testing.T) {
	rec := panggil(routerUji(pingerPalsu{}), http.MethodGet, "/acak")
	if rec.Code != http.StatusNotFound || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Errorf("tanpa Static: %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
}

func TestStaticDirKosongTidakPanik(t *testing.T) {
	h := NewRouter(Deps{Log: slog.New(slog.NewTextHandler(io.Discard, nil)), DB: pingerPalsu{}, Konten: kontenPalsu{},
		Static: &DepsStatic{Dir: filepath.Join(t.TempDir(), "tidak-ada"), NamaSitus: "AI Center"}})
	for _, p := range []string{"/", "/assets/x.css", "/admin"} {
		if rec := panggil(h, http.MethodGet, p); rec.Code != http.StatusNotFound {
			t.Errorf("%s: %d, mau 404 (berkas tidak ada)", p, rec.Code)
		}
	}
}

// Bundel ber-hash (/_astro, /admin/assets) di-cache lama dan immutable; halaman tidak.
func TestStaticBundelBerHashImmutable(t *testing.T) {
	h := routerStatis(t, false, nil)
	for _, rute := range []string{"/_astro/app.AbC123.js", "/admin/assets/index-XyZ.js"} {
		rec := panggil(h, http.MethodGet, rute)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
			t.Errorf("%s: %d cache=%q", rute, rec.Code, rec.Header().Get("Cache-Control"))
		}
	}
	if rec := panggil(h, http.MethodGet, "/_astro/../rahasia.txt"); rec.Code == http.StatusOK && strings.Contains(rec.Body.String(), "RAHASIA") {
		t.Error("traversal lewat /_astro berhasil")
	}
}

// Aset halaman masuk development ikut 404 di luar development.
func TestStaticAsetMasukHanyaDevelopment(t *testing.T) {
	for _, dev := range []bool{true, false} {
		h := routerStatis(t, dev, nil)
		for _, rute := range []string{"/admin/masuk.css", "/admin/masuk.js"} {
			rec := panggil(h, http.MethodGet, rute)
			if (rec.Code == http.StatusOK) != dev {
				t.Errorf("dev=%v %s: status %d", dev, rute, rec.Code)
			}
		}
	}
}
