package jobs

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

// Tes ujung ke ujung: PostgreSQL tes + RustFS tes (bucket ai-center-test) + Redis DB 15 + layanan admin nyata.
func TestE2ESiklusMenghapusTrashKedaluwarsaDanFileYatim(t *testing.T) {
	url := testutil.DatabaseTes(t)
	k := testutil.RustFSTes(t)
	rdb := testutil.RedisTes(t)
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

	keys := []string{
		"berita/trash-lama.png", "berita/trash-baru.png", "berita/hidup.png", "berita/yatim.png",
		"dokumentasi/lama-1.png", "dokumentasi/lama-2.png", "testimoni/lama.png", "tim/anggota.png",
	}
	t.Cleanup(func() { // hanya key milik tes ini
		for _, key := range keys {
			_ = s3.Hapus(context.Background(), key)
		}
	})
	for _, key := range keys {
		if err := s3.Put(ctx, key, []byte("x"), "image/png"); err != nil {
			t.Fatal(err)
		}
	}
	exec := func(q string) {
		t.Helper()
		if _, err := st.Pool.Exec(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO berita (slug, judul, isi, tanggal_publikasi, gambar, status, created_by, updated_by, deleted_at) VALUES
		('trash-lama', 'j', 'i', '2026-01-01', 'berita/trash-lama.png', 'draft', 'u', 'u', now() - interval '31 days'),
		('trash-baru', 'j', 'i', '2026-01-01', 'berita/trash-baru.png', 'draft', 'u', 'u', now() - interval '5 days')`)
	exec(`INSERT INTO berita (slug, judul, isi, tanggal_publikasi, gambar, status, created_by, updated_by) VALUES
		('hidup', 'j', 'i', '2026-01-01', 'berita/hidup.png', 'posted', 'u', 'u')`)
	exec(`INSERT INTO dokumentasi (slug, judul, tanggal_publikasi, status, created_by, updated_by, deleted_at) VALUES ('dok-lama', 'j', '2026-01-01', 'draft', 'u', 'u', now() - interval '40 days')`)
	exec(`INSERT INTO dokumentasi_gambar (dokumentasi_id, key, urutan) SELECT id, 'dokumentasi/lama-1.png', 1 FROM dokumentasi WHERE slug = 'dok-lama'`)
	exec(`INSERT INTO dokumentasi_gambar (dokumentasi_id, key, urutan) SELECT id, 'dokumentasi/lama-2.png', 2 FROM dokumentasi WHERE slug = 'dok-lama'`)
	exec(`INSERT INTO testimoni (nama, teks, posisi, foto, status, created_by, updated_by, deleted_at) VALUES ('n', 't', 'p', 'testimoni/lama.png', 'draft', 'u', 'u', now() - interval '33 days')`)
	exec(`INSERT INTO tim (nama, posisi, bidang_riset, foto, urutan, created_by, updated_by) VALUES ('a', 'p', 'b', 'tim/anggota.png', 1, 'u', 'u')`)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	berita := admin.BeritaLayanan{Repo: st.AdminBerita(), Media: s3, Log: log}
	dok := admin.DokumentasiLayanan{Repo: st.AdminDokumentasi(), Media: s3, Log: log}
	testi := admin.TestimoniLayanan{Repo: st.AdminTestimoni(), Media: s3, Log: log}
	sumber := func(nama string, hapus func(context.Context, int64) error) Sumber {
		return Sumber{
			Nama: nama,
			Kedaluwarsa: func(ctx context.Context, hari, batas int32) ([]int64, error) {
				return st.TrashKedaluwarsa(ctx, nama, hari, batas)
			},
			HapusPermanen: hapus,
		}
	}
	j := &Job{
		Log: log, Redis: rdb, Gudang: s3, Dipakai: st.KeyDipakai,
		Sumber: []Sumber{sumber("berita", berita.HapusPermanen), sumber("dokumentasi", dok.HapusPermanen), sumber("testimoni", testi.HapusPermanen)},
		Cfg: Config{
			Interval: time.Hour, LockTTL: 30 * time.Minute, RetensiTrashHari: 30,
			UmurMinYatim: 24 * time.Hour, MaksHapusYatim: 200,
		},
		// objek baru diunggah dianggap berumur 48 jam
		Sekarang: func() time.Time { return time.Now().Add(48 * time.Hour) },
	}

	j.Siklus(ctx)

	adaFile := func(key string) bool {
		ada, err := media.Ada(ctx, s3, key)
		if err != nil {
			t.Fatal(err)
		}
		return ada
	}
	hitung := func(q string) int {
		var n int
		if err := st.Pool.QueryRow(ctx, q).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	for _, key := range []string{"berita/trash-lama.png", "dokumentasi/lama-1.png", "dokumentasi/lama-2.png", "testimoni/lama.png", "berita/yatim.png"} {
		if adaFile(key) {
			t.Errorf("%s harus terhapus", key)
		}
	}
	for _, key := range []string{"berita/trash-baru.png", "berita/hidup.png", "tim/anggota.png"} {
		if !adaFile(key) {
			t.Errorf("%s tidak boleh terhapus", key)
		}
	}
	if n := hitung(`SELECT count(*) FROM berita`); n != 2 {
		t.Errorf("berita tersisa %d, mau 2 (trash-baru dan hidup)", n)
	}
	if hitung(`SELECT count(*) FROM dokumentasi`) != 0 || hitung(`SELECT count(*) FROM dokumentasi_gambar`) != 0 || hitung(`SELECT count(*) FROM testimoni`) != 0 {
		t.Error("dokumentasi (beserta baris gambar) dan testimoni kedaluwarsa harus terhapus")
	}
	if hitung(`SELECT count(*) FROM tim`) != 1 {
		t.Error("tim tidak punya trash dan tidak boleh tersentuh")
	}
	if rdb.Exists(ctx, KunciJob).Val() != 0 {
		t.Error("kunci job harus dilepas")
	}

	// Siklus kedua tidak mengubah apa pun (idempoten).
	j.Siklus(ctx)
	if !adaFile("berita/hidup.png") || hitung(`SELECT count(*) FROM berita`) != 2 {
		t.Error("siklus kedua tidak boleh mengubah data")
	}
}
