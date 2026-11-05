package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func lookupDari(m map[string]string) Lookup {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func dasarDev() map[string]string {
	return map[string]string{
		"APP_ENV":        "development",
		"DATABASE_URL":   "postgres://u:p@localhost:5432/x",
		"ADMIN_PASSWORD": "rahasia",
	}
}

func dasarProd() map[string]string {
	return map[string]string{
		"APP_ENV":         "production",
		"DATABASE_URL":    "postgres://u:p@db:5432/x",
		"PUBLIC_BASE_URL": "https://ai-center.example.com/",
		"ALLOWED_ORIGINS": "https://ai-center.example.com",
	}
}

func TestLoadDevelopmentDefault(t *testing.T) {
	cfg, err := Load(lookupDari(dasarDev()))
	if err != nil {
		t.Fatalf("galat tak terduga: %v", err)
	}
	if cfg.Port != 8080 || cfg.SessionCookieName != "session" || cfg.SessionCookieSecure || !cfg.DevLoginAktif {
		t.Fatalf("default dev salah: %+v", cfg)
	}
}

func TestLoadProduksiMemaksaCookieAman(t *testing.T) {
	cfg, err := Load(lookupDari(dasarProd()))
	if err != nil {
		t.Fatalf("galat tak terduga: %v", err)
	}
	if cfg.SessionCookieName != NamaCookieProduksi || !cfg.SessionCookieSecure || cfg.DevLoginAktif {
		t.Fatalf("produksi harus __Host-session + Secure tanpa dev login: %+v", cfg)
	}
	if cfg.PublicBaseURL != "https://ai-center.example.com" {
		t.Fatalf("slash akhir harus dibuang: %q", cfg.PublicBaseURL)
	}
}

func TestLoadGalat(t *testing.T) {
	cases := []struct {
		nama  string
		ubah  func(m map[string]string)
		dasar func() map[string]string
		wajib string
	}{
		{"APP_ENV kosong", func(m map[string]string) { delete(m, "APP_ENV") }, dasarDev, "APP_ENV wajib"},
		{"APP_ENV asing", func(m map[string]string) { m["APP_ENV"] = "staging" }, dasarDev, "tidak dikenal"},
		{"DATABASE_URL kosong", func(m map[string]string) { delete(m, "DATABASE_URL") }, dasarDev, "DATABASE_URL wajib"},
		{"PORT bukan angka", func(m map[string]string) { m["PORT"] = "abc" }, dasarDev, "PORT"},
		{"PORT di luar rentang", func(m map[string]string) { m["PORT"] = "70000" }, dasarDev, "PORT"},
		{"dev tanpa ADMIN_PASSWORD", func(m map[string]string) { delete(m, "ADMIN_PASSWORD") }, dasarDev, "ADMIN_PASSWORD wajib"},
		{"origin dengan path", func(m map[string]string) { m["ALLOWED_ORIGINS"] = "https://a.example.com/x" }, dasarDev, "ALLOWED_ORIGINS"},
		{"produksi nama cookie salah", func(m map[string]string) { m["SESSION_COOKIE_NAME"] = "session" }, dasarProd, "__Host-session"},
		{"produksi Secure=false", func(m map[string]string) { m["SESSION_COOKIE_SECURE"] = "false" }, dasarProd, "harus true"},
		{"produksi tanpa PUBLIC_BASE_URL", func(m map[string]string) { delete(m, "PUBLIC_BASE_URL") }, dasarProd, "PUBLIC_BASE_URL wajib"},
		{"produksi tanpa ALLOWED_ORIGINS", func(m map[string]string) { delete(m, "ALLOWED_ORIGINS") }, dasarProd, "ALLOWED_ORIGINS wajib"},
	}
	for _, c := range cases {
		t.Run(c.nama, func(t *testing.T) {
			m := c.dasar()
			c.ubah(m)
			_, err := Load(lookupDari(m))
			if err == nil || !strings.Contains(err.Error(), c.wajib) {
				t.Fatalf("diharapkan galat memuat %q, dapat: %v", c.wajib, err)
			}
		})
	}
}

func TestLoadMengumpulkanSemuaGalat(t *testing.T) {
	_, err := Load(lookupDari(map[string]string{"PORT": "0"}))
	if err == nil {
		t.Fatal("harus galat")
	}
	for _, s := range []string{"APP_ENV", "DATABASE_URL", "PORT"} {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("galat harus menyebut %s: %v", s, err)
		}
	}
}

func TestLoadDotEnvTidakMenimpas(t *testing.T) {
	p := filepath.Join(t.TempDir(), ".env")
	isi := "# komentar\nCFG_TEST_BARU=dari-berkas\nCFG_TEST_ADA=dari-berkas\n\nCFG_TEST_SAMA=a=b\n"
	if err := os.WriteFile(p, []byte(isi), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CFG_TEST_ADA", "dari-env")
	t.Cleanup(func() { os.Unsetenv("CFG_TEST_BARU"); os.Unsetenv("CFG_TEST_SAMA") })

	if err := LoadDotEnv(p); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("CFG_TEST_BARU") != "dari-berkas" {
		t.Error("variabel baru harus terisi")
	}
	if os.Getenv("CFG_TEST_ADA") != "dari-env" {
		t.Error("variabel yang sudah ada tidak boleh ditimpa")
	}
	if os.Getenv("CFG_TEST_SAMA") != "a=b" {
		t.Error("nilai yang memuat '=' harus utuh")
	}
	if err := LoadDotEnv(filepath.Join(t.TempDir(), "tidak-ada")); err != nil {
		t.Errorf("berkas tidak ada bukan galat: %v", err)
	}
}
