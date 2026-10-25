# Checklist Deploy — AI Center Ubaya

Langkah menyalakan aplikasi di server produksi. Semua nilai contoh = **DUMMY** —
isi dengan milik Anda sendiri, dan **jangan pernah** menaruh kredensial asli di
`.env.example` atau di repo.

## 1. Siapkan PostgreSQL

- Buat database (mis. `ai_center`) + user khusus dengan password kuat.
- Catat URL koneksi: `postgres://USER:PASSWORD@HOST:5432/ai_center`

## 2. Environment backend

Salin `.env.example` ke `.env`, lalu isi:

| Variabel | Contoh DUMMY | Keterangan |
| --- | --- | --- |
| `PORT` | `3000` | Port internal backend. |
| `DATABASE_URL` | `postgres://user:pass@db:5432/ai_center` | Wajib produksi; tanpa ini backend memakai PGlite. |
| `CORS_ORIGIN` | `https://ai-center.example.com` | Origin produksi. |
| `ADMIN_PASSWORD` | `ganti-dengan-password-kuat` | Wajib produksi. |
| `RUSTFS_ENDPOINT` | `http://rustfs:9000` | Endpoint S3 RustFS dari jaringan backend. |
| `RUSTFS_ACCESS_KEY` | `dummy-access-key` | Access key RustFS. |
| `RUSTFS_SECRET_KEY` | `dummy-secret-key` | Secret key RustFS. |
| `RUSTFS_BUCKET` | `ai-center-images` | Bucket gambar. |

Signing region SDK ditetapkan internal `us-east-1`. Jangan salin `.env`
backend ke image frontend.

## 3. Migrasi dan seed

```bash
cd server
bun run db:migrate
SEED_CONFIRM=RESET bun run db:seed
```

`db:seed` menghapus lalu mengisi ulang tabel. Gunakan hanya saat setup awal atau
reset development. Jangan jalankan pada database produksi berisi konten admin.

## 4. Jalankan container

Deployment root memakai:

- `deploy/backend.Dockerfile`;
- `deploy/frontend.Dockerfile`;
- `deploy/frontend.nginx.conf`;
- `deploy/nginx.conf`;
- `docker-compose.production.yml`.

```bash
docker compose -f docker-compose.production.yml config
docker compose -f docker-compose.production.yml up -d --build
```

Compose hanya mempublikasikan reverse proxy. Backend memakai jaringan internal,
dan tidak mempublikasikan port `3000` ke host.

Compose tidak menyediakan PostgreSQL, RustFS, atau TLS. Sediakan layanan tersebut
terpisah, lalu pastikan endpoint dapat dijangkau container backend.

Object gambar tersimpan di bucket RustFS. Persist volume RustFS dan jangan hapus
bucket saat deploy ulang.

## 5. URL aplikasi

- frontend: `https://ai-center.example.com/`;
- admin: `https://ai-center.example.com/admin`;
- API: `https://ai-center.example.com/api/*`;
- upload: `https://ai-center.example.com/uploads/*`.

Proxy meneruskan prefix `/api` apa adanya. Proxy tidak menambah `/api` kedua.
Rute selain `/api/*` dan `/uploads/*` memakai SPA fallback frontend.

Tanpa `VITE_API_BASE_URL`, frontend memakai `/api`. Override hanya diperlukan
untuk backend beda origin saat development.

## 6. Smoke test

- [ ] `GET /api/health` → `200` setelah skema database tersedia.
- [ ] Akses langsung `/` dan `/admin`.
- [ ] Login `/admin` dengan password produksi.
- [ ] CRUD dokumentasi, berita, kursus, tim, dan testimoni.
- [ ] Upload gambar lewat admin muncul dari `/uploads/...`.
