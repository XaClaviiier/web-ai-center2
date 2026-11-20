package httpapi

import (
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/seo"
)

// DepsStatic adalah dependensi penyajian frontend statis (folder web/, hasil build frontend-final). Satu origin
// dengan API (docs/stack.md). nil = tidak dipasang (tes API murni).
type DepsStatic struct {
	Dir           string // STATIC_DIR
	DevLoginAktif bool   // /admin/masuk hanya dilayani saat true (APP_ENV=development)
	Halaman404    []byte // isi 404.html bila ada (nil = halaman 404 bawaan)
	NamaSitus     string
}

// rutaStatis memetakan URL bersih ke berkas di STATIC_DIR. Tabel eksplisit: tidak ada tebakan ekstensi, sehingga
// path acak (mis. /index.html, /berita.html) tidak melayani apa pun dan jatuh ke 404.
var rutaStatis = map[string]string{
	"/":                            "index.html",
	"/berita":                      "berita/index.html",
	"/dokumentasi":                 "dokumentasi/index.html",
	"/tim":                         "tim/index.html",
	"/tentang-kami":                "tentang-kami/index.html",
	"/layanan/pelatihan":           "layanan/pelatihan/index.html",
	"/layanan/inference-solutions": "layanan/inference-solutions/index.html",
	"/layanan/inference-solutions/jejak-alumni":            "layanan/inference-solutions/jejak-alumni/index.html",
	"/layanan/inference-solutions/perencanaan-implan-gigi": "layanan/inference-solutions/perencanaan-implan-gigi/index.html",
	"/layanan/inference-solutions/document-intelligence":   "layanan/inference-solutions/document-intelligence/index.html",
}

// Tipe yang tidak selalu ada di tabel MIME bawaan sistem (mis. Windows tanpa registri).
var tipeTambahan = map[string]string{
	".woff2": "font/woff2",
	".webp":  "image/webp",
	".svg":   "image/svg+xml",
	".js":    "text/javascript; charset=utf-8",
	".css":   "text/css; charset=utf-8",
	".html":  tipeHTML,
}

type staticHandler struct {
	d    DepsStatic
	root fs.FS
}

func pasangStatic(r chi.Router, d DepsStatic) {
	h := staticHandler{d: d, root: os.DirFS(d.Dir)}
	for rute, berkas := range rutaStatis {
		berkas := berkas
		layani := func(w http.ResponseWriter, req *http.Request) { h.layaniBerkas(w, req, berkas, "no-cache") }
		r.Get(rute, layani)
		r.Head(rute, layani)
	}
	// Aset: tanpa daftar direktori, path dengan ".." ditolak (fs.ValidPath).
	r.Get("/assets/*", h.aset("assets", "no-cache"))
	r.Head("/assets/*", h.aset("assets", "no-cache"))
	// Bundel ber-hash (nama berubah bila isi berubah): aman di-cache lama.
	r.Get("/_astro/*", h.aset("_astro", cacheBerHash))
	r.Head("/_astro/*", h.aset("_astro", cacheBerHash))
	r.Get("/admin/assets/*", h.aset("admin/assets", cacheBerHash))
	r.Head("/admin/assets/*", h.aset("admin/assets", cacheBerHash))

	// Admin (React, hash-route di sisi klien): hanya dua pintu masuk. Tidak di-cache (RP-05).
	admin := func(w http.ResponseWriter, req *http.Request) { h.layaniBerkas(w, req, "admin/index.html", "no-store") }
	for _, p := range []string{"/admin", "/admin/"} {
		r.Get(p, admin)
		r.Head(p, admin)
	}
	// Halaman masuk development: di luar development berperilaku seperti path tak dikenal (404).
	masuk := func(w http.ResponseWriter, req *http.Request) {
		if !h.d.DevLoginAktif {
			h.tulis404(w, req)
			return
		}
		h.layaniBerkas(w, req, "admin/masuk.html", "no-store")
	}
	r.Get("/admin/masuk", masuk)
	r.Head("/admin/masuk", masuk)
	// Aset halaman masuk development; ikut 404 di luar development agar halamannya tidak terlihat.
	for _, f := range []string{"masuk.css", "masuk.js"} {
		f := f
		berkas := func(w http.ResponseWriter, req *http.Request) {
			if !h.d.DevLoginAktif {
				h.tulis404(w, req)
				return
			}
			h.layaniBerkas(w, req, "admin/"+f, "no-store")
		}
		r.Get("/admin/"+f, berkas)
		r.Head("/admin/"+f, berkas)
	}
}

// tidakDitemukan melayani 404 HTML untuk path di luar /api dan JSON di bawah /api.
func (h staticHandler) tidakDitemukan(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
		tulisGalat(w, http.StatusNotFound, kodeTidakDitemukan, "Sumber tidak ditemukan")
		return
	}
	h.tulis404(w, r)
}

func (h staticHandler) tulis404(w http.ResponseWriter, _ *http.Request) {
	tulis404HTML(w, h.d.Halaman404, h.d.NamaSitus)
}

// tulis404HTML menulis halaman 404 HTML (status 404, noindex). isi nil = halaman bawaan.
func tulis404HTML(w http.ResponseWriter, isi []byte, namaSitus string) {
	if isi == nil {
		isi = seo.Halaman404Bawaan(namaSitus)
	}
	hd := w.Header()
	hd.Set("Content-Type", tipeHTML)
	hd.Set("X-Robots-Tag", "noindex")
	hd.Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(isi)
}

// cacheBerHash untuk berkas yang namanya memuat hash isi.
const cacheBerHash = "public, max-age=31536000, immutable"

// aset melayani berkas di bawah direktori (relatif STATIC_DIR) untuk pola "/.../*".
func (h staticHandler) aset(dir, cacheControl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sisa := chi.URLParam(r, "*")
		nama := path.Join(dir, sisa)
		if !fs.ValidPath(nama) || strings.Contains(sisa, "..") {
			h.tulis404(w, r)
			return
		}
		h.layaniBerkas(w, r, nama, cacheControl)
	}
}

// layaniBerkas melayani satu berkas di bawah STATIC_DIR dengan ETag lemah (ukuran + waktu ubah) dan
// Cache-Control sesuai argumen. Direktori dan berkas yang tidak ada dibalas 404 HTML.
func (h staticHandler) layaniBerkas(w http.ResponseWriter, r *http.Request, nama, cacheControl string) {
	if !fs.ValidPath(nama) {
		h.tulis404(w, r)
		return
	}
	f, err := h.root.Open(nama)
	if err != nil {
		h.tulis404(w, r)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		h.tulis404(w, r)
		return
	}
	rs, ok := f.(readSeeker)
	if !ok {
		h.tulis404(w, r)
		return
	}
	hd := w.Header()
	if t, ok := tipeTambahan[strings.ToLower(path.Ext(nama))]; ok {
		hd.Set("Content-Type", t)
	} else if t := mime.TypeByExtension(path.Ext(nama)); t != "" {
		hd.Set("Content-Type", t)
	}
	hd.Set("Cache-Control", cacheControl)
	hd.Set("ETag", fmt.Sprintf(`W/"%x-%x"`, info.Size(), info.ModTime().UnixNano()))
	http.ServeContent(w, r, nama, info.ModTime(), rs)
}

type readSeeker interface {
	Read(p []byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
}
