-- Query admin testimoni. Alur sama dengan berita (database.md §6): draft -> posted -> retract, soft delete ke trash, pulihkan, hapus permanen.
-- Daftar admin tanpa pagination dan tanpa pencarian (keputusan pengguna). Foto opsional (NULL = tampil inisial nama).

-- name: ListTestimoniAdmin :many
SELECT * FROM testimoni
WHERE (
        (sqlc.arg(filter)::text = 'trash' AND deleted_at IS NOT NULL)
     OR (sqlc.arg(filter)::text <> 'trash' AND deleted_at IS NULL AND (sqlc.arg(filter)::text = 'semua' OR status = sqlc.arg(filter)::text))
      )
ORDER BY COALESCE(deleted_at, created_at) DESC, id DESC;

-- name: GetTestimoniAdmin :one
SELECT * FROM testimoni WHERE id = sqlc.arg(id);

-- name: BuatTestimoni :one
INSERT INTO testimoni (nama, teks, posisi, foto, created_by, updated_by)
VALUES (sqlc.arg(nama), sqlc.arg(teks), sqlc.arg(posisi), sqlc.narg(foto)::text, sqlc.arg(oleh), sqlc.arg(oleh))
RETURNING *;

-- name: UbahTestimoni :one
UPDATE testimoni
SET nama = sqlc.arg(nama), teks = sqlc.arg(teks), posisi = sqlc.arg(posisi), foto = sqlc.narg(foto)::text,
    version = version + 1, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'draft' AND deleted_at IS NULL
RETURNING *;

-- name: PostTestimoni :one
UPDATE testimoni
SET status = 'posted', version = version + 1, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'draft' AND deleted_at IS NULL
RETURNING *;

-- name: RetractTestimoni :one
UPDATE testimoni
SET status = 'draft', version = version + 1, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'posted' AND deleted_at IS NULL
RETURNING *;

-- name: HapusTestimoni :one
UPDATE testimoni
SET deleted_at = now(), updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND status = 'draft' AND deleted_at IS NULL
RETURNING *;

-- name: PulihkanTestimoni :one
UPDATE testimoni
SET deleted_at = NULL, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND deleted_at IS NOT NULL
RETURNING *;

-- name: HapusPermanenTestimoni :one
DELETE FROM testimoni WHERE id = sqlc.arg(id) AND deleted_at IS NOT NULL
RETURNING foto;

-- name: FotoTestimoniDipakai :one
SELECT EXISTS (SELECT 1 FROM testimoni WHERE foto = sqlc.arg(key));
