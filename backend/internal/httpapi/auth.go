package httpapi

import (
	"cmp"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/auth"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/ratelimit"
)

// Kode galat auth (stabil untuk klien).
const (
	kodeBelumLogin        = "unauthenticated"
	kodePasswordSalah     = "invalid_credentials"
	kodeCSRFTidakValid    = "csrf_invalid"
	kodeOriginTidakSah    = "origin_invalid"
	kodeTerlaluSering     = "rate_limited"
	kodeSessionBermasalah = "session_unavailable"
	kodeRequestTidakSah   = "invalid_request"
)

const (
	headerCSRF     = "X-CSRF-Token"
	batasBodyLogin = 1024 // byte
	bucketLogin    = "login"
)

// Sesi adalah kebijakan session yang dibutuhkan handler (diimplementasikan *auth.Manager; palsu di tes).
type Sesi interface {
	Buat(ctx context.Context, ident auth.Identitas) (string, auth.Session, error)
	Validasi(ctx context.Context, id string) (auth.Session, error)
	Hapus(ctx context.Context, id string) error
}

// DepsAuth adalah dependensi dan konfigurasi endpoint auth.
type DepsAuth struct {
	Sesi    Sesi
	Limiter ratelimit.Limiter

	DevLoginAktif bool // hanya true bila APP_ENV=development
	AdminPassword string
	// Akun dev kedua (opsional; kosong = nonaktif), identitasnya dev-admin-2.
	AdminPassword2 string
	BatasLogin     int // request per menit per IP

	CookieName   string
	CookieSecure bool
	CookieMaxAge time.Duration // = batas mutlak session (security.md §2)

	AllowedOrigins []string
}

// identitasDev adalah pengguna tunggal login development (semua pengguna sama haknya; role hanya label).
var identitasDev = auth.Identitas{Sub: "dev-admin", Nama: "Admin Dev", Role: "pegawai", Username: "admin.dev"}

// identitasDev2 adalah akun dev kedua agar fitur dua-admin (mis. soft lock) bisa dicoba di development.
var identitasDev2 = auth.Identitas{Sub: "dev-admin-2", Nama: "Admin Dev 2", Role: "pegawai", Username: "admin.dev2"}

type ctxKey int

const kunciSession ctxKey = iota

// SessionDariContext mengambil session yang sudah divalidasi middleware wajibLogin.
func SessionDariContext(ctx context.Context) (auth.Session, bool) {
	s, ok := ctx.Value(kunciSession).(auth.Session)
	return s, ok
}

type authHandler struct {
	log *slog.Logger
	DepsAuth
}

// routerAuth mendaftarkan /api/auth/*. dev-login hanya didaftarkan bila DevLoginAktif (selain itu 404 biasa).
func (h authHandler) routerAuth(r chi.Router) {
	r.Use(noStore)
	r.Use(cekOrigin(h.AllowedOrigins))
	if h.DevLoginAktif {
		r.Post("/dev-login", h.devLogin)
	}
	r.With(h.wajibLogin, wajibCSRF).Post("/logout", h.logout)
	r.With(h.wajibLogin).Get("/me", h.me)
}

type sesiJSON struct {
	Nama      string `json:"nama"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CSRFToken string `json:"csrfToken"`
}

func bentukSesi(s auth.Session) sesiJSON {
	return sesiJSON{Nama: s.Nama, Username: cmp.Or(s.Username, s.Sub), Role: s.Role, CSRFToken: s.CSRFToken}
}

func (h authHandler) devLogin(w http.ResponseWriter, r *http.Request) {
	keputusan, err := h.Limiter.Hit(r.Context(), bucketLogin, ipKlien(r), h.BatasLogin)
	if err != nil {
		h.galatSession(w, r, err)
		return
	}
	if !keputusan.Diizinkan {
		w.Header().Set("Retry-After", strconv.Itoa(keputusan.TungguDetik))
		tulisGalat(w, http.StatusTooManyRequests, kodeTerlaluSering, "Terlalu banyak percobaan login, coba lagi nanti")
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, batasBodyLogin))
	if err := dec.Decode(&body); err != nil {
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "Body harus JSON berisi password")
		return
	}
	// Kedua perbandingan selalu dijalankan (waktu konstan, tidak membocorkan akun mana yang cocok).
	cocok1 := passwordCocok(body.Password, h.AdminPassword)
	cocok2 := h.AdminPassword2 != "" && passwordCocok(body.Password, h.AdminPassword2)
	ident := identitasDev
	switch {
	case cocok1:
	case cocok2:
		ident = identitasDev2
	default:
		tulisGalat(w, http.StatusUnauthorized, kodePasswordSalah, "Password salah")
		return
	}

	// ID session selalu baru; session lama pada request yang sama (bila ada) dibuang.
	if c, err := r.Cookie(h.CookieName); err == nil {
		if err := h.Sesi.Hapus(r.Context(), c.Value); err != nil {
			h.log.Warn("gagal menghapus session lama", "err", err)
		}
	}
	id, sess, err := h.Sesi.Buat(r.Context(), ident)
	if err != nil {
		h.galatSession(w, r, err)
		return
	}
	http.SetCookie(w, h.cookie(id, int(h.CookieMaxAge.Seconds())))
	tulisJSON(w, http.StatusOK, bentukSesi(sess))
}

func (h authHandler) me(w http.ResponseWriter, r *http.Request) {
	s, _ := SessionDariContext(r.Context())
	tulisJSON(w, http.StatusOK, bentukSesi(s))
}

func (h authHandler) logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(h.CookieName)
	if err == nil {
		if err := h.Sesi.Hapus(r.Context(), c.Value); err != nil {
			h.galatSession(w, r, err)
			return
		}
	}
	http.SetCookie(w, h.cookie("", -1))
	w.WriteHeader(http.StatusNoContent)
}

// wajibLogin memvalidasi cookie session (dan memperpanjang idle). Tanpa session sah: 401; Redis bermasalah: 503.
func (h authHandler) wajibLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(h.CookieName)
		if err != nil {
			tulisGalat(w, http.StatusUnauthorized, kodeBelumLogin, "Belum login atau session berakhir")
			return
		}
		sess, err := h.Sesi.Validasi(r.Context(), c.Value)
		if errors.Is(err, auth.ErrTidakAda) {
			http.SetCookie(w, h.cookie("", -1)) // buang cookie basi
			tulisGalat(w, http.StatusUnauthorized, kodeBelumLogin, "Belum login atau session berakhir")
			return
		}
		if err != nil {
			h.galatSession(w, r, err)
			return
		}
		ctx := context.WithValue(r.Context(), kunciSession, sess)
		// Aktor untuk log aktivitas: dari session server, bukan dari body request.
		ctx = auth.DenganAktor(ctx, auth.Aktor{Sub: sess.Sub, Username: sess.Username})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h authHandler) cookie(nilai string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     h.CookieName,
		Value:    nilai,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

func (h authHandler) galatSession(w http.ResponseWriter, r *http.Request, err error) {
	h.log.Error("penyimpanan session bermasalah", "err", err, "path", r.URL.Path, "request_id", middleware.GetReqID(r.Context()))
	tulisGalat(w, http.StatusServiceUnavailable, kodeSessionBermasalah, "Layanan session sedang tidak tersedia")
}

// — middleware bersama (dipakai juga oleh route admin nanti) ——————————————————

func metodeAman(m string) bool {
	return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
}

// noStore: respons auth dan admin tidak boleh di-cache (security.md §6).
func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// cekOrigin: untuk metode yang mengubah data, header Origin wajib ada dan tercantum di ALLOWED_ORIGINS.
func cekOrigin(diizinkan []string) func(http.Handler) http.Handler {
	set := make(map[string]struct{}, len(diizinkan))
	for _, o := range diizinkan {
		set[o] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !metodeAman(r.Method) {
				if _, ok := set[r.Header.Get("Origin")]; !ok {
					tulisGalat(w, http.StatusForbidden, kodeOriginTidakSah, "Origin tidak diizinkan")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// wajibCSRF: metode yang mengubah data harus membawa X-CSRF-Token yang sama dengan token di session.
// Dipasang SETELAH wajibLogin.
func wajibCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !metodeAman(r.Method) {
			s, ok := SessionDariContext(r.Context())
			token := r.Header.Get(headerCSRF)
			if !ok || token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(s.CSRFToken)) != 1 {
				tulisGalat(w, http.StatusForbidden, kodeCSRFTidakValid, "Token CSRF tidak valid")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// passwordCocok membandingkan dalam waktu konstan (lewat hash agar panjang tidak bocor).
func passwordCocok(masukan, benar string) bool {
	a, b := sha256.Sum256([]byte(masukan)), sha256.Sum256([]byte(benar))
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}
