package admin

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// barisSatu merapikan teks satu baris: spasi/tab/baris baru dirapatkan menjadi satu spasi. Karakter kontrol ditolak.
func barisSatu(nama, s string, maks int) (string, error) {
	if !utf8.ValidString(s) {
		return "", salah("Field " + nama + " bukan teks UTF-8 yang valid")
	}
	for _, r := range s {
		if unicode.IsControl(r) && !unicode.IsSpace(r) {
			return "", salah("Field " + nama + " memuat karakter yang tidak diizinkan")
		}
	}
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return "", salah("Field " + nama + " wajib diisi")
	}
	if utf8.RuneCountInString(s) > maks {
		return "", salah("Field " + nama + " maksimal " + strconv.Itoa(maks) + " karakter")
	}
	return s, nil
}

var barisKosongBerlebih = regexp.MustCompile(`\n{3,}`)

// judulBoleh merapikan judul dan menegakkan BatasJudul, KECUALI saat Simpan (judulLama != "") dan judul tidak diubah:
// data lama yang sudah melebihi batas tetap boleh disimpan selama judulnya tidak berubah (desain admin). Judul baru
// atau yang diubah tetap wajib memenuhi batas; pembuatan memanggil ini dengan judulLama kosong.
func judulBoleh(nama, s, judulLama string) (string, error) {
	judul, err := barisSatu(nama, s, BatasJudul)
	if err == nil || judulLama == "" {
		return judul, err
	}
	if rapi, err2 := barisSatu(nama, s, math.MaxInt32); err2 == nil && rapi == judulLama {
		return rapi, nil
	}
	return "", err
}

// banyakBaris merapikan teks berparagraf: \r\n dan \r menjadi \n, paragraf dipisah tepat satu baris kosong (\n\n),
// tab diperbolehkan, karakter kontrol lain ditolak, spasi di awal/akhir dibuang.
func banyakBaris(nama, s string, maks int) (string, error) {
	if !utf8.ValidString(s) {
		return "", salah("Field " + nama + " bukan teks UTF-8 yang valid")
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return "", salah("Field " + nama + " memuat karakter yang tidak diizinkan")
		}
	}
	s = strings.TrimSpace(barisKosongBerlebih.ReplaceAllString(s, "\n\n"))
	if s == "" {
		return "", salah("Field " + nama + " wajib diisi")
	}
	if utf8.RuneCountInString(s) > maks {
		return "", salah("Field " + nama + " maksimal " + strconv.Itoa(maks) + " karakter")
	}
	return s, nil
}

// zonaWIB adalah zona waktu acuan "hari ini" (Asia/Jakarta, UTC+7, tanpa DST). FixedZone dipakai agar tidak
// bergantung pada database zona waktu sistem (mis. Windows tanpa tzdata).
var zonaWIB = time.FixedZone("WIB", 7*60*60)

// sekarang dapat diganti di tes.
var sekarang = time.Now

// tanggal memvalidasi YYYY-MM-DD (tanggal kalender yang sah) dan menolak tanggal setelah hari ini (WIB).
func tanggal(nama, s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", salah("Field " + nama + " wajib diisi")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return "", salah("Field " + nama + " harus tanggal valid berformat YYYY-MM-DD")
	}
	if hariIni := sekarang().In(zonaWIB).Format("2006-01-02"); t.Format("2006-01-02") > hariIni {
		return "", salah("Field " + nama + " tidak boleh setelah hari ini")
	}
	return t.Format("2006-01-02"), nil
}

// Slug membuat slug dari judul: huruf kecil, aksen dibuang, selain huruf/angka ASCII menjadi tanda hubung (dirapatkan),
// maksimum BatasSlug karakter. Kosong bila judul tidak memuat huruf atau angka Latin.
func Slug(judul string) string {
	var b strings.Builder
	jeda := true // true = sudah ada pemisah (atau awal), jangan menambah lagi
	for _, r := range norm.NFD.String(judul) {
		switch {
		case unicode.Is(unicode.Mn, r):
			continue // tanda aksen hasil dekomposisi
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			jeda = false
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
			jeda = false
		default:
			if !jeda {
				b.WriteByte('-')
				jeda = true
			}
		}
	}
	s := strings.TrimRight(b.String(), "-")
	if len(s) > BatasSlug {
		s = strings.TrimRight(s[:BatasSlug], "-")
	}
	return s
}
