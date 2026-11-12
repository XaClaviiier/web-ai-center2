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
