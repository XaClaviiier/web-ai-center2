package admin

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
)

// repoTimPalsu meniru aturan SQL tim (urutan dari klien, version, foto unik) di memori.
type repoTimPalsu struct {
	baris   map[int64]*Tim
	urut    int64
	dipakai map[string]int64
}

func baruRepoTimPalsu() *repoTimPalsu {
	return &repoTimPalsu{baris: map[int64]*Tim{}, dipakai: map[string]int64{}}
}

func (r *repoTimPalsu) Daftar(context.Context) ([]Tim, error) {
	out := make([]Tim, 0, len(r.baris))
	for _, t := range r.baris {
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Urutan != out[j].Urutan {
			return out[i].Urutan < out[j].Urutan
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (r *repoTimPalsu) Detail(_ context.Context, id int64) (Tim, error) {
	t, ok := r.baris[id]
	if !ok {
		return Tim{}, ErrTidakDitemukan
	}
	return *t, nil
}

func (r *repoTimPalsu) Buat(_ context.Context, in InputTim, oleh string) (Tim, error) {
	if _, ada := r.dipakai[in.Foto]; ada {
		return Tim{}, ErrGambarDipakai
	}
	r.urut++
	t := &Tim{ID: r.urut, Nama: in.Nama, Posisi: in.Posisi, BidangRiset: in.BidangRiset, FotoKey: in.Foto, Urutan: in.Urutan,
		Version: 1, DibuatPada: time.Now(), DibuatOleh: oleh, DiubahOleh: oleh}
	r.baris[t.ID] = t
	r.dipakai[in.Foto] = t.ID
	return *t, nil
}

func (r *repoTimPalsu) Ubah(_ context.Context, id int64, version int32, in InputTim, oleh string) (Tim, error) {
	t, ok := r.baris[id]
	if !ok {
		return Tim{}, ErrTidakDitemukan
	}
	if t.Version != version {
		return Tim{}, ErrVersiBentrok
	}
	if pemilik, ada := r.dipakai[in.Foto]; ada && pemilik != id {
		return Tim{}, ErrGambarDipakai
	}
	delete(r.dipakai, t.FotoKey)
	r.dipakai[in.Foto] = id
	t.Nama, t.Posisi, t.BidangRiset, t.FotoKey, t.Urutan, t.DiubahOleh = in.Nama, in.Posisi, in.BidangRiset, in.Foto, in.Urutan, oleh
	t.Version++
	return *t, nil
}

func (r *repoTimPalsu) Hapus(_ context.Context, id int64) (string, error) {
	t, ok := r.baris[id]
	if !ok {
		return "", ErrTidakDitemukan
	}
	delete(r.baris, id)
	delete(r.dipakai, t.FotoKey)
	return t.FotoKey, nil
}

func (r *repoTimPalsu) FotoDipakai(_ context.Context, key string) (bool, error) {
	_, ada := r.dipakai[key]
	return ada, nil
}

type lingkunganTim struct {
	svc  TimLayanan
	repo *repoTimPalsu
	mem  *media.Memori
}

func siapkanTim(t *testing.T) lingkunganTim {
	t.Helper()
	repo, mem := baruRepoTimPalsu(), media.NewMemori()
	return lingkunganTim{
		svc:  TimLayanan{Repo: repo, Media: mem, Log: slog.New(slog.NewTextHandler(io.Discard, nil))},
		repo: repo, mem: mem,
	}
}

// unggahKey menaruh gambar di storage palsu dan mengembalikan key-nya (sumber bebas).
func unggahKey(t *testing.T, mem *media.Memori, sumber string) string {
	t.Helper()
	key, err := media.BuatKey(sumber, "png")
	if err != nil {
		t.Fatal(err)
	}
	if err := mem.Put(context.Background(), key, png, "image/png"); err != nil {
		t.Fatal(err)
	}
	return key
}

func adaKey(t *testing.T, mem *media.Memori, key string) bool {
	t.Helper()
	ada, err := media.Ada(context.Background(), mem, key)
	if err != nil {
		t.Fatal(err)
	}
	return ada
}

func (l lingkunganTim) mentah(t *testing.T, nama string) MentahTim {
	t.Helper()
	return MentahTim{Nama: nama, Posisi: "Peneliti", BidangRiset: "Pemrosesan Bahasa Alami", Foto: unggahKey(t, l.mem, media.SumberTim), Urutan: urut(1)}
}

// urut membuat penunjuk urutan untuk MentahTim.
func urut(n int64) *int64 { return &n }

// Urutan tampil dipilih admin: daftar mengikuti angka urutan, bukan urutan pembuatan.
func TestTimBuatMemakaiUrutanDariKlien(t *testing.T) {
	l := siapkanTim(t)
	for i, nama := range []string{"Ani", "Budi", "Cici"} {
		m := l.mentah(t, nama)
		m.Urutan = urut([]int64{3, 1, 2}[i])
		tm, err := l.svc.Buat(ctx, m, "dev-admin")
		if err != nil {
			t.Fatal(err)
		}
		if tm.Urutan != int32([]int64{3, 1, 2}[i]) || tm.Version != 1 || tm.DibuatOleh != "dev-admin" || tm.FotoKey != m.Foto {
			t.Errorf("%s: %+v", nama, tm)
		}
		if !adaKey(t, l.mem, m.Foto) {
			t.Errorf("%s: foto tidak boleh terhapus saat berhasil", nama)
		}
	}
	daftar, _ := l.svc.Daftar(ctx)
	var nama []string
	for _, d := range daftar {
		nama = append(nama, d.Nama)
	}
	if strings.Join(nama, ",") != "Budi,Cici,Ani" {
		t.Errorf("urutan daftar: %v", nama)
	}
}

func TestTimValidasi(t *testing.T) {
	ok := MentahTim{Nama: "  Dr.   Ani ", Posisi: "Ketua", BidangRiset: "NLP", Foto: "tim/x.png", Urutan: urut(1)}
	in, err := ValidasiTim(ok)
	if err != nil || in.Nama != "Dr. Ani" {
		t.Fatalf("perapian: %+v %v", in, err)
	}
	kasus := map[string]func(m *MentahTim){
		"nama kosong":            func(m *MentahTim) { m.Nama = " " },
		"nama terlalu panjang":   func(m *MentahTim) { m.Nama = strings.Repeat("a", BatasNama+1) },
		"posisi kosong":          func(m *MentahTim) { m.Posisi = "" },
		"posisi terlalu panjang": func(m *MentahTim) { m.Posisi = strings.Repeat("a", BatasPosisi+1) },
		"bidang kosong":          func(m *MentahTim) { m.BidangRiset = "" },
		"bidang terlalu panjang": func(m *MentahTim) { m.BidangRiset = strings.Repeat("a", BatasBidangRiset+1) },
		"foto kosong":            func(m *MentahTim) { m.Foto = "" },
		"karakter kontrol":       func(m *MentahTim) { m.Nama = "a\x00b" },
		"urutan tidak dikirim":   func(m *MentahTim) { m.Urutan = nil },
		"urutan nol":             func(m *MentahTim) { m.Urutan = urut(0) },
		"urutan negatif":         func(m *MentahTim) { m.Urutan = urut(-3) },
		"urutan di atas int32":   func(m *MentahTim) { m.Urutan = urut(1 << 31) },
	}
	for nama, f := range kasus {
		m := ok
		f(&m)
		if _, err := ValidasiTim(m); err == nil {
			t.Errorf("%s: seharusnya ditolak", nama)
		} else {
			perluErrInput(t, err, "")
		}
	}
	m := ok
	m.BidangRiset = strings.Repeat("a", BatasBidangRiset)
	m.Urutan = urut(1<<31 - 1) // batas atas kolom integer
	if in, err := ValidasiTim(m); err != nil || in.Urutan != 1<<31-1 {
		t.Errorf("tepat di batas harus diterima: %+v %v", in, err)
	}
}

func TestTimBuatKeyDitolakDanFileBaruDihapusKembali(t *testing.T) {
	l := siapkanTim(t)
	keyBerita := unggahKey(t, l.mem, media.SumberBerita)
	for nama, key := range map[string]string{
		"sumber lain (berita)": keyBerita,
		"key sembarang":        "foto.png",
		"bentuk sah tak ada":   "tim/3f9c2a10-7b4e-4c6d-9a1f-0123456789ab.png",
	} {
		m := MentahTim{Nama: "A", Posisi: "B", BidangRiset: "C", Foto: key, Urutan: urut(1)}
		_, err := l.svc.Buat(ctx, m, "u")
		var e *ErrInput
		if !errors.As(err, &e) {
			t.Errorf("%s: diharapkan ErrInput, dapat %v", nama, err)
		}
	}
	if !adaKey(t, l.mem, keyBerita) {
		t.Error("file sumber lain tidak boleh ikut terhapus")
	}
	// validasi teks gagal dengan key tim sah yang belum dipakai: file baru dihapus kembali
	m := l.mentah(t, "")
	if _, err := l.svc.Buat(ctx, m, "u"); err == nil {
		t.Fatal("seharusnya galat")
	}
	if adaKey(t, l.mem, m.Foto) {
		t.Error("file baru yang gagal disimpan harus dihapus kembali")
	}
	if len(l.repo.baris) != 0 {
		t.Error("tidak boleh ada baris tercipta")
	}
}

func TestTimFotoKembarDitolakTanpaMenghapusMilikOrangLain(t *testing.T) {
	l := siapkanTim(t)
	a, _ := l.svc.Buat(ctx, l.mentah(t, "A"), "u")
	_, err := l.svc.Buat(ctx, MentahTim{Nama: "B", Posisi: "p", BidangRiset: "b", Foto: a.FotoKey, Urutan: urut(2)}, "u")
	var e *ErrInput
	if !errors.As(err, &e) || !strings.Contains(e.Pesan, "sudah dipakai") {
		t.Fatalf("diharapkan ErrInput 'sudah dipakai', dapat %v", err)
	}
	if !adaKey(t, l.mem, a.FotoKey) {
		t.Fatal("foto milik anggota lain tidak boleh terhapus")
	}
}

func TestTimUbah(t *testing.T) {
	l := siapkanTim(t)
	a, _ := l.svc.Buat(ctx, l.mentah(t, "Ani"), "u1")
	lama := a.FotoKey

	// ganti foto: lama dihapus setelah commit
	baru := l.mentah(t, "Ani Lestari")
	u, err := l.svc.Ubah(ctx, a.ID, baru, 1, "u2")
	if err != nil {
		t.Fatal(err)
	}
	if u.Nama != "Ani Lestari" || u.Version != 2 || u.DiubahOleh != "u2" || u.Urutan != a.Urutan || u.FotoKey != baru.Foto {
		t.Errorf("hasil salah: %+v", u)
	}
	if adaKey(t, l.mem, lama) || !adaKey(t, l.mem, baru.Foto) {
		t.Error("foto lama harus dihapus dan foto baru tersimpan")
	}

	// tanpa ganti foto: tidak ada file yang dihapus
	u2, err := l.svc.Ubah(ctx, a.ID, MentahTim{Nama: "Ani L.", Posisi: "p", BidangRiset: "b", Foto: baru.Foto, Urutan: urut(5)}, 2, "u")
	if err != nil || !adaKey(t, l.mem, baru.Foto) || u2.Version != 3 {
		t.Fatalf("tanpa ganti foto: %+v %v", u2, err)
	}
}

func TestTimUbahGagalMenghapusFileBaruTapiMenjagaLama(t *testing.T) {
	l := siapkanTim(t)
	a, _ := l.svc.Buat(ctx, l.mentah(t, "Ani"), "u")
	b, _ := l.svc.Buat(ctx, l.mentah(t, "Budi"), "u")

	// version basi
	baru := l.mentah(t, "Ani")
	if _, err := l.svc.Ubah(ctx, a.ID, baru, 99, "u"); !errors.Is(err, ErrVersiBentrok) {
		t.Fatalf("version basi: %v", err)
	}
	if adaKey(t, l.mem, baru.Foto) {
		t.Error("file baru yang gagal disimpan harus dihapus")
	}
	if !adaKey(t, l.mem, a.FotoKey) {
		t.Error("foto lama tidak boleh tersentuh")
	}
	// memakai foto milik anggota lain
	_, err := l.svc.Ubah(ctx, a.ID, MentahTim{Nama: "x", Posisi: "p", BidangRiset: "b", Foto: b.FotoKey, Urutan: urut(1)}, 1, "u")
	perluErrInput(t, err, "sudah dipakai")
	if !adaKey(t, l.mem, b.FotoKey) {
		t.Error("foto milik anggota lain tidak boleh terhapus")
	}
	// tidak ada
	if _, err := l.svc.Ubah(ctx, 999, l.mentah(t, "x"), 1, "u"); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("tidak ada: %v", err)
	}
}

func TestTimHapusLangsungPermanen(t *testing.T) {
	l := siapkanTim(t)
	a, _ := l.svc.Buat(ctx, l.mentah(t, "Ani"), "u")
	b, _ := l.svc.Buat(ctx, l.mentah(t, "Budi"), "u")

	if err := l.svc.Hapus(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := l.svc.Detail(ctx, a.ID); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("baris harus hilang: %v", err)
	}
	if adaKey(t, l.mem, a.FotoKey) {
		t.Error("foto harus dihapus")
	}
	if !adaKey(t, l.mem, b.FotoKey) {
		t.Error("foto anggota lain tidak boleh tersentuh")
	}
	if err := l.svc.Hapus(ctx, a.ID); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("hapus dua kali: %v", err)
	}
}

func TestTimHapusTetapSuksesBilaFileGagalDihapus(t *testing.T) {
	l := siapkanTim(t)
	a, _ := l.svc.Buat(ctx, l.mentah(t, "Ani"), "u")
	l.mem.Err = errors.New("rustfs mati")
	if err := l.svc.Hapus(ctx, a.ID); err != nil {
		t.Fatalf("baris sudah terhapus; kegagalan file hanya dicatat: %v", err)
	}
	if _, err := l.repo.Detail(ctx, a.ID); !errors.Is(err, ErrTidakDitemukan) {
		t.Error("baris harus sudah terhapus")
	}
}

func TestTimStorageBermasalahSaatBuat(t *testing.T) {
	l := siapkanTim(t)
	m := l.mentah(t, "Ani")
	l.mem.Err = errors.New("rustfs mati")
	if _, err := l.svc.Buat(ctx, m, "u"); !errors.Is(err, ErrStorage) {
		t.Fatalf("diharapkan ErrStorage, dapat %v", err)
	}
}
