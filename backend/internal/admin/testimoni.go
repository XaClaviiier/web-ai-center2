package admin

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
)

// Testimoni adalah satu testimoni untuk admin (semua status, termasuk trash). Alurnya sama dengan berita (database.md §6).
type Testimoni struct {
	ID          int64
	Nama        string
	Teks        string
	Posisi      string
	FotoKey     string // kosong = tanpa foto (tampil inisial nama)
	Status      string // draft | posted
	Version     int32
	DibuatPada  time.Time
	DibuatOleh  string
	DiubahOleh  string
	DihapusPada *time.Time
}

// MentahTestimoni adalah isi testimoni apa adanya dari klien. Foto kosong berarti tanpa foto.
type MentahTestimoni struct {
	Nama   string
	Teks   string
	Posisi string
	Foto   string
}

// InputTestimoni adalah isi yang sudah divalidasi dan dirapikan.
type InputTestimoni struct {
	Nama   string
	Teks   string
	Posisi string
	Foto   string
}

// TestimoniRepo adalah akses data testimoni untuk admin. Aturan status diterapkan atomik oleh implementasi;
// galat: ErrTidakDitemukan, ErrStatusTidakValid, ErrVersiBentrok, ErrGambarDipakai.
type TestimoniRepo interface {
	Daftar(ctx context.Context, filter string) ([]Testimoni, error) // tanpa pagination
	Detail(ctx context.Context, id int64) (Testimoni, error)
	Buat(ctx context.Context, in InputTestimoni, oleh string) (Testimoni, error)
	Ubah(ctx context.Context, id int64, version int32, in InputTestimoni, oleh string) (Testimoni, error)
	Post(ctx context.Context, id int64, version int32, oleh string) (Testimoni, error)
	Retract(ctx context.Context, id int64, version int32, oleh string) (Testimoni, error)
	Hapus(ctx context.Context, id int64, oleh string) error
	Pulihkan(ctx context.Context, id int64, oleh string) (Testimoni, error)
	HapusPermanen(ctx context.Context, id int64) (fotoKey string, err error)
	FotoDipakai(ctx context.Context, key string) (bool, error)
}

// TestimoniLayanan menerapkan aturan bisnis testimoni di atas TestimoniRepo dan penyimpanan gambar.
type TestimoniLayanan struct {
	Repo  TestimoniRepo
	Media media.Store
	Log   *slog.Logger
}

// ValidasiTestimoni memvalidasi dan merapikan isi testimoni (tanpa memeriksa key foto ke storage).
// teks boleh beberapa baris; nama dan posisi satu baris. Foto opsional.
func ValidasiTestimoni(m MentahTestimoni) (InputTestimoni, error) {
	nama, err := barisSatu("nama", m.Nama, BatasNama)
	if err != nil {
		return InputTestimoni{}, err
	}
	teks, err := banyakBaris("teks", m.Teks, BatasTeksTestimoni)
	if err != nil {
		return InputTestimoni{}, err
	}
	posisi, err := barisSatu("posisi", m.Posisi, BatasPosisi)
	if err != nil {
		return InputTestimoni{}, err
	}
	return InputTestimoni{Nama: nama, Teks: teks, Posisi: posisi, Foto: m.Foto}, nil
}

func (s TestimoniLayanan) file() pengelolaFile {
	return pengelolaFile{Media: s.Media, Log: s.Log, Sumber: media.SumberTestimoni, Dipakai: s.Repo.FotoDipakai}
}

func (s TestimoniLayanan) Daftar(ctx context.Context, filter string) ([]Testimoni, error) {
	return s.Repo.Daftar(ctx, filter)
}

func (s TestimoniLayanan) Detail(ctx context.Context, id int64) (Testimoni, error) {
	return s.Repo.Detail(ctx, id)
}

// Buat membuat testimoni baru berstatus draft. Foto opsional.
func (s TestimoniLayanan) Buat(ctx context.Context, m MentahTestimoni, oleh string) (hasil Testimoni, err error) {
	berhasil := false
	defer func() {
		if !berhasil {
			s.file().bersihkanBaru(m.Foto, "")
		}
	}()
	in, err := ValidasiTestimoni(m)
	if err != nil {
		return Testimoni{}, err
	}
	if in.Foto != "" {
		if err := s.file().periksaKey(ctx, in.Foto); err != nil {
			return Testimoni{}, err
		}
	}
	t, err := s.Repo.Buat(ctx, in, oleh)
	if errors.Is(err, ErrGambarDipakai) {
		return Testimoni{}, salah("Foto sudah dipakai konten lain")
	}
	if err != nil {
		return Testimoni{}, err
	}
	berhasil = true
	return t, nil
}

// Ubah menyimpan perubahan draft. Foto kosong menghapus foto; foto berbeda menggantinya (file lama dihapus setelah commit).
func (s TestimoniLayanan) Ubah(ctx context.Context, id int64, m MentahTestimoni, version int32, oleh string) (hasil Testimoni, err error) {
	lama := ""
	berhasil := false
	defer func() {
		if !berhasil {
			s.file().bersihkanBaru(m.Foto, lama)
		}
	}()

	cur, err := s.Repo.Detail(ctx, id)
	if err != nil {
		return Testimoni{}, err
	}
	lama = cur.FotoKey
	if cur.DihapusPada != nil || cur.Status != "draft" {
		return Testimoni{}, ErrStatusTidakValid
	}
	if cur.Version != version {
		return Testimoni{}, ErrVersiBentrok
	}
	in, err := ValidasiTestimoni(m)
	if err != nil {
		return Testimoni{}, err
	}
	if in.Foto != "" && in.Foto != lama {
		if err := s.file().periksaKey(ctx, in.Foto); err != nil {
			return Testimoni{}, err
		}
	}
	t, err := s.Repo.Ubah(ctx, id, version, in, oleh)
	if errors.Is(err, ErrGambarDipakai) {
		return Testimoni{}, salah("Foto sudah dipakai konten lain")
	}
	if err != nil {
		return Testimoni{}, err
	}
	berhasil = true
	if in.Foto != lama {
		s.file().hapusLama(lama) // setelah commit (juga saat foto dihapus)
	}
	return t, nil
}

// Post menerbitkan draft. Bila ada foto, diperiksa ulang masih ada di storage.
func (s TestimoniLayanan) Post(ctx context.Context, id int64, version int32, oleh string) (Testimoni, error) {
	cur, err := s.Repo.Detail(ctx, id)
	if err != nil {
		return Testimoni{}, err
	}
	if cur.DihapusPada != nil || cur.Status != "draft" {
		return Testimoni{}, ErrStatusTidakValid
	}
	if cur.Version != version {
		return Testimoni{}, ErrVersiBentrok
	}
	if cur.FotoKey != "" {
		ada, err := s.file().ada(ctx, cur.FotoKey)
		if err != nil {
			return Testimoni{}, err
		}
		if !ada {
			return Testimoni{}, ErrGambarHilang
		}
	}
	return s.Repo.Post(ctx, id, version, oleh)
}

func (s TestimoniLayanan) Retract(ctx context.Context, id int64, version int32, oleh string) (Testimoni, error) {
	return s.Repo.Retract(ctx, id, version, oleh)
}

func (s TestimoniLayanan) Hapus(ctx context.Context, id int64, oleh string) error {
	return s.Repo.Hapus(ctx, id, oleh)
}

func (s TestimoniLayanan) Pulihkan(ctx context.Context, id int64, oleh string) (Testimoni, error) {
	return s.Repo.Pulihkan(ctx, id, oleh)
}

// HapusPermanen menghapus baris dari trash, lalu (setelah commit) file fotonya bila ada.
func (s TestimoniLayanan) HapusPermanen(ctx context.Context, id int64) error {
	key, err := s.Repo.HapusPermanen(ctx, id)
	if err != nil {
		return err
	}
	s.file().hapusLama(key)
	return nil
}
