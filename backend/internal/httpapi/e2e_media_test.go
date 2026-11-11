package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

// Tes ujung ke ujung gambar: login, upload, lalu ambil lewat /media dengan RustFS nyata (bucket tes).
func TestE2EUploadLaluMedia(t *testing.T) {
	k := testutil.RustFSTes(t)
	ctx, batal := context.WithTimeout(context.Background(), 60*time.Second)
	defer batal()
	s3 := media.NewS3(media.S3Config{Endpoint: k.Endpoint, AccessKey: k.AccessKey, SecretKey: k.SecretKey, Bucket: k.Bucket})
	if err := s3.PastikanBucket(ctx, true); err != nil {
		t.Fatal(err)
	}
	var dibuat []string
	t.Cleanup(func() {
		for _, key := range dibuat {
			_ = s3.Hapus(context.Background(), key)
		}
	})

	e := siapkanAuthDenganMedia(t, nil, s3)
	isi := pngUji(t)
	id, csrf := loginBerhasil(t, e)
	body, ct := bangunMultipart(t, map[string]string{"sumber": "dokumentasi"}, bagianFile{"gambar", "foto.png", "image/png", isi})
	rec := unggah(t, e, id, csrf, body, ct)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload status %d: %s", rec.Code, rec.Body)
	}
	var h hasilUpload
	if err := json.Unmarshal(rec.Body.Bytes(), &h); err != nil {
		t.Fatal(err)
	}
	dibuat = append(dibuat, h.Key)

	// objek benar-benar ada di RustFS dengan Content-Type dari isi file
	meta, err := s3.Head(ctx, h.Key)
	if err != nil || meta.ContentType != "image/png" || meta.Ukuran != int64(len(isi)) {
		t.Fatalf("objek di RustFS salah: %+v %v", meta, err)
	}

	// diambil publik (tanpa cookie) lewat /media
	get := ambilMedia(e, http.MethodGet, h.URL, nil)
	if get.Code != http.StatusOK || !bytes.Equal(get.Body.Bytes(), isi) || get.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("GET /media: status %d, sama=%v, tipe %q", get.Code, bytes.Equal(get.Body.Bytes(), isi), get.Header().Get("Content-Type"))
	}
	etag := get.Header().Get("ETag")
	if etag == "" {
		t.Fatal("ETag dari RustFS kosong")
	}
	if rec := ambilMedia(e, http.MethodGet, h.URL, map[string]string{"If-None-Match": etag}); rec.Code != http.StatusNotModified {
		t.Errorf("If-None-Match harus 304, dapat %d", rec.Code)
	}
	if rec := ambilMedia(e, http.MethodHead, h.URL, nil); rec.Code != http.StatusOK || rec.Body.Len() != 0 {
		t.Errorf("HEAD: status %d body %d", rec.Code, rec.Body.Len())
	}

	// setelah dihapus: 404
	if err := s3.Hapus(ctx, h.Key); err != nil {
		t.Fatal(err)
	}
	if rec := ambilMedia(e, http.MethodGet, h.URL, nil); rec.Code != http.StatusNotFound {
		t.Errorf("setelah dihapus harus 404, dapat %d", rec.Code)
	}
}
