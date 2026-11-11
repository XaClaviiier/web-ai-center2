package testutil

import (
	"os"
	"testing"
)

// KonfigRustFS adalah koneksi RustFS untuk tes; Bucket selalu bucket TES.
type KonfigRustFS struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

// RustFSTes membaca RUSTFS_* dan TEST_RUSTFS_BUCKET. Tes dilewati bila TEST_RUSTFS_BUCKET kosong.
// Bila bucket tes sama dengan RUSTFS_BUCKET (bucket dev/produksi), tes gagal agar gambar sungguhan tidak terhapus.
func RustFSTes(t *testing.T) KonfigRustFS {
	t.Helper()
	MuatEnv(t)
	bucket := os.Getenv("TEST_RUSTFS_BUCKET")
	if bucket == "" {
		t.Skip("TEST_RUSTFS_BUCKET tidak di-set, tes integrasi RustFS dilewati")
	}
	if bucket == os.Getenv("RUSTFS_BUCKET") {
		t.Fatalf("TEST_RUSTFS_BUCKET (%q) tidak boleh sama dengan RUSTFS_BUCKET: tes mengosongkan bucket ini", bucket)
	}
	k := KonfigRustFS{
		Endpoint:  os.Getenv("RUSTFS_ENDPOINT"),
		AccessKey: os.Getenv("RUSTFS_ACCESS_KEY"),
		SecretKey: os.Getenv("RUSTFS_SECRET_KEY"),
		Bucket:    bucket,
	}
	if k.Endpoint == "" || k.AccessKey == "" || k.SecretKey == "" {
		t.Fatal("RUSTFS_ENDPOINT, RUSTFS_ACCESS_KEY, dan RUSTFS_SECRET_KEY wajib diisi untuk tes RustFS")
	}
	return k
}
