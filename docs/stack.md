# Stack — AI Center UBAYA CMS

Dilebur dari `docs/_arsip/00-detail-rencana-web.md` dan `docs/_arsip/01-kontrak-api.md`.

## 1. Alur fase kerja

Semua akan diawali dengan proses planning terlebih dahulu, lalu masuk ke proses prototyping yg dibuat menggunakan HTML, CSS, JS saja. Jangan melakukan implementasi langsung dengan menggunakan list stack fix. List stack fix hanya akan diimplementasikan ketika proses prototyping telah aku anggap selesai

Kriteria selesai per fase: \[Lihat requirement.md §8\].

## 2. List Stack yang akan digunakan adalah sebagai berikut

**Frontend:**

* Astro \+ TypeScript: halaman publik. Mode render: statis; data diambil di browser lewat island (TanStack Query) dari API Go.  
* React \+ vite: dashboard admin (tempat penulisan dan pengelolaan konten).  
* Tailwind CSS versi terbaru: styling.  
* shadcn/ui: komponen UI.  
* GSAP dan motion.dev: animasi.  
* Embla-carousel: Library Carousel  
* TanStack Query: pengambilan data server dan cache di browser.  
* Bun: tooling dependency, build, dan pengujian frontend.

## 3. Backend

* Go \+ Chi (framework Go): API (`/api/*`), logika bisnis, penyajian hasil build Astro (`/`) dan build admin React (`/admin`) dalam satu origin. \[RP-01\]\[RP-02\] Fallback ke index.html hanya untuk `/admin/*` (SPA admin). \[RP-07\] Rute publik tanpa fallback (lihat aturan /berita/{slug} dan /dokumentasi/{slug} di bawah).  
  Rute /berita/{slug} dan /dokumentasi/{slug}: Go menyisipkan title, meta description, canonical, OG/Twitter, JSON-LD, isi plain text (di-escape), dan data awal JSON ke kerangka Astro. Slug tidak ditemukan dibalas HTTP 404 \+ noindex. Path tanpa file statis dibalas 404.html dengan status 404. Hasil di-cache di Redis per slug dan di-invalidate saat Post dan Retract. SSR tidak dipakai (K-03 tetap). \[RP-12\] sitemap.xml dan robots.txt dilayani Go (slug valid saja).  
  SEO per entitas dan aturan slug: \[Lihat database.md §3.3\].  
  URL canonical dan OG memakai env `PUBLIC_BASE_URL`. \[RP-08\]  
* PostgreSQL: database utama.  
* RustFS (kompatibel S3): penyimpanan gambar dan media.  
* Redis: cache server dan rate limiter.  
* Cookie session HttpOnly \+ CSRF: autentikasi dan perlindungan permintaan admin.  
* E2E: \[tool, mis. Playwright\] untuk alur publik dan admin end-to-end.

## 4. Status kode saat ini

**Status kode saat ini:** kode di repo (React+Vite, Bun+Elysia, PGlite, login password) bukan basis implementasi. Fase 3 dimulai dari proyek baru sesuai stack di atas; kode lama hanya referensi konten.

## 5. Rencana BACKEND

Penggantian bahasa untuk logika backend menjadi Go, penambahan layer API ditengah-tengah frontend dan backend yang ditulis menggunakan Chi (framework Go). Kode Go ditulis sebagai proyek baru di folder `/backend` dengan struktur folder default Go; `/server` lama hanya referensi.

## 6. Penerapan Redis

Key dipisah dengan prefix: `session:`, `lock:`, `ratelimit:`, `cache:`. Semua TTL dan batas disimpan sebagai konfigurasi (environment variable)

* Session login: \[Lihat security.md §2\]  
* Soft lock: \[Lihat database.md §4\]  
* Rate limiter: \[Lihat security.md §6\]  
* Cache server: \[Lihat database.md §7\]

## 7. Ketergantungan reverse proxy

Tag \[RP-xx\] menandai bagian rencana yang bergantung pada reverse proxy atau infrastruktur di depan aplikasi Go, agar mudah dicari (grep `\[RP-`) dan dibahas bersama pihak infra. Reverse proxy hanya menentukan ke mana request dikirim berdasarkan prefix path; aplikasi yang menentukan isinya. RP-01, RP-02, dan RP-09 harus jelas sebelum backend dimulai; sisanya paling lambat terjawab sebelum fase integrasi.

| ID | Bagian | Isu | Status |
|---|---|---|---|
| RP-01 | stack.md §3 | Tabel routing prefix: /api/\*, /admin/\*, /media/\*, sitemap.xml, robots.txt, /berita/{slug}, /dokumentasi/{slug}, dan sisanya ke file statis Astro. | SEBAGIAN: ditentukan di Chi (Go); /media/\* bergantung pada RP-11. |
| RP-02 | stack.md §3 | Apakah backend menerima /api/x atau /x setelah diteruskan. | SELESAI: prefix /api dipertahankan. |
| RP-03 | security.md §6 | IP asli klien: header yang diisi proxy (X-Forwarded-For / X-Real-IP) dan daftar IP proxy tepercaya, agar rate limiter membaca IP yang benar dan tidak bisa dipalsukan. | TERBUKA: bahas dengan infra. |
| RP-04 | security.md §6; database.md §7 | Lokasi cache: lapisan cache yang dipakai (proxy, Redis aplikasi, CDN) dan path yang tidak boleh ter-cache (/admin, login). | TERBUKA: bahas dengan infra. |
| RP-05 | security.md §6 | Rate limit proxy vs aplikasi: pembagian tugas limit kasar di proxy dan limit per user/akun di aplikasi, agar tidak dihitung dua kali. | TERBUKA: bahas dengan infra. |
| RP-06 | security.md §6 dan §7 | Respons 429 dan error: memastikan proxy tidak menimpa halaman error, status 429, dan header Retry-After. | TERBUKA: bahas dengan infra. |
| RP-07 | stack.md §3 | Fallback SPA: aturan fallback ke index.html agar refresh halaman tidak 404. Hanya relevan untuk SPA. | GUGUR: publik memakai Astro statis. Fallback hanya untuk /admin (ditangani Go). |
| RP-08 | stack.md §3 (`PUBLIC_BASE_URL`) | HTTPS dan header forward: siapa yang menghentikan TLS, dan bagaimana aplikasi tahu skema dan host asli (untuk canonical dan OG). | SEBAGIAN: pakai env PUBLIC_BASE_URL; detail TLS/header dibahas dengan infra. |
| RP-09 | security.md §2 | Origin, CORS, cookie: domain frontend, API, dan admin satu origin atau beda; memengaruhi CORS dan cookie session. | SELESAI: satu origin, tanpa CORS, cookie SameSite=Lax. Dev memakai proxy dev Vite/Astro ke Go. |
| RP-10 | database.md §3.1 dan §3.2 (batas 2 MB per file); timeout: \[belum ada bagian terkait\] | Upload dan timeout: batas ukuran file media dan lama request yang diizinkan (di proxy maupun Go). | TERBUKA: bahas dengan infra. |
| RP-11 | database.md §3.3 (OG image) dan §5.2 | File statis dan media: media/OG image dilayani lewat /media/\* di Go, atau bucket publik RustFS. OG image harus URL absolut yang bisa diakses crawler. | BELUM DIPUTUSKAN. |
| RP-12 | stack.md §3 (rute /berita/{slug} dan /dokumentasi/{slug}) | Injeksi meta SEO: penyisipan title, meta description, canonical, OG/Twitter, JSON-LD, dan data awal ke HTML kerangka; slug tidak ada dibalas 404. | DIPUTUSKAN: dikerjakan di backend Go (lihat N-03). |

## 8. Lokasi dokumen

Semua dokumen (.md) berada di /docs, sehingga folder dapat lebih rapi

## 9. Kontrak API

Acuan bersama frontend (mock) dan backend (implementasi). Keputusan bisnis ada di `requirement.md`; dokumen ini hanya bentuk request dan respons.

Status: kerangka. Bagian bertanda \[diisi\] dilengkapi sebelum fase frontend stack dimulai.

### 9.1 Konvensi umum

* Base path `/api`, format JSON (UTF-8), satu origin dengan frontend (tanpa CORS).  
* Endpoint publik memakai `slug`; endpoint admin memakai `id`.  
* Nama field, tipe, dan format tanggal: \[Lihat database.md §2\].

### 9.2 Format pagination

```json
{ "items": [], "page": 1, "limit": 9, "total": 0 }
```

### 9.3 Format error

```json
{ "error": { "code": "string", "message": "Pesan dalam sentence case" } }
```

Status yang dipakai: `400`, `401`, `403`, `404`, `409`, `429` (dengan header `Retry-After`). Tidak menampilkan detail internal.

### 9.4 Endpoint publik

Hanya mengembalikan data berstatus posted.

| Method | Path | Keterangan |
|---|---|---|
| GET | `/api/health` | Cek server dan koneksi database; 503 bila database mati |
| GET | `/api/berita` | Daftar berita; parameter `page`, `limit`, `q` (pencarian judul, minimal 2 karakter) |
| GET | `/api/berita/:slug` | Detail berita |
| GET | `/api/dokumentasi` | Daftar dokumentasi; parameter `page`, `limit`, `q` (pencarian judul, minimal 2 karakter) |
| GET | `/api/dokumentasi/:slug` | Detail dokumentasi |
| GET | `/api/testimoni` | Daftar testimoni |
| GET | `/api/tim` | Daftar tim |

Layanan (Pelatihan, Inference and Solutions) tidak punya endpoint; isinya hardcode di frontend.

### 9.5 Endpoint admin

Butuh session login (cookie) dan header `X-CSRF-Token` untuk POST/PUT/DELETE \[Lihat security.md §2 dan §3\]. `:sumber` \= `berita`, `dokumentasi`, atau `testimoni`. `PUT` membawa `version`; versi tidak cocok dibalas `409`.

| Method | Path | Aksi |
|---|---|---|
| GET | `/api/admin/:sumber` | Daftar semua status |
| GET | `/api/admin/:sumber/:id` | Detail, termasuk draft |
| POST | `/api/admin/:sumber` | Tambah (hasilnya draft) |
| PUT | `/api/admin/:sumber/:id` | Edit |
| POST | `/api/admin/:sumber/:id/post` | Post |
| POST | `/api/admin/:sumber/:id/retract` | Retract |
| DELETE | `/api/admin/:sumber/:id` | Hapus (ke trash) |
| POST | `/api/admin/:sumber/:id/restore` | Pulihkan |
| \[diisi\] | \[diisi\] | Hapus permanen (Berita, Dokumentasi, Testimoni) |
| GET, POST | `/api/admin/tim` | Daftar dan tambah Tim |
| GET, PUT, DELETE | `/api/admin/tim/:id` | Detail, edit, dan hapus permanen langsung Tim |
| \[diisi\] | \[diisi\] | Pasang lock, heartbeat, lepas lock (Berita, Dokumentasi, Testimoni, Tim) |
| \[diisi\] | \[diisi\] | Unggah gambar \[menunggu keputusan: RP-10\] |

Endpoint autentikasi \[menunggu konfirmasi Keycloak\]:

| Method | Path | Aksi |
|---|---|---|
| GET | `/api/auth/login` | Mulai alur OIDC (redirect ke Keycloak) |
| GET | `/api/auth/callback` | Terima kode OIDC, buat session |
| POST | `/api/auth/logout` | Hapus session |
| GET | `/api/auth/me` | Info pengguna (nama \+ role) |
| \[diisi\] | \[diisi\] | Login development (hanya `APP_ENV=development`) \[Lihat security.md §1\] |

### 9.6 Skema JSON per entitas

Field mengikuti rencana; nama camelCase mengikuti \[database.md §2\]; kolom pendukung di \[database.md §3\].

- **Berita:** id, slug, judul, isi (plain text, paragraf dipisah `\n\n`), gambar (1), tanggal publikasi, status, version.
- **Dokumentasi:** id, slug, judul kegiatan, gambar (1–6, berurutan; pertama = headline), tanggal publikasi, status, version.
- **Testimoni:** id, nama, teks, posisi, foto (opsional), status, version.
- **Tim:** id, nama, posisi / jabatan, bidang riset, foto, urutan, version.

Penulis berita ditampilkan statis oleh frontend dan tidak dikirim API \[Lihat database.md §3.1\].

Aturan pembuatan slug: \[terbuka\] \[Lihat database.md §3.3\].

Data awal JSON yang disisipkan Go ke HTML /berita/{slug} dan /dokumentasi/{slug} memakai bentuk yang sama dengan respons `GET /api/berita/:slug` dan `GET /api/dokumentasi/:slug`.

URL gambar di respons: \[menunggu keputusan: RP-11\].

### 9.7 Contoh mock

\[diisi: satu contoh respons per endpoint publik dan admin\]

### 9.8 Environment variable

Nama yang sudah dipakai: `APP_ENV`, `SESSION_COOKIE_NAME`, `SESSION_COOKIE_SECURE`, `DATABASE_URL`, `REDIS_URL`, `RUSTFS_ENDPOINT`, `RUSTFS_ACCESS_KEY`, `RUSTFS_SECRET_KEY`, `RUSTFS_BUCKET`, `ALLOWED_ORIGINS` (pemeriksaan header `Origin` untuk CSRF), `PUBLIC_BASE_URL`, `KEYCLOAK_ISSUER_URL`, `KEYCLOAK_CLIENT_ID`, `KEYCLOAK_CLIENT_SECRET`, `KEYCLOAK_REDIRECT_URI` \[menunggu konfirmasi Keycloak\].

Nama final \[diisi\] untuk daftar yang sudah disebut di rencana:

- TTL cache, TTL session (idle 1 jam, mutlak 8 jam), TTL lock (5 menit), interval heartbeat (1 menit).
- Batas rate limiter per bucket.
- Daftar IP proxy tepercaya.

### 9.9 Rujukan silang

* Rute SEO dan data awal: \[Lihat stack.md §3\]  
* Autentikasi, session, CSRF: \[Lihat security.md §1–§3\]  
* Rate limit dan `429`: \[Lihat security.md §6\]  
* Cache konten: \[Lihat database.md §7\]
