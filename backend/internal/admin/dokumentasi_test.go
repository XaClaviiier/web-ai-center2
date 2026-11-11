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

// repoDokPalsu meniru aturan SQL dokumentasi (status, version, slug unik, key gambar unik global, atomik) di memori.
type repoDokPalsu struct {
	baris   map[int64]*Dokumentasi
	urut    int64
	dipakai map[string]int64 // key -> id dokumentasi
}

func baruRepoDokPalsu() *repoDokPalsu {
	return &repoDokPalsu{baris: map[int64]*Dokumentasi{}, dipakai: map[string]int64{}}
}

func (r *repoDokPalsu) Daftar(context.Context, ParamDaftar) ([]RingkasDokumentasi, int64, error) {
	return nil, 0, nil
}

func (r *repoDokPalsu) Detail(_ context.Context, id int64) (Dokumentasi, error) {
	d, ok := r.baris[id]
	if !ok {
		return Dokumentasi{}, ErrTidakDitemukan
	}
	salin := *d
	salin.Gambar = append([]GambarDokumentasi(nil), d.Gambar...)
	return salin, nil
}

// bentrok memeriksa apakah ada key dalam daftar yang dipakai dokumentasi LAIN (id=0 berarti dokumentasi baru).
func (r *repoDokPalsu) bentrok(g []GambarDokumentasi, id int64) bool {
	for _, x := range g {
		if pemilik, ada := r.dipakai[x.Key]; ada && pemilik != id {
			return true
		}
	}
	return false
}

func (r *repoDokPalsu) Buat(_ context.Context, slug string, in InputDokumentasi, oleh string) (Dokumentasi, bool, error) {
	for _, d := range r.baris {
		if d.Slug == slug {
			return Dokumentasi{}, false, nil
		}
	}
	if r.bentrok(in.Gambar, 0) {
		return Dokumentasi{}, false, ErrGambarDipakai
	}
	r.urut++
	d := &Dokumentasi{ID: r.urut, Slug: slug, Judul: in.Judul, TanggalPublikasi: in.TanggalPublikasi,
		Gambar: append([]GambarDokumentasi(nil), in.Gambar...), Status: "draft", Version: 1, DibuatPada: time.Now(), DibuatOleh: oleh, DiubahOleh: oleh}
	r.baris[d.ID] = d
	for _, g := range in.Gambar {
		r.dipakai[g.Key] = d.ID
	}
	return *d, true, nil
}

func (r *repoDokPalsu) draftAktif(id int64, version int32) (*Dokumentasi, error) {
	d, ok := r.baris[id]
	if !ok {
		return nil, ErrTidakDitemukan
	}
	if d.DihapusPada != nil || d.Status != "draft" {
		return nil, ErrStatusTidakValid
	}
	if d.Version != version {
		return nil, ErrVersiBentrok
	}
	return d, nil
}

func (r *repoDokPalsu) Ubah(_ context.Context, id int64, version int32, in InputDokumentasi, oleh string) (Dokumentasi, error) {
	d, err := r.draftAktif(id, version)
	if err != nil {
		return Dokumentasi{}, err
	}
	if r.bentrok(in.Gambar, id) {
		return Dokumentasi{}, ErrGambarDipakai // atomik: tidak ada perubahan apa pun
	}
	for _, g := range d.Gambar {
		delete(r.dipakai, g.Key)
	}
	for _, g := range in.Gambar {
		r.dipakai[g.Key] = id
	}
	d.Judul, d.TanggalPublikasi, d.Gambar, d.DiubahOleh = in.Judul, in.TanggalPublikasi, append([]GambarDokumentasi(nil), in.Gambar...), oleh
	d.Version++
	return *d, nil
}

func (r *repoDokPalsu) Post(_ context.Context, id int64, version int32, oleh string) (Dokumentasi, error) {
	d, err := r.draftAktif(id, version)
	if err != nil {
		return Dokumentasi{}, err
	}
	d.Status, d.DiubahOleh = "posted", oleh
	d.Version++
	return *d, nil
}

func (r *repoDokPalsu) Retract(_ context.Context, id int64, version int32, oleh string) (Dokumentasi, error) {
	d, ok := r.baris[id]
	if !ok {
		return Dokumentasi{}, ErrTidakDitemukan
	}
	if d.DihapusPada != nil || d.Status != "posted" {
		return Dokumentasi{}, ErrStatusTidakValid
	}
	if d.Version != version {
		return Dokumentasi{}, ErrVersiBentrok
	}
	d.Status, d.DiubahOleh = "draft", oleh
	d.Version++
	return *d, nil
}

func (r *repoDokPalsu) Hapus(_ context.Context, id int64, oleh string) error {
	d, ok := r.baris[id]
	if !ok {
		return ErrTidakDitemukan
	}
	if d.DihapusPada != nil || d.Status != "draft" {
		return ErrStatusTidakValid
	}
	t := time.Now()
	d.DihapusPada, d.DiubahOleh = &t, oleh
	return nil
}

func (r *repoDokPalsu) Pulihkan(_ context.Context, id int64, oleh string) (Dokumentasi, error) {
	d, ok := r.baris[id]
	if !ok {
		return Dokumentasi{}, ErrTidakDitemukan
	}
	if d.DihapusPada == nil {
		return Dokumentasi{}, ErrStatusTidakValid
	}
	d.DihapusPada, d.DiubahOleh = nil, oleh
	return *d, nil
}

func (r *repoDokPalsu) HapusPermanen(_ context.Context, id int64) ([]string, error) {
	d, ok := r.baris[id]
	if !ok {
		return nil, ErrTidakDitemukan
	}
	if d.DihapusPada == nil {
		return nil, ErrStatusTidakValid
	}
	var keys []string
	for _, g := range d.Gambar {
		keys = append(keys, g.Key)
		delete(r.dipakai, g.Key)
	}
	delete(r.baris, id)
	return keys, nil
}

func (r *repoDokPalsu) GambarDipakai(_ context.Context, key string) (bool, error) {
	_, ada := r.dipakai[key]
	return ada, nil
}

type lingkunganDok struct {
	svc  DokumentasiLayanan
	repo *repoDokPalsu
	mem  *media.Memori
}

func siapkanDok(t *testing.T) lingkunganDok {
	t.Helper()
	repo, mem := baruRepoDokPalsu(), media.NewMemori()
	return lingkunganDok{
		svc:  DokumentasiLayanan{Repo: repo, Media: mem, Log: slog.New(slog.NewTextHandler(io.Discard, nil))},
		repo: repo, mem: mem,
	}
}

// unggahN menaruh n gambar dokumentasi di storage palsu dan mengembalikan daftar GambarDokumentasi-nya.
func (l lingkunganDok) unggahN(t *testing.T, n int) []GambarDokumentasi {
	t.Helper()
	out := make([]GambarDokumentasi, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, GambarDokumentasi{Key: unggahKey(t, l.mem, media.SumberDokumentasi)})
	}
	return out
}

func (l lingkunganDok) mentah(t *testing.T, judul string, n int) MentahDokumentasi {
	t.Helper()
	return MentahDokumentasi{Judul: judul, TanggalPublikasi: "2026-11-09", Gambar: l.unggahN(t, n)}
}

func keysDok(d Dokumentasi) []string {
	var out []string
	for _, g := range d.Gambar {
		out = append(out, g.Key)
	}
	return out
}

func TestValidasiDokumentasi(t *testing.T) {
	ok := MentahDokumentasi{Judul: "  Kegiatan   Satu ", TanggalPublikasi: "2026-11-09",
		Gambar: []GambarDokumentasi{{Key: "k1", Alt: "  Foto   utama "}, {Key: "k2"}}}
	in, err := ValidasiDokumentasi(ok)
	if err != nil {
		t.Fatal(err)
	}
	if in.Judul != "Kegiatan Satu" || in.Gambar[0].Alt != "Foto utama" || in.Gambar[1].Alt != "" || len(in.Gambar) != 2 {
		t.Errorf("perapian: %+v", in)
	}
	banyak := func(n int) []GambarDokumentasi {
		var g []GambarDokumentasi
		for i := 0; i < n; i++ {
			g = append(g, GambarDokumentasi{Key: "k" + strings.Repeat("x", i)})
		}
		return g
	}
	kasus := map[string]func(m *MentahDokumentasi){
		"judul kosong":     func(m *MentahDokumentasi) { m.Judul = "" },
		"judul panjang":    func(m *MentahDokumentasi) { m.Judul = strings.Repeat("a", BatasJudul+1) },
		"tanggal salah":    func(m *MentahDokumentasi) { m.TanggalPublikasi = "kemarin" },
		"tanpa gambar":     func(m *MentahDokumentasi) { m.Gambar = nil },
		"gambar kosong []": func(m *MentahDokumentasi) { m.Gambar = []GambarDokumentasi{} },
		"tujuh gambar":     func(m *MentahDokumentasi) { m.Gambar = banyak(7) },
		"key kosong":       func(m *MentahDokumentasi) { m.Gambar = []GambarDokumentasi{{Key: "a"}, {Key: ""}} },
		"key kembar":       func(m *MentahDokumentasi) { m.Gambar = []GambarDokumentasi{{Key: "a"}, {Key: "a"}} },
		"alt terlalu panjang": func(m *MentahDokumentasi) {
			m.Gambar = []GambarDokumentasi{{Key: "a", Alt: strings.Repeat("a", BatasAlt+1)}}
		},
		"alt karakter kontrol": func(m *MentahDokumentasi) { m.Gambar = []GambarDokumentasi{{Key: "a", Alt: "a\x00b"}} },
	}
	for nama, f := range kasus {
		m := ok
		f(&m)
		if _, err := ValidasiDokumentasi(m); err == nil {
			t.Errorf("%s: seharusnya ditolak", nama)
		} else {
			perluErrInput(t, err, "")
		}
	}
	for _, n := range []int{1, 6} {
		m := ok
		m.Gambar = banyak(n)
		if _, err := ValidasiDokumentasi(m); err != nil {
			t.Errorf("%d gambar harus diterima: %v", n, err)
		}
	}
	m := ok
	m.Gambar = []GambarDokumentasi{{Key: "a", Alt: strings.Repeat("a", BatasAlt)}}
	if _, err := ValidasiDokumentasi(m); err != nil {
		t.Errorf("alt tepat di batas harus diterima: %v", err)
	}
}

func TestDokBuatBerhasilEnamGambar(t *testing.T) {
	l := siapkanDok(t)
	m := l.mentah(t, "Pelatihan AI untuk Dosen", 6)
	d, err := l.svc.Buat(ctx, m, "dev-admin")
	if err != nil {
		t.Fatal(err)
	}
	if d.Slug != "pelatihan-ai-untuk-dosen" || d.Status != "draft" || d.Version != 1 || d.DibuatOleh != "dev-admin" || len(d.Gambar) != 6 {
		t.Errorf("hasil salah: %+v", d)
	}
	for _, g := range m.Gambar {
		if !adaKey(t, l.mem, g.Key) {
			t.Errorf("gambar %s tidak boleh terhapus saat berhasil", g.Key)
		}
	}
}

func TestDokBuatSlugBentrokDiberiAkhiran(t *testing.T) {
	l := siapkanDok(t)
	var slugs []string
	for i := 0; i < 3; i++ {
		d, err := l.svc.Buat(ctx, l.mentah(t, "Judul Sama", 1), "u")
		if err != nil {
			t.Fatal(err)
		}
		slugs = append(slugs, d.Slug)
	}
	if strings.Join(slugs, ",") != "judul-sama,judul-sama-2,judul-sama-3" {
		t.Errorf("slug: %v", slugs)
	}
}

func TestDokBuatKeyDitolak(t *testing.T) {
	l := siapkanDok(t)
	keyTim := unggahKey(t, l.mem, media.SumberTim)
	baik := l.unggahN(t, 1)[0]

	kasus := map[string]string{
		"key sembarang":      "foto.png",
		"sumber lain (tim)":  keyTim,
		"bentuk sah tak ada": "dokumentasi/3f9c2a10-7b4e-4c6d-9a1f-0123456789ab.png",
		"traversal":          "../dokumentasi/x.png",
	}
	for nama, key := range kasus {
		m := MentahDokumentasi{Judul: "J", TanggalPublikasi: "2026-11-09", Gambar: []GambarDokumentasi{baik, {Key: key}}}
		_, err := l.svc.Buat(ctx, m, "u")
		var e *ErrInput
		if !errors.As(err, &e) {
			t.Errorf("%s: diharapkan ErrInput, dapat %v", nama, err)
		}
		if len(l.repo.baris) != 0 {
			t.Errorf("%s: tidak boleh ada baris tercipta", nama)
		}
	}
	if !adaKey(t, l.mem, keyTim) {
		t.Error("gambar sumber lain tidak boleh ikut terhapus")
	}
}

func TestDokBuatGagalMenghapusFileBaruTetapMenjagaYangDipakai(t *testing.T) {
	l := siapkanDok(t)

	// 1. validasi teks gagal: semua file baru yang sah dan belum dipakai dihapus kembali
	m := l.mentah(t, "", 3)
	if _, err := l.svc.Buat(ctx, m, "u"); err == nil {
		t.Fatal("seharusnya galat")
	}
	for _, g := range m.Gambar {
		if adaKey(t, l.mem, g.Key) {
			t.Errorf("file baru %s harus dihapus kembali", g.Key)
		}
	}

	// 2. satu key tidak ada di bucket: galat; key lain yang sah dan belum dipakai ikut dibersihkan
	baik1, baik2 := l.unggahN(t, 2)[0], l.unggahN(t, 1)[0]
	hilang := GambarDokumentasi{Key: "dokumentasi/3f9c2a10-7b4e-4c6d-9a1f-0123456789ab.png"}
	if _, err := l.svc.Buat(ctx, MentahDokumentasi{Judul: "J", TanggalPublikasi: "2026-11-09", Gambar: []GambarDokumentasi{baik1, hilang, baik2}}, "u"); err == nil {
		t.Fatal("seharusnya galat")
	}
	if adaKey(t, l.mem, baik1.Key) || adaKey(t, l.mem, baik2.Key) {
		t.Error("file baru yang gagal disimpan harus dihapus kembali")
	}

	// 3. key milik dokumentasi lain TIDAK boleh terhapus oleh request yang gagal
	pertama, err := l.svc.Buat(ctx, l.mentah(t, "Pertama", 2), "u")
	if err != nil {
		t.Fatal(err)
	}
	dup := MentahDokumentasi{Judul: "Kedua", TanggalPublikasi: "2026-11-09", Gambar: []GambarDokumentasi{{Key: pertama.Gambar[1].Key}}}
	_, err = l.svc.Buat(ctx, dup, "u")
	perluErrInput(t, err, "sudah dipakai")
	for _, g := range pertama.Gambar {
		if !adaKey(t, l.mem, g.Key) {
			t.Fatalf("file milik dokumentasi lain terhapus: %s", g.Key)
		}
	}
}

func TestDokBuatStorageBermasalah(t *testing.T) {
	l := siapkanDok(t)
	m := l.mentah(t, "Judul", 2)
	l.mem.Err = errors.New("rustfs mati")
	if _, err := l.svc.Buat(ctx, m, "u"); !errors.Is(err, ErrStorage) {
		t.Fatalf("diharapkan ErrStorage, dapat %v", err)
	}
}

func TestDokUbahSiklusFile(t *testing.T) {
	l := siapkanDok(t)
	awal := l.mentah(t, "Dok", 3)
	d, err := l.svc.Buat(ctx, awal, "u")
	if err != nil {
		t.Fatal(err)
	}
	a, b, c := awal.Gambar[0], awal.Gambar[1], awal.Gambar[2]
	ubah := func(g []GambarDokumentasi, version int32) Dokumentasi {
		t.Helper()
		r, err := l.svc.Ubah(ctx, d.ID, MentahDokumentasi{Judul: "Dok", TanggalPublikasi: "2026-11-09", Gambar: g}, version, "u")
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	// 1. ganti urutan: tidak ada file yang dihapus
	r := ubah([]GambarDokumentasi{c, a, b}, 1)
	if strings.Join(keysDok(r), ",") != strings.Join([]string{c.Key, a.Key, b.Key}, ",") || r.Version != 2 {
		t.Fatalf("urutan: %v", keysDok(r))
	}
	for _, g := range []GambarDokumentasi{a, b, c} {
		if !adaKey(t, l.mem, g.Key) {
			t.Fatalf("ganti urutan tidak boleh menghapus %s", g.Key)
		}
	}

	// 2. replace di urutan yang sama: key lama (b) dihapus SETELAH commit, key baru tersimpan
	baru := l.unggahN(t, 1)[0]
	r = ubah([]GambarDokumentasi{c, a, baru}, 2)
	if strings.Join(keysDok(r), ",") != strings.Join([]string{c.Key, a.Key, baru.Key}, ",") {
		t.Fatalf("replace: %v", keysDok(r))
	}
	if adaKey(t, l.mem, b.Key) || !adaKey(t, l.mem, baru.Key) {
		t.Error("replace: file lama harus dihapus dan file baru tersimpan")
	}

	// 3. hapus headline: hanya file yang dibuang yang dihapus; yang lain utuh
	r = ubah([]GambarDokumentasi{a, baru}, 3)
	if adaKey(t, l.mem, c.Key) || !adaKey(t, l.mem, a.Key) || !adaKey(t, l.mem, baru.Key) || len(r.Gambar) != 2 {
		t.Error("hapus headline: hanya c yang dihapus")
	}

	// 4. tambah di akhir
	tambah := l.unggahN(t, 2)
	r = ubah(append([]GambarDokumentasi{a, baru}, tambah...), 4)
	if len(r.Gambar) != 4 || !adaKey(t, l.mem, tambah[1].Key) {
		t.Errorf("tambah: %v", keysDok(r))
	}

	// 5. alt diperbarui tanpa operasi file
	r = ubah([]GambarDokumentasi{{Key: a.Key, Alt: "Alt baru"}, baru, tambah[0], tambah[1]}, 5)
	if r.Gambar[0].Alt != "Alt baru" {
		t.Errorf("alt: %+v", r.Gambar[0])
	}
}

func TestDokUbahVersiBentrokMenghapusFileBaruTapiMenjagaLama(t *testing.T) {
	l := siapkanDok(t)
	awal := l.mentah(t, "Dok", 2)
	d, _ := l.svc.Buat(ctx, awal, "u")
	baru := l.unggahN(t, 2)

	_, err := l.svc.Ubah(ctx, d.ID, MentahDokumentasi{Judul: "Dok", TanggalPublikasi: "2026-11-09", Gambar: []GambarDokumentasi{awal.Gambar[0], baru[0], baru[1]}}, 99, "u")
	if !errors.Is(err, ErrVersiBentrok) {
		t.Fatalf("diharapkan ErrVersiBentrok, dapat %v", err)
	}
	for _, g := range baru {
		if adaKey(t, l.mem, g.Key) {
			t.Errorf("file baru %s yang gagal disimpan harus dihapus (database.md §5)", g.Key)
		}
	}
	for _, g := range awal.Gambar {
		if !adaKey(t, l.mem, g.Key) {
			t.Errorf("gambar lama %s tidak boleh tersentuh", g.Key)
		}
	}
	got, _ := l.svc.Detail(ctx, d.ID)
	if len(got.Gambar) != 2 || got.Version != 1 {
		t.Errorf("data tidak boleh berubah: %+v", got)
	}
}

func TestDokUbahKeyBaruBermasalahTidakMengubahApaPun(t *testing.T) {
	l := siapkanDok(t)
	awal := l.mentah(t, "Dok", 2)
	d, _ := l.svc.Buat(ctx, awal, "u")
	lain, _ := l.svc.Buat(ctx, l.mentah(t, "Lain", 1), "u")
	bagus := l.unggahN(t, 1)[0]
	hilang := GambarDokumentasi{Key: "dokumentasi/3f9c2a10-7b4e-4c6d-9a1f-0123456789ab.png"}

	// 1. key baru tidak ada di bucket: ditolak sebelum menyentuh database; key baru lain yang sah dibersihkan
	_, err := l.svc.Ubah(ctx, d.ID, MentahDokumentasi{Judul: "Dok", TanggalPublikasi: "2026-11-09", Gambar: []GambarDokumentasi{awal.Gambar[0], bagus, hilang}}, 1, "u")
	perluErrInput(t, err, "belum diunggah")
	if adaKey(t, l.mem, bagus.Key) {
		t.Error("file baru yang sah ikut dibersihkan saat request gagal")
	}
	// 2. key milik dokumentasi lain: ditolak, dan file milik dokumentasi lain TIDAK terhapus
	_, err = l.svc.Ubah(ctx, d.ID, MentahDokumentasi{Judul: "Dok", TanggalPublikasi: "2026-11-09", Gambar: []GambarDokumentasi{awal.Gambar[0], lain.Gambar[0]}}, 1, "u")
	perluErrInput(t, err, "sudah dipakai")
	if !adaKey(t, l.mem, lain.Gambar[0].Key) {
		t.Fatal("file milik dokumentasi lain terhapus")
	}
	got, _ := l.svc.Detail(ctx, d.ID)
	if got.Version != 1 || len(got.Gambar) != 2 {
		t.Errorf("tidak ada perubahan yang boleh tersimpan: %+v", got)
	}
	for _, g := range awal.Gambar {
		if !adaKey(t, l.mem, g.Key) {
			t.Errorf("gambar asli %s harus utuh", g.Key)
		}
	}
}

func TestDokUbahKeyYangSudahDimilikiTidakDiperiksaUlang(t *testing.T) {
	l := siapkanDok(t)
	awal := l.mentah(t, "Dok", 2)
	d, _ := l.svc.Buat(ctx, awal, "u")
	// file gambar kedua hilang dari bucket (mis. korup); mengubah judul/urutan tidak boleh gagal karenanya
	if err := l.mem.Hapus(ctx, awal.Gambar[1].Key); err != nil {
		t.Fatal(err)
	}
	if _, err := l.svc.Ubah(ctx, d.ID, MentahDokumentasi{Judul: "Dok Baru", TanggalPublikasi: "2026-11-09", Gambar: awal.Gambar}, 1, "u"); err != nil {
		t.Fatalf("key yang sudah dimiliki tidak diperiksa ulang saat Simpan: %v", err)
	}
}

func TestDokUbahStatusDanKetidakadaan(t *testing.T) {
	l := siapkanDok(t)
	m := l.mentah(t, "Dok", 1)
	d, _ := l.svc.Buat(ctx, m, "u")
	ubah := MentahDokumentasi{Judul: "x", TanggalPublikasi: "2026-11-09", Gambar: m.Gambar}

	if _, err := l.svc.Ubah(ctx, 999, ubah, 1, "u"); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("tidak ada: %v", err)
	}
	p, _ := l.svc.Post(ctx, d.ID, 1, "u")
	if _, err := l.svc.Ubah(ctx, d.ID, ubah, p.Version, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("edit posted: %v", err)
	}
	r, _ := l.svc.Retract(ctx, d.ID, p.Version, "u")
	_ = l.svc.Hapus(ctx, d.ID, "u")
	if _, err := l.svc.Ubah(ctx, d.ID, ubah, r.Version, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("edit trash: %v", err)
	}
}

func TestDokPostMemeriksaSemuaGambar(t *testing.T) {
	l := siapkanDok(t)
	m := l.mentah(t, "Dok", 3)
	d, _ := l.svc.Buat(ctx, m, "u")

	// gambar terakhir hilang: ErrGambarHilang, tidak ada perubahan
	if err := l.mem.Hapus(ctx, m.Gambar[2].Key); err != nil {
		t.Fatal(err)
	}
	if _, err := l.svc.Post(ctx, d.ID, 1, "u"); !errors.Is(err, ErrGambarHilang) {
		t.Fatalf("diharapkan ErrGambarHilang, dapat %v", err)
	}
	if got, _ := l.svc.Detail(ctx, d.ID); got.Status != "draft" || got.Version != 1 {
		t.Error("Post yang gagal tidak boleh mengubah apa pun")
	}
	// storage bermasalah bukan "gambar hilang"
	l.mem.Err = errors.New("rustfs mati")
	if _, err := l.svc.Post(ctx, d.ID, 1, "u"); !errors.Is(err, ErrStorage) || errors.Is(err, ErrGambarHilang) {
		t.Fatalf("diharapkan ErrStorage saja, dapat %v", err)
	}
	l.mem.Err = nil
	// diunggah ulang: Post berhasil
	if err := l.mem.Put(ctx, m.Gambar[2].Key, png, "image/png"); err != nil {
		t.Fatal(err)
	}
	if p, err := l.svc.Post(ctx, d.ID, 1, "u"); err != nil || p.Status != "posted" {
		t.Fatalf("Post: %+v %v", p, err)
	}
}

func TestDokSiklusStatus(t *testing.T) {
	l := siapkanDok(t)
	d, _ := l.svc.Buat(ctx, l.mentah(t, "Dok", 2), "u")

	p, _ := l.svc.Post(ctx, d.ID, 1, "u")
	if err := l.svc.Hapus(ctx, d.ID, "u"); !errors.Is(err, ErrStatusTidakValid) {
		t.Errorf("hapus posted: %v", err)
	}
	if _, err := l.svc.Retract(ctx, d.ID, 1, "u"); !errors.Is(err, ErrVersiBentrok) {
		t.Errorf("retract version basi: %v", err)
	}
	if _, err := l.svc.Retract(ctx, d.ID, p.Version, "u"); err != nil {
		t.Fatal(err)
	}
	if err := l.svc.Hapus(ctx, d.ID, "u"); err != nil {
		t.Fatal(err)
	}
	pl, err := l.svc.Pulihkan(ctx, d.ID, "u")
	if err != nil || pl.Status != "draft" || pl.DihapusPada != nil {
		t.Fatalf("Pulihkan: %+v %v", pl, err)
	}
}

func TestDokHapusPermanenMenghapusSemuaFile(t *testing.T) {
	l := siapkanDok(t)
	m := l.mentah(t, "Dok", 4)
	d, _ := l.svc.Buat(ctx, m, "u")
	lain, _ := l.svc.Buat(ctx, l.mentah(t, "Lain", 1), "u")

	if err := l.svc.HapusPermanen(ctx, d.ID); !errors.Is(err, ErrStatusTidakValid) {
		t.Fatalf("non-trash: %v", err)
	}
	for _, g := range m.Gambar {
		if !adaKey(t, l.mem, g.Key) {
			t.Fatal("penolakan tidak boleh menghapus file")
		}
	}
	_ = l.svc.Hapus(ctx, d.ID, "u")
	for _, g := range m.Gambar {
		if !adaKey(t, l.mem, g.Key) {
			t.Fatal("soft delete tidak boleh menyentuh gambar (database.md §5)")
		}
	}
	if err := l.svc.HapusPermanen(ctx, d.ID); err != nil {
		t.Fatal(err)
	}
	for _, g := range m.Gambar {
		if adaKey(t, l.mem, g.Key) {
			t.Errorf("hapus permanen harus menghapus SEMUA file: %s masih ada", g.Key)
		}
	}
	if !adaKey(t, l.mem, lain.Gambar[0].Key) {
		t.Error("file dokumentasi lain tidak boleh tersentuh")
	}
	if err := l.svc.HapusPermanen(ctx, d.ID); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("hapus permanen dua kali: %v", err)
	}
}

func TestDokHapusPermanenTetapSuksesBilaFileGagalDihapus(t *testing.T) {
	l := siapkanDok(t)
	d, _ := l.svc.Buat(ctx, l.mentah(t, "Dok", 3), "u")
	_ = l.svc.Hapus(ctx, d.ID, "u")
	l.mem.Err = errors.New("rustfs mati")
	if err := l.svc.HapusPermanen(ctx, d.ID); err != nil {
		t.Fatalf("baris sudah terhapus; kegagalan file hanya dicatat: %v", err)
	}
	if _, err := l.repo.Detail(ctx, d.ID); !errors.Is(err, ErrTidakDitemukan) {
		t.Error("baris harus sudah terhapus")
	}
}
