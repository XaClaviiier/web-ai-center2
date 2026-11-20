package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

// e2eAdmin membungkus router asli + PostgreSQL tes + RustFS tes untuk tes ujung ke ujung konten admin.
type e2eAdmin struct {
	t        *testing.T
	ctx      context.Context
	e        lingkunganAuth
	s3       *media.S3Store
	id, csrf string
	dibuat   []string // key yang dibuat tes ini; dibersihkan di akhir (hanya key milik tes)
}

func siapkanE2EAdmin(t *testing.T) *e2eAdmin {
	t.Helper()
	url := testutil.DatabaseTes(t)
	k := testutil.RustFSTes(t)
	ctx, batal := context.WithTimeout(context.Background(), 300*time.Second)
	t.Cleanup(batal)

	st, err := store.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(st.Close)
	testutil.BersihkanKonten(t, st.Pool)

	s3 := media.NewS3(media.S3Config{Endpoint: k.Endpoint, AccessKey: k.AccessKey, SecretKey: k.SecretKey, Bucket: k.Bucket})
	if err := s3.PastikanBucket(ctx, true); err != nil {
		t.Fatal(err)
	}
	a := &e2eAdmin{t: t, ctx: ctx, s3: s3}
	t.Cleanup(func() {
		for _, key := range a.dibuat {
			_ = s3.Hapus(context.Background(), key)
		}
	})

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	a.e = siapkanAuthDenganMedia(t, func(d *DepsAuth) { d.AdminPassword2 = passDev2 }, s3, func(d *Deps) {
		d.DB, d.Konten = st, st
		d.Admin = DepsAdmin{
			Berita:      admin.BeritaLayanan{Repo: st.AdminBerita(), Media: s3, Log: log},
			Dokumentasi: admin.DokumentasiLayanan{Repo: st.AdminDokumentasi(), Media: s3, Log: log},
			Testimoni:   admin.TestimoniLayanan{Repo: st.AdminTestimoni(), Media: s3, Log: log},
			Tim:         admin.TimLayanan{Repo: st.AdminTim(), Media: s3, Log: log},
			Log:         admin.LogLayanan{Repo: st.AdminLog()},
		}
	})
	a.id, a.csrf = loginBerhasil(t, a.e)
	return a
}

func (a *e2eAdmin) panggil(metode, path, body string) *httptest.ResponseRecorder {
	return kirimAdmin(a.t, a.e.h, a.id, a.csrf, metode, path, body)
}

func (a *e2eAdmin) publik(path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	a.e.h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func (a *e2eAdmin) unggah(sumber string) string {
	a.t.Helper()
	body, ct := bangunMultipart(a.t, map[string]string{"sumber": sumber}, bagianFile{"gambar", "f.png", "image/png", pngUji(a.t)})
	rec := unggah(a.t, a.e, a.id, a.csrf, body, ct)
	if rec.Code != http.StatusCreated {
		a.t.Fatalf("upload %s status %d: %s", sumber, rec.Code, rec.Body)
	}
	var h hasilUpload
	_ = json.Unmarshal(rec.Body.Bytes(), &h)
	a.dibuat = append(a.dibuat, h.Key)
	return h.Key
}

func (a *e2eAdmin) ada(key string) bool {
	a.t.Helper()
	ada, err := media.Ada(a.ctx, a.s3, key)
	if err != nil {
		a.t.Fatal(err)
	}
	return ada
}

func (a *e2eAdmin) harus(rec *httptest.ResponseRecorder, kode int, langkah string) map[string]any {
	a.t.Helper()
	if rec.Code != kode {
		a.t.Fatalf("%s: status %d (diharapkan %d): %s", langkah, rec.Code, kode, rec.Body)
	}
	if rec.Body.Len() == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		a.t.Fatalf("%s: bukan JSON: %s", langkah, rec.Body)
	}
	return m
}

func namaDari(items []any) string {
	var out []string
	for _, i := range items {
		out = append(out, i.(map[string]any)["nama"].(string))
	}
	return strings.Join(out, ",")
}

func TestE2EAdminTimSiklusPenuh(t *testing.T) {
	a := siapkanE2EAdmin(t)
	badan := func(nama, foto string, urutan int) string {
		return `{"nama":"` + nama + `","posisi":"Peneliti","bidangRiset":"NLP","foto":"` + foto + `","urutan":` + itoaID(int64(urutan)) + `}`
	}

	// 1. buat dua anggota dengan urutan pilihan admin, langsung tampil di publik (tim tanpa status)
	foto1, foto2 := a.unggah("tim"), a.unggah("tim")
	t1 := a.harus(a.panggil("POST", "/api/admin/tim", badan("Ani", foto1, 1)), 201, "buat Ani")
	t2 := a.harus(a.panggil("POST", "/api/admin/tim", badan("Budi", foto2, 2)), 201, "buat Budi")
	if t1["urutan"].(float64) != 1 || t2["urutan"].(float64) != 2 || t1["version"].(float64) != 1 || t1["dibuatOleh"] != "dev-admin" {
		t.Fatalf("urutan/version/audit salah: %v %v", t1, t2)
	}
	pub := a.harus(a.publik("/api/tim"), 200, "publik")["items"].([]any)
	if namaDari(pub) != "Ani,Budi" || pub[0].(map[string]any)["foto"] != "/media/"+foto1 || pub[0].(map[string]any)["bidangRiset"] != "NLP" {
		t.Fatalf("publik tim salah: %v", pub)
	}
	pathAni := "/api/admin/tim/" + itoaID(int64(t1["id"].(float64)))
	pathBudi := "/api/admin/tim/" + itoaID(int64(t2["id"].(float64)))

	// 2. foto yang sama dipakai anggota lain ditolak, dan foto milik Ani TIDAK terhapus
	a.harus(a.panggil("POST", "/api/admin/tim", badan("Kembar", foto1, 3)), 400, "foto kembar")
	if !a.ada(foto1) {
		t.Fatal("foto milik anggota lain terhapus oleh request yang gagal")
	}

	// 3. Simpan version basi: 409 dan file baru dihapus kembali
	fotoBaru := a.unggah("tim")
	e409 := a.harus(a.panggil("PUT", pathAni, `{"nama":"Ani Lestari","posisi":"Ketua","bidangRiset":"NLP","foto":"`+fotoBaru+`","urutan":1,"version":99}`), 409, "version basi")
	if e409["error"].(map[string]any)["code"] != "conflict" {
		t.Errorf("kode salah: %v", e409)
	}
	if a.ada(fotoBaru) {
		t.Error("file baru yang gagal disimpan harus dihapus kembali")
	}

	// 4. Simpan benar dengan foto baru: foto lama dihapus setelah commit, urutan tetap
	fotoBaru = a.unggah("tim")
	u := a.harus(a.panggil("PUT", pathAni, `{"nama":"Ani Lestari","posisi":"Ketua","bidangRiset":"NLP","foto":"`+fotoBaru+`","urutan":1,"version":1}`), 200, "simpan")
	if u["nama"] != "Ani Lestari" || u["version"].(float64) != 2 || u["urutan"].(float64) != 1 || u["diubahOleh"] != "dev-admin" {
		t.Fatalf("hasil simpan: %v", u)
	}
	if a.ada(foto1) || !a.ada(fotoBaru) {
		t.Error("foto lama harus terhapus dan foto baru tersimpan")
	}

	// 5. Tim tidak punya status/trash: tidak ada rute post/restore/permanen
	for _, r := range []struct{ m, p string }{{"POST", pathAni + "/post"}, {"POST", pathAni + "/restore"}, {"DELETE", pathAni + "/permanen"}} {
		if rec := a.panggil(r.m, r.p, `{"version":2}`); rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s harus tidak ada: %d", r.m, r.p, rec.Code)
		}
	}

	// 6. Hapus langsung permanen: baris dan file hilang, anggota lain utuh
	if rec := a.panggil("DELETE", pathAni, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("hapus: %d %s", rec.Code, rec.Body)
	}
	if a.ada(fotoBaru) {
		t.Error("foto harus terhapus bersama anggota")
	}
	a.harus(a.panggil("GET", pathAni, ""), 404, "detail setelah hapus")
	a.harus(a.panggil("DELETE", pathAni, ""), 404, "hapus dua kali")
	pub = a.harus(a.publik("/api/tim"), 200, "publik setelah hapus")["items"].([]any)
	if namaDari(pub) != "Budi" || !a.ada(foto2) {
		t.Errorf("anggota lain harus utuh: %v", pub)
	}
	adm := a.harus(a.panggil("GET", "/api/admin/tim", ""), 200, "daftar admin")["items"].([]any)
	if namaDari(adm) != "Budi" {
		t.Errorf("daftar admin: %v", adm)
	}

	// 7. urutan dipilih admin: Cici (urutan 1) tampil sebelum Budi (urutan 2); Simpan boleh mengubah urutan
	t3 := a.harus(a.panggil("POST", "/api/admin/tim", badan("Cici", a.unggah("tim"), 1)), 201, "buat Cici")
	if t3["urutan"].(float64) != 1 {
		t.Errorf("urutan harus mengikuti klien (1): %v", t3["urutan"])
	}
	pub = a.harus(a.publik("/api/tim"), 200, "publik urutan")["items"].([]any)
	if namaDari(pub) != "Cici,Budi" {
		t.Errorf("urutan publik harus Cici,Budi: %s", namaDari(pub))
	}
	pathCici := "/api/admin/tim/" + itoaID(int64(t3["id"].(float64)))
	cici := a.harus(a.panggil("PUT", pathCici, `{"nama":"Cici","posisi":"Peneliti","bidangRiset":"NLP","foto":"`+t3["foto"].(map[string]any)["key"].(string)+`","urutan":9,"version":1}`), 200, "ubah urutan Cici")
	if cici["urutan"].(float64) != 9 {
		t.Errorf("urutan baru 9: %v", cici["urutan"])
	}
	pub = a.harus(a.publik("/api/tim"), 200, "publik setelah ubah urutan")["items"].([]any)
	if namaDari(pub) != "Budi,Cici" {
		t.Errorf("setelah urutan Cici=9 harus Budi,Cici: %s", namaDari(pub))
	}
	// urutan wajib dan harus bilangan bulat >= 1
	foto := a.unggah("tim")
	for nama, body := range map[string]string{
		"tanpa urutan": `{"nama":"X","posisi":"p","bidangRiset":"b","foto":"` + foto + `"}`,
		"urutan nol":   badan("X", foto, 0),
		"urutan minus": badan("X", foto, -4),
	} {
		a.harus(a.panggil("POST", "/api/admin/tim", body), 400, nama)
	}
	_ = pathBudi
}

func TestE2EAdminTestimoniSiklusPenuh(t *testing.T) {
	a := siapkanE2EAdmin(t)
	dasar := `"nama":"Ani","teks":"Pelatihannya sangat membantu.\r\n\r\n\r\nTerima kasih.","posisi":"Dosen"`

	// 1. buat tanpa foto (null) dan dengan foto
	tanpa := a.harus(a.panggil("POST", "/api/admin/testimoni", `{`+dasar+`,"foto":null}`), 201, "buat tanpa foto")
	foto1 := a.unggah("testimoni")
	dengan := a.harus(a.panggil("POST", "/api/admin/testimoni", `{"nama":"Budi","teks":"Bagus","posisi":"Mhs","foto":"`+foto1+`"}`), 201, "buat dengan foto")
	if tanpa["status"] != "draft" || tanpa["version"].(float64) != 1 || tanpa["teks"] != "Pelatihannya sangat membantu.\n\nTerima kasih." {
		t.Fatalf("tanpa foto: %v", tanpa)
	}
	if _, ada := tanpa["foto"]; ada {
		t.Errorf("tanpa foto: field foto harus dihilangkan: %v", tanpa)
	}
	idTanpa, idDengan := int64(tanpa["id"].(float64)), int64(dengan["id"].(float64))
	pathTanpa := "/api/admin/testimoni/" + itoaID(idTanpa)
	pathDengan := "/api/admin/testimoni/" + itoaID(idDengan)

	// 2. foto wajib hadir; foto kembar ditolak tanpa menghapus milik orang lain
	a.harus(a.panggil("POST", "/api/admin/testimoni", `{`+dasar+`}`), 400, "tanpa field foto")
	a.harus(a.panggil("POST", "/api/admin/testimoni", `{`+dasar+`,"foto":"`+foto1+`"}`), 400, "foto kembar")
	if !a.ada(foto1) {
		t.Fatal("foto milik testimoni lain terhapus oleh request yang gagal")
	}

	// 3. draft tidak tampil di publik; Post tanpa foto tidak butuh storage; Post dengan foto tampil dengan url
	if pub := a.harus(a.publik("/api/testimoni"), 200, "publik kosong")["items"].([]any); len(pub) != 0 {
		t.Fatalf("draft tidak boleh tampil di publik: %v", pub)
	}
	a.harus(a.panggil("POST", pathTanpa+"/post", `{"version":1}`), 200, "post tanpa foto")
	p := a.harus(a.panggil("POST", pathDengan+"/post", `{"version":1}`), 200, "post dengan foto")
	if p["status"] != "posted" || p["version"].(float64) != 2 {
		t.Fatalf("post: %v", p)
	}
	pub := a.harus(a.publik("/api/testimoni"), 200, "publik")["items"].([]any)
	if namaDari(pub) != "Budi,Ani" { // terbaru dibuat dulu
		t.Fatalf("urutan publik: %s", namaDari(pub))
	}
	if pub[0].(map[string]any)["foto"] != "/media/"+foto1 {
		t.Errorf("foto publik: %v", pub[0])
	}
	if _, ada := pub[1].(map[string]any)["foto"]; ada {
		t.Errorf("tanpa foto: publik harus menghilangkan foto: %v", pub[1])
	}
	if rec := a.publik("/media/" + foto1); rec.Code != 200 {
		t.Errorf("foto lewat /media: %d", rec.Code)
	}

	// 4. aturan status: posted tidak bisa diedit atau dihapus
	a.harus(a.panggil("PUT", pathDengan, `{"nama":"x","teks":"y","posisi":"z","foto":"`+foto1+`","version":2}`), 400, "edit posted")
	a.harus(a.panggil("DELETE", pathDengan, ""), 409, "hapus posted")

	// 5. Retract lalu ubah foto: ganti, hapus, tambah
	a.harus(a.panggil("POST", pathDengan+"/retract", `{"version":2}`), 200, "retract")
	foto2 := a.unggah("testimoni")
	g := a.harus(a.panggil("PUT", pathDengan, `{"nama":"Budi","teks":"Bagus","posisi":"Mhs","foto":"`+foto2+`","version":3}`), 200, "ganti foto")
	if g["version"].(float64) != 4 || a.ada(foto1) || !a.ada(foto2) {
		t.Fatalf("ganti foto: %v lama=%v baru=%v", g, a.ada(foto1), a.ada(foto2))
	}
	h := a.harus(a.panggil("PUT", pathDengan, `{"nama":"Budi","teks":"Bagus","posisi":"Mhs","foto":null,"version":4}`), 200, "hapus foto")
	if _, ada := h["foto"]; ada || a.ada(foto2) {
		t.Fatalf("hapus foto: %v file masih ada=%v", h, a.ada(foto2))
	}
	foto3 := a.unggah("testimoni")
	a.harus(a.panggil("PUT", pathDengan, `{"nama":"Budi","teks":"Bagus","posisi":"Mhs","foto":"`+foto3+`","version":5}`), 200, "tambah foto")
	if !a.ada(foto3) {
		t.Fatal("foto yang ditambahkan harus tersimpan")
	}
	// version basi: 409 dan file baru dihapus kembali
	foto4 := a.unggah("testimoni")
	a.harus(a.panggil("PUT", pathDengan, `{"nama":"Budi","teks":"Bagus","posisi":"Mhs","foto":"`+foto4+`","version":1}`), 409, "version basi")
	if a.ada(foto4) || !a.ada(foto3) {
		t.Error("file baru dihapus kembali; foto lama tidak tersentuh")
	}

	// 6. trash, daftar admin (tanpa pagination), pulihkan, hapus permanen
	a.harus(a.panggil("DELETE", pathDengan, ""), 204, "hapus")
	if !a.ada(foto3) {
		t.Fatal("soft delete tidak boleh menyentuh foto")
	}
	trash := a.harus(a.panggil("GET", "/api/admin/testimoni?status=trash", ""), 200, "trash")
	if len(trash) != 1 || namaDari(trash["items"].([]any)) != "Budi" || trash["items"].([]any)[0].(map[string]any)["dihapusPada"] == nil {
		t.Errorf("trash: %v", trash)
	}
	semua := a.harus(a.panggil("GET", "/api/admin/testimoni", ""), 200, "semua")["items"].([]any)
	if namaDari(semua) != "Ani" {
		t.Errorf("semua (tanpa trash): %s", namaDari(semua))
	}
	pl := a.harus(a.panggil("POST", pathDengan+"/restore", ""), 200, "pulihkan")
	if pl["status"] != "draft" || pl["dihapusPada"] != nil {
		t.Fatalf("pulihkan: %v", pl)
	}
	a.harus(a.panggil("DELETE", pathDengan+"/permanen", ""), 409, "permanen non-trash")
	a.harus(a.panggil("DELETE", pathDengan, ""), 204, "hapus lagi")
	a.harus(a.panggil("DELETE", pathDengan+"/permanen", ""), 204, "permanen")
	if a.ada(foto3) {
		t.Error("hapus permanen harus menghapus foto")
	}
	a.harus(a.panggil("GET", pathDengan, ""), 404, "detail setelah permanen")

	// 7. testimoni tanpa foto: retract, hapus, permanen tanpa operasi file
	a.harus(a.panggil("POST", pathTanpa+"/retract", `{"version":2}`), 200, "retract tanpa foto")
	a.harus(a.panggil("DELETE", pathTanpa, ""), 204, "hapus tanpa foto")
	a.harus(a.panggil("DELETE", pathTanpa+"/permanen", ""), 204, "permanen tanpa foto")
	if pub := a.harus(a.publik("/api/testimoni"), 200, "publik akhir")["items"].([]any); len(pub) != 0 {
		t.Errorf("publik harus kosong: %v", pub)
	}
}
