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
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/auth"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/ratelimit"
)

const (
	originSah = "http://localhost:5173"
	passDev   = "rahasia-dev"

	// batas unggah kecil agar tes ukuran cepat
	batasUnggahUji = 1 << 20
)

// storeSesi adalah auth.Store di memori dengan injeksi galat.
type storeSesi struct {
	data map[string]auth.Session
	err  error
}

func baruStoreSesi() *storeSesi { return &storeSesi{data: map[string]auth.Session{}} }

func (s *storeSesi) Simpan(_ context.Context, id string, v auth.Session, _ time.Duration) error {
	if s.err != nil {
		return s.err
	}
	s.data[id] = v
	return nil
}

func (s *storeSesi) Ambil(_ context.Context, id string) (auth.Session, error) {
	if s.err != nil {
		return auth.Session{}, s.err
	}
	v, ok := s.data[id]
	if !ok {
		return auth.Session{}, auth.ErrTidakAda
	}
	return v, nil
}

func (s *storeSesi) Perpanjang(context.Context, string, time.Duration) error { return s.err }

func (s *storeSesi) Hapus(_ context.Context, id string) error {
	if s.err != nil {
		return s.err
	}
	delete(s.data, id)
	return nil
}

// limiterPalsu mencatat pemanggilan dan mengembalikan keputusan yang diatur tes.
type limiterPalsu struct {
	keputusan ratelimit.Keputusan
	err       error
	bucket    string
	id        string
	batas     int
	dipanggil int
}

func (l *limiterPalsu) Hit(_ context.Context, bucket, id string, batas int) (ratelimit.Keputusan, error) {
	l.dipanggil++
	l.bucket, l.id, l.batas = bucket, id, batas
	return l.keputusan, l.err
}

type lingkunganAuth struct {
	h       http.Handler
	store   *storeSesi
	limiter *limiterPalsu
	media   *media.Memori
}

func siapkanAuth(t *testing.T, ubah func(d *DepsAuth), lain ...func(*Deps)) lingkunganAuth {
	t.Helper()
	mem := media.NewMemori()
	e := siapkanAuthDenganMedia(t, ubah, mem, lain...)
	e.media = mem
	return e
}

// siapkanAuthDenganMedia sama seperti siapkanAuth tetapi dengan storage gambar pilihan (mis. RustFS nyata).
// Field media pada hasilnya nil; pemanggil memakai storage yang ia berikan.
func siapkanAuthDenganMedia(t *testing.T, ubah func(d *DepsAuth), m media.Store, lain ...func(*Deps)) lingkunganAuth {
	t.Helper()
	st := baruStoreSesi()
	lim := &limiterPalsu{keputusan: ratelimit.Keputusan{Diizinkan: true}}
	d := DepsAuth{
		Sesi:           auth.NewManager(st, time.Hour, 8*time.Hour),
		Limiter:        lim,
		DevLoginAktif:  true,
		AdminPassword:  passDev,
		BatasLogin:     10,
		CookieName:     "session",
		CookieSecure:   false,
		CookieMaxAge:   8 * time.Hour,
		AllowedOrigins: []string{originSah},
	}
	if ubah != nil {
		ubah(&d)
	}
	deps := Deps{Log: slog.New(slog.NewTextHandler(io.Discard, nil)), DB: pingerPalsu{}, Konten: kontenPalsu{}, Auth: d, Media: m, UploadMaxBytes: batasUnggahUji}
	for _, f := range lain {
		f(&deps)
	}
	h := NewRouter(deps)
	return lingkunganAuth{h: h, store: st, limiter: lim}
}

type opsi struct {
	origin string
	csrf   string
	cookie string
	xff    string
}

func kirim(h http.Handler, metode, path, body string, o opsi) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metode, path, strings.NewReader(body))
	if o.origin != "" {
		req.Header.Set("Origin", o.origin)
	}
	if o.csrf != "" {
		req.Header.Set("X-CSRF-Token", o.csrf)
	}
	if o.cookie != "" {
		req.AddCookie(&http.Cookie{Name: "session", Value: o.cookie})
	}
	if o.xff != "" {
		req.Header.Set("X-Forwarded-For", o.xff)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func loginBerhasil(t *testing.T, e lingkunganAuth) (id string, csrf string) {
	t.Helper()
	rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", `{"password":"`+passDev+`"}`, opsi{origin: originSah})
	if rec.Code != http.StatusOK {
		t.Fatalf("login status %d: %s", rec.Code, rec.Body)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "session" {
			id = c.Value
		}
	}
	var m sesiJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	return id, m.CSRFToken
}

func kodeGalat(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var g bentukGalat
	if err := json.Unmarshal(rec.Body.Bytes(), &g); err != nil {
		t.Fatalf("bukan bentuk galat: %s", rec.Body)
	}
	return g.Error.Code
}

func TestDevLoginBerhasil(t *testing.T) {
	e := siapkanAuth(t, nil)
	rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", `{"password":"`+passDev+`"}`, opsi{origin: originSah})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	var m map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	if m["nama"] != "Admin Dev" || m["username"] != "admin.dev" || m["role"] != "pegawai" || m["csrfToken"] == "" || m["csrfToken"] == nil {
		t.Fatalf("isi salah: %v", m)
	}
	if _, bocor := m["sub"]; bocor || len(m) != 4 {
		t.Errorf("hanya nama, username, role, csrfToken yang boleh dikirim: %v", m)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control harus no-store, dapat %q", rec.Header().Get("Cache-Control"))
	}

	cs := rec.Result().Cookies()
	if len(cs) != 1 {
		t.Fatalf("harus satu cookie: %v", cs)
	}
	c := cs[0]
	if c.Name != "session" || !auth.IDSah(c.Value) || !c.HttpOnly || c.Path != "/" || c.SameSite != http.SameSiteLaxMode ||
		c.MaxAge != 8*3600 || c.Secure {
		t.Errorf("atribut cookie salah: %+v", c)
	}
	if got := e.store.data[c.Value]; got.Sub != "dev-admin" || got.CSRFToken != m["csrfToken"] {
		t.Errorf("session di store salah: %+v", got)
	}
}

func TestDevLoginCookieSecureBilaDikonfigurasi(t *testing.T) {
	e := siapkanAuth(t, func(d *DepsAuth) { d.CookieSecure = true; d.CookieName = "__Host-session" })
	rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", `{"password":"`+passDev+`"}`, opsi{origin: originSah})
	c := rec.Result().Cookies()[0]
	if c.Name != "__Host-session" || !c.Secure || c.Domain != "" || c.Path != "/" {
		t.Errorf("cookie __Host- harus Secure, Path=/, tanpa Domain: %+v", c)
	}
}

func TestDevLoginPasswordSalah(t *testing.T) {
	e := siapkanAuth(t, nil)
	for _, body := range []string{`{"password":"salah"}`, `{"password":""}`, `{}`} {
		rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", body, opsi{origin: originSah})
		if rec.Code != http.StatusUnauthorized || kodeGalat(t, rec) != kodePasswordSalah {
			t.Errorf("%s: status %d: %s", body, rec.Code, rec.Body)
		}
		if len(rec.Result().Cookies()) != 0 {
			t.Errorf("%s: login gagal tidak boleh mengirim cookie", body)
		}
	}
	if len(e.store.data) != 0 {
		t.Error("login gagal tidak boleh membuat session")
	}
}

func TestDevLoginBodyTidakValid(t *testing.T) {
	e := siapkanAuth(t, nil)
	besar := `{"password":"` + strings.Repeat("a", 2000) + `"}`
	for _, body := range []string{"", "bukan json", `["x"]`, besar} {
		rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", body, opsi{origin: originSah})
		if rec.Code != http.StatusBadRequest || kodeGalat(t, rec) != kodeRequestTidakSah {
			t.Errorf("body %.20q: status %d: %s", body, rec.Code, rec.Body)
		}
	}
}

func TestOriginWajibUntukMetodeUbahData(t *testing.T) {
	e := siapkanAuth(t, nil)
	benar := `{"password":"` + passDev + `"}`
	for _, o := range []string{"", "http://evil.example.com", "http://localhost:5173/", "null", "HTTP://LOCALHOST:5173"} {
		rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", benar, opsi{origin: o})
		if rec.Code != http.StatusForbidden || kodeGalat(t, rec) != kodeOriginTidakSah {
			t.Errorf("origin %q: status %d: %s", o, rec.Code, rec.Body)
		}
	}
	if e.limiter.dipanggil != 0 || len(e.store.data) != 0 {
		t.Error("origin salah harus ditolak sebelum limiter dan sebelum membuat session")
	}
	// GET tidak butuh Origin.
	if rec := kirim(e.h, http.MethodGet, "/api/auth/me", "", opsi{}); rec.Code != http.StatusUnauthorized {
		t.Errorf("GET me tanpa Origin harus diproses (401 belum login), dapat %d", rec.Code)
	}
}

func TestRateLimitLogin(t *testing.T) {
	e := siapkanAuth(t, nil)
	e.limiter.keputusan = ratelimit.Keputusan{Diizinkan: false, TungguDetik: 42}
	// Password BENAR pun ditolak saat dibatasi; X-Forwarded-For dipalsukan harus diabaikan.
	rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", `{"password":"`+passDev+`"}`, opsi{origin: originSah, xff: "9.9.9.9"})
	if rec.Code != http.StatusTooManyRequests || kodeGalat(t, rec) != kodeTerlaluSering {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Retry-After") != "42" {
		t.Errorf("Retry-After salah: %q", rec.Header().Get("Retry-After"))
	}
	if e.limiter.bucket != "login" || e.limiter.id != "192.0.2.1" || e.limiter.batas != 10 {
		t.Errorf("limiter dipanggil dengan %q/%q/%d; harus login/RemoteAddr(192.0.2.1)/10", e.limiter.bucket, e.limiter.id, e.limiter.batas)
	}
	if len(e.store.data) != 0 {
		t.Error("request yang dibatasi tidak boleh membuat session")
	}
}

func TestLimiterBermasalahMenjadi503(t *testing.T) {
	e := siapkanAuth(t, nil)
	e.limiter.err = errors.New("redis: connection refused password=rahasia")
	rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", `{"password":"`+passDev+`"}`, opsi{origin: originSah})
	if rec.Code != http.StatusServiceUnavailable || kodeGalat(t, rec) != kodeSessionBermasalah {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "rahasia") {
		t.Error("detail internal bocor")
	}
}

func TestDevLoginTidakAdaBilaBukanDevelopment(t *testing.T) {
	e := siapkanAuth(t, func(d *DepsAuth) { d.DevLoginAktif = false })
	rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", `{"password":"`+passDev+`"}`, opsi{origin: originSah})
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("route harus tidak ada (404/405), dapat %d", rec.Code)
	}
	if len(e.store.data) != 0 || e.limiter.dipanggil != 0 {
		t.Error("route yang tidak ada tidak boleh memproses apa pun")
	}
	if rec.Code == http.StatusNotFound && kodeGalat(t, rec) != kodeTidakDitemukan {
		t.Errorf("404 harus berformat standar: %s", rec.Body)
	}
}

func TestMe(t *testing.T) {
	e := siapkanAuth(t, nil)
	id, csrf := loginBerhasil(t, e)

	rec := kirim(e.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: id})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	var m map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	if m["nama"] != "Admin Dev" || m["username"] != "admin.dev" || m["role"] != "pegawai" || m["csrfToken"] != csrf || len(m) != 4 {
		t.Errorf("isi me salah: %v", m)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Error("me harus no-store")
	}
}

func TestMeTanpaSessionSah(t *testing.T) {
	e := siapkanAuth(t, nil)
	rec := kirim(e.h, http.MethodGet, "/api/auth/me", "", opsi{})
	if rec.Code != http.StatusUnauthorized || kodeGalat(t, rec) != kodeBelumLogin {
		t.Fatalf("tanpa cookie: %d %s", rec.Code, rec.Body)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Error("tanpa cookie, tidak ada yang perlu dibersihkan")
	}

	rec = kirim(e.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: strings.Repeat("a", 43)})
	if rec.Code != http.StatusUnauthorized || kodeGalat(t, rec) != kodeBelumLogin {
		t.Fatalf("cookie tak dikenal: %d %s", rec.Code, rec.Body)
	}
	cs := rec.Result().Cookies()
	if len(cs) != 1 || cs[0].MaxAge >= 0 || cs[0].Value != "" {
		t.Errorf("cookie basi harus dibersihkan: %+v", cs)
	}
}

func TestMeStoreBermasalah503(t *testing.T) {
	e := siapkanAuth(t, nil)
	id, _ := loginBerhasil(t, e)
	e.store.err = errors.New("redis mati")
	rec := kirim(e.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: id})
	if rec.Code != http.StatusServiceUnavailable || kodeGalat(t, rec) != kodeSessionBermasalah {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
}

func TestLogout(t *testing.T) {
	e := siapkanAuth(t, nil)
	id, csrf := loginBerhasil(t, e)

	// Tanpa token CSRF, token salah, atau token milik session lain: 403.
	id2, csrf2 := loginBerhasil(t, e)
	for nama, o := range map[string]opsi{
		"tanpa token":     {origin: originSah, cookie: id},
		"token salah":     {origin: originSah, cookie: id, csrf: "salah"},
		"token session b": {origin: originSah, cookie: id, csrf: csrf2},
	} {
		rec := kirim(e.h, http.MethodPost, "/api/auth/logout", "", o)
		if rec.Code != http.StatusForbidden || kodeGalat(t, rec) != kodeCSRFTidakValid {
			t.Errorf("%s: status %d: %s", nama, rec.Code, rec.Body)
		}
	}
	// Origin salah: 403 origin_invalid walau token benar.
	rec := kirim(e.h, http.MethodPost, "/api/auth/logout", "", opsi{origin: "http://evil.example.com", cookie: id, csrf: csrf})
	if rec.Code != http.StatusForbidden || kodeGalat(t, rec) != kodeOriginTidakSah {
		t.Errorf("origin salah: %d %s", rec.Code, rec.Body)
	}
	// Tanpa session: 401.
	rec = kirim(e.h, http.MethodPost, "/api/auth/logout", "", opsi{origin: originSah, csrf: csrf})
	if rec.Code != http.StatusUnauthorized || kodeGalat(t, rec) != kodeBelumLogin {
		t.Errorf("tanpa session: %d %s", rec.Code, rec.Body)
	}
	if _, ada := e.store.data[id]; !ada {
		t.Fatal("percobaan yang ditolak tidak boleh menghapus session")
	}

	// Berhasil: 204, session hilang, cookie dibersihkan; session lain tetap ada.
	rec = kirim(e.h, http.MethodPost, "/api/auth/logout", "", opsi{origin: originSah, cookie: id, csrf: csrf})
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	cs := rec.Result().Cookies()
	if len(cs) != 1 || cs[0].MaxAge >= 0 || cs[0].Value != "" || !cs[0].HttpOnly {
		t.Errorf("cookie harus dibersihkan: %+v", cs)
	}
	if _, ada := e.store.data[id]; ada {
		t.Error("session harus dihapus dari store")
	}
	if _, ada := e.store.data[id2]; !ada {
		t.Error("session lain tidak boleh ikut terhapus")
	}
	if rec := kirim(e.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: id}); rec.Code != http.StatusUnauthorized {
		t.Errorf("setelah logout me harus 401, dapat %d", rec.Code)
	}
}

func TestLoginUlangMembuangSessionLamaDanMembuatIDBaru(t *testing.T) {
	e := siapkanAuth(t, nil)
	lama, csrfLama := loginBerhasil(t, e)

	rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", `{"password":"`+passDev+`"}`, opsi{origin: originSah, cookie: lama})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	baru := rec.Result().Cookies()[0].Value
	var m sesiJSON
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	if baru == lama || m.CSRFToken == csrfLama {
		t.Error("ID session dan token CSRF harus baru pada setiap login")
	}
	if _, ada := e.store.data[lama]; ada {
		t.Error("session lama harus dibuang")
	}
}

func TestRouteAuthTidakMengubahRutePublik(t *testing.T) {
	e := siapkanAuth(t, nil)
	if rec := kirim(e.h, http.MethodGet, "/api/health", "", opsi{}); rec.Code != http.StatusOK {
		t.Errorf("health %d", rec.Code)
	}
	if rec := kirim(e.h, http.MethodGet, "/api/berita", "", opsi{}); rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") == "no-store" {
		t.Errorf("rute publik tidak boleh ikut no-store: %d %q", rec.Code, rec.Header().Get("Cache-Control"))
	}
}
