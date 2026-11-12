package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/config"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/httpapi"
)

// webNyata adalah web/ di akar repo (hasil `node scripts/bangun-web.cjs`).
func webNyata(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "web"))
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// Kerangka hasil skrip pembangun harus lolos seo.BuatKerangka (production menolak start bila tidak).
func TestWebNyataKerangkaLolosMuatSEO(t *testing.T) {
	d, err := muatSEO(cfgUji(config.EnvProduction, webNyata(t)), logUji)
	if err != nil {
		t.Fatalf("muatSEO production dengan web/ nyata: %v", err)
	}
	if d.KerangkaBerita == nil || d.KerangkaDokumentasi == nil {
		t.Fatal("kerangka berita/dokumentasi tidak termuat")
	}
	if d.Halaman404 == nil || !strings.Contains(string(d.Halaman404), `content="noindex"`) {
		t.Error("404.html tidak termuat atau tidak noindex")
	}
}

// Setiap rute statis yang dijanjikan melayani berkas yang benar-benar ada di web/.
func TestWebNyataSemuaRuteTersaji(t *testing.T) {
	dir := webNyata(t)
	d, err := muatSEO(cfgUji(config.EnvProduction, dir), logUji)
	if err != nil {
		t.Fatal(err)
	}
	h := httpapi.NewRouter(httpapi.Deps{
		Log: slog.New(slog.NewTextHandler(io.Discard, nil)), SEO: d,
		Static: &httpapi.DepsStatic{Dir: dir, DevLoginAktif: true, Halaman404: d.Halaman404, NamaSitus: "AI Center"},
	})
	for _, rute := range []string{"/", "/berita", "/dokumentasi", "/tim", "/tentang-kami", "/layanan/pelatihan", "/layanan/inference-solutions",
		"/layanan/inference-solutions/jejak-alumni", "/layanan/inference-solutions/perencanaan-implan-gigi", "/layanan/inference-solutions/document-intelligence",
		"/admin", "/admin/", "/admin/masuk", "/assets/css/tokens.css", "/assets/js/public.js", "/assets/js/konten.js", "/assets/js/masuk.js", "/assets/fonts/GeneralSans-Regular.woff2"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, rute, nil))
		if rec.Code != http.StatusOK || rec.Body.Len() == 0 {
			t.Errorf("%s: status %d, %d byte", rute, rec.Code, rec.Body.Len())
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tidak-ada", nil))
	if rec.Code != http.StatusNotFound || rec.Header().Get("X-Robots-Tag") != "noindex" || !strings.Contains(rec.Body.String(), "noindex") {
		t.Errorf("404 nyata: %d %q", rec.Code, rec.Header().Get("X-Robots-Tag"))
	}
}
