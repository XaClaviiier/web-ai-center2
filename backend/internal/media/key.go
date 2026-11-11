package media

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"regexp"
	"slices"
)

// Sumber adalah jenis konten pemilik gambar; menjadi prefix key.
const (
	SumberBerita      = "berita"
	SumberDokumentasi = "dokumentasi"
	SumberTestimoni   = "testimoni"
	SumberTim         = "tim"
)

var sumberSah = []string{SumberBerita, SumberDokumentasi, SumberTestimoni, SumberTim}

// SumberSah memeriksa apakah s salah satu jenis konten yang punya gambar.
func SumberSah(s string) bool { return slices.Contains(sumberSah, s) }

// Jenis adalah tipe gambar yang diterima (jpg, png, webp).
type Jenis struct {
	Ext         string
	ContentType string
}

// DeteksiGambar menentukan tipe dari ISI file (bukan dari nama atau header klien), security.md §7.
// Hanya JPEG, PNG, dan WebP yang diterima.
func DeteksiGambar(data []byte) (Jenis, bool) {
	if len(data) == 0 {
		return Jenis{}, false
	}
	head := data
	if len(head) > 512 {
		head = head[:512]
	}
	switch http.DetectContentType(head) {
	case "image/jpeg":
		return Jenis{Ext: "jpg", ContentType: "image/jpeg"}, true
	case "image/png":
		return Jenis{Ext: "png", ContentType: "image/png"}, true
	case "image/webp":
		return Jenis{Ext: "webp", ContentType: "image/webp"}, true
	}
	return Jenis{}, false
}

// BuatKey membuat key unik {sumber}/{uuid-v4}.{ext}. Key tidak memuat id konten karena file diunggah sebelum
// barisnya ada (keputusan pengguna; menyimpang dari pola {sumber}/{id}/{uuid}.ext di docs/database.md §5).
func BuatKey(sumber, ext string) (string, error) {
	if !SumberSah(sumber) {
		return "", fmt.Errorf("sumber %q tidak dikenal", sumber)
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("membuat uuid: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // versi 4
	b[8] = (b[8] & 0x3f) | 0x80 // varian RFC 4122
	return fmt.Sprintf("%s/%x-%x-%x-%x-%x.%s", sumber, b[0:4], b[4:6], b[6:8], b[8:10], b[10:16], ext), nil
}

var polaKey = regexp.MustCompile(`^(berita|dokumentasi|testimoni|tim)/[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}\.(jpg|png|webp)$`)

// SumberDariKey memvalidasi bentuk key buatan BuatKey dan mengembalikan sumbernya.
// Dipakai route /media dan, nanti, validasi key dari klien saat Simpan.
func SumberDariKey(key string) (string, bool) {
	m := polaKey.FindStringSubmatch(key)
	if m == nil {
		return "", false
	}
	return m[1], true
}
