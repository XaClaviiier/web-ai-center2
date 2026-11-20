-- Query publik berita: hanya status 'posted' dan belum di-trash. Urutan: tanggal_publikasi terbaru, created_at sebagai pemutus seri.
-- Dua varian (tanpa/dengan pencarian) sengaja dipisah agar rencana kueri stabil dan index trigram bisa dipakai.

-- name: ListBeritaPublik :many
SELECT id, slug, judul, ringkasan, gambar, tanggal_publikasi
FROM berita
WHERE status = 'posted' AND deleted_at IS NULL
ORDER BY tanggal_publikasi DESC, created_at DESC
LIMIT sqlc.arg(batas) OFFSET sqlc.arg(geser)::bigint;

-- name: CariBeritaPublik :many
SELECT id, slug, judul, ringkasan, gambar, tanggal_publikasi
FROM berita
WHERE status = 'posted' AND deleted_at IS NULL AND judul ILIKE sqlc.arg(pola)::text
ORDER BY tanggal_publikasi DESC, created_at DESC
LIMIT sqlc.arg(batas) OFFSET sqlc.arg(geser)::bigint;

-- name: HitungBeritaPublik :one
SELECT count(*) FROM berita WHERE status = 'posted' AND deleted_at IS NULL;

-- name: HitungCariBeritaPublik :one
SELECT count(*) FROM berita
WHERE status = 'posted' AND deleted_at IS NULL AND judul ILIKE sqlc.arg(pola)::text;

-- name: GetBeritaPublik :one
SELECT id, slug, judul, ringkasan, isi, gambar, tanggal_publikasi
FROM berita
WHERE slug = sqlc.arg(slug) AND status = 'posted' AND deleted_at IS NULL;
