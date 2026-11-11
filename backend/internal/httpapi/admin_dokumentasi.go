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

// LayananDokumentasi adalah aturan bisnis dokumentasi untuk admin (diimplementasikan admin.DokumentasiLayanan; palsu di tes).
type LayananDokumentasi interface {
	Daftar(ctx context.Context, p admin.ParamDaftar) ([]admin.RingkasDokumentasi, int64, error)
	Detail(ctx context.Context, id int64) (admin.Dokumentasi, error)
	Buat(ctx context.Context, m admin.MentahDokumentasi, oleh string) (admin.Dokumentasi, error)
	Ubah(ctx context.Context, id int64, m admin.MentahDokumentasi, version int32, oleh string) (admin.Dokumentasi, error)
	Post(ctx context.Context, id int64, version int32, oleh string) (admin.Dokumentasi, error)
	Retract(ctx context.Context, id int64, version int32, oleh string) (admin.Dokumentasi, error)
	Hapus(ctx context.Context, id int64, oleh string) error
	Pulihkan(ctx context.Context, id int64, oleh string) (admin.Dokumentasi, error)
	HapusPermanen(ctx context.Context, id int64) error
}

type dokumentasiAdminHandler struct {
	log *slog.Logger
	svc LayananDokumentasi
}

type gambarDokumentasiAdminJSON struct {
	Key string `json:"key"`
	URL string `json:"url"`
	Alt string `json:"alt,omitempty"` // dihilangkan bila kosong (frontend memakai judul sebagai cadangan)
}

// dokumentasiAdminJSON adalah detail lengkap (dengan seluruh gambar berurutan).
type dokumentasiAdminJSON struct {
	ID               int64                        `json:"id"`
	Slug             string                       `json:"slug"`
	Judul            string                       `json:"judul"`
	TanggalPublikasi string                       `json:"tanggalPublikasi"`
	Gambar           []gambarDokumentasiAdminJSON `json:"gambar"`
	Status           string                       `json:"status"`
	Version          int32                        `json:"version"`
	DibuatPada       time.Time                    `json:"dibuatPada"`
	DibuatOleh       string                       `json:"dibuatOleh"`
	DiubahOleh       string                       `json:"diubahOleh"`
	DihapusPada      *time.Time                   `json:"dihapusPada,omitempty"` // hanya di trash
}

// dokumentasiAdminRingkasJSON adalah item daftar: hanya gambar utama dan jumlahnya.
type dokumentasiAdminRingkasJSON struct {
	ID               int64            `json:"id"`
	Slug             string           `json:"slug"`
	Judul            string           `json:"judul"`
	TanggalPublikasi string           `json:"tanggalPublikasi"`
	GambarUtama      *gambarAdminJSON `json:"gambarUtama,omitempty"`
	JumlahGambar     int              `json:"jumlahGambar"`
	Status           string           `json:"status"`
	Version          int32            `json:"version"`
	DibuatPada       time.Time        `json:"dibuatPada"`
	DibuatOleh       string           `json:"dibuatOleh"`
	DiubahOleh       string           `json:"diubahOleh"`
	DihapusPada      *time.Time       `json:"dihapusPada,omitempty"`
}

func bentukDokumentasiAdmin(d admin.Dokumentasi) dokumentasiAdminJSON {
	g := make([]gambarDokumentasiAdminJSON, 0, len(d.Gambar))
	for _, x := range d.Gambar {
		g = append(g, gambarDokumentasiAdminJSON{Key: x.Key, URL: urlMedia(x.Key), Alt: x.Alt})
	}
	return dokumentasiAdminJSON{
		ID: d.ID, Slug: d.Slug, Judul: d.Judul, TanggalPublikasi: d.TanggalPublikasi, Gambar: g, Status: d.Status, Version: d.Version,
		DibuatPada: d.DibuatPada, DibuatOleh: d.DibuatOleh, DiubahOleh: d.DiubahOleh, DihapusPada: d.DihapusPada,
	}
}

func bentukDokumentasiRingkas(r admin.RingkasDokumentasi) dokumentasiAdminRingkasJSON {
	j := dokumentasiAdminRingkasJSON{
		ID: r.ID, Slug: r.Slug, Judul: r.Judul, TanggalPublikasi: r.TanggalPublikasi, JumlahGambar: r.JumlahGambar,
		Status: r.Status, Version: r.Version, DibuatPada: r.DibuatPada, DibuatOleh: r.DibuatOleh, DiubahOleh: r.DiubahOleh,
		DihapusPada: r.DihapusPada,
	}
	if r.GambarUtamaKey != "" {
		j.GambarUtama = &gambarAdminJSON{Key: r.GambarUtamaKey, URL: urlMedia(r.GambarUtamaKey)}
	}
	return j
}

// gambarMasukJSON adalah satu gambar pada body. Daftar `gambar` menyatakan KEADAAN AKHIR berurutan (1-6 item):
// urutan array = urutan gambar, key yang sudah ada dipertahankan, key baru ditambahkan, key yang hilang dihapus.
type gambarMasukJSON struct {
	Key string `json:"key"`
	Alt string `json:"alt"`
}

// POST tidak boleh membawa version; PUT wajib.
type bodyBuatDokumentasi struct {
	Judul            string            `json:"judul"`
	TanggalPublikasi string            `json:"tanggalPublikasi"`
	Gambar           []gambarMasukJSON `json:"gambar"`
}

type bodyUbahDokumentasi struct {
	bodyBuatDokumentasi
	Version *int32 `json:"version"`
}

func (b bodyBuatDokumentasi) mentah() admin.MentahDokumentasi {
	g := make([]admin.GambarDokumentasi, 0, len(b.Gambar))
	for _, x := range b.Gambar {
		g = append(g, admin.GambarDokumentasi{Key: x.Key, Alt: x.Alt})
	}
	return admin.MentahDokumentasi{Judul: b.Judul, TanggalPublikasi: b.TanggalPublikasi, Gambar: g}
}

// routerDokumentasi mendaftarkan /api/admin/dokumentasi/*. Middleware login, CSRF, dan Origin dipasang oleh grup induk.
func (h dokumentasiAdminHandler) routerDokumentasi(r chi.Router) {
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

func (h dokumentasiAdminHandler) galat(w http.ResponseWriter, r *http.Request, err error) {
	petakanGalatAdmin(h.log, w, r, err)
}

func (h dokumentasiAdminHandler) daftar(w http.ResponseWriter, r *http.Request) {
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
	out := make([]dokumentasiAdminRingkasJSON, 0, len(items))
	for _, x := range items {
		out = append(out, bentukDokumentasiRingkas(x))
	}
	tulisJSON(w, http.StatusOK, daftarJSON{Items: out, Page: q.Page, Limit: q.Limit, Total: total})
}

func (h dokumentasiAdminHandler) detail(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	d, err := h.svc.Detail(r.Context(), id)
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisDetailAdmin(w, r, bentukDokumentasiAdmin(d))
}

func (h dokumentasiAdminHandler) buat(w http.ResponseWriter, r *http.Request) {
	var body bodyBuatDokumentasi
	if !bacaJSON(w, r, &body) {
		return
	}
	d, err := h.svc.Buat(r.Context(), body.mentah(), oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/admin/dokumentasi/"+strconv.FormatInt(d.ID, 10))
	tulisJSON(w, http.StatusCreated, bentukDokumentasiAdmin(d))
}

func (h dokumentasiAdminHandler) ubah(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	var body bodyUbahDokumentasi
	if !bacaJSON(w, r, &body) {
		return
	}
	version, ok := versiWajib(w, body.Version)
	if !ok {
		return
	}
	d, err := h.svc.Ubah(r.Context(), id, body.mentah(), version, oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukDokumentasiAdmin(d))
}

// aksiBerversi menjalankan post atau retract: body {"version": n} wajib.
func (h dokumentasiAdminHandler) aksiBerversi(w http.ResponseWriter, r *http.Request, jalankan func(ctx context.Context, id int64, version int32, oleh string) (admin.Dokumentasi, error)) {
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
	d, err := jalankan(r.Context(), id, version, oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukDokumentasiAdmin(d))
}

func (h dokumentasiAdminHandler) post(w http.ResponseWriter, r *http.Request) {
	h.aksiBerversi(w, r, h.svc.Post)
}

func (h dokumentasiAdminHandler) retract(w http.ResponseWriter, r *http.Request) {
	h.aksiBerversi(w, r, h.svc.Retract)
}

func (h dokumentasiAdminHandler) pulihkan(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	d, err := h.svc.Pulihkan(r.Context(), id, oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukDokumentasiAdmin(d))
}

func (h dokumentasiAdminHandler) hapus(w http.ResponseWriter, r *http.Request) {
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

func (h dokumentasiAdminHandler) hapusPermanen(w http.ResponseWriter, r *http.Request) {
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
