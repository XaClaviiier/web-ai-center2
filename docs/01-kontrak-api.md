# Referensi API

## 1. Cara Memakai Dokumen

Acuan bersama frontend dan backend. Perubahan endpoint atau bentuk data diubah di sini lebih dulu, lalu di kode. Setiap perubahan dicatat di bagian Riwayat Perubahan.

Kolom Keterangan: **ada** = sudah ada di kode `/server`; **baru** = belum ada di kode, harus dibuat; **ubah** = ada di kode, tetapi bentuknya berubah; **dihapus** = ada di kode, harus dihapus.

## 2. Konvensi Umum

- Base path `/api`, format JSON (UTF-8), satu origin dengan frontend (tanpa CORS).
- Endpoint publik: `/api/...`. Endpoint admin (wajib login): `/api/admin/...`.
- Endpoint publik memakai `slug`; endpoint admin memakai `id`.
- Pagination: `?page=1`, 12 item per halaman untuk berita dan dokumentasi. Pencarian: `?q=` (minimal 3 karakter), digabung dengan pagination.
- Waktu memakai format ISO 8601 (UTC).
- Gaya nama field JSON dan bentuk URL gambar (dari kode lama di `server/src/repositories/mappers.ts`; URL gambar diubah ke `/media`):
  - Nama field JSON: bahasa Indonesia, camelCase (`judul`, `ringkasan`, `diperbaruiPada`). Kolom DB snake_case dipetakan ke camelCase.
  - Kolom opsional yang NULL dihilangkan dari JSON (tidak dikirim sebagai `null`).
  - Daftar dibungkus `{ "items": [...] }`; detail berupa objek langsung; galat berupa `{ "error": "<pesan>" }`.
  - URL gambar: path relatif ke origin, `/media/<key>` (ubah; sebelumnya `/uploads/<nama>`). Database hanya menyimpan `key` object di RustFS. API mengembalikan URL hasil susunan `/media/<key>`.
  - `tanggal` berita/dokumentasi berformat `YYYY-MM-DD` (tanggal terbit pilihan editor, tidak bergeser oleh zona waktu). Timestamp sistem (`createdAt`, `updatedAt`, `postedAt`, `deletedAt`) berformat ISO 8601 UTC.
- Tanpa CORS: di produksi, reverse proxy menyajikan frontend dan meneruskan `/api` dan `/media` ke backend. Di development, server dev frontend (Astro untuk halaman publik, React + Vite untuk halaman admin) meneruskan `/api` dan `/media` ke backend lewat proxy. `@elysia/cors` dan `CORS_ORIGIN` dihapus; `ALLOWED_ORIGINS` hanya dipakai memeriksa header `Origin` untuk CSRF.
- Frontend: halaman publik memakai Astro, halaman admin memakai React + Vite.

## 3. Autentikasi

- Login lewat Keycloak (OIDC); backend membuat session. Cookie session HttpOnly, dan request POST/PUT/DELETE membawa header `X-CSRF-Token`.
- Session berupa ID acak yang disimpan di Redis (tanpa `SESSION_SECRET`).
- Tidak ada pembagian role: semua pengguna yang terdaftar di Keycloak adalah admin. Prasyarat (urusan pengelola Keycloak): pendaftaran mandiri (self-registration) dimatikan.
- **Masa transisi:** Keycloak belum tersedia. Sampai siap, auth memakai mekanisme yang ada: `POST /api/admin/login` (password dari `ADMIN_PASSWORD`) lalu `Authorization: Bearer <token>`. Setelah Keycloak jadi, diganti bertahap dan login password dihapus. Selama transisi `X-CSRF-Token` tidak berlaku untuk Bearer.
- Endpoint auth (login, callback, logout, identitas pengguna):

| Method | Path | Fungsi | Keterangan |
| ------ | ---- | ------ | ---------- |
| POST | `/api/admin/login` | Login satu akun: body `{ "password" }`, balikan `{ "token" }`. Rate-limit 5 gagal / 10 menit per IP (429 + `Retry-After`; in-memory, bukan Redis). | ada (sementara; dihapus setelah Keycloak) |
| GET | `/api/auth/login` | Mulai alur OIDC: redirect ke Keycloak. | baru (rencana, menunggu Keycloak) |
| GET | `/api/auth/callback` | Terima kode OIDC, buat session di Redis, set cookie. | baru (rencana) |
| POST | `/api/auth/logout` | Hapus session. | baru (rencana) |
| GET | `/api/auth/me` | Identitas pengguna yang sedang login. | baru (rencana) |

## 4. Model Data

Sumber: `server/src/db/schema.ts` (PostgreSQL, Drizzle). Kolom `id` = `serial`. Kolom `jsonb` ditandai. Kolom/tabel **baru** menurut keputusan sesi desain (rincian akhir mengikuti plan.md, lihat bagian Pertanyaan Terbuka).

**berita**: `id`, `slug` (unik, dibuat sekali dari judul), `judul`, `ringkasan`, `isi`, `tanggal` (date), `gambar` (null; berisi key), ~~`penulis`~~ (dihapus), `created_at`.
- Kolom baru: `status` (`draft`/`posted`), `version` (naik 1 tiap simpan, Post, dan Retract), `updated_at`, `posted_at` (null), `deleted_at` (null; soft delete).

**dokumentasi**: `id`, `slug` (unik, dibuat sekali dari judul), `judul`, `deskripsi`, `tanggal` (date), `kategori`, `created_at`.
- Kolom `gambar` lama **dihapus**; datanya dimigrasi ke tabel `dokumentasi_gambar`.
- Kolom baru: `status`, `version`, `updated_at`, `posted_at`, `deleted_at` (sama seperti berita).

**dokumentasi_gambar** (tabel baru): `id`, `dokumentasi_id` (FK ke dokumentasi), `key` (object key RustFS), `urutan`, `keterangan` (null). Gambar dengan `urutan` terkecil dipakai sebagai sampul.

**tim**: `id`, `nama`, `peran`, `kredensial` (null), `foto` (null; key), `urutan`. CRUD biasa, tanpa status.

**testimoni**: `id`, `nama`, `peran`, `kutipan`, `urutan`. CRUD biasa, tanpa status.

**Dihapus total:** tabel `kursus` dan tabel `progress_pembelajaran`.

**Di luar kontrak** (ada di skema, dikelola frontend, tidak punya endpoint): `layanan`, `hero_slides`, `klien`, `profil`, `inference` (legacy).

**Redis** (bukan tabel): session store dan cache konten. Cache mencakup GET publik berita, dokumentasi, tim, dan testimoni. Cache dihapus per sumber setiap ada tulis (edit, Post, Retract, hapus, restore), dengan TTL 5 menit sebagai pengaman. Redis tidak dipakai untuk rate-limit login.

## 5. Endpoint Publik

Tanpa autentikasi. Berita dan dokumentasi hanya mengembalikan konten berstatus `posted` dan belum terhapus (draft/terhapus: 404 lewat slug, tidak muncul di daftar).

| Method | Path | Fungsi | Keterangan |
| ------ | ---- | ------ | ---------- |
| GET | `/api/health` | Cek server dan koneksi DB; 503 bila DB mati. | ada |
| GET | `/api/berita` | Daftar berita `posted`. Query: `q`, `page`, `order` (`asc`/`desc`, default `desc`). | ubah (12 per halaman, `q` min. 3 karakter, hanya `posted`) |
| GET | `/api/berita/:slug` | Detail berita `posted`; 404 bila tidak ada. | ubah (hanya `posted`) |
| GET | `/api/dokumentasi` | Daftar dokumentasi `posted`, dengan gambar sampul. Query: `q`, `page`, `order`, `kategori`. | ubah (12 per halaman, `q` min. 3 karakter, hanya `posted`) |
| GET | `/api/dokumentasi/:slug` | Detail dokumentasi `posted` beserta semua gambarnya (berurutan). | ubah (hanya `posted`, daftar gambar) |
| GET | `/api/tim` | Daftar anggota tim. | ada |
| GET | `/api/testimoni` | Daftar testimoni. | ada |
| GET | `/media/<key>` | Gambar dari bucket RustFS (baca-publik), diteruskan proxy. Bukan route backend. | baru (menggantikan `/uploads/:nama`) |
| GET | `/api/kursus`, `/api/kursus/:kode` | Katalog kursus. | dihapus |
| GET, POST | `/api/progress` | Progres belajar peserta. | dihapus |
| GET | `/uploads/:nama` | Penyaji file unggahan. | dihapus |

## 6. Endpoint Admin

Memakai Bearer token (masa transisi) atau session Keycloak. Tanpa login: 401. `:sumber` = `berita` atau `dokumentasi`. `PUT` wajib membawa `version`; 409 bila `version` sudah berubah. Post, Retract, dan simpan menaikkan `version` 1. Retract dan restore mengembalikan konten ke `draft`. Slug dibuat sekali dari judul saat pertama disimpan, lalu tidak berubah.

| Method | Path | Fungsi | Keterangan |
| ------ | ---- | ------ | ---------- |
| POST | `/api/admin/login` | Login sementara (bagian 3). | ada (sementara) |
| POST | `/api/admin/upload` | Unggah gambar (multipart field `gambar`) ke RustFS; balikan `{ "key", "url" }` (201). `key` disimpan saat Simpan, `url` (`/media/<key>`) hanya untuk pratinjau. | ubah (sebelumnya `{ url: "/uploads/..." }`) |
| POST | `/api/admin/uploads/bersihkan` | Hapus object RustFS yang tidak direferensikan (membaca berita, tim, dan `dokumentasi_gambar`). Body `{ "kering": true }` = hanya laporan. | ada (ubah: ikut `dokumentasi_gambar`) |
| GET | `/api/admin/:sumber` | Daftar semua status (draft, posted, terhapus). Query `status`, `q`, `page`. | baru |
| GET | `/api/admin/:sumber/:id` | Detail by `id`, termasuk draft. | baru |
| POST | `/api/admin/:sumber` | Buat sebagai draft; slug dari judul; 409 bila bentrok. | baru (menggantikan `POST /api/berita`, `/api/dokumentasi`) |
| PUT | `/api/admin/:sumber/:id` | Ubah konten; body membawa `version`. | baru (menggantikan `PUT /api/<sumber>/:slug`) |
| POST | `/api/admin/:sumber/:id/post` | Post: draft menjadi tampil publik. | baru |
| POST | `/api/admin/:sumber/:id/retract` | Retract: tarik kembali ke draft. | baru |
| DELETE | `/api/admin/:sumber/:id` | Hapus (soft delete). | baru (menggantikan `DELETE /api/<sumber>/:slug`, yang permanen) |
| POST | `/api/admin/:sumber/:id/restore` | Pulihkan yang terhapus; kembali ke `draft`. | baru |
| POST | `/api/admin/tim` | Tambah anggota tim (paling atas); 201. | ubah (sebelumnya `/api/tim`) |
| PUT | `/api/admin/tim/:id` | Ubah anggota tim. | ubah (sebelumnya `/api/tim/:id`) |
| DELETE | `/api/admin/tim/:id` | Hapus anggota tim. | ubah (sebelumnya `/api/tim/:id`) |
| POST | `/api/admin/testimoni` | Tambah testimoni (paling atas); 201. | ubah (sebelumnya `/api/testimoni`) |
| PUT | `/api/admin/testimoni/:id` | Ubah testimoni. | ubah (sebelumnya `/api/testimoni/:id`) |
| DELETE | `/api/admin/testimoni/:id` | Hapus testimoni. | ubah (sebelumnya `/api/testimoni/:id`) |
| POST, PUT, DELETE | `/api/kursus`, `/api/kursus/:kode` | CRUD kursus. | dihapus |
| POST, PUT, DELETE | `/api/berita`, `/api/berita/:slug`, `/api/dokumentasi`, `/api/dokumentasi/:slug` | CRUD lama (path publik). | dihapus (diganti `/api/admin/:sumber...`) |

## 7. Environment Variable

- Sudah diputuskan: `APP_ENV`, `SESSION_COOKIE_NAME`, `SESSION_COOKIE_SECURE`. Cookie `__Host-` hanya aktif di produksi.
- Dari kode saat ini (`server/src/config.ts`, `server/src/index.ts`) dan keputusan terbaru; contoh di `server/.env.example`:

| Variabel | Fungsi | Keterangan |
| -------- | ------ | ---------- |
| `PORT` | Port HTTP backend (default `3000`) | ada |
| `DATABASE_URL` | Koneksi PostgreSQL (kosong = PGlite lokal) | ada |
| `ALLOWED_ORIGINS` | Daftar origin (koma) untuk pemeriksaan header `Origin` pada CSRF, bukan CORS | baru (menggantikan `CORS_ORIGIN`) |
| `REDIS_URL` | Koneksi Redis (session + cache konten) | baru |
| `KEYCLOAK_ISSUER_URL` | URL issuer OIDC Keycloak | baru (menunggu Keycloak) |
| `KEYCLOAK_CLIENT_ID` | Client ID | baru (menunggu Keycloak) |
| `KEYCLOAK_CLIENT_SECRET` | Client secret | baru (menunggu Keycloak) |
| `KEYCLOAK_REDIRECT_URI` | Redirect URI callback (`/api/auth/callback`) | baru (menunggu Keycloak) |
| `SESSION_TTL_SECONDS` | Masa berlaku session di Redis | baru |
| `APP_ENV`, `SESSION_COOKIE_NAME`, `SESSION_COOKIE_SECURE` | Lihat butir keputusan di atas | diputuskan, belum dibaca kode |
| `RUSTFS_ENDPOINT` | Endpoint S3 RustFS (default `http://localhost:9000`) | ada |
| `RUSTFS_ACCESS_KEY`, `RUSTFS_SECRET_KEY` | Kredensial RustFS (wajib saat runtime) | ada |
| `RUSTFS_BUCKET` | Bucket gambar (default `ai-center-images`; baca-publik untuk `/media`) | ada |
| `ADMIN_PASSWORD` | Password login admin sementara (fallback `admin-dev`; di `NODE_ENV=production` server berhenti bila tidak diisi) | ada (sementara, sampai Keycloak jadi) |
| `NODE_ENV` | Mode produksi | ada |
| `CORS_ORIGIN` | Whitelist CORS | dihapus (diganti `ALLOWED_ORIGINS`) |
| `UPLOAD_DIR` | Direktori fallback filesystem untuk test | dihapus bersama `/uploads` |

- Frontend: `VITE_API_BASE_URL` (default `/api`; dengan proxy dev cukup default) dan `VITE_SITE_URL`.
- `.env.example` disesuaikan saat kode diubah, bukan di dokumen ini.

## 8. Riwayat Perubahan

| Tanggal | Perubahan | Oleh |
| ------- | --------- | ---- |
| 2026-11-04 | Isi bagian [AGENT] dari kode `/server`. plan.md tidak ditemukan, item terkait ditandai [TBD]. | Claude |
| 2026-11-04 | Terapkan keputusan sesi desain: kursus dan progres dihapus, status/version/soft delete berita dan dokumentasi, `dokumentasi_gambar`, path admin `/api/admin/...`, tanpa CORS, gambar lewat `/media`, Redis, env baru, auth transisi. | Claude |

## Pertanyaan Terbuka [TBD]

1. **plan.md** sedang dibuat. Setelah jadi, cocokkan bagian 4 dan 6 dengannya: bentuk body/respons draft, Post, Retract, restore, tipe kolom `version`, dan usulan path.
2. **Nilai `APP_ENV`, `SESSION_COOKIE_NAME`, `SESSION_COOKIE_SECURE`** diputuskan tetapi nilai dan perilakunya belum tertulis di kode.
3. **Keycloak** (realm, client, redirect URI, pendaftaran mandiri dimatikan) menunggu pengelola Keycloak. Bagian 3 dan 7 berlaku penuh setelah itu.
4. **Astro** (halaman publik): apakah dirender statis atau SSR memengaruhi pemakaian cache Redis; belum diputuskan.
5. **Migrasi data**: skrip memindahkan `dokumentasi.gambar` ke `dokumentasi_gambar`, mengubah URL `/uploads/...` menjadi key, dan menghapus data kursus/progres belum ditulis.
