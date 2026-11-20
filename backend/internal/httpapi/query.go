package httpapi

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
)

const (
	limitDefault = 9  // stack.md §9.2 dan requirement.md §4.4-§4.5
	limitMaks    = 50 // di atas ini dipotong, bukan ditolak
	panjangQMin  = 2  // stack.md §9.4
	panjangQMaks = 100
)

// daftarQuery adalah query daftar berpagination yang sudah tervalidasi.
type daftarQuery struct {
	Page  int
	Limit int
	// Pola ILIKE yang sudah di-escape dan dibungkus %...%; kosong = tanpa pencarian.
	Pola string
}

func (q daftarQuery) param() konten.ParamDaftar {
	return konten.ParamDaftar{
		Pola:  q.Pola,
		Batas: int32(q.Limit),
		Geser: int64(q.Page-1) * int64(q.Limit),
	}
}

// parseDaftar memvalidasi page, limit, dan q. Galat dikembalikan sebagai pesan untuk klien (HTTP 400).
//   - page: bilangan bulat >= 1 (default 1)
//   - limit: bilangan bulat >= 1 (default 9); di atas 50 dipotong menjadi 50
//   - q: setelah trim dan perapian spasi minimal 2 dan maksimal 100 karakter; kosong dianggap tidak ada
func parseDaftar(v url.Values) (daftarQuery, error) {
	return parseDaftarLimit(v, limitDefault)
}

// parseDaftarLimit sama seperti parseDaftar tetapi dengan limit default pilihan (daftar admin memakai 20).
func parseDaftarLimit(v url.Values, limitAwal int) (daftarQuery, error) {
	out := daftarQuery{Page: 1, Limit: limitAwal}

	if v.Has("page") {
		n, err := strconv.ParseInt(v.Get("page"), 10, 32)
		if err != nil || n < 1 {
			return daftarQuery{}, errors.New("Parameter page harus bilangan bulat minimal 1")
		}
		out.Page = int(n)
	}
	if v.Has("limit") {
		n, err := strconv.ParseInt(v.Get("limit"), 10, 32)
		if err != nil || n < 1 {
			return daftarQuery{}, errors.New("Parameter limit harus bilangan bulat minimal 1")
		}
		out.Limit = int(min(n, limitMaks))
	}
	if v.Has("q") {
		q, err := bersihkanQ(v.Get("q"))
		if err != nil {
			return daftarQuery{}, err
		}
		if q != "" {
			out.Pola = "%" + escapeLike(q) + "%"
		}
	}
	return out, nil
}

func bersihkanQ(raw string) (string, error) {
	if !utf8.ValidString(raw) {
		return "", errors.New("Parameter q bukan teks UTF-8 yang valid")
	}
	for _, r := range raw {
		if unicode.IsControl(r) && !unicode.IsSpace(r) {
			return "", errors.New("Parameter q memuat karakter yang tidak diizinkan")
		}
	}
	q := strings.Join(strings.Fields(raw), " ")
	if q == "" {
		return "", nil
	}
	n := utf8.RuneCountInString(q)
	if n < panjangQMin {
		return "", errors.New("Parameter q minimal 2 karakter")
	}
	if n > panjangQMaks {
		return "", errors.New("Parameter q maksimal 100 karakter")
	}
	return q, nil
}

// escapeLike membuat karakter wildcard LIKE (backslash, persen, garis bawah) diperlakukan sebagai teks biasa.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

var polaSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

const panjangSlugMaks = 200

// slugSah: hanya huruf kecil, angka, dan tanda hubung antar-kata. Slug yang tidak sah diperlakukan
// sama seperti slug yang tidak ada (404), supaya tidak bisa dipakai membedakan kasus.
func slugSah(s string) bool {
	return len(s) > 0 && len(s) <= panjangSlugMaks && polaSlug.MatchString(s)
}
