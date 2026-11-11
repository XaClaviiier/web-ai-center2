package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/cache"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/ratelimit"
)

const kodeQueryTidakValid = "invalid_query"

// urlMedia mengubah key objek RustFS menjadi URL relatif satu origin (/media/{key}).
// Key dibuat server (uuid), jadi tidak perlu di-encode. Kosong tetap kosong.
func urlMedia(key string) string {
	if key == "" {
		return ""
	}
	return "/media/" + key
}

// Bentuk JSON publik: camelCase; field opsional dihilangkan bila kosong.

type daftarJSON struct {
	Items any   `json:"items"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type beritaRingkasJSON struct {
	ID               int64  `json:"id"`
	Slug             string `json:"slug"`
	Judul            string `json:"judul"`
	Ringkasan        string `json:"ringkasan"`
	Gambar           string `json:"gambar"`
	TanggalPublikasi string `json:"tanggalPublikasi"`
}

type beritaDetailJSON struct {
	beritaRingkasJSON
	Isi string `json:"isi"`
}

type dokumentasiRingkasJSON struct {
	ID               int64  `json:"id"`
	Slug             string `json:"slug"`
	Judul            string `json:"judul"`
	TanggalPublikasi string `json:"tanggalPublikasi"`
	GambarUtama      string `json:"gambarUtama,omitempty"`
}

type gambarDokumentasiJSON struct {
	URL string `json:"url"`
	Alt string `json:"alt,omitempty"`
}

type dokumentasiDetailJSON struct {
	ID               int64                   `json:"id"`
	Slug             string                  `json:"slug"`
	Judul            string                  `json:"judul"`
	TanggalPublikasi string                  `json:"tanggalPublikasi"`
	Gambar           []gambarDokumentasiJSON `json:"gambar"`
}

type testimoniJSON struct {
	ID     int64  `json:"id"`
	Nama   string `json:"nama"`
	Teks   string `json:"teks"`
	Posisi string `json:"posisi"`
	Foto   string `json:"foto,omitempty"`
}

type anggotaTimJSON struct {
	ID          int64  `json:"id"`
	Nama        string `json:"nama"`
	Posisi      string `json:"posisi"`
	BidangRiset string `json:"bidangRiset"`
	Foto        string `json:"foto"`
}

func beritaRingkas(b konten.BeritaRingkas) beritaRingkasJSON {
	return beritaRingkasJSON{
		ID: b.ID, Slug: b.Slug, Judul: b.Judul, Ringkasan: b.Ringkasan,
		Gambar: urlMedia(b.GambarKey), TanggalPublikasi: b.TanggalPublikasi,
	}
}

// publik memegang handler baca konten publik.
type publik struct {
	log *slog.Logger
	src konten.BacaPublik

	cache   *cache.Cache      // nil = tanpa cache
	limiter ratelimit.Limiter // nil = tanpa limiter
	batas   int               // batas miss per IP per menit
}

func (h publik) daftarBerita(w http.ResponseWriter, r *http.Request) {
	q, err := parseDaftar(r.URL.Query())
	if err != nil {
		tulisGalat(w, http.StatusBadRequest, kodeQueryTidakValid, err.Error())
		return
	}
	h.layani(w, r, cache.Berita, subDaftar(q), func(ctx context.Context) (any, bool, error) {
		items, total, err := h.src.ListBerita(ctx, q.param())
		if err != nil {
			return nil, false, err
		}
		out := make([]beritaRingkasJSON, 0, len(items))
		for _, b := range items {
			out = append(out, beritaRingkas(b))
		}
		return daftarJSON{Items: out, Page: q.Page, Limit: q.Limit, Total: total}, bolehSimpanDaftar(q, len(items)), nil
	})
}

func (h publik) detailBerita(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !slugSah(slug) {
		tulisTidakDitemukan(w)
		return
	}
	h.layani(w, r, cache.Berita, "detail:"+slug, func(ctx context.Context) (any, bool, error) {
		b, err := h.src.BeritaBySlug(ctx, slug)
		if err != nil {
			return nil, false, err
		}
		return beritaDetailJSON{beritaRingkasJSON: beritaRingkas(b.BeritaRingkas), Isi: b.Isi}, true, nil
	})
}

func (h publik) daftarDokumentasi(w http.ResponseWriter, r *http.Request) {
	q, err := parseDaftar(r.URL.Query())
	if err != nil {
		tulisGalat(w, http.StatusBadRequest, kodeQueryTidakValid, err.Error())
		return
	}
	h.layani(w, r, cache.Dokumentasi, subDaftar(q), func(ctx context.Context) (any, bool, error) {
		items, total, err := h.src.ListDokumentasi(ctx, q.param())
		if err != nil {
			return nil, false, err
		}
		out := make([]dokumentasiRingkasJSON, 0, len(items))
		for _, d := range items {
			out = append(out, dokumentasiRingkasJSON{
				ID: d.ID, Slug: d.Slug, Judul: d.Judul, TanggalPublikasi: d.TanggalPublikasi,
				GambarUtama: urlMedia(d.GambarUtamaKey),
			})
		}
		return daftarJSON{Items: out, Page: q.Page, Limit: q.Limit, Total: total}, bolehSimpanDaftar(q, len(items)), nil
	})
}

func (h publik) detailDokumentasi(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !slugSah(slug) {
		tulisTidakDitemukan(w)
		return
	}
	h.layani(w, r, cache.Dokumentasi, "detail:"+slug, func(ctx context.Context) (any, bool, error) {
		d, err := h.src.DokumentasiBySlug(ctx, slug)
		if err != nil {
			return nil, false, err
		}
		gambar := make([]gambarDokumentasiJSON, 0, len(d.Gambar))
		for _, g := range d.Gambar {
			gambar = append(gambar, gambarDokumentasiJSON{URL: urlMedia(g.Key), Alt: g.Alt})
		}
		return dokumentasiDetailJSON{
			ID: d.ID, Slug: d.Slug, Judul: d.Judul, TanggalPublikasi: d.TanggalPublikasi, Gambar: gambar,
		}, true, nil
	})
}

func (h publik) daftarTestimoni(w http.ResponseWriter, r *http.Request) {
	h.layani(w, r, cache.Testimoni, "list", func(ctx context.Context) (any, bool, error) {
		items, err := h.src.ListTestimoni(ctx)
		if err != nil {
			return nil, false, err
		}
		out := make([]testimoniJSON, 0, len(items))
		for _, t := range items {
			out = append(out, testimoniJSON{ID: t.ID, Nama: t.Nama, Teks: t.Teks, Posisi: t.Posisi, Foto: urlMedia(t.FotoKey)})
		}
		return map[string]any{"items": out}, true, nil
	})
}

func (h publik) daftarTim(w http.ResponseWriter, r *http.Request) {
	h.layani(w, r, cache.Tim, "list", func(ctx context.Context) (any, bool, error) {
		items, err := h.src.ListTim(ctx)
		if err != nil {
			return nil, false, err
		}
		out := make([]anggotaTimJSON, 0, len(items))
		for _, a := range items {
			out = append(out, anggotaTimJSON{
				ID: a.ID, Nama: a.Nama, Posisi: a.Posisi, BidangRiset: a.BidangRiset, Foto: urlMedia(a.FotoKey),
			})
		}
		return map[string]any{"items": out}, true, nil
	})
}

func tulisTidakDitemukan(w http.ResponseWriter) {
	tulisGalat(w, http.StatusNotFound, kodeTidakDitemukan, "Sumber tidak ditemukan")
}

// galatInternal mencatat detail galat ke log dan membalas 500 generik (tanpa detail internal).
func (h publik) galatInternal(w http.ResponseWriter, r *http.Request, err error) {
	h.log.Error("gagal membaca konten", "err", err, "path", r.URL.Path, "request_id", middleware.GetReqID(r.Context()))
	tulisGalat(w, http.StatusInternalServerError, kodeInternal, "Terjadi kesalahan pada server")
}
