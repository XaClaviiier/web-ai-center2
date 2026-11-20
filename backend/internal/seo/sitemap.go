package seo

import (
	"encoding/xml"
	"fmt"
)

// BatasSitemap adalah jumlah URL maksimum per berkas sitemap (protokol sitemaps.org).
const BatasSitemap = 50000

// HalamanTetap adalah rute publik tetap yang selalu masuk sitemap ("/" = beranda).
var HalamanTetap = []string{
	"/", "/berita", "/dokumentasi", "/tim", "/tentang-kami", "/layanan/pelatihan", "/layanan/inference-solutions",
}

// URL adalah satu entri sitemap; Path diawali "/".
type URL struct {
	Path    string
	Lastmod string // YYYY-MM-DD; kosong = tanpa lastmod
}

type xmlURL struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod,omitempty"`
}

type xmlUrlset struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URL     []xmlURL `xml:"url"`
}

// Sitemap membuat sitemap.xml dari daftar URL; lebih dari BatasSitemap dipotong (dipotong=true).
func Sitemap(base string, urls []URL) (data []byte, dipotong bool, err error) {
	if len(urls) > BatasSitemap {
		urls, dipotong = urls[:BatasSitemap], true
	}
	set := xmlUrlset{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9", URL: make([]xmlURL, 0, len(urls))}
	for _, u := range urls {
		set.URL = append(set.URL, xmlURL{Loc: base + u.Path, Lastmod: u.Lastmod})
	}
	isi, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, false, fmt.Errorf("seo: encode sitemap: %w", err)
	}
	return append([]byte(xml.Header), append(isi, '\n')...), dipotong, nil
}

// Robots menghasilkan robots.txt: semua boleh kecuali /admin dan /api/, plus lokasi sitemap.
func Robots(base string) string {
	return "User-agent: *\nAllow: /\nDisallow: /admin\nDisallow: /api/\n\nSitemap: " + base + "/sitemap.xml\n"
}
