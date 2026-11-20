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

// LayananTim adalah aturan bisnis tim untuk admin (diimplementasikan admin.TimLayanan; palsu di tes).
type LayananTim interface {
	Daftar(ctx context.Context) ([]admin.Tim, error)
	Detail(ctx context.Context, id int64) (admin.Tim, error)
	Buat(ctx context.Context, m admin.MentahTim, oleh string) (admin.Tim, error)
	Ubah(ctx context.Context, id int64, m admin.MentahTim, version int32, oleh string) (admin.Tim, error)
	Hapus(ctx context.Context, id int64) error
}

type timAdminHandler struct {
	log *slog.Logger
	svc LayananTim
}

type timAdminJSON struct {
	ID          int64           `json:"id"`
	Nama        string          `json:"nama"`
	Posisi      string          `json:"posisi"`
	BidangRiset string          `json:"bidangRiset"`
	Foto        gambarAdminJSON `json:"foto"`
	Urutan      int32           `json:"urutan"` // hanya-baca; anggota baru otomatis di urutan akhir
	Version     int32           `json:"version"`
	DibuatPada  time.Time       `json:"dibuatPada"`
	DibuatOleh  string          `json:"dibuatOleh"`
	DiubahOleh  string          `json:"diubahOleh"`
}

func bentukTimAdmin(t admin.Tim) timAdminJSON {
	return timAdminJSON{
		ID: t.ID, Nama: t.Nama, Posisi: t.Posisi, BidangRiset: t.BidangRiset,
		Foto:   gambarAdminJSON{Key: t.FotoKey, URL: urlMedia(t.FotoKey)},
		Urutan: t.Urutan, Version: t.Version, DibuatPada: t.DibuatPada, DibuatOleh: t.DibuatOleh, DiubahOleh: t.DiubahOleh,
	}
}

// POST tidak boleh membawa version; PUT wajib.
type bodyBuatTim struct {
	Nama        string `json:"nama"`
	Posisi      string `json:"posisi"`
	BidangRiset string `json:"bidangRiset"`
	Foto        string `json:"foto"`
	Urutan      *int64 `json:"urutan"` // wajib; angka bulat >= 1
}

type bodyUbahTim struct {
	bodyBuatTim
	Version *int32 `json:"version"`
}

func (b bodyBuatTim) mentah() admin.MentahTim {
	return admin.MentahTim{Nama: b.Nama, Posisi: b.Posisi, BidangRiset: b.BidangRiset, Foto: b.Foto, Urutan: b.Urutan}
}

// routerTim mendaftarkan /api/admin/tim/*. Tim tanpa status dan trash: DELETE langsung permanen.
func (h timAdminHandler) routerTim(r chi.Router) {
	r.Get("/", h.daftar)
	r.Post("/", h.buat)
	r.Get("/{id}", h.detail)
	r.Put("/{id}", h.ubah)
	r.Delete("/{id}", h.hapus)
}

func (h timAdminHandler) galat(w http.ResponseWriter, r *http.Request, err error) {
	petakanGalatAdmin(h.log, w, r, err)
}

// daftar mengembalikan semua anggota berurutan (urutan, id), tanpa pagination dan tanpa filter.
func (h timAdminHandler) daftar(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Daftar(r.Context())
	if err != nil {
		h.galat(w, r, err)
		return
	}
	out := make([]timAdminJSON, 0, len(items))
	for _, t := range items {
		out = append(out, bentukTimAdmin(t))
	}
	tulisJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h timAdminHandler) detail(w http.ResponseWriter, r *http.Request) {
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
	tulisDetailAdmin(w, r, bentukTimAdmin(t))
}

func (h timAdminHandler) buat(w http.ResponseWriter, r *http.Request) {
	var body bodyBuatTim
	if !bacaJSON(w, r, &body) {
		return
	}
	t, err := h.svc.Buat(r.Context(), body.mentah(), oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/admin/tim/"+strconv.FormatInt(t.ID, 10))
	tulisJSON(w, http.StatusCreated, bentukTimAdmin(t))
}

func (h timAdminHandler) ubah(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	var body bodyUbahTim
	if !bacaJSON(w, r, &body) {
		return
	}
	version, ok := versiWajib(w, body.Version)
	if !ok {
		return
	}
	t, err := h.svc.Ubah(r.Context(), id, body.mentah(), version, oleh(r))
	if err != nil {
		h.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukTimAdmin(t))
}

// hapus menghapus anggota langsung permanen (tanpa version, sesuai database.md §6); 204.
func (h timAdminHandler) hapus(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	if err := h.svc.Hapus(r.Context(), id); err != nil {
		h.galat(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
