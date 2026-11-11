package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
)

// kunci membuat key dokumentasi sederhana untuk tes (database tidak memeriksa bentuk key).
func kunci(nama string) string { return "dokumentasi/" + nama + ".png" }

func gambarDok(nama ...string) []admin.GambarDokumentasi {
	out := make([]admin.GambarDokumentasi, 0, len(nama))
	for _, n := range nama {
		out = append(out, admin.GambarDokumentasi{Key: kunci(n)})
	}
	return out
}

func inputDok(judul string, g []admin.GambarDokumentasi) admin.InputDokumentasi {
	return admin.InputDokumentasi{Judul: judul, TanggalPublikasi: "2026-11-09", Gambar: g}
}

func buatDok(t *testing.T, s *Store, ctx context.Context, slug, judul string, nama ...string) admin.Dokumentasi {
	t.Helper()
	d, dibuat, err := s.AdminDokumentasi().Buat(ctx, slug, inputDok(judul, gambarDok(nama...)), "dev-admin")
	if err != nil || !dibuat {
		t.Fatalf("Buat %s: dibuat=%v err=%v", slug, dibuat, err)
	}
	return d
}

func urutanKey(d admin.Dokumentasi) string {
	var out []string
	for _, g := range d.Gambar {
		out = append(out, strings.TrimSuffix(strings.TrimPrefix(g.Key, "dokumentasi/"), ".png"))
	}
	return strings.Join(out, ",")
}

func hitung(t *testing.T, s *Store, ctx context.Context, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := s.Pool.QueryRow(ctx, sql, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestAdminDokumentasiBuatDenganGambarBerurutan(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	in := inputDok("Kegiatan Satu", []admin.GambarDokumentasi{{Key: kunci("a"), Alt: "Foto utama"}, {Key: kunci("b")}, {Key: kunci("c"), Alt: "Penutup"}})
	d, dibuat, err := s.AdminDokumentasi().Buat(ctx, "kegiatan-satu", in, "dev-admin")
	if err != nil || !dibuat {
		t.Fatalf("Buat: %v %v", dibuat, err)
	}
	if d.Status != "draft" || d.Version != 1 || d.Slug != "kegiatan-satu" || d.DibuatOleh != "dev-admin" || d.DihapusPada != nil || d.TanggalPublikasi != "2026-11-09" {
		t.Errorf("nilai awal salah: %+v", d)
	}
	if urutanKey(d) != "a,b,c" || d.Gambar[0].Alt != "Foto utama" || d.Gambar[1].Alt != "" || d.Gambar[2].Alt != "Penutup" {
		t.Errorf("gambar salah: %+v", d.Gambar)
	}
	var urutan []int
	rows, _ := s.Pool.Query(ctx, `SELECT urutan FROM dokumentasi_gambar WHERE dokumentasi_id=$1 ORDER BY urutan`, d.ID)
	for rows.Next() {
		var u int
		_ = rows.Scan(&u)
		urutan = append(urutan, u)
	}
	if fmt.Sprint(urutan) != "[1 2 3]" {
		t.Errorf("urutan di database harus 1..3: %v", urutan)
	}
	if n := hitung(t, s, ctx, `SELECT count(*) FROM dokumentasi_gambar WHERE dokumentasi_id=$1 AND alt IS NULL`, d.ID); n != 1 {
		t.Errorf("alt kosong harus NULL: %d", n)
	}
}

func TestAdminDokumentasiBuatRollbackTotalBilaGambarKembar(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	buatDok(t, s, ctx, "pertama", "Pertama", "a", "b")

	// gambar kedua dokumentasi baru memakai key milik dokumentasi pertama: seluruh transaksi harus dibatalkan
	_, _, err := s.AdminDokumentasi().Buat(ctx, "kedua", inputDok("Kedua", gambarDok("x", "b")), "u")
	if !errors.Is(err, admin.ErrGambarDipakai) {
		t.Fatalf("diharapkan ErrGambarDipakai, dapat %v", err)
	}
	if n := hitung(t, s, ctx, `SELECT count(*) FROM dokumentasi`); n != 1 {
		t.Errorf("baris dokumentasi 'kedua' tidak boleh tersisa: %d baris", n)
	}
	if n := hitung(t, s, ctx, `SELECT count(*) FROM dokumentasi_gambar`); n != 2 {
		t.Errorf("tidak boleh ada gambar setengah jadi (gambar 'x' harus ikut dibatalkan): %d baris", n)
	}
	// koneksi dan transaksi berikutnya tetap sehat
	if _, dibuat, err := s.AdminDokumentasi().Buat(ctx, "kedua", inputDok("Kedua", gambarDok("x")), "u"); err != nil || !dibuat {
		t.Errorf("pembuatan berikutnya harus berhasil: %v %v", dibuat, err)
	}
}

func TestAdminDokumentasiSlugBentrokTidakMenyisakanGambar(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	buatDok(t, s, ctx, "sama", "Sama", "a")
	_, dibuat, err := s.AdminDokumentasi().Buat(ctx, "sama", inputDok("Lain", gambarDok("b", "c")), "u")
	if err != nil || dibuat {
		t.Fatalf("slug bentrok: dibuat=%v err=%v", dibuat, err)
	}
	if n := hitung(t, s, ctx, `SELECT count(*) FROM dokumentasi_gambar`); n != 1 {
		t.Errorf("gambar dari percobaan yang bentrok tidak boleh tersimpan: %d", n)
	}
	if _, dibuat, err := s.AdminDokumentasi().Buat(ctx, "sama-2", inputDok("Lain", gambarDok("b", "c")), "u"); err != nil || !dibuat {
		t.Errorf("sama-2: %v %v", dibuat, err)
	}
}

func TestAdminDokumentasiUbahOperasiGambar(t *testing.T) {
	kasus := []struct {
		nama   string
		daftar []admin.GambarDokumentasi
		want   string
	}{
		{"ganti urutan (rotasi)", gambarDok("c", "a", "b"), "c,a,b"},
		{"tukar dua gambar", gambarDok("b", "a", "c"), "b,a,c"},
		{"balik semua", gambarDok("c", "b", "a"), "c,b,a"},
		{"hapus yang tengah, urutan dirapikan", gambarDok("a", "c"), "a,c"},
		{"hapus headline, berikutnya menjadi headline", gambarDok("b", "c"), "b,c"},
		{"hapus semua kecuali satu", gambarDok("b"), "b"},
		{"tambah di akhir", gambarDok("a", "b", "c", "d"), "a,b,c,d"},
		{"tambah dua, tepat enam", gambarDok("a", "b", "c", "d", "e", "f"), "a,b,c,d,e,f"},
		{"replace di urutan yang sama", gambarDok("a", "x", "c"), "a,x,c"},
		{"replace headline", gambarDok("x", "b", "c"), "x,b,c"},
		{"campur: hapus, tukar, tambah", gambarDok("c", "y", "a", "z"), "c,y,a,z"},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			s, ctx := bukaStoreTes(t)
			d := buatDok(t, s, ctx, "dok", "Dok", "a", "b", "c")
			u, err := s.AdminDokumentasi().Ubah(ctx, d.ID, 1, inputDok("Dok", k.daftar), "editor")
			if err != nil {
				t.Fatal(err)
			}
			if urutanKey(u) != k.want {
				t.Fatalf("hasil Ubah: %s, diharapkan %s", urutanKey(u), k.want)
			}
			baca, _ := s.AdminDokumentasi().Detail(ctx, d.ID)
			if urutanKey(baca) != k.want || baca.Version != 2 || baca.DiubahOleh != "editor" {
				t.Errorf("Detail: %s version=%d oleh=%s", urutanKey(baca), baca.Version, baca.DiubahOleh)
			}
			if n := hitung(t, s, ctx, `SELECT count(*) FROM dokumentasi_gambar WHERE dokumentasi_id=$1`, d.ID); n != len(k.daftar) {
				t.Errorf("jumlah baris gambar %d, diharapkan %d (tidak boleh ada sisa)", n, len(k.daftar))
			}
			// urutan di database harus tepat 1..n tanpa lubang
			for i := range k.daftar {
				if hitung(t, s, ctx, `SELECT count(*) FROM dokumentasi_gambar WHERE dokumentasi_id=$1 AND urutan=$2`, d.ID, i+1) != 1 {
					t.Errorf("urutan %d harus tepat satu baris", i+1)
				}
			}
		})
	}
}

func TestAdminDokumentasiUbahAltDanJudul(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	d := buatDok(t, s, ctx, "dok", "Dok", "a", "b")
	in := admin.InputDokumentasi{Judul: "Judul Baru", TanggalPublikasi: "2026-12-31", Gambar: []admin.GambarDokumentasi{{Key: kunci("a"), Alt: "Alt baru"}, {Key: kunci("b"), Alt: ""}}}
	u, err := s.AdminDokumentasi().Ubah(ctx, d.ID, 1, in, "u")
	if err != nil {
		t.Fatal(err)
	}
	if u.Judul != "Judul Baru" || u.TanggalPublikasi != "2026-12-31" || u.Slug != "dok" || u.Gambar[0].Alt != "Alt baru" || u.Gambar[1].Alt != "" {
		t.Errorf("hasil: %+v", u)
	}
	// alt yang dikosongkan kembali menjadi NULL
	in.Gambar[0].Alt = ""
	u, _ = s.AdminDokumentasi().Ubah(ctx, d.ID, 2, in, "u")
	if hitung(t, s, ctx, `SELECT count(*) FROM dokumentasi_gambar WHERE dokumentasi_id=$1 AND alt IS NULL`, d.ID) != 2 {
		t.Error("alt kosong harus NULL")
	}
}

func TestAdminDokumentasiUbahRollbackTotal(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	a := buatDok(t, s, ctx, "a", "A", "a1", "a2")
	b := buatDok(t, s, ctx, "b", "B", "b1")

	// 1. version basi: tidak ada yang berubah
	if _, err := s.AdminDokumentasi().Ubah(ctx, a.ID, 9, inputDok("A diubah", gambarDok("a2", "a1", "baru")), "u"); !errors.Is(err, admin.ErrVersiBentrok) {
		t.Fatalf("version basi: %v", err)
	}
	cur, _ := s.AdminDokumentasi().Detail(ctx, a.ID)
	if cur.Judul != "A" || cur.Version != 1 || urutanKey(cur) != "a1,a2" {
		t.Fatalf("version basi mengubah data: %+v", cur)
	}

	// 2. gambar dokumentasi LAIN dipakai di tengah daftar: transaksi dibatalkan, termasuk perubahan judul/urutan/hapus
	_, err := s.AdminDokumentasi().Ubah(ctx, a.ID, 1, inputDok("A diubah", gambarDok("a2", "b1", "baru")), "u")
	if !errors.Is(err, admin.ErrGambarDipakai) {
		t.Fatalf("diharapkan ErrGambarDipakai, dapat %v", err)
	}
	cur, _ = s.AdminDokumentasi().Detail(ctx, a.ID)
	if cur.Judul != "A" || cur.Version != 1 || urutanKey(cur) != "a1,a2" {
		t.Fatalf("rollback tidak total: %+v", cur)
	}
	other, _ := s.AdminDokumentasi().Detail(ctx, b.ID)
	if urutanKey(other) != "b1" {
		t.Errorf("dokumentasi lain tidak boleh tersentuh: %+v", other)
	}
	// 3. setelah kegagalan, ubah yang benar tetap berhasil (transaksi tidak tertinggal)
	if _, err := s.AdminDokumentasi().Ubah(ctx, a.ID, 1, inputDok("A diubah", gambarDok("a2", "a1")), "u"); err != nil {
		t.Errorf("ubah berikutnya harus berhasil: %v", err)
	}
}

func TestAdminDokumentasiUbahAturanStatus(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	r := s.AdminDokumentasi()
	d := buatDok(t, s, ctx, "dok", "Dok", "a")

	if _, err := r.Ubah(ctx, 99999, 1, inputDok("x", gambarDok("a")), "u"); !errors.Is(err, admin.ErrTidakDitemukan) {
		t.Errorf("tidak ada: %v", err)
	}
	p, err := r.Post(ctx, d.ID, 1, "u")
	if err != nil || p.Status != "posted" || p.Version != 2 || urutanKey(p) != "a" {
		t.Fatalf("Post: %+v %v", p, err)
	}
	if _, err := r.Ubah(ctx, d.ID, 2, inputDok("x", gambarDok("a", "z")), "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("edit posted: %v", err)
	}
	// status salah didahulukan, dan tidak ada gambar yang tersisip oleh percobaan yang ditolak
	if _, err := r.Ubah(ctx, d.ID, 1, inputDok("x", gambarDok("a", "z")), "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("status salah harus didahulukan: %v", err)
	}
	if hitung(t, s, ctx, `SELECT count(*) FROM dokumentasi_gambar WHERE dokumentasi_id=$1`, d.ID) != 1 {
		t.Error("percobaan yang ditolak tidak boleh mengubah gambar")
	}
}

func TestAdminDokumentasiSiklusStatus(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	r := s.AdminDokumentasi()
	d := buatDok(t, s, ctx, "dok", "Dok", "a", "b")

	if _, err := r.Retract(ctx, d.ID, 1, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("retract draft: %v", err)
	}
	if _, err := r.Post(ctx, d.ID, 9, "u"); !errors.Is(err, admin.ErrVersiBentrok) {
		t.Errorf("post version salah: %v", err)
	}
	p, _ := r.Post(ctx, d.ID, 1, "u")
	if err := r.Hapus(ctx, d.ID, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("hapus posted: %v", err)
	}
	rt, err := r.Retract(ctx, d.ID, p.Version, "u")
	if err != nil || rt.Status != "draft" || rt.Version != 3 || len(rt.Gambar) != 2 {
		t.Fatalf("Retract: %+v %v", rt, err)
	}
	if err := r.Hapus(ctx, d.ID, "penghapus"); err != nil {
		t.Fatal(err)
	}
	got, _ := r.Detail(ctx, d.ID)
	if got.DihapusPada == nil || got.Version != 3 || got.DiubahOleh != "penghapus" || len(got.Gambar) != 2 {
		t.Errorf("soft delete (gambar utuh, version tidak naik): %+v", got)
	}
	pl, err := r.Pulihkan(ctx, d.ID, "u")
	if err != nil || pl.DihapusPada != nil || pl.Status != "draft" || len(pl.Gambar) != 2 {
		t.Fatalf("Pulihkan: %+v %v", pl, err)
	}
	if _, err := r.Pulihkan(ctx, d.ID, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("pulihkan non-trash: %v", err)
	}
}

func TestAdminDokumentasiHapusPermanenMenghapusSemuaGambar(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	r := s.AdminDokumentasi()
	a := buatDok(t, s, ctx, "a", "A", "a1", "a2", "a3")
	b := buatDok(t, s, ctx, "b", "B", "b1")

	if _, err := r.HapusPermanen(ctx, a.ID); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Fatalf("permanen non-trash: %v", err)
	}
	if hitung(t, s, ctx, `SELECT count(*) FROM dokumentasi_gambar WHERE dokumentasi_id=$1`, a.ID) != 3 {
		t.Fatal("penolakan tidak boleh menghapus gambar")
	}
	_ = r.Hapus(ctx, a.ID, "u")
	keys, err := r.HapusPermanen(ctx, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(keys, ",") != kunci("a1")+","+kunci("a2")+","+kunci("a3") {
		t.Errorf("harus mengembalikan SEMUA key berurutan: %v", keys)
	}
	if hitung(t, s, ctx, `SELECT count(*) FROM dokumentasi WHERE id=$1`, a.ID) != 0 ||
		hitung(t, s, ctx, `SELECT count(*) FROM dokumentasi_gambar WHERE dokumentasi_id=$1`, a.ID) != 0 {
		t.Error("baris dokumentasi dan gambarnya harus hilang")
	}
	if got, _ := r.Detail(ctx, b.ID); urutanKey(got) != "b1" {
		t.Error("dokumentasi lain harus utuh")
	}
	if _, err := r.HapusPermanen(ctx, a.ID); !errors.Is(err, admin.ErrTidakDitemukan) {
		t.Errorf("hapus permanen dua kali: %v", err)
	}
}

func TestAdminDokumentasiGambarDipakai(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	buatDok(t, s, ctx, "a", "A", "a1")
	if ada, err := s.AdminDokumentasi().GambarDipakai(ctx, kunci("a1")); err != nil || !ada {
		t.Errorf("a1: %v %v", ada, err)
	}
	if ada, _ := s.AdminDokumentasi().GambarDipakai(ctx, kunci("bebas")); ada {
		t.Error("key bebas tidak boleh dilaporkan dipakai")
	}
}

func TestAdminDokumentasiDaftarGambarUtamaFilterDanPencarian(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	r := s.AdminDokumentasi()
	a := buatDok(t, s, ctx, "a", "Alpha Pelatihan", "a1", "a2", "a3")
	b := buatDok(t, s, ctx, "b", "Beta Seminar", "b1")
	c := buatDok(t, s, ctx, "c", "Gamma Pelatihan", "c1", "c2")
	buatDok(t, s, ctx, "d", "Delta 50%_ Diskon", "d1")
	if _, err := r.Post(ctx, b.ID, 1, "u"); err != nil {
		t.Fatal(err)
	}
	if err := r.Hapus(ctx, c.ID, "u"); err != nil {
		t.Fatal(err)
	}

	cari := func(f, pola string, batas int32, geser int64) ([]admin.RingkasDokumentasi, int64) {
		items, total, err := r.Daftar(ctx, admin.ParamDaftar{Filter: f, Pola: pola, Batas: batas, Geser: geser})
		if err != nil {
			t.Fatal(err)
		}
		return items, total
	}
	slugs := func(items []admin.RingkasDokumentasi) string {
		var o []string
		for _, i := range items {
			o = append(o, i.Slug)
		}
		return strings.Join(o, ",")
	}

	items, total := cari("semua", "", 20, 0)
	if slugs(items) != "d,b,a" || total != 3 {
		t.Fatalf("semua: %s total=%d", slugs(items), total)
	}
	// gambar utama = urutan 1; jumlah gambar benar
	if items[2].GambarUtamaKey != kunci("a1") || items[2].JumlahGambar != 3 || items[1].JumlahGambar != 1 {
		t.Errorf("gambar utama/jumlah: %+v", items[2])
	}
	if got, tot := cari("draft", "", 20, 0); slugs(got) != "d,a" || tot != 2 {
		t.Errorf("draft: %s %d", slugs(got), tot)
	}
	if got, tot := cari("posted", "", 20, 0); slugs(got) != "b" || tot != 1 {
		t.Errorf("posted: %s %d", slugs(got), tot)
	}
	got, tot := cari("trash", "", 20, 0)
	if slugs(got) != "c" || tot != 1 || got[0].DihapusPada == nil || got[0].JumlahGambar != 2 {
		t.Errorf("trash: %s %d %+v", slugs(got), tot, got)
	}
	if got, tot := cari("semua", "", 2, 2); slugs(got) != "a" || tot != 3 {
		t.Errorf("halaman 2: %s %d", slugs(got), tot)
	}
	if got, tot := cari("semua", "", 2, 100); len(got) != 0 || tot != 3 || got == nil {
		t.Errorf("di luar rentang: %v %d", got, tot)
	}
	if got, tot := cari("semua", "%pelatihan%", 20, 0); slugs(got) != "a" || tot != 1 {
		t.Errorf("cari (trash tidak ikut): %s %d", slugs(got), tot)
	}
	if got, tot := cari("semua", `%50\%\_%`, 20, 0); slugs(got) != "d" || tot != 1 {
		t.Errorf("wildcard literal: %s %d", slugs(got), tot)
	}
	_ = a
}

func TestAdminDokumentasiTampilDiPublikDenganUrutanTerkini(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	r := s.AdminDokumentasi()
	d := buatDok(t, s, ctx, "terbit", "Terbit", "a", "b", "c")

	if _, err := s.DokumentasiBySlug(ctx, "terbit"); !errors.Is(err, konten.ErrTidakDitemukan) {
		t.Fatalf("draft tidak boleh tampil di publik: %v", err)
	}
	// ubah urutan saat masih draft, lalu Post: publik harus mengikuti urutan terkini
	u, err := r.Ubah(ctx, d.ID, 1, inputDok("Terbit", gambarDok("c", "a", "b")), "u")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Post(ctx, d.ID, u.Version, "u"); err != nil {
		t.Fatal(err)
	}
	pub, err := s.DokumentasiBySlug(ctx, "terbit")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, g := range pub.Gambar {
		got = append(got, g.Key)
	}
	if strings.Join(got, ",") != kunci("c")+","+kunci("a")+","+kunci("b") {
		t.Errorf("publik harus berurutan c,a,b: %v", got)
	}
	list, _, _ := s.ListDokumentasi(ctx, konten.ParamDaftar{Batas: 10})
	if len(list) != 1 || list[0].GambarUtamaKey != kunci("c") {
		t.Errorf("headline publik harus gambar urutan 1 (c): %+v", list)
	}
}
