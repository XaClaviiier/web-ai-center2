package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/cache"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/lock"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

const passDev2 = "rahasia-dev-dua"

type sesiUji struct{ id, csrf string }

type lingkLock struct {
	h   http.Handler
	rdb *redis.Client
	svc *layananPalsu
	tim *layananTimPalsu
	a   sesiUji // dev-admin ("Admin Dev")
	b   sesiUji // dev-admin-2 ("Admin Dev 2")
}

func loginDengan(t *testing.T, h http.Handler, pass string) sesiUji {
	t.Helper()
	rec := kirim(h, http.MethodPost, "/api/auth/dev-login", `{"password":"`+pass+`"}`, opsi{origin: originSah})
	if rec.Code != http.StatusOK {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}
	var s sesiUji
	for _, c := range rec.Result().Cookies() {
		if c.Name == "session" {
			s.id = c.Value
		}
	}
	var body struct {
		CSRFToken string `json:"csrfToken"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	s.csrf = body.CSRFToken
	return s
}

func siapkanLock(t *testing.T) lingkLock {
	t.Helper()
	rdb := testutil.RedisTes(t)
	svc := &layananPalsu{hasil: contohBerita()}
	tim := &layananTimPalsu{hasil: admin.Tim{ID: 3}}
	e := siapkanAuth(t, func(d *DepsAuth) { d.AdminPassword2 = passDev2 }, func(d *Deps) {
		d.Admin = DepsAdmin{Berita: svc, Tim: tim}
		d.Lock = &lock.Lock{C: rdb, TTL: time.Minute}
		d.LockHeartbeatDetik = 20
		d.Cache = &cache.Cache{C: rdb}
	})
	return lingkLock{h: e.h, rdb: rdb, svc: svc, tim: tim, a: loginDengan(t, e.h, passDev), b: loginDengan(t, e.h, passDev2)}
}

func (e lingkLock) kirim(t *testing.T, s sesiUji, metode, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	return kirimAdmin(t, e.h, s.id, s.csrf, metode, path, body)
}

func peta(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("bukan JSON: %s", rec.Body)
	}
	return m
}

const bodyUbahUji = `{"judul":"x","isi":"y","tanggalPublikasi":"09-11-2026","gambar":"berita/3f9c2a10-7b4e-4c6d-9a1f-0123456789ab.png","version":3}`

func TestAkunDevKeduaPunyaIdentitasSendiri(t *testing.T) {
	e := siapkanLock(t)
	for _, c := range []struct {
		s    sesiUji
		nama string
	}{{e.a, "Admin Dev"}, {e.b, "Admin Dev 2"}} {
		rec := e.kirim(t, c.s, "GET", "/api/auth/me", "")
		if rec.Code != 200 || peta(t, rec)["nama"] != c.nama {
			t.Errorf("me: %d %s (mau %s)", rec.Code, rec.Body, c.nama)
		}
	}
	// Password acak tetap ditolak, juga saat akun kedua tidak dikonfigurasi.
	if rec := kirim(e.h, http.MethodPost, "/api/auth/dev-login", `{"password":"salah"}`, opsi{origin: originSah}); rec.Code != http.StatusUnauthorized {
		t.Errorf("password salah harus 401, dapat %d", rec.Code)
	}
	tanpa := siapkanAuth(t, nil)
	if rec := kirim(tanpa.h, http.MethodPost, "/api/auth/dev-login", `{"password":""}`, opsi{origin: originSah}); rec.Code != http.StatusUnauthorized {
		t.Errorf("password kosong tidak boleh cocok dengan akun kedua yang kosong: %d", rec.Code)
	}
}

func TestLockPasangDuaAdminSalingMengunci(t *testing.T) {
	e := siapkanLock(t)
	rec := e.kirim(t, e.a, "POST", "/api/admin/berita/7/lock", "")
	m := peta(t, rec)
	if rec.Code != 200 || m["milikSaya"] != true || m["heartbeatDetik"] != float64(20) || m["pemegang"].(map[string]any)["nama"] != "Admin Dev" {
		t.Fatalf("A pasang: %d %s", rec.Code, rec.Body)
	}
	rec = e.kirim(t, e.b, "POST", "/api/admin/berita/7/lock", "")
	if rec.Code != statusTerkunci || kodeGalat(t, rec) != "locked" {
		t.Fatalf("B harus 423 locked: %d %s", rec.Code, rec.Body)
	}
	g := peta(t, rec)["error"].(map[string]any)
	if g["pemegang"].(map[string]any)["nama"] != "Admin Dev" {
		t.Errorf("423 harus menyebut nama pemegang: %s", rec.Body)
	}
	if rec := e.kirim(t, e.a, "POST", "/api/admin/berita/7/lock", ""); rec.Code != 200 {
		t.Errorf("A pasang ulang harus idempoten: %d", rec.Code)
	}
}

func TestLockDetailMenampilkanInfo(t *testing.T) {
	e := siapkanLock(t)
	m := peta(t, e.kirim(t, e.a, "GET", "/api/admin/berita/7", ""))
	v, ada := m["lock"]
	if !ada || v != nil {
		t.Fatalf("tanpa lock, field lock harus ada dan null: %v ada=%v", v, ada)
	}
	if m["slug"] != "judul-tujuh" {
		t.Errorf("isi detail asli harus tetap ada: %v", m["slug"])
	}
	e.kirim(t, e.a, "POST", "/api/admin/berita/7/lock", "")
	la := peta(t, e.kirim(t, e.a, "GET", "/api/admin/berita/7", ""))["lock"].(map[string]any)
	lb := peta(t, e.kirim(t, e.b, "GET", "/api/admin/berita/7", ""))["lock"].(map[string]any)
	if la["milikSaya"] != true || lb["milikSaya"] != false || lb["pemegang"].(map[string]any)["nama"] != "Admin Dev" || lb["kedaluwarsaPada"] == nil {
		t.Errorf("info lock salah: A=%v B=%v", la, lb)
	}
	// Daftar: item yang dikunci ORANG LAIN ditandai dikunciOleh; kunci milik sendiri tidak.
	e.svc.daftar, e.svc.total = []admin.Berita{contohBerita()}, 1
	daftar := func(s sesiUji) map[string]any {
		rec := e.kirim(t, s, "GET", "/api/admin/berita", "")
		if rec.Code != 200 || !json.Valid(rec.Body.Bytes()) {
			t.Fatalf("daftar: %d %s", rec.Code, rec.Body)
		}
		return peta(t, rec)["items"].([]any)[0].(map[string]any)
	}
	if v, ada := daftar(e.b)["dikunciOleh"]; !ada || v != "Admin Dev" {
		t.Errorf("B melihat item dikunci A: dikunciOleh=%v ada=%v", v, ada)
	}
	if v, ada := daftar(e.a)["dikunciOleh"]; ada {
		t.Errorf("pemegang kunci sendiri tidak ditandai: %v", v)
	}
	if m := peta(t, e.kirim(t, e.b, "GET", "/api/admin/berita", "")); m["total"] == nil || m["page"] == nil {
		t.Errorf("meta daftar harus utuh: %v", m)
	}
	e.kirim(t, e.a, "DELETE", "/api/admin/berita/7/lock", "")
	if _, ada := daftar(e.b)["dikunciOleh"]; ada {
		t.Error("setelah kunci dilepas, tanda hilang")
	}
}

func TestLockMenolakPerubahanOlehOrangLain(t *testing.T) {
	e := siapkanLock(t)
	e.kirim(t, e.a, "POST", "/api/admin/berita/7/lock", "")

	semua := []struct{ metode, path, body string }{
		{"PUT", "/api/admin/berita/7", bodyUbahUji},
		{"POST", "/api/admin/berita/7/post", `{"version":3}`},
		{"POST", "/api/admin/berita/7/retract", `{"version":3}`},
		{"DELETE", "/api/admin/berita/7", ""},
		{"POST", "/api/admin/berita/7/restore", ""},
		{"DELETE", "/api/admin/berita/7/permanen", ""},
	}
	for _, c := range semua {
		e.svc.aksi = ""
		rec := e.kirim(t, e.b, c.metode, c.path, c.body)
		if rec.Code != statusTerkunci || kodeGalat(t, rec) != "locked" {
			t.Errorf("%s %s oleh B: %d %s", c.metode, c.path, rec.Code, rec.Body)
		}
		if e.svc.aksi != "" {
			t.Errorf("%s %s: layanan tidak boleh dipanggil saat terkunci (aksi %q)", c.metode, c.path, e.svc.aksi)
		}
	}
	// Item lain tidak terpengaruh.
	if rec := e.kirim(t, e.b, "PUT", "/api/admin/berita/8", bodyUbahUji); rec.Code != 200 {
		t.Errorf("item lain harus bebas: %d %s", rec.Code, rec.Body)
	}
}

func TestLockDilepasSetelahSimpanDanPemegangBebas(t *testing.T) {
	e := siapkanLock(t)
	ada := func() bool { return e.rdb.Exists(context.Background(), "lock:berita:7").Val() == 1 }

	e.kirim(t, e.a, "POST", "/api/admin/berita/7/lock", "")
	if rec := e.kirim(t, e.a, "PUT", "/api/admin/berita/7", bodyUbahUji); rec.Code != 200 || e.svc.aksi != "ubah" {
		t.Fatalf("pemegang harus bisa menyimpan: %d %s", rec.Code, rec.Body)
	}
	if ada() {
		t.Error("lock harus dilepas setelah Simpan sukses")
	}
	// Tanpa lock (mis. kedaluwarsa) Simpan tetap jalan; version jadi pengaman.
	if rec := e.kirim(t, e.b, "PUT", "/api/admin/berita/7", bodyUbahUji); rec.Code != 200 {
		t.Errorf("tanpa lock harus lolos: %d", rec.Code)
	}
	// Simpan yang gagal (409) tidak melepas lock.
	e.kirim(t, e.a, "POST", "/api/admin/berita/7/lock", "")
	e.svc.err = admin.ErrVersiBentrok
	if rec := e.kirim(t, e.a, "PUT", "/api/admin/berita/7", bodyUbahUji); rec.Code != http.StatusConflict {
		t.Fatalf("mau 409, dapat %d", rec.Code)
	}
	if !ada() {
		t.Error("simpan gagal tidak boleh melepas lock")
	}
	e.svc.err = nil
	// Post oleh pemegang sukses lalu menghapus lock.
	if rec := e.kirim(t, e.a, "POST", "/api/admin/berita/7/post", `{"version":3}`); rec.Code != 200 {
		t.Fatalf("post: %d %s", rec.Code, rec.Body)
	}
	if ada() {
		t.Error("lock harus dihapus setelah Post sukses")
	}
}

func TestLockHeartbeatDanLepasEndpoint(t *testing.T) {
	e := siapkanLock(t)
	ada := func() bool { return e.rdb.Exists(context.Background(), "lock:berita:7").Val() == 1 }

	if rec := e.kirim(t, e.a, "PUT", "/api/admin/berita/7/lock", ""); rec.Code != statusLockHilang || kodeGalat(t, rec) != "lock_lost" {
		t.Errorf("heartbeat tanpa lock: %d %s", rec.Code, rec.Body)
	}
	e.kirim(t, e.a, "POST", "/api/admin/berita/7/lock", "")
	e.rdb.PExpire(context.Background(), "lock:berita:7", 5*time.Second)
	rec := e.kirim(t, e.a, "PUT", "/api/admin/berita/7/lock", "")
	if rec.Code != 200 || peta(t, rec)["milikSaya"] != true {
		t.Fatalf("heartbeat pemegang: %d %s", rec.Code, rec.Body)
	}
	if ttl := e.rdb.PTTL(context.Background(), "lock:berita:7").Val(); ttl < 50*time.Second {
		t.Errorf("TTL harus diperpanjang: %v", ttl)
	}
	if rec := e.kirim(t, e.b, "PUT", "/api/admin/berita/7/lock", ""); rec.Code != statusLockHilang || kodeGalat(t, rec) != "lock_lost" {
		t.Errorf("heartbeat oleh bukan pemegang: %d %s", rec.Code, rec.Body)
	}
	if rec := e.kirim(t, e.b, "DELETE", "/api/admin/berita/7/lock", ""); rec.Code != 204 || !ada() {
		t.Errorf("B tidak boleh melepas lock A (status %d, ada=%v)", rec.Code, ada())
	}
	if rec := e.kirim(t, e.a, "DELETE", "/api/admin/berita/7/lock", ""); rec.Code != 204 || ada() {
		t.Errorf("A harus bisa melepas (status %d, ada=%v)", rec.Code, ada())
	}
	if rec := e.kirim(t, e.a, "DELETE", "/api/admin/berita/7/lock", ""); rec.Code != 204 {
		t.Errorf("lepas ulang harus idempoten: %d", rec.Code)
	}
}

func TestLockHanyaUntukDraftYangAda(t *testing.T) {
	e := siapkanLock(t)
	e.svc.hasil.Status = "posted"
	if rec := e.kirim(t, e.a, "POST", "/api/admin/berita/7/lock", ""); rec.Code != http.StatusConflict || kodeGalat(t, rec) != "status_tidak_valid" {
		t.Errorf("posted: %d %s", rec.Code, rec.Body)
	}
	e.svc.hasil.Status = "draft"
	hapus := time.Now()
	e.svc.hasil.DihapusPada = &hapus
	if rec := e.kirim(t, e.a, "POST", "/api/admin/berita/7/lock", ""); rec.Code != http.StatusConflict {
		t.Errorf("trash: %d", rec.Code)
	}
	e.svc.hasil.DihapusPada = nil
	e.svc.err = admin.ErrTidakDitemukan
	if rec := e.kirim(t, e.a, "POST", "/api/admin/berita/7/lock", ""); rec.Code != http.StatusNotFound {
		t.Errorf("tidak ada: %d", rec.Code)
	}
	e.svc.err = nil
	if rec := e.kirim(t, e.a, "POST", "/api/admin/berita/abc/lock", ""); rec.Code != http.StatusNotFound {
		t.Errorf("id tidak sah: %d", rec.Code)
	}
	if e.rdb.Exists(context.Background(), "lock:berita:7").Val() != 0 {
		t.Error("tidak boleh ada kunci yang terbentuk dari percobaan gagal")
	}
}

func TestLockTimTanpaAturanDraftDanTidakMembatalkanCache(t *testing.T) {
	e := siapkanLock(t)
	if rec := e.kirim(t, e.a, "POST", "/api/admin/tim/3/lock", ""); rec.Code != 200 {
		t.Fatalf("pasang tim: %d %s", rec.Code, rec.Body)
	}
	e.kirim(t, e.a, "PUT", "/api/admin/tim/3/lock", "")
	if rec := e.kirim(t, e.b, "POST", "/api/admin/tim/3/lock", ""); rec.Code != statusTerkunci {
		t.Errorf("B harus 423: %d", rec.Code)
	}
	e.tim.aksi = ""
	if rec := e.kirim(t, e.b, "DELETE", "/api/admin/tim/3", ""); rec.Code != statusTerkunci || e.tim.aksi != "" {
		t.Errorf("hapus tim oleh B: %d (aksi %q)", rec.Code, e.tim.aksi)
	}
	e.kirim(t, e.a, "DELETE", "/api/admin/tim/3/lock", "")
	if e.rdb.Exists(context.Background(), "cache:ver:tim").Val() != 0 {
		t.Error("memasang, heartbeat, dan melepas lock tidak boleh membatalkan cache tim")
	}
}

func TestLockGagalTertutupBilaRedisMati(t *testing.T) {
	mati := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 200 * time.Millisecond, MaxRetries: -1})
	t.Cleanup(func() { _ = mati.Close() })
	svc := &layananPalsu{hasil: contohBerita(), daftar: []admin.Berita{contohBerita()}, total: 1} // daftar berisi item, jadi kunci harus dibaca
	e := siapkanAuth(t, nil, func(d *Deps) {
		d.Admin = DepsAdmin{Berita: svc}
		d.Lock = &lock.Lock{C: mati}
	})
	id, csrf := loginBerhasil(t, e)
	for _, c := range []struct{ metode, path, body string }{
		{"POST", "/api/admin/berita/7/lock", ""},
		{"PUT", "/api/admin/berita/7/lock", ""},
		{"PUT", "/api/admin/berita/7", bodyUbahUji},
		{"GET", "/api/admin/berita/7", ""},
		{"GET", "/api/admin/berita", ""},
	} {
		svc.aksi = ""
		rec := kirimAdmin(t, e.h, id, csrf, c.metode, c.path, c.body)
		if rec.Code != http.StatusServiceUnavailable || kodeGalat(t, rec) != "session_unavailable" {
			t.Errorf("%s %s: %d %s", c.metode, c.path, rec.Code, rec.Body)
		}
		if c.path == "/api/admin/berita/7" && c.metode == "PUT" && svc.aksi != "" {
			t.Error("PUT tidak boleh menyentuh layanan bila lock tidak bisa diperiksa")
		}
	}
}
