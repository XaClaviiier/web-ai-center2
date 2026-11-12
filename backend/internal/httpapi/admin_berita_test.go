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

// layananPalsu mencatat panggilan terakhir dan mengembalikan nilai yang diatur tes.
type layananPalsu struct {
	err    error
	hasil  admin.Berita
	daftar []admin.Berita
	total  int64

	aksi    string
	param   admin.ParamDaftar
	id      int64
	version int32
	oleh    string
	mentah  admin.MentahBerita
}

func (l *layananPalsu) Daftar(_ context.Context, p admin.ParamDaftar) ([]admin.Berita, int64, error) {
	l.aksi, l.param = "daftar", p
	return l.daftar, l.total, l.err
}
func (l *layananPalsu) Detail(_ context.Context, id int64) (admin.Berita, error) {
	l.aksi, l.id = "detail", id
	return l.hasil, l.err
}
func (l *layananPalsu) Buat(_ context.Context, m admin.MentahBerita, oleh string) (admin.Berita, error) {
	l.aksi, l.mentah, l.oleh = "buat", m, oleh
	return l.hasil, l.err
}
func (l *layananPalsu) Ubah(_ context.Context, id int64, m admin.MentahBerita, v int32, oleh string) (admin.Berita, error) {
	l.aksi, l.id, l.mentah, l.version, l.oleh = "ubah", id, m, v, oleh
	return l.hasil, l.err
}
func (l *layananPalsu) Post(_ context.Context, id int64, v int32, oleh string) (admin.Berita, error) {
	l.aksi, l.id, l.version, l.oleh = "post", id, v, oleh
	return l.hasil, l.err
}
func (l *layananPalsu) Retract(_ context.Context, id int64, v int32, oleh string) (admin.Berita, error) {
	l.aksi, l.id, l.version, l.oleh = "retract", id, v, oleh
	return l.hasil, l.err
}
func (l *layananPalsu) Hapus(_ context.Context, id int64, oleh string) error {
	l.aksi, l.id, l.oleh = "hapus", id, oleh
	return l.err
}
func (l *layananPalsu) Pulihkan(_ context.Context, id int64, oleh string) (admin.Berita, error) {
	l.aksi, l.id, l.oleh = "pulihkan", id, oleh
	return l.hasil, l.err
}
func (l *layananPalsu) HapusPermanen(_ context.Context, id int64) error {
	l.aksi, l.id = "permanen", id
	return l.err
}

var waktuUji = time.Date(2026, 11, 9, 3, 0, 0, 0, time.UTC)

func contohBerita() admin.Berita {
	return admin.Berita{
		ID: 7, Slug: "judul-tujuh", Judul: "Judul Tujuh", Ringkasan: "Ringkas.", Isi: "Ringkas.\n\nLanjut.",
		TanggalPublikasi: "2026-11-09", GambarKey: "berita/3f9c2a10-7b4e-4c6d-9a1f-0123456789ab.png",
		Status: "draft", Version: 3, DibuatPada: waktuUji, DibuatOleh: "dev-admin", DiubahOleh: "editor-2",
	}
}

type lingkunganBerita struct {
	lingkunganAuth
	svc *layananPalsu
}

func siapkanBerita(t *testing.T) lingkunganBerita {
	t.Helper()
	svc := &layananPalsu{hasil: contohBerita()}
	e := siapkanAuth(t, nil, func(d *Deps) { d.Admin = DepsAdmin{Berita: svc} })
	return lingkunganBerita{lingkunganAuth: e, svc: svc}
}

// kirimJSON mengirim request admin sebagai pengguna login (cookie, CSRF, Origin) dengan body JSON.
func (e lingkunganBerita) kirimJSON(t *testing.T, id, csrf, metode, path, body string) *httptest.ResponseRecorder {
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
	e.h.ServeHTTP(rec, req)
	return rec
}

func (e lingkunganBerita) login(t *testing.T) (id, csrf string) {
	t.Helper()
	return loginBerhasil(t, e.lingkunganAuth)
}

func TestAdminBeritaButuhLoginCSRFDanOrigin(t *testing.T) {
	e := siapkanBerita(t)
	id, csrf := e.login(t)
	rute := []struct{ metode, path, body string }{
		{"GET", "/api/admin/berita", ""},
		{"GET", "/api/admin/berita/7", ""},
		{"POST", "/api/admin/berita", `{}`},
		{"PUT", "/api/admin/berita/7", `{}`},
		{"DELETE", "/api/admin/berita/7", ""},
		{"POST", "/api/admin/berita/7/post", `{"version":1}`},
		{"POST", "/api/admin/berita/7/retract", `{"version":1}`},
		{"POST", "/api/admin/berita/7/restore", ""},
		{"DELETE", "/api/admin/berita/7/permanen", ""},
	}
	for _, r := range rute {
		ubah := r.metode != "GET"
		// tanpa login
		e.svc.aksi = ""
		if rec := e.kirimJSON(t, "", "", r.metode, r.path, r.body); rec.Code != http.StatusUnauthorized && !(ubah && rec.Code == http.StatusForbidden) {
			t.Errorf("%s %s tanpa login: %d", r.metode, r.path, rec.Code)
		}
		if e.svc.aksi != "" {
			t.Errorf("%s %s tanpa login: layanan tidak boleh dipanggil (aksi=%q)", r.metode, r.path, e.svc.aksi)
		}
		// login tetapi tanpa CSRF: hanya metode ubah-data yang ditolak (GET sah dan memang sampai ke layanan)
		e.svc.aksi = ""
		rec := e.kirimJSON(t, id, "", r.metode, r.path, r.body)
		if ubah && (rec.Code != http.StatusForbidden || kodeGalat(t, rec) != kodeCSRFTidakValid) {
			t.Errorf("%s %s tanpa CSRF: %d %s", r.metode, r.path, rec.Code, rec.Body)
		}
		if ubah && e.svc.aksi != "" {
			t.Errorf("%s %s tanpa CSRF: layanan tidak boleh dipanggil (aksi=%q)", r.metode, r.path, e.svc.aksi)
		}
		if !ubah && rec.Code == http.StatusForbidden {
			t.Errorf("%s %s: GET tidak boleh butuh CSRF", r.metode, r.path)
		}
		e.svc.aksi = ""
		// Origin asing
		req := httptest.NewRequest(r.metode, r.path, strings.NewReader(r.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://evil.example.com")
		req.Header.Set("X-CSRF-Token", csrf)
		req.AddCookie(&http.Cookie{Name: "session", Value: id})
		rec = httptest.NewRecorder()
		e.h.ServeHTTP(rec, req)
		if ubah && (rec.Code != http.StatusForbidden || kodeGalat(t, rec) != kodeOriginTidakSah) {
			t.Errorf("%s %s Origin asing: %d %s", r.metode, r.path, rec.Code, rec.Body)
		}
		if ubah && e.svc.aksi != "" {
			t.Errorf("%s %s Origin asing: layanan tidak boleh dipanggil (aksi=%q)", r.metode, r.path, e.svc.aksi)
		}
	}
}

func TestAdminBeritaNoStore(t *testing.T) {
	e := siapkanBerita(t)
	id, _ := e.login(t)
	rec := e.kirimJSON(t, id, "", "GET", "/api/admin/berita/7", "")
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("respons admin harus no-store: %q", rec.Header().Get("Cache-Control"))
	}
}

func TestAdminBeritaDaftar(t *testing.T) {
	e := siapkanBerita(t)
	id, _ := e.login(t)
	dihapus := waktuUji.Add(time.Hour)
	b2 := contohBerita()
	b2.ID, b2.Slug, b2.DihapusPada = 8, "dua", &dihapus
	e.svc.daftar, e.svc.total = []admin.Berita{contohBerita(), b2}, 42

	rec := e.kirimJSON(t, id, "", "GET", "/api/admin/berita", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if e.svc.param.Filter != "semua" || e.svc.param.Batas != limitDaftarAdminDefault || e.svc.param.Geser != 0 || e.svc.param.Pola != "" {
		t.Errorf("default salah: %+v", e.svc.param)
	}
	var m map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	if m["total"].(float64) != 42 || m["page"].(float64) != 1 || m["limit"].(float64) != limitDaftarAdminDefault {
		t.Errorf("meta salah: %v", m)
	}
	// pagination tetap tersedia bila diminta eksplisit
	e.kirimJSON(t, id, "", "GET", "/api/admin/berita?page=3&limit=20", "")
	if e.svc.param.Batas != 20 || e.svc.param.Geser != 40 {
		t.Errorf("page=3&limit=20 -> %+v", e.svc.param)
	}
	items := m["items"].([]any)
	i0, i1 := items[0].(map[string]any), items[1].(map[string]any)
	if _, ada := i0["isi"]; ada {
		t.Errorf("daftar tidak boleh memuat isi: %v", i0)
	}
	if _, ada := i0["dihapusPada"]; ada {
		t.Errorf("dihapusPada hanya untuk item trash: %v", i0)
	}
	if i1["dihapusPada"] == nil {
		t.Errorf("item trash harus memuat dihapusPada: %v", i1)
	}
	want := map[string]any{
		"id": 7.0, "slug": "judul-tujuh", "judul": "Judul Tujuh", "ringkasan": "Ringkas.", "tanggalPublikasi": "2026-11-09",
		"status": "draft", "version": 3.0, "dibuatPada": "2026-11-09T03:00:00Z", "dibuatOleh": "dev-admin", "diubahOleh": "editor-2",
	}
	for k, v := range want {
		if i0[k] != v {
			t.Errorf("%s = %v, diharapkan %v", k, i0[k], v)
		}
	}
	g := i0["gambar"].(map[string]any)
	if g["key"] != contohBerita().GambarKey || g["url"] != "/media/"+contohBerita().GambarKey {
		t.Errorf("gambar salah: %v", g)
	}
}

func TestAdminBeritaDaftarParameter(t *testing.T) {
	e := siapkanBerita(t)
	id, _ := e.login(t)
	rec := e.kirimJSON(t, id, "", "GET", "/api/admin/berita?status=trash&page=3&limit=500&q=ab%25", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	p := e.svc.param
	if p.Filter != "trash" || p.Batas != 50 || p.Geser != 100 || p.Pola != `%ab\%%` {
		t.Errorf("param ke layanan salah: %+v", p)
	}
	for _, q := range []string{"status=hapus", "status=", "status=SEMUA", "page=0", "limit=0", "q=a", "page=x"} {
		rec := e.kirimJSON(t, id, "", "GET", "/api/admin/berita?"+q, "")
		if rec.Code != http.StatusBadRequest || kodeGalat(t, rec) != kodeQueryTidakValid {
			t.Errorf("?%s: %d %s", q, rec.Code, rec.Body)
		}
	}
	for _, f := range []string{"semua", "draft", "posted", "trash"} {
		if rec := e.kirimJSON(t, id, "", "GET", "/api/admin/berita?status="+f, ""); rec.Code != http.StatusOK {
			t.Errorf("status=%s harus diterima: %d", f, rec.Code)
		}
	}
}

func TestAdminBeritaDetail(t *testing.T) {
	e := siapkanBerita(t)
	id, _ := e.login(t)
	rec := e.kirimJSON(t, id, "", "GET", "/api/admin/berita/7", "")
	if rec.Code != http.StatusOK || e.svc.id != 7 {
		t.Fatalf("status %d id=%d", rec.Code, e.svc.id)
	}
	var m map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	if m["isi"] != "Ringkas.\n\nLanjut." {
		t.Errorf("detail harus memuat isi: %v", m)
	}
	for _, bad := range []string{"abc", "0", "-1", "1.5", "9999999999999999999999"} {
		e.svc.aksi = ""
		rec := e.kirimJSON(t, id, "", "GET", "/api/admin/berita/"+bad, "")
		if rec.Code != http.StatusNotFound || e.svc.aksi != "" {
			t.Errorf("id %q: status %d, layanan dipanggil=%q", bad, rec.Code, e.svc.aksi)
		}
	}
}

func TestAdminBeritaBuat(t *testing.T) {
	e := siapkanBerita(t)
	id, csrf := e.login(t)
	body := `{"judul":"Judul","isi":"Isi","tanggalPublikasi":"2026-11-09","gambar":"berita/k.png"}`
	rec := e.kirimJSON(t, id, csrf, "POST", "/api/admin/berita", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Location") != "/api/admin/berita/7" {
		t.Errorf("Location salah: %q", rec.Header().Get("Location"))
	}
	m := e.svc.mentah
	if e.svc.aksi != "buat" || m.Judul != "Judul" || m.Isi != "Isi" || m.TanggalPublikasi != "2026-11-09" || m.Gambar != "berita/k.png" || e.svc.oleh != "dev-admin" {
		t.Errorf("argumen ke layanan salah: %+v oleh=%q", e.svc, e.svc.oleh)
	}
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if out["isi"] == nil {
		t.Errorf("respons buat harus memuat detail lengkap: %v", out)
	}
}

func TestAdminBeritaBodyJSONKetat(t *testing.T) {
	e := siapkanBerita(t)
	id, csrf := e.login(t)
	ok := `"judul":"J","isi":"I","tanggalPublikasi":"2026-11-09","gambar":"berita/k.png"`
	kasus := map[string]struct {
		body  string
		ctype string
		kode  int
	}{
		"field tak dikenal":        {`{` + ok + `,"status":"posted"}`, "application/json", 400},
		"POST tidak boleh version": {`{` + ok + `,"version":1}`, "application/json", 400},
		"bukan JSON":               {`bukan json`, "application/json", 400},
		"body kosong":              {``, "application/json", 400},
		"array":                    {`[1,2]`, "application/json", 400},
		"data setelah objek":       {`{` + ok + `} {}`, "application/json", 400},
		"tipe field salah":         {`{"judul":123}`, "application/json", 400},
		"content-type salah":       {`{` + ok + `}`, "text/plain", 415},
		"form-encoded":             {`judul=x`, "application/x-www-form-urlencoded", 415},
		"terlalu besar":            {`{"judul":"` + strings.Repeat("a", batasBodyJSON) + `"}`, "application/json", 413},
	}
	for nama, k := range kasus {
		req := httptest.NewRequest("POST", "/api/admin/berita", strings.NewReader(k.body))
		req.Header.Set("Content-Type", k.ctype)
		req.Header.Set("Origin", originSah)
		req.Header.Set("X-CSRF-Token", csrf)
		req.AddCookie(&http.Cookie{Name: "session", Value: id})
		rec := httptest.NewRecorder()
		e.h.ServeHTTP(rec, req)
		if rec.Code != k.kode {
			t.Errorf("%s: status %d (diharapkan %d): %s", nama, rec.Code, k.kode, rec.Body)
		}
	}
	if e.svc.aksi != "" {
		t.Errorf("body tidak valid tidak boleh sampai ke layanan (aksi=%q)", e.svc.aksi)
	}
}

func TestAdminBeritaUbahWajibVersion(t *testing.T) {
	e := siapkanBerita(t)
	id, csrf := e.login(t)
	ok := `"judul":"J","isi":"I","tanggalPublikasi":"2026-11-09","gambar":"berita/k.png"`
	for _, body := range []string{`{` + ok + `}`, `{` + ok + `,"version":0}`, `{` + ok + `,"version":-2}`, `{` + ok + `,"version":"1"}`, `{` + ok + `,"version":null}`} {
		rec := e.kirimJSON(t, id, csrf, "PUT", "/api/admin/berita/7", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d", body, rec.Code)
		}
	}
	if e.svc.aksi != "" {
		t.Error("tanpa version tidak boleh sampai ke layanan")
	}
	rec := e.kirimJSON(t, id, csrf, "PUT", "/api/admin/berita/7", `{`+ok+`,"version":3}`)
	if rec.Code != http.StatusOK || e.svc.aksi != "ubah" || e.svc.id != 7 || e.svc.version != 3 || e.svc.oleh != "dev-admin" || e.svc.mentah.Gambar != "berita/k.png" {
		t.Errorf("status %d, argumen: %+v", rec.Code, e.svc)
	}
}

func TestAdminBeritaPostRetract(t *testing.T) {
	e := siapkanBerita(t)
	id, csrf := e.login(t)
	for _, aksi := range []string{"post", "retract"} {
		path := "/api/admin/berita/7/" + aksi
		for _, body := range []string{``, `{}`, `{"version":0}`, `{"version":"2"}`, `{"version":2,"x":1}`} {
			e.svc.aksi = ""
			if rec := e.kirimJSON(t, id, csrf, "POST", path, body); rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnsupportedMediaType || e.svc.aksi != "" {
				t.Errorf("%s %q: status %d, layanan=%q", aksi, body, rec.Code, e.svc.aksi)
			}
		}
		rec := e.kirimJSON(t, id, csrf, "POST", path, `{"version":2}`)
		if rec.Code != http.StatusOK || e.svc.aksi != aksi || e.svc.id != 7 || e.svc.version != 2 || e.svc.oleh != "dev-admin" {
			t.Errorf("%s: status %d, %+v", aksi, rec.Code, e.svc)
		}
	}
}

func TestAdminBeritaHapusPulihkanPermanen(t *testing.T) {
	e := siapkanBerita(t)
	id, csrf := e.login(t)

	rec := e.kirimJSON(t, id, csrf, "DELETE", "/api/admin/berita/7", "")
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 || e.svc.aksi != "hapus" || e.svc.oleh != "dev-admin" {
		t.Errorf("hapus: %d %q %+v", rec.Code, rec.Body, e.svc)
	}
	rec = e.kirimJSON(t, id, csrf, "POST", "/api/admin/berita/7/restore", "")
	if rec.Code != http.StatusOK || e.svc.aksi != "pulihkan" || !strings.Contains(rec.Body.String(), `"slug":"judul-tujuh"`) {
		t.Errorf("pulihkan: %d %s", rec.Code, rec.Body)
	}
	rec = e.kirimJSON(t, id, csrf, "DELETE", "/api/admin/berita/7/permanen", "")
	if rec.Code != http.StatusNoContent || e.svc.aksi != "permanen" || e.svc.id != 7 {
		t.Errorf("permanen: %d %+v", rec.Code, e.svc)
	}
	for _, r := range []struct{ m, p string }{{"DELETE", "/api/admin/berita/abc"}, {"POST", "/api/admin/berita/0/restore"}, {"DELETE", "/api/admin/berita/-5/permanen"}} {
		e.svc.aksi = ""
		if rec := e.kirimJSON(t, id, csrf, r.m, r.p, ""); rec.Code != http.StatusNotFound || e.svc.aksi != "" {
			t.Errorf("%s %s: %d layanan=%q", r.m, r.p, rec.Code, e.svc.aksi)
		}
	}
}

func TestAdminBeritaPemetaanGalat(t *testing.T) {
	e := siapkanBerita(t)
	id, csrf := e.login(t)
	kasus := []struct {
		nama string
		err  error
		kode int
		code string
	}{
		{"input salah", &admin.ErrInput{Pesan: "Field judul wajib diisi"}, 400, kodeRequestTidakSah},
		{"input salah terbungkus", fmt.Errorf("x: %w", &admin.ErrInput{Pesan: "Field isi wajib diisi"}), 400, kodeRequestTidakSah},
		{"tidak ditemukan", admin.ErrTidakDitemukan, 404, kodeTidakDitemukan},
		{"status tidak valid", admin.ErrStatusTidakValid, 409, kodeStatusTidakValid},
		{"version bentrok", admin.ErrVersiBentrok, 409, kodeVersiBentrok},
		{"gambar hilang", admin.ErrGambarHilang, 409, kodeGambarHilang},
		{"storage", fmt.Errorf("%w: rustfs mati access_key=RAHASIA", admin.ErrStorage), 503, kodeStorageBermasalah},
		{"galat lain", errors.New("pq: koneksi putus password=RAHASIA"), 500, kodeInternal},
	}
	aksi := []struct{ metode, path, body string }{
		{"POST", "/api/admin/berita", `{"judul":"J","isi":"I","tanggalPublikasi":"2026-11-09","gambar":"b"}`},
		{"PUT", "/api/admin/berita/7", `{"judul":"J","isi":"I","tanggalPublikasi":"2026-11-09","gambar":"b","version":1}`},
		{"POST", "/api/admin/berita/7/post", `{"version":1}`},
		{"POST", "/api/admin/berita/7/retract", `{"version":1}`},
		{"DELETE", "/api/admin/berita/7", ``},
		{"POST", "/api/admin/berita/7/restore", ``},
		{"DELETE", "/api/admin/berita/7/permanen", ``},
		{"GET", "/api/admin/berita/7", ``},
		{"GET", "/api/admin/berita", ``},
	}
	for _, k := range kasus {
		e.svc.err = k.err
		for _, a := range aksi {
			rec := e.kirimJSON(t, id, csrf, a.metode, a.path, a.body)
			if rec.Code != k.kode || kodeGalat(t, rec) != k.code {
				t.Errorf("%s: %s %s -> %d %s", k.nama, a.metode, a.path, rec.Code, rec.Body)
			}
			if strings.Contains(rec.Body.String(), "RAHASIA") || strings.Contains(rec.Body.String(), "pq:") {
				t.Errorf("%s: detail internal bocor: %s", k.nama, rec.Body)
			}
		}
	}
	// pesan ErrInput diteruskan apa adanya ke klien
	e.svc.err = &admin.ErrInput{Pesan: "Field judul wajib diisi"}
	rec := e.kirimJSON(t, id, csrf, "POST", "/api/admin/berita", `{"judul":"J","isi":"I","tanggalPublikasi":"2026-11-09","gambar":"b"}`)
	if !strings.Contains(rec.Body.String(), "Field judul wajib diisi") {
		t.Errorf("pesan validasi harus sampai ke klien: %s", rec.Body)
	}
}

func itoaID(n int64) string { return fmt.Sprintf("%d", n) }
