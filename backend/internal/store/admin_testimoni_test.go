package store

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
)

func inputTestimoni(nama, foto string) admin.InputTestimoni {
	return admin.InputTestimoni{Nama: nama, Teks: "Sangat membantu.\n\nTerima kasih.", Posisi: "Dosen", Foto: foto}
}

func buatTestimoni(t *testing.T, s *Store, ctx context.Context, nama, foto string) admin.Testimoni {
	t.Helper()
	tm, err := s.AdminTestimoni().Buat(ctx, inputTestimoni(nama, foto), "dev-admin")
	if err != nil {
		t.Fatalf("Buat %s: %v", nama, err)
	}
	return tm
}

func namaTestimoni(items []admin.Testimoni) string {
	var out []string
	for _, i := range items {
		out = append(out, i.Nama)
	}
	return strings.Join(out, ",")
}

func TestAdminTestimoniBuatDenganDanTanpaFoto(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	a := buatTestimoni(t, s, ctx, "Ani", "testimoni/ani.png")
	b := buatTestimoni(t, s, ctx, "Budi", "")
	c := buatTestimoni(t, s, ctx, "Cici", "") // banyak NULL foto diperbolehkan (UNIQUE mengizinkan NULL ganda)

	if a.Status != "draft" || a.Version != 1 || a.FotoKey != "testimoni/ani.png" || a.DibuatOleh != "dev-admin" || a.DihapusPada != nil {
		t.Errorf("nilai awal salah: %+v", a)
	}
	if b.FotoKey != "" || c.FotoKey != "" {
		t.Errorf("tanpa foto harus kosong: %q %q", b.FotoKey, c.FotoKey)
	}
	d, err := s.AdminTestimoni().Detail(ctx, a.ID)
	if err != nil || d.Teks != "Sangat membantu.\n\nTerima kasih." {
		t.Errorf("Detail: %+v %v", d, err)
	}
	var nulls int
	if err := s.Pool.QueryRow(ctx, `SELECT count(*) FROM testimoni WHERE foto IS NULL`).Scan(&nulls); err != nil || nulls != 2 {
		t.Errorf("tanpa foto harus tersimpan sebagai NULL, bukan string kosong: %d %v", nulls, err)
	}
}

func TestAdminTestimoniFotoUnik(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	a := buatTestimoni(t, s, ctx, "Ani", "testimoni/ani.png")
	if _, err := s.AdminTestimoni().Buat(ctx, inputTestimoni("Lain", a.FotoKey), "u"); !errors.Is(err, admin.ErrGambarDipakai) {
		t.Fatalf("Buat foto kembar: %v", err)
	}
	b := buatTestimoni(t, s, ctx, "Budi", "")
	if _, err := s.AdminTestimoni().Ubah(ctx, b.ID, 1, inputTestimoni("Budi", a.FotoKey), "u"); !errors.Is(err, admin.ErrGambarDipakai) {
		t.Fatalf("Ubah ke foto milik lain: %v", err)
	}
	if dipakai, err := s.AdminTestimoni().FotoDipakai(ctx, a.FotoKey); err != nil || !dipakai {
		t.Errorf("FotoDipakai: %v %v", dipakai, err)
	}
	// kunci kosong tidak pernah dianggap "dipakai" (NULL bukan key)
	if dipakai, err := s.AdminTestimoni().FotoDipakai(ctx, ""); err != nil || dipakai {
		t.Errorf("key kosong tidak boleh dianggap dipakai: %v %v", dipakai, err)
	}
}

func TestAdminTestimoniUbahFotoTigaKeadaan(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	a := buatTestimoni(t, s, ctx, "Ani", "")

	u, err := s.AdminTestimoni().Ubah(ctx, a.ID, 1, inputTestimoni("Ani", "testimoni/baru.png"), "editor") // tambah foto
	if err != nil || u.FotoKey != "testimoni/baru.png" || u.Version != 2 || u.DiubahOleh != "editor" {
		t.Fatalf("tambah foto: %+v %v", u, err)
	}
	u, err = s.AdminTestimoni().Ubah(ctx, a.ID, 2, inputTestimoni("Ani", ""), "u") // hapus foto -> NULL
	if err != nil || u.FotoKey != "" || u.Version != 3 {
		t.Fatalf("hapus foto: %+v %v", u, err)
	}
	var nulls int
	_ = s.Pool.QueryRow(ctx, `SELECT count(*) FROM testimoni WHERE id=$1 AND foto IS NULL`, a.ID).Scan(&nulls)
	if nulls != 1 {
		t.Error("foto yang dihapus harus NULL di database")
	}
}

func TestAdminTestimoniAturanVersionDanStatus(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	a := buatTestimoni(t, s, ctx, "Ani", "")
	r := s.AdminTestimoni()

	if _, err := r.Ubah(ctx, a.ID, 9, inputTestimoni("x", ""), "u"); !errors.Is(err, admin.ErrVersiBentrok) {
		t.Errorf("version basi: %v", err)
	}
	if _, err := r.Ubah(ctx, 99999, 1, inputTestimoni("x", ""), "u"); !errors.Is(err, admin.ErrTidakDitemukan) {
		t.Errorf("tidak ada: %v", err)
	}
	p, err := r.Post(ctx, a.ID, 1, "u")
	if err != nil || p.Status != "posted" || p.Version != 2 {
		t.Fatalf("Post: %+v %v", p, err)
	}
	if _, err := r.Ubah(ctx, a.ID, 2, inputTestimoni("x", ""), "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("edit posted: %v", err)
	}
	if _, err := r.Ubah(ctx, a.ID, 1, inputTestimoni("x", ""), "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("status salah harus didahulukan daripada version salah: %v", err)
	}
	if err := r.Hapus(ctx, a.ID, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("hapus posted: %v", err)
	}
	if _, err := r.Post(ctx, a.ID, 2, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("post dua kali: %v", err)
	}
	if _, err := r.Retract(ctx, a.ID, 1, "u"); !errors.Is(err, admin.ErrVersiBentrok) {
		t.Errorf("retract version basi: %v", err)
	}
	rt, err := r.Retract(ctx, a.ID, 2, "u")
	if err != nil || rt.Status != "draft" || rt.Version != 3 {
		t.Fatalf("Retract: %+v %v", rt, err)
	}
	if _, err := r.Retract(ctx, a.ID, 3, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("retract draft: %v", err)
	}
}

func TestAdminTestimoniTrashPulihkanPermanen(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	r := s.AdminTestimoni()
	a := buatTestimoni(t, s, ctx, "Ani", "testimoni/ani.png")
	b := buatTestimoni(t, s, ctx, "Budi", "")

	if err := r.Hapus(ctx, a.ID, "penghapus"); err != nil {
		t.Fatal(err)
	}
	d, _ := r.Detail(ctx, a.ID)
	if d.DihapusPada == nil || d.Version != 1 || d.DiubahOleh != "penghapus" {
		t.Errorf("setelah Hapus (version tidak naik): %+v", d)
	}
	if err := r.Hapus(ctx, a.ID, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("hapus dua kali: %v", err)
	}
	if _, err := r.HapusPermanen(ctx, b.ID); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("permanen non-trash: %v", err)
	}
	pl, err := r.Pulihkan(ctx, a.ID, "u")
	if err != nil || pl.DihapusPada != nil || pl.Status != "draft" {
		t.Fatalf("Pulihkan: %+v %v", pl, err)
	}
	if _, err := r.Pulihkan(ctx, a.ID, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("pulihkan non-trash: %v", err)
	}

	_ = r.Hapus(ctx, a.ID, "u")
	foto, err := r.HapusPermanen(ctx, a.ID)
	if err != nil || foto != "testimoni/ani.png" {
		t.Fatalf("HapusPermanen dengan foto: %q %v", foto, err)
	}
	_ = r.Hapus(ctx, b.ID, "u")
	foto, err = r.HapusPermanen(ctx, b.ID)
	if err != nil || foto != "" {
		t.Fatalf("HapusPermanen tanpa foto harus mengembalikan key kosong: %q %v", foto, err)
	}
	for _, f := range []func() error{
		func() error { return r.Hapus(ctx, a.ID, "u") },
		func() error { _, e := r.Pulihkan(ctx, a.ID, "u"); return e },
		func() error { _, e := r.HapusPermanen(ctx, a.ID); return e },
		func() error { _, e := r.Post(ctx, a.ID, 1, "u"); return e },
	} {
		if err := f(); !errors.Is(err, admin.ErrTidakDitemukan) {
			t.Errorf("aksi pada baris yang sudah hilang: %v", err)
		}
	}
}

func TestAdminTestimoniDaftarFilterTanpaPagination(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	r := s.AdminTestimoni()
	a := buatTestimoni(t, s, ctx, "Ani", "")
	b := buatTestimoni(t, s, ctx, "Budi", "")
	c := buatTestimoni(t, s, ctx, "Cici", "")
	d := buatTestimoni(t, s, ctx, "Dedi", "")
	if _, err := r.Post(ctx, b.ID, 1, "u"); err != nil {
		t.Fatal(err)
	}
	// hapus Cici lalu Ani: trash berurutan dari yang terakhir dihapus (Ani lebih dulu)
	for _, id := range []int64{c.ID, a.ID} {
		if err := r.Hapus(ctx, id, "u"); err != nil {
			t.Fatal(err)
		}
	}
	list := func(f string) string {
		items, err := r.Daftar(ctx, f)
		if err != nil {
			t.Fatal(err)
		}
		return namaTestimoni(items)
	}
	if got := list("semua"); got != "Dedi,Budi" {
		t.Errorf("semua (draft + posted, tanpa trash, terbaru dibuat dulu): %s", got)
	}
	if got := list("draft"); got != "Dedi" {
		t.Errorf("draft: %s", got)
	}
	if got := list("posted"); got != "Budi" {
		t.Errorf("posted: %s", got)
	}
	if got := list("trash"); got != "Ani,Cici" {
		t.Errorf("trash dari yang terakhir dihapus: %s", got)
	}
	_ = d

	// tanpa pagination: semua baris kembali sekaligus
	for i := 0; i < 30; i++ {
		buatTestimoni(t, s, ctx, "Banyak", "")
	}
	items, err := r.Daftar(ctx, "semua")
	if err != nil || len(items) != 32 {
		t.Errorf("tanpa pagination harus mengembalikan semua (32): %d %v", len(items), err)
	}
}

func TestAdminTestimoniPostMenampilkanDiPublik(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	r := s.AdminTestimoni()
	a := buatTestimoni(t, s, ctx, "Ani", "testimoni/ani.png")

	if pub, _ := s.ListTestimoni(ctx); len(pub) != 0 {
		t.Fatal("draft tidak boleh tampil di publik")
	}
	p, _ := r.Post(ctx, a.ID, 1, "u")
	pub, err := s.ListTestimoni(ctx)
	if err != nil || len(pub) != 1 || pub[0].Nama != "Ani" || pub[0].FotoKey != "testimoni/ani.png" {
		t.Fatalf("posted harus tampil: %+v %v", pub, err)
	}
	if _, err := r.Retract(ctx, a.ID, p.Version, "u"); err != nil {
		t.Fatal(err)
	}
	if pub, _ := s.ListTestimoni(ctx); len(pub) != 0 {
		t.Fatal("setelah Retract tidak boleh tampil lagi")
	}
}
