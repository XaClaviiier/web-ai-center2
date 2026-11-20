-- Log aktivitas admin (migrasi 00003). Hanya baca dari aplikasi; baris ditulis trigger catat_aktivitas.
-- Filter tanggal berupa batas UTC [dari, sampai) yang dihitung pemanggil dari hari Asia/Jakarta (NULL = tanpa batas).

-- name: ListLogAktivitas :many
SELECT id, terjadi_pada, pengguna_sub, username, jenis, aksi, modul, item_id, label_item
FROM log_aktivitas
WHERE (sqlc.arg(username)::text = '' OR username = sqlc.arg(username)::text)
  AND (sqlc.arg(jenis)::text = '' OR jenis = sqlc.arg(jenis)::text)
  AND (sqlc.narg(dari)::timestamptz IS NULL OR terjadi_pada >= sqlc.narg(dari)::timestamptz)
  AND (sqlc.narg(sampai)::timestamptz IS NULL OR terjadi_pada < sqlc.narg(sampai)::timestamptz)
ORDER BY terjadi_pada DESC, id DESC
LIMIT sqlc.arg(batas) OFFSET sqlc.arg(geser)::bigint;

-- name: HitungLogAktivitas :one
SELECT count(*) FROM log_aktivitas
WHERE (sqlc.arg(username)::text = '' OR username = sqlc.arg(username)::text)
  AND (sqlc.arg(jenis)::text = '' OR jenis = sqlc.arg(jenis)::text)
  AND (sqlc.narg(dari)::timestamptz IS NULL OR terjadi_pada >= sqlc.narg(dari)::timestamptz)
  AND (sqlc.narg(sampai)::timestamptz IS NULL OR terjadi_pada < sqlc.narg(sampai)::timestamptz);

-- Pilihan filter username: semua username yang pernah tercatat.
-- name: ListUsernameLog :many
SELECT DISTINCT username FROM log_aktivitas ORDER BY username;

-- Retensi bergulir: baris lebih tua dari N bulan kalender (jam database). Dihapus dalam batch terlama dulu agar
-- transaksi pendek; HitungLogKedaluwarsa dipakai mode dry-run. Satu-satunya jalur DELETE pada log dari aplikasi.
-- name: HitungLogKedaluwarsa :one
SELECT count(*) FROM log_aktivitas
WHERE terjadi_pada < now() - make_interval(months => sqlc.arg(bulan)::int);

-- name: HapusLogKedaluwarsa :execrows
DELETE FROM log_aktivitas
WHERE id IN (
    SELECT id FROM log_aktivitas
    WHERE terjadi_pada < now() - make_interval(months => sqlc.arg(bulan)::int)
    ORDER BY id
    LIMIT sqlc.arg(batas)::int
);
