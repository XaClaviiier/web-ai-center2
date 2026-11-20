package httpapi

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/ratelimit"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
)

const (
	bucketAdminBaca  = "admin-read"
	bucketAdminTulis = "admin-write"

	kunciIPKlien ctxKey = iota + 100
)

// ipKlien mengembalikan IP klien: hasil middleware ipKlienTepercaya bila terpasang, selain itu alamat koneksi.
func ipKlien(r *http.Request) string {
	if ip, ok := r.Context().Value(kunciIPKlien).(string); ok && ip != "" {
		return ip
	}
	return ipKoneksi(r)
}

func ipKoneksi(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ipKlienTepercaya menentukan IP klien sekali per request (RP-03). X-Forwarded-For hanya dipercaya bila alamat
// koneksi berada di daftar proxy tepercaya; selama daftar kosong, header diabaikan (klien bisa memalsukannya).
// Alamat yang dipakai adalah entri paling kanan yang bukan proxy tepercaya.
func ipKlienTepercaya(proxy []netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if len(proxy) == 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ipDariXFF(proxy, ipKoneksi(r), r.Header.Values("X-Forwarded-For"))
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), kunciIPKlien, ip)))
		})
	}
}

func tepercaya(proxy []netip.Prefix, a netip.Addr) bool {
	for _, p := range proxy {
		if p.Contains(a) {
			return true
		}
	}
	return false
}

func ipDariXFF(proxy []netip.Prefix, koneksi string, xff []string) string {
	sambung, err := netip.ParseAddr(koneksi)
	if err != nil || !tepercaya(proxy, sambung.Unmap()) {
		return koneksi
	}
	var rantai []string
	for _, h := range xff {
		for _, bagian := range strings.Split(h, ",") {
			if b := strings.TrimSpace(bagian); b != "" {
				rantai = append(rantai, b)
			}
		}
	}
	for i := len(rantai) - 1; i >= 0; i-- {
		a, err := netip.ParseAddr(rantai[i])
		if err != nil {
			return koneksi // rantai rusak: jangan percaya apa pun dari header
		}
		a = a.Unmap()
		if !tepercaya(proxy, a) {
			return a.String()
		}
	}
	return koneksi
}

// batasAdmin membatasi request admin per user ID (sub dari session), jadi dua admin dari IP yang sama tidak saling
// memengaruhi. Dipasang SETELAH wajibLogin; setiap request yang sampai ke sini dihitung, termasuk yang nanti ditolak CSRF.
// Gagal tertutup: bila Redis bermasalah, 503 (admin memang butuh Redis untuk session). batas <= 0 = mati.
func batasAdmin(lim ratelimit.Limiter, baca, tulis int, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if lim == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bucket, batas := bucketAdminTulis, tulis
			if metodeAman(r.Method) {
				bucket, batas = bucketAdminBaca, baca
			}
			if batas > 0 {
				sess, _ := SessionDariContext(r.Context())
				kep, err := lim.Hit(r.Context(), bucket, sess.Sub, batas)
				if err != nil {
					log.Error("limiter admin bermasalah", "err", err, "path", r.URL.Path, "request_id", middleware.GetReqID(r.Context()))
					tulisGalat(w, http.StatusServiceUnavailable, kodeSessionBermasalah, "Layanan session sedang tidak tersedia")
					return
				}
				if !kep.Diizinkan {
					w.Header().Set("Retry-After", strconv.Itoa(kep.TungguDetik))
					tulisGalat(w, http.StatusTooManyRequests, kodeTerlaluSering, "Terlalu banyak permintaan, coba lagi nanti")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
