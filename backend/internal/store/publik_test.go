package store

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

func bukaStoreTes(t *testing.T) (*Store, context.Context) {
	t.Helper()
	url := testutil.DatabaseTes(t)
	ctx, batal := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(batal)
	s, err := Open(ctx, url)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(s.Close)
	testutil.BersihkanKonten(t, s.Pool)
	return s, ctx
}

type fixtureBerita struct {
	slug, judul, isi, tanggal, dibuat, status string
	trash                                     bool
}

func sisipBerita(t *testing.T, ctx context.Context, s *Store, f fixtureBerita) {
	t.Helper()
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO berita (slug, judul, isi, tanggal_publikasi, gambar, status, created_at, created_by, updated_by, deleted_at)
		 VALUES ($1, $2, $3, $4::date, 'berita/x/'||$1||'.jpg', $5, $6::timestamptz, 'u', 'u', CASE WHEN $7::boolean THEN now() END)`,
		f.slug, f.judul, f.isi, f.tanggal, f.status, f.dibuat, f.trash)
	if err != nil {
		t.Fatalf("sisip berita %s: %v", f.slug, err)
	}
}

func slugs(items []konten.BeritaRingkas) string {
	var out []string
	for _, i := range items {
		out = append(out, i.Slug)
	}
	return strings.Join(out, ",")
}

func seedBerita(t *testing.T, ctx context.Context, s *Store) {
	t.Helper()
	for _, f := range []fixtureBerita{
		{"b1", "Berita Satu", "Paragraf pertama.\n\nParagraf kedua", "2026-03-01", "2026-03-01T08:00:00Z", "posted", false},
		{"b2", "Berita Dua", "Isi dua", "2026-03-01", "2026-03-01T09:00:00Z", "posted", false}, // tanggal sama dengan b1, dibuat lebih akhir
		{"b3", "Berita Tiga", "Isi tiga", "2026-04-01", "2026-04-01T08:00:00Z", "posted", false},
		{"b4", "Draft rahasia", "Isi draft", "2026-05-01", "2026-05-01T08:00:00Z", "draft", false},
		{"b5", "Berita terhapus", "Isi terhapus", "2026-05-02", "2026-05-02T08:00:00Z", "posted", true},
		{"b6", "Diskon 50%_ spesial", strings.Repeat("x", 300), "2026-01-01", "2026-01-01T08:00:00Z", "posted", false},
	} {
		sisipBerita(t, ctx, s, f)
	}
}

func TestListBeritaUrutanDanFilterStatus(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	seedBerita(t, ctx, s)

	items, total, err := s.ListBerita(ctx, konten.ParamDaftar{Batas: 10})
	if err != nil {
		t.Fatal(err)
	}
	// terbaru dulu; b2 sebelum b1 (tanggal sama, created_at lebih akhir); draft dan trash tidak ikut.
	if got := slugs(items); got != "b3,b2,b1,b6" || total != 4 {
		t.Fatalf("urutan/filter salah: %s total=%d", got, total)
	}
	if items[0].TanggalPublikasi != "2026-04-01" || items[0].GambarKey != "berita/x/b3.jpg" {
		t.Errorf("field item salah: %+v", items[0])
	}
}

func TestListBeritaPagination(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	seedBerita(t, ctx, s)

	items, total, err := s.ListBerita(ctx, konten.ParamDaftar{Batas: 2, Geser: 2})
	if err != nil || slugs(items) != "b1,b6" || total != 4 {
		t.Fatalf("halaman 2: %s total=%d err=%v", slugs(items), total, err)
	}
	items, total, err = s.ListBerita(ctx, konten.ParamDaftar{Batas: 2, Geser: 1000})
	if err != nil || len(items) != 0 || total != 4 || items == nil {
		t.Fatalf("di luar rentang harus kosong non-nil dengan total benar: %v total=%d err=%v", items, total, err)
	}
}

func TestListBeritaPencarian(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	seedBerita(t, ctx, s)

	cases := []struct {
		nama, pola, want string
		total            int64
	}{
		{"tak peka huruf besar", "%BERITA%", "b3,b2,b1", 3},
		{"draft tidak ditemukan", "%rahasia%", "", 0},
		{"trash tidak ditemukan", "%terhapus%", "", 0},
		{"wildcard di-escape cocok literal", `%50\%\_%`, "b6", 1},
		{"persen literal tidak mencocokkan semua", `%a\%%`, "", 0},
		{"garis bawah literal", `%\_ %`, "b6", 1},
	}
	for _, c := range cases {
		t.Run(c.nama, func(t *testing.T) {
			items, total, err := s.ListBerita(ctx, konten.ParamDaftar{Pola: c.pola, Batas: 10})
			if err != nil {
				t.Fatal(err)
			}
			if slugs(items) != c.want || total != c.total {
				t.Fatalf("dapat %q total=%d, diharapkan %q total=%d", slugs(items), total, c.want, c.total)
			}
		})
	}
}

func TestRingkasanDariParagrafPertama(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	seedBerita(t, ctx, s)

	d, err := s.BeritaBySlug(ctx, "b1")
	if err != nil {
		t.Fatal(err)
	}
	if d.Ringkasan != "Paragraf pertama." || d.Isi != "Paragraf pertama.\n\nParagraf kedua" {
		t.Errorf("ringkasan/isi salah: %q / %q", d.Ringkasan, d.Isi)
	}
	panjang, err := s.BeritaBySlug(ctx, "b6")
	if err != nil {
		t.Fatal(err)
	}
	if len([]rune(panjang.Ringkasan)) != 200 {
		t.Errorf("ringkasan harus dipotong 200 karakter, dapat %d", len([]rune(panjang.Ringkasan)))
	}
}

func TestBeritaBySlugTidakDitemukan(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	seedBerita(t, ctx, s)
	for _, slug := range []string{"tidak-ada", "b4", "b5"} { // tidak ada, draft, di-trash
		if _, err := s.BeritaBySlug(ctx, slug); !errors.Is(err, konten.ErrTidakDitemukan) {
			t.Errorf("%s: diharapkan ErrTidakDitemukan, dapat %v", slug, err)
		}
	}
}

func TestDokumentasi(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	exec := func(q string, a ...any) {
		t.Helper()
		if _, err := s.Pool.Exec(ctx, q, a...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	ins := `INSERT INTO dokumentasi (slug, judul, tanggal_publikasi, status, created_at, created_by, updated_by, deleted_at)
	        VALUES ($1, $2, $3::date, $4, $5::timestamptz, 'u', 'u', CASE WHEN $6::boolean THEN now() END)`
	exec(ins, "d1", "Kegiatan Satu", "2026-02-01", "posted", "2026-02-01T08:00:00Z", false)
	exec(ins, "d2", "Kegiatan Dua", "2026-03-01", "posted", "2026-03-01T08:00:00Z", false)
	exec(ins, "d3", "Kegiatan Draft", "2026-04-01", "draft", "2026-04-01T08:00:00Z", false)
	exec(ins, "d4", "Kegiatan Terhapus", "2026-05-01", "posted", "2026-05-01T08:00:00Z", true)
	// gambar d1 disisipkan terbalik untuk membuktikan urutan dari kolom urutan
	exec(`INSERT INTO dokumentasi_gambar (dokumentasi_id, key, urutan) SELECT id, 'dokumentasi/d1/dua.jpg', 2 FROM dokumentasi WHERE slug='d1'`)
	exec(`INSERT INTO dokumentasi_gambar (dokumentasi_id, key, urutan) SELECT id, 'dokumentasi/d1/satu.jpg', 1 FROM dokumentasi WHERE slug='d1'`)

	items, total, err := s.ListDokumentasi(ctx, konten.ParamDaftar{Batas: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(items) != 2 || items[0].Slug != "d2" || items[1].Slug != "d1" {
		t.Fatalf("urutan/filter salah: %+v total=%d", items, total)
	}
	if items[0].GambarUtamaKey != "" || items[1].GambarUtamaKey != "dokumentasi/d1/satu.jpg" {
		t.Errorf("gambar utama salah: %+v", items)
	}
	// daftar publik memuat jumlah semua gambar (d2 tanpa gambar = 0, d1 = 2), juga pada hasil pencarian
	if items[0].JumlahGambar != 0 || items[1].JumlahGambar != 2 {
		t.Errorf("jumlah gambar salah: %+v", items)
	}

	cari, total, err := s.ListDokumentasi(ctx, konten.ParamDaftar{Pola: "%satu%", Batas: 10})
	if err != nil || total != 1 || len(cari) != 1 || cari[0].Slug != "d1" || cari[0].JumlahGambar != 2 {
		t.Fatalf("pencarian salah: %+v total=%d err=%v", cari, total, err)
	}

	d, err := s.DokumentasiBySlug(ctx, "d1")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Gambar) != 2 || d.Gambar[0].Key != "dokumentasi/d1/satu.jpg" ||
		d.Gambar[1].Key != "dokumentasi/d1/dua.jpg" {
		t.Errorf("gambar detail salah: %+v", d.Gambar)
	}
	for _, slug := range []string{"d3", "d4", "tidak-ada"} {
		if _, err := s.DokumentasiBySlug(ctx, slug); !errors.Is(err, konten.ErrTidakDitemukan) {
			t.Errorf("%s: diharapkan ErrTidakDitemukan, dapat %v", slug, err)
		}
	}
}

func TestTestimoniDanTim(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	exec := func(q string, a ...any) {
		t.Helper()
		if _, err := s.Pool.Exec(ctx, q, a...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	ins := `INSERT INTO testimoni (nama, teks, posisi, foto, status, created_at, created_by, updated_by, deleted_at)
	        VALUES ($1, 'teks', 'posisi', $2::text, $3, $4::timestamptz, 'u', 'u', CASE WHEN $5::boolean THEN now() END)`
	exec(ins, "Lama", "testimoni/1/f.jpg", "posted", "2026-01-01T08:00:00Z", false)
	exec(ins, "Baru", nil, "posted", "2026-02-01T08:00:00Z", false)
	exec(ins, "Draft", nil, "draft", "2026-03-01T08:00:00Z", false)
	exec(ins, "Terhapus", nil, "posted", "2026-04-01T08:00:00Z", true)

	ts, err := s.ListTestimoni(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 2 || ts[0].Nama != "Baru" || ts[1].Nama != "Lama" || ts[0].FotoKey != "" || ts[1].FotoKey != "testimoni/1/f.jpg" {
		t.Fatalf("testimoni salah: %+v", ts)
	}

	tim := `INSERT INTO tim (nama, posisi, bidang_riset, foto, urutan, created_by, updated_by) VALUES ($1::text, 'p', 'b', 'tim/' || $1::text || '.jpg', $2, 'u', 'u')`
	exec(tim, "Ketiga", 2)
	exec(tim, "Pertama", 1)
	exec(tim, "Kedua", 1) // urutan sama dengan Pertama, id lebih besar
	anggota, err := s.ListTim(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(anggota) != 3 || anggota[0].Nama != "Pertama" || anggota[1].Nama != "Kedua" || anggota[2].Nama != "Ketiga" {
		t.Fatalf("urutan tim salah: %+v", anggota)
	}
	if anggota[0].BidangRiset != "b" || anggota[0].FotoKey != "tim/Pertama.jpg" {
		t.Errorf("field tim salah: %+v", anggota[0])
	}
}

func TestDaftarKosongBukanNil(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	b, _, err := s.ListBerita(ctx, konten.ParamDaftar{Batas: 5})
	if err != nil || b == nil || len(b) != 0 {
		t.Errorf("berita: %v %v", b, err)
	}
	ts, err := s.ListTestimoni(ctx)
	if err != nil || ts == nil {
		t.Errorf("testimoni: %v %v", ts, err)
	}
	tm, err := s.ListTim(ctx)
	if err != nil || tm == nil {
		t.Errorf("tim: %v %v", tm, err)
	}
}
