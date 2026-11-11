// Package media menangani penyimpanan gambar di RustFS (S3-compatible): aturan key, deteksi tipe dari isi file,
// dan antarmuka Store dengan implementasi S3 dan memori (untuk tes). Rujukan: docs/database.md §5, docs/security.md §7.
package media

import (
	"context"
	"errors"
	"io"
	"time"
)

// ErrTidakAda berarti objek tidak ada di storage.
var ErrTidakAda = errors.New("objek tidak ditemukan")

// Meta adalah metadata objek tanpa isinya.
type Meta struct {
	ContentType string
	Ukuran      int64
	ETag        string // sudah termasuk tanda kutip, siap dipakai di header
}

// Objek adalah isi objek beserta metadatanya. Pemanggil wajib menutup Body.
type Objek struct {
	Meta
	Body io.ReadCloser
}

// Info adalah satu entri hasil Daftar (untuk job pembersih file yatim).
type Info struct {
	Key    string
	Ukuran int64
	Diubah time.Time
}

// Store adalah kontrak penyimpanan gambar. Gambar dibatasi beberapa MB sehingga Put menerima []byte utuh.
type Store interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
	// Get dan Head mengembalikan ErrTidakAda bila key tidak ada.
	Get(ctx context.Context, key string) (Objek, error)
	Head(ctx context.Context, key string) (Meta, error)
	// Hapus idempotent: menghapus key yang tidak ada bukan galat.
	Hapus(ctx context.Context, key string) error
}

// Pendaftar mendaftar objek berdasarkan prefix.
type Pendaftar interface {
	Daftar(ctx context.Context, prefix string) ([]Info, error)
}

// Ada memeriksa keberadaan key lewat Head.
func Ada(ctx context.Context, s Store, key string) (bool, error) {
	_, err := s.Head(ctx, key)
	if errors.Is(err, ErrTidakAda) {
		return false, nil
	}
	return err == nil, err
}
