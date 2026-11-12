-- Keputusan pengguna: teks alternatif gambar dokumentasi tidak dipakai (form admin tidak punya input alt).
-- Alt di halaman publik memakai teks bawaan dari judul. Migrasi baru (bukan mengubah 00001) agar database yang sudah
-- berisi data uji tidak perlu direset.

-- +goose Up
ALTER TABLE dokumentasi_gambar DROP COLUMN alt;

-- +goose Down
ALTER TABLE dokumentasi_gambar ADD COLUMN alt text;
