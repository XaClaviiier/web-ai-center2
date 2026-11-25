# Panduan Kontribusi

Aturan kerja tim untuk repo ini. Aturan lama sempat ada di `CLAUDE.md` lokal (dikeluarkan dari repo pada `72f3ca4`); isinya dipindah ke sini agar semua anggota dapat membacanya. Perubahan aturan juga lewat PR.

Cara menyiapkan dan menjalankan proyek ada di `README.md`.

## 1. Alur kerja Git

- **Jangan commit langsung ke `main`.** Semua perubahan lewat branch dan pull request.
- **Satu branch per tugas**, berumur pendek (beberapa hari). Nama branch:
  - `feat/<topik>` untuk fitur, `fix/<topik>` untuk perbaikan, `docs/<topik>` untuk dokumen, `test/<topik>` untuk tes, `ci/<topik>` untuk CI.
- **Pesan commit:** `<jenis>(<area>): <ringkasan>` dalam bahasa Indonesia, mis. `fix(e2e): tes keluar menangani batas login 429`. Jelaskan *mengapa* di badan pesan bila tidak jelas dari judulnya.
- **PR kecil dengan satu tujuan.** Pisahkan commit "hapus/pindah berkas" dari commit "ubah logika" agar peninjau dapat melihat apa yang hilang.
- **Penghapusan dan pemindahan berkas** pakai `git mv`/`git rm`, dan sebutkan alasannya di deskripsi PR.
- **Sebelum commit:** periksa `git status` dan `git diff --staged`; pastikan tidak ada berkas yang tertinggal atau rahasia yang ikut.
- Selesaikan konflik merge di branch tugas, jalankan tes lagi, lalu minta tinjauan ulang.

## 2. Checklist pull request

Sebelum meminta tinjauan:

- [ ] Tes yang relevan dijalankan dan lulus (repo belum punya CI aktif, jadi tes **wajib dijalankan manual**; daftar perintah di bagian "Tes" `README.md`).
- [ ] `backend/.env.example` diperbarui bila ada variabel baru.
- [ ] Dokumen diperbarui bila struktur, perintah, atau perilaku berubah.
- [ ] Perubahan major dicatat sesuai bagian 4.
- [ ] Tidak ada rahasia, `.env`, atau berkas lokal.
- [ ] Penghapusan dan pemindahan dijelaskan.

Minimal satu anggota lain meninjau sebelum merge. Penulis PR tidak dapat menyetujui PR-nya sendiri di GitHub.

## 3. Aturan kode

- **Bahasa:** komentar kode dan dokumen berbahasa Indonesia.
- **`frontend-final/`:**
  1. Diskusikan dulu dengan pemilik tampilan sebelum mengubahnya.
  2. Jangan merusak struktur tampilan (card, CSS, tata letak); perubahan bersifat fungsional atau integrasi.
  3. Catat setiap perubahan di `docs/backend/perubahan-kode.md`.
  4. Setelah berubah, bangun ulang `web/` dengan `node scripts/bangun-web-final.cjs`, lalu jalankan tes.
- **`web/` adalah hasil build.** Jangan menyuntingnya dengan tangan.
- **Kontrak API mengikuti frontend** (tanggal `DD-MM-YYYY`, kode galat `conflict`/`bukan_draft`/`locked`/`lock_lost`, payload dokumentasi `gambarDaftar`). Bila backend belum memenuhi kebutuhan frontend, backend yang disesuaikan.
- **Database:**
  - perubahan skema selalu lewat migrasi **baru** di `backend/migrations/`; migrasi yang sudah berjalan tidak diubah;
  - setelah mengubah `queries/` atau `migrations/`, jalankan `sqlc generate`; kode di `backend/internal/store/db` adalah hasil generate dan tidak diedit tangan.
- **Tes Go** wajib `go test -p 1 ./...`: paket yang memakai `ai_center_test` tidak boleh paralel. Tes yang mengosongkan tabel juga mengosongkan `log_aktivitas` (`testutil.BersihkanKonten`).
- **RustFS:** jangan pernah mendaftar-lalu-menghapus isi bucket dev (`RUSTFS_BUCKET`) untuk membersihkan; hapus hanya key yang Anda buat sendiri. Tes hanya memakai bucket `ai-center-test`.
- **Library baru:** pin versi tepat dan jalankan `govulncheck ./...` (untuk Go).

## 4. Pencatatan perubahan

Tiga berkas di `docs/backend/` mencatat perubahan code base:

- **`perubahan-kode.md`:** hanya perubahan **major**, satu entri per langkah, entri terbaru di atas. Perubahan minor yang masih penting dikelompokkan di dalam entri tahap/langkahnya. Setiap entri yang merujuk `docs/` menyebut (a) bagian yang dirujuk, (b) keputusan yang melengkapi bagian yang masih `[diisi]`/`[terbuka]`, dan (c) penyimpangan beserta alasan dan siapa yang memutuskan.
- **`req.md`:** setiap library atau alat baru: nama, versi tepat, alasan, hasil `govulncheck`, dan catatan kompatibilitas.
- **`schema-baru.md`:** setiap perubahan struktur atau atribut database, dicocokkan dengan database yang benar-benar terbentuk.

Jangan menandai sebuah langkah selesai sebelum entrinya mencantumkan hasil verifikasi yang benar-benar dijalankan.

## 5. Dokumen perencanaan (`docs/`)

- Implementasi merujuk ke dokumen di `docs/`. Bila dokumen belum menjawab sesuatu (`[diisi]`, `[terbuka]`, `[menunggu ...]`), tanyakan ke pemilik keputusan; jangan menebak.
- Mengubah dokumen di `docs/` (selain tiga berkas pencatat di atas) hanya lewat PR yang disetujui pemilik keputusan.

## 6. Rahasia

- `backend/.env`, `keycloak/secrets/`, dan `.env` lain tidak pernah di-commit (sudah di-`.gitignore`).
- Nilai di `.env.example` dan `docker-compose.yml` hanya untuk development. Production memakai rahasia baru (lihat `docs/deploy.md` bila sudah tersedia).
- Bila rahasia sempat ter-commit, anggap bocor: ganti nilainya, jangan hanya menghapus commit.

## 7. Identitas Git

Pakai satu nama dan email yang sama di semua komputer (`git config --global user.name` dan `user.email`), sesuai akun GitHub Anda. Identitas lama yang berbeda dipetakan di `.mailmap`.
