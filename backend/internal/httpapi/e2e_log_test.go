package httpapi

import (
	"fmt"
	"strings"
	"testing"
)

// Log aktivitas end-to-end: HTTP admin -> database nyata -> trigger -> GET /api/admin/log-aktivitas.
func TestE2ELogAktivitas(t *testing.T) {
	a := siapkanE2EAdmin(t)
	badan := func(nama, foto string, urutan int) string {
		return `{"nama":"` + nama + `","posisi":"Peneliti","bidangRiset":"NLP","foto":"` + foto + `","urutan":` + itoaID(int64(urutan)) + `}`
	}
	baca := func(q string) map[string]any {
		return a.harus(a.panggil("GET", "/api/admin/log-aktivitas"+q, ""), 200, "log "+q)
	}
	ringkas := func(m map[string]any) string {
		var out []string
		for _, x := range m["items"].([]any) {
			e := x.(map[string]any)
			out = append(out, fmt.Sprintf("%s/%s/%s:%s@%s", e["jenis"], e["aksi"], e["modul"], e["labelItem"], e["username"]))
		}
		return strings.Join(out, " | ")
	}

	if m := baca(""); m["total"].(float64) != 0 || len(m["usernames"].([]any)) != 0 {
		t.Fatalf("log awal harus kosong: %v", m)
	}

	// Akun A: tambah testimoni, Simpan, Post; akun B: Retract dan hapus
	foto := a.unggah("tim")
	tm := a.harus(a.panggil("POST", "/api/admin/tim", badan("Ani", foto, 1)), 201, "buat tim")
	pathTim := "/api/admin/tim/" + itoaID(int64(tm["id"].(float64)))
	tes := a.harus(a.panggil("POST", "/api/admin/testimoni", `{"nama":"Budi","teks":"Bagus","posisi":"Mhs","foto":null}`), 201, "buat testimoni")
	pathT := "/api/admin/testimoni/" + itoaID(int64(tes["id"].(float64)))
	a.harus(a.panggil("POST", pathT+"/post", `{"version":1}`), 200, "post")

	// gagal (version basi, status salah) tidak tercatat
	a.harus(a.panggil("PUT", pathT, `{"nama":"X","teks":"y","posisi":"z","foto":null,"version":1}`), 400, "edit posted")
	a.harus(a.panggil("POST", pathT+"/post", `{"version":99}`), 409, "version basi")

	// akun B (login lain) memakai cookie sendiri
	b := loginDengan(t, a.e.h, passDev2)
	a.id, a.csrf = b.id, b.csrf
	a.harus(a.panggil("POST", pathT+"/retract", `{"version":2}`), 200, "retract oleh B")
	a.harus(a.panggil("DELETE", pathT, ""), 204, "hapus ke trash oleh B")
	a.harus(a.panggil("DELETE", pathTim, ""), 204, "hapus tim oleh B")

	m := baca("")
	if m["total"].(float64) != 6 || m["limit"].(float64) != 10 {
		t.Fatalf("meta: %v", m)
	}
	// terbaru dulu
	want := strings.Join([]string{
		"hapus/hapus_permanen/tim:Ani@admin.dev2",
		"hapus/hapus/testimoni:Budi@admin.dev2",
		"ubah/retract/testimoni:Budi@admin.dev2",
		"ubah/post/testimoni:Budi@admin.dev",
		"tambah/tambah/testimoni:Budi@admin.dev",
		"tambah/tambah/tim:Ani@admin.dev",
	}, " | ")
	if got := ringkas(m); got != want {
		t.Errorf("log salah:\n got: %s\nmau: %s", got, want)
	}
	if u := m["usernames"].([]any); len(u) != 2 || u[0] != "admin.dev" || u[1] != "admin.dev2" {
		t.Errorf("usernames: %v", u)
	}
	e0 := m["items"].([]any)[0].(map[string]any)
	if e0["penggunaSub"] != "dev-admin-2" || e0["itemId"].(float64) != tm["id"].(float64) {
		t.Errorf("sub/itemId: %v", e0)
	}
	// payload tidak tersimpan: hanya field metadata
	for _, k := range []string{"isi", "teks", "foto", "password", "csrfToken"} {
		if _, ada := e0[k]; ada {
			t.Errorf("log tidak boleh memuat %s", k)
		}
	}

	// filter
	if got := ringkas(baca("?username=admin.dev")); got != "ubah/post/testimoni:Budi@admin.dev | tambah/tambah/testimoni:Budi@admin.dev | tambah/tambah/tim:Ani@admin.dev" {
		t.Errorf("filter username: %s", got)
	}
	if baca("?jenis=hapus")["total"].(float64) != 2 || baca("?jenis=tambah")["total"].(float64) != 2 || baca("?jenis=ubah")["total"].(float64) != 2 {
		t.Error("filter jenis salah")
	}
	if baca("?dari=2999-01-01")["total"].(float64) != 0 {
		t.Error("filter dari di masa depan harus kosong")
	}
	if p := baca("?limit=4&page=9"); p["page"].(float64) != 2 || len(p["items"].([]any)) != 2 {
		t.Errorf("halaman melewati akhir dibatasi ke terakhir: %v", p)
	}
	a.harus(a.panggil("GET", "/api/admin/log-aktivitas?jenis=post", ""), 400, "jenis salah")
	a.harus(a.panggil("GET", "/api/admin/log-aktivitas?dari=2026-11-12&sampai=2026-11-10", ""), 400, "rentang terbalik")

	// log admin tidak bisa dibaca publik dan tidak ada rute tulis
	if rec := a.publik("/api/admin/log-aktivitas"); rec.Code != 401 {
		t.Errorf("publik: %d", rec.Code)
	}
	a.harus(a.panggil("POST", "/api/admin/log-aktivitas", `{}`), 405, "tulis log")
}
