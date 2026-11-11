-- Query publik dokumentasi. Gambar utama = gambar urutan 1 (headline).

-- name: ListDokumentasiPublik :many
SELECT d.id, d.slug, d.judul, d.tanggal_publikasi, g.key AS gambar_utama
FROM dokumentasi d
LEFT JOIN dokumentasi_gambar g ON g.dokumentasi_id = d.id AND g.urutan = 1
WHERE d.status = 'posted' AND d.deleted_at IS NULL
ORDER BY d.tanggal_publikasi DESC, d.created_at DESC
LIMIT sqlc.arg(batas) OFFSET sqlc.arg(geser)::bigint;

-- name: CariDokumentasiPublik :many
SELECT d.id, d.slug, d.judul, d.tanggal_publikasi, g.key AS gambar_utama
FROM dokumentasi d
LEFT JOIN dokumentasi_gambar g ON g.dokumentasi_id = d.id AND g.urutan = 1
WHERE d.status = 'posted' AND d.deleted_at IS NULL AND d.judul ILIKE sqlc.arg(pola)::text
ORDER BY d.tanggal_publikasi DESC, d.created_at DESC
LIMIT sqlc.arg(batas) OFFSET sqlc.arg(geser)::bigint;

-- name: HitungDokumentasiPublik :one
SELECT count(*) FROM dokumentasi WHERE status = 'posted' AND deleted_at IS NULL;

-- name: HitungCariDokumentasiPublik :one
SELECT count(*) FROM dokumentasi
WHERE status = 'posted' AND deleted_at IS NULL AND judul ILIKE sqlc.arg(pola)::text;

-- name: GetDokumentasiPublik :one
SELECT id, slug, judul, tanggal_publikasi
FROM dokumentasi
WHERE slug = sqlc.arg(slug) AND status = 'posted' AND deleted_at IS NULL;

-- name: ListGambarDokumentasi :many
SELECT key, alt FROM dokumentasi_gambar WHERE dokumentasi_id = sqlc.arg(dokumentasi_id) ORDER BY urutan;
