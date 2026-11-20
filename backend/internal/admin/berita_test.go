package admin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
)

// repoPalsu meniru aturan SQL (status, version, slug unik, gambar unik) di memori. Semantik SQL sebenarnya diuji di paket store.
type repoPalsu struct {
	baris   map[int64]*Berita
	urut    int64
	err     error // diteruskan ke semua operasi bila diisi
	dipakai map[string]int64
}

func baruRepoPalsu() *repoPalsu {
	return &repoPalsu{baris: map[int64]*Berita{}, dipakai: map[string]int64{}}
}

func (r *repoPalsu) Daftar(context.Context, ParamDaftar) ([]Berita, int64, error) {
	return nil, 0, r.err
}

func (r *repoPalsu) Detail(_ context.Context, id int64) (Berita, error) {
	if r.err != nil {
		return Berita{}, r.err
	}
	b, ok := r.baris[id]
	if !ok {
		return Berita{}, ErrTidakDitemukan
	}
	return *b, nil
}

func (r *repoPalsu) Buat(_ context.Context, slug string, in InputBerita, oleh string) (Berita, bool, error) {
	if r.err != nil {
		return Berita{}, false, r.err
	}
	for _, b := range r.baris {
		if b.Slug == slug {
			return Berita{}, false, nil
		}
	}
	if _, ada := r.dipakai[in.Gambar]; ada {
		return Berita{}, false, ErrGambarDipakai
	}
	r.urut++
	b := &Berita{ID: r.urut, Slug: slug, Judul: in.Judul, Isi: in.Isi, TanggalPublikasi: in.TanggalPublikasi,
		GambarKey: in.Gambar, Status: "draft", Version: 1, DibuatPada: time.Now(), DibuatOleh: oleh, DiubahOleh: oleh}
	r.baris[b.ID] = b
	r.dipakai[in.Gambar] = b.ID
	return *b, true, nil
}

func (r *repoPalsu) draftAktif(id int64, version int32) (*Berita, error) {
	b, ok := r.baris[id]
	if !ok {
		return nil, ErrTidakDitemukan
	}
	if b.DihapusPada != nil || b.Status != "draft" {
		return nil, ErrStatusTidakValid
	}
	if b.Version != version {
		return nil, ErrVersiBentrok
	}
	return b, nil
}

func (r *repoPalsu) Ubah(_ context.Context, id int64, version int32, in InputBerita, oleh string) (Berita, error) {
	if r.err != nil {
		return Berita{}, r.err
	}
	b, err := r.draftAktif(id, version)
	if err != nil {
		return Berita{}, err
	}
	if pemilik, ada := r.dipakai[in.Gambar]; ada && pemilik != id {
		return Berita{}, ErrGambarDipakai
	}
	delete(r.dipakai, b.GambarKey)
	r.dipakai[in.Gambar] = id
	b.Judul, b.Isi, b.TanggalPublikasi, b.GambarKey, b.DiubahOleh = in.Judul, in.Isi, in.TanggalPublikasi, in.Gambar, oleh
	b.Version++
	return *b, nil
}

func (r *repoPalsu) Post(_ context.Context, id int64, version int32, oleh string) (Berita, error) {
	b, err := r.draftAktif(id, version)
	if err != nil {
		return Berita{}, err
	}
	b.Status, b.DiubahOleh = "posted", oleh
	b.Version++
	return *b, nil
}

func (r *repoPalsu) Retract(_ context.Context, id int64, version int32, oleh string) (Berita, error) {
	b, ok := r.baris[id]
	if !ok {
		return Berita{}, ErrTidakDitemukan
	}
	if b.DihapusPada != nil || b.Status != "posted" {
		return Berita{}, ErrStatusTidakValid
	}
	if b.Version != version {
		return Berita{}, ErrVersiBentrok
	}
	b.Status, b.DiubahOleh = "draft", oleh
	b.Version++
	return *b, nil
}

func (r *repoPalsu) Hapus(_ context.Context, id int64, oleh string) error {
	b, ok := r.baris[id]
	if !ok {
		return ErrTidakDitemukan
	}
	if b.DihapusPada != nil || b.Status != "draft" {
		return ErrStatusTidakValid
	}
	t := time.Now()
	b.DihapusPada, b.DiubahOleh = &t, oleh
	return nil
}

func (r *repoPalsu) Pulihkan(_ context.Context, id int64, oleh string) (Berita, error) {
	b, ok := r.baris[id]
	if !ok {
		return Berita{}, ErrTidakDitemukan
	}
	if b.DihapusPada == nil {
		return Berita{}, ErrStatusTidakValid
	}
	b.DihapusPada, b.DiubahOleh = nil, oleh
	return *b, nil
}

func (r *repoPalsu) HapusPermanen(_ context.Context, id int64) (string, error) {
	b, ok := r.baris[id]
	if !ok {
		return "", ErrTidakDitemukan
	}
	if b.DihapusPada == nil {
		return "", ErrStatusTidakValid
	}
	delete(r.baris, id)
	delete(r.dipakai, b.GambarKey)
	return b.GambarKey, nil
}

func (r *repoPalsu) GambarDipakai(_ context.Context, key string) (bool, error) {
	_, ada := r.dipakai[key]
	return ada, r.err
}

type lingkungan struct {
	svc  BeritaLayanan
	repo *repoPalsu
	mem  *media.Memori
}

func siapkan(t *testing.T) lingkungan {
	t.Helper()
	repo, mem := baruRepoPalsu(), media.NewMemori()
	return lingkungan{
		svc:  BeritaLayanan{Repo: repo, Media: mem, Log: slog.New(slog.NewTextHandler(io.Discard, nil))},
		repo: repo, mem: mem,
	}
}

var png = []byte("\x89PNG\r\n\x1a\n" + "data")

// unggah menaruh gambar di storage palsu dan mengembalikan key-nya.
func (l lingkungan) unggah(t *testing.T, sumber string) string {
	t.Helper()
	key, err := media.BuatKey(sumber, "png")
	if err != nil {
		t.Fatal(err)
	}
	if err := l.mem.Put(context.Background(), key, png, "image/png"); err != nil {
		t.Fatal(err)
	}
	return key
}

func (l lingkungan) mentah(t *testing.T, judul string) MentahBerita {
	t.Helper()
	return MentahBerita{Judul: judul, Isi: "Paragraf satu.\n\nParagraf dua.", TanggalPublikasi: "09-11-2026", Gambar: l.unggah(t, media.SumberBerita)}
}

func adaFile(t *testing.T, l lingkungan, key string) bool {
	t.Helper()
	ada, err := media.Ada(context.Background(), l.mem, key)
	if err != nil {
		t.Fatal(err)
	}
	return ada
}

func perluErrInput(t *testing.T, err error, mengandung string) {
	t.Helper()
	var e *ErrInput
	if !errors.As(err, &e) {
		t.Fatalf("diharapkan ErrInput, dapat %v", err)
	}
	if mengandung != "" && !strings.Contains(e.Pesan, mengandung) {
		t.Fatalf("pesan %q tidak memuat %q", e.Pesan, mengandung)
	}
}

var ctx = context.Background()

func TestBuatBerhasil(t *testing.T) {
	l := siapkan(t)
	m := l.mentah(t, "Pelatihan AI untuk Dosen")
	b, err := l.svc.Buat(ctx, m, "dev-admin")
	if err != nil {
		t.Fatal(err)
	}
	if b.Slug != "pelatihan-ai-untuk-dosen" || b.Status != "draft" || b.Version != 1 || b.DibuatOleh != "dev-admin" || b.GambarKey != m.Gambar {
		t.Errorf("hasil salah: %+v", b)
	}
	if !adaFile(t, l, m.Gambar) {
		t.Error("gambar tidak boleh terhapus saat berhasil")
	}
}

func TestBuatSlugBentrokDiberiAkhiran(t *testing.T) {
	l := siapkan(t)
	var slugs []string
	for i := 0; i < 3; i++ {
		b, err := l.svc.Buat(ctx, l.mentah(t, "Judul Sama"), "u")
		if err != nil {
			t.Fatal(err)
		}
		slugs = append(slugs, b.Slug)
	}
	if strings.Join(slugs, ",") != "judul-sama,judul-sama-2,judul-sama-3" {
		t.Errorf("slug: %v", slugs)
	}
}

func TestBuatValidasiMenolak(t *testing.T) {
	l := siapkan(t)
	kasus := map[string]func(m *MentahBerita){
		"judul kosong":      func(m *MentahBerita) { m.Judul = "" },
		"judul tanpa huruf": func(m *MentahBerita) { m.Judul = "!!! ???" },
		"tanggal salah":     func(m *MentahBerita) { m.TanggalPublikasi = "x" },
	}
	for nama, f := range kasus {
		m := l.mentah(t, "Judul")
		f(&m)
		if _, err := l.svc.Buat(ctx, m, "u"); err == nil {
			t.Errorf("%s: seharusnya galat", nama)
		}
	}
	if len(l.repo.baris) != 0 {
		t.Error("tidak boleh ada baris tercipta")
	}
}

func TestBuatKeyGambarDitolak(t *testing.T) {
	l := siapkan(t)
	m := l.mentah(t, "Judul")
	keyTim := l.unggah(t, media.SumberTim)

	kasus := map[string]string{
		"key sembarang":      "foto.png",
		"traversal":          "../berita/x.png",
		"sumber lain (tim)":  keyTim,
		"bentuk sah tak ada": "berita/3f9c2a10-7b4e-4c6d-9a1f-0123456789ab.png",
		"ekstensi asing":     strings.Replace(m.Gambar, ".png", ".gif", 1),
	}
	for nama, key := range kasus {
		mm := m
		mm.Gambar = key
		_, err := l.svc.Buat(ctx, mm, "u")
		perluErrInput(t, err, "")
		if nama == "bentuk sah tak ada" {
			perluErrInput(t, err, "belum diunggah")
		}
	}
	if len(l.repo.baris) != 0 {
		t.Error("tidak boleh ada baris tercipta")
	}
	if !adaFile(t, l, keyTim) {
		t.Error("gambar sumber lain tidak boleh ikut terhapus")
	}
}

func TestBuatGagalMenghapusFileBaruTapiAmanBagiYangDipakai(t *testing.T) {
	l := siapkan(t)
	// 1. gagal validasi teks: file baru yang sah dan tidak dipakai dihapus kembali (database.md §5.2)
	m := l.mentah(t, "")
	if _, err := l.svc.Buat(ctx, m, "u"); err == nil {
		t.Fatal("seharusnya galat")
	}
	if adaFile(t, l, m.Gambar) {
		t.Error("file baru yang gagal disimpan harus dihapus kembali")
	}

	// 2. berhasil membuat; lalu request lain memakai key yang sama dan gagal: key milik berita pertama TIDAK boleh terhapus
	pertama, err := l.svc.Buat(ctx, l.mentah(t, "Pertama"), "u")
	if err != nil {
		t.Fatal(err)
	}
	dup := MentahBerita{Judul: "Kedua", Isi: "isi", TanggalPublikasi: "09-11-2026", Gambar: pertama.GambarKey}
	_, err = l.svc.Buat(ctx, dup, "u")
	perluErrInput(t, err, "sudah dipakai")
	if !adaFile(t, l, pertama.GambarKey) {
		t.Fatal("file milik berita lain tidak boleh terhapus oleh request yang gagal")
	}

	// 3. key tidak sah tidak pernah dihapus (tidak ada yang dihapus selain key bentukan sendiri)
	asing := "berita/3f9c2a10-7b4e-4c6d-9a1f-0123456789ab.gif"
	if err := l.mem.Put(ctx, asing, png, "image/gif"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.svc.Buat(ctx, MentahBerita{Judul: "x", Isi: "y", TanggalPublikasi: "09-11-2026", Gambar: asing}, "u"); err == nil {
		t.Fatal("seharusnya galat")
	}
	if !adaFile(t, l, asing) {
		t.Error("key berbentuk tidak sah tidak boleh dihapus")
	}
}

func TestBuatStorageBermasalah(t *testing.T) {
	l := siapkan(t)
	m := l.mentah(t, "Judul")
	l.mem.Err = errors.New("rustfs mati")
	_, err := l.svc.Buat(ctx, m, "u")
	if !errors.Is(err, ErrStorage) {
		t.Fatalf("diharapkan ErrStorage, dapat %v", err)
	}
}

func TestUbahBerhasilDanGantiGambar(t *testing.T) {
	l := siapkan(t)
	b, _ := l.svc.Buat(ctx, l.mentah(t, "Judul Awal"), "u1")
	lama := b.GambarKey
	baru := l.mentah(t, "Judul Baru")

	u, err := l.svc.Ubah(ctx, b.ID, baru, b.Version, "u2")
	if err != nil {
		t.Fatal(err)
	}
	if u.Judul != "Judul Baru" || u.Version != b.Version+1 || u.DiubahOleh != "u2" || u.GambarKey != baru.Gambar {
		t.Errorf("hasil salah: %+v", u)
	}
	if u.Slug != b.Slug {
		t.Errorf("slug tidak boleh berubah: %q -> %q", b.Slug, u.Slug)
	}
	if adaFile(t, l, lama) {
		t.Error("gambar lama harus dihapus setelah commit")
	}
	if !adaFile(t, l, baru.Gambar) {
		t.Error("gambar baru harus tersimpan")
	}
}

func TestUbahTanpaGantiGambarTidakMenghapusApaPun(t *testing.T) {
	l := siapkan(t)
	b, _ := l.svc.Buat(ctx, l.mentah(t, "Judul"), "u")
	m := MentahBerita{Judul: "Judul Diedit", Isi: "isi baru", TanggalPublikasi: "10-11-2026", Gambar: b.GambarKey}
	u, err := l.svc.Ubah(ctx, b.ID, m, b.Version, "u")
	if err != nil {
		t.Fatal(err)
	}
	if u.GambarKey != b.GambarKey || !adaFile(t, l, b.GambarKey) {
		t.Error("gambar yang sama harus tetap ada")
	}
}

func TestUbahVersiBentrokMenghapusFileBaru(t *testing.T) {
	l := siapkan(t)
	b, _ := l.svc.Buat(ctx, l.mentah(t, "Judul"), "u")
	baru := l.mentah(t, "Judul")
	_, err := l.svc.Ubah(ctx, b.ID, baru, b.Version+5, "u")
	if !errors.Is(err, ErrVersiBentrok) {
		t.Fatalf("diharapkan ErrVersiBentrok, dapat %v", err)
	}
	if adaFile(t, l, baru.Gambar) {
		t.Error("file baru yang gagal disimpan harus dihapus (database.md §5)")
	}
	if !adaFile(t, l, b.GambarKey) {
		t.Error("gambar lama tidak boleh tersentuh")
	}
}

func TestUbahStatusDanKetidakadaan(t *testing.T) {
	l := siapkan(t)
	b, _ := l.svc.Buat(ctx, l.mentah(t, "Judul"), "u")

	if _, err := l.svc.Ubah(ctx, 999, l.mentah(t, "x"), 1, "u"); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("tidak ada: %v", err)
	}
	p, _ := l.svc.Post(ctx, b.ID, b.Version, "u")
	if _, err := l.svc.Ubah(ctx, b.ID, l.mentah(t, "x"), p.Version, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("edit posted harus ErrStatusTidakValid, dapat %v", err)
	}
	r, _ := l.svc.Retract(ctx, b.ID, p.Version, "u")
	if err := l.svc.Hapus(ctx, b.ID, "u"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.svc.Ubah(ctx, b.ID, l.mentah(t, "x"), r.Version, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("edit trash harus ErrStatusTidakValid, dapat %v", err)
	}
}

func TestSiklusStatus(t *testing.T) {
	l := siapkan(t)
	b, _ := l.svc.Buat(ctx, l.mentah(t, "Judul"), "u")

	p, err := l.svc.Post(ctx, b.ID, 1, "u")
	if err != nil || p.Status != "posted" || p.Version != 2 {
		t.Fatalf("Post: %+v %v", p, err)
	}
	if _, err := l.svc.Post(ctx, b.ID, 2, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("Post ulang pada posted: %v", err)
	}
	if err := l.svc.Hapus(ctx, b.ID, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("hapus posted harus ditolak (retract dulu): %v", err)
	}
	if _, err := l.svc.Retract(ctx, b.ID, 1, "u"); !errors.Is(err, ErrVersiBentrok) {
		t.Errorf("Retract versi basi: %v", err)
	}
	r, err := l.svc.Retract(ctx, b.ID, 2, "u")
	if err != nil || r.Status != "draft" || r.Version != 3 {
		t.Fatalf("Retract: %+v %v", r, err)
	}
	if _, err := l.svc.Retract(ctx, b.ID, 3, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("Retract pada draft: %v", err)
	}
	if err := l.svc.Hapus(ctx, b.ID, "u"); err != nil {
		t.Fatal(err)
	}
	if err := l.svc.Hapus(ctx, b.ID, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("hapus dua kali: %v", err)
	}
	pl, err := l.svc.Pulihkan(ctx, b.ID, "u")
	if err != nil || pl.Status != "draft" || pl.DihapusPada != nil || pl.Version != 3 {
		t.Fatalf("Pulihkan (kembali ke draft, version tidak naik): %+v %v", pl, err)
	}
	if _, err := l.svc.Pulihkan(ctx, b.ID, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("pulihkan yang tidak di-trash: %v", err)
	}
}

func TestPostMemeriksaGambarMasihAda(t *testing.T) {
	l := siapkan(t)
	b, _ := l.svc.Buat(ctx, l.mentah(t, "Judul"), "u")
	if err := l.mem.Hapus(ctx, b.GambarKey); err != nil {
		t.Fatal(err)
	}
	if _, err := l.svc.Post(ctx, b.ID, b.Version, "u"); !errors.Is(err, ErrGambarHilang) {
		t.Fatalf("diharapkan ErrGambarHilang, dapat %v", err)
	}
	if got, _ := l.svc.Detail(ctx, b.ID); got.Status != "draft" || got.Version != b.Version {
		t.Error("Post yang gagal tidak boleh mengubah apa pun")
	}
}

func TestPostStorageBermasalahBukanGambarHilang(t *testing.T) {
	l := siapkan(t)
	b, _ := l.svc.Buat(ctx, l.mentah(t, "Judul"), "u")
	l.mem.Err = errors.New("rustfs mati")
	_, err := l.svc.Post(ctx, b.ID, b.Version, "u")
	if !errors.Is(err, ErrStorage) || errors.Is(err, ErrGambarHilang) {
		t.Fatalf("diharapkan ErrStorage saja, dapat %v", err)
	}
}

func TestHapusPermanen(t *testing.T) {
	l := siapkan(t)
	b, _ := l.svc.Buat(ctx, l.mentah(t, "Judul"), "u")

	if err := l.svc.HapusPermanen(ctx, b.ID); !errors.Is(err, ErrStatusTidakValid) {
		t.Fatalf("hapus permanen non-trash harus ditolak: %v", err)
	}
	if !adaFile(t, l, b.GambarKey) {
		t.Fatal("penolakan tidak boleh menghapus file")
	}
	if err := l.svc.Hapus(ctx, b.ID, "u"); err != nil {
		t.Fatal(err)
	}
	if !adaFile(t, l, b.GambarKey) {
		t.Fatal("soft delete tidak boleh menyentuh gambar (database.md §5)")
	}
	if err := l.svc.HapusPermanen(ctx, b.ID); err != nil {
		t.Fatal(err)
	}
	if adaFile(t, l, b.GambarKey) {
		t.Error("hapus permanen harus menghapus file gambar")
	}
	if _, err := l.svc.Detail(ctx, b.ID); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("baris harus hilang: %v", err)
	}
	if err := l.svc.HapusPermanen(ctx, b.ID); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("hapus permanen dua kali: %v", err)
	}
}

func TestHapusPermanenTetapSuksesBilaFileGagalDihapus(t *testing.T) {
	l := siapkan(t)
	b, _ := l.svc.Buat(ctx, l.mentah(t, "Judul"), "u")
	_ = l.svc.Hapus(ctx, b.ID, "u")
	l.mem.Err = errors.New("rustfs mati")
	if err := l.svc.HapusPermanen(ctx, b.ID); err != nil {
		t.Fatalf("baris sudah terhapus; kegagalan file hanya dicatat (job yatim membersihkannya): %v", err)
	}
	if _, err := l.repo.Detail(ctx, b.ID); !errors.Is(err, ErrTidakDitemukan) {
		t.Error("baris harus sudah terhapus")
	}
}

func TestWaktuOperasiFileDipakai(t *testing.T) {
	if waktuOperasiFile <= 0 {
		t.Fatal(fmt.Sprint("waktuOperasiFile harus positif: ", waktuOperasiFile))
	}
}

// Data lama berjudul >70 karakter (dibuat sebelum batas 70) boleh disimpan selama judulnya tidak diubah.
func TestUbahJudulLamaDiAtasBatas(t *testing.T) {
	l := siapkan(t)
	b, err := l.svc.Buat(ctx, l.mentah(t, "Judul Awal"), "u")
	if err != nil {
		t.Fatal(err)
	}
	panjang := strings.Repeat("L", 100)
	l.repo.baris[b.ID].Judul = panjang // data lama yang lolos batas lama (200)

	m := MentahBerita{Judul: panjang, Isi: "isi baru", TanggalPublikasi: "09-11-2026", Gambar: b.GambarKey}
	u, err := l.svc.Ubah(ctx, b.ID, m, b.Version, "u")
	if err != nil || u.Judul != panjang || u.Isi != "isi baru" {
		t.Fatalf("judul lama tidak diubah harus lolos: %+v %v", u, err)
	}
	m.Judul = panjang + "!"
	if _, err := l.svc.Ubah(ctx, b.ID, m, u.Version, "u"); err == nil {
		t.Error("judul panjang yang diubah harus ditolak")
	}
	m.Judul = strings.Repeat("p", 70)
	if _, err := l.svc.Ubah(ctx, b.ID, m, u.Version, "u"); err != nil {
		t.Errorf("judul dipendekkan harus lolos: %v", err)
	}
}
