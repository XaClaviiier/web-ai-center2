# Panduan Deploy Production

Panduan ringkas menjalankan CMS di production. Keycloak production dibahas di `keycloak/README.md` bagian 2.

## Arsitektur

```
Internet --443--> reverse proxy (TLS) --http--> backend CMS (image ini, :8080)
                        |                          |-- PostgreSQL 15   (jaringan internal)
                        |                          |-- Redis 7         (jaringan internal)
                        |                          '-- RustFS          (jaringan internal)
                        '--443--> Keycloak (/realms/ai-center/ dan /resources/ saja)
```

Hanya port 443 di reverse proxy yang terbuka ke internet. PostgreSQL, Redis, RustFS, dan konsol admin Keycloak tidak boleh dipublikasikan. `docker-compose.yml` di akar repo hanya untuk development: ia mempublikasikan port database dengan password dev.

## 1. Bangun image

Dari akar repo, tandai dengan SHA commit agar setiap rilis dapat dilacak dan di-rollback:

```bash
docker build -f backend/Dockerfile -t ai-center-cms:$(git rev-parse --short HEAD) .
```

Image berisi binary backend, folder `web/` (sudah dibangun dan di-commit), migrasi, dan `goose`. Image berjalan sebagai user non-root dan default `APP_ENV=production`. Base image dipin ke digest.

## 2. Siapkan sebelum rilis pertama

- **Domain dan sertifikat TLS** di reverse proxy (Caddy atau nginx; Let's Encrypt).
- **Rahasia baru khusus production.** Jangan memakai nilai dev dari `.env.example` atau `docker-compose.yml`. Simpan sebagai berkas (Docker secrets atau secret manager) dan rujuk lewat variabel `*_FILE`:
  - client secret Keycloak (`KEYCLOAK_CLIENT_SECRET_FILE`);
  - kunci session tepat 32 byte acak (`SESSION_ENCRYPTION_KEY_FILE`);
  - password PostgreSQL, Redis, dan kunci RustFS.
- **Bucket RustFS** dibuat lebih dulu (`RUSTFS_BUCKET`); production tidak membuatnya otomatis.
- **Role database** dengan hak minimum untuk aplikasi (lihat `docs/database.md`).
- **Backup** otomatis untuk PostgreSQL CMS, database Keycloak, dan bucket RustFS. Uji pemulihan sekali sebelum go-live.

## 3. Variabel environment wajib

| Variabel | Nilai production |
|---|---|
| `APP_ENV` | `production` (default image) |
| `PUBLIC_BASE_URL` | Alamat publik `https://...`; server menolak start bila `http://` |
| `ALLOWED_ORIGINS` | Origin `https://` yang boleh menulis (minimal `PUBLIC_BASE_URL`) |
| `SESSION_COOKIE_NAME` / `SESSION_COOKIE_SECURE` | `__Host-session` / `true` |
| `KEYCLOAK_ISSUER` | `https://<host-keycloak>/realms/ai-center` |
| `TRUSTED_PROXIES` | IP atau CIDR reverse proxy. Bila kosong, semua klien terlihat sebagai IP proxy dan berbagi satu batas laju |
| `DATABASE_URL`, `REDIS_URL`, `RUSTFS_*` | Alamat di jaringan internal |
| `JOBS_ENABLED` / `JOBS_DRY_RUN` | `true` / `true` pada hari pertama (lihat langkah 6) |

Daftar lengkap dan nilai default ada di `backend/.env.example`. Server memeriksa konfigurasi saat start dan menolak berjalan dengan pesan jelas bila ada yang salah.

## 4. Migrasi database

Jalankan sebagai langkah terpisah **sebelum** versi baru menerima trafik, memakai image yang sama:

```bash
docker run --rm --network <jaringan-internal> --entrypoint /app/goose ai-center-cms:<sha> \
  -dir /app/migrations postgres "<DATABASE_URL>" up
```

Tulis migrasi agar kompatibel dengan versi sebelumnya (tambah dulu, hapus di rilis berikutnya), supaya rollback image tetap aman.

## 5. Rilis

1. Jalankan semua tes (bagian "Tes" di `README.md`) dan suite Playwright terhadap **staging** yang meniru production (Keycloak TLS, domain https).
2. Migrasi (langkah 4).
3. Ganti container backend ke image baru.
4. Smoke test: `GET /api/health` 200, halaman publik terbuka, login admin lewat Keycloak berhasil.
5. Rollback: jalankan kembali image versi sebelumnya (tag SHA lama).

## 6. Hari-hari pertama

- **Job pembersihan:** hari pertama `JOBS_DRY_RUN=true`. Periksa log `dry-run: ... akan dihapus`, lalu set `false`.
- **Pantau:** log production berformat JSON. Beri alert pada lonjakan 5xx, lonjakan 429 (tanda `TRUSTED_PROXIES` atau batas laju salah), dan kegagalan job. Pasang uptime check pada `/api/health`.
- **Keycloak:** hapus operator bootstrap setelah operator permanen dengan MFA siap (`keycloak/README.md` langkah 7).
