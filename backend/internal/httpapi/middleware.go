package httpapi

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// logRequest mencatat satu baris log per request (metode, path, status, durasi, request id).
func logRequest(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			mulai := time.Now()
			next.ServeHTTP(ww, r)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"ms", time.Since(mulai).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}

// pulihkanPanic mengubah panic handler menjadi 500 berformat standar; stack trace hanya ke log.
func pulihkanPanic(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					if rec == http.ErrAbortHandler {
						panic(rec)
					}
					log.Error("panic di handler",
						"panic", rec,
						"stack", string(debug.Stack()),
						"request_id", middleware.GetReqID(r.Context()),
					)
					tulisGalat(w, http.StatusInternalServerError, kodeInternal, "Terjadi kesalahan pada server")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
