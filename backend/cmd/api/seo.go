package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/config"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/httpapi"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/seo"
)

// muatSEO membaca kerangka detail (hasil build Astro) dan 404.html dari STATIC_DIR (docs/stack.md bagian 3, RP-12).
// Kerangka hilang atau penandanya tidak lengkap: production menolak start; development hanya memberi peringatan
// dan route detail membalas 503 sampai build tersedia.
func muatSEO(cfg config.Config, log *slog.Logger) (*httpapi.DepsSEO, error) {
	d := &httpapi.DepsSEO{BaseURL: cfg.PublicBaseURL, NamaSitus: cfg.SiteName}

	muatKerangka := func(nama string) (*seo.Kerangka, error) {
		path := filepath.Join(cfg.StaticDir, nama, "_kerangka", "index.html")
		isi, err := os.ReadFile(path)
		if err == nil {
			var k *seo.Kerangka
			if k, err = seo.BuatKerangka(string(isi)); err == nil {
				return k, nil
			}
		}
		err = fmt.Errorf("kerangka %s (%s): %w", nama, path, err)
		if cfg.AppEnv == config.EnvProduction {
			return nil, err
		}
		log.Warn("kerangka SEO tidak tersedia; route detail membalas 503 sampai build ada", "err", err)
		return nil, nil
	}

	var err error
	var galat []error
	if d.KerangkaBerita, err = muatKerangka("berita"); err != nil {
		galat = append(galat, err)
	}
	if d.KerangkaDokumentasi, err = muatKerangka("dokumentasi"); err != nil {
		galat = append(galat, err)
	}
	if err := errors.Join(galat...); err != nil {
		return nil, err
	}

	if isi, err := os.ReadFile(filepath.Join(cfg.StaticDir, "404.html")); err == nil {
		d.Halaman404 = seo.Halaman404(isi)
	}
	return d, nil
}
