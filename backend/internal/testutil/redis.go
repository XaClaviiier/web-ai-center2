package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

// DBRedisTes adalah satu-satunya database Redis yang boleh dikosongkan oleh tes.
const DBRedisTes = 15

// RedisTes membuka klien ke TEST_REDIS_URL dan mengosongkan database-nya sebelum dan sesudah tes.
// Tes dilewati bila TEST_REDIS_URL tidak di-set. Sebelum FLUSHDB, URL dipastikan menunjuk DB 15 supaya
// data dev (DB 0) tidak pernah terhapus.
func RedisTes(t *testing.T) *redis.Client {
	t.Helper()
	MuatEnv(t)
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("TEST_REDIS_URL tidak di-set, tes integrasi Redis dilewati")
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("TEST_REDIS_URL tidak valid: %v", err)
	}
	if opt.DB != DBRedisTes {
		t.Fatalf("TEST_REDIS_URL harus menunjuk DB %d (tes mengosongkan database ini), dapat DB %d", DBRedisTes, opt.DB)
	}
	c := redis.NewClient(opt)
	bersih := func() {
		if err := c.FlushDB(context.Background()).Err(); err != nil {
			t.Fatalf("mengosongkan Redis tes: %v", err)
		}
	}
	// Close didaftarkan lebih dulu agar dijalankan terakhir (t.Cleanup berjalan terbalik).
	t.Cleanup(func() { _ = c.Close() })
	if err := c.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("Redis tes tidak terjangkau: %v", err)
	}
	bersih()
	t.Cleanup(bersih)
	return c
}
