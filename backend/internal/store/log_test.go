package store

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/auth"
)

// ringkasLog membaca seluruh log (terlama dulu) sebagai "jenis/aksi/modul#id:label@username".
func ringkasLog(t *testing.T, s *Store, ctx context.Context) []string {
	t.Helper()
	rows, err := s.Pool.Query(ctx, `SELECT jenis||'/'||aksi||'/'||modul||'#'||item_id||':'||label_item||'@'||username FROM log_aktivitas ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var x string
		if err := rows.Scan(&x); err != nil {
			t.Fatal(err)
		}
		out = append(out, x)
	}
	return out
}

func samaDengan(t *testing.T, got []string, mau ...string) {
	t.Helper()
	if strings.Join(got, " | ") != strings.Join(mau, " | ") {
		t.Errorf("log salah:\n got: %v\nmau: %v", got, mau)
	}
}

func TestLogSiklusTestimoniTercatatDenganAktor(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	ctx = auth.DenganAktor(ctx, auth.Aktor{Sub: "dev-admin", Username: "admin.dev"})

	a := buatTestimoni(t, s, ctx, "Ani", "")
	u, err := s.AdminTestimoni().Ubah(ctx, a.ID, 1, inputTestimoni("Ani Lestari", ""), "dev-admin")
	if err != nil {
		t.Fatal(err)
	}
	p, _ := s.AdminTestimoni().Post(ctx, a.ID, u.Version, "dev-admin")
	r, _ := s.AdminTestimoni().Retract(ctx, a.ID, p.Version, "dev-admin")
	if r.Status != "draft" {
		t.Fatalf("retract: %+v", r)
	}
	if err := s.AdminTestimoni().Hapus(ctx, a.ID, "dev-admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AdminTestimoni().Pulihkan(ctx, a.ID, "dev-admin"); err != nil {
		t.Fatal(err)
	}
	if err := s.AdminTestimoni().Hapus(ctx, a.ID, "dev-admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AdminTestimoni().HapusPermanen(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	id := formatInt(a.ID)
	samaDengan(t, ringkasLog(t, s, ctx),
		"tambah/tambah/testimoni#"+id+":Ani@admin.dev",
		"ubah/ubah/testimoni#"+id+":Ani Lestari@admin.dev",
		"ubah/post/testimoni#"+id+":Ani Lestari@admin.dev",
		"ubah/retract/testimoni#"+id+":Ani Lestari@admin.dev",
		"hapus/hapus/testimoni#"+id+":Ani Lestari@admin.dev",
		"ubah/restore/testimoni#"+id+":Ani Lestari@admin.dev",
		"hapus/hapus/testimoni#"+id+":Ani Lestari@admin.dev",
		"hapus/hapus_permanen/testimoni#"+id+":Ani Lestari@admin.dev",
	)
	// sub tersimpan dari sesi; label tetap ada setelah item dihapus permanen
	var sub string
	if err := s.Pool.QueryRow(ctx, `SELECT DISTINCT pengguna_sub FROM log_aktivitas`).Scan(&sub); err != nil || sub != "dev-admin" {
		t.Errorf("pengguna_sub = %q %v", sub, err)
	}
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func TestLogTanpaAktorTidakTercatat(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	a := buatTestimoni(t, s, ctx, "Ani", "") // ctx tanpa aktor: aksi sistem (job purge, impor, tes)
	if err := s.AdminTestimoni().Hapus(ctx, a.ID, "x"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AdminTestimoni().HapusPermanen(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	samaDengan(t, ringkasLog(t, s, ctx))
}

func TestLogUsernameKosongMemakaiSub(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	ctx = auth.DenganAktor(ctx, auth.Aktor{Sub: "sub-lama"}) // sesi lama tanpa username
	buatTestimoni(t, s, ctx, "Budi", "")
	got := ringkasLog(t, s, ctx)
	if len(got) != 1 || !strings.HasSuffix(got[0], "@sub-lama") {
		t.Errorf("username harus jatuh ke sub: %v", got)
	}
}

func TestLogGagalTidakTercatat(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	ctx = auth.DenganAktor(ctx, auth.Aktor{Sub: "dev-admin", Username: "admin.dev"})
	a := buatTestimoni(t, s, ctx, "Ani", "")
	// version basi, tidak ada, dan status salah: tidak ada perubahan, tidak ada log baru
	if _, err := s.AdminTestimoni().Ubah(ctx, a.ID, 99, inputTestimoni("X", ""), "u"); !errors.Is(err, admin.ErrVersiBentrok) {
		t.Fatalf("version basi: %v", err)
	}
	if _, err := s.AdminTestimoni().Ubah(ctx, 99999, 1, inputTestimoni("X", ""), "u"); !errors.Is(err, admin.ErrTidakDitemukan) {
		t.Fatalf("tidak ada: %v", err)
	}
	if _, err := s.AdminTestimoni().Retract(ctx, a.ID, 1, "u"); !errors.Is(err, admin.ErrStatusTidakValid) {
		t.Fatalf("retract draft: %v", err)
	}
	if got := ringkasLog(t, s, ctx); len(got) != 1 {
		t.Errorf("hanya pembuatan yang tercatat: %v", got)
	}
}

// Perubahan dan log satu transaksi: INSERT yang gagal (foto kembar) tidak meninggalkan log.
func TestLogAtomikBersamaPerubahan(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	ctx = auth.DenganAktor(ctx, auth.Aktor{Sub: "dev-admin", Username: "admin.dev"})
	buatTestimoni(t, s, ctx, "Ani", "testimoni/sama.png")
	if _, err := s.AdminTestimoni().Buat(ctx, inputTestimoni("Kembar", "testimoni/sama.png"), "u"); !errors.Is(err, admin.ErrGambarDipakai) {
		t.Fatalf("foto kembar: %v", err)
	}
	if got := ringkasLog(t, s, ctx); len(got) != 1 {
		t.Errorf("INSERT gagal tidak boleh meninggalkan log: %v", got)
	}
}

func TestLogDokumentasiDanTimDalamTransaksiBerlapis(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	ctx = auth.DenganAktor(ctx, auth.Aktor{Sub: "dev-admin", Username: "admin.dev"})
	d := buatDok(t, s, ctx, "dok", "Dokumentasi", "a", "b") // Buat: transaksi multi-statement
	if _, err := s.AdminDokumentasi().Ubah(ctx, d.ID, 1, inputDok("Dokumentasi 2", gambarDok("b", "a")), "u"); err != nil {
		t.Fatal(err)
	}
	tm := buatTim(t, s, ctx, "Ani")
	if _, err := s.AdminTim().Hapus(ctx, tm.ID); err != nil {
		t.Fatal(err)
	}
	// dokumentasi_gambar tidak dicatat sendiri; Ubah dokumentasi tercatat satu kali
	samaDengan(t, ringkasLog(t, s, ctx),
		"tambah/tambah/dokumentasi#"+formatInt(d.ID)+":Dokumentasi@admin.dev",
		"ubah/ubah/dokumentasi#"+formatInt(d.ID)+":Dokumentasi 2@admin.dev",
		"tambah/tambah/tim#"+formatInt(tm.ID)+":Ani@admin.dev",
		"hapus/hapus_permanen/tim#"+formatInt(tm.ID)+":Ani@admin.dev",
	)
}

func TestLogLabelDipotong140(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	ctx = auth.DenganAktor(ctx, auth.Aktor{Sub: "dev-admin", Username: "admin.dev"})
	buatTestimoni(t, s, ctx, strings.Repeat("n", 300), "")
	var n int
	if err := s.Pool.QueryRow(ctx, `SELECT length(label_item) FROM log_aktivitas`).Scan(&n); err != nil || n != 140 {
		t.Errorf("label harus 140 karakter: %d %v", n, err)
	}
}

func TestLogBacaFilterUrutanDanHalaman(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	// isi langsung (waktu terkontrol): 3 baris, dua username, dua hari WIB
	_, err := s.Pool.Exec(ctx, `INSERT INTO log_aktivitas (terjadi_pada, pengguna_sub, username, jenis, aksi, modul, item_id, label_item) VALUES
		('2026-11-10 16:59:00+00', 's1', 'ani',  'tambah', 'tambah', 'berita', 1, 'A'),
		('2026-11-10 17:00:00+00', 's2', 'budi', 'ubah',   'post',   'berita', 1, 'A'),
		('2026-11-10 17:00:00+00', 's1', 'ani',  'hapus',  'hapus',  'tim',    2, 'B')`)
	if err != nil {
		t.Fatal(err)
	}
	svc := admin.LogLayanan{Repo: s.AdminLog()}
	h, err := svc.Daftar(ctx, admin.FilterLog{})
	if err != nil || h.Total != 3 || h.Limit != 10 || h.Page != 1 {
		t.Fatalf("semua: %+v %v", h, err)
	}
	// terbaru dulu; waktu sama dipecah id menurun
	if h.Items[0].Modul != "tim" || h.Items[1].Aksi != "post" || h.Items[2].Aksi != "tambah" {
		t.Errorf("urutan salah: %+v", h.Items)
	}
	if strings.Join(h.Usernames, ",") != "ani,budi" {
		t.Errorf("usernames: %v", h.Usernames)
	}
	// 2026-11-10 17:00 UTC = 11-11 00:00 WIB: hari WIB 10 hanya memuat baris 16:59 UTC; hari 11 memuat dua lainnya
	if h, _ := svc.Daftar(ctx, admin.FilterLog{Dari: "2026-11-10", Sampai: "2026-11-10"}); h.Total != 1 || h.Items[0].Aksi != "tambah" {
		t.Errorf("hari WIB 10: %+v", h)
	}
	if h, _ := svc.Daftar(ctx, admin.FilterLog{Dari: "2026-11-11", Sampai: "2026-11-11"}); h.Total != 2 {
		t.Errorf("hari WIB 11: %+v", h)
	}
	if h, _ := svc.Daftar(ctx, admin.FilterLog{Username: "budi"}); h.Total != 1 {
		t.Errorf("username budi: %+v", h)
	}
	if h, _ := svc.Daftar(ctx, admin.FilterLog{Jenis: "ubah"}); h.Total != 1 {
		t.Errorf("jenis ubah: %+v", h)
	}
	// halaman melewati akhir dibatasi ke halaman terakhir
	if h, _ := svc.Daftar(ctx, admin.FilterLog{Page: 9, Limit: 2}); h.Page != 2 || len(h.Items) != 1 || h.Total != 3 {
		t.Errorf("clamp halaman: %+v", h)
	}
	// validasi
	for nama, f := range map[string]admin.FilterLog{
		"jenis salah":         {Jenis: "post"},
		"tanggal salah":       {Dari: "kemarin"},
		"tanggal mustahil":    {Dari: "2026-02-30"},
		"dari setelah sampai": {Dari: "2026-11-12", Sampai: "2026-11-10"},
	} {
		if _, err := svc.Daftar(ctx, f); err == nil {
			t.Errorf("%s: seharusnya ditolak", nama)
		} else {
			var e *admin.ErrInput
			if !errors.As(err, &e) {
				t.Errorf("%s: bukan ErrInput: %v", nama, err)
			}
		}
	}
}
