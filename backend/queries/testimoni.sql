-- name: ListTestimoniPublik :many
SELECT id, nama, teks, posisi, foto
FROM testimoni
WHERE status = 'posted' AND deleted_at IS NULL
ORDER BY created_at DESC, id DESC;
