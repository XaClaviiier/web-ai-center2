-- name: ListTim :many
SELECT id, nama, posisi, bidang_riset, foto FROM tim ORDER BY urutan, id;
