package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Kode galat stabil untuk klien (stack.md §9.3). Pesan dalam sentence case, tanpa detail internal.
const (
	kodeTidakDitemukan = "not_found"
	kodeMetodeSalah    = "method_not_allowed"
	kodeInternal       = "internal_error"
	kodeDBTidakAda     = "db_unavailable"
)

type bentukGalat struct {
	Error isiGalat `json:"error"`
}

type isiGalat struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func tulisJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("gagal menulis respons JSON", "err", err)
	}
}

func tulisGalat(w http.ResponseWriter, status int, kode, pesan string) {
	tulisJSON(w, status, bentukGalat{Error: isiGalat{Code: kode, Message: pesan}})
}
