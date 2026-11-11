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

type layananTestimoniPalsu struct {
	err    error
	hasil  admin.Testimoni
	daftar []admin.Testimoni

	aksi    string
	filter  string
	id      int64
	version int32
	oleh    string
	mentah  admin.MentahTestimoni
}

func (l *layananTestimoniPalsu) Daftar(_ context.Context, filter string) ([]admin.Testimoni, error) {
	l.aksi, l.filter = "daftar", filter
	return l.daftar, l.err
}
func (l *layananTestimoniPalsu) Detail(_ context.Context, id int64) (admin.Testimoni, error) {
	l.aksi, l.id = "detail", id
	return l.hasil, l.err
}
func (l *layananTestimoniPalsu) Buat(_ context.Context, m admin.MentahTestimoni, oleh string) (admin.Testimoni, error) {
	l.aksi, l.mentah, l.oleh = "buat", m, oleh
	return l.hasil, l.err
}
func (l *layananTestimoniPalsu) Ubah(_ context.Context, id int64, m admin.MentahTestimoni, v int32, oleh string) (admin.Testimoni, error) {
	l.aksi, l.id, l.mentah, l.version, l.oleh = "ubah", id, m, v, oleh
	return l.hasil, l.err
}
func (l *layananTestimoniPalsu) Post(_ context.Context, id int64, v int32, oleh string) (admin.Testimoni, error) {
	l.aksi, l.id, l.version, l.oleh = "post", id, v, oleh
	return l.hasil, l.err
}
func (l *layananTestimoniPalsu) Retract(_ context.Context, id int64, v int32, oleh string) (admin.Testimoni, error) {
	l.aksi, l.id, l.version, l.oleh = "retract", id, v, oleh
	return l.hasil, l.err
}
func (l *layananTestimoniPalsu) Hapus(_ context.Context, id int64, oleh string) error {
	l.aksi, l.id, l.oleh = "hapus", id, oleh
	return l.err
}
func (l *layananTestimoniPalsu) Pulihkan(_ context.Context, id int64, oleh string) (admin.Testimoni, error) {
	l.aksi, l.id, l.oleh = "pulihkan", id, oleh
	return l.hasil, l.err
}
func (l *layananTestimoniPalsu) HapusPermanen(_ context.Context, id int64) error {
	l.aksi, l.id = "permanen", id
	return l.err
}

func contohTestimoni() admin.Testimoni {
	return admin.Testimoni{
		ID: 9, Nama: "Ani", Teks: "Sangat membantu.", Posisi: "Dosen", FotoKey: "testimoni/3f9c2a10-7b4e-4c6d-9a1f-0123456789ab.png",
		Status: "draft", Version: 2, DibuatPada: waktuUji, DibuatOleh: "dev-admin", DiubahOleh: "editor-2",
	}
}

type lingkunganTestimoniHTTP struct {
	lingkunganAuth
	svc *layananTestimoniPalsu
}

func siapkanTestimoniHTTP(t *testing.T) lingkunganTestimoniHTTP {
	t.Helper()
	svc := &layananTestimoniPalsu{hasil: contohTestimoni()}
	e := siapkanAuth(t, nil, func(d *Deps) { d.Admin = DepsAdmin{Testimoni: svc} })
	return lingkunganTestimoniHTTP{lingkunganAuth: e, svc: svc}
}

func TestAdminTestimoniButuhLoginCSRFDanOrigin(t *testing.T) {
	e := siapkanTestimoniHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	rute := []struct{ metode, path, body string }{
		{"GET", "/api/admin/testimoni", ""},
		{"GET", "/api/admin/testimoni/9", ""},
		{"POST", "/api/admin/testimoni", `{}`},
		{"PUT", "/api/admin/testimoni/9", `{}`},
		{"DELETE", "/api/admin/testimoni/9", ""},
		{"POST", "/api/admin/testimoni/9/post", `{"version":1}`},
		{"POST", "/api/admin/testimoni/9/retract", `{"version":1}`},
		{"POST", "/api/admin/testimoni/9/restore", ""},
		{"DELETE", "/api/admin/testimoni/9/permanen", ""},
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

func TestAdminTestimoniDaftar(t *testing.T) {
	e := siapkanTestimoniHTTP(t)
	id, _ := loginBerhasil(t, e.lingkunganAuth)
	hapus := waktuUji.Add(time.Hour)
	tanpaFoto := contohTestimoni()
	tanpaFoto.ID, tanpaFoto.FotoKey, tanpaFoto.DihapusPada = 10, "", &hapus
	e.svc.daftar = []admin.Testimoni{contohTestimoni(), tanpaFoto}

	rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/testimoni", "")
	if rec.Code != http.StatusOK || e.svc.filter != "semua" {
		t.Fatalf("status %d filter=%q: %s", rec.Code, e.svc.filter, rec.Body)
	}
	var m map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	if len(m) != 1 {
		t.Errorf("tanpa pagination: hanya {items}: %v", m)
	}
	items := m["items"].([]any)
	i0, i1 := items[0].(map[string]any), items[1].(map[string]any)
	want := map[string]any{
		"id": 9.0, "nama": "Ani", "teks": "Sangat membantu.", "posisi": "Dosen", "status": "draft", "version": 2.0,
		"dibuatPada": "2026-11-09T03:00:00Z", "dibuatOleh": "dev-admin", "diubahOleh": "editor-2",
	}
	for k, v := range want {
		if i0[k] != v {
			t.Errorf("%s = %v, diharapkan %v", k, i0[k], v)
		}
	}
	foto := i0["foto"].(map[string]any)
	if foto["key"] != contohTestimoni().FotoKey || foto["url"] != "/media/"+contohTestimoni().FotoKey {
		t.Errorf("foto salah: %v", foto)
	}
	if _, ada := i0["dihapusPada"]; ada {
		t.Errorf("dihapusPada hanya untuk trash: %v", i0)
	}
	if _, ada := i1["foto"]; ada {
		t.Errorf("tanpa foto: field foto harus dihilangkan: %v", i1)
	}
	if i1["dihapusPada"] == nil {
		t.Errorf("trash harus memuat dihapusPada: %v", i1)
	}
}

func TestAdminTestimoniFilterStatusDanParameterDiabaikan(t *testing.T) {
	e := siapkanTestimoniHTTP(t)
	id, _ := loginBerhasil(t, e.lingkunganAuth)
	for _, f := range []string{"semua", "draft", "posted", "trash"} {
		if rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/testimoni?status="+f, ""); rec.Code != http.StatusOK || e.svc.filter != f {
			t.Errorf("status=%s: %d filter=%q", f, rec.Code, e.svc.filter)
		}
	}
	for _, q := range []string{"status=hapus", "status=", "status=SEMUA"} {
		if rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/testimoni?"+q, ""); rec.Code != http.StatusBadRequest || kodeGalat(t, rec) != kodeQueryTidakValid {
			t.Errorf("?%s: %d", q, rec.Code)
		}
	}
	// page/limit/q tidak berlaku untuk testimoni (tanpa pagination dan tanpa pencarian): diabaikan, bukan galat
	rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/testimoni?page=0&limit=x&q=a", "")
	if rec.Code != http.StatusOK {
		t.Errorf("parameter yang tidak berlaku harus diabaikan: %d %s", rec.Code, rec.Body)
	}
}

func TestAdminTestimoniFotoWajibHadir(t *testing.T) {
	e := siapkanTestimoniHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	dasar := `"nama":"Ani","teks":"Bagus","posisi":"Dosen"`

	// tidak ada field foto sama sekali: 400 (bukan "tetap", bukan "hapus")
	for _, r := range []struct{ metode, path, body string }{
		{"POST", "/api/admin/testimoni", `{` + dasar + `}`},
		{"PUT", "/api/admin/testimoni/9", `{` + dasar + `,"version":2}`},
	} {
		e.svc.aksi = ""
		rec := kirimAdmin(t, e.h, id, csrf, r.metode, r.path, r.body)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "foto") || e.svc.aksi != "" {
			t.Errorf("%s tanpa foto: %d %s aksi=%q", r.metode, rec.Code, rec.Body, e.svc.aksi)
		}
	}
	// tiga bentuk yang sah: key, null, string kosong
	kasus := []struct{ foto, want string }{
		{`"testimoni/k.png"`, "testimoni/k.png"},
		{`null`, ""},
		{`""`, ""},
	}
	for _, k := range kasus {
		rec := kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/testimoni", `{`+dasar+`,"foto":`+k.foto+`}`)
		if rec.Code != http.StatusCreated || e.svc.mentah.Foto != k.want {
			t.Errorf("foto=%s: %d foto ke layanan=%q (diharapkan %q)", k.foto, rec.Code, e.svc.mentah.Foto, k.want)
		}
		rec = kirimAdmin(t, e.h, id, csrf, "PUT", "/api/admin/testimoni/9", `{`+dasar+`,"foto":`+k.foto+`,"version":2}`)
		if rec.Code != http.StatusOK || e.svc.mentah.Foto != k.want {
			t.Errorf("PUT foto=%s: %d foto ke layanan=%q", k.foto, rec.Code, e.svc.mentah.Foto)
		}
	}
	// tipe salah
	for _, foto := range []string{`123`, `true`, `["a"]`, `{"key":"a"}`} {
		e.svc.aksi = ""
		rec := kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/testimoni", `{`+dasar+`,"foto":`+foto+`}`)
		if rec.Code != http.StatusBadRequest || e.svc.aksi != "" {
			t.Errorf("foto=%s harus 400: %d", foto, rec.Code)
		}
	}
}

func TestAdminTestimoniBuatUbahDanJSONKetat(t *testing.T) {
	e := siapkanTestimoniHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	ok := `"nama":"Ani","teks":"Bagus","posisi":"Dosen","foto":null`

	rec := kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/testimoni", `{`+ok+`}`)
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/api/admin/testimoni/9" {
		t.Fatalf("buat: %d %q", rec.Code, rec.Header().Get("Location"))
	}
	m := e.svc.mentah
	if e.svc.aksi != "buat" || m.Nama != "Ani" || m.Teks != "Bagus" || m.Posisi != "Dosen" || e.svc.oleh != "dev-admin" {
		t.Errorf("argumen: %+v", e.svc)
	}
	e.svc.aksi = ""
	for nama, body := range map[string]string{
		"POST tidak boleh version": `{` + ok + `,"version":1}`,
		"field tak dikenal":        `{` + ok + `,"status":"posted"}`,
		"urutan tidak ada":         `{` + ok + `,"urutan":3}`,
		"bukan JSON":               `x`,
	} {
		if rec := kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/testimoni", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: %d", nama, rec.Code)
		}
	}
	for _, body := range []string{`{` + ok + `}`, `{` + ok + `,"version":0}`, `{` + ok + `,"version":"2"}`} {
		if rec := kirimAdmin(t, e.h, id, csrf, "PUT", "/api/admin/testimoni/9", body); rec.Code != http.StatusBadRequest {
			t.Errorf("PUT %s: %d", body, rec.Code)
		}
	}
	if e.svc.aksi != "" {
		t.Errorf("body tidak valid tidak boleh sampai ke layanan (aksi=%q)", e.svc.aksi)
	}
	rec = kirimAdmin(t, e.h, id, csrf, "PUT", "/api/admin/testimoni/9", `{`+ok+`,"version":2}`)
	if rec.Code != http.StatusOK || e.svc.aksi != "ubah" || e.svc.version != 2 || e.svc.id != 9 {
		t.Errorf("ubah: %d %+v", rec.Code, e.svc)
	}
}

func TestAdminTestimoniAksiStatus(t *testing.T) {
	e := siapkanTestimoniHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	for _, aksi := range []string{"post", "retract"} {
		path := "/api/admin/testimoni/9/" + aksi
		for _, body := range []string{``, `{}`, `{"version":0}`, `{"version":"2"}`, `{"version":2,"x":1}`} {
			e.svc.aksi = ""
			if rec := kirimAdmin(t, e.h, id, csrf, "POST", path, body); (rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnsupportedMediaType) || e.svc.aksi != "" {
				t.Errorf("%s %q: %d aksi=%q", aksi, body, rec.Code, e.svc.aksi)
			}
		}
		rec := kirimAdmin(t, e.h, id, csrf, "POST", path, `{"version":2}`)
		if rec.Code != http.StatusOK || e.svc.aksi != aksi || e.svc.id != 9 || e.svc.version != 2 || e.svc.oleh != "dev-admin" {
			t.Errorf("%s: %d %+v", aksi, rec.Code, e.svc)
		}
	}
	rec := kirimAdmin(t, e.h, id, csrf, "DELETE", "/api/admin/testimoni/9", "")
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 || e.svc.aksi != "hapus" {
		t.Errorf("hapus: %d %+v", rec.Code, e.svc)
	}
	rec = kirimAdmin(t, e.h, id, csrf, "POST", "/api/admin/testimoni/9/restore", "")
	if rec.Code != http.StatusOK || e.svc.aksi != "pulihkan" || !strings.Contains(rec.Body.String(), `"nama":"Ani"`) {
		t.Errorf("pulihkan: %d %s", rec.Code, rec.Body)
	}
	rec = kirimAdmin(t, e.h, id, csrf, "DELETE", "/api/admin/testimoni/9/permanen", "")
	if rec.Code != http.StatusNoContent || e.svc.aksi != "permanen" || e.svc.id != 9 {
		t.Errorf("permanen: %d %+v", rec.Code, e.svc)
	}
	for _, r := range []struct{ m, p string }{{"GET", "/api/admin/testimoni/abc"}, {"DELETE", "/api/admin/testimoni/0"}, {"POST", "/api/admin/testimoni/-1/restore"}, {"DELETE", "/api/admin/testimoni/x/permanen"}} {
		e.svc.aksi = ""
		if rec := kirimAdmin(t, e.h, id, csrf, r.m, r.p, ""); rec.Code != http.StatusNotFound || e.svc.aksi != "" {
			t.Errorf("%s %s: %d aksi=%q", r.m, r.p, rec.Code, e.svc.aksi)
		}
	}
}

func TestAdminTestimoniPemetaanGalat(t *testing.T) {
	e := siapkanTestimoniHTTP(t)
	id, csrf := loginBerhasil(t, e.lingkunganAuth)
	kasus := []struct {
		err  error
		kode int
		code string
	}{
		{&admin.ErrInput{Pesan: "Field nama wajib diisi"}, 400, kodeRequestTidakSah},
		{admin.ErrTidakDitemukan, 404, kodeTidakDitemukan},
		{admin.ErrStatusTidakValid, 409, kodeStatusTidakValid},
		{admin.ErrVersiBentrok, 409, kodeVersiBentrok},
		{admin.ErrGambarHilang, 409, kodeGambarHilang},
		{fmt.Errorf("%w: rustfs mati access_key=RAHASIA", admin.ErrStorage), 503, kodeStorageBermasalah},
		{errors.New("pq: koneksi putus password=RAHASIA"), 500, kodeInternal},
	}
	ok := `"nama":"A","teks":"T","posisi":"P","foto":null`
	aksi := []struct{ metode, path, body string }{
		{"POST", "/api/admin/testimoni", `{` + ok + `}`},
		{"PUT", "/api/admin/testimoni/9", `{` + ok + `,"version":1}`},
		{"POST", "/api/admin/testimoni/9/post", `{"version":1}`},
		{"POST", "/api/admin/testimoni/9/retract", `{"version":1}`},
		{"DELETE", "/api/admin/testimoni/9", ``},
		{"POST", "/api/admin/testimoni/9/restore", ``},
		{"DELETE", "/api/admin/testimoni/9/permanen", ``},
		{"GET", "/api/admin/testimoni/9", ``},
		{"GET", "/api/admin/testimoni", ``},
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
