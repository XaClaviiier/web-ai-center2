// Package konten berisi tipe domain konten publik dan antarmuka pembacanya.
// Paket ini tidak bergantung pada HTTP maupun database, sehingga handler dan store
// sama-sama bergantung padanya (dan handler mudah diuji dengan implementasi palsu).
package konten

import (
	"context"
	"errors"
)

// ErrTidakDitemukan dikembalikan bila konten tidak ada, belum di-Post, atau sudah di-trash.
var ErrTidakDitemukan = errors.New("konten tidak ditemukan")

// ParamDaftar adalah parameter daftar berpagination yang sudah tervalidasi.
type ParamDaftar struct {
	// Pola adalah pola ILIKE yang sudah di-escape dan dibungkus %...%; kosong = tanpa pencarian.
	Pola  string
	Batas int32
	Geser int64
}

// BeritaRingkas adalah item daftar berita.
type BeritaRingkas struct {
	ID               int64
	Slug             string
	Judul            string
	Ringkasan        string
	GambarKey        string
	TanggalPublikasi string // YYYY-MM-DD
}

// BeritaDetail adalah satu berita lengkap.
type BeritaDetail struct {
	BeritaRingkas
	Isi string
}

// DokumentasiRingkas adalah item daftar dokumentasi.
type DokumentasiRingkas struct {
	ID               int64
	Slug             string
	Judul            string
	TanggalPublikasi string // YYYY-MM-DD
	GambarUtamaKey   string // kosong bila tidak ada gambar urutan 1
	JumlahGambar     int    // semua gambar milik dokumentasi
}

// GambarDokumentasi adalah satu gambar dokumentasi (berurutan).
type GambarDokumentasi struct {
	Key string
}

// DokumentasiDetail adalah satu dokumentasi lengkap dengan seluruh gambarnya.
type DokumentasiDetail struct {
	ID               int64
	Slug             string
	Judul            string
	TanggalPublikasi string
	Gambar           []GambarDokumentasi
}

// Testimoni adalah satu testimoni yang sudah di-Post.
type Testimoni struct {
	ID      int64
	Nama    string
	Teks    string
	Posisi  string
	FotoKey string // kosong = tanpa foto
}

// AnggotaTim adalah satu anggota tim.
type AnggotaTim struct {
	ID          int64
	Nama        string
	Posisi      string
	BidangRiset string
	FotoKey     string
}

// BacaPublik adalah kontrak baca konten publik (hanya item posted dan belum di-trash).
type BacaPublik interface {
	ListBerita(ctx context.Context, p ParamDaftar) (items []BeritaRingkas, total int64, err error)
	BeritaBySlug(ctx context.Context, slug string) (BeritaDetail, error)

	ListDokumentasi(ctx context.Context, p ParamDaftar) (items []DokumentasiRingkas, total int64, err error)
	DokumentasiBySlug(ctx context.Context, slug string) (DokumentasiDetail, error)

	ListTestimoni(ctx context.Context) ([]Testimoni, error)
	ListTim(ctx context.Context) ([]AnggotaTim, error)
}

// EntriSitemap adalah satu URL konten untuk sitemap.xml.
type EntriSitemap struct {
	Slug    string
	Tanggal string // tanggal publikasi, YYYY-MM-DD
}

// BacaSitemap membaca slug item posted dan belum di-trash, terbaru dulu, paling banyak batas entri.
type BacaSitemap interface {
	SitemapBerita(ctx context.Context, batas int32) ([]EntriSitemap, error)
	SitemapDokumentasi(ctx context.Context, batas int32) ([]EntriSitemap, error)
}
