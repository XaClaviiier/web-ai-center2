-- Query untuk job berkala: item trash kedaluwarsa (index parsial *_trash_idx) dan pemakaian key gambar.
-- Jam yang dipakai adalah jam database (now()), bukan jam aplikasi.

-- name: TrashBeritaKedaluwarsa :many
SELECT id FROM berita
WHERE deleted_at IS NOT NULL AND deleted_at < now() - make_interval(days => sqlc.arg(hari)::int)
ORDER BY deleted_at
LIMIT sqlc.arg(batas);

-- name: TrashDokumentasiKedaluwarsa :many
SELECT id FROM dokumentasi
WHERE deleted_at IS NOT NULL AND deleted_at < now() - make_interval(days => sqlc.arg(hari)::int)
ORDER BY deleted_at
LIMIT sqlc.arg(batas);

-- name: TrashTestimoniKedaluwarsa :many
SELECT id FROM testimoni
WHERE deleted_at IS NOT NULL AND deleted_at < now() - make_interval(days => sqlc.arg(hari)::int)
ORDER BY deleted_at
LIMIT sqlc.arg(batas);

-- name: KeyGambarDipakai :one
-- Dipakai oleh baris mana pun pada keempat jenis konten, termasuk draft dan trash.
SELECT (
    EXISTS (SELECT 1 FROM berita WHERE gambar = sqlc.arg(key)::text)
 OR EXISTS (SELECT 1 FROM dokumentasi_gambar WHERE key = sqlc.arg(key)::text)
 OR EXISTS (SELECT 1 FROM testimoni WHERE foto = sqlc.arg(key)::text)
 OR EXISTS (SELECT 1 FROM tim WHERE foto = sqlc.arg(key)::text)
)::boolean AS dipakai;
