package httpapi

import (
	"cmp"
	"context"
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/oauth2"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/auth"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/ratelimit"
)

// Kode galat auth (stabil untuk klien).
const (
	kodeBelumLogin        = "unauthenticated"
	kodeCSRFTidakValid    = "csrf_invalid"
	kodeOriginTidakSah    = "origin_invalid"
	kodeTerlaluSering     = "rate_limited"
	kodeSessionBermasalah = "session_unavailable"
	kodeRequestTidakSah   = "invalid_request"
)

// Kode galat login yang dikirim ke halaman admin lewat /admin/?galat=<kode> (stabil untuk klien).
const (
	galatAksesDitolak     = "akses_ditolak"     // terverifikasi tetapi tidak punya role direktur/pegawai
	galatLoginKedaluwarsa = "login_kedaluwarsa" // state/pengikat tidak ada, kedaluwarsa, atau dipakai ulang
	galatLoginGagal       = "login_gagal"       // galat penyedia, token/tanda tangan/claim tidak sah, jaringan
)

const (
	headerCSRF     = "X-CSRF-Token"
	bucketLogin    = "login"
	bucketCallback = "login-callback"
	halamanAdmin   = "/admin/"
	// batasAkhiriSSO membatasi panggilan logout ke Keycloak agar logout lokal tidak tertahan lama.
	batasAkhiriSSO = 3 * time.Second
)

// Sesi adalah kebijakan session yang dibutuhkan handler (diimplementasikan *auth.Manager; palsu di tes).
type Sesi interface {
	Buat(ctx context.Context, ident auth.Identitas, tok auth.Token) (string, auth.Session, error)
	// Validasi memeriksa session, menyegarkan token bila perlu, dan memperpanjang idle.
	Validasi(ctx context.Context, id string) (auth.Session, error)
	// Baca memeriksa session tanpa menyegarkan token (logout tidak boleh tertahan oleh penyedia identitas).
	Baca(ctx context.Context, id string) (auth.Session, error)
	Hapus(ctx context.Context, id string) error
}

// PenyediaLogin adalah penyedia identitas OIDC (diimplementasikan *auth.OIDC).
type PenyediaLogin interface {
	URLMasuk(state, nonce, verifier string, pilihAkun bool) string
	Tukar(ctx context.Context, kode, verifier, nonce string) (auth.Identitas, auth.Token, error)
	AkhiriSSO(ctx context.Context, refresh string) error
}

// PenyimpanLogin menyimpan transaksi login sampai callback (diimplementasikan auth.RedisTransaksi).
type PenyimpanLogin interface {
	Simpan(ctx context.Context, state string, t auth.TransaksiLogin, ttl time.Duration) error
	// Ambil mengambil sekaligus menghapus transaksi; auth.ErrTransaksiTidakAda bila tidak ada.
	Ambil(ctx context.Context, state string) (auth.TransaksiLogin, error)
}

// DepsAuth adalah dependensi dan konfigurasi endpoint auth.
type DepsAuth struct {
	Sesi      Sesi
	Limiter   ratelimit.Limiter
	OIDC      PenyediaLogin
	Transaksi PenyimpanLogin

	BatasLogin int // request per menit per IP, untuk /login dan /callback masing-masing

	CookieName   string
	CookieSecure bool
	CookieMaxAge time.Duration // = batas mutlak session (security.md §2)

	AllowedOrigins []string
}

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

// routerAuth mendaftarkan /api/auth/*: login dan callback Keycloak, logout, me.
func (h authHandler) routerAuth(r chi.Router) {
	r.Use(noStore)
	r.Use(cekOrigin(h.AllowedOrigins))
	r.Get("/login", h.login)
	r.Get("/callback", h.callback)
	r.With(h.wajibSessionTanpaSegar, wajibCSRF).Post("/logout", h.logout)
	r.With(h.wajibLogin).Get("/me", h.me)
}

type sesiJSON struct {
	Sub       string `json:"sub"`
	Nama      string `json:"nama"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CSRFToken string `json:"csrfToken"`
}

func bentukSesi(s auth.Session) sesiJSON {
	return sesiJSON{Sub: s.Sub, Nama: s.Nama, Username: cmp.Or(s.Username, s.Sub), Role: s.Role, CSRFToken: s.CSRFToken}
}

// login memulai Authorization Code + PKCE: transaksi (verifier, nonce, hash pengikat) disimpan di Redis dengan key
// dari state, cookie pengikat dipasang di browser, lalu browser dialihkan ke Keycloak.
func (h authHandler) login(w http.ResponseWriter, r *http.Request) {
	if !h.batasiLogin(w, r, bucketLogin, false) {
		return
	}
	state, err1 := auth.TokenAcak()
	nonce, err2 := auth.TokenAcak()
	ikatan, err3 := auth.TokenAcak()
	if err := errors.Join(err1, err2, err3); err != nil {
		h.galatSession(w, r, err)
		return
	}
	verifier := oauth2.GenerateVerifier()
	t := auth.TransaksiLogin{Verifier: verifier, Nonce: nonce, HashIkatan: auth.Hash(ikatan)}
	if err := h.Transaksi.Simpan(r.Context(), state, t, auth.MasaTransaksiLogin); err != nil {
		h.galatSession(w, r, err)
		return
	}
	http.SetCookie(w, h.cookieIkatan(ikatan, int(auth.MasaTransaksiLogin.Seconds())))
	http.Redirect(w, r, h.OIDC.URLMasuk(state, nonce, verifier, r.URL.Query().Get("pilih") == "1"), http.StatusFound)
}

// callback menyelesaikan login. Semua kegagalan dialihkan ke /admin/?galat=<kode>; kode, token, dan state tidak
// pernah dicatat di log.
func (h authHandler) callback(w http.ResponseWriter, r *http.Request) {
	// Cookie pengikat selalu dibuang: transaksi hanya sekali pakai, berhasil atau gagal.
	http.SetCookie(w, h.cookieIkatan("", -1))
	if !h.batasiLogin(w, r, bucketCallback, true) {
		return
	}
	q := r.URL.Query()
	state := q.Get("state")
	if state == "" {
		h.gagalLogin(w, r, galatLoginKedaluwarsa, nil)
		return
	}
	// GETDEL: state yang sama tidak bisa dipakai dua kali, juga bila percobaan ini gagal.
	t, err := h.Transaksi.Ambil(r.Context(), state)
	if errors.Is(err, auth.ErrTransaksiTidakAda) {
		h.gagalLogin(w, r, galatLoginKedaluwarsa, nil)
		return
	}
	if err != nil {
		h.gagalLogin(w, r, galatLoginGagal, err)
		return
	}
	c, err := r.Cookie(h.namaCookieIkatan())
	if err != nil || subtle.ConstantTimeCompare([]byte(auth.Hash(c.Value)), []byte(t.HashIkatan)) != 1 {
		h.gagalLogin(w, r, galatLoginKedaluwarsa, nil)
		return
	}
	if q.Get("error") != "" {
		h.gagalLogin(w, r, galatLoginGagal, errors.New("penyedia identitas mengembalikan galat otorisasi"))
		return
	}
	kode := q.Get("code")
	if kode == "" {
		h.gagalLogin(w, r, galatLoginGagal, errors.New("callback tanpa kode otorisasi"))
		return
	}
	ident, tok, err := h.OIDC.Tukar(r.Context(), kode, t.Verifier, t.Nonce)
	if errors.Is(err, auth.ErrAksesDitolak) {
		h.gagalLogin(w, r, galatAksesDitolak, nil)
		return
	}
	if err != nil {
		h.gagalLogin(w, r, galatLoginGagal, err)
		return
	}

	// ID session selalu baru; session lama pada browser yang sama (bila ada) dibuang.
	if c, err := r.Cookie(h.CookieName); err == nil {
		if err := h.Sesi.Hapus(r.Context(), c.Value); err != nil {
			h.log.Warn("gagal menghapus session lama", "err", err)
		}
	}
	id, _, err := h.Sesi.Buat(r.Context(), ident, tok)
	if err != nil {
		h.gagalLogin(w, r, galatLoginGagal, err)
		return
	}
	http.SetCookie(w, h.cookie(id, int(h.CookieMaxAge.Seconds())))
	http.Redirect(w, r, halamanAdmin, http.StatusFound)
}

// batasiLogin menerapkan batas per IP. alihkan=true (callback): penolakan dialihkan ke halaman admin.
func (h authHandler) batasiLogin(w http.ResponseWriter, r *http.Request, bucket string, alihkan bool) bool {
	keputusan, err := h.Limiter.Hit(r.Context(), bucket, ipKlien(r), h.BatasLogin)
	switch {
	case err != nil && alihkan:
		h.gagalLogin(w, r, galatLoginGagal, err)
	case err != nil:
		h.galatSession(w, r, err)
	case keputusan.Diizinkan:
		return true
	case alihkan:
		h.gagalLogin(w, r, galatLoginGagal, nil)
	default:
		w.Header().Set("Retry-After", strconv.Itoa(keputusan.TungguDetik))
		tulisGalat(w, http.StatusTooManyRequests, kodeTerlaluSering, "Terlalu banyak percobaan login, coba lagi nanti")
	}
	return false
}

// gagalLogin mengalihkan ke halaman admin dengan kode galat. err (bila ada) dicatat tanpa isi query.
func (h authHandler) gagalLogin(w http.ResponseWriter, r *http.Request, kode string, err error) {
	if err != nil {
		h.log.Warn("login gagal", "kode", kode, "err", err, "request_id", middleware.GetReqID(r.Context()))
	}
	http.Redirect(w, r, halamanAdmin+"?galat="+url.QueryEscape(kode), http.StatusFound)
}

func (h authHandler) me(w http.ResponseWriter, r *http.Request) {
	s, _ := SessionDariContext(r.Context())
	tulisJSON(w, http.StatusOK, bentukSesi(s))
}

// logout menghapus session lokal lalu (upaya terbaik, batas waktu pendek) mengakhiri session SSO Keycloak.
func (h authHandler) logout(w http.ResponseWriter, r *http.Request) {
	s, _ := SessionDariContext(r.Context())
	c, err := r.Cookie(h.CookieName)
	if err == nil {
		if err := h.Sesi.Hapus(r.Context(), c.Value); err != nil {
			h.galatSession(w, r, err)
			return
		}
	}
	http.SetCookie(w, h.cookie("", -1))
	if s.Refresh != "" && h.OIDC != nil {
		ctx, batal := context.WithTimeout(context.WithoutCancel(r.Context()), batasAkhiriSSO)
		if err := h.OIDC.AkhiriSSO(ctx, s.Refresh); err != nil {
			h.log.Warn("gagal mengakhiri session SSO Keycloak", "err", err, "request_id", middleware.GetReqID(r.Context()))
		}
		batal()
	}
	w.WriteHeader(http.StatusNoContent)
}

// wajibLogin memvalidasi cookie session (menyegarkan token bila perlu, memperpanjang idle). Tanpa session sah: 401;
// Redis atau penyedia identitas tidak terjangkau: 503.
func (h authHandler) wajibLogin(next http.Handler) http.Handler {
	return h.wajibSession(next, true)
}

// wajibSessionTanpaSegar sama seperti wajibLogin tetapi tanpa penyegaran token (untuk logout).
func (h authHandler) wajibSessionTanpaSegar(next http.Handler) http.Handler {
	return h.wajibSession(next, false)
}

func (h authHandler) wajibSession(next http.Handler, segarkan bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(h.CookieName)
		if err != nil {
			tulisGalat(w, http.StatusUnauthorized, kodeBelumLogin, "Belum login atau session berakhir")
			return
		}
		var sess auth.Session
		if segarkan {
			sess, err = h.Sesi.Validasi(r.Context(), c.Value)
		} else {
			sess, err = h.Sesi.Baca(r.Context(), c.Value)
		}
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

// namaCookieIkatan mengikuti pola cookie session: awalan __Host- (wajib Secure, Path=/, tanpa Domain) di production.
func (h authHandler) namaCookieIkatan() string {
	if strings.HasPrefix(h.CookieName, "__Host-") {
		return "__Host-login"
	}
	return "login"
}

// cookieIkatan mengikat transaksi login ke browser yang memulainya. SameSite=Lax: ikut terkirim pada navigasi
// GET tingkat atas kembali dari Keycloak.
func (h authHandler) cookieIkatan(nilai string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     h.namaCookieIkatan(),
		Value:    nilai,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

func (h authHandler) galatSession(w http.ResponseWriter, r *http.Request, err error) {
	h.log.Error("layanan session bermasalah", "err", err, "path", r.URL.Path, "request_id", middleware.GetReqID(r.Context()))
	tulisGalat(w, http.StatusServiceUnavailable, kodeSessionBermasalah, "Layanan session sedang tidak tersedia")
}

// — middleware bersama (dipakai juga oleh route admin) ——————————————————————

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
