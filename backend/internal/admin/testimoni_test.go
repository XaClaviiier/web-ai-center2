package admin

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
)

// repoTestimoniPalsu meniru aturan SQL testimoni (status, version, foto opsional dan unik) di memori.
type repoTestimoniPalsu struct {
	baris   map[int64]*Testimoni
	urut    int64
	dipakai map[string]int64
}

func baruRepoTestimoniPalsu() *repoTestimoniPalsu {
	return &repoTestimoniPalsu{baris: map[int64]*Testimoni{}, dipakai: map[string]int64{}}
}

func (r *repoTestimoniPalsu) Daftar(context.Context, string) ([]Testimoni, error) { return nil, nil }

func (r *repoTestimoniPalsu) Detail(_ context.Context, id int64) (Testimoni, error) {
	t, ok := r.baris[id]
	if !ok {
		return Testimoni{}, ErrTidakDitemukan
	}
	return *t, nil
}

func (r *repoTestimoniPalsu) pakai(foto string, id int64) error {
	if foto == "" {
		return nil
	}
	if pemilik, ada := r.dipakai[foto]; ada && pemilik != id {
		return ErrGambarDipakai
	}
	return nil
}

func (r *repoTestimoniPalsu) Buat(_ context.Context, in InputTestimoni, oleh string) (Testimoni, error) {
	if err := r.pakai(in.Foto, 0); err != nil {
		return Testimoni{}, err
	}
	r.urut++
	t := &Testimoni{ID: r.urut, Nama: in.Nama, Teks: in.Teks, Posisi: in.Posisi, FotoKey: in.Foto, Status: "draft", Version: 1,
		DibuatPada: time.Now(), DibuatOleh: oleh, DiubahOleh: oleh}
	r.baris[t.ID] = t
	if in.Foto != "" {
		r.dipakai[in.Foto] = t.ID
	}
	return *t, nil
}

func (r *repoTestimoniPalsu) draftAktif(id int64, version int32) (*Testimoni, error) {
	t, ok := r.baris[id]
	if !ok {
		return nil, ErrTidakDitemukan
	}
	if t.DihapusPada != nil || t.Status != "draft" {
		return nil, ErrStatusTidakValid
	}
	if t.Version != version {
		return nil, ErrVersiBentrok
	}
	return t, nil
}

func (r *repoTestimoniPalsu) Ubah(_ context.Context, id int64, version int32, in InputTestimoni, oleh string) (Testimoni, error) {
	t, err := r.draftAktif(id, version)
	if err != nil {
		return Testimoni{}, err
	}
	if err := r.pakai(in.Foto, id); err != nil {
		return Testimoni{}, err
	}
	delete(r.dipakai, t.FotoKey)
	if in.Foto != "" {
		r.dipakai[in.Foto] = id
	}
	t.Nama, t.Teks, t.Posisi, t.FotoKey, t.DiubahOleh = in.Nama, in.Teks, in.Posisi, in.Foto, oleh
	t.Version++
	return *t, nil
}

func (r *repoTestimoniPalsu) Post(_ context.Context, id int64, version int32, oleh string) (Testimoni, error) {
	t, err := r.draftAktif(id, version)
	if err != nil {
		return Testimoni{}, err
	}
	t.Status, t.DiubahOleh = "posted", oleh
	t.Version++
	return *t, nil
}

func (r *repoTestimoniPalsu) Retract(_ context.Context, id int64, version int32, oleh string) (Testimoni, error) {
	t, ok := r.baris[id]
	if !ok {
		return Testimoni{}, ErrTidakDitemukan
	}
	if t.DihapusPada != nil || t.Status != "posted" {
		return Testimoni{}, ErrStatusTidakValid
	}
	if t.Version != version {
		return Testimoni{}, ErrVersiBentrok
	}
	t.Status, t.DiubahOleh = "draft", oleh
	t.Version++
	return *t, nil
}

func (r *repoTestimoniPalsu) Hapus(_ context.Context, id int64, oleh string) error {
	t, ok := r.baris[id]
	if !ok {
		return ErrTidakDitemukan
	}
	if t.DihapusPada != nil || t.Status != "draft" {
		return ErrStatusTidakValid
	}
	d := time.Now()
	t.DihapusPada, t.DiubahOleh = &d, oleh
	return nil
}

func (r *repoTestimoniPalsu) Pulihkan(_ context.Context, id int64, oleh string) (Testimoni, error) {
	t, ok := r.baris[id]
	if !ok {
		return Testimoni{}, ErrTidakDitemukan
	}
	if t.DihapusPada == nil {
		return Testimoni{}, ErrStatusTidakValid
	}
	t.DihapusPada, t.DiubahOleh = nil, oleh
	return *t, nil
}

func (r *repoTestimoniPalsu) HapusPermanen(_ context.Context, id int64) (string, error) {
	t, ok := r.baris[id]
	if !ok {
		return "", ErrTidakDitemukan
	}
	if t.DihapusPada == nil {
		return "", ErrStatusTidakValid
	}
	delete(r.baris, id)
	delete(r.dipakai, t.FotoKey)
	return t.FotoKey, nil
}

func (r *repoTestimoniPalsu) FotoDipakai(_ context.Context, key string) (bool, error) {
	_, ada := r.dipakai[key]
	return ada, nil
}

type lingkunganTestimoni struct {
	svc  TestimoniLayanan
	repo *repoTestimoniPalsu
	mem  *media.Memori
}

func siapkanTestimoni(t *testing.T) lingkunganTestimoni {
	t.Helper()
	repo, mem := baruRepoTestimoniPalsu(), media.NewMemori()
	return lingkunganTestimoni{
		svc:  TestimoniLayanan{Repo: repo, Media: mem, Log: slog.New(slog.NewTextHandler(io.Discard, nil))},
		repo: repo, mem: mem,
	}
}

func (l lingkunganTestimoni) denganFoto(t *testing.T, nama string) MentahTestimoni {
	t.Helper()
	return MentahTestimoni{Nama: nama, Teks: "Pelatihannya sangat membantu.", Posisi: "Dosen", Foto: unggahKey(t, l.mem, media.SumberTestimoni)}
}

func tanpaFoto(nama string) MentahTestimoni {
	return MentahTestimoni{Nama: nama, Teks: "Bagus sekali.", Posisi: "Mahasiswa"}
}

func TestTestimoniFotoOpsional(t *testing.T) {
	l := siapkanTestimoni(t)
	a, err := l.svc.Buat(ctx, tanpaFoto("Ani"), "dev-admin")
	if err != nil {
		t.Fatal(err)
	}
	if a.FotoKey != "" || a.Status != "draft" || a.Version != 1 || a.DibuatOleh != "dev-admin" {
		t.Errorf("tanpa foto: %+v", a)
	}
	m := l.denganFoto(t, "Budi")
	b, err := l.svc.Buat(ctx, m, "u")
	if err != nil || b.FotoKey != m.Foto || !adaKey(t, l.mem, m.Foto) {
		t.Fatalf("dengan foto: %+v %v", b, err)
	}
	// storage mati tidak mengganggu pembuatan TANPA foto (tidak ada yang diperiksa)
	l.mem.Err = errors.New("rustfs mati")
	if _, err := l.svc.Buat(ctx, tanpaFoto("Cici"), "u"); err != nil {
		t.Errorf("tanpa foto tidak boleh bergantung pada storage: %v", err)
	}
}

func TestTestimoniValidasi(t *testing.T) {
	in, err := ValidasiTestimoni(MentahTestimoni{Nama: "  Ani   B ", Teks: "Baris 1\r\n\r\n\r\nBaris 2", Posisi: " Dosen "})
	if err != nil || in.Nama != "Ani B" || in.Teks != "Baris 1\n\nBaris 2" || in.Posisi != "Dosen" {
		t.Fatalf("perapian: %+v %v", in, err)
	}
	kasus := map[string]func(m *MentahTestimoni){
		"nama kosong":     func(m *MentahTestimoni) { m.Nama = "" },
		"nama panjang":    func(m *MentahTestimoni) { m.Nama = strings.Repeat("a", BatasNama+1) },
		"teks kosong":     func(m *MentahTestimoni) { m.Teks = "  " },
		"teks panjang":    func(m *MentahTestimoni) { m.Teks = strings.Repeat("a", BatasTeksTestimoni+1) },
		"posisi kosong":   func(m *MentahTestimoni) { m.Posisi = "" },
		"posisi panjang":  func(m *MentahTestimoni) { m.Posisi = strings.Repeat("a", BatasPosisi+1) },
		"kontrol di teks": func(m *MentahTestimoni) { m.Teks = "a\x00b" },
	}
	for nama, f := range kasus {
		m := MentahTestimoni{Nama: "A", Teks: "T", Posisi: "P"}
		f(&m)
		if _, err := ValidasiTestimoni(m); err == nil {
			t.Errorf("%s: seharusnya ditolak", nama)
		} else {
			perluErrInput(t, err, "")
		}
	}
	m := MentahTestimoni{Nama: "A", Teks: strings.Repeat("a", BatasTeksTestimoni), Posisi: "P"}
	if _, err := ValidasiTestimoni(m); err != nil {
		t.Errorf("teks tepat di batas harus diterima: %v", err)
	}
}

func TestTestimoniBuatFotoDitolakDanDibersihkan(t *testing.T) {
	l := siapkanTestimoni(t)
	keyTim := unggahKey(t, l.mem, media.SumberTim)
	m := tanpaFoto("Ani")
	m.Foto = keyTim
	perluErrInput(t, func() error { _, e := l.svc.Buat(ctx, m, "u"); return e }(), "")
	if !adaKey(t, l.mem, keyTim) {
		t.Error("foto sumber lain tidak boleh terhapus")
	}
	// validasi gagal dengan foto testimoni sah: file baru dihapus kembali
	g := l.denganFoto(t, "")
	if _, err := l.svc.Buat(ctx, g, "u"); err == nil {
		t.Fatal("seharusnya galat")
	}
	if adaKey(t, l.mem, g.Foto) {
		t.Error("file baru yang gagal disimpan harus dihapus kembali")
	}
}

func TestTestimoniUbahFotoTigaKeadaan(t *testing.T) {
	l := siapkanTestimoni(t)
	m := l.denganFoto(t, "Ani")
	a, _ := l.svc.Buat(ctx, m, "u")

	// 1. ganti foto: lama dihapus setelah commit
	baru := l.denganFoto(t, "Ani")
	u, err := l.svc.Ubah(ctx, a.ID, baru, 1, "u")
	if err != nil || u.FotoKey != baru.Foto || u.Version != 2 {
		t.Fatalf("ganti: %+v %v", u, err)
	}
	if adaKey(t, l.mem, m.Foto) || !adaKey(t, l.mem, baru.Foto) {
		t.Error("foto lama harus dihapus dan foto baru tersimpan")
	}
	// 2. foto sama: tidak ada yang dihapus
	if _, err := l.svc.Ubah(ctx, a.ID, baru, 2, "u"); err != nil || !adaKey(t, l.mem, baru.Foto) {
		t.Fatalf("foto sama: %v", err)
	}
	// 3. hapus foto (kosong): file lama dihapus setelah commit, tampil inisial
	tanpa := baru
	tanpa.Foto = ""
	u3, err := l.svc.Ubah(ctx, a.ID, tanpa, 3, "u")
	if err != nil || u3.FotoKey != "" {
		t.Fatalf("hapus foto: %+v %v", u3, err)
	}
	if adaKey(t, l.mem, baru.Foto) {
		t.Error("file foto yang dihapus harus ikut terhapus dari storage")
	}
	// 4. tambah foto pada testimoni tanpa foto
	tambah := l.denganFoto(t, "Ani")
	u4, err := l.svc.Ubah(ctx, a.ID, tambah, 4, "u")
	if err != nil || u4.FotoKey != tambah.Foto || !adaKey(t, l.mem, tambah.Foto) {
		t.Fatalf("tambah foto: %+v %v", u4, err)
	}
}

func TestTestimoniUbahGagalMenghapusFileBaru(t *testing.T) {
	l := siapkanTestimoni(t)
	a, _ := l.svc.Buat(ctx, l.denganFoto(t, "Ani"), "u")
	baru := l.denganFoto(t, "Ani")
	if _, err := l.svc.Ubah(ctx, a.ID, baru, 99, "u"); !errors.Is(err, ErrVersiBentrok) {
		t.Fatalf("version basi: %v", err)
	}
	if adaKey(t, l.mem, baru.Foto) {
		t.Error("file baru yang gagal disimpan harus dihapus")
	}
	if !adaKey(t, l.mem, a.FotoKey) {
		t.Error("foto lama tidak boleh tersentuh")
	}
}

func TestTestimoniSiklusStatus(t *testing.T) {
	l := siapkanTestimoni(t)
	a, _ := l.svc.Buat(ctx, tanpaFoto("Ani"), "u")

	p, err := l.svc.Post(ctx, a.ID, 1, "u")
	if err != nil || p.Status != "posted" || p.Version != 2 {
		t.Fatalf("Post: %+v %v", p, err)
	}
	if _, err := l.svc.Ubah(ctx, a.ID, tanpaFoto("x"), 2, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("edit posted: %v", err)
	}
	if err := l.svc.Hapus(ctx, a.ID, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("hapus posted: %v", err)
	}
	if _, err := l.svc.Retract(ctx, a.ID, 1, "u"); !errors.Is(err, ErrVersiBentrok) {
		t.Errorf("retract version basi: %v", err)
	}
	r, err := l.svc.Retract(ctx, a.ID, 2, "u")
	if err != nil || r.Status != "draft" || r.Version != 3 {
		t.Fatalf("Retract: %+v %v", r, err)
	}
	if err := l.svc.Hapus(ctx, a.ID, "u"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.svc.Ubah(ctx, a.ID, tanpaFoto("x"), 3, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("edit trash: %v", err)
	}
	pl, err := l.svc.Pulihkan(ctx, a.ID, "u")
	if err != nil || pl.Status != "draft" || pl.DihapusPada != nil || pl.Version != 3 {
		t.Fatalf("Pulihkan: %+v %v", pl, err)
	}
	if _, err := l.svc.Pulihkan(ctx, a.ID, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("pulihkan non-trash: %v", err)
	}
}

func TestTestimoniPostPemeriksaanFoto(t *testing.T) {
	l := siapkanTestimoni(t)
	// tanpa foto: Post tidak menyentuh storage sama sekali
	a, _ := l.svc.Buat(ctx, tanpaFoto("Ani"), "u")
	l.mem.Err = errors.New("rustfs mati")
	if _, err := l.svc.Post(ctx, a.ID, 1, "u"); err != nil {
		t.Fatalf("Post tanpa foto tidak boleh bergantung pada storage: %v", err)
	}
	l.mem.Err = nil

	// dengan foto yang hilang dari bucket
	b, _ := l.svc.Buat(ctx, l.denganFoto(t, "Budi"), "u")
	if err := l.mem.Hapus(ctx, b.FotoKey); err != nil {
		t.Fatal(err)
	}
	if _, err := l.svc.Post(ctx, b.ID, 1, "u"); !errors.Is(err, ErrGambarHilang) {
		t.Fatalf("diharapkan ErrGambarHilang, dapat %v", err)
	}
	// storage bermasalah bukan "gambar hilang"
	c, _ := l.svc.Buat(ctx, l.denganFoto(t, "Cici"), "u")
	l.mem.Err = errors.New("rustfs mati")
	if _, err := l.svc.Post(ctx, c.ID, 1, "u"); !errors.Is(err, ErrStorage) || errors.Is(err, ErrGambarHilang) {
		t.Fatalf("diharapkan ErrStorage saja, dapat %v", err)
	}
}

func TestTestimoniHapusPermanen(t *testing.T) {
	l := siapkanTestimoni(t)
	denga, _ := l.svc.Buat(ctx, l.denganFoto(t, "Ani"), "u")
	tanpa, _ := l.svc.Buat(ctx, tanpaFoto("Budi"), "u")

	if err := l.svc.HapusPermanen(ctx, denga.ID); !errors.Is(err, ErrStatusTidakValid) {
		t.Fatalf("non-trash: %v", err)
	}
	if !adaKey(t, l.mem, denga.FotoKey) {
		t.Fatal("penolakan tidak boleh menghapus file")
	}
	_ = l.svc.Hapus(ctx, denga.ID, "u")
	if !adaKey(t, l.mem, denga.FotoKey) {
		t.Fatal("soft delete tidak boleh menyentuh foto")
	}
	if err := l.svc.HapusPermanen(ctx, denga.ID); err != nil {
		t.Fatal(err)
	}
	if adaKey(t, l.mem, denga.FotoKey) {
		t.Error("hapus permanen harus menghapus foto")
	}
	// tanpa foto: tidak ada operasi file (storage mati tidak mengganggu)
	_ = l.svc.Hapus(ctx, tanpa.ID, "u")
	l.mem.Err = errors.New("rustfs mati")
	if err := l.svc.HapusPermanen(ctx, tanpa.ID); err != nil {
		t.Errorf("tanpa foto tidak boleh gagal karena storage: %v", err)
	}
	if err := l.svc.HapusPermanen(ctx, 999); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("tidak ada: %v", err)
	}
}
