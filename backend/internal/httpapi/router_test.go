package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
)

type pingerPalsu struct{ err error }

func (p pingerPalsu) Ping(context.Context) error { return p.err }

func routerUji(p Pinger) http.Handler {
	return routerUjiDengan(p, kontenPalsu{})
}

func routerUjiDengan(p Pinger, k konten.BacaPublik) http.Handler {
	return NewRouter(Deps{Log: slog.New(slog.NewTextHandler(io.Discard, nil)), DB: p, Konten: k})
}

func panggil(h http.Handler, metode, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(metode, path, nil))
	return rec
}

func TestHealthOK(t *testing.T) {
	rec := panggil(routerUji(pingerPalsu{}), http.MethodGet, "/api/health")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["status"] != "ok" || got["db"] != "ok" {
		t.Fatalf("isi salah: %v", got)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type %q", ct)
	}
}

func TestHealthDBMati(t *testing.T) {
	rec := panggil(routerUji(pingerPalsu{err: errors.New("koneksi putus: password=rahasia")}), http.MethodGet, "/api/health")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d", rec.Code)
	}
	var got bentukGalat
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Error.Code != kodeDBTidakAda {
		t.Fatalf("kode %q", got.Error.Code)
	}
	if strings.Contains(rec.Body.String(), "rahasia") {
		t.Fatal("detail internal bocor ke klien")
	}
}

func TestNotFoundFormatStandar(t *testing.T) {
	rec := panggil(routerUji(pingerPalsu{}), http.MethodGet, "/api/tidak-ada")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
	var got bentukGalat
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || got.Error.Code != kodeTidakDitemukan {
		t.Fatalf("bentuk galat salah: %s (%v)", rec.Body.String(), err)
	}
}

func TestMetodeSalah(t *testing.T) {
	rec := panggil(routerUji(pingerPalsu{}), http.MethodPost, "/api/health")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestPanicMenjadi500TanpaBocor(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := pulihkanPanic(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("rahasia internal")
	}))
	rec := panggil(h, http.MethodGet, "/")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "rahasia") {
		t.Fatal("pesan panic bocor ke klien")
	}
}

func TestHeaderKeamananSemuaRespons(t *testing.T) {
	for _, rute := range []string{"/api/health", "/api/tidak-ada"} {
		rec := panggil(routerUji(pingerPalsu{}), http.MethodGet, rute)
		h := rec.Header()
		if h.Get("X-Content-Type-Options") != "nosniff" || h.Get("X-Frame-Options") != "DENY" || h.Get("Referrer-Policy") == "" {
			t.Errorf("%s: header keamanan kurang: %v", rute, h)
		}
		if csp := h.Get("Content-Security-Policy"); !strings.Contains(csp, "frame-ancestors 'none'") || !strings.Contains(csp, "script-src 'self'") || strings.Contains(csp, "script-src 'self' 'unsafe") {
			t.Errorf("%s: CSP tidak sesuai: %q", rute, csp)
		}
		if h.Get("Strict-Transport-Security") != "" {
			t.Errorf("%s: HSTS tidak boleh aktif tanpa Deps.HSTS", rute)
		}
	}
}

func TestHSTSHanyaBilaDiaktifkan(t *testing.T) {
	h := NewRouter(Deps{Log: slog.New(slog.NewTextHandler(io.Discard, nil)), DB: pingerPalsu{}, Konten: kontenPalsu{}, HSTS: true})
	if got := panggil(h, http.MethodGet, "/api/health").Header().Get("Strict-Transport-Security"); !strings.HasPrefix(got, "max-age=") {
		t.Errorf("HSTS = %q", got)
	}
}

// Hash skrip inline dari build frontend masuk ke script-src; tanpa hash hanya 'self', dan 'unsafe-inline' tidak pernah ada.
func TestKebijakanKontenHashSkrip(t *testing.T) {
	if got := kebijakanKonten(nil); !strings.Contains(got, "script-src 'self';") {
		t.Errorf("tanpa hash: %s", got)
	}
	got := kebijakanKonten([]string{"sha256-AAAA", "sha256-BBBB"})
	if !strings.Contains(got, "script-src 'self' 'sha256-AAAA' 'sha256-BBBB';") {
		t.Errorf("dengan hash: %s", got)
	}
	if strings.Contains(strings.SplitN(got, "style-src", 2)[0], "unsafe-inline") {
		t.Error("script-src tidak boleh memuat unsafe-inline")
	}
}
