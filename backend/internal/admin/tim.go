package admin

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
)

// Tim adalah satu anggota tim untuk admin. Tim tidak punya status dan trash (database.md §6).
type Tim struct {
	ID          int64
	Nama        string
	Posisi      string
	BidangRiset string
	FotoKey     string
	Urutan      int32 // urutan tampil: diisi admin, angka kecil tampil lebih dulu
	Version     int32
	DibuatPada  time.Time
	DibuatOleh  string
	DiubahOleh  string
}

// MentahTim adalah isi anggota tim apa adanya dari klien.
type MentahTim struct {
	Nama        string
	Posisi      string
	BidangRiset string
	Foto        string // key RustFS; wajib
	Urutan      *int64 // wajib (nil = tidak dikirim); angka bulat >= 1
}

// InputTim adalah isi yang sudah divalidasi dan dirapikan.
type InputTim struct {
	Nama        string
	Posisi      string
	BidangRiset string
	Foto        string
	Urutan      int32
}

// TimRepo adalah akses data tim untuk admin; galat: ErrTidakDitemukan, ErrVersiBentrok, ErrGambarDipakai.
type TimRepo interface {
	Daftar(ctx context.Context) ([]Tim, error) // berurutan: urutan, id
	Detail(ctx context.Context, id int64) (Tim, error)
	Buat(ctx context.Context, in InputTim, oleh string) (Tim, error)
	Ubah(ctx context.Context, id int64, version int32, in InputTim, oleh string) (Tim, error)
	Hapus(ctx context.Context, id int64) (fotoKey string, err error) // langsung permanen
	FotoDipakai(ctx context.Context, key string) (bool, error)
}

// TimLayanan menerapkan aturan bisnis tim di atas TimRepo dan penyimpanan gambar.
type TimLayanan struct {
	Repo  TimRepo
	Media media.Store
	Log   *slog.Logger
}

// ValidasiTim memvalidasi dan merapikan isi anggota tim (tanpa memeriksa key foto ke storage).
func ValidasiTim(m MentahTim) (InputTim, error) {
	nama, err := barisSatu("nama", m.Nama, BatasNama)
	if err != nil {
		return InputTim{}, err
	}
	posisi, err := barisSatu("posisi", m.Posisi, BatasPosisi)
	if err != nil {
		return InputTim{}, err
	}
	bidang, err := barisSatu("bidangRiset", m.BidangRiset, BatasBidangRiset)
	if err != nil {
		return InputTim{}, err
	}
	if m.Foto == "" {
		return InputTim{}, salah("Field foto wajib diisi")
	}
	// Urutan tampil: angka bulat >= 1 (docs/requirement.md); batas atas hanya rentang kolom integer.
	if m.Urutan == nil {
		return InputTim{}, salah("Field urutan wajib diisi")
	}
	if *m.Urutan < 1 || *m.Urutan > math.MaxInt32 {
		return InputTim{}, salah("Field urutan harus bilangan bulat minimal 1")
	}
	return InputTim{Nama: nama, Posisi: posisi, BidangRiset: bidang, Foto: m.Foto, Urutan: int32(*m.Urutan)}, nil
}

func (s TimLayanan) file() pengelolaFile {
	return pengelolaFile{Media: s.Media, Log: s.Log, Sumber: media.SumberTim, Dipakai: s.Repo.FotoDipakai}
}

func (s TimLayanan) Daftar(ctx context.Context) ([]Tim, error) { return s.Repo.Daftar(ctx) }

func (s TimLayanan) Detail(ctx context.Context, id int64) (Tim, error) { return s.Repo.Detail(ctx, id) }

// Buat menambah anggota tim baru dengan urutan tampil pilihan admin.
func (s TimLayanan) Buat(ctx context.Context, m MentahTim, oleh string) (hasil Tim, err error) {
	berhasil := false
	defer func() {
		if !berhasil {
			s.file().bersihkanBaru(m.Foto, "")
		}
	}()
	in, err := ValidasiTim(m)
	if err != nil {
		return Tim{}, err
	}
	if err := s.file().periksaKey(ctx, in.Foto); err != nil {
		return Tim{}, err
	}
	t, err := s.Repo.Buat(ctx, in, oleh)
	if errors.Is(err, ErrGambarDipakai) {
		return Tim{}, salah("Foto sudah dipakai konten lain")
	}
	if err != nil {
		return Tim{}, err
	}
	berhasil = true
	return t, nil
}

// Ubah menyimpan perubahan; version harus cocok. Tim tidak punya aturan draft.
func (s TimLayanan) Ubah(ctx context.Context, id int64, m MentahTim, version int32, oleh string) (hasil Tim, err error) {
	lama := ""
	berhasil := false
	defer func() {
		if !berhasil {
			s.file().bersihkanBaru(m.Foto, lama)
		}
	}()

	cur, err := s.Repo.Detail(ctx, id)
	if err != nil {
		return Tim{}, err
	}
	lama = cur.FotoKey
	if cur.Version != version {
		return Tim{}, ErrVersiBentrok
	}
	in, err := ValidasiTim(m)
	if err != nil {
		return Tim{}, err
	}
	if in.Foto != lama {
		if err := s.file().periksaKey(ctx, in.Foto); err != nil {
			return Tim{}, err
		}
	}
	t, err := s.Repo.Ubah(ctx, id, version, in, oleh)
	if errors.Is(err, ErrGambarDipakai) {
		return Tim{}, salah("Foto sudah dipakai konten lain")
	}
	if err != nil {
		return Tim{}, err
	}
	berhasil = true
	if in.Foto != lama {
		s.file().hapusLama(lama) // setelah commit
	}
	return t, nil
}

// Hapus menghapus anggota langsung permanen, lalu (setelah commit) file fotonya.
func (s TimLayanan) Hapus(ctx context.Context, id int64) error {
	key, err := s.Repo.Hapus(ctx, id)
	if err != nil {
		return err
	}
	s.file().hapusLama(key)
	return nil
}
