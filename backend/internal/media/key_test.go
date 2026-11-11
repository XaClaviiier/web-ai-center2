package media

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"regexp"
	"strings"
	"testing"
)

func pngKecil(t testing.TB) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func jpegKecil(t testing.TB) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// webpKecil hanya memuat header RIFF/WEBP; cukup untuk deteksi dari isi (sniffing).
func webpKecil() []byte {
	return append([]byte("RIFF\x1a\x00\x00\x00WEBPVP8 "), make([]byte, 20)...)
}

func TestDeteksiGambar(t *testing.T) {
	cases := []struct {
		nama string
		data []byte
		ext  string
		ok   bool
	}{
		{"png", pngKecil(t), "png", true},
		{"jpeg", jpegKecil(t), "jpg", true},
		{"webp", webpKecil(), "webp", true},
		{"kosong", nil, "", false},
		{"gif ditolak", []byte("GIF89a\x01\x00\x01\x00"), "", false},
		{"bmp ditolak", []byte("BM\x00\x00\x00\x00\x00\x00\x00\x00"), "", false},
		{"pdf ditolak", []byte("%PDF-1.7\n..."), "", false},
		{"html ditolak", []byte("<!DOCTYPE html><html><script>alert(1)</script></html>"), "", false},
		{"svg ditolak", []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`), "", false},
		{"teks biasa ditolak", []byte("halo dunia"), "", false},
		{"biner acak ditolak", []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, "", false},
		{"riff bukan webp ditolak", []byte("RIFF\x1a\x00\x00\x00WAVEfmt "), "", false},
		{"php menyamar png: isi php ditolak", []byte("<?php system($_GET['x']); ?>"), "", false},
	}
	for _, c := range cases {
		t.Run(c.nama, func(t *testing.T) {
			j, ok := DeteksiGambar(c.data)
			if ok != c.ok || j.Ext != c.ext {
				t.Fatalf("dapat ext=%q ok=%v; diharapkan ext=%q ok=%v", j.Ext, ok, c.ext, c.ok)
			}
			if ok && !strings.HasPrefix(j.ContentType, "image/") {
				t.Errorf("ContentType salah: %q", j.ContentType)
			}
		})
	}
}

var uuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestBuatKey(t *testing.T) {
	sudah := map[string]bool{}
	for i := 0; i < 2000; i++ {
		key, err := BuatKey(SumberBerita, "jpg")
		if err != nil {
			t.Fatal(err)
		}
		if sudah[key] {
			t.Fatalf("key kembar: %s", key)
		}
		sudah[key] = true
		if !strings.HasPrefix(key, "berita/") || !strings.HasSuffix(key, ".jpg") {
			t.Fatalf("bentuk key salah: %s", key)
		}
		if !uuidV4.MatchString(strings.TrimSuffix(strings.TrimPrefix(key, "berita/"), ".jpg")) {
			t.Fatalf("bagian tengah bukan UUID v4: %s", key)
		}
		if _, ok := SumberDariKey(key); !ok {
			t.Fatalf("key buatan sendiri harus lolos validasi: %s", key)
		}
	}
	if _, err := BuatKey("rahasia", "jpg"); err == nil {
		t.Error("sumber asing harus galat")
	}
	if _, err := BuatKey("../berita", "jpg"); err == nil {
		t.Error("sumber dengan path harus galat")
	}
}

func TestSumberDariKey(t *testing.T) {
	const u = "3f9c2a10-7b4e-4c6d-9a1f-0123456789ab"
	ok := map[string]string{
		"berita/" + u + ".jpg":      "berita",
		"dokumentasi/" + u + ".png": "dokumentasi",
		"testimoni/" + u + ".webp":  "testimoni",
		"tim/" + u + ".jpg":         "tim",
	}
	for key, sumber := range ok {
		got, valid := SumberDariKey(key)
		if !valid || got != sumber {
			t.Errorf("%s: dapat %q %v", key, got, valid)
		}
	}
	buruk := []string{
		"", "berita/", "berita/x.jpg",
		"rahasia/" + u + ".jpg", // sumber asing
		"berita/" + u + ".gif",  // ekstensi asing
		"berita/" + u + ".JPG",  // huruf besar
		"BERITA/" + u + ".jpg",  // sumber huruf besar
		"berita/" + strings.ToUpper(u) + ".jpg",
		"berita/" + u + ".jpg/", // garis miring akhir
		"/berita/" + u + ".jpg", // awalan /
		"berita//" + u + ".jpg",
		"../berita/" + u + ".jpg", // traversal
		"berita/../tim/" + u + ".jpg",
		"berita/" + u + ".jpg\x00.php",                    // null byte
		"berita/" + u + ".jpg ",                           // spasi
		"berita/" + u + ".jpg\n",                          // baris baru (regex ^$ tidak boleh lolos)
		"berita/3f9c2a10-7b4e-1c6d-9a1f-0123456789ab.jpg", // bukan UUID v4
		"berita/3f9c2a10-7b4e-4c6d-7a1f-0123456789ab.jpg", // varian salah
		"berita/" + u + "x.jpg",
	}
	for _, key := range buruk {
		if _, valid := SumberDariKey(key); valid {
			t.Errorf("%q seharusnya tidak sah", key)
		}
	}
}

func TestSumberSah(t *testing.T) {
	for _, s := range []string{"berita", "dokumentasi", "testimoni", "tim"} {
		if !SumberSah(s) {
			t.Errorf("%s harus sah", s)
		}
	}
	for _, s := range []string{"", "Berita", "layanan", "tim/", "berita ", "kursus"} {
		if SumberSah(s) {
			t.Errorf("%q tidak boleh sah", s)
		}
	}
}
