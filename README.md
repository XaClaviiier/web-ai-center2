# UBAYA AI Center

Situs web UBAYA AI Center: halaman publik (beranda, berita, dokumentasi, tim, tentang kami, layanan, dan tiga demo konsep) beserta admin untuk mengelola konten (berita, dokumentasi, testimoni, tim) dan melihat Log Aktivitas. Satu server Go menyajikan frontend statis, API, dan gambar dari satu origin (tanpa CORS).

## Cara kerja singkat

- **Publik:** halaman statis hasil build Astro. Halaman detail berita/dokumentasi memakai satu kerangka; Go menyisipkan `<title>`, meta, Open Graph, JSON-LD, dan data awal per slug (SEO), lalu React mengambil alih tampilan.
- **Admin (`/admin`):** aplikasi React satu halaman. Alur konten: draft, Post, Retract, trash, pulihkan, hapus permanen. Penguncian edit (soft lock) dan `version` (optimistic locking) mencegah dua admin saling menimpa.
- **Gambar:** diunggah lewat `/api/admin/upload`, disimpan di RustFS (S3), dilayani publik lewat `/media/*`.
- **Login:** Keycloak (OIDC Authorization Code + PKCE). `/admin/` tanpa session diarahkan ke Keycloak; hanya client role `direktur`/`pegawai` pada client `ai-center-cms` yang diterima. Konfigurasi dan langkah operator: `keycloak/README.md`.

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
| Autentikasi | Keycloak (OIDC; go-oidc + x/oauth2 di backend) | 26.8.1 (`keycloak/`) |

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
- Docker untuk PostgreSQL 15, Redis 7, dan RustFS (`docker-compose.yml` di akar repo) serta Keycloak dev (`keycloak/compose.dev.yml`)
- Bun >= 1.4.4 (build frontend). Build memakai `bun --bun` karena Astro 7 menolak Node di bawah 22.23.4
- Node.js (untuk skrip `scripts/*.cjs` dan Playwright) dan `npm install` di akar repo
- Opsional: `sqlc` v1.31.1 dan `govulncheck`

## Menjalankan proyek

1. **Layanan:** dari akar repo, `docker compose up -d` menyalakan PostgreSQL 15 (database `ai_center`), Redis 7, dan RustFS 1.0.3; data tersimpan di `pgdata/`, `redis_data/`, `rustfs-data/` (ter-ignore). Database tes dibuat sekali secara manual: `docker compose exec db psql -U user -d ai_center -c "CREATE DATABASE ai_center_test"`. Bucket RustFS dibuat otomatis oleh server di development (di production harus dibuat dulu), dan bucket tes `ai-center-test` dibuat otomatis oleh tes. Port 5432, 6379, 8080, 8180, 9000, dan 9001 harus bebas.
2. **Konfigurasi:** salin `backend/.env.example` ke `backend/.env`. Kredensial database/Redis/RustFS di dalamnya sudah cocok dengan `docker-compose.yml`; `ALLOWED_ORIGINS` harus tetap memuat `PUBLIC_BASE_URL` (tanpanya semua tulis admin 403). Keycloak dev (port 8180) disiapkan sesuai `keycloak/README.md`; client secret, kunci session, dan akun uji (`uji-direktur`, `uji-pegawai`, `uji-tanpa-role`) dapat dibuat otomatis dengan `bash keycloak/siapkan-dev.sh` setelah Keycloak jalan. `backend/.env` dan `keycloak/secrets/` ter-ignore git dan tidak boleh di-commit.
3. **Migrasi database:** `.env` tidak dimuat otomatis ke shell, jadi ambil URL-nya dari berkas. Migrasikan juga database tes:
   ```
   cd backend
   go run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -dir migrations postgres "$(sed -n 's/^DATABASE_URL=//p' .env | tr -d '\r')" up
   go run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -dir migrations postgres "$(sed -n 's/^TEST_DATABASE_URL=//p' .env | tr -d '\r')" up
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
   Admin: `http://localhost:8080/admin/` (login lewat Keycloak). Di development, set `JOBS_ENABLED=false` agar job pembersihan tidak menyentuh bucket dev.

## Tes

```
cd backend
go vet ./...
go test -p 1 ./... -count=1 -timeout 15m       # -p 1 wajib; butuh Docker menyala; beberapa menit
govulncheck ./...

# dari akar repo
node scripts/tests/web.test.cjs                # tes statis atas web/ (REBUILD=1 untuk build ulang dulu)
bun test scripts/tests/api.test.ts             # klien API dengan fetch palsu
npm run test:e2e                               # Playwright; server + Keycloak harus jalan (BASE_URL, KC_USER/KC_PASS, KC_USER2/KC_PASS2 untuk tes dua akun, KC_USER_TANPA_ROLE/KC_PASS_TANPA_ROLE)
```

Tes memakai `ai_center_test`, Redis DB 15, dan bucket `ai-center-test`; data dev tidak disentuh. Spec e2e admin memberi jeda sekitar 61 detik agar batas tulis admin tidak terlampaui. Tes alur pengguna dengan agen Claude ada di `scripts/e2e-agent/` (lihat README di sana).

Masalah umum di Windows + Docker Desktop:
- Tes kunci Redis (`TestAntiStampede`, `TestDuaInstanceBersamaanHanyaSatuBekerja`) gagal bila `.env` memakai `localhost`: pakai `127.0.0.1` seperti di `.env.example`.
- Setelah WSL atau Docker Desktop restart, port container bisa basi (server gagal dengan `EOF` ke Postgres/Redis): jalankan `docker compose restart`.

## Keamanan singkat

Session di Redis dengan cookie `HttpOnly`; metode ubah-data wajib `Origin` yang diizinkan dan header `X-CSRF-Token`; rate limit per IP/akun; CSP ketat tanpa `unsafe-inline` (skrip inline Astro diizinkan lewat hash); tipe gambar diperiksa dari isi file. Detail: `docs/security.md`.

## Dokumen

`docs/stack.md` (stack dan kontrak API), `docs/requirement.md`, `docs/database.md`, `docs/security.md`, `docs/design.md`, dan `docs/backend/` (catatan perubahan, rencana merge, skema). Aturan kerja untuk agen/kontributor ada di `CLAUDE.md`.
