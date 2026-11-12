package admin

import (
	"context"
	"time"
)

// Jenis aktivitas yang dapat difilter (docs/requirement.md, Log Aktivitas).
const (
	JenisTambah = "tambah"
	JenisUbah   = "ubah"
	JenisHapus  = "hapus"
)

// BatasLogPerHalaman: log memakai 10 entri per halaman (daftar CRUD memakai aturan sendiri).
const (
	LogLimitDefault  = 10
	LogLimitMaks     = 50
	BatasUsernameLog = 100
)

// zonaLog adalah acuan hari untuk rentang tanggal (Asia/Jakarta, UTC+7 tanpa DST).
var zonaLog = time.FixedZone("WIB", 7*60*60)

// EntriLog adalah satu catatan aktivitas (baca-saja).
type EntriLog struct {
	ID          int64
	TerjadiPada time.Time // UTC
	PenggunaSub string
	Username    string
	Jenis       string
	Aksi        string
	Modul       string
	ItemID      int64
	LabelItem   string
}

// FilterLog adalah parameter daftar log dari klien. Dari/Sampai berformat YYYY-MM-DD (hari Asia/Jakarta, inklusif).
type FilterLog struct {
	Page, Limit  int
	Username     string
	Jenis        string
	Dari, Sampai string
}

// ParamRepoLog adalah filter yang sudah divalidasi: rentang waktu UTC setengah terbuka [Dari, Sampai).
type ParamRepoLog struct {
	Username     string
	Jenis        string
	Dari, Sampai *time.Time
	Batas        int32
	Geser        int64
}

// HasilLog adalah satu halaman log beserta daftar username untuk pilihan filter.
type HasilLog struct {
	Items     []EntriLog
	Page      int
	Limit     int
	Total     int64
	Usernames []string
}

// LogRepo adalah akses data log (baca saja; log ditulis trigger database di transaksi perubahan data).
type LogRepo interface {
	Daftar(ctx context.Context, p ParamRepoLog) ([]EntriLog, error)
	Hitung(ctx context.Context, p ParamRepoLog) (int64, error)
	Usernames(ctx context.Context) ([]string, error)
}

// LogLayanan menerapkan aturan baca log.
type LogLayanan struct{ Repo LogRepo }

func hariJakarta(nama, s string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", s, zonaLog)
	if err != nil {
		return time.Time{}, salah("Parameter " + nama + " harus tanggal valid berformat YYYY-MM-DD")
	}
	return t, nil
}

// Daftar memvalidasi filter lalu mengembalikan satu halaman. Nomor halaman yang melebihi halaman terakhir dibatasi
// ke halaman terakhir (mengikuti prototipe); urutan terbaru dulu, id menurun sebagai pemutus seri.
func (s LogLayanan) Daftar(ctx context.Context, f FilterLog) (HasilLog, error) {
	switch f.Jenis {
	case "", JenisTambah, JenisUbah, JenisHapus:
	default:
		return HasilLog{}, salah("Parameter jenis harus tambah, ubah, atau hapus")
	}
	p := ParamRepoLog{Username: f.Username, Jenis: f.Jenis}
	var dari, sampai time.Time
	if f.Dari != "" {
		t, err := hariJakarta("dari", f.Dari)
		if err != nil {
			return HasilLog{}, err
		}
		dari = t
		u := t.UTC()
		p.Dari = &u
	}
	if f.Sampai != "" {
		t, err := hariJakarta("sampai", f.Sampai)
		if err != nil {
			return HasilLog{}, err
		}
		sampai = t
		u := t.AddDate(0, 0, 1).UTC() // awal hari sesudah tanggal akhir -> inklusif
		p.Sampai = &u
	}
	if f.Dari != "" && f.Sampai != "" && dari.After(sampai) {
		return HasilLog{}, salah("Tanggal awal harus sebelum atau sama dengan tanggal akhir")
	}
	limit := f.Limit
	if limit < 1 {
		limit = LogLimitDefault
	}
	limit = min(limit, LogLimitMaks)

	total, err := s.Repo.Hitung(ctx, p)
	if err != nil {
		return HasilLog{}, err
	}
	halaman := max(1, f.Page)
	terakhir := max(1, int((total+int64(limit)-1)/int64(limit)))
	halaman = min(halaman, terakhir)
	p.Batas, p.Geser = int32(limit), int64(halaman-1)*int64(limit)
	items, err := s.Repo.Daftar(ctx, p)
	if err != nil {
		return HasilLog{}, err
	}
	nama, err := s.Repo.Usernames(ctx)
	if err != nil {
		return HasilLog{}, err
	}
	return HasilLog{Items: items, Page: halaman, Limit: limit, Total: total, Usernames: nama}, nil
}
