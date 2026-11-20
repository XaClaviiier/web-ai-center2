package store

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
)

func inputTim(nama, foto string) admin.InputTim {
	return admin.InputTim{Nama: nama, Posisi: "Peneliti", BidangRiset: "NLP", Foto: foto, Urutan: 1}
}

func buatTim(t *testing.T, s *Store, ctx context.Context, nama string) admin.Tim {
	t.Helper()
	tm, err := s.AdminTim().Buat(ctx, inputTim(nama, "tim/"+strings.ToLower(nama)+".png"), "dev-admin")
	if err != nil {
		t.Fatalf("Buat %s: %v", nama, err)
	}
	return tm
}

func namaTim(items []admin.Tim) string {
	var out []string
	for _, i := range items {
		out = append(out, i.Nama)
	}
	return strings.Join(out, ",")
}

// buatTimUrutan membuat anggota dengan urutan pilihan admin.
func buatTimUrutan(t *testing.T, s *Store, ctx context.Context, nama string, urutan int32) admin.Tim {
	t.Helper()
	in := inputTim(nama, "tim/"+strings.ToLower(nama)+".png")
	in.Urutan = urutan
	tm, err := s.AdminTim().Buat(ctx, in, "dev-admin")
	if err != nil {
		t.Fatalf("Buat %s: %v", nama, err)
	}
	return tm
}

func TestAdminTimUrutanDariKlien(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	a := buatTimUrutan(t, s, ctx, "Ani", 3)
	b := buatTimUrutan(t, s, ctx, "Budi", 1)
	c := buatTimUrutan(t, s, ctx, "Cici", 2)
	if a.Urutan != 3 || b.Urutan != 1 || c.Urutan != 2 {
		t.Fatalf("urutan harus mengikuti klien (3,1,2): %d %d %d", a.Urutan, b.Urutan, c.Urutan)
	}
	if items, _ := s.AdminTim().Daftar(ctx); namaTim(items) != "Budi,Cici,Ani" {
		t.Fatalf("daftar mengikuti urutan: %s", namaTim(items))
	}
	// Simpan boleh mengubah urutan
	in := inputTim("Ani", a.FotoKey)
	in.Urutan = 1
	u, err := s.AdminTim().Ubah(ctx, a.ID, a.Version, in, "editor")
	if err != nil || u.Urutan != 1 {
		t.Fatalf("ubah urutan: %+v %v", u, err)
	}
	// urutan sama dengan Budi (1): pemutus seri id, Budi dibuat lebih dulu? Ani lebih kecil id, jadi lebih dulu
	if items, _ := s.AdminTim().Daftar(ctx); namaTim(items) != "Ani,Budi,Cici" {
		t.Fatalf("setelah ubah urutan: %s", namaTim(items))
	}
	if a.Version != 1 || a.DibuatOleh != "dev-admin" || a.DiubahOleh != "dev-admin" || a.FotoKey != "tim/ani.png" || a.DibuatPada.IsZero() {
		t.Errorf("nilai awal salah: %+v", a)
	}
	if _, err := s.AdminTim().Hapus(ctx, b.ID); err != nil {
		t.Fatal(err)
	}
	if items, _ := s.AdminTim().Daftar(ctx); namaTim(items) != "Ani,Cici" {
		t.Errorf("daftar setelah hapus: %s", namaTim(items))
	}
}

func TestAdminTimDaftarPemutusSeriId(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	// dua anggota dengan urutan sama (mis. pembuatan serentak): pemutus seri id
	for _, nama := range []string{"Pertama", "Kedua"} {
		if _, err := s.Pool.Exec(ctx, `INSERT INTO tim (nama, posisi, bidang_riset, foto, urutan, created_by, updated_by) VALUES ($1::text, 'p', 'b', 'tim/' || $1::text || '.png', 5, 'u', 'u')`, nama); err != nil {
			t.Fatal(err)
		}
	}
	items, err := s.AdminTim().Daftar(ctx)
	if err != nil || namaTim(items) != "Pertama,Kedua" {
		t.Fatalf("pemutus seri id: %s %v", namaTim(items), err)
	}
}

func TestAdminTimDaftarKosongBukanNil(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	items, err := s.AdminTim().Daftar(ctx)
	if err != nil || items == nil || len(items) != 0 {
		t.Errorf("harus slice kosong non-nil: %v %v", items, err)
	}
}

func TestAdminTimFotoUnik(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	a := buatTim(t, s, ctx, "Ani")
	if _, err := s.AdminTim().Buat(ctx, inputTim("Lain", a.FotoKey), "u"); !errors.Is(err, admin.ErrGambarDipakai) {
		t.Fatalf("Buat dengan foto yang sama: %v", err)
	}
	b := buatTim(t, s, ctx, "Budi")
	if _, err := s.AdminTim().Ubah(ctx, b.ID, 1, inputTim("Budi", a.FotoKey), "u"); !errors.Is(err, admin.ErrGambarDipakai) {
		t.Fatalf("Ubah ke foto milik anggota lain: %v", err)
	}
	if _, err := s.AdminTim().Ubah(ctx, b.ID, 1, inputTim("Budi X", b.FotoKey), "u"); err != nil {
		t.Errorf("foto sendiri harus boleh: %v", err)
	}
	if dipakai, err := s.AdminTim().FotoDipakai(ctx, a.FotoKey); err != nil || !dipakai {
		t.Errorf("FotoDipakai(a) = %v %v", dipakai, err)
	}
	if dipakai, _ := s.AdminTim().FotoDipakai(ctx, "tim/bebas.png"); dipakai {
		t.Error("key bebas tidak boleh dilaporkan dipakai")
	}
}

func TestAdminTimUbahVersionDanKetidakadaan(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	a := buatTim(t, s, ctx, "Ani")
	buatTim(t, s, ctx, "Budi")

	u, err := s.AdminTim().Ubah(ctx, a.ID, 1, inputTim("Ani Lestari", "tim/ani-baru.png"), "editor")
	if err != nil {
		t.Fatal(err)
	}
	if u.Nama != "Ani Lestari" || u.Version != 2 || u.DiubahOleh != "editor" || u.DibuatOleh != "dev-admin" || u.Urutan != a.Urutan || u.FotoKey != "tim/ani-baru.png" {
		t.Errorf("hasil Ubah salah: %+v", u)
	}
	if _, err := s.AdminTim().Ubah(ctx, a.ID, 1, inputTim("Basi", "tim/x.png"), "u"); !errors.Is(err, admin.ErrVersiBentrok) {
		t.Errorf("version basi: %v", err)
	}
	if _, err := s.AdminTim().Ubah(ctx, 99999, 1, inputTim("x", "tim/y.png"), "u"); !errors.Is(err, admin.ErrTidakDitemukan) {
		t.Errorf("tidak ada: %v", err)
	}
	// tim tidak punya aturan status: tetap bisa diubah berulang selama version benar
	if _, err := s.AdminTim().Ubah(ctx, a.ID, 2, inputTim("Ani Lagi", "tim/ani-baru.png"), "u"); err != nil {
		t.Errorf("ubah berulang: %v", err)
	}
	d, _ := s.AdminTim().Detail(ctx, a.ID)
	if d.Version != 3 || d.Nama != "Ani Lagi" {
		t.Errorf("detail: %+v", d)
	}
}

func TestAdminTimHapusLangsungPermanen(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	a := buatTim(t, s, ctx, "Ani")
	b := buatTim(t, s, ctx, "Budi")

	foto, err := s.AdminTim().Hapus(ctx, a.ID)
	if err != nil || foto != "tim/ani.png" {
		t.Fatalf("Hapus: %q %v", foto, err)
	}
	if _, err := s.AdminTim().Detail(ctx, a.ID); !errors.Is(err, admin.ErrTidakDitemukan) {
		t.Errorf("baris harus hilang: %v", err)
	}
	if _, err := s.AdminTim().Hapus(ctx, a.ID); !errors.Is(err, admin.ErrTidakDitemukan) {
		t.Errorf("hapus dua kali: %v", err)
	}
	if _, err := s.AdminTim().Detail(ctx, b.ID); err != nil {
		t.Errorf("anggota lain harus utuh: %v", err)
	}
	// tidak ada trash: tidak ada kolom status/deleted_at pada tabel tim
	var jumlah int
	if err := s.Pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.columns WHERE table_name='tim' AND column_name IN ('status','deleted_at')`).Scan(&jumlah); err != nil || jumlah != 0 {
		t.Errorf("tabel tim tidak boleh punya status/deleted_at: %d %v", jumlah, err)
	}
}

func TestAdminTimTampilLangsungDiPublik(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	buatTim(t, s, ctx, "Ani")
	pub, err := s.ListTim(ctx)
	if err != nil || len(pub) != 1 || pub[0].Nama != "Ani" || pub[0].FotoKey != "tim/ani.png" {
		t.Fatalf("publik: %+v %v", pub, err)
	}
}
