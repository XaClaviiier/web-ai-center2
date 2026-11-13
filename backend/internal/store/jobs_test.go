package store

import (
	"fmt"
	"strings"
	"testing"
)

func idsStr(ids []int64) string {
	var out []string
	for _, i := range ids {
		out = append(out, fmt.Sprint(i))
	}
	return strings.Join(out, ",")
}

func TestTrashKedaluwarsaHanyaYangLewatRetensiTerlamaDuluDanBerbatas(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := s.Pool.Exec(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	// berita: id 1 = 40 hari lalu, id 2 = 31 hari lalu, id 3 = 29 hari lalu (belum), id 4 = tidak di-trash
	for i, hari := range []int{40, 31, 29} {
		exec(`INSERT INTO berita (slug, judul, isi, tanggal_publikasi, gambar, status, created_by, updated_by, deleted_at)
			VALUES ($1, 'j', 'i', '2026-01-01', 'berita/' || $1 || '.png', 'draft', 'u', 'u', now() - make_interval(days => $2::int))`,
			fmt.Sprintf("b%d", i+1), hari)
	}
	exec(`INSERT INTO berita (slug, judul, isi, tanggal_publikasi, gambar, status, created_by, updated_by) VALUES ('b4', 'j', 'i', '2026-01-01', 'berita/b4.png', 'posted', 'u', 'u')`)
	exec(`INSERT INTO dokumentasi (slug, judul, tanggal_publikasi, status, created_by, updated_by, deleted_at) VALUES ('d1', 'j', '2026-01-01', 'draft', 'u', 'u', now() - interval '35 days')`)
	exec(`INSERT INTO testimoni (nama, teks, posisi, status, created_by, updated_by, deleted_at) VALUES ('n', 't', 'p', 'draft', 'u', 'u', now() - interval '32 days')`)
	exec(`INSERT INTO testimoni (nama, teks, posisi, status, created_by, updated_by) VALUES ('hidup', 't', 'p', 'posted', 'u', 'u')`)

	got, err := s.TrashKedaluwarsa(ctx, "berita", 30, 100)
	if err != nil || idsStr(got) != "1,2" {
		t.Errorf("berita: %v %v (mau 1,2 terlama dulu)", got, err)
	}
	if got, _ = s.TrashKedaluwarsa(ctx, "berita", 30, 1); idsStr(got) != "1" {
		t.Errorf("batas 1: %v", got)
	}
	if got, _ = s.TrashKedaluwarsa(ctx, "berita", 45, 100); len(got) != 0 {
		t.Errorf("retensi 45 hari: %v", got)
	}
	if got, _ = s.TrashKedaluwarsa(ctx, "dokumentasi", 30, 100); idsStr(got) != "1" {
		t.Errorf("dokumentasi: %v", got)
	}
	if got, _ = s.TrashKedaluwarsa(ctx, "testimoni", 30, 100); idsStr(got) != "1" {
		t.Errorf("testimoni: %v", got)
	}
	if _, err := s.TrashKedaluwarsa(ctx, "tim", 30, 100); err == nil {
		t.Error("tim tidak punya trash: harus galat")
	}
}

func TestKeyDipakaiMemeriksaKeempatTabelTermasukDraftDanTrash(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	for _, q := range []string{
		`INSERT INTO berita (slug, judul, isi, tanggal_publikasi, gambar, status, created_by, updated_by, deleted_at) VALUES ('b', 'j', 'i', '2026-01-01', 'berita/trash.png', 'draft', 'u', 'u', now())`,
		`INSERT INTO berita (slug, judul, isi, tanggal_publikasi, gambar, status, created_by, updated_by) VALUES ('c', 'j', 'i', '2026-01-01', 'berita/draft.png', 'draft', 'u', 'u')`,
		`INSERT INTO dokumentasi (slug, judul, tanggal_publikasi, status, created_by, updated_by) VALUES ('d', 'j', '2026-01-01', 'draft', 'u', 'u')`,
		`INSERT INTO dokumentasi_gambar (dokumentasi_id, key, urutan) SELECT id, 'dokumentasi/g.png', 1 FROM dokumentasi WHERE slug = 'd'`,
		`INSERT INTO testimoni (nama, teks, posisi, foto, status, created_by, updated_by) VALUES ('n', 't', 'p', 'testimoni/f.png', 'draft', 'u', 'u')`,
		`INSERT INTO tim (nama, posisi, bidang_riset, foto, urutan, created_by, updated_by) VALUES ('n', 'p', 'b', 'tim/f.png', 1, 'u', 'u')`,
	} {
		if _, err := s.Pool.Exec(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	for _, k := range []string{"berita/trash.png", "berita/draft.png", "dokumentasi/g.png", "testimoni/f.png", "tim/f.png"} {
		if ada, err := s.KeyDipakai(ctx, k); err != nil || !ada {
			t.Errorf("%s harus terdeteksi dipakai: %v %v", k, ada, err)
		}
	}
	for _, k := range []string{"berita/lain.png", "tim/trash.png", "", "berita/"} {
		if ada, err := s.KeyDipakai(ctx, k); err != nil || ada {
			t.Errorf("%q tidak dipakai: %v %v", k, ada, err)
		}
	}
}

func TestHapusLogKedaluwarsaHanyaYangLebihTuaDariBulanDanBerbatas(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := s.Pool.Exec(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`DELETE FROM log_aktivitas`)
	// umur dalam bulan: 5, 4 (kedaluwarsa), 2, 0 (masih berlaku); lalu 3 bulan + 1 hari (kedaluwarsa)
	for i, umur := range []string{"5 months", "4 months", "2 months", "0 days", "3 months 1 day"} {
		exec(`INSERT INTO log_aktivitas (terjadi_pada, pengguna_sub, username, jenis, aksi, modul, item_id, label_item)
			VALUES (now() - $1::interval, 's', 'u', 'tambah', 'tambah', 'berita', $2, 'x')`, umur, i+1)
	}
	if n, err := s.HitungLogKedaluwarsa(ctx, 3); err != nil || n != 3 {
		t.Fatalf("hitung = %d %v (mau 3)", n, err)
	}
	n, err := s.HapusLogKedaluwarsa(ctx, 3, 2)
	if err != nil || n != 2 {
		t.Fatalf("batch pertama = %d %v (mau 2, berbatas)", n, err)
	}
	if n, err = s.HapusLogKedaluwarsa(ctx, 3, 100); err != nil || n != 1 {
		t.Fatalf("batch kedua = %d %v (mau 1)", n, err)
	}
	var sisa []int64
	rows, err := s.Pool.Query(ctx, `SELECT item_id FROM log_aktivitas ORDER BY item_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		sisa = append(sisa, id)
	}
	if idsStr(sisa) != "3,4" {
		t.Errorf("sisa item_id = %s (mau 3,4: umur 2 bulan dan baru)", idsStr(sisa))
	}
	if n, _ = s.HapusLogKedaluwarsa(ctx, 3, 100); n != 0 {
		t.Errorf("tanpa kedaluwarsa harus 0, dapat %d", n)
	}
}
