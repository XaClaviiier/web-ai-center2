package httpapi

import (
	"net/http"
	"strings"
	"testing"
)

func keysDariGambar(t *testing.T, m map[string]any) string {
	t.Helper()
	var out []string
	for _, g := range m["gambar"].([]any) {
		out = append(out, g.(map[string]any)["key"].(string))
	}
	return strings.Join(out, ",")
}

// daftarGambar membuat JSON array gambar [{"key":..}] dari daftar key.
func daftarGambar(keys ...string) string {
	var out []string
	for _, key := range keys {
		out = append(out, `{"key":"`+key+`"}`)
	}
	return `[` + strings.Join(out, ",") + `]`
}

func TestE2EAdminDokumentasiSiklusPenuh(t *testing.T) {
	a := siapkanE2EAdmin(t)
	badan := func(judul, gambar string) string {
		return `{"judul":"` + judul + `","tanggalPublikasi":"09-11-2026","gambarDaftar":` + gambar + `}`
	}

	// 1. buat dengan tiga gambar
	k := []string{a.unggah("dokumentasi"), a.unggah("dokumentasi"), a.unggah("dokumentasi"), a.unggah("dokumentasi")}
	d := a.harus(a.panggil("POST", "/api/admin/dokumentasi", badan("Pelatihan AI", daftarGambar(k[0], k[1], k[2]))), 201, "buat")
	if d["slug"] != "pelatihan-ai" || d["status"] != "draft" || d["version"].(float64) != 1 || d["dibuatOleh"] != "dev-admin" || keysDariGambar(t, d) != strings.Join(k[:3], ",") {
		t.Fatalf("hasil buat salah: %v", d)
	}
	if _, ada := d["gambar"].([]any)[0].(map[string]any)["alt"]; ada {
		t.Errorf("field alt sudah dihapus: %v", d["gambar"])
	}
	a.harus(a.panggil("POST", "/api/admin/dokumentasi", badan("Dengan Alt", `[{"key":"`+k[3]+`","alt":"x"}]`)), 400, "alt tidak dikenal lagi")
	path := "/api/admin/dokumentasi/" + itoaID(int64(d["id"].(float64)))

	// 2. draft tidak tampil di publik; judul sama -> slug diberi akhiran
	if rec := a.publik("/api/dokumentasi/pelatihan-ai"); rec.Code != 404 {
		t.Fatalf("draft di publik harus 404, dapat %d", rec.Code)
	}
	d2 := a.harus(a.panggil("POST", "/api/admin/dokumentasi", badan("Pelatihan AI", daftarGambar(k[3]))), 201, "buat kedua")
	if d2["slug"] != "pelatihan-ai-2" {
		t.Errorf("slug bentrok harus pelatihan-ai-2: %v", d2["slug"])
	}
	path2 := "/api/admin/dokumentasi/" + itoaID(int64(d2["id"].(float64)))

	// 3. validasi daftar gambar: kosong, tujuh, kembar, dan key milik dokumentasi lain
	a.harus(a.panggil("POST", "/api/admin/dokumentasi", badan("Kosong", `[]`)), 400, "0 gambar")
	var tujuh []string
	for i := 0; i < 7; i++ {
		tujuh = append(tujuh, a.unggah("dokumentasi"))
	}
	a.harus(a.panggil("POST", "/api/admin/dokumentasi", badan("Tujuh", daftarGambar(tujuh...))), 400, "7 gambar")
	for _, key := range tujuh {
		if a.ada(key) {
			t.Errorf("file baru yang gagal disimpan harus dihapus kembali: %s", key)
		}
	}
	a.harus(a.panggil("POST", "/api/admin/dokumentasi", badan("Kembar", daftarGambar(k[0], k[0]))), 400, "key kembar")
	a.harus(a.panggil("POST", "/api/admin/dokumentasi", badan("Curian", daftarGambar(k[1]))), 400, "key milik dokumentasi lain")
	for _, key := range k {
		if !a.ada(key) {
			t.Fatalf("file milik dokumentasi lain terhapus oleh request yang gagal: %s", key)
		}
	}

	// 4. Simpan version basi: 409, file baru dihapus kembali, data tidak berubah
	n1 := a.unggah("dokumentasi")
	a.harus(a.panggil("PUT", path, `{"judul":"Baru","tanggalPublikasi":"10-11-2026","gambarDaftar":`+daftarGambar(k[0], n1)+`,"version":99}`), 409, "version basi")
	if a.ada(n1) {
		t.Error("file baru yang gagal disimpan harus dihapus kembali")
	}
	if cur := a.harus(a.panggil("GET", path, ""), 200, "detail"); cur["judul"] != "Pelatihan AI" || keysDariGambar(t, cur) != strings.Join(k[:3], ",") || cur["version"].(float64) != 1 {
		t.Fatalf("data berubah padahal gagal: %v", cur)
	}

	// 5. ganti urutan: tidak ada file yang dihapus
	u := a.harus(a.panggil("PUT", path, `{"judul":"Pelatihan AI","tanggalPublikasi":"09-11-2026","gambarDaftar":`+daftarGambar(k[2], k[0], k[1])+`,"version":1}`), 200, "urutan")
	if keysDariGambar(t, u) != k[2]+","+k[0]+","+k[1] || u["version"].(float64) != 2 {
		t.Fatalf("urutan: %v", u)
	}
	for _, key := range k[:3] {
		if !a.ada(key) {
			t.Fatalf("mengganti urutan tidak boleh menghapus file: %s", key)
		}
	}

	// 6. replace (k1 -> n2 di posisi tengah), hapus (k0), tambah (n3 di akhir) dalam SATU simpan
	n2, n3 := a.unggah("dokumentasi"), a.unggah("dokumentasi")
	u = a.harus(a.panggil("PUT", path, `{"judul":"Pelatihan AI","tanggalPublikasi":"09-11-2026","gambarDaftar":`+daftarGambar(k[2], n2, n3)+`,"version":2}`), 200, "replace+hapus+tambah")
	if keysDariGambar(t, u) != k[2]+","+n2+","+n3 || u["version"].(float64) != 3 {
		t.Fatalf("hasil: %v", u)
	}
	if a.ada(k[0]) || a.ada(k[1]) {
		t.Error("gambar yang dibuang (k0 dan k1) harus dihapus setelah commit")
	}
	if !a.ada(k[2]) || !a.ada(n2) || !a.ada(n3) {
		t.Error("gambar yang dipertahankan dan yang baru harus ada")
	}
	if !a.ada(k[3]) {
		t.Error("gambar dokumentasi lain tidak boleh tersentuh")
	}

	// 7. Post memeriksa SEMUA gambar: satu hilang -> 409 gambar_hilang; setelah dipulihkan -> berhasil
	if err := a.s3.Hapus(a.ctx, n3); err != nil {
		t.Fatal(err)
	}
	e409 := a.harus(a.panggil("POST", path+"/post", `{"version":3}`), 409, "post gambar hilang")
	if e409["error"].(map[string]any)["code"] != "gambar_hilang" {
		t.Errorf("kode salah: %v", e409)
	}
	if err := a.s3.Put(a.ctx, n3, pngUji(t), "image/png"); err != nil {
		t.Fatal(err)
	}
	p := a.harus(a.panggil("POST", path+"/post", `{"version":3}`), 200, "post")
	if p["status"] != "posted" || p["version"].(float64) != 4 {
		t.Fatalf("post: %v", p)
	}

	// 8. publik: detail memuat semua gambar berurutan, headline di daftar = gambar pertama, gambar lewat /media
	pub := a.harus(a.publik("/api/dokumentasi/pelatihan-ai"), 200, "publik detail")
	var urls []string
	for _, g := range pub["gambar"].([]any) {
		urls = append(urls, g.(map[string]any)["url"].(string))
	}
	if strings.Join(urls, ",") != "/media/"+k[2]+",/media/"+n2+",/media/"+n3 {
		t.Errorf("gambar publik: %v", urls)
	}
	list := a.harus(a.publik("/api/dokumentasi"), 200, "publik daftar")
	if list["total"].(float64) != 1 || list["items"].([]any)[0].(map[string]any)["gambarUtama"] != "/media/"+k[2] {
		t.Errorf("headline publik harus gambar urutan 1: %v", list)
	}
	for _, u := range urls {
		if rec := a.publik(u); rec.Code != 200 {
			t.Errorf("%s: %d", u, rec.Code)
		}
	}

	// 9. aturan status
	a.harus(a.panggil("PUT", path, `{"judul":"x","tanggalPublikasi":"09-11-2026","gambarDaftar":`+daftarGambar(k[2])+`,"version":4}`), 400, "edit posted")
	a.harus(a.panggil("DELETE", path, ""), 409, "hapus posted")
	r := a.harus(a.panggil("POST", path+"/retract", `{"version":4}`), 200, "retract")
	if r["status"] != "draft" || r["version"].(float64) != 5 || len(r["gambar"].([]any)) != 3 {
		t.Fatalf("retract: %v", r)
	}
	if rec := a.publik("/api/dokumentasi/pelatihan-ai"); rec.Code != 404 {
		t.Errorf("setelah retract harus 404 di publik: %d", rec.Code)
	}

	// 10. daftar admin: gambarUtama dan jumlahGambar; trash tidak menyentuh gambar
	semua := a.harus(a.panggil("GET", "/api/admin/dokumentasi", ""), 200, "daftar")
	first := semua["items"].([]any)
	var pertama map[string]any
	for _, it := range first {
		if it.(map[string]any)["slug"] == "pelatihan-ai" {
			pertama = it.(map[string]any)
		}
	}
	if pertama == nil || pertama["jumlahGambar"].(float64) != 3 || pertama["gambarUtama"].(map[string]any)["key"] != k[2] {
		t.Errorf("daftar admin: %v", pertama)
	}
	a.harus(a.panggil("DELETE", path, ""), 204, "hapus")
	for _, key := range []string{k[2], n2, n3} {
		if !a.ada(key) {
			t.Fatalf("soft delete tidak boleh menyentuh gambar: %s", key)
		}
	}
	trash := a.harus(a.panggil("GET", "/api/admin/dokumentasi?status=trash", ""), 200, "trash")
	if trash["total"].(float64) != 1 || trash["items"].([]any)[0].(map[string]any)["dihapusPada"] == nil {
		t.Errorf("trash: %v", trash)
	}
	pl := a.harus(a.panggil("POST", path+"/restore", ""), 200, "pulihkan")
	if pl["status"] != "draft" || pl["dihapusPada"] != nil || len(pl["gambar"].([]any)) != 3 {
		t.Fatalf("pulihkan: %v", pl)
	}

	// 11. hapus permanen menghapus baris dan SEMUA file gambarnya; dokumentasi lain utuh
	a.harus(a.panggil("DELETE", path+"/permanen", ""), 409, "permanen non-trash")
	a.harus(a.panggil("DELETE", path, ""), 204, "hapus lagi")
	a.harus(a.panggil("DELETE", path+"/permanen", ""), 204, "permanen")
	for _, key := range []string{k[2], n2, n3} {
		if a.ada(key) {
			t.Errorf("hapus permanen harus menghapus SEMUA file: %s masih ada", key)
		}
	}
	a.harus(a.panggil("GET", path, ""), 404, "detail setelah permanen")
	if rec := a.publik("/media/" + n2); rec.Code != http.StatusNotFound {
		t.Errorf("/media setelah dihapus harus 404: %d", rec.Code)
	}
	if kedua := a.harus(a.panggil("GET", path2, ""), 200, "dokumentasi kedua"); keysDariGambar(t, kedua) != k[3] || !a.ada(k[3]) {
		t.Errorf("dokumentasi kedua harus utuh: %v", kedua)
	}
}
