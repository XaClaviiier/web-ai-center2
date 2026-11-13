package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
)

type layananDokPalsu struct {
	err    error
	hasil  admin.Dokumentasi
	daftar []admin.RingkasDokumentasi
	total  int64

	aksi    string
	param   admin.ParamDaftar
	id      int64
	version int32
	oleh    string
	mentah  admin.MentahDokumentasi
}

func (l *layananDokPalsu) Daftar(_ context.Context, p admin.ParamDaftar) ([]admin.RingkasDokumentasi, int64, error) {
	l.aksi, l.param = "daftar", p
	return l.daftar, l.total, l.err
}
func (l *layananDokPalsu) Detail(_ context.Context, id int64) (admin.Dokumentasi, error) {
	l.aksi, l.id = "detail", id
	return l.hasil, l.err
}
func (l *layananDokPalsu) Buat(_ context.Context, m admin.MentahDokumentasi, oleh string) (admin.Dokumentasi, error) {
	l.aksi, l.mentah, l.oleh = "buat", m, oleh
	return l.hasil, l.err
}
func (l *layananDokPalsu) Ubah(_ context.Context, id int64, m admin.MentahDokumentasi, v int32, oleh string) (admin.Dokumentasi, error) {
	l.aksi, l.id, l.mentah, l.version, l.oleh = "ubah", id, m, v, oleh
	return l.hasil, l.err
}
func (l *layananDokPalsu) Post(_ context.Context, id int64, v int32, oleh string) (admin.Dokumentasi, error) {
	l.aksi, l.id, l.version, l.oleh = "post", id, v, oleh
	return l.hasil, l.err
}
func (l *layananDokPalsu) Retract(_ context.Context, id int64, v int32, oleh string) (admin.Dokumentasi, error) {
	l.aksi, l.id, l.version, l.oleh = "retract", id, v, oleh
	return l.hasil, l.err
}
func (l *layananDokPalsu) Hapus(_ context.Context, id int64, oleh string) error {
	l.aksi, l.id, l.oleh = "hapus", id, oleh
	return l.err
}
func (l *layananDokPalsu) Pulihkan(_ context.Context, id int64, oleh string) (admin.Dokumentasi, error) {
	l.aksi, l.id, l.oleh = "pulihkan", id, oleh
	return l.hasil, l.err
}
func (l *layananDokPalsu) HapusPermanen(_ context.Context, id int64) error {
	l.aksi, l.id = "permanen", id
	return l.err
}

const (
	keyDok1 = "dokumentasi/3f9c2a10-7b4e-4c6d-9a1f-0123456789a1.png"
	keyDok2 = "dokumentasi/3f9c2a10-7b4e-4c6d-9a1f-0123456789a2.png"
)

func contohDokumentasi() admin.Dokumentasi {
	return admin.Dokumentasi{
		ID: 12, Slug: "kegiatan-dua", Judul: "Kegiatan Dua", TanggalPublikasi: "2026-11-09",
		Gambar: []admin.GambarDokumentasi{{Key: keyDok1}, {Key: keyDok2}},
		Status: "draft", Version: 4, DibuatPada: waktuUji, DibuatOleh: "dev-admin", DiubahOleh: "editor-2",
	}
}

type lingkunganDokHTTP struct {
	lingkunganAuth
	svc *layananDokPalsu
}

func siapkanDokHTTP(t *testing.T) lingkunganDokHTTP {
	t.Helper()
	svc := &layananDokPalsu{hasil: contohDokumentasi()}
	e := siapkanAuth(t, nil, func(d *Deps) { d.Admin = DepsAdmin{Dokumentasi: svc} })
	return lingkunganDokHTTP{lingkunganAuth: e, svc: svc}
}

const badanDokOK = `"judul":"Kegiatan Dua","tanggalPublikasi":"09-11-2026","gambarDaftar":[{"key":"` + keyDok1 + `"},{"key":"` + keyDok2 + `"}]`

func TestAdminDokumentasiButuhLoginCSRFDanOrigin(t *testing.T) {
	e := siapkanDokHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	rute := []struct{ metode, path, body string }{
		{"GET", "/api/admin/dokumentasi", ""},
		{"GET", "/api/admin/dokumentasi/12", ""},
		{"POST", "/api/admin/dokumentasi", `{}`},
		{"PUT", "/api/admin/dokumentasi/12", `{}`},
		{"DELETE", "/api/admin/dokumentasi/12", ""},
		{"POST", "/api/admin/dokumentasi/12/post", `{"version":1}`},
		{"POST", "/api/admin/dokumentasi/12/retract", `{"version":1}`},
		{"POST", "/api/admin/dokumentasi/12/restore", ""},
		{"DELETE", "/api/admin/dokumentasi/12/permanen", ""},
	}
	for _, r := range rute {
		ubah := r.metode != "GET"
		e.svc.aksi = ""
		if rec := kirimAdmin(t, e.h, "", "", r.metode, r.path, r.body); rec.Code != http.StatusUnauthorized && !(ubah && rec.Code == http.StatusForbidden) {
			t.Errorf("%s %s tanpa login: %d", r.metode, r.path, rec.Code)
		}
		if e.svc.aksi != "" {
			t.Errorf("%s %s tanpa login: layanan tidak boleh dipanggil", r.metode, r.path)
		}
		e.svc.aksi = ""
		rec := kirimAdmin(t, e.h, id, "", r.metode, r.path, r.body)
		if ubah && (rec.Code != http.StatusForbidden || kodeGalat(t, rec) != kodeCSRFTidakValid || e.svc.aksi != "") {
			t.Errorf("%s %s tanpa CSRF: %d %s", r.metode, r.path, rec.Code, rec.Body)
		}
		if !ubah && rec.Code == http.StatusForbidden {
			t.Errorf("%s %s: GET tidak boleh butuh CSRF", r.metode, r.path)
		}
		e.svc.aksi = ""
		req := httptest.NewRequest(r.metode, r.path, strings.NewReader(r.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://evil.example.com")
		req.Header.Set("X-CSRF-Token", csrf)
		req.AddCookie(&http.Cookie{Name: "session", Value: id})
		rec = httptest.NewRecorder()
		e.h.ServeHTTP(rec, req)
		if ubah && (rec.Code != http.StatusForbidden || kodeGalat(t, rec) != kodeOriginTidakSah || e.svc.aksi != "") {
			t.Errorf("%s %s Origin asing: %d", r.metode, r.path, rec.Code)
		}
	}
}

func TestAdminDokumentasiDaftar(t *testing.T) {
	e := siapkanDokHTTP(t)
	id, _ := loginBerhasil(t, e.lingkunganAuth)
	hapus := waktuUji.Add(time.Hour)
	e.svc.daftar = []admin.RingkasDokumentasi{
		{ID: 12, Slug: "a", Judul: "A", TanggalPublikasi: "2026-11-09", GambarUtamaKey: keyDok1, JumlahGambar: 3, Status: "draft", Version: 4,
			DibuatPada: waktuUji, DibuatOleh: "dev-admin", DiubahOleh: "editor-2"},
		{ID: 13, Slug: "b", Judul: "B", TanggalPublikasi: "2026-11-08", JumlahGambar: 0, Status: "draft", Version: 1,
			DibuatPada: waktuUji, DibuatOleh: "dev-admin", DiubahOleh: "dev-admin", DihapusPada: &hapus},
	}
	e.svc.total = 42

	rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/dokumentasi", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if p := e.svc.param; p.Filter != "semua" || p.Batas != limitDaftarAdminDefault || p.Geser != 0 || p.Pola != "" {
		t.Errorf("default salah: %+v", p)
	}
	var m map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	if m["total"].(float64) != 42 || m["page"].(float64) != 1 || m["limit"].(float64) != limitDaftarAdminDefault {
		t.Errorf("meta: %v", m)
	}
	items := m["items"].([]any)
	i0, i1 := items[0].(map[string]any), items[1].(map[string]any)
	if _, ada := i0["gambar"]; ada {
		t.Errorf("daftar tidak memuat seluruh gambar: %v", i0)
	}
	gu := i0["gambarUtama"].(map[string]any)
	if gu["key"] != keyDok1 || gu["url"] != "/media/"+keyDok1 || i0["jumlahGambar"].(float64) != 3 || i0["status"] != "draft" || i0["version"].(float64) != 4 {
		t.Errorf("item 0: %v", i0)
	}
	if _, ada := i0["dihapusPada"]; ada {
		t.Errorf("dihapusPada hanya untuk trash: %v", i0)
	}
	if _, ada := i1["gambarUtama"]; ada || i1["dihapusPada"] == nil || i1["jumlahGambar"].(float64) != 0 {
		t.Errorf("item 1 (tanpa gambar, di trash): %v", i1)
	}
}

func TestAdminDokumentasiDaftarParameter(t *testing.T) {
	e := siapkanDokHTTP(t)
	id, _ := loginBerhasil(t, e.lingkunganAuth)
	rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/dokumentasi?status=trash&page=3&limit=500&q=ab%25", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if p := e.svc.param; p.Filter != "trash" || p.Batas != 50 || p.Geser != 100 || p.Pola != `%ab\%%` {
		t.Errorf("param ke layanan salah: %+v", p)
	}
	for _, q := range []string{"status=hapus", "status=", "page=0", "limit=0", "q=a", "page=x"} {
		if rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/dokumentasi?"+q, ""); rec.Code != http.StatusBadRequest || kodeGalat(t, rec) != kodeQueryTidakValid {
			t.Errorf("?%s: %d %s", q, rec.Code, rec.Body)
		}
	}
}

func TestAdminDokumentasiDetailMemuatGambarBerurutan(t *testing.T) {
	e := siapkanDokHTTP(t)
	id, _ := loginBerhasil(t, e.lingkunganAuth)
	rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/dokumentasi/12", "")
	if rec.Code != http.StatusOK || e.svc.id != 12 {
		t.Fatalf("status %d id=%d", rec.Code, e.svc.id)
	}
	var m map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	g := m["gambar"].([]any)
	g0, g1 := g[0].(map[string]any), g[1].(map[string]any)
	if len(g) != 2 || g0["key"] != keyDok1 || g0["url"] != "/media/"+keyDok1 || g1["key"] != keyDok2 {
		t.Errorf("gambar salah: %v", g)
	}
	if _, ada := g0["alt"]; ada {
		t.Errorf("field alt sudah dihapus: %v", g0)
	}
	if _, ada := m["gambarUtama"]; ada {
		t.Errorf("detail tidak memakai gambarUtama: %v", m)
	}
	for _, bad := range []string{"abc", "0", "-1", "1.5"} {
		e.svc.aksi = ""
		if rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/dokumentasi/"+bad, ""); rec.Code != http.StatusNotFound || e.svc.aksi != "" {
			t.Errorf("id %q: %d aksi=%q", bad, rec.Code, e.svc.aksi)
		}
	}
}

func TestAdminDokumentasiBuatMeneruskanDaftarGambarBerurutan(t *testing.T) {
	e := siapkanDokHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	rec := kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/dokumentasi", `{`+badanDokOK+`}`)
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/api/admin/dokumentasi/12" {
		t.Fatalf("buat: %d %q %s", rec.Code, rec.Header().Get("Location"), rec.Body)
	}
	m := e.svc.mentah
	if e.svc.aksi != "buat" || m.Judul != "Kegiatan Dua" || m.TanggalPublikasi != "09-11-2026" || e.svc.oleh != "dev-admin" {
		t.Errorf("argumen: %+v", e.svc)
	}
	if len(m.Gambar) != 2 || m.Gambar[0].Key != keyDok1 || m.Gambar[1].Key != keyDok2 {
		t.Errorf("daftar gambar harus diteruskan berurutan: %+v", m.Gambar)
	}
}

func TestAdminDokumentasiBodyJSONKetat(t *testing.T) {
	e := siapkanDokHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	kasus := map[string]string{
		"POST tidak boleh version":  `{` + badanDokOK + `,"version":1}`,
		"field tak dikenal":         `{` + badanDokOK + `,"status":"posted"}`,
		"item gambar: field asing":  `{"judul":"J","tanggalPublikasi":"09-11-2026","gambarDaftar":[{"key":"k","urutan":1}]}`,
		"gambar bukan array":        `{"judul":"J","tanggalPublikasi":"09-11-2026","gambarDaftar":"k"}`,
		"gambar array string":       `{"judul":"J","tanggalPublikasi":"09-11-2026","gambarDaftar":["k"]}`,
		"item gambar key bukan str": `{"judul":"J","tanggalPublikasi":"09-11-2026","gambarDaftar":[{"key":1}]}`,
		"alt tidak dikenal lagi":    `{"judul":"J","tanggalPublikasi":"09-11-2026","gambarDaftar":[{"key":"k","alt":"x"}]}`,
		"bukan JSON":                `x`,
	}
	for nama, body := range kasus {
		if rec := kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/dokumentasi", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d", nama, rec.Code)
		}
	}
	if e.svc.aksi != "" {
		t.Errorf("body tidak valid tidak boleh sampai ke layanan (aksi=%q)", e.svc.aksi)
	}
	// gambar tidak dikirim / kosong: tidak ditolak di lapisan HTTP, tetapi diteruskan apa adanya ke layanan (aturan 1-6 milik layanan)
	for nama, body := range map[string]string{
		"tanpa field gambar": `{"judul":"J","tanggalPublikasi":"09-11-2026"}`,
		"gambar kosong":      `{"judul":"J","tanggalPublikasi":"09-11-2026","gambarDaftar":[]}`,
	} {
		e.svc.aksi = ""
		kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/dokumentasi", body)
		if e.svc.aksi != "buat" || len(e.svc.mentah.Gambar) != 0 {
			t.Errorf("%s: harus diteruskan dengan daftar kosong: aksi=%q %+v", nama, e.svc.aksi, e.svc.mentah.Gambar)
		}
	}
	req := httptest.NewRequest("POST", "/api/admin/dokumentasi", strings.NewReader(`{`+badanDokOK+`}`))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Origin", originSah)
	req.Header.Set("X-CSRF-Token", csrf)
	req.AddCookie(&http.Cookie{Name: "session", Value: id})
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("content-type salah: %d", rec.Code)
	}
}

func TestAdminDokumentasiUbahWajibVersion(t *testing.T) {
	e := siapkanDokHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	for _, body := range []string{`{` + badanDokOK + `}`, `{` + badanDokOK + `,"version":0}`, `{` + badanDokOK + `,"version":"4"}`, `{` + badanDokOK + `,"version":null}`} {
		if rec := kirimAdmin(t, e.h, id, csrf, "PUT", "/api/admin/dokumentasi/12", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: %d", body, rec.Code)
		}
	}
	if e.svc.aksi != "" {
		t.Error("tanpa version tidak boleh sampai ke layanan")
	}
	rec := kirimAdmin(t, e.h, id, csrf, "PUT", "/api/admin/dokumentasi/12", `{`+badanDokOK+`,"version":4}`)
	if rec.Code != http.StatusOK || e.svc.aksi != "ubah" || e.svc.id != 12 || e.svc.version != 4 || e.svc.oleh != "dev-admin" || len(e.svc.mentah.Gambar) != 2 {
		t.Errorf("ubah: %d %+v", rec.Code, e.svc)
	}
}

func TestAdminDokumentasiAksiStatus(t *testing.T) {
	e := siapkanDokHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	for _, aksi := range []string{"post", "retract"} {
		path := "/api/admin/dokumentasi/12/" + aksi
		for _, body := range []string{``, `{}`, `{"version":0}`, `{"version":"4"}`, `{"version":4,"x":1}`} {
			e.svc.aksi = ""
			if rec := kirimAdmin(t, e.h, id, csrf, "POST", path, body); (rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnsupportedMediaType) || e.svc.aksi != "" {
				t.Errorf("%s %q: %d aksi=%q", aksi, body, rec.Code, e.svc.aksi)
			}
		}
		rec := kirimAdmin(t, e.h, id, csrf, "POST", path, `{"version":4}`)
		if rec.Code != http.StatusOK || e.svc.aksi != aksi || e.svc.id != 12 || e.svc.version != 4 || e.svc.oleh != "dev-admin" || !strings.Contains(rec.Body.String(), `"gambar":[`) {
			t.Errorf("%s: %d %+v", aksi, rec.Code, e.svc)
		}
	}
	rec := kirimAdmin(t, e.h, id, csrf, "DELETE", "/api/admin/dokumentasi/12", "")
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 || e.svc.aksi != "hapus" {
		t.Errorf("hapus: %d %+v", rec.Code, e.svc)
	}
	rec = kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/dokumentasi/12/restore", "")
	if rec.Code != http.StatusOK || e.svc.aksi != "pulihkan" || !strings.Contains(rec.Body.String(), `"slug":"kegiatan-dua"`) {
		t.Errorf("pulihkan: %d %s", rec.Code, rec.Body)
	}
	rec = kirimAdmin(t, e.h, id, csrf, "DELETE", "/api/admin/dokumentasi/12/permanen", "")
	if rec.Code != http.StatusNoContent || e.svc.aksi != "permanen" || e.svc.id != 12 {
		t.Errorf("permanen: %d %+v", rec.Code, e.svc)
	}
	for _, r := range []struct{ m, p string }{{"DELETE", "/api/admin/dokumentasi/abc"}, {"POST", "/api/admin/dokumentasi/0/restore"}, {"DELETE", "/api/admin/dokumentasi/-1/permanen"}} {
		e.svc.aksi = ""
		if rec := kirimAdmin(t, e.h, id, csrf, r.m, r.p, ""); rec.Code != http.StatusNotFound || e.svc.aksi != "" {
			t.Errorf("%s %s: %d aksi=%q", r.m, r.p, rec.Code, e.svc.aksi)
		}
	}
}

func TestAdminDokumentasiPemetaanGalat(t *testing.T) {
	e := siapkanDokHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	kasus := []struct {
		err  error
		kode int
		code string
	}{
		{&admin.ErrInput{Pesan: "Field gambar harus berisi 1 sampai 6 gambar"}, 400, kodeRequestTidakSah},
		{admin.ErrTidakDitemukan, 404, kodeTidakDitemukan},
		{admin.ErrStatusTidakValid, 409, kodeStatusTidakValid},
		{admin.ErrVersiBentrok, 409, kodeVersiBentrok},
		{admin.ErrGambarHilang, 409, kodeGambarHilang},
		{fmt.Errorf("%w: rustfs mati access_key=RAHASIA", admin.ErrStorage), 503, kodeStorageBermasalah},
		{errors.New("pq: koneksi putus password=RAHASIA"), 500, kodeInternal},
	}
	aksi := []struct{ metode, path, body string }{
		{"POST", "/api/admin/dokumentasi", `{` + badanDokOK + `}`},
		{"PUT", "/api/admin/dokumentasi/12", `{` + badanDokOK + `,"version":1}`},
		{"POST", "/api/admin/dokumentasi/12/post", `{"version":1}`},
		{"POST", "/api/admin/dokumentasi/12/retract", `{"version":1}`},
		{"DELETE", "/api/admin/dokumentasi/12", ``},
		{"POST", "/api/admin/dokumentasi/12/restore", ``},
		{"DELETE", "/api/admin/dokumentasi/12/permanen", ``},
		{"GET", "/api/admin/dokumentasi/12", ``},
		{"GET", "/api/admin/dokumentasi", ``},
	}
	for _, k := range kasus {
		e.svc.err = k.err
		for _, a := range aksi {
			rec := kirimAdmin(t, e.h, id, csrf, a.metode, a.path, a.body)
			kode, code := k.kode, k.code
			if code == kodeStatusTidakValid && a.metode == "PUT" { // Simpan pada item bukan draft: 400 bukan_draft (frontend-final)
				kode, code = 400, kodeBukanDraft
			}
			if rec.Code != kode || kodeGalat(t, rec) != code {
				t.Errorf("%v: %s %s -> %d %s", k.err, a.metode, a.path, rec.Code, rec.Body)
			}
			if strings.Contains(rec.Body.String(), "RAHASIA") || strings.Contains(rec.Body.String(), "pq:") {
				t.Errorf("detail internal bocor: %s", rec.Body)
			}
		}
	}
}
