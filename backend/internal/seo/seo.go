// Package seo merakit meta SEO, isi awal, dan data awal untuk halaman detail berita/dokumentasi, serta
// sitemap.xml dan robots.txt. Paket ini murni (tanpa HTTP dan database) dan semua teks yang masuk ke HTML
// di-escape di sini (docs/stack.md bagian 3, RP-12).
package seo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
)

// Penanda di kerangka HTML Astro yang diganti oleh Go. Masing-masing wajib muncul tepat sekali.
const (
	PenandaHead   = "<!--seo:head-->"
	PenandaKonten = "<!--seo:konten-->"
	PenandaData   = "<!--seo:data-->"
)

// Kerangka adalah HTML kerangka (hasil build Astro) yang bagian-bagiannya diganti saat merender.
type Kerangka struct {
	html string
}

// BuatKerangka memvalidasi bahwa ketiga penanda ada tepat sekali.
func BuatKerangka(isi string) (*Kerangka, error) {
	var galat []error
	for _, p := range []string{PenandaHead, PenandaKonten, PenandaData} {
		if n := strings.Count(isi, p); n != 1 {
			galat = append(galat, fmt.Errorf("penanda %s harus muncul tepat sekali, ditemukan %d", p, n))
		}
	}
	if err := errors.Join(galat...); err != nil {
		return nil, err
	}
	return &Kerangka{html: isi}, nil
}

// Render mengganti penanda dengan head, konten, dan data (sudah berupa HTML aman). Penggantian satu lintasan,
// jadi isi yang disisipkan tidak pernah ditafsirkan ulang sebagai penanda.
func (k *Kerangka) Render(head, konten, data string) []byte {
	return []byte(strings.NewReplacer(PenandaHead, head, PenandaKonten, konten, PenandaData, data).Replace(k.html))
}

// Halaman404 menyisipkan <meta name="robots" content="noindex"> sebelum </head> bila belum ada.
func Halaman404(isi []byte) []byte {
	s := string(isi)
	if strings.Contains(strings.ToLower(s), `name="robots"`) {
		return isi
	}
	i := strings.Index(strings.ToLower(s), "</head>")
	if i < 0 {
		return isi
	}
	return []byte(s[:i] + `<meta name="robots" content="noindex">` + s[i:])
}

// Halaman404Bawaan dipakai bila 404.html tidak ada.
func Halaman404Bawaan(namaSitus string) []byte {
	return []byte(`<!doctype html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">` +
		`<meta name="robots" content="noindex"><title>Halaman tidak ditemukan | ` + html.EscapeString(namaSitus) + `</title></head>` +
		`<body><h1>Halaman tidak ditemukan</h1><p><a href="/">Kembali ke beranda</a></p></body></html>`)
}

// Meta adalah data untuk blok <head>.
type Meta struct {
	NamaSitus string
	Judul     string // sudah termasuk nama situs, mis. "Judul | AI Center"
	Deskripsi string
	Canonical string
	Gambar    string // URL absolut; kosong = tanpa og:image
	JSONLD    any
}

// Head menghasilkan HTML untuk <head>: title, description, canonical, OG, Twitter, dan JSON-LD.
func (m Meta) Head() (string, error) {
	e := html.EscapeString
	var b strings.Builder
	b.WriteString("<title>" + e(m.Judul) + "</title>\n")
	b.WriteString(`<meta name="description" content="` + e(m.Deskripsi) + `">` + "\n")
	b.WriteString(`<link rel="canonical" href="` + e(m.Canonical) + `">` + "\n")
	og := func(prop, nilai string) {
		b.WriteString(`<meta property="` + prop + `" content="` + e(nilai) + `">` + "\n")
	}
	tw := func(nama, nilai string) {
		b.WriteString(`<meta name="` + nama + `" content="` + e(nilai) + `">` + "\n")
	}
	og("og:type", "article")
	og("og:title", m.Judul)
	og("og:description", m.Deskripsi)
	og("og:url", m.Canonical)
	og("og:site_name", m.NamaSitus)
	tw("twitter:card", "summary_large_image")
	tw("twitter:title", m.Judul)
	tw("twitter:description", m.Deskripsi)
	if m.Gambar != "" {
		og("og:image", m.Gambar)
		tw("twitter:image", m.Gambar)
	}
	if m.JSONLD != nil {
		s, err := ScriptJSON(`type="application/ld+json"`, m.JSONLD)
		if err != nil {
			return "", err
		}
		b.WriteString(s + "\n")
	}
	return b.String(), nil
}

// ScriptJSON menulis <script {atribut}>JSON</script> dengan JSON di-escape HTML (<, >, &, U+2028/9), sehingga
// isi data tidak bisa menutup tag script.
func ScriptJSON(atribut string, v any) (string, error) {
	mentah, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("seo: encode JSON: %w", err)
	}
	return scriptDariJSON(atribut, mentah), nil
}

// DataAwal membungkus JSON yang sudah jadi (mis. respons API publik) sebagai <script id="data-awal">.
func DataAwal(jsonMentah []byte) string {
	return scriptDariJSON(`id="data-awal" type="application/json"`, bytes.TrimSpace(jsonMentah))
}

func scriptDariJSON(atribut string, mentah []byte) string {
	var b bytes.Buffer
	json.HTMLEscape(&b, mentah)
	return "<script " + atribut + ">" + b.String() + "</script>"
}

const maksDeskripsi = 160

// DeskripsiBerita mengambil awal isi: spasi dan baris baru dirapatkan, dipotong di batas kata pada
// maksimum 160 karakter (termasuk "…") bila lebih panjang.
func DeskripsiBerita(isi string) string {
	t := strings.Join(strings.Fields(isi), " ")
	r := []rune(t)
	if len(r) <= maksDeskripsi {
		return t
	}
	potong := r[:maksDeskripsi-1]
	for i := len(potong) - 1; i >= maksDeskripsi/2; i-- {
		if potong[i] == ' ' {
			potong = potong[:i]
			break
		}
	}
	return strings.TrimRight(string(potong), " ,.;:-") + "…"
}

// DeskripsiDokumentasi adalah deskripsi default dokumentasi (docs/database.md bagian 3.3).
func DeskripsiDokumentasi(judul, tanggal string) string {
	return "Dokumentasi kegiatan " + judul + " — " + tanggal
}

// Gambar adalah satu gambar dokumentasi untuk isi awal.
type Gambar struct {
	URL string
}

// KontenBerita menghasilkan <article> berisi judul dan satu paragraf per blok teks (dipisah baris kosong).
func KontenBerita(judul, isi string) string {
	var b strings.Builder
	b.WriteString(`<article id="konten-awal"><h1>` + html.EscapeString(judul) + "</h1>")
	for _, p := range strings.Split(strings.ReplaceAll(isi, "\r\n", "\n"), "\n\n") {
		if p = strings.TrimSpace(p); p != "" {
			b.WriteString("<p>" + html.EscapeString(p) + "</p>")
		}
	}
	b.WriteString("</article>")
	return b.String()
}

// KontenDokumentasi menghasilkan <article> berisi judul, deskripsi, dan gambar berurutan (alt = judul).
func KontenDokumentasi(judul, deskripsi string, gambar []Gambar) string {
	e := html.EscapeString
	var b strings.Builder
	b.WriteString(`<article id="konten-awal"><h1>` + e(judul) + "</h1><p>" + e(deskripsi) + "</p>")
	for _, g := range gambar {
		b.WriteString(`<figure><img src="` + e(g.URL) + `" alt="` + e(judul) + `" loading="lazy"></figure>`)
	}
	b.WriteString("</article>")
	return b.String()
}

// JSONLDBerita adalah data terstruktur NewsArticle (tanpa author: datanya tidak ada).
func JSONLDBerita(namaSitus, canonical, judul, tanggal, gambar string) map[string]any {
	m := map[string]any{
		"@context":         "https://schema.org",
		"@type":            "NewsArticle",
		"headline":         judul,
		"datePublished":    tanggal,
		"mainEntityOfPage": map[string]any{"@type": "WebPage", "@id": canonical},
		"publisher":        map[string]any{"@type": "Organization", "name": namaSitus},
	}
	if gambar != "" {
		m["image"] = []string{gambar}
	}
	return m
}

// JSONLDDokumentasi adalah data terstruktur ImageGallery berisi semua gambar.
func JSONLDDokumentasi(canonical, judul, deskripsi, tanggal string, gambar []string) map[string]any {
	m := map[string]any{
		"@context":         "https://schema.org",
		"@type":            "ImageGallery",
		"name":             judul,
		"description":      deskripsi,
		"datePublished":    tanggal,
		"mainEntityOfPage": map[string]any{"@type": "WebPage", "@id": canonical},
	}
	if len(gambar) > 0 {
		m["image"] = gambar
	}
	return m
}
