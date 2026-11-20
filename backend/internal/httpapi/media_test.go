package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strconv"
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
)

func pngUji(t testing.TB) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(1, 1, color.RGBA{G: 200, A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

type bagianFile struct {
	field, nama, tipe string
	isi               []byte
}

// bangunMultipart membuat body multipart dengan field teks dan file; mengembalikan body dan Content-Type.
func bangunMultipart(t testing.TB, fields map[string]string, files ...bagianFile) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range files {
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", `form-data; name="`+f.field+`"; filename="`+f.nama+`"`)
		h.Set("Content-Type", f.tipe)
		w, err := mw.CreatePart(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(f.isi); err != nil {
			t.Fatal(err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, mw.FormDataContentType()
}

// unggah mengirim POST /api/admin/upload sebagai pengguna yang sudah login.
func unggah(t testing.TB, e lingkunganAuth, id, csrf string, body *bytes.Buffer, ctype string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/upload", body)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("Origin", originSah)
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}
	if id != "" {
		req.AddCookie(&http.Cookie{Name: "session", Value: id})
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	return rec
}

func uploadBerhasil(t *testing.T, e lingkunganAuth, sumber string, isi []byte) hasilUpload {
	t.Helper()
	id, csrf := loginBerhasil(t, e)
	body, ct := bangunMultipart(t, map[string]string{"sumber": sumber}, bagianFile{"gambar", "foto.png", "image/png", isi})
	rec := unggah(t, e, id, csrf, body, ct)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload status %d: %s", rec.Code, rec.Body)
	}
	var h hasilUpload
	if err := json.Unmarshal(rec.Body.Bytes(), &h); err != nil {
		t.Fatal(err)
	}
	return h
}

func TestUploadBerhasil(t *testing.T) {
	e := siapkanAuth(t, nil)
	isi := pngUji(t)
	h := uploadBerhasil(t, e, "berita", isi)

	if _, ok := media.SumberDariKey(h.Key); !ok || !strings.HasPrefix(h.Key, "berita/") || !strings.HasSuffix(h.Key, ".png") {
		t.Fatalf("key salah: %q", h.Key)
	}
	if h.URL != "/media/"+h.Key {
		t.Errorf("url salah: %q", h.URL)
	}
	meta, err := e.media.Head(context.Background(), h.Key)
	if err != nil || meta.ContentType != "image/png" || meta.Ukuran != int64(len(isi)) {
		t.Errorf("objek di storage salah: %+v %v", meta, err)
	}
}

func TestUploadSemuaSumberDanNamaFileTidakDipercaya(t *testing.T) {
	e := siapkanAuth(t, nil)
	id, csrf := loginBerhasil(t, e)
	for _, sumber := range []string{"berita", "dokumentasi", "testimoni", "tim"} {
		// nama file berbahaya dan Content-Type bohong tidak boleh memengaruhi key/ekstensi
		body, ct := bangunMultipart(t, map[string]string{"sumber": sumber},
			bagianFile{"gambar", "../../etc/passwd.php", "text/html", pngUji(t)})
		rec := unggah(t, e, id, csrf, body, ct)
		if rec.Code != http.StatusCreated {
			t.Fatalf("%s: status %d: %s", sumber, rec.Code, rec.Body)
		}
		var h hasilUpload
		_ = json.Unmarshal(rec.Body.Bytes(), &h)
		if s, ok := media.SumberDariKey(h.Key); !ok || s != sumber || !strings.HasSuffix(h.Key, ".png") {
			t.Errorf("%s: key %q", sumber, h.Key)
		}
		if strings.Contains(h.Key, "passwd") || strings.Contains(h.Key, "..") {
			t.Errorf("nama file klien bocor ke key: %q", h.Key)
		}
	}
	if e.media.Jumlah() != 4 {
		t.Errorf("harus 4 objek, dapat %d", e.media.Jumlah())
	}
}

func TestUploadTipeDariIsiBukanDariKlien(t *testing.T) {
	e := siapkanAuth(t, nil)
	id, csrf := loginBerhasil(t, e)
	tolak := map[string][]byte{
		"html bernama jpg": []byte("<html><script>alert(1)</script></html>"),
		"svg bernama png":  []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>1</script></svg>`),
		"gif":              []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00;"),
		"pdf":              []byte("%PDF-1.7 isi"),
		"teks":             []byte("bukan gambar"),
		"php":              []byte("<?php echo 1; ?>"),
	}
	for nama, isi := range tolak {
		body, ct := bangunMultipart(t, map[string]string{"sumber": "berita"}, bagianFile{"gambar", "foto.jpg", "image/jpeg", isi})
		rec := unggah(t, e, id, csrf, body, ct)
		if rec.Code != http.StatusUnsupportedMediaType || kodeGalat(t, rec) != kodeTipeTidakDidukung {
			t.Errorf("%s: status %d: %s", nama, rec.Code, rec.Body)
		}
	}
	if e.media.Jumlah() != 0 {
		t.Error("file ditolak tidak boleh tersimpan")
	}
}

func TestUploadUkuranMelebihiBatas(t *testing.T) {
	e := siapkanAuth(t, nil)
	id, csrf := loginBerhasil(t, e)

	// Sedikit di atas batas: ditolak lewat ukuran file.
	besar := append(pngUji(t), bytes.Repeat([]byte{0}, batasUnggahUji)...)
	body, ct := bangunMultipart(t, map[string]string{"sumber": "berita"}, bagianFile{"gambar", "a.png", "image/png", besar})
	rec := unggah(t, e, id, csrf, body, ct)
	if rec.Code != http.StatusRequestEntityTooLarge || kodeGalat(t, rec) != kodeTerlaluBesar {
		t.Fatalf("sedikit di atas batas: status %d: %s", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), "1 MB") {
		t.Errorf("pesan harus menyebut batas dalam MB: %s", rec.Body)
	}

	// Jauh di atas batas: ditolak oleh pembatas body sebelum sempat dibaca semuanya.
	raksasa := append(pngUji(t), bytes.Repeat([]byte{0}, 5*batasUnggahUji)...)
	body, ct = bangunMultipart(t, map[string]string{"sumber": "berita"}, bagianFile{"gambar", "a.png", "image/png", raksasa})
	rec = unggah(t, e, id, csrf, body, ct)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("jauh di atas batas: status %d: %s", rec.Code, rec.Body)
	}
	if e.media.Jumlah() != 0 {
		t.Error("file terlalu besar tidak boleh tersimpan")
	}

	// Tepat di batas diterima.
	pas := append(pngUji(t), bytes.Repeat([]byte{0}, batasUnggahUji-len(pngUji(t)))...)
	body, ct = bangunMultipart(t, map[string]string{"sumber": "berita"}, bagianFile{"gambar", "a.png", "image/png", pas})
	if rec := unggah(t, e, id, csrf, body, ct); rec.Code != http.StatusCreated {
		t.Errorf("tepat di batas harus diterima: %d %s", rec.Code, rec.Body)
	}
}

func TestUploadRequestTidakValid(t *testing.T) {
	e := siapkanAuth(t, nil)
	id, csrf := loginBerhasil(t, e)
	png := pngUji(t)

	kasus := map[string]struct {
		fields map[string]string
		files  []bagianFile
	}{
		"tanpa sumber":       {nil, []bagianFile{{"gambar", "a.png", "image/png", png}}},
		"sumber asing":       {map[string]string{"sumber": "kursus"}, []bagianFile{{"gambar", "a.png", "image/png", png}}},
		"sumber dengan path": {map[string]string{"sumber": "../tim"}, []bagianFile{{"gambar", "a.png", "image/png", png}}},
		"tanpa file":         {map[string]string{"sumber": "berita"}, nil},
		"field file salah":   {map[string]string{"sumber": "berita"}, []bagianFile{{"file", "a.png", "image/png", png}}},
		"dua file":           {map[string]string{"sumber": "berita"}, []bagianFile{{"gambar", "a.png", "image/png", png}, {"gambar", "b.png", "image/png", png}}},
		"file kosong":        {map[string]string{"sumber": "berita"}, []bagianFile{{"gambar", "a.png", "image/png", nil}}},
	}
	for nama, k := range kasus {
		body, ct := bangunMultipart(t, k.fields, k.files...)
		rec := unggah(t, e, id, csrf, body, ct)
		if rec.Code != http.StatusBadRequest || kodeGalat(t, rec) != kodeRequestTidakSah {
			t.Errorf("%s: status %d: %s", nama, rec.Code, rec.Body)
		}
	}

	// Bukan multipart.
	rec := unggah(t, e, id, csrf, bytes.NewBufferString(`{"sumber":"berita"}`), "application/json")
	if rec.Code != http.StatusBadRequest || kodeGalat(t, rec) != kodeRequestTidakSah {
		t.Errorf("JSON: status %d: %s", rec.Code, rec.Body)
	}
	// Multipart rusak.
	rec = unggah(t, e, id, csrf, bytes.NewBufferString("--x\r\nrusak"), "multipart/form-data; boundary=x")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("multipart rusak: status %d: %s", rec.Code, rec.Body)
	}
	if e.media.Jumlah() != 0 {
		t.Error("request tidak valid tidak boleh menyimpan apa pun")
	}
}

func TestUploadButuhLoginCSRFDanOrigin(t *testing.T) {
	e := siapkanAuth(t, nil)
	id, csrf := loginBerhasil(t, e)
	buat := func() (*bytes.Buffer, string) {
		return bangunMultipart(t, map[string]string{"sumber": "berita"}, bagianFile{"gambar", "a.png", "image/png", pngUji(t)})
	}

	body, ct := buat()
	if rec := unggah(t, e, "", "", body, ct); rec.Code != http.StatusUnauthorized {
		t.Errorf("tanpa login: %d", rec.Code)
	}
	body, ct = buat()
	if rec := unggah(t, e, id, "", body, ct); rec.Code != http.StatusForbidden || kodeGalat(t, rec) != kodeCSRFTidakValid {
		t.Errorf("tanpa CSRF: %d %s", rec.Code, rec.Body)
	}
	body, ct = buat()
	if rec := unggah(t, e, id, "salah", body, ct); rec.Code != http.StatusForbidden {
		t.Errorf("CSRF salah: %d", rec.Code)
	}
	body, ct = buat()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/upload", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-CSRF-Token", csrf)
	req.AddCookie(&http.Cookie{Name: "session", Value: id})
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req) // tanpa Origin
	if rec.Code != http.StatusForbidden || kodeGalat(t, rec) != kodeOriginTidakSah {
		t.Errorf("tanpa Origin: %d %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Error("respons admin harus no-store")
	}
	if e.media.Jumlah() != 0 {
		t.Error("request yang ditolak tidak boleh menyimpan apa pun")
	}
}

func TestUploadStorageBermasalah503(t *testing.T) {
	e := siapkanAuth(t, nil)
	id, csrf := loginBerhasil(t, e)
	e.media.Err = errors.New("rustfs: connection refused access_key=RAHASIA")
	body, ct := bangunMultipart(t, map[string]string{"sumber": "berita"}, bagianFile{"gambar", "a.png", "image/png", pngUji(t)})
	rec := unggah(t, e, id, csrf, body, ct)
	if rec.Code != http.StatusServiceUnavailable || kodeGalat(t, rec) != kodeStorageBermasalah {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "RAHASIA") {
		t.Error("detail internal bocor")
	}
}

// — GET/HEAD /media/{key} ————————————————————————————————————————————————

func ambilMedia(e lingkunganAuth, metode, path string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metode, path, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	return rec
}

func TestSajikanMedia(t *testing.T) {
	e := siapkanAuth(t, nil)
	isi := pngUji(t)
	h := uploadBerhasil(t, e, "tim", isi)

	rec := ambilMedia(e, http.MethodGet, h.URL, nil) // tanpa cookie: publik
	if rec.Code != http.StatusOK || !bytes.Equal(rec.Body.Bytes(), isi) {
		t.Fatalf("GET: status %d, panjang %d", rec.Code, rec.Body.Len())
	}
	want := map[string]string{
		"Content-Type":            "image/png",
		"X-Content-Type-Options":  "nosniff",
		"Content-Security-Policy": "default-src 'none'; sandbox",
		"Cache-Control":           "public, max-age=31536000, immutable",
	}
	for k, v := range want {
		if rec.Header().Get(k) != v {
			t.Errorf("header %s = %q, diharapkan %q", k, rec.Header().Get(k), v)
		}
	}
	if rec.Header().Get("ETag") == "" || rec.Header().Get("Content-Length") == "" {
		t.Errorf("ETag/Content-Length wajib ada: %v", rec.Header())
	}
}

func TestSajikanMediaHEAD(t *testing.T) {
	e := siapkanAuth(t, nil)
	isi := pngUji(t)
	h := uploadBerhasil(t, e, "berita", isi)

	rec := ambilMedia(e, http.MethodHead, h.URL, nil)
	if rec.Code != http.StatusOK || rec.Body.Len() != 0 {
		t.Fatalf("HEAD: status %d, body %d byte", rec.Code, rec.Body.Len())
	}
	if rec.Header().Get("Content-Length") != strconv.Itoa(len(isi)) || rec.Header().Get("Content-Type") != "image/png" {
		t.Errorf("header HEAD salah: %v", rec.Header())
	}
}

func TestSajikanMedia304(t *testing.T) {
	e := siapkanAuth(t, nil)
	h := uploadBerhasil(t, e, "berita", pngUji(t))
	etag := ambilMedia(e, http.MethodGet, h.URL, nil).Header().Get("ETag")

	for nama, v := range map[string]string{"sama": etag, "weak": "W/" + etag, "daftar": `"lain", ` + etag, "bintang": "*"} {
		rec := ambilMedia(e, http.MethodGet, h.URL, map[string]string{"If-None-Match": v})
		if rec.Code != http.StatusNotModified || rec.Body.Len() != 0 {
			t.Errorf("%s: status %d, body %d", nama, rec.Code, rec.Body.Len())
		}
		if rec.Header().Get("ETag") != etag || rec.Header().Get("Cache-Control") == "" {
			t.Errorf("%s: 304 harus membawa ETag dan Cache-Control: %v", nama, rec.Header())
		}
	}
	rec := ambilMedia(e, http.MethodGet, h.URL, map[string]string{"If-None-Match": `"beda"`})
	if rec.Code != http.StatusOK {
		t.Errorf("ETag beda harus 200, dapat %d", rec.Code)
	}
}

func TestSajikanMediaKeyTidakSahAtauTidakAda404(t *testing.T) {
	e := siapkanAuth(t, nil)
	h := uploadBerhasil(t, e, "berita", pngUji(t))
	const u = "3f9c2a10-7b4e-4c6d-9a1f-0123456789ab"
	for _, path := range []string{
		"/media/berita/" + u + ".png",             // bentuk sah tetapi tidak ada
		"/media/",                                 // kosong
		"/media/berita",                           // bukan key
		"/media/../api/health",                    // traversal
		"/media/berita/../tim/" + u + ".png",      // traversal
		"/media/%2e%2e/etc/passwd",                // traversal ter-encode
		"/media/rahasia/" + u + ".png",            // sumber asing
		"/media/berita/" + u + ".gif",             // ekstensi asing
		strings.Replace(h.URL, ".png", ".PNG", 1), // huruf besar
		h.URL + "/",                               // garis miring akhir
		h.URL + "%00.php",                         // null byte
	} {
		rec := ambilMedia(e, http.MethodGet, path, nil)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: status %d, diharapkan 404", path, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "png") && rec.Code == http.StatusNotFound && strings.Contains(rec.Header().Get("Content-Type"), "image") {
			t.Errorf("%s: 404 tidak boleh berisi gambar", path)
		}
	}
}

func TestSajikanMediaMetodeLainDitolak(t *testing.T) {
	e := siapkanAuth(t, nil)
	h := uploadBerhasil(t, e, "berita", pngUji(t))
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		if rec := ambilMedia(e, m, h.URL, nil); rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: status %d, diharapkan 405", m, rec.Code)
		}
	}
	if e.media.Jumlah() != 1 {
		t.Error("metode ubah-data ke /media tidak boleh mengubah storage")
	}
}

func TestSajikanMediaContentTypeAsingDinetralkan(t *testing.T) {
	e := siapkanAuth(t, nil)
	const u = "3f9c2a10-7b4e-4c6d-9a1f-0123456789ab"
	key := "berita/" + u + ".png"
	// objek asing yang entah bagaimana masuk bucket dengan Content-Type berbahaya
	if err := e.media.Put(context.Background(), key, []byte("<html><script>alert(1)</script></html>"), "text/html"); err != nil {
		t.Fatal(err)
	}
	rec := ambilMedia(e, http.MethodGet, "/media/"+key, nil)
	if rec.Header().Get("Content-Type") != "application/octet-stream" {
		t.Errorf("Content-Type berbahaya harus dinetralkan, dapat %q", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("nosniff wajib")
	}
}

func TestSajikanMediaStorageBermasalah503(t *testing.T) {
	e := siapkanAuth(t, nil)
	h := uploadBerhasil(t, e, "berita", pngUji(t))
	e.media.Err = errors.New("rustfs mati secret=RAHASIA")
	for _, m := range []string{http.MethodGet, http.MethodHead} {
		rec := ambilMedia(e, m, h.URL, nil)
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("%s: status %d", m, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "RAHASIA") {
			t.Error("detail internal bocor")
		}
	}
}

func TestMediaTidakMemakaiNoStore(t *testing.T) {
	e := siapkanAuth(t, nil)
	h := uploadBerhasil(t, e, "berita", pngUji(t))
	if cc := ambilMedia(e, http.MethodGet, h.URL, nil).Header().Get("Cache-Control"); strings.Contains(cc, "no-store") {
		t.Errorf("gambar publik harus bisa di-cache: %q", cc)
	}
}
