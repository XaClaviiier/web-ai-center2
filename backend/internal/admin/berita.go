package admin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
)

// BeritaRepo adalah akses data berita untuk admin. Implementasi menjaga atomisitas dan aturan status lewat SQL;
// galat yang dikembalikan: ErrTidakDitemukan, ErrStatusTidakValid, ErrVersiBentrok, ErrGambarDipakai.
type BeritaRepo interface {
	Daftar(ctx context.Context, p ParamDaftar) ([]Berita, int64, error)
	Detail(ctx context.Context, id int64) (Berita, error) // semua status, termasuk trash
	// Buat menyisipkan draft baru; dibuat=false bila slug sudah dipakai (pemanggil mencoba slug lain).
	Buat(ctx context.Context, slug string, in InputBerita, oleh string) (b Berita, dibuat bool, err error)
	Ubah(ctx context.Context, id int64, version int32, in InputBerita, oleh string) (Berita, error)
	Post(ctx context.Context, id int64, version int32, oleh string) (Berita, error)
	Retract(ctx context.Context, id int64, version int32, oleh string) (Berita, error)
	Hapus(ctx context.Context, id int64, oleh string) error                    // soft delete (hanya draft)
	Pulihkan(ctx context.Context, id int64, oleh string) (Berita, error)       // trash -> draft
	HapusPermanen(ctx context.Context, id int64) (gambarKey string, err error) // hanya dari trash
	GambarDipakai(ctx context.Context, key string) (bool, error)
}

// BeritaLayanan menerapkan aturan bisnis berita di atas BeritaRepo dan penyimpanan gambar.
type BeritaLayanan struct {
	Repo  BeritaRepo
	Media media.Store
	Log   *slog.Logger
}

// ValidasiBerita memvalidasi dan merapikan isi berita dari klien (tanpa memeriksa key gambar ke storage).
func ValidasiBerita(m MentahBerita) (InputBerita, error) {
	judul, err := barisSatu("judul", m.Judul, BatasJudul)
	if err != nil {
		return InputBerita{}, err
	}
	isi, err := banyakBaris("isi", m.Isi, BatasIsi)
	if err != nil {
		return InputBerita{}, err
	}
	tgl, err := tanggal("tanggalPublikasi", m.TanggalPublikasi)
	if err != nil {
		return InputBerita{}, err
	}
	if m.Gambar == "" {
		return InputBerita{}, salah("Field gambar wajib diisi")
	}
	return InputBerita{Judul: judul, Isi: isi, TanggalPublikasi: tgl, Gambar: m.Gambar}, nil
}

func (s BeritaLayanan) Daftar(ctx context.Context, p ParamDaftar) ([]Berita, int64, error) {
	return s.Repo.Daftar(ctx, p)
}

func (s BeritaLayanan) Detail(ctx context.Context, id int64) (Berita, error) {
	return s.Repo.Detail(ctx, id)
}

// file adalah pengelola file gambar berita (aturan bersama di berkas.go).
func (s BeritaLayanan) file() pengelolaFile {
	return pengelolaFile{Media: s.Media, Log: s.Log, Sumber: media.SumberBerita, Dipakai: s.Repo.GambarDipakai}
}

func (s BeritaLayanan) periksaKey(ctx context.Context, key string) error {
	return s.file().periksaKey(ctx, key)
}

func (s BeritaLayanan) bersihkanBaru(key, lama string) { s.file().bersihkanBaru(key, lama) }

func (s BeritaLayanan) hapusLama(key string) { s.file().hapusLama(key) }

// Buat membuat berita baru berstatus draft. Slug dibuat dari judul; bentrok diberi akhiran -2, -3, dst.
func (s BeritaLayanan) Buat(ctx context.Context, m MentahBerita, oleh string) (hasil Berita, err error) {
	berhasil := false
	defer func() {
		if !berhasil {
			s.bersihkanBaru(m.Gambar, "")
		}
	}()

	in, err := ValidasiBerita(m)
	if err != nil {
		return Berita{}, err
	}
	dasar := Slug(in.Judul)
	if dasar == "" {
		return Berita{}, salah("Judul harus memuat huruf atau angka")
	}
	if err := s.periksaKey(ctx, in.Gambar); err != nil {
		return Berita{}, err
	}
	for n := 1; n <= 1000; n++ {
		slug := dasar
		if n > 1 {
			slug = dasar + "-" + strconv.Itoa(n)
		}
		b, dibuat, err := s.Repo.Buat(ctx, slug, in, oleh)
		if errors.Is(err, ErrGambarDipakai) {
			return Berita{}, salah("Gambar sudah dipakai konten lain")
		}
		if err != nil {
			return Berita{}, err
		}
		if dibuat {
			berhasil = true
			return b, nil
		}
	}
	return Berita{}, errors.New("tidak menemukan slug yang bebas setelah 1000 percobaan")
}

// Ubah menyimpan perubahan draft. Hanya untuk draft yang tidak di-trash, dan version harus cocok.
func (s BeritaLayanan) Ubah(ctx context.Context, id int64, m MentahBerita, version int32, oleh string) (hasil Berita, err error) {
	lama := ""
	berhasil := false
	defer func() {
		if !berhasil {
			s.bersihkanBaru(m.Gambar, lama)
		}
	}()

	cur, err := s.Repo.Detail(ctx, id)
	if err != nil {
		return Berita{}, err
	}
	lama = cur.GambarKey
	if cur.DihapusPada != nil || cur.Status != "draft" {
		return Berita{}, ErrStatusTidakValid
	}
	if cur.Version != version {
		return Berita{}, ErrVersiBentrok
	}
	in, err := ValidasiBerita(m)
	if err != nil {
		return Berita{}, err
	}
	if in.Gambar != lama {
		if err := s.periksaKey(ctx, in.Gambar); err != nil {
			return Berita{}, err
		}
	}
	b, err := s.Repo.Ubah(ctx, id, version, in, oleh)
	if errors.Is(err, ErrGambarDipakai) {
		return Berita{}, salah("Gambar sudah dipakai konten lain")
	}
	if err != nil {
		return Berita{}, err
	}
	berhasil = true
	if in.Gambar != lama {
		s.hapusLama(lama) // setelah commit
	}
	return b, nil
}

// Post menerbitkan draft. Gambar diperiksa ulang masih ada di storage.
func (s BeritaLayanan) Post(ctx context.Context, id int64, version int32, oleh string) (Berita, error) {
	cur, err := s.Repo.Detail(ctx, id)
	if err != nil {
		return Berita{}, err
	}
	if cur.DihapusPada != nil || cur.Status != "draft" {
		return Berita{}, ErrStatusTidakValid
	}
	if cur.Version != version {
		return Berita{}, ErrVersiBentrok
	}
	ada, err := media.Ada(ctx, s.Media, cur.GambarKey)
	if err != nil {
		return Berita{}, fmt.Errorf("%w: %v", ErrStorage, err)
	}
	if !ada {
		return Berita{}, ErrGambarHilang
	}
	return s.Repo.Post(ctx, id, version, oleh)
}

func (s BeritaLayanan) Retract(ctx context.Context, id int64, version int32, oleh string) (Berita, error) {
	return s.Repo.Retract(ctx, id, version, oleh)
}

// Hapus memindahkan draft ke trash (soft delete). Gambar tidak disentuh.
func (s BeritaLayanan) Hapus(ctx context.Context, id int64, oleh string) error {
	return s.Repo.Hapus(ctx, id, oleh)
}

// Pulihkan mengembalikan item trash menjadi draft.
func (s BeritaLayanan) Pulihkan(ctx context.Context, id int64, oleh string) (Berita, error) {
	return s.Repo.Pulihkan(ctx, id, oleh)
}

// HapusPermanen menghapus baris dari trash, lalu (setelah commit) file gambarnya.
func (s BeritaLayanan) HapusPermanen(ctx context.Context, id int64) error {
	key, err := s.Repo.HapusPermanen(ctx, id)
	if err != nil {
		return err
	}
	s.hapusLama(key)
	return nil
}
