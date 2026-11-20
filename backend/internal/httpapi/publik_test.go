package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
)

// kontenPalsu adalah implementasi palsu konten.BacaPublik untuk tes handler.
type kontenPalsu struct {
	berita       []konten.BeritaRingkas
	total        int64
	beritaDetail konten.BeritaDetail
	dokum        []konten.DokumentasiRingkas
	dokumDetail  konten.DokumentasiDetail
	testimoni    []konten.Testimoni
	tim          []konten.AnggotaTim
	err          error
	tangkap      *konten.ParamDaftar
}

func (k kontenPalsu) ListBerita(_ context.Context, p konten.ParamDaftar) ([]konten.BeritaRingkas, int64, error) {
	if k.tangkap != nil {
		*k.tangkap = p
	}
	return k.berita, k.total, k.err
}

func (k kontenPalsu) BeritaBySlug(_ context.Context, slug string) (konten.BeritaDetail, error) {
	if k.err != nil {
		return konten.BeritaDetail{}, k.err
	}
	if k.beritaDetail.Slug != slug {
		return konten.BeritaDetail{}, konten.ErrTidakDitemukan
	}
	return k.beritaDetail, nil
}

func (k kontenPalsu) ListDokumentasi(_ context.Context, p konten.ParamDaftar) ([]konten.DokumentasiRingkas, int64, error) {
	if k.tangkap != nil {
		*k.tangkap = p
	}
	return k.dokum, k.total, k.err
}

func (k kontenPalsu) DokumentasiBySlug(_ context.Context, slug string) (konten.DokumentasiDetail, error) {
	if k.err != nil {
		return konten.DokumentasiDetail{}, k.err
	}
	if k.dokumDetail.Slug != slug {
		return konten.DokumentasiDetail{}, konten.ErrTidakDitemukan
	}
	return k.dokumDetail, nil
}

func (k kontenPalsu) ListTestimoni(context.Context) ([]konten.Testimoni, error) {
	return k.testimoni, k.err
}

func (k kontenPalsu) ListTim(context.Context) ([]konten.AnggotaTim, error) {
	return k.tim, k.err
}

func routerKonten(k kontenPalsu) http.Handler {
	return routerUjiDengan(pingerPalsu{}, k)
}

func TestParseDaftar(t *testing.T) {
	cases := []struct {
		nama  string
		query string
		page  int
		limit int
		pola  string
		galat string
	}{
		{"default", "", 1, 9, "", ""},
		{"page dan limit", "page=2&limit=5", 2, 5, "", ""},
		{"limit di atas 50 dipotong", "limit=500", 1, 50, "", ""},
		{"limit tepat 50", "limit=50", 1, 50, "", ""},
		{"limit nol ditolak", "limit=0", 0, 0, "", "limit"},
		{"limit negatif ditolak", "limit=-3", 0, 0, "", "limit"},
		{"limit bukan angka ditolak", "limit=abc", 0, 0, "", "limit"},
		{"page nol ditolak", "page=0", 0, 0, "", "page"},
		{"page bukan angka ditolak", "page=satu", 0, 0, "", "page"},
		{"page terlalu besar ditolak", "page=99999999999", 0, 0, "", "page"},
		{"q satu karakter ditolak", "q=a", 0, 0, "", "minimal 2"},
		{"q kosong dianggap tidak ada", "q=", 1, 9, "", ""},
		{"q spasi saja dianggap tidak ada", "q=%20%20", 1, 9, "", ""},
		{"q dirapikan", "q=%20%20ab%20%20%20cd%20", 1, 9, "%ab cd%", ""},
		{"q wildcard di-escape", "q=" + url.QueryEscape(`50%_a\b`), 1, 9, `%50\%\_a\\b%`, ""},
		{"q terlalu panjang ditolak", "q=" + strings.Repeat("a", 101), 0, 0, "", "maksimal 100"},
		{"q 100 karakter diterima", "q=" + strings.Repeat("a", 100), 1, 9, "%" + strings.Repeat("a", 100) + "%", ""},
		{"q karakter kontrol ditolak", "q=ab%00cd", 0, 0, "", "tidak diizinkan"},
		{"q UTF-8 rusak ditolak", "q=%ff%fe", 0, 0, "", "UTF-8"},
	}
	for _, c := range cases {
		t.Run(c.nama, func(t *testing.T) {
			v, err := url.ParseQuery(c.query)
			if err != nil {
				t.Fatal(err)
			}
			got, gerr := parseDaftar(v)
			if c.galat != "" {
				if gerr == nil || !strings.Contains(gerr.Error(), c.galat) {
					t.Fatalf("diharapkan galat memuat %q, dapat: %v", c.galat, gerr)
				}
				return
			}
			if gerr != nil {
				t.Fatalf("galat tak terduga: %v", gerr)
			}
			if got.Page != c.page || got.Limit != c.limit || got.Pola != c.pola {
				t.Fatalf("dapat %+v, diharapkan page=%d limit=%d pola=%q", got, c.page, c.limit, c.pola)
			}
		})
	}
}

func TestParamGeser(t *testing.T) {
	p := daftarQuery{Page: 3, Limit: 5}.param()
	if p.Batas != 5 || p.Geser != 10 {
		t.Fatalf("param salah: %+v", p)
	}
}

func TestSlugSah(t *testing.T) {
	ok := []string{"a", "berita-pertama", "abc-123", "2026-rencana"}
	buruk := []string{"", "Berita", "a_b", "-a", "a-", "a--b", "a b", "a/b", "../etc", "%2e", strings.Repeat("a", 201)}
	for _, s := range ok {
		if !slugSah(s) {
			t.Errorf("%q seharusnya sah", s)
		}
	}
	for _, s := range buruk {
		if slugSah(s) {
			t.Errorf("%q seharusnya tidak sah", s)
		}
	}
}

func decode(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("bukan JSON: %v\n%s", err, body)
	}
	return m
}

func TestDaftarBeritaBentukDanPagination(t *testing.T) {
	var tangkap konten.ParamDaftar
	k := kontenPalsu{
		berita: []konten.BeritaRingkas{{ID: 7, Slug: "satu", Judul: "Satu", Ringkasan: "Ringkas", GambarKey: "berita/7/a.jpg", TanggalPublikasi: "2026-11-05"}},
		total:  23, tangkap: &tangkap,
	}
	rec := panggil(routerKonten(k), http.MethodGet, "/api/berita?page=3&limit=5&q=pel")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	m := decode(t, rec.Body.String())
	if m["page"].(float64) != 3 || m["limit"].(float64) != 5 || m["total"].(float64) != 23 {
		t.Fatalf("meta salah: %v", m)
	}
	item := m["items"].([]any)[0].(map[string]any)
	want := map[string]any{"id": 7.0, "slug": "satu", "judul": "Satu", "ringkasan": "Ringkas", "gambar": "/media/berita/7/a.jpg", "tanggalPublikasi": "05-11-2026"}
	if len(item) != len(want) {
		t.Fatalf("field berlebih/kurang: %v", item)
	}
	for k, v := range want {
		if item[k] != v {
			t.Errorf("%s = %v, diharapkan %v", k, item[k], v)
		}
	}
	if tangkap.Batas != 5 || tangkap.Geser != 10 || tangkap.Pola != "%pel%" {
		t.Fatalf("param ke store salah: %+v", tangkap)
	}
	for _, terlarang := range []string{"status", "version", "createdBy", "updatedBy", "deletedAt", "created_by"} {
		if strings.Contains(rec.Body.String(), terlarang) {
			t.Errorf("respons publik tidak boleh memuat %q", terlarang)
		}
	}
}

func TestDaftarKosongBerupaArrayBukanNull(t *testing.T) {
	for _, path := range []string{"/api/berita?page=99", "/api/dokumentasi", "/api/testimoni", "/api/tim"} {
		rec := panggil(routerKonten(kontenPalsu{}), http.MethodGet, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status %d", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"items":[]`) {
			t.Errorf("%s: items harus [] bukan null: %s", path, rec.Body)
		}
	}
}

func TestHalamanDiLuarRentangTetap200DenganTotal(t *testing.T) {
	rec := panggil(routerKonten(kontenPalsu{total: 4}), http.MethodGet, "/api/berita?page=50")
	m := decode(t, rec.Body.String())
	if rec.Code != http.StatusOK || m["total"].(float64) != 4 || m["page"].(float64) != 50 {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body)
	}
}

func TestQueryTidakValidMenjadi400(t *testing.T) {
	for _, path := range []string{"/api/berita?page=0", "/api/berita?limit=x", "/api/berita?q=a", "/api/dokumentasi?page=-1", "/api/dokumentasi?q=a"} {
		rec := panggil(routerKonten(kontenPalsu{}), http.MethodGet, path)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s status %d", path, rec.Code)
		}
		var g bentukGalat
		if err := json.Unmarshal(rec.Body.Bytes(), &g); err != nil || g.Error.Code != kodeQueryTidakValid || g.Error.Message == "" {
			t.Errorf("%s: bentuk galat salah: %s", path, rec.Body)
		}
	}
}

func TestDetailBerita(t *testing.T) {
	k := kontenPalsu{beritaDetail: konten.BeritaDetail{
		BeritaRingkas: konten.BeritaRingkas{ID: 1, Slug: "ada", Judul: "Ada", Ringkasan: "R", GambarKey: "berita/1/x.png", TanggalPublikasi: "2026-01-02"},
		Isi:           "Isi lengkap",
	}}
	rec := panggil(routerKonten(k), http.MethodGet, "/api/berita/ada")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	m := decode(t, rec.Body.String())
	if m["isi"] != "Isi lengkap" || m["gambar"] != "/media/berita/1/x.png" || m["slug"] != "ada" {
		t.Fatalf("isi salah: %v", m)
	}
}

func TestDetail404UntukTidakAdaDanSlugTidakSah(t *testing.T) {
	r := routerKonten(kontenPalsu{})
	for _, path := range []string{"/api/berita/tidak-ada", "/api/berita/Besar", "/api/berita/a_b", "/api/dokumentasi/tidak-ada", "/api/dokumentasi/..%2Fx"} {
		rec := panggil(r, http.MethodGet, path)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s status %d, diharapkan 404", path, rec.Code)
			continue
		}
		var g bentukGalat
		if err := json.Unmarshal(rec.Body.Bytes(), &g); err != nil || g.Error.Code != kodeTidakDitemukan {
			t.Errorf("%s: bentuk galat salah: %s", path, rec.Body)
		}
	}
}

func TestGalatStoreMenjadi500TanpaBocor(t *testing.T) {
	r := routerKonten(kontenPalsu{err: errors.New("pq: koneksi putus password=rahasia")})
	for _, path := range []string{"/api/berita", "/api/berita/ada", "/api/dokumentasi", "/api/dokumentasi/ada", "/api/testimoni", "/api/tim"} {
		rec := panggil(r, http.MethodGet, path)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("%s status %d", path, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "rahasia") || strings.Contains(rec.Body.String(), "koneksi") {
			t.Errorf("%s: detail internal bocor: %s", path, rec.Body)
		}
	}
}

func TestDokumentasi(t *testing.T) {
	k := kontenPalsu{
		dokum: []konten.DokumentasiRingkas{
			{ID: 1, Slug: "a", Judul: "A", TanggalPublikasi: "2026-02-03", GambarUtamaKey: "dokumentasi/1/h.jpg", JumlahGambar: 2},
			{ID: 2, Slug: "b", Judul: "B", TanggalPublikasi: "2026-02-02"},
		},
		total: 2,
		dokumDetail: konten.DokumentasiDetail{
			ID: 1, Slug: "a", Judul: "A", TanggalPublikasi: "2026-02-03",
			Gambar: []konten.GambarDokumentasi{{Key: "dokumentasi/1/h.jpg"}, {Key: "dokumentasi/1/2.jpg"}},
		},
	}
	r := routerKonten(k)

	m := decode(t, panggil(r, http.MethodGet, "/api/dokumentasi").Body.String())
	items := m["items"].([]any)
	if items[0].(map[string]any)["gambarUtama"] != "/media/dokumentasi/1/h.jpg" {
		t.Errorf("gambarUtama salah: %v", items[0])
	}
	// jumlahGambar selalu hadir (0 bila tanpa gambar) agar frontend dapat menampilkan "N Foto"
	if items[0].(map[string]any)["jumlahGambar"] != 2.0 || items[1].(map[string]any)["jumlahGambar"] != 0.0 {
		t.Errorf("jumlahGambar salah: %v %v", items[0], items[1])
	}
	if _, ada := items[1].(map[string]any)["gambarUtama"]; ada {
		t.Errorf("gambarUtama kosong harus dihilangkan: %v", items[1])
	}

	rec := panggil(r, http.MethodGet, "/api/dokumentasi/a")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	gambar := decode(t, rec.Body.String())["gambar"].([]any)
	g0, g1 := gambar[0].(map[string]any), gambar[1].(map[string]any)
	if len(gambar) != 2 || g0["url"] != "/media/dokumentasi/1/h.jpg" {
		t.Errorf("gambar[0] salah: %v", g0)
	}
	if _, ada := g1["alt"]; ada || g1["url"] != "/media/dokumentasi/1/2.jpg" {
		t.Errorf("gambar[1] salah (field alt sudah dihapus): %v", g1)
	}
}

func TestTestimoniDanTim(t *testing.T) {
	k := kontenPalsu{
		testimoni: []konten.Testimoni{{ID: 1, Nama: "Ani", Teks: "Bagus", Posisi: "Dosen", FotoKey: "testimoni/1/f.jpg"}, {ID: 2, Nama: "Budi", Teks: "Oke", Posisi: "Mhs"}},
		tim:       []konten.AnggotaTim{{ID: 3, Nama: "Cici", Posisi: "Ketua", BidangRiset: "NLP", FotoKey: "tim/3/f.jpg"}},
	}
	r := routerKonten(k)

	ts := decode(t, panggil(r, http.MethodGet, "/api/testimoni").Body.String())["items"].([]any)
	if ts[0].(map[string]any)["foto"] != "/media/testimoni/1/f.jpg" {
		t.Errorf("foto salah: %v", ts[0])
	}
	if _, ada := ts[1].(map[string]any)["foto"]; ada {
		t.Errorf("foto kosong harus dihilangkan: %v", ts[1])
	}

	rec := panggil(r, http.MethodGet, "/api/tim")
	m := decode(t, rec.Body.String())
	a := m["items"].([]any)[0].(map[string]any)
	if a["bidangRiset"] != "NLP" || a["foto"] != "/media/tim/3/f.jpg" || a["nama"] != "Cici" {
		t.Errorf("anggota tim salah: %v", a)
	}
	if _, ada := m["total"]; ada {
		t.Errorf("tim tanpa pagination, tidak boleh ada total: %v", m)
	}
	if strings.Contains(rec.Body.String(), "urutan") || strings.Contains(rec.Body.String(), "version") {
		t.Errorf("urutan/version tidak boleh bocor ke publik: %s", rec.Body)
	}
}
