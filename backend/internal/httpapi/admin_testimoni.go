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

// LayananTestimoni adalah aturan bisnis testimoni untuk admin (diimplementasikan admin.TestimoniLayanan; palsu di tes).
type LayananTestimoni interface {
	Daftar(ctx context.Context, filter string) ([]admin.Testimoni, error)
	Detail(ctx context.Context, id int64) (admin.Testimoni, error)
	Buat(ctx context.Context, m admin.MentahTestimoni, oleh string) (admin.Testimoni, error)
	Ubah(ctx context.Context, id int64, m admin.MentahTestimoni, version int32, oleh string) (admin.Testimoni, error)
	Post(ctx context.Context, id int64, version int32, oleh string) (admin.Testimoni, error)
	Retract(ctx context.Context, id int64, version int32, oleh string) (admin.Testimoni, error)
	Hapus(ctx context.Context, id int64, oleh string) error
	Pulihkan(ctx context.Context, id int64, oleh string) (admin.Testimoni, error)
	HapusPermanen(ctx context.Context, id int64) error
}

type testimoniAdminHandler struct {
	log *slog.Logger
	svc LayananTestimoni
}

type testimoniAdminJSON struct {
	ID          int64            `json:"id"`
	Nama        string           `json:"nama"`
	Teks        string           `json:"teks"`
	Posisi      string           `json:"posisi"`
	Foto        *gambarAdminJSON `json:"foto,omitempty"` // dihilangkan bila tanpa foto (tampil inisial nama)
	Status      string           `json:"status"`
	Version     int32            `json:"version"`
	DibuatPada  time.Time        `json:"dibuatPada"`
	DibuatOleh  string           `json:"dibuatOleh"`
	DiubahOleh  string           `json:"diubahOleh"`
	DihapusPada *time.Time       `json:"dihapusPada,omitempty"` // hanya di trash
}

func bentukTestimoniAdmin(t admin.Testimoni) testimoniAdminJSON {
	j := testimoniAdminJSON{
		ID: t.ID, Nama: t.Nama, Teks: t.Teks, Posisi: t.Posisi, Status: t.Status, Version: t.Version,
		DibuatPada: t.DibuatPada, DibuatOleh: t.DibuatOleh, DiubahOleh: t.DiubahOleh, DihapusPada: t.DihapusPada,
	}
	if t.FotoKey != "" {
		j.Foto = &gambarAdminJSON{Key: t.FotoKey, URL: urlMedia(t.FotoKey)}
	}
	return j
}

// Field foto WAJIB hadir: string (key) atau null (tanpa foto). POST tidak boleh membawa version; PUT wajib.
type bodyBuatTestimoni struct {
	Nama   string       `json:"nama"`
	Teks   string       `json:"teks"`
	Posisi string       `json:"posisi"`
	Foto   fotoOpsional `json:"foto"`
}

type bodyUbahTestimoni struct {
	bodyBuatTestimoni
	Version *int32 `json:"version"`
}

func (b bodyBuatTestimoni) mentah() admin.MentahTestimoni {
	return admin.MentahTestimoni{Nama: b.Nama, Teks: b.Teks, Posisi: b.Posisi, Foto: b.Foto.Nilai}
}

// fotoHadir menolak body yang tidak menyertakan field foto sama sekali (bukan "tetap" dan bukan "hapus").
func fotoHadir(w http.ResponseWriter, f fotoOpsional) bool {
	if !f.Ada {
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "Field foto wajib hadir: isi key gambar, atau null bila tanpa foto")
		return false
	}
	return true
}

// routerTestimoni mendaftarkan /api/admin/testimoni/*. Alur sama dengan berita.
func (h testimoniAdminHandler) routerTestimoni(r chi.Router) {
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

func (h testimoniAdminHandler) galat(w http.ResponseWriter, r *http.Request, err error) {
	petakanGalatAdmin(h.log, w, r, err)
}

// daftar mengembalikan testimoni menurut ?status= tanpa pagination dan tanpa pencarian.
func (h testimoniAdminHandler) daftar(w http.ResponseWriter, r *http.Request) {
	filter, ok := filterStatus(w, r)
	if !ok {
		return
	}
	items, err := h.svc.Daftar(r.Context(), filter)
	if err != nil {
		h.galat(w, r, err)
		return
	}
	out := make([]testimoniAdminJSON, 0, len(items))
	for _, t := range items {
		out = append(out, bentukTestimoniAdmin(t))
	}
	tulisJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h testimoniAdminHandler) detail(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	t, err := h.svc.Detail(r.Context(), id)
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukTestimoniAdmin(t))
}

func (h testimoniAdminHandler) buat(w http.ResponseWriter, r *http.Request) {
	var body bodyBuatTestimoni
	if !bacaJSON(w, r, &body) || !fotoHadir(w, body.Foto) {
		return
	}
	t, err := h.svc.Buat(r.Context(), body.mentah(), oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/admin/testimoni/"+strconv.FormatInt(t.ID, 10))
	tulisJSON(w, http.StatusCreated, bentukTestimoniAdmin(t))
}

func (h testimoniAdminHandler) ubah(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	var body bodyUbahTestimoni
	if !bacaJSON(w, r, &body) {
		return
	}
	version, ok := versiWajib(w, body.Version)
	if !ok || !fotoHadir(w, body.Foto) {
		return
	}
	t, err := h.svc.Ubah(r.Context(), id, body.mentah(), version, oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukTestimoniAdmin(t))
}

// aksiBerversi menjalankan post atau retract: body {"version": n} wajib.
func (h testimoniAdminHandler) aksiBerversi(w http.ResponseWriter, r *http.Request, jalankan func(ctx context.Context, id int64, version int32, oleh string) (admin.Testimoni, error)) {
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
	t, err := jalankan(r.Context(), id, version, oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukTestimoniAdmin(t))
}

func (h testimoniAdminHandler) post(w http.ResponseWriter, r *http.Request) {
	h.aksiBerversi(w, r, h.svc.Post)
}

func (h testimoniAdminHandler) retract(w http.ResponseWriter, r *http.Request) {
	h.aksiBerversi(w, r, h.svc.Retract)
}

func (h testimoniAdminHandler) pulihkan(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	t, err := h.svc.Pulihkan(r.Context(), id, oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukTestimoniAdmin(t))
}

func (h testimoniAdminHandler) hapus(w http.ResponseWriter, r *http.Request) {
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

func (h testimoniAdminHandler) hapusPermanen(w http.ResponseWriter, r *http.Request) {
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
