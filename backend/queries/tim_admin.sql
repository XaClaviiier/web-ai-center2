-- Query admin tim. Tim tidak punya status dan trash: hapus langsung permanen (database.md §6).
-- urutan tampil manual: diisi admin (angka bulat >= 1, kecil tampil lebih dulu). Urutan sama diperbolehkan; pemutus seri id.

-- name: ListTimAdmin :many
SELECT * FROM tim ORDER BY urutan, id;

-- name: GetTimAdmin :one
SELECT * FROM tim WHERE id = sqlc.arg(id);

-- name: BuatTim :one
INSERT INTO tim (nama, posisi, bidang_riset, foto, urutan, created_by, updated_by)
VALUES (sqlc.arg(nama), sqlc.arg(posisi), sqlc.arg(bidang_riset), sqlc.arg(foto), sqlc.arg(urutan), sqlc.arg(oleh), sqlc.arg(oleh))
RETURNING *;

-- name: UbahTim :one
UPDATE tim
SET nama = sqlc.arg(nama), posisi = sqlc.arg(posisi), bidang_riset = sqlc.arg(bidang_riset), foto = sqlc.arg(foto), urutan = sqlc.arg(urutan),
    version = version + 1, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version)
RETURNING *;

-- name: HapusTim :one
DELETE FROM tim WHERE id = sqlc.arg(id) RETURNING foto;

-- name: FotoTimDipakai :one
SELECT EXISTS (SELECT 1 FROM tim WHERE foto = sqlc.arg(key));
