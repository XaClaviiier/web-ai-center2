# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Status repo: kode lama vs rencana baru

Repo ini punya dua lapisan yang berbeda. Baca `docs/stack.md` §4–§5 sebelum mengubah apa pun.

- **Kode di `src/` dan `server/` adalah implementasi lama** (React + Vite, Bun + Elysia, PGlite/Postgres, login password). Menurut `docs/stack.md`, kode ini **bukan basis implementasi**, hanya referensi konten.
- **Rencana target** (belum ada kodenya): halaman publik Astro (statis, data lewat island + TanStack Query), dashboard admin React + Vite, backend Go + Chi di folder baru `/backend`, PostgreSQL, RustFS, Redis, Keycloak (OIDC), semuanya satu origin. Alur kerja: planning, lalu prototipe HTML/CSS/JS saja, baru implementasi stack final setelah prototipe dinyatakan selesai.
- Jangan memperluas kode lama (mis. menambah fitur di `server/`) kecuali diminta. Tanyakan dulu bila tugas ambigu antara memperbaiki kode lama dan mengerjakan rencana baru.

## Dokumen perencanaan (`docs/`, bahasa Indonesia)

`requirement.md` (kebutuhan, rute, status draft/posted), `stack.md` (stack, routing, penandaan `[RP-xx]` untuk bagian yang bergantung reverse proxy, kontrak API di §9), `database.md` (skema dan field), `security.md` (session, CSRF, rate limit, sanitasi), `design.md`. Dokumen lama ada di `docs/_arsip/`. `docs/backend/` dan `CLAUDE.md` ikut di-track git (agar ikut ke branch `Fullstack`); `backend/.env` tetap di-`.gitignore` dan tidak boleh pernah ikut. Skema database: `docs/database.md`.

Keputusan penting dari dokumen: satu origin tanpa CORS (`ALLOWED_ORIGINS` hanya untuk cek header `Origin` CSRF); URL gambar lewat `/media/*`; berita/dokumentasi punya `status`, `version` (optimistic locking, 409 bila basi), soft delete; endpoint admin `/api/admin/:sumber/:id`, endpoint publik memakai `slug`; konten Layanan/Pelatihan di-hardcode di frontend.

## Aturan wajib untuk dokumen dan perubahan kode

- **Dilarang mengubah file `.md` apa pun di `docs/`** (termasuk `docs/_arsip/`). Bila isinya perlu berubah atau tampak keliru/bertentangan, laporkan ke pengguna dan jangan edit. Satu-satunya pengecualian adalah tiga berkas pencatat di `docs/backend/` yang diatur di bawah: `perubahan-kode.md`, `req.md`, dan `schema-baru.md`.
- **Semua implementasi merujuk ke apa yang tercatat di `.md` dalam `docs/`.** Bila dokumen belum menjawab sesuatu (ada tanda `[diisi]`, `[terbuka]`, `[menunggu ...]`), jangan menebak: tanyakan ke pengguna.

### Pencatatan perubahan (wajib, dalam sesi yang sama dengan perubahannya)

Tiga berkas di `docs/backend/` adalah satu-satunya `.md` di `docs/` yang boleh Anda edit, masing-masing untuk satu jenis perubahan:

- **`perubahan-kode.md`: hanya perubahan MAJOR.** Satu entri per langkah pengerjaan, entri terbaru di paling atas, format entri ada di dalam berkas itu.
  - Perubahan **minor tetapi masih sangat penting** (perbaikan, penyesuaian konfigurasi, pemasangan alat, koreksi tes) tidak diberi entri sendiri. Kelompokkan berdasarkan **tahap** tempat task minor itu terjadi (mis. "Tahap 2, langkah 3") di dalam entri tahap/langkah yang bersangkutan, seperti pada laporan-laporan yang sudah ada. Perubahan sepele (format, ganti nama lokal, komentar) tidak dicatat.
  - **Setiap langkah yang menyentuh `docs/`** wajib mendaftar secara eksplisit di entrinya: (a) bagian `docs/` yang menjadi rujukan, (b) detail atau keputusan yang **melengkapi** bagian dokumen yang masih `[diisi]`/`[terbuka]`/tidak lengkap, dan (c) setiap **penyimpangan** dari isi `docs/` beserta alasan dan siapa yang memutuskan. Dokumen `docs/` sendiri tetap tidak diubah.
  - Keputusan teknis penting dari pengguna (nama env, batas angka, aturan) dicatat di entri yang menerapkannya, bukan di berkas `.md` baru.
- **`req.md`: setiap penambahan library atau alat baru.** Cantumkan nama modul, versi tepat yang terpasang, alasan/reputasi, hasil `govulncheck`, dan catatan kompatibilitas bila ada. Library baru juga disebut singkat di entri `perubahan-kode.md` langkahnya. Perubahan strategi tes dan layanan eksternal juga dicatat di sini.
- **`schema-baru.md`: setiap perubahan struktur atau atribut database** (tabel, kolom, tipe, constraint, index, ekstensi, migrasi). Perbarui tabel kolom, daftar index, bagian "Perbedaan dari skema lama", dan "Gaps" agar sesuai database yang benar-benar terbentuk (cocokkan dengan `information_schema`/`pg_constraint`/`pg_indexes`), lalu rujuk perubahannya dari entri `perubahan-kode.md`.
- Langkah yang tidak menyentuh kode (hanya pertanyaan atau diskusi) tidak dicatat. Jangan mengklaim sebuah langkah selesai sebelum entrinya mencantumkan hasil verifikasi yang benar-benar dijalankan.

## Backend Go baru (`backend/`)

Ini jalur kerja aktif; detail stack dan versi di `docs/backend/req.md`, skema di `docs/backend/schema-baru.md`. `go.mod` memakai `go 1.27.4` (toolchain diunduh otomatis; instalasi sistem 1.27.3; jangan menulis `go.mod` dengan `Set-Content -Encoding utf8` di PowerShell karena menambah BOM dan merusaknya). Jalankan `govulncheck ./...` di awal sesi: advisori baru bisa terbit kapan saja. Go dan `sqlc` ada di `C:\Program Files\Go\bin` dan `%USERPROFILE%\go\bin` (belum tentu di PATH terminal yang sudah terbuka). Konfigurasi dari `backend/.env` (lihat `.env.example`; database `ai_center`, tes di `ai_center_test`, bukan `mydb` milik kode lama).

```
cd backend
go run ./cmd/api                        # server di :8080
go test -p 1 ./... -count=1 -timeout 15m   # -p 1 wajib (paket yang memakai ai_center_test tidak boleh paralel); butuh Docker (PostgreSQL, Redis, RustFS) menyala; suite penuh berjalan beberapa menit, jalankan di latar belakang
go test ./internal/config -run TestLoadGalat   # satu tes
go vet ./... ; gofmt -l .
sqlc generate                           # setelah mengubah queries/ atau migrations/
go run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -dir migrations postgres "$DATABASE_URL" up
govulncheck ./...
```

Auth admin (sementara, sampai Keycloak siap): `POST /api/auth/dev-login` (hanya `APP_ENV=development`, password = `ADMIN_PASSWORD`), `GET /api/auth/me`, `POST /api/auth/logout`. Session di Redis (`session:{id}`), cookie `HttpOnly`; metode ubah-data wajib membawa `Origin` yang ada di `ALLOWED_ORIGINS` dan header `X-CSRF-Token` (ambil dari respons login atau `me`). Route admin memakai middleware `wajibLogin` + `wajibCSRF` (+ `cekOrigin`, `noStore`) dari `internal/httpapi/auth.go`. `REDIS_URL` wajib; tes Redis hanya memakai DB 15.

Gambar: `POST /api/admin/upload` (multipart `sumber` + `gambar`, tipe dicek dari isi, batas `UPLOAD_MAX_MB` default 2) menyimpan ke RustFS dengan key `{sumber}/{uuid}.{ext}` dan mengembalikan `{key,url}`; `GET/HEAD /media/{key}` melayani publik lewat Go (bucket privat). Key dari klien saat Simpan WAJIB divalidasi (`media.SumberDariKey` + sumber cocok + `media.Ada`) dan satu key tidak boleh dipakai dua baris. Tes RustFS hanya memakai bucket `ai-center-test`. **Jangan pernah mendaftar-lalu-menghapus isi bucket `RUSTFS_BUCKET` (dev) untuk membersihkan; hapus hanya key yang Anda buat sendiri.**

Admin konten (berita, dokumentasi, testimoni, dan tim sudah selesai): `/api/admin/dokumentasi` memakai pola status yang sama dengan berita, tetapi gambarnya dinyatakan sebagai KEADAAN AKHIR `gambar:[{key,alt?}]` (1-6 item berurutan; server menurunkan sendiri replace/hapus/tambah/ganti urutan dalam satu transaksi, constraint urutan DEFERRED; file yang dibuang dihapus setelah commit; hapus permanen menghapus semua file; Post memeriksa semua gambar). `/api/admin/berita` dan `/api/admin/testimoni` (daftar `?status=semua|draft|posted|trash`, detail, buat, `PUT` dengan `version`, `post`, `retract`, `restore`, soft delete, `DELETE .../permanen`; daftar testimoni tanpa pagination dan `q`, field `foto` wajib hadir: key atau `null`), `/api/admin/tim` (tanpa status/trash, `DELETE` langsung permanen tanpa `version`, anggota baru di urutan akhir, daftar tanpa pagination). Kode bersama: `internal/admin/berkas.go` (siklus file gambar), `internal/httpapi/admin_umum.go` (pemetaan galat, `versiWajib`, `filterStatus`), `internal/store/admin_umum.go`; repositori diambil lewat adaptor `st.AdminBerita()/AdminTestimoni()/AdminTim()`. Lapisan: handler `internal/httpapi/admin_berita.go` (tipis) -> aturan bisnis `internal/admin` (validasi, slug, status, siklus file gambar; tanpa HTTP/SQL) -> repositori `internal/store/admin_berita.go` + `queries/berita_admin.sql` (aturan status diterapkan atomik di SQL; galat diklasifikasi dari baris yang ada). Aturan: edit/hapus hanya draft, `version` wajib pada `PUT`/`post`/`retract` (409 `version_conflict`), status salah 409 `status_tidak_valid`, `version` hanya naik pada Simpan/Post/Retract, file gambar lama dihapus SETELAH commit, file baru yang gagal disimpan dihapus kembali hanya bila key sah, tidak dipakai baris mana pun, dan ada di bucket. Dokumentasi, testimoni, tim belum dikerjakan.

Struktur: `cmd/api` (entry), `internal/{config,httpapi,store,konten,admin,auth,ratelimit,media,testutil}`, `migrations/` (hanya `00001_skema_awal.sql` dipakai sampai ada data nyata), `queries/` + `sqlc.yaml` (kode di `internal/store/db` adalah hasil generate, jangan diedit). Tahap kerja mengikuti urutan di `docs/backend/perubahan-kode.md`.

## Perintah (kode lama)

Frontend (root, Node >= 20):

```
npm run dev            # Vite di :5173
npm run build          # tsc --noEmit && vite build
npm run typecheck      # sama dengan lint
npm test               # vitest run (jsdom)
npx vitest run src/lib/api.test.ts      # satu file test
npx vitest run -t "nama test"           # satu test menurut nama
```

Backend lama (`server/`, memakai Bun, bukan Node):

```
bun install
bun run dev            # watch, port 3000
bun test               # bun:test (BUKAN vitest; vitest mengecualikan server/)
bun test src/lib/tulis.test.ts          # satu file test
bun run typecheck
bun run db:migrate / db:seed
```

Jalankan frontend dan backend sebagai dua proses terpisah. Frontend membaca `VITE_API_BASE_URL` dari `.env` di root (diisi `http://localhost:3000/api`); tanpa itu default-nya `/api` pada origin Vite dan menghasilkan 404 karena `vite.config.ts` tidak punya proxy. `.env` hanya dibaca saat Vite start, jadi restart setelah mengubahnya. Backend butuh `server/.env` (lihat `server/.env.example`); `ADMIN_PASSWORD` adalah password login `/admin`.

## Arsitektur kode lama

- **Frontend (`src/`)**: React 19 + react-router, Tailwind 4. Semua fetch lewat `src/lib/api.ts` (`ambilJson`/`ambilDaftar` jatuh ke data dummy di `src/data/` bila backend gagal; `kirimJsonAdmin`/`kirimFileAdmin` untuk tulis admin justru MELEMPAR `ErrorApi`). Mode `test` tidak pernah fetch, supaya test deterministik. Halaman admin tunggal `src/pages/Admin.tsx` (login, token Bearer di localStorage). Teks UI ada di `src/lib/kamusBahasa.ts`. `vite.config.ts` punya plugin kustom (peta chunk rute, inline CSS saat build).
- **Backend (`server/src/`)**: Elysia. `index.ts` merakit config, DB, storage RustFS, lalu `createApp` (`app.ts`) yang memasang `@elysia/cors`, penanganan error, dan `registerRoutes` (`routes.ts`, semua route dalam satu file). Akses data lewat interface `Repositories` (`repositories/types.ts`) dengan implementasi Drizzle (`drizzle.ts`) dan mapper baris-ke-respons (`mappers.ts`; kolom NULL dihilangkan dari JSON). Validasi input di `lib/tulis.ts`, query list di `lib/query.ts`, rate-limit in-memory di `lib/rateLimit.ts`, auth token turunan `ADMIN_PASSWORD` di `auth.ts`, gambar di RustFS (`lib/upload.ts`, `lib/orphanUpload.ts`). Skema di `db/schema.ts`. Test memakai repository/storage tiruan.
- Konvensi respons: daftar `{ items: [...] }`, galat `{ error: "pesan" }`; nama field JSON bahasa Indonesia camelCase.
- Deploy: `server/DEPLOY.md`, `deploy/` (Dockerfile + nginx), `docker-compose.production.yml`.

## Catatan kerja

- Semua komentar kode dan dokumen berbahasa Indonesia; ikuti gaya itu.
- Bila ada dua salinan repo (mis. di `Isolated_web_test\...` dan `orca\workspaces\...`), `.env` dan `node_modules` tidak ikut tersalin; pastikan dev server dijalankan dari folder yang punya `.env`.
- `README.md` di root kosong.