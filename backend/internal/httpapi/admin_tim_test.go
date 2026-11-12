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

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
)

// kirimAdmin mengirim request admin sebagai pengguna login (cookie, CSRF, Origin sah) dengan body JSON.
func kirimAdmin(t *testing.T, h http.Handler, id, csrf, metode, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(metode, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Origin", originSah)
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}
	if id != "" {
		req.AddCookie(&http.Cookie{Name: "session", Value: id})
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

type layananTimPalsu struct {
	err    error
	hasil  admin.Tim
	daftar []admin.Tim

	aksi    string
	id      int64
	version int32
	oleh    string
	mentah  admin.MentahTim
}

func (l *layananTimPalsu) Daftar(context.Context) ([]admin.Tim, error) {
	l.aksi = "daftar"
	return l.daftar, l.err
}
func (l *layananTimPalsu) Detail(_ context.Context, id int64) (admin.Tim, error) {
	l.aksi, l.id = "detail", id
	return l.hasil, l.err
}
func (l *layananTimPalsu) Buat(_ context.Context, m admin.MentahTim, oleh string) (admin.Tim, error) {
	l.aksi, l.mentah, l.oleh = "buat", m, oleh
	return l.hasil, l.err
}
func (l *layananTimPalsu) Ubah(_ context.Context, id int64, m admin.MentahTim, v int32, oleh string) (admin.Tim, error) {
	l.aksi, l.id, l.mentah, l.version, l.oleh = "ubah", id, m, v, oleh
	return l.hasil, l.err
}
func (l *layananTimPalsu) Hapus(_ context.Context, id int64) error {
	l.aksi, l.id = "hapus", id
	return l.err
}

func contohTim() admin.Tim {
	return admin.Tim{
		ID: 4, Nama: "Dr. Ani", Posisi: "Ketua", BidangRiset: "NLP", FotoKey: "tim/3f9c2a10-7b4e-4c6d-9a1f-0123456789ab.png",
		Urutan: 2, Version: 5, DibuatPada: waktuUji, DibuatOleh: "dev-admin", DiubahOleh: "editor-2",
	}
}

type lingkunganTimHTTP struct {
	lingkunganAuth
	svc *layananTimPalsu
}

func siapkanTimHTTP(t *testing.T) lingkunganTimHTTP {
	t.Helper()
	svc := &layananTimPalsu{hasil: contohTim()}
	e := siapkanAuth(t, nil, func(d *Deps) { d.Admin = DepsAdmin{Tim: svc} })
	return lingkunganTimHTTP{lingkunganAuth: e, svc: svc}
}

func TestAdminTimButuhLoginCSRFDanOrigin(t *testing.T) {
	e := siapkanTimHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	rute := []struct{ metode, path, body string }{
		{"GET", "/api/admin/tim", ""},
		{"GET", "/api/admin/tim/4", ""},
		{"POST", "/api/admin/tim", `{}`},
		{"PUT", "/api/admin/tim/4", `{}`},
		{"DELETE", "/api/admin/tim/4", ""},
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
			t.Errorf("%s %s tanpa CSRF: %d %s (aksi=%q)", r.metode, r.path, rec.Code, rec.Body, e.svc.aksi)
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
			t.Errorf("%s %s Origin asing: %d %s", r.metode, r.path, rec.Code, rec.Body)
		}
	}
}

func TestAdminTimDaftarTanpaPagination(t *testing.T) {
	e := siapkanTimHTTP(t)
	id, _ := loginBerhasil(t, e.lingkunganAuth)
	e.svc.daftar = []admin.Tim{contohTim()}
	for _, q := range []string{"", "?page=2&limit=1&q=ani&status=trash"} { // parameter pagination/pencarian/filter diabaikan
		rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/tim"+q, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("%q: status %d: %s", q, rec.Code, rec.Body)
		}
		var m map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &m)
		if len(m) != 1 || m["items"] == nil {
			t.Errorf("%q: respons hanya boleh {items}: %v", q, m)
		}
		item := m["items"].([]any)[0].(map[string]any)
		want := map[string]any{
			"id": 4.0, "nama": "Dr. Ani", "posisi": "Ketua", "bidangRiset": "NLP", "urutan": 2.0, "version": 5.0,
			"dibuatPada": "2026-11-09T03:00:00Z", "dibuatOleh": "dev-admin", "diubahOleh": "editor-2",
		}
		for k, v := range want {
			if item[k] != v {
				t.Errorf("%s = %v, diharapkan %v", k, item[k], v)
			}
		}
		foto := item["foto"].(map[string]any)
		if foto["key"] != contohTim().FotoKey || foto["url"] != "/media/"+contohTim().FotoKey {
			t.Errorf("foto salah: %v", foto)
		}
		if len(item) != len(want)+1 {
			t.Errorf("field berlebih/kurang: %v", item)
		}
	}
	e.svc.daftar = nil
	rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/tim", "")
	if !strings.Contains(rec.Body.String(), `"items":[]`) {
		t.Errorf("daftar kosong harus []: %s", rec.Body)
	}
}

func TestAdminTimDetailDanIdTidakSah(t *testing.T) {
	e := siapkanTimHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	if rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/tim/4", ""); rec.Code != http.StatusOK || e.svc.id != 4 {
		t.Fatalf("detail: %d id=%d", rec.Code, e.svc.id)
	}
	for _, bad := range []string{"abc", "0", "-1", "1.5", "99999999999999999999"} {
		for _, m := range []string{"GET", "DELETE"} {
			e.svc.aksi = ""
			rec := kirimAdmin(t, e.h, id, csrf, m, "/api/admin/tim/"+bad, "")
			if rec.Code != http.StatusNotFound || e.svc.aksi != "" {
				t.Errorf("%s id %q: %d aksi=%q", m, bad, rec.Code, e.svc.aksi)
			}
		}
	}
}

func TestAdminTimBuatDanJSONKetat(t *testing.T) {
	e := siapkanTimHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	ok := `"nama":"Dr. Ani","posisi":"Ketua","bidangRiset":"NLP","foto":"tim/k.png","urutan":2`

	rec := kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/tim", `{`+ok+`}`)
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/api/admin/tim/4" {
		t.Fatalf("buat: %d %q %s", rec.Code, rec.Header().Get("Location"), rec.Body)
	}
	m := e.svc.mentah
	if e.svc.aksi != "buat" || m.Nama != "Dr. Ani" || m.Posisi != "Ketua" || m.BidangRiset != "NLP" || m.Foto != "tim/k.png" || m.Urutan == nil || *m.Urutan != 2 || e.svc.oleh != "dev-admin" {
		t.Errorf("argumen ke layanan salah: %+v oleh=%q", m, e.svc.oleh)
	}

	e.svc.aksi = ""
	kasus := map[string]string{
		"urutan pecahan":            `{"nama":"A","posisi":"B","bidangRiset":"C","foto":"tim/k.png","urutan":1.5}`,
		"urutan teks":               `{"nama":"A","posisi":"B","bidangRiset":"C","foto":"tim/k.png","urutan":"1"}`,
		"POST tidak boleh version":   `{` + ok + `,"version":1}`,
		"field tak dikenal":          `{` + ok + `,"status":"posted"}`,
		"bukan JSON":                 `x`,
		"tipe salah":                 `{"nama":1}`,
	}
	for nama, body := range kasus {
		if rec := kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/tim", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d", nama, rec.Code)
		}
	}
	req := httptest.NewRequest("POST", "/api/admin/tim", strings.NewReader(`{`+ok+`}`))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Origin", originSah)
	req.Header.Set("X-CSRF-Token", csrf)
	req.AddCookie(&http.Cookie{Name: "session", Value: id})
	rec = httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("content-type salah: %d", rec.Code)
	}
	if e.svc.aksi != "" {
		t.Errorf("body tidak valid tidak boleh sampai ke layanan (aksi=%q)", e.svc.aksi)
	}
}

func TestAdminTimUbahWajibVersion(t *testing.T) {
	e := siapkanTimHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	ok := `"nama":"A","posisi":"B","bidangRiset":"C","foto":"tim/k.png","urutan":3`
	for _, body := range []string{`{` + ok + `}`, `{` + ok + `,"version":0}`, `{` + ok + `,"version":"1"}`, `{` + ok + `,"version":null}`} {
		if rec := kirimAdmin(t, e.h, id, csrf, "PUT", "/api/admin/tim/4", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: %d", body, rec.Code)
		}
	}
	if e.svc.aksi != "" {
		t.Error("tanpa version tidak boleh sampai ke layanan")
	}
	rec := kirimAdmin(t, e.h, id, csrf, "PUT", "/api/admin/tim/4", `{`+ok+`,"version":5}`)
	if rec.Code != http.StatusOK || e.svc.aksi != "ubah" || e.svc.id != 4 || e.svc.version != 5 || e.svc.oleh != "dev-admin" {
		t.Errorf("ubah: %d %+v", rec.Code, e.svc)
	}
}

func TestAdminTimHapusLangsungTanpaVersion(t *testing.T) {
	e := siapkanTimHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	rec := kirimAdmin(t, e.h, id, csrf, "DELETE", "/api/admin/tim/4", "")
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 || e.svc.aksi != "hapus" || e.svc.id != 4 {
		t.Errorf("hapus: %d %q %+v", rec.Code, rec.Body, e.svc)
	}
	// rute trash/status tidak ada untuk tim
	for _, r := range []struct{ m, p string }{{"POST", "/api/admin/tim/4/post"}, {"POST", "/api/admin/tim/4/retract"}, {"POST", "/api/admin/tim/4/restore"}, {"DELETE", "/api/admin/tim/4/permanen"}} {
		rec := kirimAdmin(t, e.h, id, csrf, r.m, r.p, `{"version":1}`)
		if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s harus tidak ada: %d", r.m, r.p, rec.Code)
		}
	}
}

func TestAdminTimPemetaanGalat(t *testing.T) {
	e := siapkanTimHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	kasus := []struct {
		err  error
		kode int
		code string
	}{
		{&admin.ErrInput{Pesan: "Field nama wajib diisi"}, 400, kodeRequestTidakSah},
		{admin.ErrTidakDitemukan, 404, kodeTidakDitemukan},
		{admin.ErrVersiBentrok, 409, kodeVersiBentrok},
		{fmt.Errorf("%w: rustfs mati access_key=RAHASIA", admin.ErrStorage), 503, kodeStorageBermasalah},
		{errors.New("pq: koneksi putus password=RAHASIA"), 500, kodeInternal},
	}
	ok := `"nama":"A","posisi":"B","bidangRiset":"C","foto":"tim/k.png","urutan":3`
	aksi := []struct{ metode, path, body string }{
		{"POST", "/api/admin/tim", `{` + ok + `}`},
		{"PUT", "/api/admin/tim/4", `{` + ok + `,"version":1}`},
		{"DELETE", "/api/admin/tim/4", ``},
		{"GET", "/api/admin/tim/4", ``},
		{"GET", "/api/admin/tim", ``},
	}
	for _, k := range kasus {
		e.svc.err = k.err
		for _, a := range aksi {
			rec := kirimAdmin(t, e.h, id, csrf, a.metode, a.path, a.body)
			if rec.Code != k.kode || kodeGalat(t, rec) != k.code {
				t.Errorf("%v: %s %s -> %d %s", k.err, a.metode, a.path, rec.Code, rec.Body)
			}
			if strings.Contains(rec.Body.String(), "RAHASIA") || strings.Contains(rec.Body.String(), "pq:") {
				t.Errorf("detail internal bocor: %s", rec.Body)
			}
		}
	}
}
