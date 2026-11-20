package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/config"
)

const kerangkaSah = `<html><head><!--seo:head--></head><body><!--seo:konten--><!--seo:data--></body></html>`

func tulisKerangka(t *testing.T, dir, nama, isi string) {
	t.Helper()
	p := filepath.Join(dir, nama, "_kerangka")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "index.html"), []byte(isi), 0o644); err != nil {
		t.Fatal(err)
	}
}

func cfgUji(env, dir string) config.Config {
	return config.Config{AppEnv: env, StaticDir: dir, SiteName: "AI Center", PublicBaseURL: "https://situs.test"}
}

var logUji = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestMuatSEOMemuatKerangkaDan404(t *testing.T) {
	dir := t.TempDir()
	tulisKerangka(t, dir, "berita", kerangkaSah)
	tulisKerangka(t, dir, "dokumentasi", kerangkaSah)
	if err := os.WriteFile(filepath.Join(dir, "404.html"), []byte("<html><head></head><body>x</body></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := muatSEO(cfgUji(config.EnvProduction, dir), logUji)
	if err != nil || d.KerangkaBerita == nil || d.KerangkaDokumentasi == nil {
		t.Fatalf("mau berhasil: %+v %v", d, err)
	}
	if !strings.Contains(string(d.Halaman404), `content="noindex"`) || d.BaseURL != "https://situs.test" || d.NamaSitus != "AI Center" {
		t.Errorf("404.html/konfigurasi salah: %+v", d)
	}
}

func TestMuatSEOKerangkaHilangAtauRusak(t *testing.T) {
	kosong := t.TempDir()
	if _, err := muatSEO(cfgUji(config.EnvProduction, kosong), logUji); err == nil || !strings.Contains(err.Error(), "kerangka berita") || !strings.Contains(err.Error(), "kerangka dokumentasi") {
		t.Errorf("production tanpa kerangka harus gagal untuk keduanya: %v", err)
	}
	d, err := muatSEO(cfgUji(config.EnvDevelopment, kosong), logUji)
	if err != nil || d.KerangkaBerita != nil || d.KerangkaDokumentasi != nil || d.Halaman404 != nil {
		t.Errorf("development tanpa kerangka harus lolos dengan nil: %+v %v", d, err)
	}

	rusak := t.TempDir()
	tulisKerangka(t, rusak, "berita", `<html><head></head><body></body></html>`) // tanpa penanda
	tulisKerangka(t, rusak, "dokumentasi", kerangkaSah)
	if _, err := muatSEO(cfgUji(config.EnvProduction, rusak), logUji); err == nil || !strings.Contains(err.Error(), "penanda") {
		t.Errorf("production dengan penanda tidak lengkap harus gagal: %v", err)
	}
	d, err = muatSEO(cfgUji(config.EnvDevelopment, rusak), logUji)
	if err != nil || d.KerangkaBerita != nil || d.KerangkaDokumentasi == nil {
		t.Errorf("development: berita nil (rusak), dokumentasi ada: %+v %v", d, err)
	}
}
