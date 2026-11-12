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

// Tes ujung ke ujung seluruh siklus berita lewat HTTP: PostgreSQL tes + RustFS tes nyata, auth palsu.
func TestE2EAdminBeritaSiklusPenuh(t *testing.T) {
	url := testutil.DatabaseTes(t)
	k := testutil.RustFSTes(t)
	ctx, batal := context.WithTimeout(context.Background(), 90*time.Second)
	defer batal()

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
	var dibuat []string // semua key yang dibuat tes ini; dibersihkan di akhir (hanya key milik tes)
	t.Cleanup(func() {
		for _, key := range dibuat {
			_ = s3.Hapus(context.Background(), key)
		}
	})

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	e := siapkanAuthDenganMedia(t, nil, s3, func(d *Deps) {
		d.DB, d.Konten = st, st
		d.Admin = DepsAdmin{Berita: admin.BeritaLayanan{Repo: st.AdminBerita(), Media: s3, Log: log}}
	})
	id, csrf := loginBerhasil(t, e)

	panggil := func(metode, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(metode, path, strings.NewReader(body))
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Origin", originSah)
		req.Header.Set("X-CSRF-Token", csrf)
		req.AddCookie(&http.Cookie{Name: "session", Value: id})
		rec := httptest.NewRecorder()
		e.h.ServeHTTP(rec, req)
		return rec
	}
	publik := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		e.h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}
	unggahBaru := func() string {
		t.Helper()
		body, ct := bangunMultipart(t, map[string]string{"sumber": "berita"}, bagianFile{"gambar", "f.png", "image/png", pngUji(t)})
		rec := unggah(t, e, id, csrf, body, ct)
		if rec.Code != http.StatusCreated {
			t.Fatalf("upload status %d: %s", rec.Code, rec.Body)
		}
		var h hasilUpload
		_ = json.Unmarshal(rec.Body.Bytes(), &h)
		dibuat = append(dibuat, h.Key)
		return h.Key
	}
	adaDiStorage := func(key string) bool {
		ada, err := media.Ada(ctx, s3, key)
		if err != nil {
			t.Fatal(err)
		}
		return ada
	}
	badan := func(key, judul string) string {
		return `{"judul":"` + judul + `","isi":"Paragraf satu.\r\n\r\nParagraf dua.","tanggalPublikasi":"2026-11-09","gambar":"` + key + `"}`
	}
	baca := func(rec *httptest.ResponseRecorder) map[string]any {
		t.Helper()
		var m map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
			t.Fatalf("bukan JSON: %s", rec.Body)
		}
		return m
	}
	harus := func(rec *httptest.ResponseRecorder, kode int, langkah string) map[string]any {
		t.Helper()
		if rec.Code != kode {
			t.Fatalf("%s: status %d (diharapkan %d): %s", langkah, rec.Code, kode, rec.Body)
		}
		if rec.Body.Len() == 0 {
			return nil
		}
		return baca(rec)
	}

	// 1. buat draft
	key1 := unggahBaru()
	b := harus(panggil("POST", "/api/admin/berita", badan(key1, "Pelatihan AI")), 201, "buat")
	idB := int64(b["id"].(float64))
	path := "/api/admin/berita/" + itoaID(idB)
	if b["slug"] != "pelatihan-ai" || b["status"] != "draft" || b["version"].(float64) != 1 || b["dibuatOleh"] != "dev-admin" {
		t.Fatalf("hasil buat salah: %v", b)
	}
	if b["isi"] != "Paragraf satu.\n\nParagraf dua." || b["ringkasan"] != "Paragraf satu." {
		t.Errorf("isi/ringkasan salah (CRLF harus dinormalkan): %q / %q", b["isi"], b["ringkasan"])
	}

	// 2. draft tidak tampil di publik
	if rec := publik("/api/berita/pelatihan-ai"); rec.Code != 404 {
		t.Fatalf("draft di publik harus 404, dapat %d", rec.Code)
	}

	// 3. judul sama -> slug diberi akhiran
	key2 := unggahBaru()
	b2 := harus(panggil("POST", "/api/admin/berita", badan(key2, "Pelatihan AI")), 201, "buat kedua")
	if b2["slug"] != "pelatihan-ai-2" {
		t.Errorf("slug bentrok harus pelatihan-ai-2, dapat %v", b2["slug"])
	}

	// 4. gambar yang sama dipakai dua berita -> ditolak, dan file milik berita pertama TIDAK terhapus
	rec := panggil("POST", "/api/admin/berita", badan(key1, "Duplikat Gambar"))
	harus(rec, 400, "gambar kembar")
	if !adaDiStorage(key1) {
		t.Fatal("file milik berita lain terhapus oleh request yang gagal")
	}

	// 5. Simpan dengan version basi -> 409 dan file baru dihapus kembali (database.md §5)
	keyBaru := unggahBaru()
	rec = panggil("PUT", path, `{"judul":"Pelatihan AI Lanjut","isi":"Isi baru","tanggalPublikasi":"2026-11-10","gambar":"`+keyBaru+`","version":99}`)
	e409 := harus(rec, 409, "simpan version basi")
	if e409["error"].(map[string]any)["code"] != "version_conflict" {
		t.Errorf("kode salah: %v", e409)
	}
	if adaDiStorage(keyBaru) {
		t.Error("file baru yang gagal disimpan harus dihapus kembali")
	}
	if !adaDiStorage(key1) {
		t.Fatal("gambar lama tidak boleh tersentuh saat Simpan gagal")
	}

	// 6. Simpan dengan version benar dan gambar baru -> gambar lama dihapus setelah commit
	keyBaru = unggahBaru()
	u := harus(panggil("PUT", path, `{"judul":"Pelatihan AI Lanjut","isi":"Isi baru","tanggalPublikasi":"2026-11-10","gambar":"`+keyBaru+`","version":1}`), 200, "simpan")
	if u["judul"] != "Pelatihan AI Lanjut" || u["version"].(float64) != 2 || u["slug"] != "pelatihan-ai" || u["tanggalPublikasi"] != "2026-11-10" {
		t.Fatalf("hasil simpan salah: %v", u)
	}
	if adaDiStorage(key1) {
		t.Error("gambar lama harus dihapus setelah commit")
	}
	if !adaDiStorage(keyBaru) {
		t.Fatal("gambar baru harus ada")
	}

	// 7. Post -> tampil di publik (detail, daftar, dan gambar lewat /media)
	p := harus(panggil("POST", path+"/post", `{"version":2}`), 200, "post")
	if p["status"] != "posted" || p["version"].(float64) != 3 {
		t.Fatalf("hasil post salah: %v", p)
	}
	pub := harus(publik("/api/berita/pelatihan-ai"), 200, "publik detail")
	if pub["judul"] != "Pelatihan AI Lanjut" || pub["gambar"] != "/media/"+keyBaru {
		t.Errorf("publik salah: %v", pub)
	}
	for _, terlarang := range []string{"status", "version", "dibuatOleh", "diubahOleh"} {
		if _, ada := pub[terlarang]; ada {
			t.Errorf("respons publik tidak boleh memuat %q", terlarang)
		}
	}
	if daftar := baca(publik("/api/berita")); daftar["total"].(float64) != 1 {
		t.Errorf("publik harus memuat 1 berita posted: %v", daftar)
	}
	if rec := publik("/media/" + keyBaru); rec.Code != 200 || rec.Header().Get("Content-Type") != "image/png" {
		t.Errorf("gambar lewat /media: %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}

	// 8. aturan status: posted tidak bisa diedit/dihapus; harus Retract dulu
	harus(panggil("PUT", path, `{"judul":"x","isi":"y","tanggalPublikasi":"2026-11-10","gambar":"`+keyBaru+`","version":3}`), 409, "edit posted")
	harus(panggil("DELETE", path, ""), 409, "hapus posted")
	harus(panggil("POST", path+"/post", `{"version":3}`), 409, "post ulang")

	// 9. Retract -> hilang dari publik
	r := harus(panggil("POST", path+"/retract", `{"version":3}`), 200, "retract")
	if r["status"] != "draft" || r["version"].(float64) != 4 {
		t.Fatalf("hasil retract salah: %v", r)
	}
	if rec := publik("/api/berita/pelatihan-ai"); rec.Code != 404 {
		t.Errorf("setelah retract harus 404 di publik, dapat %d", rec.Code)
	}

	// 10. soft delete -> masuk trash, tidak di daftar biasa, gambar tidak disentuh
	harus(panggil("DELETE", path, ""), 204, "hapus")
	if !adaDiStorage(keyBaru) {
		t.Fatal("soft delete tidak boleh menghapus gambar")
	}
	trash := baca(panggil("GET", "/api/admin/berita?status=trash", ""))
	if trash["total"].(float64) != 1 || trash["items"].([]any)[0].(map[string]any)["dihapusPada"] == nil {
		t.Errorf("trash salah: %v", trash)
	}
	if semua := baca(panggil("GET", "/api/admin/berita", "")); semua["total"].(float64) != 1 {
		t.Errorf("daftar biasa tidak boleh memuat item trash: %v", semua)
	}
	harus(panggil("PUT", path, `{"judul":"x","isi":"y","tanggalPublikasi":"2026-11-10","gambar":"`+keyBaru+`","version":4}`), 409, "edit trash")

	// 11. pulihkan -> draft; hapus lagi; hapus permanen -> baris dan file hilang
	pl := harus(panggil("POST", path+"/restore", ""), 200, "pulihkan")
	if pl["status"] != "draft" || pl["dihapusPada"] != nil || pl["version"].(float64) != 4 {
		t.Fatalf("hasil pulihkan salah: %v", pl)
	}
	harus(panggil("DELETE", path+"/permanen", ""), 409, "permanen non-trash")
	if !adaDiStorage(keyBaru) {
		t.Fatal("penolakan hapus permanen tidak boleh menghapus gambar")
	}
	harus(panggil("DELETE", path, ""), 204, "hapus lagi")
	harus(panggil("DELETE", path+"/permanen", ""), 204, "hapus permanen")
	if adaDiStorage(keyBaru) {
		t.Error("hapus permanen harus menghapus file gambar")
	}
	harus(panggil("GET", path, ""), 404, "detail setelah permanen")
	if rec := publik("/media/" + keyBaru); rec.Code != 404 {
		t.Errorf("/media setelah dihapus harus 404, dapat %d", rec.Code)
	}

	// 12. berita kedua tetap utuh
	harus(panggil("GET", "/api/admin/berita/"+itoaID(int64(b2["id"].(float64))), ""), 200, "berita kedua")
	if !adaDiStorage(key2) {
		t.Error("gambar berita kedua tidak boleh tersentuh")
	}
}
