# req.md: Stack dan Library Backend (Go + Chi)

Daftar stack, library, dan versi untuk backend baru di folder `/backend`. Rujukan arsitektur: `docs/stack.md`, `docs/database.md`, `docs/security.md`. Dokumen ini hanya mencatat pilihan teknis dan versinya, tidak mengubah keputusan di `docs/`.

Tanggal verifikasi versi: 2026-11-04. Versi dicek lewat `go list -m` (proxy Go) dan kerentanan dicek dengan `govulncheck`.

## 1. Toolchain

| Komponen | Versi | Catatan |
| -------- | ----- | ------- |
| Go | **1.27.4** (toolchain efektif untuk modul ini); instalasi sistem 1.27.3 | Instalasi sistem terpasang via `winget install GoLang.Go` (1.27.3). Sejak 2026-11-11 `go.mod` memakai `go 1.27.4` karena GO-2026-6859 (lihat §5), sehingga `go` mengunduh toolchain 1.27.4 otomatis (`GOTOOLCHAIN=auto`, diverifikasi `sum.golang.org`) tanpa mengubah instalasi sistem. Pasang 1.27.4 di sistem (`winget upgrade GoLang.Go`) setelah `winget` memuatnya. |
| govulncheck | v1.9.0 | Pemindai kerentanan resmi tim Go (`golang.org/x/vuln`). Jalankan `govulncheck ./...` sebelum rilis dan setiap menaikkan versi dependensi. |
| sqlc (alat dev, bukan dependensi runtime) | **v1.31.1** | Membangkitkan kode Go dari SQL. Terpasang dari binari rilis resmi (`go install` gagal: `The paging file is too small`). Lokasi: `C:\Users\KMira\go\bin\sqlc.exe` (`$(go env GOPATH)\bin`). Lihat bukti checksum di §1.1. |

### 1.1 Bukti checksum sqlc

- Sumber: `https://github.com/sqlc-dev/sqlc/releases/download/v1.31.1/sqlc_1.31.1_windows_amd64.zip` (15.284.588 byte).
- SHA256 zip yang diharapkan (digest aset dari GitHub Releases API, rilis tidak menyertakan file checksum terpisah): `352711fa7dcb05dcdfefca0ad71b2c9a74fd090f8d7fc609419de4cbc725429f`.
- SHA256 zip hasil unduhan: `352711fa7dcb05dcdfefca0ad71b2c9a74fd090f8d7fc609419de4cbc725429f` (**cocok**).
- SHA256 `sqlc.exe` hasil ekstrak: `22ce316349ac153a4aacaa3912981f73980a6f84861f4c602ae6e3075ca326e4`.
- `sqlc version` mengembalikan `v1.31.1`.
- Pemasangan ulang di mesin lain: unduh zip yang sama, cocokkan hash zip dengan nilai di atas sebelum mengekstrak.

## 2. Library runtime

| Fungsi | Library | Versi | Alasan dan reputasi |
| ------ | ------- | ----- | ------------------- |
| Router HTTP | `github.com/go-chi/chi/v5` | **v5.5.0** | Dipilih di `docs/stack.md`. Kompatibel `net/http` standar, tanpa dependensi luar, sangat luas dipakai. Memakai paket `middleware` bawaan (RequestID, RealIP perlu dibatasi ke proxy tepercaya, Recoverer, Timeout). |
| PostgreSQL | `github.com/jackc/pgx/v5` (+ `pgxpool`) | **v5.12.0** | Driver PostgreSQL de facto untuk Go, dirawat aktif, query berparameter secara bawaan (selaras `security.md` §7). |
| Redis | `github.com/redis/go-redis/v9` | **v9.23.0** | Klien resmi organisasi Redis. Dipakai untuk session, soft lock (`SET NX`), rate limiter (`INCR` + `EXPIRE`), cache. |
| S3 / RustFS | `github.com/aws/aws-sdk-go-v2` | **v1.47.5** | SDK resmi AWS. |
| | `.../aws-sdk-go-v2/service/s3` | **v1.114.6** | Klien S3 untuk RustFS (endpoint kustom, path-style, region ditetapkan `us-east-1` seperti backend lama). |
| | `.../aws-sdk-go-v2/config` | v1.33.13 | Pemuatan konfigurasi. |
| | `.../aws-sdk-go-v2/credentials` | v1.20.13 | Kredensial statis dari env. |
| OIDC (Keycloak) | `github.com/coreos/go-oidc/v3` | **v3.21.0** | Verifikasi ID token (penerbit, tanda tangan, masa berlaku). Dipakai luas (mis. ekosistem Kubernetes). Tarik `go-jose/v4` v4.1.4. Baru dipakai setelah Keycloak siap. |
| OAuth2 + PKCE | `golang.org/x/oauth2` | **v0.37.0** | Library resmi tim Go; mendukung Authorization Code + PKCE (S256). |
| Anti-stampede cache | `golang.org/x/sync` (`singleflight`) | **v0.24.0** | Library resmi tim Go. Sebelumnya hanya dependensi tidak langsung; kini langsung dipakai `internal/cache` agar banyak request bersamaan pada key cache dingin hanya membaca database sekali. `govulncheck`: No vulnerabilities found. Tidak ada masalah kompatibilitas (go 1.27.4).
| Migrasi skema | `github.com/pressly/goose/v3` | **v3.29.0** | Migrasi SQL berurutan, bisa dipakai sebagai library maupun CLI. |

Dependensi transitif yang ikut terbawa (hasil `go mod tidy` pada modul uji): `golang.org/x/crypto` v0.58.0, `golang.org/x/net` v0.61.0, `golang.org/x/text` v0.43.0. Versinya dikunci oleh `go.sum`.

## 3. Standard library yang dipakai (tanpa dependensi tambahan)

- `log/slog`: log terstruktur (detail error hanya ke log, bukan ke klien).
- `crypto/rand`: ID session acak 256 bit dan token CSRF; `crypto/subtle`: perbandingan password/token waktu-konstan (login development).
- `net/http` (`DetectContentType`): cek tipe gambar dari isi file (jpg/png/webp), ukuran dibatasi 2 MB.
- `encoding/json`, `net/http/httptest`, `testing`: serialisasi dan pengujian.
- `os` (`Getenv`): konfigurasi dari environment variable; validasi dan nilai default ditulis sendiri di paket `config`.

Sengaja tidak dipakai: ORM (GORM dan sejenisnya), framework web selain Chi, library validasi pihak ketiga, library testing pihak ketiga. Alasannya: surface dependensi kecil, query SQL eksplisit dan berparameter, sesuai prinsip keamanan di `docs/security.md`.

## 4. Layanan eksternal (sudah berjalan di Docker milik pengguna)

| Layanan | Image | Port | Catatan |
| ------- | ----- | ---- | ------- |
| PostgreSQL | `postgres:15` | 5432 | Database utama. |
| Redis | `redis:7-alpine` | 6379 | Session, lock, rate limiter, cache (prefix `session:`, `lock:`, `ratelimit:`, `cache:`). |
| RustFS | `rustfs/rustfs:latest` | 9000 (S3), 9001 (konsol) | Penyimpanan gambar. Tag `latest` disarankan dikunci ke versi tertentu agar build ulang tidak berubah diam-diam. |
| Keycloak | (belum ada) | | Menunggu pihak pengelola. Sementara memakai login development (`APP_ENV=development`, lihat `docs/security.md` §1). |

Compose dikelola pengguna sendiri; backend tidak membawa compose.

### 4.1 Pengujian integrasi (database, Redis)

Tes memakai layanan Docker di atas, tetapi data dev tidak pernah disentuh:

| Layanan | Yang dipakai tes | Pengaman |
| ------- | ---------------- | -------- |
| PostgreSQL | database `ai_center_test` (`TEST_DATABASE_URL`) | tabel konten di-`TRUNCATE` sebelum dan sesudah tes; `ai_center` dan `mydb` tidak disentuh |
| Redis | database index **15** (`TEST_REDIS_URL`) | sebelum `FLUSHDB`, `testutil.RedisTes` memastikan URL menunjuk DB 15; selain itu tes gagal. DB 0 (dev) tidak pernah dikosongkan |
| RustFS | bucket `ai-center-test` (`TEST_RUSTFS_BUCKET`) | dipakai tes `media` dan tes ujung-ke-ujung upload; dibuat bila belum ada dan dikosongkan sebelum dan sesudah tes. `testutil.RustFSTes` menggagalkan tes bila bucket tes sama dengan `RUSTFS_BUCKET` |

- Tes dilewati otomatis (SKIP) bila `TEST_DATABASE_URL`/`TEST_REDIS_URL` kosong; bila diisi tetapi layanan mati, tes GAGAL.
- Paket yang memakai database/Redis dijalankan berurutan: `go test -p 1 ./... -count=1`.
- **Lama dan batas waktu (diperbarui 2026-11-11):** suite penuh kini berjalan beberapa menit (paket `httpapi` sekitar 160 detik karena tes ujung-ke-ujung memakai banyak operasi RustFS, `store` sekitar 90-135 detik). Pakai `-timeout 15m` untuk suite penuh (batas bawaan Go 10 menit pernah terlampaui), dan jalankan lewat latar belakang atau dengan keluaran ke berkas. Tes ujung-ke-ujung memakai batas konteks 300 detik.
- **Catatan lingkungan:** pernah terjadi kelambatan tidak terduga (dua tes `store` gagal `context deadline exceeded`, satu run `store` melewati 10 menit) yang hilang dengan sendirinya dan tidak berkaitan dengan perubahan kode; penyebab belum diketahui (dugaan: kelambatan Docker/WSL). Bila terulang, ambil `pg_stat_activity` dan `pg_locks` selagi macet sebelum menjalankan ulang.
- **Tes cache dan limiter publik (2026-11-11):** `internal/cache` dan `httpapi/publik_cache_test.go` memakai Redis DB 15 nyata (hit tanpa menyentuh database/limiter, 304 tidak menambah hitungan, anti-stampede 20 request bersamaan = 1 pembacaan, 429 + `Retry-After`, Redis mati = situs publik tetap 200, pembatalan cache admin) serta satu tes ujung-ke-ujung dengan PostgreSQL tes (Retract langsung 404).
- Handler diuji dengan implementasi palsu (tanpa layanan), sedangkan `store`, `auth` (RedisStore), `ratelimit`, dan satu tes ujung-ke-ujung berita memakai layanan nyata.

## 5. Hasil pemeriksaan keamanan dependensi

- `govulncheck ./...` pada modul uji yang mengimpor semua library runtime di atas: **No vulnerabilities found** (database vuln.go.dev, diperbarui 2026-11-02, Go 1.27.3).
- Pembaruan 2026-11-09: `govulncheck ./...` pada modul `backend` yang sebenarnya: **No vulnerabilities found**. Dependensi langsung terpasang: chi v5.5.0, pgx v5.12.0, go-redis v9.23.0, aws-sdk-go-v2 v1.47.5, aws-sdk-go-v2/service/s3 v1.114.6, aws-sdk-go-v2/credentials v1.20.13 (tarik smithy-go v1.28.7). Library `config` AWS tidak dipakai (konfigurasi dibuat langsung). goose tetap dijalankan sebagai alat CLI, go-oidc dan oauth2 baru ditambahkan ke `go.mod` saat Keycloak dikerjakan.
- Pembaruan 2026-11-09 (CRUD admin berita): ditambah dependensi langsung **`golang.org/x/text` v0.43.0** (paket `unicode/norm`, untuk membuang aksen saat membuat slug dari judul, mis. "Café" menjadi "cafe"). Alasan: pustaka resmi tim Go, sebelumnya sudah ada sebagai dependensi tidak langsung (v0.41.0) dan dinaikkan ke v0.43.0, menaikkan juga `golang.org/x/sync` ke v0.24.0. `govulncheck ./...` setelah penambahan: **No vulnerabilities found**. Tidak ada library baru lain pada langkah ini.
- **Temuan 2026-11-11: GO-2026-6859** (`net/http` di pustaka standar, "HTTP/1 server crash due to MaxBytesReader header race"), ditemukan ulang oleh `govulncheck` setelah database kerentanan diperbarui; ditemukan di Go 1.27.3 dan diperbaiki di **Go 1.27.4**. Relevan karena kode memakai `http.MaxBytesReader` (upload, body JSON admin, login). Hasil `govulncheck` bersih yang dicatat di atas (2026-11-02 sampai 2026-11-09) tidak lagi berlaku untuk toolchain 1.27.3. Perbaikan: `go 1.27.4` di `go.mod` (toolchain otomatis, tanpa menyentuh instalasi sistem). Verifikasi: `govulncheck ./...` = No vulnerabilities found; seluruh tes (242) lulus dengan toolchain 1.27.4. Aturan perawatan tambahan: jalankan `govulncheck ./...` di awal setiap sesi kerja, bukan hanya saat menambah dependensi, karena advisori baru bisa terbit kapan saja.
- Catatan kompatibilitas RustFS: klien S3 memakai path-style, region `us-east-1`, dan checksum permintaan/respons SDK hanya bila diwajibkan (`RequestChecksumCalculation`/`ResponseChecksumValidation` = `WhenRequired`); tanpa itu SDK baru menambahkan header checksum yang tidak selalu didukung server S3 non-AWS.
- Semua versi di atas adalah rilis terbaru yang stabil pada tanggal verifikasi (tanpa pra-rilis). Catatan: proxy Go menunjukkan `pgx` v5.12.0 dirilis 2026-10-31, sangat baru. Bila ada masalah kompatibilitas, turunkan ke v5.11.x lalu cek lagi dengan `govulncheck`.
- Aturan perawatan: versi dikunci lewat `go.mod`/`go.sum`; naikkan versi hanya dengan `go get <modul>@<versi>`, jalankan ulang `govulncheck ./...` dan tes, lalu perbarui dokumen ini.

## 6. Hal yang belum diputuskan

- Struktur folder Go sudah diputuskan: modul `github.com/XaClaviiier/web-ai-center2/backend`, `cmd/api`, `internal/{config,httpapi,store,konten,auth,ratelimit,testutil}`.
- Paket gambar (resize/optimasi) tidak dipakai; bila `docs/` kelak mensyaratkannya, tambahkan di sini.

## 7. Alat uji e2e dan audit keamanan (2026-11-12)

- **`@playwright/test` 1.65.0** (devDependency root, versi tepat tanpa `^`; Chromium diunduh ke cache pengguna `%LOCALAPPDATA%\ms-playwright`, bukan ke repo). Alasan: spesifikasi e2e browser nyata di `scripts/e2e/*.spec.cjs` (`npm run test:e2e`; butuh server dev berjalan, variabel `BASE_URL`, `PW`, `PW2`). Reputasi: runner resmi Microsoft Playwright. `npm audit`: **0 vulnerabilities**. Kompatibilitas: berjalan di Node 22.19.0; spesifikasi memakai CommonJS (`.cjs`) agar sesuai `scripts/`. Strategi tes: spesifikasi memakai worker tunggal (batas tulis admin 30/menit per akun), data uji berlabel `[E2E-PW]` yang dibuat dan dibersihkan sendiri.
- **Skill agen (bukan dependensi proyek, terpasang global di `~/.claude/skills`):** `security-audit` dari `github.com/cloudflare/security-audit-skill` (audit statis; keluaran di luar repo, `~/security-audit-skill/web-ai-center2/run-1/`) dan `playwright` dari `github.com/openai/skills` (hanya dipakai sebagai panduan alur; tidak ada kode dari skill yang masuk repo).

## 8. Alat build frontend-final (2026-11-13)

- **Bun 1.4.3** (terpasang global di mesin pengembang; `packageManager` frontend meminta 1.4.4, hanya peringatan engines). Dipakai `scripts/bangun-web-final.cjs` untuk `bun install --frozen-lockfile` dan `bun --bun run build` di salinan kerja `.bangun/frontend-final/` (ter-ignore git). `--bun` diperlukan karena Node 22.19.0 di mesin ini ditolak Astro 7 (butuh >=22.23.4); Bun melaporkan versi Node yang memenuhi. Variabel `BUN_BIN` mengganti path biner (nama `BUN` tidak dipakai karena sudah berisi direktori instalasi Bun).
- **Dependensi `frontend-final/`** (dikunci `frontend-final/bun.lock`, dipilih tim frontend; tidak diubah): astro 7.4.11, @astrojs/react 7.1.8, react/react-dom 19.3.1, @tanstack/react-query 5.107.1, gsap 3.15.1, motion 14.13.0, embla-carousel-react/autoplay 8.6.0, @radix-ui/react-slot 1.5.0 dan react-alert-dialog 1.2.1, class-variance-authority 0.7.1, clsx 2.1.1, tailwind-merge 3.7.0; dev: vite 8.3.8, @vitejs/plugin-react 6.3.0, tailwindcss + @tailwindcss/vite 4.3.5, typescript 6.0.3, @astrojs/check 0.9.12, @playwright/test 1.65.0. Hasil `bun audit` pada 399 paket: **No vulnerabilities found** (2026-11-13).
- Perubahan strategi tes: `scripts/e2e/*.spec.cjs` ditulis untuk DOM prototipe lama dan BELUM disesuaikan ke DOM frontend-final (langkah 7 rencana). Verifikasi sementara memakai smoke test browser sekali pakai (tidak masuk repo).