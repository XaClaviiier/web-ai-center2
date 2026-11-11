package store

import (
	"context"
	"fmt"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
)

var _ konten.BacaSitemap = (*Store)(nil)

func (s *Store) SitemapBerita(ctx context.Context, batas int32) ([]konten.EntriSitemap, error) {
	baris, err := s.Q.SitemapBerita(ctx, batas)
	if err != nil {
		return nil, fmt.Errorf("sitemap berita: %w", err)
	}
	out := make([]konten.EntriSitemap, 0, len(baris))
	for _, r := range baris {
		out = append(out, konten.EntriSitemap{Slug: r.Slug, Tanggal: tanggal(r.TanggalPublikasi)})
	}
	return out, nil
}

func (s *Store) SitemapDokumentasi(ctx context.Context, batas int32) ([]konten.EntriSitemap, error) {
	baris, err := s.Q.SitemapDokumentasi(ctx, batas)
	if err != nil {
		return nil, fmt.Errorf("sitemap dokumentasi: %w", err)
	}
	out := make([]konten.EntriSitemap, 0, len(baris))
	for _, r := range baris {
		out = append(out, konten.EntriSitemap{Slug: r.Slug, Tanggal: tanggal(r.TanggalPublikasi)})
	}
	return out, nil
}
