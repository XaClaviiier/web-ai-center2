package httpapi

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/auth"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

// Tes ujung ke ujung login Keycloak dengan penyedia OIDC tiruan (httptest) yang menandatangani token RS256 nyata.

const (
	klienUji    = "ai-center-cms"
	rahasiaUji  = "rahasia-klien-uji"
	redirectUji = "http://localhost:8080/api/auth/callback"
	subUji      = "sub-budi"
	kidUji      = "kunci-uji"
)

var (
	kunciRSAUji     = sync.OnceValue(func() *rsa.PrivateKey { return buatKunciRSA() })
	kunciRSALainUji = sync.OnceValue(func() *rsa.PrivateKey { return buatKunciRSA() })
)

func buatKunciRSA() *rsa.PrivateKey {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return k
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

type kodeIdP struct{ nonce, tantangan, redirect string }

// idpUji meniru endpoint Keycloak yang dipakai backend: discovery, JWKS, token (authorization_code + PKCE,
// refresh_token dengan rotasi), dan logout back-channel.
type idpUji struct {
	srv    *httptest.Server
	issuer string

	mu           sync.Mutex
	kode         map[string]kodeIdP
	refresh      map[string]string // refresh token aktif -> sub
	diterbitkan  []string          // semua refresh token yang pernah diterbitkan
	nama         string
	username     string
	roles        []string
	masaAkses    time.Duration
	ubahAkses    func(map[string]any)
	ubahID       func(map[string]any)
	kunciTanda   *rsa.PrivateKey // nil = kunci yang diumumkan di JWKS
	tolakSegar   bool
	panggilSegar int
	logout       []url.Values
	statusLogout int
}

func baruIdP(t *testing.T) *idpUji {
	t.Helper()
	p := &idpUji{kode: map[string]kodeIdP{}, refresh: map[string]string{}, nama: "Budi Santoso", username: "budi",
		roles: []string{"pegawai", "offline_access"}, masaAkses: 5 * time.Minute, statusLogout: http.StatusNoContent}
	mux := http.NewServeMux()
	p.srv = httptest.NewServer(mux)
	t.Cleanup(p.srv.Close)
	p.issuer = p.srv.URL + "/realms/ai-center"
	dasar := "/realms/ai-center/protocol/openid-connect"
	mux.HandleFunc("GET /realms/ai-center/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		tulisJSONIdP(w, http.StatusOK, map[string]any{
			"issuer":                                p.issuer,
			"authorization_endpoint":                p.srv.URL + dasar + "/auth",
			"token_endpoint":                        p.srv.URL + dasar + "/token",
			"jwks_uri":                              p.srv.URL + dasar + "/certs",
			"end_session_endpoint":                  p.srv.URL + dasar + "/logout",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("GET "+dasar+"/certs", func(w http.ResponseWriter, _ *http.Request) {
		pub := kunciRSAUji().PublicKey
		tulisJSONIdP(w, http.StatusOK, map[string]any{"keys": []map[string]any{{
			"kty": "RSA", "kid": kidUji, "use": "sig", "alg": "RS256",
			"n": b64(pub.N.Bytes()), "e": b64(big.NewInt(int64(pub.E)).Bytes()),
		}}})
	})
	mux.HandleFunc("POST "+dasar+"/token", p.token)
	mux.HandleFunc("POST "+dasar+"/logout", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		p.mu.Lock()
		defer p.mu.Unlock()
		p.logout = append(p.logout, r.PostForm)
		w.WriteHeader(p.statusLogout)
	})
	return p
}

func tulisJSONIdP(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// berikanKode meniru langkah authorize: pengguna sudah login di Keycloak, kode diterbitkan untuk nonce dan tantangan PKCE.
func (p *idpUji) berikanKode(nonce, tantangan, redirect string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	kode := "kode-" + b64(acakUji(16))
	p.kode[kode] = kodeIdP{nonce: nonce, tantangan: tantangan, redirect: redirect}
	return kode
}

func acakUji(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func (p *idpUji) token(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	f := r.PostForm
	if f.Get("client_id") != klienUji || f.Get("client_secret") != rahasiaUji {
		tulisJSONIdP(w, http.StatusUnauthorized, map[string]string{"error": "invalid_client"})
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	switch f.Get("grant_type") {
	case "authorization_code":
		k, ok := p.kode[f.Get("code")]
		delete(p.kode, f.Get("code"))
		sum := sha256.Sum256([]byte(f.Get("code_verifier")))
		if !ok || k.redirect != f.Get("redirect_uri") || b64(sum[:]) != k.tantangan {
			tulisJSONIdP(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
			return
		}
		p.terbitkan(w, subUji, k.nonce)
	case "refresh_token":
		p.panggilSegar++
		sub, ok := p.refresh[f.Get("refresh_token")]
		if p.tolakSegar || !ok {
			tulisJSONIdP(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant", "error_description": "Session not active"})
			return
		}
		delete(p.refresh, f.Get("refresh_token"))
		p.terbitkan(w, sub, "")
	default:
		tulisJSONIdP(w, http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
	}
}

// terbitkan menulis respons token seperti Keycloak: access token (Bearer) dan ID token (ID) bertanda tangan RS256.
func (p *idpUji) terbitkan(w http.ResponseWriter, sub, nonce string) {
	now := time.Now()
	akses := map[string]any{
		"iss": p.issuer, "aud": []string{klienUji, "account"}, "azp": klienUji, "sub": sub, "typ": "Bearer",
		"exp": now.Add(p.masaAkses).Unix(), "iat": now.Unix(), "preferred_username": p.username,
		"resource_access": map[string]any{klienUji: map[string]any{"roles": p.roles}, "account": map[string]any{"roles": []string{"direktur"}}},
	}
	id := map[string]any{
		"iss": p.issuer, "aud": klienUji, "azp": klienUji, "sub": sub, "typ": "ID",
		"exp": now.Add(5 * time.Minute).Unix(), "iat": now.Unix(), "name": p.nama, "preferred_username": p.username,
	}
	if nonce != "" {
		id["nonce"] = nonce
	}
	if p.ubahAkses != nil {
		p.ubahAkses(akses)
	}
	if p.ubahID != nil {
		p.ubahID(id)
	}
	rt := "rt-" + b64(acakUji(24))
	p.refresh[rt] = sub
	p.diterbitkan = append(p.diterbitkan, rt)
	tulisJSONIdP(w, http.StatusOK, map[string]any{
		"access_token": p.tandatangani(akses), "token_type": "Bearer", "expires_in": int(p.masaAkses.Seconds()),
		"refresh_token": rt, "id_token": p.tandatangani(id),
	})
}

func (p *idpUji) tandatangani(klaim map[string]any) string {
	kunci := kunciRSAUji()
	if p.kunciTanda != nil {
		kunci = p.kunciTanda
	}
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": kidUji, "typ": "JWT"})
	isi, _ := json.Marshal(klaim)
	masukan := b64(header) + "." + b64(isi)
	h := sha256.Sum256([]byte(masukan))
	sig, err := rsa.SignPKCS1v15(rand.Reader, kunci, crypto.SHA256, h[:])
	if err != nil {
		panic(err)
	}
	return masukan + "." + b64(sig)
}

func (p *idpUji) atur(f func(p *idpUji)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	f(p)
}

func (p *idpUji) jumlahSegar() (n int) {
	p.atur(func(p *idpUji) { n = p.panggilSegar })
	return n
}

func (p *idpUji) daftarLogout() (v []url.Values) {
	p.atur(func(p *idpUji) { v = slices.Clone(p.logout) })
	return v
}

func (p *idpUji) refreshDiterbitkan() (v []string) {
	p.atur(func(p *idpUji) { v = slices.Clone(p.diterbitkan) })
	return v
}

func (p *idpUji) kodeBelumDipakai(kode string) (ada bool) {
	p.atur(func(p *idpUji) { _, ada = p.kode[kode] })
	return ada
}

type lingkOIDC struct {
	lingkunganAuth
	idp *idpUji
	rdb *redis.Client // nil = store session dan transaksi di memori
}

// siapkanOIDC memasang *auth.OIDC nyata (discovery ke idpUji) sebagai penyedia login dan penyegar session.
func siapkanOIDC(t *testing.T, redisNyata bool) lingkOIDC {
	t.Helper()
	idp := baruIdP(t)
	o, err := auth.BukaOIDC(context.Background(), auth.KonfigOIDC{Issuer: idp.issuer, ClientID: klienUji, ClientSecret: rahasiaUji, RedirectURL: redirectUji})
	if err != nil {
		t.Fatalf("discovery: %v", err)
	}
	var rdb *redis.Client
	if redisNyata {
		rdb = testutil.RedisTes(t)
	}
	e := siapkanAuth(t, func(d *DepsAuth) {
		d.OIDC = o
		if rdb != nil {
			sandi, err := auth.BuatSandi(bytes.Repeat([]byte{7}, auth.PanjangKunciSandi))
			if err != nil {
				t.Fatal(err)
			}
			d.Sesi = auth.NewManager(auth.RedisStore{C: rdb, Sandi: sandi}, time.Hour, 8*time.Hour)
			d.Transaksi = auth.RedisTransaksi{C: rdb}
		}
		d.Sesi.(*auth.Manager).Penyegar = o
	})
	return lingkOIDC{lingkunganAuth: e, idp: idp, rdb: rdb}
}

// mulai memanggil /api/auth/login dan mengembalikan URL authorize serta nilai cookie pengikat.
func (l lingkOIDC) mulai(t *testing.T, query string) (*url.URL, string) {
	t.Helper()
	rec := kirim(l.h, http.MethodGet, "/api/auth/login"+query, "", opsi{})
	if rec.Code != http.StatusFound {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}
	loc, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	c := cookieBernama(rec, "login")
	if c == nil {
		t.Fatal("cookie pengikat tidak dipasang")
	}
	return loc, c.Value
}

// loginPenuh menjalankan login → authorize (kode dari idpUji) → callback.
func (l lingkOIDC) loginPenuh(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	loc, ikatan := l.mulai(t, "")
	q := loc.Query()
	kode := l.idp.berikanKode(q.Get("nonce"), q.Get("code_challenge"), q.Get("redirect_uri"))
	return panggilCallback(l.h, url.Values{"state": {q.Get("state")}, "code": {kode}}, "login", ikatan, "")
}

// sessionDariCallback memastikan callback berhasil dan mengembalikan ID session dari cookie.
func sessionDariCallback(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin/" {
		t.Fatalf("callback harus ke /admin/, dapat %d %q", rec.Code, rec.Header().Get("Location"))
	}
	c := cookieBernama(rec, "session")
	if c == nil || !auth.IDSah(c.Value) {
		t.Fatalf("cookie session tidak dipasang: %v", rec.Result().Cookies())
	}
	return c.Value
}

func TestOIDCLoginBerhasil(t *testing.T) {
	l := siapkanOIDC(t, false)
	loc, ikatan := l.mulai(t, "")
	q := loc.Query()
	if loc.Scheme+"://"+loc.Host != l.idp.srv.URL || !strings.HasSuffix(loc.Path, "/protocol/openid-connect/auth") {
		t.Errorf("harus dialihkan ke authorize Keycloak: %s", loc)
	}
	for k, v := range map[string]string{"response_type": "code", "client_id": klienUji, "scope": "openid", "code_challenge_method": "S256", "redirect_uri": redirectUji} {
		if q.Get(k) != v {
			t.Errorf("%s = %q, mau %q", k, q.Get(k), v)
		}
	}
	if q.Get("state") == "" || q.Get("nonce") == "" || len(q.Get("code_challenge")) != 43 || q.Has("prompt") {
		t.Errorf("state/nonce/challenge/prompt salah: %v", q)
	}

	kode := l.idp.berikanKode(q.Get("nonce"), q.Get("code_challenge"), q.Get("redirect_uri"))
	rec := panggilCallback(l.h, url.Values{"state": {q.Get("state")}, "code": {kode}}, "login", ikatan, "")
	id := sessionDariCallback(t, rec)
	if b := cookieBernama(rec, "login"); b == nil || b.MaxAge >= 0 {
		t.Error("cookie pengikat harus dibuang")
	}
	m := peta(t, kirim(l.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: id}))
	s := l.store.data[id]
	if m["sub"] != subUji || m["nama"] != "Budi Santoso" || m["username"] != "budi" || m["role"] != "pegawai" || m["csrfToken"] != s.CSRFToken || len(m) != 5 {
		t.Errorf("me: %v", m)
	}
	// Token tidak pernah dikirim ke browser; session server menyimpan refresh token dan masa berlaku access token.
	if strings.Contains(rec.Body.String(), "rt-") || s.Refresh == "" || s.AksesHabis.Before(time.Now().Add(4*time.Minute)) {
		t.Errorf("token session salah: %+v", s)
	}
}

func TestOIDCRoleDirekturDanFallbackNama(t *testing.T) {
	l := siapkanOIDC(t, false)
	l.idp.atur(func(p *idpUji) {
		p.roles = []string{"pegawai", "direktur"}
		p.ubahID = func(m map[string]any) { delete(m, "name") }
	})
	id := sessionDariCallback(t, l.loginPenuh(t))
	m := peta(t, kirim(l.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: id}))
	if m["role"] != "direktur" || m["nama"] != "budi" || m["username"] != "budi" {
		t.Errorf("me: %v", m)
	}
}

func TestOIDCPilihAkunMemintaPromptLogin(t *testing.T) {
	l := siapkanOIDC(t, false)
	loc, _ := l.mulai(t, "?pilih=1")
	if loc.Query().Get("prompt") != "login" {
		t.Errorf("pilih=1 harus prompt=login: %s", loc)
	}
}

func TestOIDCStateReplayDanPengikatSalah(t *testing.T) {
	l := siapkanOIDC(t, false)
	loc, ikatan := l.mulai(t, "")
	q := loc.Query()
	kode := l.idp.berikanKode(q.Get("nonce"), q.Get("code_challenge"), q.Get("redirect_uri"))
	callback := url.Values{"state": {q.Get("state")}, "code": {kode}}
	sessionDariCallback(t, panggilCallback(l.h, callback, "login", ikatan, ""))
	harusDialihkanGalat(t, panggilCallback(l.h, callback, "login", ikatan, ""), galatLoginKedaluwarsa)

	loc, _ = l.mulai(t, "")
	q = loc.Query()
	kode = l.idp.berikanKode(q.Get("nonce"), q.Get("code_challenge"), q.Get("redirect_uri"))
	harusDialihkanGalat(t, panggilCallback(l.h, url.Values{"state": {q.Get("state")}, "code": {kode}}, "login", "pengikat-penyerang", ""), galatLoginKedaluwarsa)
	if !l.idp.kodeBelumDipakai(kode) {
		t.Error("pengikat salah harus ditolak sebelum kode ditukar ke Keycloak")
	}
	if len(l.store.data) != 1 {
		t.Errorf("hanya login pertama yang membuat session: %d", len(l.store.data))
	}
}

func TestOIDCTokenTidakSahDitolak(t *testing.T) {
	lalu := time.Now().Add(-time.Minute).Unix()
	kasus := []struct {
		nama      string
		atur      func(p *idpUji)
		pkceSalah bool
	}{
		{nama: "nonce salah", atur: func(p *idpUji) { p.ubahID = func(m map[string]any) { m["nonce"] = "nonce-lain" } }},
		{nama: "tanpa nonce", atur: func(p *idpUji) { p.ubahID = func(m map[string]any) { delete(m, "nonce") } }},
		{nama: "audience ID token salah", atur: func(p *idpUji) { p.ubahID = func(m map[string]any) { m["aud"] = "klien-lain" } }},
		{nama: "audience access token salah", atur: func(p *idpUji) { p.ubahAkses = func(m map[string]any) { m["aud"] = []string{"account"} } }},
		{nama: "azp ID token salah", atur: func(p *idpUji) { p.ubahID = func(m map[string]any) { m["azp"] = "klien-lain" } }},
		{nama: "azp access token salah", atur: func(p *idpUji) { p.ubahAkses = func(m map[string]any) { m["azp"] = "klien-lain" } }},
		{nama: "access token tanpa azp", atur: func(p *idpUji) { p.ubahAkses = func(m map[string]any) { delete(m, "azp") } }},
		{nama: "issuer ID token salah", atur: func(p *idpUji) { p.ubahID = func(m map[string]any) { m["iss"] = "http://evil.example.com/realms/ai-center" } }},
		{nama: "issuer access token salah", atur: func(p *idpUji) { p.ubahAkses = func(m map[string]any) { m["iss"] = "http://evil.example.com/realms/ai-center" } }},
		{nama: "tanda tangan kunci lain", atur: func(p *idpUji) { p.kunciTanda = kunciRSALainUji() }},
		{nama: "sub ID dan access berbeda", atur: func(p *idpUji) { p.ubahID = func(m map[string]any) { m["sub"] = "sub-lain" } }},
		{nama: "tanpa sub", atur: func(p *idpUji) {
			p.ubahAkses = func(m map[string]any) { delete(m, "sub") }
			p.ubahID = func(m map[string]any) { delete(m, "sub") }
		}},
		{nama: "access token kedaluwarsa", atur: func(p *idpUji) { p.ubahAkses = func(m map[string]any) { m["exp"] = lalu } }},
		{nama: "ID token kedaluwarsa", atur: func(p *idpUji) { p.ubahID = func(m map[string]any) { m["exp"] = lalu } }},
		{nama: "ID token dipakai sebagai access token", atur: func(p *idpUji) { p.ubahAkses = func(m map[string]any) { m["typ"] = "ID" } }},
		{nama: "PKCE verifier tidak cocok", pkceSalah: true},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			l := siapkanOIDC(t, false)
			if k.atur != nil {
				l.idp.atur(k.atur)
			}
			loc, ikatan := l.mulai(t, "")
			q := loc.Query()
			tantangan := q.Get("code_challenge")
			if k.pkceSalah {
				sum := sha256.Sum256([]byte("verifier-penyerang"))
				tantangan = b64(sum[:])
			}
			kode := l.idp.berikanKode(q.Get("nonce"), tantangan, q.Get("redirect_uri"))
			harusDialihkanGalat(t, panggilCallback(l.h, url.Values{"state": {q.Get("state")}, "code": {kode}}, "login", ikatan, ""), galatLoginGagal)
			if len(l.store.data) != 0 {
				t.Error("token tidak sah tidak boleh membuat session")
			}
		})
	}
}

func TestOIDCTanpaRoleAksesDitolak(t *testing.T) {
	for nama, atur := range map[string]func(p *idpUji){
		"role lain":             func(p *idpUji) { p.roles = []string{"offline_access", "Direktur"} },
		"tanpa resource_access": func(p *idpUji) { p.ubahAkses = func(m map[string]any) { delete(m, "resource_access") } },
	} {
		t.Run(nama, func(t *testing.T) {
			l := siapkanOIDC(t, false)
			l.idp.atur(atur)
			harusDialihkanGalat(t, l.loginPenuh(t), galatAksesDitolak)
			if len(l.store.data) != 0 {
				t.Error("pengguna tanpa role tidak boleh mendapat session")
			}
		})
	}
}

// Access token habis: session disegarkan di server, identitas dan masa berlaku diperbarui, refresh token dirotasi.
func TestOIDCPenyegaranMemperbaruiSession(t *testing.T) {
	l := siapkanOIDC(t, false)
	l.idp.atur(func(p *idpUji) { p.masaAkses = 10 * time.Second }) // di dalam kelonggaran 30 detik: langsung perlu disegarkan
	id := sessionDariCallback(t, l.loginPenuh(t))
	rtAwal := l.store.data[id].Refresh
	l.idp.atur(func(p *idpUji) {
		p.masaAkses = 5 * time.Minute
		p.nama, p.username, p.roles = "Budi Baru", "budi.baru", []string{"direktur"}
	})

	m := peta(t, kirim(l.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: id}))
	if m["nama"] != "Budi Baru" || m["username"] != "budi.baru" || m["role"] != "direktur" || m["sub"] != subUji {
		t.Errorf("identitas harus diperbarui dari token baru: %v", m)
	}
	s := l.store.data[id]
	if n := l.idp.jumlahSegar(); n != 1 || s.Refresh == rtAwal || s.Refresh == "" || s.AksesHabis.Before(time.Now().Add(4*time.Minute)) {
		t.Errorf("penyegaran: panggil=%d refresh berganti=%v habis=%v", n, s.Refresh != rtAwal, s.AksesHabis)
	}
	if rec := kirim(l.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: id}); rec.Code != http.StatusOK || l.idp.jumlahSegar() != 1 {
		t.Errorf("token masih berlaku: tidak boleh disegarkan lagi (%d, panggil=%d)", rec.Code, l.idp.jumlahSegar())
	}
}

func TestOIDCPenyegaranDitolakMengakhiriSession(t *testing.T) {
	for nama, atur := range map[string]func(p *idpUji){
		"invalid_grant":      func(p *idpUji) { p.tolakSegar = true },
		"role dicabut":       func(p *idpUji) { p.roles = []string{"offline_access"} },
		"sub berganti":       func(p *idpUji) { p.ubahAkses = func(m map[string]any) { m["sub"] = "sub-lain" } },
		"tanda tangan salah": func(p *idpUji) { p.kunciTanda = kunciRSALainUji() },
	} {
		t.Run(nama, func(t *testing.T) {
			l := siapkanOIDC(t, false)
			l.idp.atur(func(p *idpUji) { p.masaAkses = 10 * time.Second })
			id := sessionDariCallback(t, l.loginPenuh(t))
			l.idp.atur(func(p *idpUji) { p.masaAkses = 5 * time.Minute })
			l.idp.atur(atur)
			rec := kirim(l.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: id})
			if rec.Code != http.StatusUnauthorized || kodeGalat(t, rec) != kodeBelumLogin {
				t.Fatalf("mau 401, dapat %d %s", rec.Code, rec.Body)
			}
			if c := cookieBernama(rec, "session"); c == nil || c.MaxAge >= 0 {
				t.Error("cookie session harus dibersihkan")
			}
			if _, ada := l.store.data[id]; ada {
				t.Error("session harus dihapus")
			}
		})
	}
}

// Keycloak tidak terjangkau saat penyegaran diperlukan: 503, session tidak diperpanjang dan tidak dihapus.
func TestOIDCPenyegaranPenyediaMati503(t *testing.T) {
	l := siapkanOIDC(t, false)
	l.idp.atur(func(p *idpUji) { p.masaAkses = 10 * time.Second })
	id := sessionDariCallback(t, l.loginPenuh(t))
	l.idp.srv.Close()
	rec := kirim(l.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: id})
	if rec.Code != http.StatusServiceUnavailable || kodeGalat(t, rec) != kodeSessionBermasalah {
		t.Fatalf("mau 503, dapat %d %s", rec.Code, rec.Body)
	}
	if _, ada := l.store.data[id]; !ada {
		t.Error("galat jaringan tidak boleh menghapus session")
	}
}

// Dengan Redis nyata: transaksi login di Redis, refresh token tersimpan terenkripsi.
func TestOIDCRefreshTokenTerenkripsiDiRedis(t *testing.T) {
	l := siapkanOIDC(t, true)
	rec := l.loginPenuh(t)
	id := sessionDariCallback(t, rec)
	ctx := context.Background()
	mentah := l.rdb.Get(ctx, "session:"+id).Val()
	if mentah == "" || !strings.Contains(mentah, `"rt":`) {
		t.Fatalf("session tidak tersimpan: %q", mentah)
	}
	for _, rt := range l.idp.refreshDiterbitkan() {
		if strings.Contains(mentah, rt) {
			t.Fatal("refresh token tersimpan sebagai teks biasa di Redis")
		}
	}
	if keys := l.rdb.Keys(ctx, "oidc:login:*").Val(); len(keys) != 0 {
		t.Errorf("transaksi login harus terhapus setelah dipakai: %v", keys)
	}
	if rec := kirim(l.h, http.MethodGet, "/api/auth/me", "", opsi{cookie: id}); rec.Code != http.StatusOK {
		t.Errorf("me: %d %s", rec.Code, rec.Body)
	}
}

func TestOIDCLogoutMengakhiriSSOUpayaTerbaik(t *testing.T) {
	l := siapkanOIDC(t, false)
	logout := func(id string) *httptest.ResponseRecorder {
		t.Helper()
		csrf := l.store.data[id].CSRFToken
		return kirim(l.h, http.MethodPost, "/api/auth/logout", "", opsi{origin: originSah, cookie: id, csrf: csrf})
	}

	id := sessionDariCallback(t, l.loginPenuh(t))
	rt := l.store.data[id].Refresh
	if rec := logout(id); rec.Code != http.StatusNoContent {
		t.Fatalf("logout: %d %s", rec.Code, rec.Body)
	}
	daftar := l.idp.daftarLogout()
	if len(daftar) != 1 {
		t.Fatalf("Keycloak harus dipanggil sekali: %d", len(daftar))
	}
	f := daftar[0]
	if f.Get("client_id") != klienUji || f.Get("client_secret") != rahasiaUji || f.Get("refresh_token") != rt {
		t.Errorf("form logout salah: client_id=%q refresh cocok=%v", f.Get("client_id"), f.Get("refresh_token") == rt)
	}

	// Keycloak menolak, lalu tidak terjangkau: logout lokal tetap 204.
	id = sessionDariCallback(t, l.loginPenuh(t))
	l.idp.atur(func(p *idpUji) { p.statusLogout = http.StatusInternalServerError })
	if rec := logout(id); rec.Code != http.StatusNoContent {
		t.Errorf("Keycloak 500: logout harus tetap 204, dapat %d", rec.Code)
	}
	id = sessionDariCallback(t, l.loginPenuh(t))
	l.idp.srv.Close()
	if rec := logout(id); rec.Code != http.StatusNoContent {
		t.Errorf("Keycloak mati: logout harus tetap 204, dapat %d", rec.Code)
	}
	if len(l.store.data) != 0 {
		t.Error("semua session lokal harus terhapus")
	}
	if !slices.ContainsFunc(l.idp.daftarLogout(), func(v url.Values) bool { return v.Get("refresh_token") != "" }) {
		t.Error("logout ke Keycloak harus membawa refresh token")
	}
}
