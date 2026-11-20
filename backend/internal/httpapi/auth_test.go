package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/auth"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/ratelimit"
)

const (
	originSah = "http://localhost:5173"

	// batas unggah kecil agar tes ukuran cepat
	batasUnggahUji = 1 << 20
)

// Identitas sintetis HANYA untuk tes: session dibuat langsung lewat manager yang disuntikkan (tanpa login).
var (
	identitasUji  = auth.Identitas{Sub: "dev-admin", Nama: "Admin Dev", Role: "pegawai", Username: "admin.dev"}
	identitasUji2 = auth.Identitas{Sub: "dev-admin-2", Nama: "Admin Dev 2", Role: "pegawai", Username: "admin.dev2"}
)

// storeSesi adalah auth.Store di memori dengan injeksi galat.
type storeSesi struct {
	data  map[string]auth.Session
	kunci map[string]string
	err   error
}

func baruStoreSesi() *storeSesi {
	return &storeSesi{data: map[string]auth.Session{}, kunci: map[string]string{}}
}

func (s *storeSesi) Simpan(_ context.Context, id string, v auth.Session, _ time.Duration) error {
	if s.err != nil {
		return s.err
	}
	s.data[id] = v
	return nil
}

func (s *storeSesi) GantiBilaAda(_ context.Context, id string, v auth.Session, _ time.Duration) error {
	if s.err != nil {
		return s.err
	}
	if _, ok := s.data[id]; !ok {
		return auth.ErrTidakAda
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

func (s *storeSesi) KunciSegar(_ context.Context, id, pemilik string, _ time.Duration) (bool, error) {
	if _, ada := s.kunci[id]; ada {
		return false, nil
	}
	s.kunci[id] = pemilik
	return true, nil
}

func (s *storeSesi) LepasKunciSegar(_ context.Context, id, pemilik string) error {
	if s.kunci[id] == pemilik {
		delete(s.kunci, id)
	}
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

// penyediaPalsu meniru Keycloak untuk tes handler (verifikasi token nyata diuji di oidc_test.go).
type penyediaPalsu struct {
	pilihAkun       bool
	state, nonce    string
	verifier        string
	kode            string
	ident           auth.Identitas
	tok             auth.Token
	errTukar        error
	errAkhiri       error
	akhiriDipanggil int
	refreshDiakhiri string
	tukarDipanggil  int
}

func (p *penyediaPalsu) URLMasuk(state, nonce, verifier string, pilihAkun bool) string {
	p.state, p.nonce, p.verifier, p.pilihAkun = state, nonce, verifier, pilihAkun
	return "https://idp.test/auth?state=" + url.QueryEscape(state)
}

func (p *penyediaPalsu) Tukar(_ context.Context, kode, verifier, nonce string) (auth.Identitas, auth.Token, error) {
	p.tukarDipanggil++
	p.kode, p.verifier, p.nonce = kode, verifier, nonce
	return p.ident, p.tok, p.errTukar
}

func (p *penyediaPalsu) AkhiriSSO(_ context.Context, refresh string) error {
	p.akhiriDipanggil++
	p.refreshDiakhiri = refresh
	return p.errAkhiri
}

// transaksiMem meniru auth.RedisTransaksi (sekali pakai).
type transaksiMem struct {
	data map[string]auth.TransaksiLogin
	err  error
}

func (m *transaksiMem) Simpan(_ context.Context, state string, t auth.TransaksiLogin, _ time.Duration) error {
	if m.err != nil {
		return m.err
	}
	m.data[state] = t
	return nil
}

func (m *transaksiMem) Ambil(_ context.Context, state string) (auth.TransaksiLogin, error) {
	if m.err != nil {
		return auth.TransaksiLogin{}, m.err
	}
	t, ok := m.data[state]
	if !ok {
		return auth.TransaksiLogin{}, auth.ErrTransaksiTidakAda
	}
	delete(m.data, state)
	return t, nil
}

type lingkunganAuth struct {
	h         http.Handler
	sesi      Sesi
	store     *storeSesi
	limiter   *limiterPalsu
	penyedia  *penyediaPalsu
	transaksi *transaksiMem
	media     *media.Memori
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
	pen := &penyediaPalsu{}
	tr := &transaksiMem{data: map[string]auth.TransaksiLogin{}}
	d := DepsAuth{
		Sesi:           auth.NewManager(st, time.Hour, 8*time.Hour),
		Limiter:        lim,
		OIDC:           pen,
		Transaksi:      tr,
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
	return lingkunganAuth{h: h, sesi: d.Sesi, store: st, limiter: lim, penyedia: pen, transaksi: tr}
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

type sesiUji struct{ id, csrf string }

// loginSebagai membuat session langsung lewat manager yang disuntikkan (bukan endpoint login) untuk identitas uji.
func loginSebagai(t *testing.T, e lingkunganAuth, ident auth.Identitas) sesiUji {
	t.Helper()
	id, s, err := e.sesi.Buat(context.Background(), ident, auth.Token{})
	if err != nil {
		t.Fatalf("membuat session uji: %v", err)
	}
	return sesiUji{id: id, csrf: s.CSRFToken}
}

// loginBerhasil membuat session untuk identitasUji (dev-admin/admin.dev).
func loginBerhasil(t *testing.T, e lingkunganAuth) (id string, csrf string) {
	t.Helper()
	s := loginSebagai(t, e, identitasUji)
	return s.id, s.csrf
}

func kodeGalat(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var g bentukGalat
	if err := json.Unmarshal(rec.Body.Bytes(), &g); err != nil {
		t.Fatalf("bukan bentuk galat: %s", rec.Body)
	}
	return g.Error.Code
}

func cookieBernama(rec *httptest.ResponseRecorder, nama string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == nama {
			return c
		}
	}
	return nil
}

// panggilCallback memanggil /api/auth/callback dengan query dan cookie pengikat (kosong = tanpa cookie).
func panggilCallback(h http.Handler, query url.Values, namaIkatan, ikatan, session string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback?"+query.Encode(), nil)
	if ikatan != "" {
		req.AddCookie(&http.Cookie{Name: namaIkatan, Value: ikatan})
	}
	if session != "" {
		req.AddCookie(&http.Cookie{Name: "session", Value: session})
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func harusDialihkanGalat(t *testing.T, rec *httptest.ResponseRecorder, kode string) {
	t.Helper()
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin/?galat="+kode {
		t.Fatalf("mau 302 ke /admin/?galat=%s, dapat %d %q", kode, rec.Code, rec.Header().Get("Location"))
	}
	if c := cookieBernama(rec, "session"); c != nil && c.MaxAge >= 0 {
		t.Errorf("login gagal tidak boleh memasang cookie session: %+v", c)
	}
}

// mulaiLogin memanggil /api/auth/login dan mengembalikan state (dari penyedia palsu) dan cookie pengikat.
func mulaiLogin(t *testing.T, e lingkunganAuth, query string) (state string, ikatan *http.Cookie, rec *httptest.ResponseRecorder) {
	t.Helper()
	rec = kirim(e.h, http.MethodGet, "/api/auth/login"+query, "", opsi{})
	if rec.Code != http.StatusFound {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}
	ikatan = cookieBernama(rec, "login")
	if ikatan == nil {
		t.Fatalf("cookie pengikat tidak dipasang: %v", rec.Result().Cookies())
	}
	return e.penyedia.state, ikatan, rec
}

func TestLoginMengalihkanKePenyediaDenganTransaksi(t *testing.T) {
	e := siapkanAuth(t, nil)
	state, ikatan, rec := mulaiLogin(t, e, "")
	if rec.Header().Get("Location") != "https://idp.test/auth?state="+url.QueryEscape(state) || rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("redirect: %q cc=%q", rec.Header().Get("Location"), rec.Header().Get("Cache-Control"))
	}
	if !ikatan.HttpOnly || ikatan.SameSite != http.SameSiteLaxMode || ikatan.Path != "/" || ikatan.MaxAge != 300 || ikatan.Secure || !auth.IDSah(ikatan.Value) {
		t.Errorf("atribut cookie pengikat salah: %+v", ikatan)
	}
	if e.penyedia.pilihAkun {
		t.Error("tanpa pilih=1 tidak boleh prompt=login")
	}
	tr, ada := e.transaksi.data[state]
	if !ada || tr.Nonce != e.penyedia.nonce || tr.Verifier != e.penyedia.verifier || tr.HashIkatan != auth.Hash(ikatan.Value) {
		t.Fatalf("transaksi salah: %+v", tr)
	}
	if !auth.IDSah(state) || !auth.IDSah(tr.Nonce) || len(tr.Verifier) < 43 || state == tr.Nonce {
		t.Errorf("state/nonce/verifier harus acak 256 bit: %q %q %q", state, tr.Nonce, tr.Verifier)
	}
	if cookieBernama(rec, "session") != nil {
		t.Error("login belum membuat session")
	}

	mulaiLogin(t, e, "?pilih=1")
	if !e.penyedia.pilihAkun {
		t.Error("pilih=1 harus meminta pemilihan akun (prompt=login)")
	}
}

func TestLoginCookiePengikatHostPrefixDiProduksi(t *testing.T) {
	e := siapkanAuth(t, func(d *DepsAuth) { d.CookieSecure = true; d.CookieName = "__Host-session" })
	rec := kirim(e.h, http.MethodGet, "/api/auth/login", "", opsi{})
	c := cookieBernama(rec, "__Host-login")
	if c == nil || !c.Secure || c.Domain != "" || c.Path != "/" || !c.HttpOnly {
		t.Errorf("cookie pengikat __Host- harus Secure, Path=/, tanpa Domain: %+v", rec.Result().Cookies())
	}
}

func TestLoginTransaksiGagalDisimpan503(t *testing.T) {
	e := siapkanAuth(t, nil)
	e.transaksi.err = errors.New("redis mati")
	rec := kirim(e.h, http.MethodGet, "/api/auth/login", "", opsi{})
	if rec.Code != http.StatusServiceUnavailable || kodeGalat(t, rec) != kodeSessionBermasalah || cookieBernama(rec, "login") != nil {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
}

func TestCallbackBerhasilMembuatSessionBaru(t *testing.T) {
	e := siapkanAuth(t, nil)
	lama, _ := loginBerhasil(t, e)
	state, ikatan, _ := mulaiLogin(t, e, "")
	e.penyedia.ident = auth.Identitas{Sub: "sub-1", Nama: "Budi", Role: "direktur", Username: "budi"}
	e.penyedia.tok = auth.Token{Refresh: "rt", AksesHabis: time.Now().Add(5 * time.Minute)}
	nonce, verifier := e.penyedia.nonce, e.penyedia.verifier

	rec := panggilCallback(e.h, url.Values{"state": {state}, "code": {"kode-1"}}, "login", ikatan.Value, lama)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin/" || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("callback: %d %q", rec.Code, rec.Header().Get("Location"))
	}
	if e.penyedia.kode != "kode-1" || e.penyedia.nonce != nonce || e.penyedia.verifier != verifier {
		t.Errorf("Tukar harus menerima kode, nonce, dan verifier transaksi: %+v", e.penyedia)
	}
	c := cookieBernama(rec, "session")
	if c == nil || !auth.IDSah(c.Value) || c.Value == lama || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.MaxAge != 8*3600 {
		t.Fatalf("cookie session salah: %+v", c)
	}
	if b := cookieBernama(rec, "login"); b == nil || b.MaxAge >= 0 {
		t.Errorf("cookie pengikat harus dibuang: %+v", b)
	}
	if _, ada := e.store.data[lama]; ada {
		t.Error("session lama harus dibuang")
	}
	s := e.store.data[c.Value]
	if s.Sub != "sub-1" || s.Refresh != "rt" || s.CSRFToken == "" {
		t.Errorf("session tersimpan salah: %+v", s)
	}
	m := peta(t, kirim(e.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: c.Value}))
	if m["sub"] != "sub-1" || m["nama"] != "Budi" || m["username"] != "budi" || m["role"] != "direktur" || m["csrfToken"] != s.CSRFToken {
		t.Errorf("me: %v", m)
	}
}

func TestCallbackStateDanPengikat(t *testing.T) {
	e := siapkanAuth(t, nil)
	e.penyedia.ident = identitasUji
	e.penyedia.tok = auth.Token{Refresh: "rt", AksesHabis: time.Now().Add(5 * time.Minute)}

	// Tanpa state, state tak dikenal, atau tanpa cookie pengikat: login_kedaluwarsa.
	harusDialihkanGalat(t, panggilCallback(e.h, url.Values{"code": {"k"}}, "login", "x", ""), galatLoginKedaluwarsa)
	harusDialihkanGalat(t, panggilCallback(e.h, url.Values{"state": {"asing"}, "code": {"k"}}, "login", "x", ""), galatLoginKedaluwarsa)
	state, _, _ := mulaiLogin(t, e, "")
	harusDialihkanGalat(t, panggilCallback(e.h, url.Values{"state": {state}, "code": {"k"}}, "login", "", ""), galatLoginKedaluwarsa)

	// Pengikat salah: ditolak dan state ikut hangus (sekali pakai).
	state, ikatan, _ := mulaiLogin(t, e, "")
	harusDialihkanGalat(t, panggilCallback(e.h, url.Values{"state": {state}, "code": {"k"}}, "login", "pengikat-lain", ""), galatLoginKedaluwarsa)
	harusDialihkanGalat(t, panggilCallback(e.h, url.Values{"state": {state}, "code": {"k"}}, "login", ikatan.Value, ""), galatLoginKedaluwarsa)

	// Replay: state yang sudah berhasil dipakai tidak bisa dipakai lagi.
	state, ikatan, _ = mulaiLogin(t, e, "")
	if rec := panggilCallback(e.h, url.Values{"state": {state}, "code": {"k"}}, "login", ikatan.Value, ""); rec.Header().Get("Location") != "/admin/" {
		t.Fatalf("callback pertama harus berhasil: %q", rec.Header().Get("Location"))
	}
	harusDialihkanGalat(t, panggilCallback(e.h, url.Values{"state": {state}, "code": {"k"}}, "login", ikatan.Value, ""), galatLoginKedaluwarsa)
	if e.penyedia.tukarDipanggil != 1 {
		t.Errorf("hanya callback yang sah boleh menukar kode, dipanggil %d kali", e.penyedia.tukarDipanggil)
	}
	if len(e.store.data) != 1 {
		t.Errorf("hanya satu session yang boleh dibuat: %d", len(e.store.data))
	}
}

func TestCallbackGalatPenyedia(t *testing.T) {
	kasus := []struct {
		nama  string
		query url.Values
		err   error
		kode  string
	}{
		{"error dari Keycloak", url.Values{"error": {"access_denied"}}, nil, galatLoginGagal},
		{"tanpa kode", url.Values{}, nil, galatLoginGagal},
		{"token tidak sah", url.Values{"code": {"k"}}, errors.New("oidc: signature tidak sah"), galatLoginGagal},
		{"role tidak ada", url.Values{"code": {"k"}}, auth.ErrAksesDitolak, galatAksesDitolak},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			e := siapkanAuth(t, nil)
			e.penyedia.errTukar = k.err
			state, ikatan, _ := mulaiLogin(t, e, "")
			k.query.Set("state", state)
			harusDialihkanGalat(t, panggilCallback(e.h, k.query, "login", ikatan.Value, ""), k.kode)
			if len(e.store.data) != 0 {
				t.Error("login gagal tidak boleh membuat session")
			}
		})
	}
}

func TestOriginWajibUntukMetodeUbahData(t *testing.T) {
	e := siapkanAuth(t, nil)
	id, csrf := loginBerhasil(t, e)
	for _, o := range []string{"", "http://evil.example.com", "http://localhost:5173/", "null", "HTTP://LOCALHOST:5173"} {
		rec := kirim(e.h, http.MethodPost, "/api/auth/logout", "", opsi{origin: o, cookie: id, csrf: csrf})
		if rec.Code != http.StatusForbidden || kodeGalat(t, rec) != kodeOriginTidakSah {
			t.Errorf("origin %q: status %d: %s", o, rec.Code, rec.Body)
		}
	}
	if _, ada := e.store.data[id]; !ada {
		t.Error("origin salah harus ditolak sebelum menghapus session")
	}
	// GET tidak butuh Origin.
	if rec := kirim(e.h, http.MethodGet, "/api/auth/me", "", opsi{}); rec.Code != http.StatusUnauthorized {
		t.Errorf("GET me tanpa Origin harus diproses (401 belum login), dapat %d", rec.Code)
	}
}

func TestRateLimitLogin(t *testing.T) {
	e := siapkanAuth(t, nil)
	e.limiter.keputusan = ratelimit.Keputusan{Diizinkan: false, TungguDetik: 42}
	// X-Forwarded-For dipalsukan harus diabaikan.
	rec := kirim(e.h, http.MethodGet, "/api/auth/login", "", opsi{xff: "9.9.9.9"})
	if rec.Code != http.StatusTooManyRequests || kodeGalat(t, rec) != kodeTerlaluSering {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Retry-After") != "42" {
		t.Errorf("Retry-After salah: %q", rec.Header().Get("Retry-After"))
	}
	if e.limiter.bucket != "login" || e.limiter.id != "192.0.2.1" || e.limiter.batas != 10 {
		t.Errorf("limiter dipanggil dengan %q/%q/%d; harus login/RemoteAddr(192.0.2.1)/10", e.limiter.bucket, e.limiter.id, e.limiter.batas)
	}
	if len(e.transaksi.data) != 0 || cookieBernama(rec, "login") != nil {
		t.Error("request yang dibatasi tidak boleh membuat transaksi")
	}

	// Callback dibatasi di bucket sendiri; penolakan dialihkan ke halaman admin.
	rec = panggilCallback(e.h, url.Values{"state": {"s"}, "code": {"k"}}, "login", "x", "")
	harusDialihkanGalat(t, rec, galatLoginGagal)
	if e.limiter.bucket != "login-callback" || e.penyedia.tukarDipanggil != 0 {
		t.Errorf("callback: bucket %q, tukar %d", e.limiter.bucket, e.penyedia.tukarDipanggil)
	}
}

func TestLimiterBermasalahMenjadi503(t *testing.T) {
	e := siapkanAuth(t, nil)
	e.limiter.err = errors.New("redis: connection refused password=rahasia")
	rec := kirim(e.h, http.MethodGet, "/api/auth/login", "", opsi{})
	if rec.Code != http.StatusServiceUnavailable || kodeGalat(t, rec) != kodeSessionBermasalah {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "rahasia") {
		t.Error("detail internal bocor")
	}
	harusDialihkanGalat(t, panggilCallback(e.h, url.Values{"state": {"s"}}, "login", "x", ""), galatLoginGagal)
}

// Login development sudah dihapus: tidak ada rute password.
func TestDevLoginTidakAda(t *testing.T) {
	e := siapkanAuth(t, nil)
	rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", `{"password":"apa saja"}`, opsi{origin: originSah})
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
	if m["sub"] != "dev-admin" || m["nama"] != "Admin Dev" || m["username"] != "admin.dev" || m["role"] != "pegawai" || m["csrfToken"] != csrf || len(m) != 5 {
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
	if e.penyedia.akhiriDipanggil != 0 {
		t.Error("session tanpa refresh token tidak perlu mengakhiri SSO")
	}
}

// Logout mengakhiri SSO Keycloak dengan refresh token session; kegagalan penyedia tidak menghalangi logout lokal.
func TestLogoutMengakhiriSSOUpayaTerbaik(t *testing.T) {
	for _, errPenyedia := range []error{nil, errors.New("dial tcp: connection refused")} {
		e := siapkanAuth(t, nil)
		id, s, err := e.sesi.Buat(context.Background(), identitasUji, auth.Token{Refresh: "rt-logout", AksesHabis: time.Now().Add(time.Minute)})
		if err != nil {
			t.Fatal(err)
		}
		e.penyedia.errAkhiri = errPenyedia
		rec := kirim(e.h, http.MethodPost, "/api/auth/logout", "", opsi{origin: originSah, cookie: id, csrf: s.CSRFToken})
		if rec.Code != http.StatusNoContent {
			t.Fatalf("penyedia %v: logout harus 204, dapat %d %s", errPenyedia, rec.Code, rec.Body)
		}
		if e.penyedia.akhiriDipanggil != 1 || e.penyedia.refreshDiakhiri != "rt-logout" {
			t.Errorf("AkhiriSSO: %d %q", e.penyedia.akhiriDipanggil, e.penyedia.refreshDiakhiri)
		}
		if _, ada := e.store.data[id]; ada {
			t.Error("session lokal harus terhapus")
		}
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
