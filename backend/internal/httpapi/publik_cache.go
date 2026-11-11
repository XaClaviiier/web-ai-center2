package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/cache"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
)

const (
	bucketPublikMiss = "public-miss"
	// Pemuatan bersama (singleflight) tidak boleh ikut batal bila pemanggil pertama memutus koneksi.
	batasWaktuMuat = 10 * time.Second
	// Boleh disimpan di cache bersama dan browser, tetapi WAJIB divalidasi tiap kali (ETag/If-None-Match),
	// supaya Retract langsung terlihat tanpa jendela basi (docs/stack.md §6).
	cacheControlPublik = "public, no-cache"
)

// muatFn memuat dan membentuk payload JSON publik. Hasil kedua menyatakan apakah payload boleh disimpan di cache
// (mis. halaman daftar kosong di luar rentang tidak disimpan supaya key tak terbatas tidak memenuhi Redis).
type muatFn func(ctx context.Context) (payload any, simpan bool, err error)

// subDaftar adalah bagian key cache untuk daftar; kosong (tidak di-cache) bila ada pencarian q, karena teks
// bebas membuat jumlah key tak terbatas.
func subDaftar(q daftarQuery) string {
	if q.Pola != "" {
		return ""
	}
	return fmt.Sprintf("list:p%d:l%d", q.Page, q.Limit)
}

func bolehSimpanDaftar(q daftarQuery, jumlah int) bool {
	return jumlah > 0 || q.Page == 1
}

// layani melayani GET publik dengan alur: cache hit -> (304 atau 200) tanpa menyentuh limiter; miss -> limiter per IP
// -> muat dari database (satu pemuat per key) -> simpan -> kirim. sub kosong = tidak di-cache.
// Redis yang bermasalah tidak pernah menjatuhkan situs publik: cache dan limiter gagal terbuka (dicatat di log).
func (h publik) layani(w http.ResponseWriter, r *http.Request, jenis, sub string, muat muatFn) {
	ctx := r.Context()
	var kunci cache.Kunci
	pakaiCache := h.cache != nil && sub != ""
	if pakaiCache {
		k, err := h.cache.Kunci(ctx, jenis, sub)
		if err == nil {
			kunci = k
			body, ada, err := h.cache.Ambil(ctx, k)
			if err == nil && ada {
				kirimJSONPublik(w, r, body)
				return
			}
			if err != nil {
				h.catatRedis("baca cache", err, r)
				pakaiCache = false
			}
		} else {
			h.catatRedis("baca versi cache", err, r)
			pakaiCache = false
		}
	}

	if h.limiter != nil && h.batas > 0 {
		kep, err := h.limiter.Hit(ctx, bucketPublikMiss, ipKlien(r), h.batas)
		if err != nil {
			h.catatRedis("limiter publik", err, r)
		} else if !kep.Diizinkan {
			w.Header().Set("Retry-After", strconv.Itoa(kep.TungguDetik))
			tulisGalat(w, http.StatusTooManyRequests, kodeTerlaluSering, "Terlalu banyak permintaan, coba lagi nanti")
			return
		}
	}

	// Bila dipakai bersama (singleflight), konteks dilepas dari pembatalan satu request.
	muatCtx := ctx
	if pakaiCache {
		c, batal := context.WithTimeout(context.WithoutCancel(ctx), batasWaktuMuat)
		defer batal()
		muatCtx = c
	}
	jalankan := func() ([]byte, error) {
		payload, simpan, err := muat(muatCtx)
		if err != nil {
			return nil, err
		}
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encode respons: %w", err)
		}
		b = append(b, '\n')
		if pakaiCache && simpan {
			if err := h.cache.Simpan(muatCtx, kunci, b); err != nil {
				h.catatRedis("tulis cache", err, r)
			}
		}
		return b, nil
	}
	var (
		body []byte
		err  error
	)
	if pakaiCache {
		body, err = h.cache.Satukan(kunci, jalankan)
	} else {
		body, err = jalankan()
	}
	if err != nil {
		if errors.Is(err, konten.ErrTidakDitemukan) {
			tulisTidakDitemukan(w)
			return
		}
		h.galatInternal(w, r, err)
		return
	}
	kirimJSONPublik(w, r, body)
}

func (h publik) catatRedis(apa string, err error, r *http.Request) {
	h.log.Error("redis bermasalah, dilewati", "op", apa, "err", err, "path", r.URL.Path, "request_id", middleware.GetReqID(r.Context()))
}

// kirimJSONPublik membalas 200 dengan ETag kuat (hash isi), atau 304 tanpa body bila If-None-Match cocok.
func kirimJSONPublik(w http.ResponseWriter, r *http.Request, body []byte) {
	sum := sha256.Sum256(body)
	etag := `"` + hex.EncodeToString(sum[:16]) + `"`
	h := w.Header()
	h.Set("Cache-Control", cacheControlPublik)
	h.Set("ETag", etag)
	if cocokETag(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	h.Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
