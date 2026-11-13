package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/config"
)

func catatPeringatan(cfg config.Config) string {
	var buf bytes.Buffer
	peringatanKonfigurasi(cfg, slog.New(slog.NewTextHandler(&buf, nil)))
	return buf.String()
}

func TestPeringatanKonfigurasi(t *testing.T) {
	if got := catatPeringatan(config.Config{AppEnv: config.EnvProduction}); !strings.Contains(got, "TRUSTED_PROXIES kosong") {
		t.Errorf("production tanpa TRUSTED_PROXIES harus memberi WARN: %q", got)
	}
	if got := catatPeringatan(config.Config{AppEnv: config.EnvDevelopment, PublicBaseURL: "https://ai.example.ac.id"}); !strings.Contains(got, "APP_ENV bukan production") {
		t.Errorf("development di host publik harus memberi WARN: %q", got)
	}
	for _, cfg := range []config.Config{
		{AppEnv: config.EnvDevelopment, PublicBaseURL: "http://localhost:8080"},
		{AppEnv: config.EnvTest, PublicBaseURL: "http://127.0.0.1:8080"},
		{AppEnv: config.EnvProduction, PublicBaseURL: "https://ai.example.ac.id", TrustedProxies: config.Config{}.TrustedProxies},
	} {
		got := catatPeringatan(cfg)
		if cfg.AppEnv != config.EnvProduction && got != "" {
			t.Errorf("tidak boleh ada WARN untuk %+v: %q", cfg, got)
		}
	}
}
