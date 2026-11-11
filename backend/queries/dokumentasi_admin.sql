-- Query admin dokumentasi. Alur status sama dengan berita (database.md §6). Gambar: 1-6 per dokumentasi di tabel dokumentasi_gambar,
-- urutan 1 = headline. Operasi gambar saat Simpan dinyatakan sebagai daftar akhir berurutan (diselesaikan repositori dalam satu transaksi).
-- Constraint UNIQUE (dokumentasi_id, urutan) bersifat DEFERRABLE INITIALLY DEFERRED agar urutan boleh ditukar di dalam satu transaksi.

-- name: ListDokumentasiAdmin :many
SELECT d.id, d.slug, d.judul, d.tanggal_publikasi, d.status, d.version, d.created_at, d.created_by, d.updated_by, d.deleted_at,
       g.key AS gambar_utama,
       (SELECT count(*) FROM dokumentasi_gambar x WHERE x.dokumentasi_id = d.id)::int AS jumlah_gambar
FROM dokumentasi d
LEFT JOIN dokumentasi_gambar g ON g.dokumentasi_id = d.id AND g.urutan = 1
WHERE (
        (sqlc.arg(filter)::text = 'trash' AND d.deleted_at IS NOT NULL)
     OR (sqlc.arg(filter)::text <> 'trash' AND d.deleted_at IS NULL AND (sqlc.arg(filter)::text = 'semua' OR d.status = sqlc.arg(filter)::text))
      )
  AND (sqlc.arg(pola)::text = '' OR d.judul ILIKE sqlc.arg(pola)::text)
ORDER BY COALESCE(d.deleted_at, d.created_at) DESC, d.id DESC
LIMIT sqlc.arg(batas) OFFSET sqlc.arg(geser)::bigint;

-- name: HitungDokumentasiAdmin :one
SELECT count(*) FROM dokumentasi
WHERE (
        (sqlc.arg(filter)::text = 'trash' AND deleted_at IS NOT NULL)
     OR (sqlc.arg(filter)::text <> 'trash' AND deleted_at IS NULL AND (sqlc.arg(filter)::text = 'semua' OR status = sqlc.arg(filter)::text))
      )
  AND (sqlc.arg(pola)::text = '' OR judul ILIKE sqlc.arg(pola)::text);

-- name: GetDokumentasiAdmin :one
SELECT * FROM dokumentasi WHERE id = sqlc.arg(id);

-- Slug bentrok tidak menggagalkan transaksi: ON CONFLICT DO NOTHING mengembalikan nol baris, pemanggil mencoba akhiran berikutnya.
-- name: BuatDokumentasi :one
INSERT INTO dokumentasi (slug, judul, tanggal_publikasi, created_by, updated_by)
VALUES (sqlc.arg(slug), sqlc.arg(judul), sqlc.arg(tanggal_publikasi), sqlc.arg(oleh), sqlc.arg(oleh))
ON CONFLICT (slug) DO NOTHING
RETURNING *;

-- name: UbahDokumentasi :one
UPDATE dokumentasi
SET judul = sqlc.arg(judul), tanggal_publikasi = sqlc.arg(tanggal_publikasi), version = version + 1, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'draft' AND deleted_at IS NULL
RETURNING *;

-- name: PostDokumentasi :one
UPDATE dokumentasi
SET status = 'posted', version = version + 1, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'draft' AND deleted_at IS NULL
RETURNING *;

-- name: RetractDokumentasi :one
UPDATE dokumentasi
SET status = 'draft', version = version + 1, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version) AND status = 'posted' AND deleted_at IS NULL
RETURNING *;

-- name: HapusDokumentasi :one
UPDATE dokumentasi
SET deleted_at = now(), updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND status = 'draft' AND deleted_at IS NULL
RETURNING *;

-- name: PulihkanDokumentasi :one
UPDATE dokumentasi
SET deleted_at = NULL, updated_by = sqlc.arg(oleh)
WHERE id = sqlc.arg(id) AND deleted_at IS NOT NULL
RETURNING *;

-- Baris gambar ikut terhapus lewat ON DELETE CASCADE; pemanggil membaca daftar key SEBELUM menghapus (dalam transaksi yang sama).
-- name: HapusPermanenDokumentasi :one
DELETE FROM dokumentasi WHERE id = sqlc.arg(id) AND deleted_at IS NOT NULL
RETURNING id;

-- Buang baris gambar yang key-nya tidak ada dalam daftar akhir.
-- name: HapusGambarDokumentasiTakDipertahankan :exec
DELETE FROM dokumentasi_gambar
WHERE dokumentasi_id = sqlc.arg(dokumentasi_id) AND NOT (key = ANY(sqlc.arg(dipertahankan)::text[]));

-- Perbarui urutan dan alt gambar yang sudah ada milik dokumentasi ini; 0 baris berarti key baru (pemanggil menyisipkannya).
-- name: PerbaruiGambarDokumentasi :execrows
UPDATE dokumentasi_gambar
SET urutan = sqlc.arg(urutan), alt = sqlc.narg(alt)::text
WHERE dokumentasi_id = sqlc.arg(dokumentasi_id) AND key = sqlc.arg(key);

-- name: TambahGambarDokumentasi :exec
INSERT INTO dokumentasi_gambar (dokumentasi_id, key, urutan, alt)
VALUES (sqlc.arg(dokumentasi_id), sqlc.arg(key), sqlc.arg(urutan), sqlc.narg(alt)::text);

-- Apakah key dipakai baris gambar dokumentasi mana pun (aman-hapus file baru yang gagal disimpan).
-- name: GambarDokumentasiDipakai :one
SELECT EXISTS (SELECT 1 FROM dokumentasi_gambar WHERE key = sqlc.arg(key));
