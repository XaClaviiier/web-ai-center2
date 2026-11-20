package media

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

// bukaS3Tes membuka S3Store ke bucket TES (dibuat bila belum ada) dan mengosongkannya sebelum dan sesudah tes.
func bukaS3Tes(t *testing.T) (*S3Store, context.Context) {
	t.Helper()
	k := testutil.RustFSTes(t)
	ctx, batal := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(batal)
	s := NewS3(S3Config{Endpoint: k.Endpoint, AccessKey: k.AccessKey, SecretKey: k.SecretKey, Bucket: k.Bucket})
	if err := s.PastikanBucket(ctx, true); err != nil {
		t.Fatalf("PastikanBucket: %v", err)
	}
	kosongkan := func() {
		daftar, err := s.Daftar(context.Background(), "")
		if err != nil {
			t.Fatalf("daftar bucket tes: %v", err)
		}
		for _, o := range daftar {
			if err := s.Hapus(context.Background(), o.Key); err != nil {
				t.Fatalf("hapus %s: %v", o.Key, err)
			}
		}
	}
	kosongkan()
	t.Cleanup(kosongkan)
	return s, ctx
}

func TestS3PutGetHeadHapus(t *testing.T) {
	s, ctx := bukaS3Tes(t)
	data := pngKecil(t)
	key, _ := BuatKey(SumberBerita, "png")

	if ada, err := Ada(ctx, s, key); err != nil || ada {
		t.Fatalf("sebelum Put harus tidak ada: %v %v", ada, err)
	}
	if err := s.Put(ctx, key, data, "image/png"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if ada, err := Ada(ctx, s, key); err != nil || !ada {
		t.Fatalf("sesudah Put harus ada: %v %v", ada, err)
	}

	meta, err := s.Head(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if meta.ContentType != "image/png" || meta.Ukuran != int64(len(data)) || meta.ETag == "" {
		t.Errorf("metadata salah: %+v", meta)
	}

	o, err := s.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Body.Close()
	got, _ := io.ReadAll(o.Body)
	if string(got) != string(data) || o.ContentType != "image/png" || o.ETag != meta.ETag {
		t.Errorf("isi/metadata Get tidak sama dengan yang disimpan")
	}

	if err := s.Hapus(ctx, key); err != nil {
		t.Fatal(err)
	}
	if ada, _ := Ada(ctx, s, key); ada {
		t.Error("sesudah Hapus harus tidak ada")
	}
	if err := s.Hapus(ctx, key); err != nil {
		t.Errorf("Hapus key yang tidak ada harus idempotent, dapat %v", err)
	}
}

func TestS3TidakAda(t *testing.T) {
	s, ctx := bukaS3Tes(t)
	key, _ := BuatKey(SumberTim, "jpg")
	if _, err := s.Get(ctx, key); !errors.Is(err, ErrTidakAda) {
		t.Errorf("Get: diharapkan ErrTidakAda, dapat %v", err)
	}
	if _, err := s.Head(ctx, key); !errors.Is(err, ErrTidakAda) {
		t.Errorf("Head: diharapkan ErrTidakAda, dapat %v", err)
	}
}

func TestS3Daftar(t *testing.T) {
	s, ctx := bukaS3Tes(t)
	for _, sumber := range []string{SumberBerita, SumberBerita, SumberTim} {
		key, _ := BuatKey(sumber, "jpg")
		if err := s.Put(ctx, key, jpegKecil(t), "image/jpeg"); err != nil {
			t.Fatal(err)
		}
	}
	semua, err := s.Daftar(ctx, "")
	if err != nil || len(semua) != 3 {
		t.Fatalf("semua: %d %v", len(semua), err)
	}
	berita, err := s.Daftar(ctx, "berita/")
	if err != nil || len(berita) != 2 {
		t.Fatalf("prefix berita/: %d %v", len(berita), err)
	}
	if berita[0].Ukuran == 0 || berita[0].Diubah.IsZero() {
		t.Errorf("Info tidak lengkap: %+v", berita[0])
	}
}

func TestS3BucketTidakAda(t *testing.T) {
	k := testutil.RustFSTes(t)
	ctx, batal := context.WithTimeout(context.Background(), 30*time.Second)
	defer batal()
	s := NewS3(S3Config{Endpoint: k.Endpoint, AccessKey: k.AccessKey, SecretKey: k.SecretKey, Bucket: "ai-center-bucket-yang-tidak-ada"})
	if err := s.PastikanBucket(ctx, false); err == nil {
		t.Fatal("bucket tidak ada dengan buat=false harus galat")
	}
}

func TestS3KredensialSalahBukanTidakAda(t *testing.T) {
	k := testutil.RustFSTes(t)
	ctx, batal := context.WithTimeout(context.Background(), 30*time.Second)
	defer batal()
	s := NewS3(S3Config{Endpoint: k.Endpoint, AccessKey: k.AccessKey, SecretKey: "kunci-salah", Bucket: k.Bucket})
	err := s.Put(ctx, "berita/x.jpg", []byte("x"), "image/jpeg")
	if err == nil || errors.Is(err, ErrTidakAda) {
		t.Fatalf("kredensial salah harus galat biasa, dapat %v", err)
	}
	if _, err := s.Head(ctx, "berita/x.jpg"); err == nil || errors.Is(err, ErrTidakAda) {
		t.Fatalf("Head dengan kredensial salah harus galat biasa, bukan ErrTidakAda: %v", err)
	}
}

func TestMemoriMemenuhiKontrakYangSama(t *testing.T) {
	m := NewMemori()
	ctx := context.Background()
	key, _ := BuatKey(SumberTestimoni, "webp")
	if _, err := m.Get(ctx, key); !errors.Is(err, ErrTidakAda) {
		t.Fatalf("Get kosong: %v", err)
	}
	if err := m.Put(ctx, key, webpKecil(), "image/webp"); err != nil {
		t.Fatal(err)
	}
	meta, err := m.Head(ctx, key)
	if err != nil || meta.ContentType != "image/webp" || meta.ETag == "" || meta.Ukuran != int64(len(webpKecil())) {
		t.Fatalf("Head: %+v %v", meta, err)
	}
	if err := m.Hapus(ctx, key); err != nil || m.Jumlah() != 0 {
		t.Fatalf("Hapus: %v jumlah=%d", err, m.Jumlah())
	}
	m.Err = errors.New("storage mati")
	if err := m.Put(ctx, key, nil, ""); err == nil {
		t.Error("Err harus diteruskan")
	}
}
