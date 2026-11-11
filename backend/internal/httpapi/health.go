package httpapi

import (
	"context"
	"net/http"
	"time"
)

const batasWaktuPing = 2 * time.Second

// health: 200 {"status":"ok","db":"ok"}; 503 bila database tidak terjangkau (stack.md §9.4).
// Kegagalan tidak di-log di sini: endpoint ini bisa di-poll berkala, sinyalnya cukup status 503.
func health(db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, batal := context.WithTimeout(r.Context(), batasWaktuPing)
		defer batal()
		if err := db.Ping(ctx); err != nil {
			tulisGalat(w, http.StatusServiceUnavailable, kodeDBTidakAda, "Database tidak terjangkau")
			return
		}
		tulisJSON(w, http.StatusOK, map[string]string{"status": "ok", "db": "ok"})
	}
}
