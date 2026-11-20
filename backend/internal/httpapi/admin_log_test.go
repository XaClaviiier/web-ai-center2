package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
)

type logPalsu struct {
	f     admin.FilterLog
	hasil admin.HasilLog
	err   error
	n     int
}

func (l *logPalsu) Daftar(_ context.Context, f admin.FilterLog) (admin.HasilLog, error) {
	l.f, l.n = f, l.n+1
	return l.hasil, l.err
}

func siapkanLog(t *testing.T) (lingkunganAuth, *logPalsu) {
	t.Helper()
	l := &logPalsu{hasil: admin.HasilLog{
		Items: []admin.EntriLog{{ID: 9, TerjadiPada: time.Date(2026, 11, 12, 3, 4, 5, 0, time.UTC), PenggunaSub: "dev-admin", Username: "admin.dev",
			Jenis: "ubah", Aksi: "post", Modul: "berita", ItemID: 7, LabelItem: "Judul Tujuh"}},
		Page: 1, Limit: 10, Total: 1, Usernames: []string{"admin.dev", "budi"},
	}}
	return siapkanAuth(t, nil, func(d *Deps) { d.Admin = DepsAdmin{Log: l} }), l
}

func TestAdminLogDaftarBentukDanFilter(t *testing.T) {
	e, l := siapkanLog(t)
	id, _ := loginBerhasil(t, e)
	rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/log-aktivitas?page=2&limit=5&username=%20admin.dev%20&jenis=ubah&dari=2026-11-01&sampai=2026-11-12", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("log admin harus no-store: %q", rec.Header().Get("Cache-Control"))
	}
	want := admin.FilterLog{Page: 2, Limit: 5, Username: "admin.dev", Jenis: "ubah", Dari: "2026-11-01", Sampai: "2026-11-12"}
	if l.f != want {
		t.Errorf("filter salah: %+v", l.f)
	}
	m := peta(t, rec)
	if m["page"].(float64) != 1 || m["limit"].(float64) != 10 || m["total"].(float64) != 1 {
		t.Errorf("meta: %v", m)
	}
	if u := m["usernames"].([]any); len(u) != 2 || u[1] != "budi" {
		t.Errorf("usernames: %v", u)
	}
	e0 := m["items"].([]any)[0].(map[string]any)
	for k, v := range map[string]any{"id": 9.0, "terjadiPada": "2026-11-12T03:04:05Z", "penggunaSub": "dev-admin", "username": "admin.dev",
		"jenis": "ubah", "aksi": "post", "modul": "berita", "itemId": 7.0, "labelItem": "Judul Tujuh"} {
		if e0[k] != v {
			t.Errorf("%s = %v, mau %v", k, e0[k], v)
		}
	}
	if len(e0) != 9 {
		t.Errorf("field berlebih/kurang: %v", e0)
	}
}

func TestAdminLogDefaultDanQueryTidakValid(t *testing.T) {
	e, l := siapkanLog(t)
	id, _ := loginBerhasil(t, e)
	if rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/log-aktivitas", ""); rec.Code != 200 || l.f != (admin.FilterLog{}) {
		t.Errorf("default: %d %+v", rec.Code, l.f)
	}
	l.n = 0
	for _, q := range []string{"page=0", "page=abc", "limit=-1", "limit=1.5", "username=" + strings.Repeat("a", 101)} {
		if rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/log-aktivitas?"+q, ""); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: %d, mau 400", q, rec.Code)
		}
	}
	if l.n != 0 {
		t.Error("query tidak valid tidak boleh sampai ke layanan")
	}
	// galat validasi dari layanan -> 400
	l.err = &admin.ErrInput{Pesan: "Parameter dari harus tanggal valid berformat YYYY-MM-DD"}
	if rec := kirimAdmin(t, e.h, id, "", "GET", "/api/admin/log-aktivitas?dari=besok", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("ErrInput: %d", rec.Code)
	}
}

func TestAdminLogWajibLoginDanBacaSaja(t *testing.T) {
	e, _ := siapkanLog(t)
	if rec := kirimAdmin(t, e.h, "", "", "GET", "/api/admin/log-aktivitas", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("tanpa login: %d", rec.Code)
	}
	id, csrf := loginBerhasil(t, e)
	for _, m := range []string{"POST", "PUT", "PATCH", "DELETE"} {
		for _, p := range []string{"/api/admin/log-aktivitas", "/api/admin/log-aktivitas/1"} {
			if rec := kirimAdmin(t, e.h, id, csrf, m, p, `{}`); rec.Code != http.StatusMethodNotAllowed && rec.Code != http.StatusNotFound {
				t.Errorf("%s %s harus tidak ada: %d", m, p, rec.Code)
			}
		}
	}
}

func TestAuthMeMemuatUsername(t *testing.T) {
	e, _ := siapkanLog(t)
	id, _ := loginBerhasil(t, e)
	m := peta(t, kirimAdmin(t, e.h, id, "", "GET", "/api/auth/me", ""))
	if m["username"] != "admin.dev" || m["nama"] != "Admin Dev" {
		t.Errorf("/me: %v", m)
	}
}
