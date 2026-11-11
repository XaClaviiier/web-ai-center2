package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
)

const limitDaftarAdminDefault = 20

// LayananBerita adalah aturan bisnis berita untuk admin (diimplementasikan admin.BeritaLayanan; palsu di tes).
type LayananBerita interface {
	Daftar(ctx context.Context, p admin.ParamDaftar) ([]admin.Berita, int64, error)
	Detail(ctx context.Context, id int64) (admin.Berita, error)
	Buat(ctx context.Context, m admin.MentahBerita, oleh string) (admin.Berita, error)
	Ubah(ctx context.Context, id int64, m admin.MentahBerita, version int32, oleh string) (admin.Berita, error)
	Post(ctx context.Context, id int64, version int32, oleh string) (admin.Berita, error)
	Retract(ctx context.Context, id int64, version int32, oleh string) (admin.Berita, error)
	Hapus(ctx context.Context, id int64, oleh string) error
	Pulihkan(ctx context.Context, id int64, oleh string) (admin.Berita, error)
	HapusPermanen(ctx context.Context, id int64) error
}

type beritaAdminHandler struct {
	log *slog.Logger
	svc LayananBerita
}

// Bentuk JSON admin (camelCase). Gambar dikirim sebagai {key,url}; saat menulis, klien mengirim hanya key.
type gambarAdminJSON struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

type beritaAdminJSON struct {
	ID               int64           `json:"id"`
	Slug             string          `json:"slug"`
	Judul            string          `json:"judul"`
	Ringkasan        string          `json:"ringkasan"`
	Isi              *string         `json:"isi,omitempty"` // hanya di detail
	TanggalPublikasi string          `json:"tanggalPublikasi"`
	Gambar           gambarAdminJSON `json:"gambar"`
	Status           string          `json:"status"`
	Version          int32           `json:"version"`
	DibuatPada       time.Time       `json:"dibuatPada"`
	DibuatOleh       string          `json:"dibuatOleh"`
	DiubahOleh       string          `json:"diubahOleh"`
	DihapusPada      *time.Time      `json:"dihapusPada,omitempty"` // hanya di trash
}

func bentukBeritaAdmin(b admin.Berita, denganIsi bool) beritaAdminJSON {
	j := beritaAdminJSON{
		ID: b.ID, Slug: b.Slug, Judul: b.Judul, Ringkasan: b.Ringkasan, TanggalPublikasi: b.TanggalPublikasi,
		Gambar: gambarAdminJSON{Key: b.GambarKey, URL: urlMedia(b.GambarKey)},
		Status: b.Status, Version: b.Version, DibuatPada: b.DibuatPada, DibuatOleh: b.DibuatOleh, DiubahOleh: b.DiubahOleh,
		DihapusPada: b.DihapusPada,
	}
	if denganIsi {
		isi := b.Isi
		j.Isi = &isi
	}
	return j
}

// Body request. POST tidak boleh membawa version; PUT, post, dan retract wajib.
type bodyBuatBerita struct {
	Judul            string `json:"judul"`
	Isi              string `json:"isi"`
	TanggalPublikasi string `json:"tanggalPublikasi"`
	Gambar           string `json:"gambar"`
}

type bodyUbahBerita struct {
	bodyBuatBerita
	Version *int32 `json:"version"`
}

func (b bodyBuatBerita) mentah() admin.MentahBerita {
	return admin.MentahBerita{Judul: b.Judul, Isi: b.Isi, TanggalPublikasi: b.TanggalPublikasi, Gambar: b.Gambar}
}

// routerBerita mendaftarkan /api/admin/berita/*. Middleware login, CSRF, dan Origin dipasang oleh grup induk.
func (h beritaAdminHandler) routerBerita(r chi.Router) {
	r.Get("/", h.daftar)
	r.Post("/", h.buat)
	r.Get("/{id}", h.detail)
	r.Put("/{id}", h.ubah)
	r.Delete("/{id}", h.hapus)
	r.Post("/{id}/post", h.post)
	r.Post("/{id}/retract", h.retract)
	r.Post("/{id}/restore", h.pulihkan)
	r.Delete("/{id}/permanen", h.hapusPermanen)
}

func (h beritaAdminHandler) daftar(w http.ResponseWriter, r *http.Request) {
	filter, ok := filterStatus(w, r)
	if !ok {
		return
	}
	q, err := parseDaftarLimit(r.URL.Query(), limitDaftarAdminDefault)
	if err != nil {
		tulisGalat(w, http.StatusBadRequest, kodeQueryTidakValid, err.Error())
		return
	}
	items, total, err := h.svc.Daftar(r.Context(), admin.ParamDaftar{Filter: filter, Pola: q.Pola, Batas: int32(q.Limit), Geser: int64(q.Page-1) * int64(q.Limit)})
	if err != nil {
		h.galat(w, r, err)
		return
	}
	out := make([]beritaAdminJSON, 0, len(items))
	for _, b := range items {
		out = append(out, bentukBeritaAdmin(b, false))
	}
	tulisJSON(w, http.StatusOK, daftarJSON{Items: out, Page: q.Page, Limit: q.Limit, Total: total})
}

func (h beritaAdminHandler) detail(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	b, err := h.svc.Detail(r.Context(), id)
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisDetailAdmin(w, r, bentukBeritaAdmin(b, true))
}

func (h beritaAdminHandler) buat(w http.ResponseWriter, r *http.Request) {
	var body bodyBuatBerita
	if !bacaJSON(w, r, &body) {
		return
	}
	b, err := h.svc.Buat(r.Context(), body.mentah(), oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/admin/berita/"+strconv.FormatInt(b.ID, 10))
	tulisJSON(w, http.StatusCreated, bentukBeritaAdmin(b, true))
}

func (h beritaAdminHandler) ubah(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	var body bodyUbahBerita
	if !bacaJSON(w, r, &body) {
		return
	}
	version, ok := versiWajib(w, body.Version)
	if !ok {
		return
	}
	b, err := h.svc.Ubah(r.Context(), id, body.mentah(), version, oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukBeritaAdmin(b, true))
}

// aksiBerversi menjalankan post atau retract: body {"version": n} wajib.
func (h beritaAdminHandler) aksiBerversi(w http.ResponseWriter, r *http.Request, jalankan func(ctx context.Context, id int64, version int32, oleh string) (admin.Berita, error)) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	var body bodyVersi
	if !bacaJSON(w, r, &body) {
		return
	}
	version, ok := versiWajib(w, body.Version)
	if !ok {
		return
	}
	b, err := jalankan(r.Context(), id, version, oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukBeritaAdmin(b, true))
}

func (h beritaAdminHandler) post(w http.ResponseWriter, r *http.Request) {
	h.aksiBerversi(w, r, h.svc.Post)
}
func (h beritaAdminHandler) retract(w http.ResponseWriter, r *http.Request) {
	h.aksiBerversi(w, r, h.svc.Retract)
}

func (h beritaAdminHandler) pulihkan(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	b, err := h.svc.Pulihkan(r.Context(), id, oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukBeritaAdmin(b, true))
}

func (h beritaAdminHandler) hapus(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	if err := h.svc.Hapus(r.Context(), id, oleh(r)); err != nil {
		h.galat(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h beritaAdminHandler) hapusPermanen(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	if err := h.svc.HapusPermanen(r.Context(), id); err != nil {
		h.galat(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h beritaAdminHandler) galat(w http.ResponseWriter, r *http.Request, err error) {
	petakanGalatAdmin(h.log, w, r, err)
}
