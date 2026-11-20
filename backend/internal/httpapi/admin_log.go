package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
)

// LayananLog adalah aturan baca log aktivitas (dipenuhi admin.LogLayanan).
type LayananLog interface {
	Daftar(ctx context.Context, f admin.FilterLog) (admin.HasilLog, error)
}

type logAdminHandler struct {
	log *slog.Logger
	svc LayananLog
}

type entriLogJSON struct {
	ID          int64     `json:"id"`
	TerjadiPada time.Time `json:"terjadiPada"`
	PenggunaSub string    `json:"penggunaSub"`
	Username    string    `json:"username"`
	Jenis       string    `json:"jenis"`
	Aksi        string    `json:"aksi"`
	Modul       string    `json:"modul"`
	ItemID      int64     `json:"itemId"`
	LabelItem   string    `json:"labelItem"`
}

type daftarLogJSON struct {
	Items     []entriLogJSON `json:"items"`
	Page      int            `json:"page"`
	Limit     int            `json:"limit"`
	Total     int64          `json:"total"`
	Usernames []string       `json:"usernames"`
}

// parseFilterLog membaca query: page, limit (default 10, maks 50), username, jenis, dari, sampai (YYYY-MM-DD).
func parseFilterLog(v url.Values) (admin.FilterLog, string) {
	f := admin.FilterLog{Username: strings.TrimSpace(v.Get("username")), Jenis: v.Get("jenis"), Dari: v.Get("dari"), Sampai: v.Get("sampai")}
	for nama, dst := range map[string]*int{"page": &f.Page, "limit": &f.Limit} {
		if !v.Has(nama) {
			continue
		}
		n, err := strconv.ParseInt(v.Get(nama), 10, 32)
		if err != nil || n < 1 {
			return f, "Parameter " + nama + " harus bilangan bulat minimal 1"
		}
		*dst = int(n)
	}
	if !utf8.ValidString(f.Username) || utf8.RuneCountInString(f.Username) > admin.BatasUsernameLog {
		return f, "Parameter username tidak valid"
	}
	return f, ""
}

// daftar melayani GET /api/admin/log-aktivitas. Baca-saja: tidak ada rute tulis untuk log.
func (h logAdminHandler) daftar(w http.ResponseWriter, r *http.Request) {
	f, pesan := parseFilterLog(r.URL.Query())
	if pesan != "" {
		tulisGalat(w, http.StatusBadRequest, kodeQueryTidakValid, pesan)
		return
	}
	hasil, err := h.svc.Daftar(r.Context(), f)
	if err != nil {
		petakanGalatAdmin(h.log, w, r, err)
		return
	}
	items := make([]entriLogJSON, 0, len(hasil.Items))
	for _, e := range hasil.Items {
		items = append(items, entriLogJSON{
			ID: e.ID, TerjadiPada: e.TerjadiPada, PenggunaSub: e.PenggunaSub, Username: e.Username, Jenis: e.Jenis,
			Aksi: e.Aksi, Modul: e.Modul, ItemID: e.ItemID, LabelItem: e.LabelItem,
		})
	}
	tulisJSON(w, http.StatusOK, daftarLogJSON{Items: items, Page: hasil.Page, Limit: hasil.Limit, Total: hasil.Total, Usernames: hasil.Usernames})
}
