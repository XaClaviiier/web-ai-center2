package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/cache"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/seo"
)

const kodeKerangkaTidakAda = "kerangka_tidak_ada"

// DepsSEO adalah dependensi halaman detail ber-SEO, sitemap.xml, dan robots.txt (docs/stack.md bagian 3, RP-12).
type DepsSEO struct {
	BaseURL   string // PUBLIC_BASE_URL tanpa garis miring akhir
	NamaSitus string

	// Kerangka HTML Astro; nil = belum ada build (route detail membalas 503 kerangka_tidak_ada).
	KerangkaBerita      *seo.Kerangka
	KerangkaDokumentasi *seo.Kerangka

	// Isi 404.html bila ada (nil = halaman 404 bawaan).
	Halaman404 []byte

	Sitemap konten.BacaSitemap // nil = sitemap hanya berisi halaman tetap
}

// seoHandler memakai publik untuk jalur cache, ETag, dan limiter yang sama dengan API publik.
type seoHandler struct {
	publik
	s DepsSEO
}

func (h seoHandler) tulis404(w http.ResponseWriter) {
	isi := h.s.Halaman404
	if isi == nil {
		isi = seo.Halaman404Bawaan(h.s.NamaSitus)
	}
	hd := w.Header()
	hd.Set("Content-Type", tipeHTML)
	hd.Set("X-Robots-Tag", "noindex")
	hd.Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(isi)
}

func (h seoHandler) berita(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !slugSah(slug) {
		h.tulis404(w)
		return
	}
	k := h.s.KerangkaBerita
	if k == nil {
		tulisGalat(w, http.StatusServiceUnavailable, kodeKerangkaTidakAda, "Kerangka halaman belum tersedia")
		return
	}
	h.layaniBytes(w, r, cache.Berita, "html:"+slug, tipeHTML, h.tulis404, func(ctx context.Context) ([]byte, bool, error) {
		b, err := h.src.BeritaBySlug(ctx, slug)
		if err != nil {
			return nil, false, err
		}
		data, err := json.Marshal(beritaDetailJSON{beritaRingkasJSON: beritaRingkas(b.BeritaRingkas), Isi: b.Isi})
		if err != nil {
			return nil, false, fmt.Errorf("encode data awal: %w", err)
		}
		canonical := h.s.BaseURL + "/berita/" + b.Slug
		gambar := h.urlAbsolut(b.GambarKey)
		deskripsi := seo.DeskripsiBerita(b.Isi)
		head, err := seo.Meta{
			NamaSitus: h.s.NamaSitus, Judul: b.Judul + " | " + h.s.NamaSitus, Deskripsi: deskripsi,
			Canonical: canonical, Gambar: gambar,
			JSONLD: seo.JSONLDBerita(h.s.NamaSitus, canonical, b.Judul, b.TanggalPublikasi, gambar),
		}.Head()
		if err != nil {
			return nil, false, err
		}
		return k.Render(head, seo.KontenBerita(b.Judul, b.Isi), seo.DataAwal(data)), true, nil
	})
}

func (h seoHandler) dokumentasi(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !slugSah(slug) {
		h.tulis404(w)
		return
	}
	k := h.s.KerangkaDokumentasi
	if k == nil {
		tulisGalat(w, http.StatusServiceUnavailable, kodeKerangkaTidakAda, "Kerangka halaman belum tersedia")
		return
	}
	h.layaniBytes(w, r, cache.Dokumentasi, "html:"+slug, tipeHTML, h.tulis404, func(ctx context.Context) ([]byte, bool, error) {
		d, err := h.src.DokumentasiBySlug(ctx, slug)
		if err != nil {
			return nil, false, err
		}
		jsonGambar := make([]gambarDokumentasiJSON, 0, len(d.Gambar))
		isiGambar := make([]seo.Gambar, 0, len(d.Gambar))
		urlGambar := make([]string, 0, len(d.Gambar))
		for _, g := range d.Gambar {
			jsonGambar = append(jsonGambar, gambarDokumentasiJSON{URL: urlMedia(g.Key), Alt: g.Alt})
			isiGambar = append(isiGambar, seo.Gambar{URL: urlMedia(g.Key), Alt: g.Alt})
			urlGambar = append(urlGambar, h.urlAbsolut(g.Key))
		}
		data, err := json.Marshal(dokumentasiDetailJSON{
			ID: d.ID, Slug: d.Slug, Judul: d.Judul, TanggalPublikasi: d.TanggalPublikasi, Gambar: jsonGambar,
		})
		if err != nil {
			return nil, false, fmt.Errorf("encode data awal: %w", err)
		}
		canonical := h.s.BaseURL + "/dokumentasi/" + d.Slug
		deskripsi := seo.DeskripsiDokumentasi(d.Judul, d.TanggalPublikasi)
		utama := ""
		if len(urlGambar) > 0 {
			utama = urlGambar[0] // headline = urutan 1
		}
		head, err := seo.Meta{
			NamaSitus: h.s.NamaSitus, Judul: d.Judul + " | " + h.s.NamaSitus, Deskripsi: deskripsi,
			Canonical: canonical, Gambar: utama,
			JSONLD: seo.JSONLDDokumentasi(canonical, d.Judul, deskripsi, d.TanggalPublikasi, urlGambar),
		}.Head()
		if err != nil {
			return nil, false, err
		}
		return k.Render(head, seo.KontenDokumentasi(d.Judul, deskripsi, isiGambar), seo.DataAwal(data)), true, nil
	})
}

func (h seoHandler) urlAbsolut(key string) string {
	if key == "" {
		return ""
	}
	return h.s.BaseURL + urlMedia(key)
}

// sitemap: halaman tetap + semua berita dan dokumentasi posted. Dicache dan dibatalkan bersama berita/dokumentasi.
func (h seoHandler) sitemap(w http.ResponseWriter, r *http.Request) {
	h.layaniBytes(w, r, cache.Sitemap, "xml", tipeXML, tulisTidakDitemukan, func(ctx context.Context) ([]byte, bool, error) {
		urls := make([]seo.URL, 0, len(seo.HalamanTetap))
		for _, p := range seo.HalamanTetap {
			urls = append(urls, seo.URL{Path: p})
		}
		if h.s.Sitemap != nil {
			berita, err := h.s.Sitemap.SitemapBerita(ctx, seo.BatasSitemap)
			if err != nil {
				return nil, false, err
			}
			for _, e := range berita {
				urls = append(urls, seo.URL{Path: "/berita/" + e.Slug, Lastmod: e.Tanggal})
			}
			dok, err := h.s.Sitemap.SitemapDokumentasi(ctx, seo.BatasSitemap)
			if err != nil {
				return nil, false, err
			}
			for _, e := range dok {
				urls = append(urls, seo.URL{Path: "/dokumentasi/" + e.Slug, Lastmod: e.Tanggal})
			}
		}
		data, dipotong, err := seo.Sitemap(h.s.BaseURL, urls)
		if err != nil {
			return nil, false, err
		}
		if dipotong {
			h.log.Warn("sitemap melebihi batas dan dipotong", "batas", seo.BatasSitemap, "total", len(urls))
		}
		return data, true, nil
	})
}

func (h seoHandler) robots(w http.ResponseWriter, r *http.Request) {
	kirimPublik(w, r, "text/plain; charset=utf-8", []byte(seo.Robots(h.s.BaseURL)))
}
