// Package httpapi merakit router Chi, middleware, dan handler HTTP.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"net/netip"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/cache"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/lock"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/ratelimit"
)

// Pinger memeriksa koneksi ke dependensi (database). Antarmuka kecil agar handler mudah diuji.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Deps adalah dependensi yang disuntikkan ke router.
type Deps struct {
	Log    *slog.Logger
	DB     Pinger
	Konten konten.BacaPublik
	Auth   DepsAuth

	Media          media.Store
	UploadMaxBytes int64 // batas ukuran satu gambar (UPLOAD_MAX_MB)

	Admin DepsAdmin

	// Cache dan limiter baca publik; nil = dimatikan (tes tanpa Redis).
	Cache         *cache.Cache
	PublikLimiter ratelimit.Limiter
	BatasPublik   int // batas miss per IP per menit

	// Limiter admin per user ID per menit (0 = mati) memakai Auth.Limiter; ProxyTepercaya untuk X-Forwarded-For (RP-03).
	BatasAdminBaca  int
	BatasAdminTulis int
	ProxyTepercaya  []netip.Prefix

	// Soft lock edit (docs/database.md section 4); nil = dimatikan. LockHeartbeatDetik disarankan ke klien.
	Lock               *lock.Lock
	LockHeartbeatDetik int

	// Halaman detail ber-SEO, sitemap.xml, dan robots.txt; nil = tidak dipasang.
	SEO *DepsSEO

	// Frontend statis (web/): halaman publik, /assets, /admin, 404 HTML; nil = tidak dipasang (404 selalu JSON).
	Static *DepsStatic

	// HSTS mengaktifkan Strict-Transport-Security (true hanya di production).
	HSTS bool
	// HashSkrip adalah hash CSP ("sha256-...") skrip inline yang diizinkan (hasil build frontend); kosong = hanya 'self'.
	HashSkrip []string
}

// DepsAdmin adalah layanan penulisan konten oleh admin.
type DepsAdmin struct {
	Berita      LayananBerita
	Dokumentasi LayananDokumentasi
	Testimoni   LayananTestimoni
	Tim         LayananTim
	Log         LayananLog // log aktivitas (baca-saja); nil = rute tidak dipasang
}

// NewRouter menyusun handler utama. Semua endpoint API berada di bawah /api.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(headerKeamanan(d.HSTS, d.HashSkrip))
	r.Use(ipKlienTepercaya(d.ProxyTepercaya))
	r.Use(logRequest(d.Log))
	r.Use(pulihkanPanic(d.Log))

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		tulisGalat(w, http.StatusNotFound, kodeTidakDitemukan, "Sumber tidak ditemukan")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		tulisGalat(w, http.StatusMethodNotAllowed, kodeMetodeSalah, "Metode tidak diizinkan")
	})

	pub := publik{log: d.Log, src: d.Konten, cache: d.Cache, limiter: d.PublikLimiter, batas: d.BatasPublik}
	mh := mediaHandler{log: d.Log, store: d.Media, maxBytes: d.UploadMaxBytes}
	r.Route("/api", func(api chi.Router) {
		api.Get("/health", health(d.DB))

		// Baca publik (tanpa autentikasi; hanya konten posted dan belum di-trash).
		api.Get("/berita", pub.daftarBerita)
		api.Get("/berita/{slug}", pub.detailBerita)
		api.Get("/dokumentasi", pub.daftarDokumentasi)
		api.Get("/dokumentasi/{slug}", pub.detailDokumentasi)
		api.Get("/testimoni", pub.daftarTestimoni)
		api.Get("/tim", pub.daftarTim)

		// Auth: dev-login (hanya development), logout, me. Keycloak menggantikan dev-login nanti.
		ah := authHandler{log: d.Log, DepsAuth: d.Auth}
		api.Route("/auth", ah.routerAuth)

		// Admin: semua route di sini wajib Origin sah, login, dan token CSRF (metode ubah-data).
		api.Route("/admin", func(adm chi.Router) {
			adm.Use(noStore, cekOrigin(d.Auth.AllowedOrigins), ah.wajibLogin, batasAdmin(d.Auth.Limiter, d.BatasAdminBaca, d.BatasAdminTulis, d.Log), wajibCSRF)
			adm.Post("/upload", mh.upload)
			if d.Admin.Log != nil {
				adm.Get("/log-aktivitas", logAdminHandler{log: d.Log, svc: d.Admin.Log}.daftar)
			}
			adm.Route("/berita", denganBatalCache(d.Cache, []string{cache.Berita, cache.Sitemap}, false, d.Log, denganLock(d.lockBerita(), beritaAdminHandler{log: d.Log, svc: d.Admin.Berita}.routerBerita)))
			adm.Route("/dokumentasi", denganBatalCache(d.Cache, []string{cache.Dokumentasi, cache.Sitemap}, false, d.Log, denganLock(d.lockDokumentasi(), dokumentasiAdminHandler{log: d.Log, svc: d.Admin.Dokumentasi}.routerDokumentasi)))
			adm.Route("/testimoni", denganBatalCache(d.Cache, []string{cache.Testimoni}, false, d.Log, denganLock(d.lockTestimoni(), testimoniAdminHandler{log: d.Log, svc: d.Admin.Testimoni}.routerTestimoni)))
			adm.Route("/tim", denganBatalCache(d.Cache, []string{cache.Tim}, true, d.Log, denganLock(d.lockTim(), timAdminHandler{log: d.Log, svc: d.Admin.Tim}.routerTim)))
		})
	})

	// HTML detail ber-SEO (kerangka Astro + meta), sitemap.xml, robots.txt: dilayani Go (RP-12).
	if d.SEO != nil {
		sh := seoHandler{publik: pub, s: *d.SEO}
		// GET dan HEAD (crawler/pemeriksa tautan memakai HEAD); server Go membuang badan untuk HEAD.
		for _, rute := range []struct {
			pola string
			h    http.HandlerFunc
		}{{"/berita/{slug}", sh.berita}, {"/dokumentasi/{slug}", sh.dokumentasi}, {"/sitemap.xml", sh.sitemap}, {"/robots.txt", sh.robots}} {
			r.Get(rute.pola, rute.h)
			r.Head(rute.pola, rute.h)
		}
	}

	// Frontend statis (web/) dan 404 HTML untuk path di luar /api. Dipasang setelah rute SEO agar
	// /berita (daftar, statis) dan /berita/{slug} (SEO) hidup berdampingan.
	if d.Static != nil {
		pasangStatic(r, *d.Static)
		r.NotFound(staticHandler{d: *d.Static}.tidakDitemukan)
	}

	// Gambar disajikan di luar /api (URL publik /media/{key}); GET/HEAD saja, tanpa login.
	r.Get("/media/*", mh.sajikan)
	r.Head("/media/*", mh.sajikan)
	return r
}
