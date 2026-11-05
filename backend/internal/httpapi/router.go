// Package httpapi merakit router Chi, middleware, dan handler HTTP.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Pinger memeriksa koneksi ke dependensi (database). Antarmuka kecil agar handler mudah diuji.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Deps adalah dependensi yang disuntikkan ke router.
type Deps struct {
	Log *slog.Logger
	DB  Pinger
}

// NewRouter menyusun handler utama. Semua endpoint API berada di bawah /api.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(logRequest(d.Log))
	r.Use(pulihkanPanic(d.Log))

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		tulisGalat(w, http.StatusNotFound, kodeTidakDitemukan, "Sumber tidak ditemukan")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		tulisGalat(w, http.StatusMethodNotAllowed, kodeMetodeSalah, "Metode tidak diizinkan")
	})

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", health(d.DB))
	})
	return r
}
