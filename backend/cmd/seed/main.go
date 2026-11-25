// Command seed mengisi database DEVELOPMENT dengan konten contoh agar clone baru tidak kosong.
// Isi teks diambil dari seed stack lama (server/src/db/seed.ts di riwayat git, commit 917b2d5^); gambar berupa
// placeholder PNG yang dibuat di sini lalu diunggah ke RustFS. Semua baris ditandai created_by = 'seed-dev'.
//
//	cd backend && go run ./cmd/seed
//
// Aman dijalankan berulang: baris yang sudah ada (slug untuk berita/dokumentasi, nama untuk tim/testimoni) dilewati.
// Menolak berjalan bila APP_ENV=production. Konten production diisi lewat halaman admin, bukan seed.
// Menghapus data seed: DELETE FROM <tabel> WHERE created_by = 'seed-dev'; file gambarnya dibersihkan job file yatim.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/cache"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/config"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store"
)

const penanda = "seed-dev"

func main() {
	if err := jalan(); err != nil {
		fmt.Fprintln(os.Stderr, "seed gagal:", err)
		os.Exit(1)
	}
}

// bolehJalan menolak production: seed hanya untuk development, test, dan staging.
func bolehJalan(appEnv string) error {
	if appEnv == config.EnvProduction {
		return errors.New("seed tidak boleh dijalankan dengan APP_ENV=production")
	}
	return nil
}

func jalan() error {
	if err := config.LoadDotEnv(".env"); err != nil {
		return fmt.Errorf("membaca .env: %w", err)
	}
	if err := bolehJalan(os.Getenv("APP_ENV")); err != nil {
		return err
	}
	for _, k := range []string{"DATABASE_URL", "REDIS_URL", "RUSTFS_ENDPOINT", "RUSTFS_ACCESS_KEY", "RUSTFS_SECRET_KEY", "RUSTFS_BUCKET"} {
		if os.Getenv(k) == "" {
			return fmt.Errorf("%s wajib diisi (lihat backend/.env.example)", k)
		}
	}

	ctx, batal := context.WithTimeout(context.Background(), 2*time.Minute)
	defer batal()
	st, err := store.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer st.Close()
	gambar := media.NewS3(media.S3Config{
		Endpoint:  os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey: os.Getenv("RUSTFS_ACCESS_KEY"),
		SecretKey: os.Getenv("RUSTFS_SECRET_KEY"),
		Bucket:    os.Getenv("RUSTFS_BUCKET"),
	})
	if err := gambar.PastikanBucket(ctx, true); err != nil {
		return fmt.Errorf("storage gambar: %w", err)
	}

	s := &penyemai{ctx: ctx, st: st, gambar: gambar}
	for _, b := range dataBerita {
		s.berita(b)
	}
	for i, d := range dataDokumentasi {
		s.dokumentasi(d, i)
	}
	for _, t := range dataTim {
		s.tim(t)
	}
	for _, t := range dataTestimoni {
		s.testimoni(t)
	}
	if s.err != nil {
		return s.err
	}

	// Seed menulis langsung ke database, jadi cache publik di Redis dibatalkan agar konten baru langsung tampil.
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return fmt.Errorf("REDIS_URL: %w", err)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()
	c := &cache.Cache{C: rdb}
	for _, jenis := range []string{cache.Berita, cache.Dokumentasi, cache.Testimoni, cache.Tim, cache.Sitemap} {
		if err := c.Batalkan(ctx, jenis); err != nil {
			return err
		}
	}
	fmt.Printf("selesai: %d dibuat, %d sudah ada (dilewati); cache publik dibatalkan\n", s.dibuat, s.dilewati)
	return nil
}

type penyemai struct {
	ctx              context.Context
	st               *store.Store
	gambar           *media.S3Store
	dibuat, dilewati int
	err              error
}

// ada memeriksa baris dengan nilai kolom tertentu; tabel dan kolom berasal dari kode ini, bukan input.
func (s *penyemai) ada(tabel, kolom, nilai string) bool {
	if s.err != nil {
		return true
	}
	var n int
	if err := s.st.Pool.QueryRow(s.ctx, "SELECT count(*) FROM "+tabel+" WHERE "+kolom+" = $1", nilai).Scan(&n); err != nil {
		s.err = err
		return true
	}
	if n > 0 {
		s.dilewati++
	}
	return n > 0
}

// unggah membuat placeholder PNG berwarna lalu menyimpannya di RustFS, mengembalikan key-nya.
// shortcut: bila INSERT sesudahnya gagal, file menjadi yatim dan dibersihkan job berkala; cukup untuk data dev.
func (s *penyemai) unggah(sumber string, lebar, tinggi int, warna color.RGBA) string {
	if s.err != nil {
		return ""
	}
	img := image.NewRGBA(image.Rect(0, 0, lebar, tinggi))
	for y := 0; y < tinggi; y++ {
		gelap := uint8(60 * y / tinggi) // gradasi vertikal tipis agar placeholder tidak datar
		for x := 0; x < lebar; x++ {
			img.SetRGBA(x, y, color.RGBA{warna.R - min(warna.R, gelap), warna.G - min(warna.G, gelap), warna.B - min(warna.B, gelap), 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		s.err = err
		return ""
	}
	key, err := media.BuatKey(sumber, "png")
	if err == nil {
		err = s.gambar.Put(s.ctx, key, buf.Bytes(), "image/png")
	}
	if err != nil {
		s.err = fmt.Errorf("unggah %s: %w", sumber, err)
		return ""
	}
	return key
}

func (s *penyemai) exec(sql string, args ...any) {
	if s.err != nil {
		return
	}
	if _, err := s.st.Pool.Exec(s.ctx, sql, args...); err != nil {
		s.err = err
		return
	}
	s.dibuat++
}

var palet = []color.RGBA{{6, 116, 253, 255}, {10, 34, 64, 255}, {255, 107, 44, 255}, {32, 150, 120, 255}, {120, 90, 200, 255}, {200, 160, 40, 255}}

func (s *penyemai) berita(b berita) {
	if s.ada("berita", "slug", b.slug) {
		return
	}
	key := s.unggah(media.SumberBerita, 1200, 675, palet[0])
	s.exec(`INSERT INTO berita (slug, judul, isi, tanggal_publikasi, gambar, status, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, 'posted', $6, $6)`, b.slug, b.judul, b.isi, b.tanggal, key, penanda)
}

func (s *penyemai) dokumentasi(d dokumentasi, i int) {
	if s.ada("dokumentasi", "slug", d.slug) {
		return
	}
	key := s.unggah(media.SumberDokumentasi, 1200, 800, palet[i%len(palet)])
	if s.err != nil {
		return
	}
	tx, err := s.st.Pool.Begin(s.ctx)
	if err != nil {
		s.err = err
		return
	}
	defer tx.Rollback(s.ctx) //nolint:errcheck // tidak berefek setelah Commit
	var id int64
	if err := tx.QueryRow(s.ctx, `INSERT INTO dokumentasi (slug, judul, tanggal_publikasi, status, created_by, updated_by)
		VALUES ($1, $2, $3, 'posted', $4, $4) RETURNING id`, d.slug, d.judul, d.tanggal, penanda).Scan(&id); err != nil {
		s.err = err
		return
	}
	if _, err := tx.Exec(s.ctx, `INSERT INTO dokumentasi_gambar (dokumentasi_id, key, urutan) VALUES ($1, $2, 1)`, id, key); err != nil {
		s.err = err
		return
	}
	if err := tx.Commit(s.ctx); err != nil {
		s.err = err
		return
	}
	s.dibuat++
}

func (s *penyemai) tim(t tim) {
	if s.ada("tim", "nama", t.nama) {
		return
	}
	key := s.unggah(media.SumberTim, 400, 400, palet[1])
	s.exec(`INSERT INTO tim (nama, posisi, bidang_riset, foto, urutan, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $6)`, t.nama, t.posisi, "Belum diisi (data contoh)", key, t.urutan, penanda)
}

func (s *penyemai) testimoni(t testimoni) {
	if s.ada("testimoni", "nama", t.nama) {
		return
	}
	s.exec(`INSERT INTO testimoni (nama, teks, posisi, status, created_by, updated_by)
		VALUES ($1, $2, $3, 'posted', $4, $4)`, t.nama, t.teks, t.posisi, penanda)
}
