package admin

import (
	"errors"
	"strings"
	"testing"
)

func TestSlug(t *testing.T) {
	cases := []struct{ judul, want string }{
		{"Pelatihan AI untuk Dosen", "pelatihan-ai-untuk-dosen"},
		{"  Spasi   Berlebih  ", "spasi-berlebih"},
		{"Kerja Sama: UBAYA & Mitra (2026)!", "kerja-sama-ubaya-mitra-2026"},
		{"Café Résumé Naïve", "cafe-resume-naive"},
		{"Ünïcödé Çāñ", "unicode-can"},
		{"---Judul---", "judul"},
		{"a_b/c\\d.e", "a-b-c-d-e"},
		{"UPPER lower 123", "upper-lower-123"},
		{"", ""},
		{"!!!", ""},
		{"日本語", ""},
		{"Judul 日本語 campur", "judul-campur"},
		{"<script>alert(1)</script>", "script-alert-1-script"},
		{"../../etc/passwd", "etc-passwd"},
	}
	for _, c := range cases {
		if got := Slug(c.judul); got != c.want {
			t.Errorf("Slug(%q) = %q, diharapkan %q", c.judul, got, c.want)
		}
	}
}

func TestSlugDipotongDanTidakBerakhirDenganTandaHubung(t *testing.T) {
	got := Slug(strings.Repeat("kata ", 100))
	if len(got) > BatasSlug {
		t.Fatalf("panjang %d melebihi %d", len(got), BatasSlug)
	}
	if strings.HasSuffix(got, "-") || strings.HasPrefix(got, "-") {
		t.Errorf("tidak boleh diawali/diakhiri tanda hubung: %q", got)
	}
	// dipotong tepat di tanda hubung
	got = Slug(strings.Repeat("a", BatasSlug-1) + " b")
	if strings.HasSuffix(got, "-") || len(got) > BatasSlug {
		t.Errorf("pemotongan di batas salah: %q (%d)", got, len(got))
	}
}

func TestBarisSatu(t *testing.T) {
	ok := map[string]string{
		"Judul Biasa":         "Judul Biasa",
		"  rapat   spasi  ":   "rapat spasi",
		"dua\nbaris\tdan tab": "dua baris dan tab",
		"Kuliah Tamu é":       "Kuliah Tamu é",
	}
	for in, want := range ok {
		got, err := barisSatu("judul", in, 50)
		if err != nil || got != want {
			t.Errorf("%q: dapat %q %v, diharapkan %q", in, got, err, want)
		}
	}
	buruk := map[string]string{
		"":                         "wajib",
		"   \n\t ":                 "wajib",
		"ab\x00cd":                 "tidak diizinkan",
		"ab\x07cd":                 "tidak diizinkan",
		"ab\x1bcd":                 "tidak diizinkan",
		string([]byte{0xff, 0xfe}): "UTF-8",
		strings.Repeat("x", 51):    "maksimal 50",
	}
	for in, pesan := range buruk {
		_, err := barisSatu("judul", in, 50)
		var e *ErrInput
		if !errors.As(err, &e) || !strings.Contains(e.Pesan, pesan) {
			t.Errorf("%q: diharapkan ErrInput memuat %q, dapat %v", in, pesan, err)
		}
	}
	if _, err := barisSatu("judul", strings.Repeat("é", 50), 50); err != nil {
		t.Errorf("batas dihitung per karakter, bukan byte: %v", err)
	}
}

func TestBanyakBaris(t *testing.T) {
	cases := []struct{ in, want string }{
		{"satu paragraf", "satu paragraf"},
		{"a\r\nb", "a\nb"},
		{"a\rb", "a\nb"},
		{"p1\n\np2", "p1\n\np2"},
		{"p1\n\n\n\n\np2", "p1\n\np2"},
		{"\n\n  awal dan akhir  \n\n", "awal dan akhir"},
		{"kolom\tbertab", "kolom\tbertab"},
		{"p1\r\n\r\n\r\n\r\np2", "p1\n\np2"},
	}
	for _, c := range cases {
		got, err := banyakBaris("isi", c.in, 100)
		if err != nil || got != c.want {
			t.Errorf("%q: dapat %q %v, diharapkan %q", c.in, got, err, c.want)
		}
	}
	for _, in := range []string{"", " \n\n ", "a\x00b", "a\x1bb", strings.Repeat("x", 101)} {
		if _, err := banyakBaris("isi", in, 100); err == nil {
			t.Errorf("%q seharusnya ditolak", in)
		}
	}
}

func TestTanggal(t *testing.T) {
	for _, in := range []string{"2026-11-09", " 2026-02-28 ", "2024-02-29"} {
		if got, err := tanggal("t", in); err != nil || got != strings.TrimSpace(in) {
			t.Errorf("%q: %q %v", in, got, err)
		}
	}
	for _, in := range []string{"", "2026-13-01", "2026-02-30", "2025-02-29", "09-11-2026", "2026/11/09", "kemarin", "2026-11-09T00:00:00Z", "0000-00-00"} {
		if _, err := tanggal("t", in); err == nil {
			t.Errorf("%q seharusnya ditolak", in)
		}
	}
}

func TestValidasiBerita(t *testing.T) {
	ok := MentahBerita{Judul: "  Judul   Berita ", Isi: "Paragraf satu.\r\n\r\n\r\nParagraf dua.", TanggalPublikasi: "2026-11-09", Gambar: "berita/x.png"}
	in, err := ValidasiBerita(ok)
	if err != nil {
		t.Fatal(err)
	}
	if in.Judul != "Judul Berita" || in.Isi != "Paragraf satu.\n\nParagraf dua." || in.TanggalPublikasi != "2026-11-09" {
		t.Errorf("hasil perapian salah: %+v", in)
	}

	ubah := map[string]func(m *MentahBerita){
		"judul kosong":          func(m *MentahBerita) { m.Judul = "" },
		"judul terlalu panjang": func(m *MentahBerita) { m.Judul = strings.Repeat("a", BatasJudul+1) },
		"isi kosong":            func(m *MentahBerita) { m.Isi = "  " },
		"isi terlalu panjang":   func(m *MentahBerita) { m.Isi = strings.Repeat("a", BatasIsi+1) },
		"tanggal salah":         func(m *MentahBerita) { m.TanggalPublikasi = "besok" },
		"tanggal kosong":        func(m *MentahBerita) { m.TanggalPublikasi = "" },
		"gambar kosong":         func(m *MentahBerita) { m.Gambar = "" },
	}
	for nama, f := range ubah {
		m := ok
		f(&m)
		_, err := ValidasiBerita(m)
		var e *ErrInput
		if !errors.As(err, &e) {
			t.Errorf("%s: diharapkan ErrInput, dapat %v", nama, err)
		}
	}
	// tepat di batas diterima
	m := ok
	m.Judul = strings.Repeat("a", BatasJudul)
	m.Isi = strings.Repeat("b", BatasIsi)
	if _, err := ValidasiBerita(m); err != nil {
		t.Errorf("tepat di batas harus diterima: %v", err)
	}
}
