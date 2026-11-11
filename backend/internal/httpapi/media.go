package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
)

const (
	kodeTerlaluBesar      = "payload_too_large"
	kodeTipeTidakDidukung = "unsupported_media_type"
	kodeStorageBermasalah = "storage_unavailable"

	// Ruang untuk batas multipart (boundary, nama field) di luar isi file.
	overheadMultipart = 64 << 10
	waktuStorage      = 20 * time.Second
)

type mediaHandler struct {
	log      *slog.Logger
	store    media.Store
	maxBytes int64
}

type hasilUpload struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

// upload menerima SATU gambar (multipart: field "sumber" dan file "gambar"), memeriksa tipe dari isinya,
// menyimpannya ke RustFS dengan key acak, lalu mengembalikan {key,url}. Key disimpan klien saat Simpan.
func (h mediaHandler) upload(w http.ResponseWriter, r *http.Request) {
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "multipart/form-data" {
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "Gunakan multipart/form-data dengan field sumber dan gambar")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes+overheadMultipart)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var terlalu *http.MaxBytesError
		if errors.As(err, &terlalu) {
			h.tolakTerlaluBesar(w)
			return
		}
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "Body multipart tidak valid")
		return
	}
	defer r.MultipartForm.RemoveAll() //nolint:errcheck // pembersihan berkas sementara

	sumber := ""
	if v := r.MultipartForm.Value["sumber"]; len(v) == 1 {
		sumber = v[0]
	}
	if !media.SumberSah(sumber) {
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "Field sumber harus berita, dokumentasi, testimoni, atau tim")
		return
	}
	files := r.MultipartForm.File["gambar"]
	if len(files) != 1 {
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "Kirim tepat satu file pada field gambar")
		return
	}
	fh := files[0]
	if fh.Size > h.maxBytes {
		h.tolakTerlaluBesar(w)
		return
	}
	f, err := fh.Open()
	if err != nil {
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "File tidak dapat dibaca")
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, h.maxBytes+1))
	if err != nil {
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "File tidak dapat dibaca")
		return
	}
	if int64(len(data)) > h.maxBytes {
		h.tolakTerlaluBesar(w)
		return
	}
	if len(data) == 0 {
		tulisGalat(w, http.StatusBadRequest, kodeRequestTidakSah, "File kosong")
		return
	}
	// Tipe ditentukan dari isi; nama file dan Content-Type dari klien tidak dipercaya dan tidak dipakai.
	jenis, ok := media.DeteksiGambar(data)
	if !ok {
		tulisGalat(w, http.StatusUnsupportedMediaType, kodeTipeTidakDidukung, "Tipe file harus JPEG, PNG, atau WebP")
		return
	}
	key, err := media.BuatKey(sumber, jenis.Ext)
	if err != nil {
		h.galatStorage(w, r, err)
		return
	}
	ctx, batal := context.WithTimeout(r.Context(), waktuStorage)
	defer batal()
	if err := h.store.Put(ctx, key, data, jenis.ContentType); err != nil {
		h.galatStorage(w, r, err)
		return
	}
	tulisJSON(w, http.StatusCreated, hasilUpload{Key: key, URL: urlMedia(key)})
}

func (h mediaHandler) tolakTerlaluBesar(w http.ResponseWriter) {
	mb := float64(h.maxBytes) / (1 << 20)
	tulisGalat(w, http.StatusRequestEntityTooLarge, kodeTerlaluBesar,
		"Ukuran file melebihi batas "+strconv.FormatFloat(mb, 'f', -1, 64)+" MB")
}

// tipeAman adalah satu-satunya Content-Type yang boleh dikirim /media (pertahanan berlapis bila objek asing masuk bucket).
var tipeAman = map[string]struct{}{"image/jpeg": {}, "image/png": {}, "image/webp": {}}

// sajikan melayani GET/HEAD /media/{key} tanpa login (key acak tak tertebak). Key harus berbentuk buatan kita,
// sehingga path traversal atau objek asing tidak bisa diminta. Rentang (Range) tidak didukung: selalu respons penuh.
func (h mediaHandler) sajikan(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "*")
	if _, ok := media.SumberDariKey(key); !ok {
		tulisTidakDitemukan(w)
		return
	}
	ctx, batal := context.WithTimeout(r.Context(), waktuStorage)
	defer batal()

	var (
		meta media.Meta
		body io.ReadCloser
		err  error
	)
	if r.Method == http.MethodHead {
		meta, err = h.store.Head(ctx, key)
	} else {
		var o media.Objek
		o, err = h.store.Get(ctx, key)
		meta, body = o.Meta, o.Body
	}
	if err != nil {
		if errors.Is(err, media.ErrTidakAda) {
			tulisTidakDitemukan(w)
			return
		}
		h.galatStorage(w, r, err)
		return
	}
	if body != nil {
		defer body.Close()
	}

	hd := w.Header()
	ctype := meta.ContentType
	if _, ok := tipeAman[ctype]; !ok {
		ctype = "application/octet-stream"
	}
	hd.Set("Content-Type", ctype)
	hd.Set("X-Content-Type-Options", "nosniff")
	hd.Set("Content-Security-Policy", "default-src 'none'; sandbox")
	hd.Set("Cache-Control", "public, max-age=31536000, immutable")
	if meta.ETag != "" {
		hd.Set("ETag", meta.ETag)
		if cocokETag(r.Header.Get("If-None-Match"), meta.ETag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	hd.Set("Content-Length", strconv.FormatInt(meta.Ukuran, 10))
	w.WriteHeader(http.StatusOK)
	if body != nil {
		_, _ = io.Copy(w, body) // galat di tengah pengiriman hanya berarti klien memutus koneksi
	}
}

// cocokETag memeriksa header If-None-Match (daftar ETag atau *), mengabaikan penanda weak "W/".
func cocokETag(header, etag string) bool {
	if header == "" {
		return false
	}
	for _, t := range strings.Split(header, ",") {
		t = strings.TrimPrefix(strings.TrimSpace(t), "W/")
		if t == "*" || t == etag {
			return true
		}
	}
	return false
}

func (h mediaHandler) galatStorage(w http.ResponseWriter, r *http.Request, err error) {
	h.log.Error("storage gambar bermasalah", "err", err, "path", r.URL.Path, "request_id", middleware.GetReqID(r.Context()))
	tulisGalat(w, http.StatusServiceUnavailable, kodeStorageBermasalah, "Penyimpanan gambar sedang tidak tersedia")
}
