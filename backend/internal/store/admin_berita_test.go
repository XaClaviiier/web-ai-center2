package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
)

func inputBerita(judul, gambar string) admin.InputBerita {
	return admin.InputBerita{Judul: judul, Isi: "Paragraf satu.\n\nParagraf dua.", TanggalPublikasi: "2026-11-09", Gambar: gambar}
}

func buatDraft(t *testing.T, s *Store, ctx context.Context, slug, judul string) admin.Berita {
	t.Helper()
	b, dibuat, err := s.AdminBerita().Buat(ctx, slug, inputBerita(judul, "berita/"+slug+".png"), "dev-admin")
	if err != nil || !dibuat {
		t.Fatalf("Buat %s: dibuat=%v err=%v", slug, dibuat, err)
	}
	return b
}

func TestAdminBuatDanDetail(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	b := buatDraft(t, s, ctx, "judul-satu", "Judul Satu")
	if b.Status != "draft" || b.Version != 1 || b.DibuatOleh != "dev-admin" || b.DiubahOleh != "dev-admin" || b.DihapusPada != nil {
		t.Errorf("nilai awal salah: %+v", b)
	}
	if b.Ringkasan != "Paragraf satu." || b.TanggalPublikasi != "2026-11-09" || b.DibuatPada.IsZero() || b.DibuatPada.Location().String() != "UTC" {
		t.Errorf("kolom turunan/tanggal salah: %+v", b)
	}
	d, err := s.AdminBerita().Detail(ctx, b.ID)
	if err != nil || d.Isi != "Paragraf satu.\n\nParagraf dua." || d.GambarKey != "berita/judul-satu.png" {
		t.Fatalf("Detail: %+v %v", d, err)
	}
	if _, err := s.AdminBerita().Detail(ctx, 99999); !errors.Is(err, admin.ErrTidakDitemukan) {
		t.Errorf("Detail tak ada: %v", err)
	}
}

func TestAdminSlugBentrokTidakMenggagalkan(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	buatDraft(t, s, ctx, "sama", "Sama")
	_, dibuat, err := s.AdminBerita().Buat(ctx, "sama", inputBerita("Lain", "berita/lain.png"), "u")
	if err != nil || dibuat {
		t.Fatalf("slug bentrok harus dibuat=false tanpa galat: dibuat=%v err=%v", dibuat, err)
	}
	// koneksi tetap sehat dan baris tidak tercipta
	if _, total, _ := s.AdminBerita().Daftar(ctx, admin.ParamDaftar{Filter: "semua", Batas: 10}); total != 1 {
		t.Errorf("total harus 1, dapat %d", total)
	}
	// slug berikutnya dapat dipakai
	if _, dibuat, err := s.AdminBerita().Buat(ctx, "sama-2", inputBerita("Lain", "berita/lain.png"), "u"); err != nil || !dibuat {
		t.Errorf("sama-2: dibuat=%v err=%v", dibuat, err)
	}
}

func TestAdminGambarUnikDiDatabase(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	a := buatDraft(t, s, ctx, "a", "A")
	_, _, err := s.AdminBerita().Buat(ctx, "b", inputBerita("B", a.GambarKey), "u")
	if !errors.Is(err, admin.ErrGambarDipakai) {
		t.Fatalf("Buat dengan gambar yang sama harus ErrGambarDipakai, dapat %v", err)
	}
	b := buatDraft(t, s, ctx, "b", "B")
	if _, err := s.AdminBerita().Ubah(ctx, b.ID, b.Version, inputBerita("B", a.GambarKey), "u"); !errors.Is(err, admin.ErrGambarDipakai) {
		t.Fatalf("Ubah dengan gambar milik berita lain harus ErrGambarDipakai, dapat %v", err)
	}
	// memakai gambar sendiri lagi tidak melanggar
	if _, err := s.AdminBerita().Ubah(ctx, b.ID, b.Version, inputBerita("B diedit", b.GambarKey), "u"); err != nil {
		t.Errorf("gambar sendiri harus boleh: %v", err)
	}
	if dipakai, err := s.AdminBerita().GambarDipakai(ctx, a.GambarKey); err != nil || !dipakai {
		t.Errorf("GambarDipakai(a) = %v %v", dipakai, err)
	}
	if dipakai, _ := s.AdminBerita().GambarDipakai(ctx, "berita/tidak-dipakai.png"); dipakai {
		t.Error("key bebas tidak boleh dilaporkan dipakai")
	}
}

func TestAdminUbahAturanVersionDanStatus(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	b := buatDraft(t, s, ctx, "u", "Awal")

	u, err := s.AdminBerita().Ubah(ctx, b.ID, 1, inputBerita("Baru", b.GambarKey), "editor")
	if err != nil {
		t.Fatal(err)
	}
	if u.Judul != "Baru" || u.Version != 2 || u.DiubahOleh != "editor" || u.DibuatOleh != "dev-admin" || u.Slug != "u" {
		t.Errorf("hasil Ubah salah: %+v", u)
	}
	if _, err := s.AdminBerita().Ubah(ctx, b.ID, 1, inputBerita("Basi", b.GambarKey), "x"); !errors.Is(err, admin.ErrVersiBentrok) {
		t.Errorf("version basi: %v", err)
	}
	if _, err := s.AdminBerita().Ubah(ctx, 99999, 1, inputBerita("x", "berita/x.png"), "x"); !errors.Is(err, admin.ErrTidakDitemukan) {
		t.Errorf("tak ada: %v", err)
	}
	p, err := s.AdminBerita().Post(ctx, b.ID, 2, "editor")
	if err != nil || p.Status != "posted" || p.Version != 3 {
		t.Fatalf("Post: %+v %v", p, err)
	}
	if _, err := s.AdminBerita().Ubah(ctx, b.ID, 3, inputBerita("x", b.GambarKey), "x"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("edit posted harus ditolak: %v", err)
	}
	// state salah didahulukan daripada version salah
	if _, err := s.AdminBerita().Ubah(ctx, b.ID, 1, inputBerita("x", b.GambarKey), "x"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("status salah harus didahulukan: %v", err)
	}
}

func TestAdminPostRetractHapusPulihkanPermanen(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	b := buatDraft(t, s, ctx, "siklus", "Siklus")

	if _, err := s.AdminBerita().Retract(ctx, b.ID, 1, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("retract draft: %v", err)
	}
	if _, err := s.AdminBerita().Post(ctx, b.ID, 9, "u"); !errors.Is(err, admin.ErrVersiBentrok) {
		t.Errorf("post version salah: %v", err)
	}
	p, _ := s.AdminBerita().Post(ctx, b.ID, 1, "u")
	if _, err := s.AdminBerita().Post(ctx, b.ID, 2, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("post dua kali: %v", err)
	}
	if err := s.AdminBerita().Hapus(ctx, b.ID, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("hapus posted: %v", err)
	}
	if _, err := s.AdminBerita().Retract(ctx, b.ID, p.Version-1, "u"); !errors.Is(err, admin.ErrVersiBentrok) {
		t.Errorf("retract version basi: %v", err)
	}
	r, err := s.AdminBerita().Retract(ctx, b.ID, p.Version, "u")
	if err != nil || r.Status != "draft" || r.Version != 3 {
		t.Fatalf("Retract: %+v %v", r, err)
	}

	// Hapus: tidak menaikkan version (database.md §4), mengisi dihapusPada, mengubah diubahOleh
	if err := s.AdminBerita().Hapus(ctx, b.ID, "penghapus"); err != nil {
		t.Fatal(err)
	}
	d, _ := s.AdminBerita().Detail(ctx, b.ID)
	if d.DihapusPada == nil || d.Version != 3 || d.DiubahOleh != "penghapus" || d.Status != "draft" {
		t.Errorf("setelah Hapus: %+v", d)
	}
	if err := s.AdminBerita().Hapus(ctx, b.ID, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("hapus dua kali: %v", err)
	}
	if _, err := s.AdminBerita().Post(ctx, b.ID, 3, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("post item trash: %v", err)
	}

	pl, err := s.AdminBerita().Pulihkan(ctx, b.ID, "pemulih")
	if err != nil || pl.DihapusPada != nil || pl.Status != "draft" || pl.Version != 3 || pl.DiubahOleh != "pemulih" {
		t.Fatalf("Pulihkan: %+v %v", pl, err)
	}
	if _, err := s.AdminBerita().Pulihkan(ctx, b.ID, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("pulihkan yang tidak di-trash: %v", err)
	}
	if _, err := s.AdminBerita().HapusPermanen(ctx, b.ID); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Errorf("hapus permanen non-trash: %v", err)
	}
	if _, err := s.AdminBerita().Detail(ctx, b.ID); err != nil {
		t.Fatal("penolakan tidak boleh menghapus baris")
	}

	_ = s.AdminBerita().Hapus(ctx, b.ID, "u")
	key, err := s.AdminBerita().HapusPermanen(ctx, b.ID)
	if err != nil || key != "berita/siklus.png" {
		t.Fatalf("HapusPermanen: %q %v", key, err)
	}
	if _, err := s.AdminBerita().Detail(ctx, b.ID); !errors.Is(err, admin.ErrTidakDitemukan) {
		t.Errorf("baris harus hilang: %v", err)
	}
	for _, f := range []func() error{
		func() error { return s.AdminBerita().Hapus(ctx, b.ID, "u") },
		func() error { _, e := s.AdminBerita().Pulihkan(ctx, b.ID, "u"); return e },
		func() error { _, e := s.AdminBerita().HapusPermanen(ctx, b.ID); return e },
		func() error { _, e := s.AdminBerita().Post(ctx, b.ID, 1, "u"); return e },
		func() error { _, e := s.AdminBerita().Retract(ctx, b.ID, 1, "u"); return e },
	} {
		if err := f(); !errors.Is(err, admin.ErrTidakDitemukan) {
			t.Errorf("aksi pada baris yang sudah hilang harus ErrTidakDitemukan: %v", err)
		}
	}
}

func TestAdminPostMenampilkanDiPublikDanRetractMenyembunyikan(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	b := buatDraft(t, s, ctx, "terbit", "Terbit")

	if _, err := s.BeritaBySlug(ctx, "terbit"); err == nil {
		t.Fatal("draft tidak boleh tampil di publik")
	}
	p, _ := s.AdminBerita().Post(ctx, b.ID, 1, "u")
	if _, err := s.BeritaBySlug(ctx, "terbit"); err != nil {
		t.Fatalf("posted harus tampil di publik: %v", err)
	}
	if _, err := s.AdminBerita().Retract(ctx, b.ID, p.Version, "u"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.BeritaBySlug(ctx, "terbit"); err == nil {
		t.Fatal("setelah Retract tidak boleh tampil lagi di publik")
	}
}

func TestAdminDaftarFilterUrutanDanPencarian(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	buatDraft(t, s, ctx, "a", "Alpha Pelatihan")
	b := buatDraft(t, s, ctx, "b", "Beta Seminar")
	c := buatDraft(t, s, ctx, "c", "Gamma Pelatihan")
	buatDraft(t, s, ctx, "d", "Delta 50%_ Diskon")
	if _, err := s.AdminBerita().Post(ctx, b.ID, 1, "u"); err != nil {
		t.Fatal(err)
	}
	if err := s.AdminBerita().Hapus(ctx, c.ID, "u"); err != nil {
		t.Fatal(err)
	}

	judul := func(f, pola string, batas int32, geser int64) (string, int64) {
		items, total, err := s.AdminBerita().Daftar(ctx, admin.ParamDaftar{Filter: f, Pola: pola, Batas: batas, Geser: geser})
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, i := range items {
			out = append(out, i.Slug)
		}
		return strings.Join(out, ","), total
	}

	// semua = draft + posted tanpa trash, terbaru dibuat dulu
	if got, total := judul("semua", "", 20, 0); got != "d,b,a" || total != 3 {
		t.Errorf("semua: %s total=%d", got, total)
	}
	if got, total := judul("draft", "", 20, 0); got != "d,a" || total != 2 {
		t.Errorf("draft: %s total=%d", got, total)
	}
	if got, total := judul("posted", "", 20, 0); got != "b" || total != 1 {
		t.Errorf("posted: %s total=%d", got, total)
	}
	if got, total := judul("trash", "", 20, 0); got != "c" || total != 1 {
		t.Errorf("trash: %s total=%d", got, total)
	}
	// pagination
	if got, total := judul("semua", "", 2, 2); got != "a" || total != 3 {
		t.Errorf("halaman 2: %s total=%d", got, total)
	}
	if got, total := judul("semua", "", 2, 100); got != "" || total != 3 {
		t.Errorf("di luar rentang: %q total=%d", got, total)
	}
	// pencarian: tak peka huruf besar, hormati filter, wildcard di-escape
	if got, total := judul("semua", "%pelatihan%", 20, 0); got != "a" || total != 1 {
		t.Errorf("cari 'pelatihan' di semua (c di-trash tidak ikut): %s total=%d", got, total)
	}
	if got, total := judul("trash", "%PELATIHAN%", 20, 0); got != "c" || total != 1 {
		t.Errorf("cari di trash: %s total=%d", got, total)
	}
	if got, total := judul("semua", `%50\%\_%`, 20, 0); got != "d" || total != 1 {
		t.Errorf("wildcard literal: %s total=%d", got, total)
	}
}

func TestAdminTrashDiurutkanDariYangTerakhirDihapus(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	var ids []int64
	for i := 0; i < 3; i++ {
		b := buatDraft(t, s, ctx, fmt.Sprintf("t%d", i), fmt.Sprintf("T%d", i))
		ids = append(ids, b.ID)
	}
	// hapus urutan terbalik dari pembuatan: t1, t2, t0
	for _, i := range []int{1, 2, 0} {
		if err := s.AdminBerita().Hapus(ctx, ids[i], "u"); err != nil {
			t.Fatal(err)
		}
	}
	items, _, err := s.AdminBerita().Daftar(ctx, admin.ParamDaftar{Filter: "trash", Batas: 10})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, i := range items {
		got = append(got, i.Slug)
	}
	if strings.Join(got, ",") != "t0,t2,t1" {
		t.Errorf("trash harus dari yang terakhir dihapus: %v", got)
	}
}
