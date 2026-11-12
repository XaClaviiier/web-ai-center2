package admin

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
)

// Batas dokumentasi (database.md §3.2, §5.1, requirement.md §5.2).
const (
	MaksGambarDokumentasi = 6 // jumlah gambar per dokumentasi: 1..6
)

// GambarDokumentasi adalah satu gambar dokumentasi (key RustFS). Teks alt tidak disimpan: frontend memakai judul.
type GambarDokumentasi struct {
	Key string
}

// Dokumentasi adalah satu dokumentasi lengkap untuk admin (semua status, termasuk trash).
type Dokumentasi struct {
	ID               int64
	Slug             string
	Judul            string
	TanggalPublikasi string              // YYYY-MM-DD
	Gambar           []GambarDokumentasi // berurutan; indeks 0 = headline (urutan 1)
	Status           string              // draft | posted
	Version          int32
	DibuatPada       time.Time
	DibuatOleh       string
	DiubahOleh       string
	DihapusPada      *time.Time
}

// RingkasDokumentasi adalah item daftar admin: tanpa seluruh gambar, hanya headline dan jumlahnya.
type RingkasDokumentasi struct {
	ID               int64
	Slug             string
	Judul            string
	TanggalPublikasi string
	GambarUtamaKey   string // kosong bila tidak ada gambar
	JumlahGambar     int
	Status           string
	Version          int32
	DibuatPada       time.Time
	DibuatOleh       string
	DiubahOleh       string
	DihapusPada      *time.Time
}

// MentahDokumentasi adalah isi dokumentasi apa adanya dari klien. Gambar adalah KEADAAN AKHIR berurutan
// (bukan daftar operasi): urutan array menjadi urutan gambar 1..n.
type MentahDokumentasi struct {
	Judul            string
	TanggalPublikasi string
	Gambar           []GambarDokumentasi
}

// InputDokumentasi adalah isi yang sudah divalidasi dan dirapikan.
type InputDokumentasi struct {
	Judul            string
	TanggalPublikasi string
	Gambar           []GambarDokumentasi
}

// DokumentasiRepo adalah akses data dokumentasi untuk admin. Buat dan Ubah menjaga atomisitas dokumentasi beserta
// seluruh gambarnya dalam SATU transaksi. Galat: ErrTidakDitemukan, ErrStatusTidakValid, ErrVersiBentrok, ErrGambarDipakai.
type DokumentasiRepo interface {
	Daftar(ctx context.Context, p ParamDaftar) ([]RingkasDokumentasi, int64, error)
	Detail(ctx context.Context, id int64) (Dokumentasi, error) // semua status, termasuk trash
	// Buat menyisipkan draft baru beserta gambarnya; dibuat=false bila slug sudah dipakai (tidak ada yang tersimpan).
	Buat(ctx context.Context, slug string, in InputDokumentasi, oleh string) (d Dokumentasi, dibuat bool, err error)
	// Ubah menyimpan judul, tanggal, dan menyelaraskan gambar ke daftar akhir: key lama dipertahankan (urutan dan alt
	// diperbarui), key baru ditambahkan, key yang hilang dari daftar dihapus. Semua atau tidak sama sekali.
	Ubah(ctx context.Context, id int64, version int32, in InputDokumentasi, oleh string) (Dokumentasi, error)
	Post(ctx context.Context, id int64, version int32, oleh string) (Dokumentasi, error)
	Retract(ctx context.Context, id int64, version int32, oleh string) (Dokumentasi, error)
	Hapus(ctx context.Context, id int64, oleh string) error // soft delete (hanya draft)
	Pulihkan(ctx context.Context, id int64, oleh string) (Dokumentasi, error)
	// HapusPermanen menghapus baris dokumentasi dan seluruh baris gambarnya dari trash, dan mengembalikan semua key-nya.
	HapusPermanen(ctx context.Context, id int64) (keys []string, err error)
	GambarDipakai(ctx context.Context, key string) (bool, error)
}

// DokumentasiLayanan menerapkan aturan bisnis dokumentasi di atas DokumentasiRepo dan penyimpanan gambar.
type DokumentasiLayanan struct {
	Repo  DokumentasiRepo
	Media media.Store
	Log   *slog.Logger
}

// ValidasiDokumentasi memvalidasi dan merapikan isi dokumentasi (tanpa memeriksa key gambar ke storage).
func ValidasiDokumentasi(m MentahDokumentasi) (InputDokumentasi, error) { return validasiDokumentasi(m, "") }

// validasiDokumentasi seperti ValidasiDokumentasi; judulLama (Simpan) mengizinkan judul lama >BatasJudul yang tidak diubah.
func validasiDokumentasi(m MentahDokumentasi, judulLama string) (InputDokumentasi, error) {
	judul, err := judulBoleh("judul", m.Judul, judulLama)
	if err != nil {
		return InputDokumentasi{}, err
	}
	tgl, err := tanggal("tanggalPublikasi", m.TanggalPublikasi)
	if err != nil {
		return InputDokumentasi{}, err
	}
	if n := len(m.Gambar); n < 1 || n > MaksGambarDokumentasi {
		return InputDokumentasi{}, salah("Field gambar harus berisi 1 sampai " + strconv.Itoa(MaksGambarDokumentasi) + " gambar")
	}
	sudah := make(map[string]struct{}, len(m.Gambar))
	gambar := make([]GambarDokumentasi, 0, len(m.Gambar))
	for i, g := range m.Gambar {
		if g.Key == "" {
			return InputDokumentasi{}, salah("Field gambar[" + strconv.Itoa(i+1) + "].key wajib diisi")
		}
		if _, ada := sudah[g.Key]; ada {
			return InputDokumentasi{}, salah("Field gambar memuat key yang sama lebih dari sekali")
		}
		sudah[g.Key] = struct{}{}
		gambar = append(gambar, GambarDokumentasi{Key: g.Key})
	}
	return InputDokumentasi{Judul: judul, TanggalPublikasi: tgl, Gambar: gambar}, nil
}

func (s DokumentasiLayanan) file() pengelolaFile {
	return pengelolaFile{Media: s.Media, Log: s.Log, Sumber: media.SumberDokumentasi, Dipakai: s.Repo.GambarDipakai}
}

func (s DokumentasiLayanan) Daftar(ctx context.Context, p ParamDaftar) ([]RingkasDokumentasi, int64, error) {
	return s.Repo.Daftar(ctx, p)
}

func (s DokumentasiLayanan) Detail(ctx context.Context, id int64) (Dokumentasi, error) {
	return s.Repo.Detail(ctx, id)
}

func kunciSet(g []GambarDokumentasi) map[string]struct{} {
	m := make(map[string]struct{}, len(g))
	for _, x := range g {
		m[x.Key] = struct{}{}
	}
	return m
}

// bersihkanBaru menghapus file baru yang gagal disimpan (database.md §5.1); key yang sudah dimiliki dokumentasi ini
// (lama) tidak pernah dihapus, dan pengaman pengelolaFile berlaku untuk tiap key.
func (s DokumentasiLayanan) bersihkanBaru(kandidat []GambarDokumentasi, lama map[string]struct{}) {
	for _, g := range kandidat {
		if _, punya := lama[g.Key]; punya {
			continue
		}
		s.file().bersihkanBaru(g.Key, "")
	}
}

// periksaKeyBaru memeriksa key yang belum dimiliki dokumentasi ini (bentuk, sumber, ada di storage).
func (s DokumentasiLayanan) periksaKeyBaru(ctx context.Context, gambar []GambarDokumentasi, lama map[string]struct{}) error {
	for _, g := range gambar {
		if _, punya := lama[g.Key]; punya {
			continue
		}
		if err := s.file().periksaKey(ctx, g.Key); err != nil {
			return err
		}
	}
	return nil
}

// Buat membuat dokumentasi baru berstatus draft dengan 1-6 gambar. Slug dari judul; bentrok diberi akhiran -2, -3, dst.
func (s DokumentasiLayanan) Buat(ctx context.Context, m MentahDokumentasi, oleh string) (hasil Dokumentasi, err error) {
	berhasil := false
	defer func() {
		if !berhasil {
			s.bersihkanBaru(m.Gambar, nil)
		}
	}()

	in, err := ValidasiDokumentasi(m)
	if err != nil {
		return Dokumentasi{}, err
	}
	dasar := Slug(in.Judul)
	if dasar == "" {
		return Dokumentasi{}, salah("Judul harus memuat huruf atau angka")
	}
	if err := s.periksaKeyBaru(ctx, in.Gambar, nil); err != nil {
		return Dokumentasi{}, err
	}
	for n := 1; n <= 1000; n++ {
		slug := dasar
		if n > 1 {
			slug = dasar + "-" + strconv.Itoa(n)
		}
		d, dibuat, err := s.Repo.Buat(ctx, slug, in, oleh)
		if errors.Is(err, ErrGambarDipakai) {
			return Dokumentasi{}, salah("Salah satu gambar sudah dipakai konten lain")
		}
		if err != nil {
			return Dokumentasi{}, err
		}
		if dibuat {
			berhasil = true
			return d, nil
		}
	}
	return Dokumentasi{}, errors.New("tidak menemukan slug yang bebas setelah 1000 percobaan")
}

// Ubah menyimpan perubahan draft dan menyelaraskan gambar ke daftar akhir (replace, hapus, tambah, atau ganti urutan).
// File yang tidak lagi dipakai dihapus SETELAH commit; file baru yang gagal disimpan dihapus kembali.
func (s DokumentasiLayanan) Ubah(ctx context.Context, id int64, m MentahDokumentasi, version int32, oleh string) (hasil Dokumentasi, err error) {
	var lama map[string]struct{}
	berhasil := false
	defer func() {
		if !berhasil {
			s.bersihkanBaru(m.Gambar, lama)
		}
	}()

	cur, err := s.Repo.Detail(ctx, id)
	if err != nil {
		return Dokumentasi{}, err
	}
	lama = kunciSet(cur.Gambar)
	if cur.DihapusPada != nil || cur.Status != "draft" {
		return Dokumentasi{}, ErrStatusTidakValid
	}
	if cur.Version != version {
		return Dokumentasi{}, ErrVersiBentrok
	}
	in, err := validasiDokumentasi(m, cur.Judul)
	if err != nil {
		return Dokumentasi{}, err
	}
	if err := s.periksaKeyBaru(ctx, in.Gambar, lama); err != nil {
		return Dokumentasi{}, err
	}
	d, err := s.Repo.Ubah(ctx, id, version, in, oleh)
	if errors.Is(err, ErrGambarDipakai) {
		return Dokumentasi{}, salah("Salah satu gambar sudah dipakai konten lain")
	}
	if err != nil {
		return Dokumentasi{}, err
	}
	berhasil = true
	baru := kunciSet(in.Gambar)
	for _, g := range cur.Gambar {
		if _, masihAda := baru[g.Key]; !masihAda {
			s.file().hapusLama(g.Key) // setelah commit
		}
	}
	return d, nil
}

// Post menerbitkan draft. Semua gambar (maksimum 6) diperiksa ulang masih ada di storage.
func (s DokumentasiLayanan) Post(ctx context.Context, id int64, version int32, oleh string) (Dokumentasi, error) {
	cur, err := s.Repo.Detail(ctx, id)
	if err != nil {
		return Dokumentasi{}, err
	}
	if cur.DihapusPada != nil || cur.Status != "draft" {
		return Dokumentasi{}, ErrStatusTidakValid
	}
	if cur.Version != version {
		return Dokumentasi{}, ErrVersiBentrok
	}
	for _, g := range cur.Gambar {
		ada, err := s.file().ada(ctx, g.Key)
		if err != nil {
			return Dokumentasi{}, err
		}
		if !ada {
			return Dokumentasi{}, ErrGambarHilang
		}
	}
	return s.Repo.Post(ctx, id, version, oleh)
}

func (s DokumentasiLayanan) Retract(ctx context.Context, id int64, version int32, oleh string) (Dokumentasi, error) {
	return s.Repo.Retract(ctx, id, version, oleh)
}

// Hapus memindahkan draft ke trash (soft delete). Gambar tidak disentuh.
func (s DokumentasiLayanan) Hapus(ctx context.Context, id int64, oleh string) error {
	return s.Repo.Hapus(ctx, id, oleh)
}

// Pulihkan mengembalikan item trash menjadi draft.
func (s DokumentasiLayanan) Pulihkan(ctx context.Context, id int64, oleh string) (Dokumentasi, error) {
	return s.Repo.Pulihkan(ctx, id, oleh)
}

// HapusPermanen menghapus baris dari trash, lalu (setelah commit) SEMUA file gambarnya.
func (s DokumentasiLayanan) HapusPermanen(ctx context.Context, id int64) error {
	keys, err := s.Repo.HapusPermanen(ctx, id)
	if err != nil {
		return err
	}
	for _, k := range keys {
		s.file().hapusLama(k)
	}
	return nil
}
