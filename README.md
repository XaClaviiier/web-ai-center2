# UBAYA AI Center

Situs web UBAYA AI Center: halaman publik (beranda, berita, dokumentasi, tim, tentang kami, layanan, dan tiga demo konsep) beserta admin untuk mengelola konten (berita, dokumentasi, testimoni, tim) dan melihat Log Aktivitas. Satu server Go menyajikan frontend statis, API, dan gambar dari satu origin (tanpa CORS).

## Cara kerja singkat

- **Publik:** halaman statis hasil build Astro. Halaman detail berita/dokumentasi memakai satu kerangka; Go menyisipkan `<title>`, meta, Open Graph, JSON-LD, dan data awal per slug (SEO), lalu React mengambil alih tampilan.
- **Admin (`/admin`):** aplikasi React satu halaman. Alur konten: draft, Post, Retract, trash, pulihkan, hapus permanen. Penguncian edit (soft lock) dan `version` (optimistic locking) mencegah dua admin saling menimpa.
- **Gambar:** diunggah lewat `/api/admin/upload`, disimpan di RustFS (S3), dilayani publik lewat `/media/*`.
- **Login:** sementara `dev-login` (hanya `APP_ENV=development`) lewat `/admin/masuk`. Keycloak menyusul.

## Stack

| Bagian | Teknologi | Versi |
| ------ | --------- | ----- |
| Backend | Go + Chi | Go 1.27.4, chi v5.5.0 |
| Database | PostgreSQL (driver pgx v5.12.0, migrasi goose v3.29.0, kode query sqlc v1.31.1) | 15 |
| Cache, session, lock, rate limit | Redis (go-redis v9.23.0) | 7 |
| Penyimpanan gambar | RustFS (S3-compatible, aws-sdk-go-v2 s3 v1.114.6) | 1.0.3 (image `rustfs/rustfs:1.0.3`) |
| Frontend publik | Astro + React (island) | Astro 7.4.11, React 19.3.1 |
| Admin | React + Vite + TanStack Query | Vite 8.3.8, TanStack Query 5.107.1 |
| Gaya | Tailwind CSS (+ CSS buatan tim) | 4.3.5 |
| Bahasa frontend | TypeScript | 6.0.3 |
| Animasi/UI | GSAP 3.15.1, motion 14.13.0, Embla 8.6.0, Radix UI | lihat `frontend-final/package.json` |
| Build frontend, tes unit | Bun | >= 1.4.4 (terpasang 1.4.3 memberi peringatan, tetap jalan) |
| Tes end-to-end | Playwright | 1.65.0 |
| Autentikasi produksi (belum) | Keycloak (OIDC) | menunggu pengelola |

Versi backend lengkap, alasan pemilihan, dan hasil `govulncheck`: `docs/backend/req.md`. Skema database: `docs/backend/schema-baru.md`.

## Struktur folder

| Folder | Isi |
| ------ | --- |
| `backend/` | Server Go (`cmd/api`, `internal/*`, `migrations/`, `queries/`) |
| `frontend-final/` | Sumber frontend (Astro + React admin + klien API di `src/api/api.ts`) |
| `web/` | HASIL BUILD `frontend-final/`, disajikan Go, di-commit. Jangan disunting tangan |
| `scripts/` | Pembangun `web/`, tes Node/Bun, tes e2e Playwright |
| `docs/` | Dokumen perencanaan (bahasa Indonesia) dan catatan kerja backend (`docs/backend/`) |

## Prasyarat

- Go 1.27.x (toolchain 1.27.4 diunduh otomatis oleh `go.mod`)
- Docker untuk PostgreSQL 15, Redis 7, dan RustFS (compose dikelola sendiri; repo tidak membawa compose)
- Bun >= 1.4.4 (build frontend). Build memakai `bun --bun` karena Astro 7 menolak Node di bawah 22.23.4
- Node.js (untuk skrip `scripts/*.cjs` dan Playwright) dan `npm install` di akar repo
- Opsional: `sqlc` v1.31.1 dan `govulncheck`

## Menjalankan proyek

1. **Layanan:** nyalakan PostgreSQL (`postgres:15`), Redis (`redis:7-alpine`), dan RustFS (`rustfs/rustfs:1.0.3`, kunci versinya; jangan `latest`) di Docker. Buat database `ai_center` (dan `ai_center_test` untuk tes) secara manual (`CREATE DATABASE`); server tidak membuatnya. Bucket RustFS dibuat otomatis oleh server di development (di production harus dibuat dulu), dan bucket tes `ai-center-test` dibuat otomatis oleh tes.
2. **Konfigurasi:** salin `backend/.env.example` ke `backend/.env`, lalu isi (koneksi database/Redis/RustFS, `ADMIN_PASSWORD`, `STATIC_DIR` mengarah ke folder `web/`). `backend/.env` ter-ignore git dan tidak boleh di-commit.
3. **Migrasi database:**
   ```
   cd backend
   go run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -dir migrations postgres "$DATABASE_URL" up
   ```
4. **Frontend:** `web/` sudah di-commit. Bangun ulang hanya bila `frontend-final/` berubah:
   ```
   node scripts/bangun-web-final.cjs              # dari akar repo; tambah --tanpa-install bila dependensi sudah terpasang
   ```
   Skrip membangun langsung di `frontend-final/` lalu menulis ulang `web/` seluruhnya (termasuk `web/csp-hash-skrip.txt`, hash CSP yang dibaca Go).
5. **Server:**
   ```
   cd backend
   go run ./cmd/api                               # http://localhost:8080
   ```
   Admin: `http://localhost:8080/admin/`, masuk dengan `ADMIN_PASSWORD` (hanya `APP_ENV=development`). Di development, set `JOBS_ENABLED=false` agar job pembersihan tidak menyentuh bucket dev.

## Tes

```
cd backend
go vet ./...
go test -p 1 ./... -count=1 -timeout 15m       # -p 1 wajib; butuh Docker menyala; beberapa menit
govulncheck ./...

# dari akar repo
node scripts/tests/web.test.cjs                # tes statis atas web/ (REBUILD=1 untuk build ulang dulu)
bun test scripts/tests/api.test.ts             # klien API dengan fetch palsu
npm run test:e2e                               # Playwright; server harus jalan (BASE_URL, PW=ADMIN_PASSWORD, PW2=ADMIN_PASSWORD_2 opsional)
```

Tes memakai `ai_center_test`, Redis DB 15, dan bucket `ai-center-test`; data dev tidak disentuh. Spec e2e admin memberi jeda sekitar 61 detik agar batas tulis admin tidak terlampaui.

## Keamanan singkat

Session di Redis dengan cookie `HttpOnly`; metode ubah-data wajib `Origin` yang diizinkan dan header `X-CSRF-Token`; rate limit per IP/akun; CSP ketat tanpa `unsafe-inline` (skrip inline Astro diizinkan lewat hash); tipe gambar diperiksa dari isi file. Detail: `docs/security.md`.

## Dokumen

`docs/stack.md` (stack dan kontrak API), `docs/requirement.md`, `docs/database.md`, `docs/security.md`, `docs/design.md`, dan `docs/backend/` (catatan perubahan, rencana merge, skema). Aturan kerja untuk agen/kontributor ada di `CLAUDE.md`.
