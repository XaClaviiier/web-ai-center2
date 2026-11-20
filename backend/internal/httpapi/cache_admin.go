package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/cache"
)

const batasWaktuBatalCache = 3 * time.Second

// denganBatalCache membungkus router admin satu sumber agar cache publiknya dibatalkan saat konten publik berubah:
//   - berita, dokumentasi, testimoni: saat Post atau Retract berhasil (draft dan trash tidak pernah ada di cache);
//   - tim (semuaMutasi=true): saat tambah, edit, atau hapus berhasil, karena tim tidak punya status.
//
// Pembatalan dijalankan SEBELUM header respons terkirim (setelah perubahan tersimpan di database), jadi begitu
// admin menerima sukses, permintaan publik berikutnya sudah membaca data baru (Retract langsung 404).
func denganBatalCache(c *cache.Cache, jenis []string, semuaMutasi bool, log *slog.Logger, router func(chi.Router)) func(chi.Router) {
	return func(r chi.Router) {
		if c != nil {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					if req.Method == http.MethodGet || req.Method == http.MethodHead {
						next.ServeHTTP(w, req)
						return
					}
					pw := &penulisBatal{ResponseWriter: w, sebelum: func(status int) {
						if status < 200 || status > 299 || !(semuaMutasi || aksiPublikasi(req.URL.Path)) || strings.HasSuffix(req.URL.Path, "/lock") {
							return
						}
						ctx, batal := context.WithTimeout(context.WithoutCancel(req.Context()), batasWaktuBatalCache)
						defer batal()
						for _, j := range jenis {
							if err := c.Batalkan(ctx, j); err != nil {
								// Perubahan sudah tersimpan; data lama paling lama bertahan sampai TTL cache.
								log.Error("gagal membatalkan cache publik", "jenis", j, "err", err, "path", req.URL.Path, "request_id", middleware.GetReqID(req.Context()))
							}
						}
					}}
					next.ServeHTTP(pw, req)
				})
			})
		}
		router(r)
	}
}

func aksiPublikasi(path string) bool {
	return strings.HasSuffix(path, "/post") || strings.HasSuffix(path, "/retract")
}

// penulisBatal memanggil sebelum() tepat sekali, sebelum status dikirim ke klien.
type penulisBatal struct {
	http.ResponseWriter
	sebelum func(status int)
	sudah   bool
}

func (p *penulisBatal) WriteHeader(code int) {
	if !p.sudah {
		p.sudah = true
		p.sebelum(code)
	}
	p.ResponseWriter.WriteHeader(code)
}

func (p *penulisBatal) Write(b []byte) (int, error) {
	if !p.sudah {
		p.WriteHeader(http.StatusOK)
	}
	return p.ResponseWriter.Write(b)
}

func (p *penulisBatal) Unwrap() http.ResponseWriter { return p.ResponseWriter }
