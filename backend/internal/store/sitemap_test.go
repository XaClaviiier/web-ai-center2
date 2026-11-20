package store

import (
	"strings"
	"testing"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
)

func slugSitemap(e []konten.EntriSitemap) string {
	var out []string
	for _, x := range e {
		out = append(out, x.Slug+"@"+x.Tanggal)
	}
	return strings.Join(out, ",")
}

func TestSitemapHanyaPostedBelumTrashTerbaruDuluDanBerbatas(t *testing.T) {
	s, ctx := bukaStoreTes(t)
	sisipBerita(t, ctx, s, fixtureBerita{slug: "lama", judul: "Lama", isi: "i", tanggal: "2026-01-01", dibuat: "2026-01-01T00:00:00Z", status: "posted"})
	sisipBerita(t, ctx, s, fixtureBerita{slug: "baru", judul: "Baru", isi: "i", tanggal: "2026-03-01", dibuat: "2026-03-01T00:00:00Z", status: "posted"})
	sisipBerita(t, ctx, s, fixtureBerita{slug: "draft", judul: "D", isi: "i", tanggal: "2026-04-01", dibuat: "2026-04-01T00:00:00Z", status: "draft"})
	sisipBerita(t, ctx, s, fixtureBerita{slug: "buang", judul: "B", isi: "i", tanggal: "2026-05-01", dibuat: "2026-05-01T00:00:00Z", status: "posted", trash: true})
	for _, q := range []string{
		`INSERT INTO dokumentasi (slug, judul, tanggal_publikasi, status, created_by, updated_by) VALUES ('dok-posted', 'P', '2026-02-02', 'posted', 'u', 'u')`,
		`INSERT INTO dokumentasi (slug, judul, tanggal_publikasi, status, created_by, updated_by) VALUES ('dok-draft', 'D', '2026-02-03', 'draft', 'u', 'u')`,
		`INSERT INTO dokumentasi (slug, judul, tanggal_publikasi, status, created_by, updated_by, deleted_at) VALUES ('dok-buang', 'B', '2026-02-04', 'posted', 'u', 'u', now())`,
	} {
		if _, err := s.Pool.Exec(ctx, q); err != nil {
			t.Fatal(err)
		}
	}

	b, err := s.SitemapBerita(ctx, 10)
	if err != nil || slugSitemap(b) != "baru@2026-03-01,lama@2026-01-01" {
		t.Errorf("berita: %q %v", slugSitemap(b), err)
	}
	if b, _ = s.SitemapBerita(ctx, 1); slugSitemap(b) != "baru@2026-03-01" {
		t.Errorf("batas 1 harus memberi yang terbaru: %q", slugSitemap(b))
	}
	d, err := s.SitemapDokumentasi(ctx, 10)
	if err != nil || slugSitemap(d) != "dok-posted@2026-02-02" {
		t.Errorf("dokumentasi: %q %v", slugSitemap(d), err)
	}
}
