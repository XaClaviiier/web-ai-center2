# Kontrak API — AI Center UBAYA CMS

Acuan bersama frontend (mock) dan backend (implementasi). Keputusan bisnis ada di `00-detail-rencana-web.md`; dokumen ini hanya bentuk request dan respons.

Status: kerangka. Bagian bertanda \[diisi\] dilengkapi sebelum fase frontend stack dimulai.

## 1. Endpoint publik

Hanya mengembalikan data berstatus posted.

| Method | Path | Keterangan |
|---|---|---|
| GET | `/api/berita` | Daftar berita; parameter `page`, `limit`, `q` (pencarian judul) |
| GET | `/api/berita/:slug` | Detail berita |
| GET | `/api/dokumentasi` | Daftar dokumentasi; parameter `page`, `limit`, `q` (pencarian judul) |
| GET | `/api/dokumentasi/:slug` | Detail dokumentasi |
| GET | `/api/testimoni` | Daftar testimoni |
| GET | `/api/tim` | Daftar tim |

Layanan (Pelatihan, Inference and Solutions) tidak punya endpoint; isinya hardcode di frontend.

## 2. Endpoint admin

Butuh session login (cookie) dan header `X-CSRF-Token` untuk POST/PUT/DELETE. Path detail \[diisi\].

| Aksi | Berlaku untuk |
|---|---|
| Daftar (semua status), detail | Berita, Dokumentasi, Testimoni, Tim |
| Tambah, edit | Berita, Dokumentasi, Testimoni, Tim |
| Post, Retract | Berita, Dokumentasi, Testimoni |
| Hapus (ke trash), Pulihkan, Hapus permanen | Berita, Dokumentasi, Testimoni |
| Hapus permanen langsung | Tim |
| Pasang lock, heartbeat, lepas lock | Berita, Dokumentasi, Testimoni, Tim |
| Login, callback, logout, info pengguna (nama + role) | Autentikasi |

## 3. Skema JSON per entitas

Field mengikuti rencana; nama dan tipe final \[diisi\].

- **Berita:** id, slug, judul, isi (plain text, paragraf dipisah `\n\n`), gambar (1), tanggal publikasi, judul SEO (opsional), deskripsi SEO (opsional), status, penulis, version.
- **Dokumentasi:** id, slug, judul kegiatan, gambar (1–6, berurutan; pertama = headline), tanggal publikasi, judul SEO (opsional), deskripsi SEO (opsional), status, penulis, version.

Aturan pembuatan slug: \[terbuka\] (lihat `00-detail-rencana-web.md`).

Data awal JSON yang disisipkan Go ke HTML /berita/{slug} dan /dokumentasi/{slug} memakai bentuk yang sama dengan respons `GET /api/berita/:slug` dan `GET /api/dokumentasi/:slug`.
- **Testimoni:** id, nama, teks, posisi, foto (opsional), status, version.
- **Tim:** id, nama, posisi / jabatan, bidang riset, foto, version.

## 4. Format pagination

```json
{ "data": [], "page": 1, "limit": 9, "total": 0 }
```

## 5. Format error

```json
{ "error": { "code": "string", "message": "Pesan dalam sentence case" } }
```

Status yang dipakai: `400`, `401`, `403`, `404`, `409`, `429` (dengan header `Retry-After`). Tidak menampilkan detail internal.

## 6. Contoh mock

\[diisi: satu contoh respons per endpoint publik dan admin\]

## 7. Environment variable

Nama final \[diisi\]. Daftar yang sudah disebut di rencana:

- TTL cache, TTL session (idle 1 jam, mutlak 8 jam), TTL lock (5 menit), interval heartbeat (1 menit).
- Batas rate limiter per bucket.
- `APP_ENV`, nama cookie session, flag `Secure`.
- Daftar IP proxy tepercaya.
- `PUBLIC_BASE_URL`.
- Koneksi PostgreSQL, Redis, RustFS, Keycloak.
