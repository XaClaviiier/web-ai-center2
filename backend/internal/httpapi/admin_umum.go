package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
)

// Kode galat admin (stabil untuk klien).
const (
	kodeStatusTidakValid = "status_tidak_valid"
	kodeVersiBentrok     = "conflict" // dulu version_conflict; disamakan dengan frontend-final (keputusan 2026-11-13)
	kodeBukanDraft       = "bukan_draft"
	kodeGambarHilang     = "gambar_hilang"
)

// idDariPath membaca {id} dari path: bilangan bulat >= 1; selain itu dianggap tidak ada (404).
func idDariPath(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	return id, err == nil && id >= 1
}

// oleh mengembalikan identitas (sub) pengguna dari session untuk kolom audit created_by/updated_by.
func oleh(r *http.Request) string {
	s, _ := SessionDariContext(r.Context())
	return s.Sub
}

// bodyVersi adalah body aksi yang hanya membawa version (post, retract).
type bodyVersi struct {
	Version *int32 `json:"version"`
}

// versiWajib memastikan version hadir dan >= 1; bila tidak, 400 sudah ditulis dan false dikembalikan.
func versiWajib(w http.ResponseWriter, v *int32) (int32, bool) {
	if v == nil || *v < 1 {
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "Field version wajib diisi (bilangan bulat minimal 1)")
		return 0, false
	}
	return *v, true
}

// fotoOpsional mendeteksi apakah field foto HADIR dalam JSON dan nilainya: string (key) atau null (tanpa foto).
// Field yang tidak dikirim sama sekali dibedakan dari null agar tidak ada tebakan "tetap" atau "hapus".
type fotoOpsional struct {
	Ada   bool
	Nilai string // kosong = tanpa foto
}

func (f *fotoOpsional) UnmarshalJSON(b []byte) error {
	f.Ada = true
	if string(b) == "null" {
		f.Nilai = ""
		return nil
	}
	return json.Unmarshal(b, &f.Nilai)
}

// petakanGalatAdmin memetakan galat domain ke respons HTTP; selain itu 500 generik (detail hanya di log).
func petakanGalatAdmin(log *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	var in *admin.ErrInput
	switch {
	case errors.As(err, &in):
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, in.Pesan)
	case errors.Is(err, admin.ErrTidakDitemukan):
		tulisTidakDitemukan(w)
	case errors.Is(err, admin.ErrStatusTidakValid) && r.Method == http.MethodPut:
		// Simpan (PUT) pada item bukan draft: kode dan status mengikuti frontend-final.
		tulisGalat(w, http.StatusBadRequest, kodeBukanDraft, "Hanya data berstatus draft yang bisa diedit")
	case errors.Is(err, admin.ErrStatusTidakValid):
		tulisGalat(w, http.StatusConflict, kodeStatusTidakValid, "Aksi tidak diizinkan pada status konten saat ini")
	case errors.Is(err, admin.ErrVersiBentrok):
		tulisGalat(w, http.StatusConflict, kodeVersiBentrok, "Konten sudah diubah pihak lain; muat ulang lalu coba lagi")
	case errors.Is(err, admin.ErrGambarHilang):
		tulisGalat(w, http.StatusConflict, kodeGambarHilang, "Gambar konten tidak ditemukan di penyimpanan; unggah ulang")
	case errors.Is(err, admin.ErrStorage):
		log.Error("storage gambar bermasalah", "err", err, "path", r.URL.Path, "request_id", middleware.GetReqID(r.Context()))
		tulisGalat(w, http.StatusServiceUnavailable, kodeStorageBermasalah, "Penyimpanan gambar sedang tidak tersedia")
	default:
		log.Error("gagal memproses konten admin", "err", err, "method", r.Method, "path", r.URL.Path, "request_id", middleware.GetReqID(r.Context()))
		tulisGalat(w, http.StatusInternalServerError, kodeInternal, "Terjadi kesalahan pada server")
	}
}

// filterStatus membaca ?status= (default semua); bila tidak sah, 400 sudah ditulis dan false dikembalikan.
func filterStatus(w http.ResponseWriter, r *http.Request) (string, bool) {
	v := r.URL.Query()
	if !v.Has("status") {
		return admin.FilterSemua, true
	}
	f := v.Get("status")
	if !admin.FilterSah(f) {
		tulisGalat(w, http.StatusBadRequest, kodeQueryTidakValid, "Parameter status harus semua, draft, posted, atau trash")
		return "", false
	}
	return f, true
}
