package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/auth"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/ratelimit"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

func proxyUji() []netip.Prefix {
	return []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("fd00::/8")}
}

func TestIPDariXFF(t *testing.T) {
	cases := []struct {
		nama, koneksi string
		xff           []string
		mau           string
	}{
		{"koneksi bukan proxy: header diabaikan", "8.8.8.8", []string{"1.2.3.4"}, "8.8.8.8"},
		{"proxy tanpa header", "10.0.0.5", nil, "10.0.0.5"},
		{"satu klien", "10.0.0.5", []string{"1.2.3.4"}, "1.2.3.4"},
		{"entri kanan yang bukan proxy, bukan yang paling kiri (dipalsukan klien)", "10.0.0.5", []string{"6.6.6.6, 1.2.3.4, 10.0.0.9"}, "1.2.3.4"},
		{"beberapa header digabung berurutan", "10.0.0.5", []string{"6.6.6.6", "1.2.3.4, 10.0.0.9"}, "1.2.3.4"},
		{"semua proxy", "10.0.0.5", []string{"10.0.0.7, 10.0.0.9"}, "10.0.0.5"},
		{"rantai rusak tidak dipercaya", "10.0.0.5", []string{"1.2.3.4, bukan-ip"}, "10.0.0.5"},
		{"IPv6 dalam bentuk dipetakan IPv4", "10.0.0.5", []string{"::ffff:1.2.3.4"}, "1.2.3.4"},
		{"proxy IPv6", "fd00::1", []string{"2001:db8::7"}, "2001:db8::7"},
	}
	for _, c := range cases {
		if got := ipDariXFF(proxyUji(), c.koneksi, c.xff); got != c.mau {
			t.Errorf("%s: dapat %q, mau %q", c.nama, got, c.mau)
		}
	}
}

func TestLoginMemakaiIPKlienDariProxyTepercaya(t *testing.T) {
	e := siapkanAuth(t, nil, func(d *Deps) { d.ProxyTepercaya = proxyUji() })
	dari := func(remote, xff string) string {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/dev-login", nil)
		req.RemoteAddr = remote
		req.Header.Set("Origin", originSah)
		if xff != "" {
			req.Header.Set("X-Forwarded-For", xff)
		}
		e.h.ServeHTTP(httptest.NewRecorder(), req)
		return e.limiter.id
	}
	if got := dari("10.0.0.5:5000", "1.2.3.4, 10.0.0.9"); got != "1.2.3.4" {
		t.Errorf("lewat proxy tepercaya: %q", got)
	}
	if got := dari("8.8.8.8:5000", "1.2.3.4"); got != "8.8.8.8" {
		t.Errorf("header dari koneksi biasa harus diabaikan: %q", got)
	}
}

func TestTanpaProxyTepercayaHeaderDiabaikan(t *testing.T) {
	e := siapkanAuth(t, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/dev-login", nil)
	req.RemoteAddr = "10.0.0.5:5000"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.Header.Set("Origin", originSah)
	e.h.ServeHTTP(httptest.NewRecorder(), req)
	if e.limiter.id != "10.0.0.5" {
		t.Errorf("default aman: IP harus alamat koneksi, dapat %q", e.limiter.id)
	}
}

// rantaiAdmin memasang batasAdmin di belakang middleware yang menanam session dengan sub tertentu.
func rantaiAdmin(lim ratelimit.Limiter, baca, tulis int) http.Handler {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	h := batasAdmin(lim, baca, tulis, slog.New(slog.NewTextHandler(io.Discard, nil)))(ok)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sub := r.Header.Get("X-Sub")
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), kunciSession, auth.Session{Sub: sub})))
	})
}

func panggilSub(h http.Handler, sub, metode string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metode, "/x", nil)
	req.Header.Set("X-Sub", sub)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestBatasAdminBacaTulisTerpisahDanPerUser(t *testing.T) {
	rdb := testutil.RedisTes(t)
	h := rantaiAdmin(ratelimit.Redis{C: rdb, Jendela: time.Minute}, 3, 2)

	for i := 1; i <= 3; i++ {
		if rec := panggilSub(h, "admin-a", http.MethodGet); rec.Code != http.StatusNoContent {
			t.Fatalf("baca ke-%d: %d", i, rec.Code)
		}
	}
	rec := panggilSub(h, "admin-a", http.MethodGet)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("baca ke-4 harus 429, dapat %d", rec.Code)
	}
	if ra, err := strconv.Atoi(rec.Header().Get("Retry-After")); err != nil || ra < 1 || ra > 60 {
		t.Errorf("Retry-After = %q", rec.Header().Get("Retry-After"))
	}
	// Bucket tulis terpisah dari baca; tulis dibatasi 2.
	for i, mau := range []int{204, 204, 429} {
		if rec := panggilSub(h, "admin-a", http.MethodPost); rec.Code != mau {
			t.Fatalf("tulis ke-%d: %d, mau %d", i+1, rec.Code, mau)
		}
	}
	for _, m := range []string{http.MethodPut, http.MethodDelete} {
		if rec := panggilSub(h, "admin-a", m); rec.Code != http.StatusTooManyRequests {
			t.Errorf("%s harus masuk bucket tulis (429), dapat %d", m, rec.Code)
		}
	}
	// Admin lain (walau dari IP yang sama) tidak terpengaruh.
	if rec := panggilSub(h, "admin-b", http.MethodGet); rec.Code != http.StatusNoContent {
		t.Errorf("admin lain tidak boleh ikut terbatasi: %d", rec.Code)
	}
	if rec := panggilSub(h, "admin-b", http.MethodPost); rec.Code != http.StatusNoContent {
		t.Errorf("admin lain tidak boleh ikut terbatasi (tulis): %d", rec.Code)
	}
}

func TestBatasAdminGagalTertutupBilaRedisBermasalah(t *testing.T) {
	h := rantaiAdmin(&limiterPalsu{err: errors.New("redis mati")}, 10, 10)
	rec := panggilSub(h, "a", http.MethodGet)
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "session_unavailable") {
		t.Errorf("harus 503 session_unavailable, dapat %d %s", rec.Code, rec.Body)
	}
}

func TestBatasAdminNolMematikan(t *testing.T) {
	h := rantaiAdmin(&limiterPalsu{err: errors.New("tidak boleh dipanggil")}, 0, 0)
	if rec := panggilSub(h, "a", http.MethodPost); rec.Code != http.StatusNoContent {
		t.Errorf("batas 0 = mati, dapat %d", rec.Code)
	}
}

// Di router asli: limiter terpasang setelah login (tanpa login 401, tidak dihitung) dan sebelum CSRF
// (request yang ditolak CSRF tetap dihitung).
func TestBatasAdminDiRouterSetelahLoginSebelumCSRF(t *testing.T) {
	rdb := testutil.RedisTes(t)
	lim := ratelimit.Redis{C: rdb, Jendela: time.Minute}
	e := siapkanAuth(t, func(d *DepsAuth) { d.Limiter = lim }, func(d *Deps) {
		d.BatasAdminBaca, d.BatasAdminTulis = 5, 2
	})
	id, csrf := loginBerhasil(t, e)

	if rec := kirim(e.h, "GET", "/api/admin/berita", "", opsi{}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("tanpa login harus 401, dapat %d", rec.Code)
	}
	if n, _ := rdb.Get(context.Background(), "ratelimit:admin-read:dev-admin").Int(); n != 0 {
		t.Errorf("request tanpa login tidak boleh dihitung, hitungan %d", n)
	}
	buruk := opsi{origin: originSah, csrf: "salah", cookie: id}
	for i := 0; i < 2; i++ {
		if rec := kirim(e.h, "DELETE", "/api/admin/berita/1", "", buruk); rec.Code != http.StatusForbidden {
			t.Fatalf("CSRF salah harus 403, dapat %d", rec.Code)
		}
	}
	if rec := kirim(e.h, "DELETE", "/api/admin/berita/1", "", opsi{origin: originSah, csrf: csrf, cookie: id}); rec.Code != http.StatusTooManyRequests {
		t.Errorf("tulis ke-3 harus 429 (dua penolakan CSRF ikut dihitung), dapat %d", rec.Code)
	}
}
