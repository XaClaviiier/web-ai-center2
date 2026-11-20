-- Sitemap: hanya slug dan tanggal publikasi item posted dan belum di-trash. Terbaru dulu; LIMIT mencegah sitemap tak terbatas.

-- name: SitemapBerita :many
SELECT slug, tanggal_publikasi FROM berita
WHERE status = 'posted' AND deleted_at IS NULL
ORDER BY tanggal_publikasi DESC, created_at DESC
LIMIT sqlc.arg(batas);

-- name: SitemapDokumentasi :many
SELECT slug, tanggal_publikasi FROM dokumentasi
WHERE status = 'posted' AND deleted_at IS NULL
ORDER BY tanggal_publikasi DESC, created_at DESC
LIMIT sqlc.arg(batas);
