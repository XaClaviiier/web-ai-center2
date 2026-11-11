-- Query admin berita. Filter status: 'semua' (draft + posted, tanpa trash), 'draft', 'posted', 'trash'.
-- Aturan status (database.md §6, requirement.md §5.1): edit dan hapus hanya untuk draft; trash selalu berstatus draft.
-- version naik 1 pada Simpan, Post, dan Retract saja (database.md §4); hapus dan pulihkan tidak menaikkannya.

-- name: ListBeritaAdmin :many
SELECT * FROM berita
WHERE (
        (sqlc.arg(filter)::text = 'trash' AND deleted_at IS NOT NULL)
     OR (sqlc.arg(filter)::text <> 'trash' AND deleted_at IS NULL AND (sqlc.arg(filter)::text = 'semua' OR status = sqlc.arg(filter)::text))
      )
  AND (sqlc.arg(pola)::text = '' OR judul ILIKE sqlc.arg(pola)::text)
ORDER BY COALESCE(deleted_at, created_at) DESC, id DESC
LIMIT sqlc.arg(batas) OFFSET sqlc.arg(geser)::bigint;

-- name: HitungBeritaAdmin :one
SELECT count(*) FROM berita
WHERE (
        (sqlc.arg(filter)::text = 'trash' AND deleted_at IS NOT NULL)
     OR (sqlc.arg(filter)::text <> 'trash' AND deleted_at IS NULL AND (sqlc.arg(filter)::text = 'semua' OR status = sqlc.arg(filter)::text))
      )
  AND (sqlc.arg(pola)::text = '' OR judul ILIKE sqlc.arg(pola)::text);

-- name: GetBeritaAdmin :one
SELECT * FROM berita WHERE id = sqlc.arg(id);

-- Slug bentrok tidak menggagalkan transaksi: ON CONFLICT DO NOTHING mengembalikan nol baris, pemanggil mencoba akhiran berikutnya.
-- name: BuatBerita :one
INSERT INTO berita (slug, judul, isi, tanggal_publikasi, gambar, created_by, updated_by)
VALUES (sqlc.arg(slug), sqlc.arg(judul), sqlc.arg(isi), sqlc.arg(tanggal_publikasi), sqlc.arg(gambar), sqlc.arg(oleh), sqlc.arg(oleh))
ON CONFLICT (slug) DO NOTHING
RETURNING *;

-- name: UbahBerita :one
UPDATE berita
SET judul = sqlc.arg(judul), isi = sqlc.arg(isi), tanggal_publikasi = sqlc.arg(tanggal_publikasi), gambar = sqlc.arg(gambar),
    version = version + 1, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'draft' AND deleted_at IS NULL
RETURNING *;

-- name: PostBerita :one
UPDATE berita
SET status = 'posted', version = version + 1, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'draft' AND deleted_at IS NULL
RETURNING *;

-- name: RetractBerita :one
UPDATE berita
SET status = 'draft', version = version + 1, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'posted' AND deleted_at IS NULL
RETURNING *;

-- name: HapusBerita :one
UPDATE berita
SET deleted_at = now(), updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND status = 'draft' AND deleted_at IS NULL
RETURNING *;

-- name: PulihkanBerita :one
UPDATE berita
SET deleted_at = NULL, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND deleted_at IS NOT NULL
RETURNING *;

-- name: HapusPermanenBerita :one
DELETE FROM berita WHERE id = sqlc.arg(id) AND deleted_at IS NOT NULL
RETURNING gambar;

-- Apakah key gambar dipakai baris berita mana pun (aman-hapus file baru yang gagal disimpan).
-- name: GambarBeritaDipakai :one
SELECT EXISTS (SELECT 1 FROM berita WHERE gambar = sqlc.arg(key));
