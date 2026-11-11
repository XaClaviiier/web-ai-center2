-- Query admin tim. Tim tidak punya status dan trash: hapus langsung permanen (database.md §6).
-- urutan tampil manual; anggota baru ditaruh paling akhir (maksimum + 1). Dua pembuatan serentak bisa berurutan sama; pemutus seri id.

-- name: ListTimAdmin :many
SELECT * FROM tim ORDER BY urutan, id;

-- name: GetTimAdmin :one
SELECT * FROM tim WHERE id = sqlc.arg(id);

-- name: BuatTim :one
INSERT INTO tim (nama, posisi, bidang_riset, foto, urutan, created_by, updated_by)
SELECT sqlc.arg(nama), sqlc.arg(posisi), sqlc.arg(bidang_riset), sqlc.arg(foto), COALESCE(MAX(urutan), 0) + 1, sqlc.arg(oleh), sqlc.arg(oleh)
FROM tim
RETURNING *;

-- name: UbahTim :one
UPDATE tim
SET nama = sqlc.arg(nama), posisi = sqlc.arg(posisi), bidang_riset = sqlc.arg(bidang_riset), foto = sqlc.arg(foto),
    version = version + 1, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version)
RETURNING *;

-- name: HapusTim :one
DELETE FROM tim WHERE id = sqlc.arg(id) RETURNING foto;

-- name: FotoTimDipakai :one
SELECT EXISTS (SELECT 1 FROM tim WHERE foto = sqlc.arg(key));
