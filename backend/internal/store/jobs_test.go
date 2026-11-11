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
