// Package testutil berisi bantuan tes integrasi (memuat .env, URL database tes).
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/config"
)

// MuatEnv memuat backend/.env (dicari naik dari direktori tes sampai ketemu go.mod).
func MuatEnv(t *testing.T) {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
	if err := config.LoadDotEnv(filepath.Join(dir, ".env")); err != nil {
		t.Fatal(err)
	}
}

// DatabaseTes mengembalikan TEST_DATABASE_URL; tes dilewati bila tidak ada (mis. di CI tanpa database).
// Tes integrasi HANYA boleh memakai database tes, tidak pernah ai_center.
func DatabaseTes(t *testing.T) string {
	t.Helper()
	MuatEnv(t)
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL tidak di-set, tes integrasi dilewati")
	}
	return url
}
