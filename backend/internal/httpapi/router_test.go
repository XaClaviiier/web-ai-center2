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
