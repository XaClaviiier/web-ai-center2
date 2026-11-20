package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
)

// Batas body JSON admin: cukup untuk isi 20.000 karakter (hingga 4 byte/karakter) plus field lain.
const batasBodyJSON = 256 << 10

// bacaJSON membaca body JSON ketat ke dst: Content-Type application/json, ukuran dibatasi, field tak dikenal ditolak,
// dan tidak boleh ada data setelah objek. Bila gagal, respons galat sudah ditulis dan false dikembalikan.
func bacaJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "application/json" {
		tulisGalat(w, http.StatusUnsupportedMediaType, kodeTipeTidakDidukung, "Content-Type harus application/json")
		return false
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, batasBodyJSON))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var terlalu *http.MaxBytesError
		if errors.As(err, &terlalu) {
			tulisGalat(w, http.StatusRequestEntityTooLarge, kodeTerlaluBesar, "Body terlalu besar")
			return false
		}
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "Body harus berupa JSON yang valid dengan field yang dikenal")
		return false
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "Body memuat data setelah objek JSON")
		return false
	}
	return true
}
