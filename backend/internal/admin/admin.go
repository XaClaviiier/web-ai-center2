// Package admin berisi aturan bisnis penulisan konten oleh admin: validasi input, slug, status (draft/posted/trash),
// optimistic locking, dan siklus hidup file gambar. Tidak bergantung pada HTTP maupun SQL (rujukan: docs/database.md §3-§6,
// docs/requirement.md §5).
package admin

import (
	"errors"
	"time"
)

// Galat domain. Handler HTTP memetakannya ke kode status; galat lain dianggap kesalahan internal.
var (
	ErrTidakDitemukan   = errors.New("konten tidak ditemukan")
	ErrStatusTidakValid = errors.New("aksi tidak diizinkan pada status konten saat ini")
	ErrVersiBentrok     = errors.New("version tidak cocok: konten sudah diubah pihak lain")
	ErrGambarHilang     = errors.New("gambar konten tidak ditemukan di penyimpanan")
	ErrGambarDipakai    = errors.New("gambar sudah dipakai konten lain")
	// ErrStorage membungkus kegagalan penyimpanan gambar (RustFS) agar dibalas 503, bukan 500.
	ErrStorage = errors.New("penyimpanan gambar bermasalah")
)

// ErrInput adalah galat validasi input klien (HTTP 400); Pesan aman ditampilkan ke klien.
type ErrInput struct{ Pesan string }

func (e *ErrInput) Error() string { return e.Pesan }

func salah(pesan string) error { return &ErrInput{Pesan: pesan} }

// Filter daftar admin.
const (
	FilterSemua  = "semua" // draft + posted, tanpa trash
	FilterDraft  = "draft"
	FilterPosted = "posted"
	FilterTrash  = "trash"
)

// FilterSah memeriksa nilai query status.
func FilterSah(f string) bool {
	switch f {
	case FilterSemua, FilterDraft, FilterPosted, FilterTrash:
		return true
	}
	return false
}

// ParamDaftar adalah parameter daftar admin yang sudah tervalidasi.
type ParamDaftar struct {
	Filter string
	Pola   string // pola ILIKE ter-escape, kosong = tanpa pencarian
	Batas  int32
	Geser  int64
}

// Batas panjang teks (karakter, setelah dirapikan). Satu tempat agar mudah disesuaikan (keputusan pengguna).
const (
	BatasJudul = 70 // docs/requirement.md; frontend (MAKS_JUDUL) memakai angka sama. Semua judul lebih panjang ditolak.
	BatasIsi   = 20000
	BatasSlug  = 120
)

// Berita adalah satu baris berita lengkap untuk admin (semua status, termasuk trash).
type Berita struct {
	ID               int64
	Slug             string
	Judul            string
	Ringkasan        string
	Isi              string
	TanggalPublikasi string // YYYY-MM-DD
	GambarKey        string
	Status           string // draft | posted
	Version          int32
	DibuatPada       time.Time
	DibuatOleh       string
	DiubahOleh       string
	DihapusPada      *time.Time // non-nil = di trash
}

// InputBerita adalah isi berita yang sudah divalidasi dan dirapikan.
type InputBerita struct {
	Judul            string
	Isi              string
	TanggalPublikasi string // YYYY-MM-DD
	Gambar           string // key RustFS
}

// MentahBerita adalah isi berita apa adanya dari klien (belum divalidasi).
type MentahBerita struct {
	Judul            string
	Isi              string
	TanggalPublikasi string
	Gambar           string
}

// Batas panjang teks tim dan testimoni (karakter, setelah dirapikan); satu tempat agar mudah disesuaikan.
const (
	BatasNama          = 100
	BatasPosisi        = 100
	BatasBidangRiset   = 150
	BatasTeksTestimoni = 1000
)
