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

// kebijakanKonten menyusun CSP untuk halaman dan API: hanya sumber same-origin. Astro (frontend-final) memasang skrip
// inline kecil untuk hidrasi island; hash SHA-256-nya dihitung saat build (scripts/bangun-web-final.cjs, web/csp-hash-skrip.txt)
// dan diberikan lewat hashSkrip, sehingga script-src tidak butuh 'unsafe-inline' atau nonce (data-awal dan JSON-LD berupa
// blok data, bukan skrip). frame-src hanya mengizinkan peta Google di halaman Tentang Kami (iframe maps.google.com yang
// dialihkan ke www.google.com); style-src memakai 'unsafe-inline' karena frontend mengatur properti CSS lewat atribut style
// dan /admin/masuk memuat blok <style>. blob: di img-src/connect-src untuk pratinjau gambar yang baru dipilih admin dan
// pembacaannya saat Simpan (URL blob: dibuat browser dari berkas lokal pengguna). frame-ancestors 'none' mencegah clickjacking. /media menetapkan CSP-nya sendiri
// (default-src 'none'; sandbox) yang menimpa nilai ini karena diset belakangan di handler.
func kebijakanKonten(hashSkrip []string) string {
	skrip := "'self'"
	for _, h := range hashSkrip {
		skrip += " '" + h + "'"
	}
	return "default-src 'self'; script-src " + skrip + "; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; " +
		"font-src 'self'; connect-src 'self' blob:; frame-src https://maps.google.com https://www.google.com; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"
}

// headerKeamanan memasang header keamanan standar pada semua respons. hsts hanya diaktifkan di production
// (browser mengabaikan HSTS lewat http, tetapi dev di localhost tidak perlu menyimpannya).
func headerKeamanan(hsts bool, hashSkrip []string) func(http.Handler) http.Handler {
	csp := kebijakanKonten(hashSkrip)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Content-Security-Policy", csp)
			if hsts {
				h.Set("Strict-Transport-Security", "max-age=31536000")
			}
			next.ServeHTTP(w, r)
		})
	}
}
