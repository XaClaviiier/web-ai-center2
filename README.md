# Web AI Center — Frontend

CMS frontend AI Center Universitas Surabaya. Terhubung ke backend
(`VITE_API_BASE_URL`, lihat `.env.example`) — bila backend mati, env kosong, atau
mode test, semua halaman otomatis memakai data dummy (lihat `src/lib/api.ts`).

## Tech stack

- React 19 + TypeScript + Vite 7 (SPA, React Router v7)
- Tailwind CSS v4 + `motion` (animasi carousel)
- Vitest + Testing Library + jsdom (unit test)
- Node.js ≥ 20

## Halaman

| Rute | Isi |
|---|---|
| `/` → `/beranda` | Hero carousel (2 slide foto), Layanan (2 kartu), highlight Dokumentasi (3 card), Testimoni, highlight Berita (3 card) |
| `/tim` | Daftar tim + pencarian |
| `/tentang-kami` | Profil, alamat |
| `/berita`, `/berita/:slug` | Daftar + pencarian, detail |
| `/dokumentasi`, `/dokumentasi/:slug` | Daftar + pencarian, detail |

## Arti penanda di kode

- `TODO_BACKEND` — bagian yang **sengaja belum diintegrasikan** + alasannya di komentar (menunggu keputusan konten/endpoint/akun). Sebagian besar endpoint sudah terintegrasi lewat `src/lib/useApiData.ts`.
- `TODO_ASSET` — file aset (mis. foto hero) perlu diganti/ditambah di `public/`.
- Limit 3 card hanya di highlight beranda; halaman `/berita` dan `/dokumentasi` me-render semua data tanpa batas.

## File `.test.tsx` / `.test.ts`

Unit test kecil per unit (input → output + 1–2 edge case). Dijalankan via `npm test`. Bukan integration/e2e test.

Catatan keputusan: 53 file test (39 frontend + 14 backend) **sengaja ikut
di-commit** walaupun aturan kerja awal menyebut file test lokal saja — supaya bukti
unit test ikut terlihat di repo untuk laporan KP dan gerbang bisa dijalankan di
mesin lain.

## Folder `lighthouse_report/`

Arsip hasil ukur Lighthouse (JSON mentah). File `prod*-*.json` = hasil build produksi (acuan skor). Ukur ulang selalu terhadap build produksi, bukan dev server.

## Menjalankan di laptop lain

```bash
npm install
npm run dev      # development (http://localhost:5173)
npm test         # unit test
npm run build    # build produksi ke dist/
npm run preview  # sajikan hasil build (http://localhost:4173)
```

Catatan: butuh Node ≥ 20 dan akses internet (instalasi pertama). Nilai asli (API key dsb.) hanya di `.env` lokal — lihat `.env.example`, jangan commit `.env`.

Menjalankan **dengan data dari database**: salin `.env.example` → `.env`, jalankan
backend `server/` (bagian bawah) dulu, baru `npm run dev`. Tanpa `.env`, web tetap
jalan penuh dengan data dummy.

## Deployment (produksi)

Arsitektur target:

- frontend: container Nginx, hanya menyajikan `dist/`;
- backend: container Bun, hanya dapat diakses jaringan internal;
- reverse proxy: `https://ai-center.example.com/`;
- frontend admin: `https://ai-center.example.com/admin`;
- API: `https://ai-center.example.com/api/*`;
- PostgreSQL dan RustFS: layanan eksternal atau jaringan Docker yang sama.

### Environment backend

Salin `server/.env.example` ke `server/.env`, lalu isi nilai produksi:

- `DATABASE_URL`;
- `CORS_ORIGIN=https://ai-center.example.com`;
- `ADMIN_PASSWORD`;
- `RUSTFS_ENDPOINT`, `RUSTFS_ACCESS_KEY`, `RUSTFS_SECRET_KEY`, `RUSTFS_BUCKET`.

Signing region SDK ditetapkan internal `us-east-1`.
Jangan salin backend `.env` ke frontend atau image frontend.

### Build dan jalankan container

File deployment:

- `deploy/frontend.Dockerfile`;
- `deploy/backend.Dockerfile`;
- `deploy/frontend.nginx.conf`;
- `deploy/nginx.conf`;
- `docker-compose.production.yml`.

Compose hanya mempublikasikan port reverse proxy `80`. Port backend hanya `expose`
di jaringan Compose.

```bash
docker compose -f docker-compose.production.yml config
docker compose -f docker-compose.production.yml up -d --build
```

Compose belum menyediakan PostgreSQL, RustFS, atau TLS certificate. Pastikan
`DATABASE_URL` dan `RUSTFS_ENDPOINT` dapat dijangkau dari container backend.
Pasang terminasi TLS di depan proxy, lalu arahkan HTTPS ke port proxy.

Setelah database baru siap, jalankan migrasi dari container backend:

```bash
docker compose -f docker-compose.production.yml exec backend bun run db:migrate
```

`db:seed` menghapus dan mengisi ulang tabel. Perintah ini hanya untuk setup awal
atau reset development, dan sekarang wajib memakai `SEED_CONFIRM=RESET`.
Jangan jalankan seed pada database produksi berisi konten admin.

### Frontend dan reverse proxy

Tanpa `VITE_API_BASE_URL`, frontend memakai `/api`. Override tetap tersedia untuk
development atau backend beda origin:

```bash
VITE_API_BASE_URL=http://localhost:3000/api
```

Vite memakai `base: '/'`. Nginx meneruskan `/api/*` dan `/uploads/*` ke backend,
serta meneruskan rute lain ke frontend. Backend menerima prefix `/api` apa adanya;
proxy tidak menambah prefix kedua.

SPA fallback mendukung akses langsung ke `/`, `/admin`, dan semua rute React.
Upload gambar tersimpan di RustFS dan dibaca melalui `/uploads/<nama>`.

### Checklist smoke test

- [ ] `GET https://ai-center.example.com/api/health` mengembalikan status `200`.
- [ ] `https://ai-center.example.com/` tampil.
- [ ] `https://ai-center.example.com/admin` tampil saat dibuka langsung.
- [ ] Login admin berhasil dengan `ADMIN_PASSWORD`.
- [ ] CRUD dokumentasi, berita, kursus, tim, dan testimoni berhasil.
- [ ] Upload gambar tampil dari URL `/uploads/...`.

## Backend `server/` (baru)

API **baca publik + tulis admin** — Bun + Elysia + Drizzle ORM + PostgreSQL.
Dev lokal tidak butuh instalasi Postgres: kosongkan `DATABASE_URL` → otomatis
memakai **PGlite** (Postgres in-process, data di `server/.pglite/`,
di-gitignore). Upload gambar memakai AWS SDK S3 v3 ke RustFS.
Isi `.env` dari `server/.env.example` untuk koneksi PostgreSQL dan RustFS
(nilai dummy saja di sana).

```bash
cd server
bun install
bun run db:migrate      # jalankan migrasi SQL (drizzle/)
SEED_CONFIRM=RESET bun run db:seed  # reset CMS; destruktif
bun run dev             # http://localhost:3000/api/health (watch)
bun test                # unit test backend (bun:test)
bun run typecheck       # tsc --noEmit
bun run db:generate     # generate migrasi — hanya setelah ubah schema.ts
```

Seed hanya mengisi lima konten CMS. Tabel konten static legacy dipertahankan,
tetapi tidak diisi ulang dan tidak memiliki route API.

Unit test route memakai fake repository. Query Drizzle diverifikasi lewat
typecheck dan smoke test database terpisah.

### Endpoint

| Endpoint | Query | Catatan |
|---|---|---|
| `GET /api/health` | — | `{ status: "ok" }` |
| `GET /api/berita` | `q`, `page`, `limit`, `order` | `{ items }`; highlight beranda: `?limit=3` |
| `GET /api/berita/:slug` | — | 404 `{ error }` bila tidak ada |
| `GET /api/dokumentasi` | `q`, `kategori`, `page`, `limit`, `order` | `{ items }` |
| `GET /api/dokumentasi/:slug` | — | 404 `{ error }` |
| `GET /api/tim` | `q`, `page`, `limit` | `{ items }` |
| `GET /api/kursus` · `/api/kursus/:kode` | `q`, `page`, `limit` | `{ items }` / 404; kode case-insensitive (`r01`) |
| `GET /api/progress` | `pesertaId` (wajib), `kode` | **publik tanpa akun** — `{ items }` progres milik 1 peserta |
| `POST /api/progress` | body `{ pesertaId, kode, state }` | **publik tanpa akun** — UPSERT per (peserta, kode); rate-limit 60 tulis/menit/IP |
| `GET /api/testimoni` | — | `{ items }` |
| `POST /api/admin/login` | body `{ password }` | rate-limit 5 gagal / 10 menit per IP → 429; balikan `{ token }` |
| `POST` · `PUT` · `DELETE /api/dokumentasi[/:slug]` | — | Bearer token; 400/404/409 eksplisit |
| `POST` · `PUT` · `DELETE /api/berita[/:slug]` | — | Bearer token; slug dibuat server dari judul |
| `POST` · `PUT` · `DELETE /api/kursus[/:kode]` | — | Bearer token; `kode` unik |
| `POST` · `PUT` · `DELETE /api/tim[/:id]` | — | Bearer token; POST menyimpan urutan 0 dalam transaksi |
| `POST` · `PUT` · `DELETE /api/testimoni[/:id]` | — | Bearer token; POST menyimpan urutan 0 dalam transaksi |
| `POST /api/admin/upload` | multipart `gambar` | file ≤ 2 MB (jpg/png/webp) → `{ url }` `/uploads/<nama>` |
| `POST /api/admin/uploads/bersihkan` | `{ kering? }` | hapus file unggahan tidak terpakai; `kering: true` = mode uji |

Semua field teks punya batas panjang di `server/src/lib/tulis.ts` (Prioritas 2) —
body yang lewat batas ditolak 400 dengan pesan menyebut field + batasnya.

CORS: whitelist origin lewat `CORS_ORIGIN` (default `http://localhost:5173`).

### Status integrasi frontend

Terintegrasi lewat `src/lib/useApiData.ts` (`useApiDaftar`/`useApiObjek`).
Frontend memakai `/api` pada same-origin production; `VITE_API_BASE_URL` dapat
dioverride. Fallback dipakai saat backend mati, override kosong, respons gagal,
atau unit test:

- Beranda: testimoni (tepat 3 teratas), highlight berita & dokumentasi
- `/berita` + `/berita/:slug`, `/dokumentasi` + `/:slug`, `/tim`
- katalog kursus + detail (`/api/kursus`, `/api/kursus/:kode`)
- **progres belajar** halaman detail kursus: localStorage tetap sumber utama,
  disinkronkan ke backend **tanpa akun** (`GET`/`POST /api/progress`, identitas =
  `pesertaId` acak di browser) — lihat `src/lib/progresPeserta.ts`; saat keduanya
  punya data: **progres lebih tinggi menang**, seri → jejak belajar lebih banyak,
  seri penuh → lokal (`gabungStateBelajar`)

Konten berikut **sudah bukan bagian dari backend** — sumber satu-satunya ada di
file frontend (diedit via editor kode, bukan lewat `/admin`):

- hero slide → `src/data/hero.ts`, layanan → `src/data/layanan.ts`
- visi/misi profil → `src/data/profil.ts`, konten inference → `src/data/inference.ts`

Kontak footer/WA **sengaja statis** dari situs lama (nomor + pesan persis link
`wa.me` di sana) — lihat `src/data/kontak.ts`.

Masih `TODO_BACKEND` (sadar — menunggu keputusan, bukan lupa):

- konten video & kuis per modul — sistemnya sudah ada (`modul.video` +
  `modul.quiz`); isi lewat template `docs/data-kursus.md`;
- akun peserta lintas perangkat belum ada; progres saat ini memakai identitas
  browser `pesertaId`.

CRUD admin kini **5 tab**: **dokumentasi**, **berita**, **kursus**, **tim** &
**testimoni**. Hanya lima jenis konten itu yang masuk CMS. Konten hero, layanan,
profil, inference, kontak, footer, dan klien tetap statis di frontend.

Duanya khusus form **Tim** & **Testimoni**: saat **Tambah baru** field urutan
**tidak ditampilkan** — item baru otomatis paling atas (server menggeser semua
baris lama +1, tidak ada yang dihapus). Testimoni baru langsung jadi **#1** di
slider beranda (beranda selalu tampil tepat 3; yang lama turun urutan dan tetap
tersimpan, bisa diubah lewat admin). Urutan manual tetap bisa diatur saat
**Ubah**. Tombol **Keluar** kini bergaris merah mencolok dan menampilkan
konfirmasi "Yakin keluar dari dashboard admin?".

Yang sengaja TIDAK dikelola CMS: **klien** (section Our Client dikomentari,
keputusan rapat) & konten statis di atas.

Field **Kategori** (form dokumentasi) kini berupa **dropdown opsi umum** —
Workshop, Kunjungan, Demo (`OPSI_KATEGORI` di `src/pages/Admin.tsx`) — plus
pilihan **"Lainnya (ketik sendiri)"** yang membuka input teks (maks 100
karakter, sejajar server). Kategori custom lama tetap tampil di input saat
diedit, jadi daftar pilihan tidak menumpuk dan tidak ambigu; menambah kategori
umum cukup satu entri di `OPSI_KATEGORI`.

Form **Kursus** juga berubah: **Target peserta** kini berupa **dropdown 1
pilihan** dari 4 label katalog (Mahasiswa, Dosen, Guru, Masyarakat umum) —
label selalu persis sama dengan filter katalog publik (dulu salah ketik membuat
kursus tidak muncul di filter mana pun), dan **Kode dibuat otomatis** dari
peserta terpilih + nomor berikutnya (prefiks M/D/G/U, `src/lib/kodeKursus.ts`):
Mahasiswa → M01, Guru → G01. Kode read-only saat tambah dan tetap terkunci saat
edit; kode lama (R01/E01/P01) berprefiks lain sehingga tidak ikut menaikkan
nomor. Kursus lama yang targetnya lebih dari satu tinggal pilih salah satunya
(catatan "Data lama punya N target …" mengingatkan bahwa simpan hanya
menyimpan pilihan itu) — Simpan tanpa target peserta ditolak dengan pesan
eksplisit.

Catatan hero (kini statis di `src/data/hero.ts`, tanpa tab admin): varian `srcSet`
harus menunjuk file yang benar-benar ada (kalau tidak, slide tetap menampilkan
gambar lama), dan slide tidak boleh dikosongkan (hero beranda jadi kosong).

## Ringkasan pengerjaan

1. Scaffold Vite React-TS + Tailwind + motion + Vitest + router; desain diadaptasi dari rapat sebelumnya.
2. Komponen reusable: layout header/footer (logo Ubaya → AI Center, dropdown Konten), hero carousel + client carousel otomatis (dots, jeda saat hover/fokus, hormat reduced-motion), slider testimoni, card berita/dokumentasi/tim dengan slot gambar.
3. Data dummy bertanda backend untuk hero, berita, dokumentasi, tim, klien, testimoni.
4. Optimasi Lighthouse hingga desktop 100/100/100/100 dan mobile 95/100/100/100 (robots.txt, dimensi gambar, target sentuh, kompresi + WebP responsif, code-splitting route, preload LCP).
5. Keputusan rapat 16 Sept: layanan tinggal 2 (Pelatihan + Inference Solution), Our Client dikomentari (kode dipertahankan), testimoni naik ke posisi Our Client dengan background `bg-soft`, kontak footer/Tentang Kami disentralisasi di `src/data/kontak.ts` (email `aicenter@unit.ubaya.ac.id`, WA 0895-6342-22240, website LPPM).
6. Unit test: frontend **238 test / 30 file** (lib, data, komponen, halaman, routing); backend **86 test**. `tsc` bersih di dua sisi.
7. Integrasi frontend↔backend: tabel + route `GET /api/kursus` (+ migrasi `0001`), lapisan `src/lib/api.ts` + hook `useApiData`, wiring semua halaman konten; seed disinkronkan dengan data frontend (tim 6 anggota, teks layanan, gambar hero).
8. Prioritas 2 (CRUD profil & tim, batas panjang field, pembersih gambar tidak terpakai) lalu konten halaman **Inference** pindah ke CMS: tabel `inference` (+ migrasi `0003`), `GET`/`PUT /api/inference`, tab Inference di `/admin`, dan `LayananDetail` memakai `useApiObjek('/inference', …)`.
9. Lengkapi CRUD admin: **testimoni** & **slide hero** (`POST`/`PUT`/`DELETE /api/testimoni[/:id]` dan `/api/hero-slides[/:id]`, tab Testimoni & Hero di `/admin`) — `id` & `urutan` ikut kontrak respons `GET`-nya, lalu `referensiGambar` juga membaca `image`/`srcSet` hero supaya pembersih gambar tidak terpakai tidak menghapus varian slide. Istilah "gambar yatim" diganti jadi **"gambar tidak terpakai"** (UI, kode: `lib/orphanUpload.ts` + `findOrphanUploads`, dan README). Gerbang saat ini: frontend **381 test / 36 file**, backend **221 test / 14 file**, `tsc` bersih di dua sisi + build produksi OK.
10. Audit backend terhadap 10 aturan kerja (`docs/prompt-persiapan-opencode.md`) — semua
    terpenuhi kecuali dua keputusan sadar yang dicatat di README ini: file test
    ikut repo, dan `drizzle.ts` tanpa unit test. Lalu putaran persiapan final:
    `RUSTFS_*` didokumentasikan di `server/.env.example`, penanda
    `TODO_BACKEND` di `robots.txt` diganti catatan deploy (penanda hanya untuk
    bagian yang butuh backend), `VITE_SITE_URL` dijelaskan fungsinya, serta
    section **Deployment (produksi)** + checklist sebelum final di README.
11. Field **Kategori** form dokumentasi di `/admin` → **dropdown opsi umum**
    (Workshop/Kunjungan/Demo) + "Lainnya (ketik sendiri)" yang membuka input
    teks (maks 100, sejajar server) — mengganti input bebas yang membuat
    kategori menumpuk & ambigu; nilai custom lama tetap terbaca saat diedit
    (dropdown menunjuk "Lainnya"). Gerbang saat ini: frontend **384 test /
    36 file** (3 test baru di `Admin.test.tsx`), `tsc -b` bersih, build OK;
    backend 221 test / 14 file.
12. Form **Kursus**: target peserta jadi **dropdown 1 pilihan dari 4 label
    katalog** (sebelumnya teks bebas; respons lanjutan: dari centang kembali ke
    dropdown) + **kode auto generate** dari peserta terpilih
    (`lib/kodeKursus.ts`: prefiks M/D/G/U + nomor terbesar berikutnya,
    read-only saat tambah, terkunci saat edit), guard "Target peserta wajib
    dipilih (menentukan kode otomatis)" sebelum request, dan catatan "Data lama
    punya N target …" untuk kursus multi-target lama. Gerbang saat ini:
    frontend **393 test / 37 file** (5 test lib baru + 4 test admin),
    `tsc -b` bersih, build OK; backend 221 test / 14 file.
13. **Optimasi Lighthouse** pada build produksi: peta chunk ditaruh di
    `index.html` sebelum IIFE pemakainya (modulepreload kini jalan ±73 ms),
    semua rute di-lazy kecuali NotFound, dan CSS di-inline ke `<style>` oleh
    plugin `inlineCssBuild` → **mobile 93 / desktop 99** (audit Lighthouse CLI
    headless tanpa ekstensi; dev `:5173` tidak dipakai untuk audit).
14. **CRUD Layanan**: tab **Layanan** di `/admin` (form nama/tagline/deskripsi +
    editor daftar `fitur`, slug read-only) → `POST`/`PUT`/`DELETE
    /api/layanan[/:slug]` dengan guard token; kunci baris = `slug` hasil
    `slugDariJudul(nama)` (bentrok → 409; slug dipegang tetap saat edit),
    validasi `validasiLayanan` di `server/src/lib/tulis.ts` (1–15 item fitur,
    maks 300 karakter per item). Gerbang saat ini: frontend **396 test /
    37 file** (3 test admin baru), `tsc -b` bersih, build OK; backend **232
    test / 14 file** (11 test baru: `validasiLayanan` + rute layanan).
15. **Katalog kursus dirapikan ala Coursera**: kartu katalog
    (`/layanan/pelatihan`) disusun ulang (eyebrow "Ubaya AI Center · kode",
    judul, baris fakta "level · durasi · N modul · ±X jam", instruktur, satu
    CTA — klaim "Sertifikat" dihapus dari kartu); hero detail kursus dapat
    fakta **Estimasi belajar** dari helper baru `src/lib/kursus.ts`
    (`menitDariMeta`/`totalMenitKursus`/`labelDurasiTotal`, 10 unit test).
    Data asli yang perlu dikonfirmasi user dikumpulkan di
    **`docs/data-kursus.md`** (template isian level/format/instruktur/
    sertifikat/video/kuis + penanda `TODO_KONTEN` di `src/data/pelatihan.ts`).
16. **Sistem video & asesmen per modul**: field `modul` kini boleh memuat
    `video` (URL http(s) ≤ 300, divalidasi `bacaVideoModul`) dan `quiz`
    (≤ 10 soal × 2–6 opsi × 200 karakter, `kunci` = index opsi —
    `bacaKuisModul` di `server/src/lib/tulis.ts`; baris lama tanpa field ini
    tetap valid → dinormalkan `null`/`[]`). Tab **Kursus** di `/admin` punya
    input URL video + editor soal (radio jawaban benar, +/− opsi, buang soal
    kosong saat kirim via `kuisSiapKirim`). Ruang belajar (`LearningWorkspace`)
    meng-embed YouTube/Vimeo mode nocookie (`src/lib/video.ts`, URL lain →
    tautan luar), dan checkpoint memakai kuis nyata bila tersedia (lulus =
    semua soal benar → `lulusQuiz`); tanpa data → perilaku lama (blok
    "video menyusul" + checkpoint dummy). Gerbang saat ini: frontend **416
    test / 39 file**, `tsc -b` bersih, build OK; backend **238 test / 14
    file** (6 test baru: validasi video/kuis + rute kursus).
17. **Audit pre-final** (10 Okt 2026) terhadap 10 aturan kerja + kesiapan
    deploy: pindai rahasia (`.env` ter-ignore, `.env.example` dummy semua),
    penanda TODO tiap butir ber-alasan, nol referensi WordPress (komentar
    "plugin WordPress" di `pelatihan.ts` diparafrase & typo nomor WA di
    `kontak.ts` dibetulkan `6289563422240`), semua tautan internal + anchor
    (`#layanan`/`#kontak`) cocok dengan peta rute `App.tsx` (skrip audit
    lokal), angka test basi di README diperbarui (53 file). Verifikasi:
    smoke API **36/36**, smoke halaman **11/11** (h1 benar, konsol 0
    error/warning), health `{"status":"ok","db":"ok"}`, gerbang ulang FE
    **416/416** + BE **238/238**, dan **Lighthouse mobile 93 / desktop 100**
    (a11y, best-practices, SEO 100 di dua mode; TBT mobile 30 ms) terhadap
    build produksi — arsip `lighthouse_report/prod-2026-10-10-*.json`
    (lokal/gitignored).
18. **Refactor admin + urutan otomatis** (15 Okt 2026): 4 tab **Layanan/
    Profil/Inference/Hero dihapus** — halaman publiknya tetap, kontennya jadi
    statis frontend (`src/data/{hero,layanan,profil,inference}.ts`); endpoint
    lama di server dibiarkan utuh (tak dipakai FE). Backend: `naikkanSemuaUrutan()`
    (`server/src/repositories/`) dipanggil sebelum `POST /api/tim` &
    `/api/testimoni` → semua baris lama digeser +1 lalu item baru disimpan
    `urutan: 0` (testimoni baru = #1 beranda, lama turun; tidak ada yang
    dihapus). Form admin: field urutan disembunyikan saat **Tambah** (tetap ada
    saat **Ubah**); beranda membatasi `slice(0, 3)` testimoni; tombol **Keluar**
    bergaris merah + `window.confirm("Yakin keluar dari dashboard admin?")`.
    Gerbang saat ini: frontend **406 test / 39 file** (`tsc -b` bersih, build
    OK), backend **238 test / 14 file**.
19. **Geist Mono bobot nyata + Lighthouse ulang** (15 Okt 2026): screenshot
    spesifikasi Figma dicek ulang — palet Nuansa Biru (6 primer + 4
    pelengkap) & font (Cabinet Grotesk judul / General Sans isi /
    Geist Mono teknis + fallback Helvetica) sudah cocok 100% dengan token
    `src/index.css`. Satu celah ditemukan: `public/fonts/GeistMono-Regular.woff2`
    statis bobot 400, padahal kicker/tag memakai `font-mono font-bold` (5
    tempat) → bold sintesis browser. Diganti **file variabel resmi Google
    Fonts** (OFL, subset latin, `GeistMono-Variable.woff2`, sumbu 100–900,
    `font-weight: 100 900`) → bold sungguhan (diverifikasi: cetak 400 ≠ 700
    via canvas, `document.fonts.check('700 …')` true). Test baru: rentang
    bobot mono + seluruh `url('/fonts/*.woff2')` di CSS wajib ada di disk.
    Gerbang: FE **408/408**, BE **238/238**, `tsc` bersih; **Lighthouse
    ulang** terhadap build hasil refactor: **mobile 92 / desktop 99**
    (a11y, best-practices, SEO 100/100 di dua mode; TBT mobile 30 ms) —
    arsip `lighthouse_report/prod-2026-10-15-*.json` (lokal/gitignored).
20. **Progres LMS tersimpan di backend tanpa akun** (15 Okt 2026): menutup
    `TODO_BACKEND` progress per peserta tanpa menunggu keputusan akun —
    tabel `progress_pembelajaran` (PK gabungan `peserta_id` + `kode`, migrasi
    `0004`), repository + mapper `toProgresPeserta`, dua endpoint **publik**:
    `GET /api/progress?pesertaId=&kode=` (filter ketat 8–64 `[A-Za-z0-9-]`,
    kode uppercase) dan `POST /api/progress` (UPSERT per peserta+kursus,
    `state` objek ≤ 32 KB, rate-limit **60 tulis/menit/IP** → 429 +
    `Retry-After`). Identitas peserta = `pesertaId` acak per-browser
    (`src/lib/progresPeserta.ts`); localStorage tetap sumber utama (offline),
    backend = salinan cadangan; saat keduanya ada → **progres lebih tinggi
    menang, seri → jejak belajar lebih banyak, seri penuh → lokal**
    (`gabungStateBelajar` + `jejakStateBelajar`), dan state kosong tidak
    dikirim (`stateBelajarTerisi`) supaya tak ada baris percuma. Semua
    kegagalan sinkron disenyapkan (konsol tetap bersih). TODO inference
    diklarifikasi: `/api/inference` = CMS konten halaman, bukan prediktor —
    output nyata tetap menunggu keputusan model. Gerbang: **FE 424 test /
    40 file**, **BE 255 test**, `tsc` bersih dua sisi, build OK; uji langsung
    API (create → get → upsert 1 baris → 400 untuk input invalid) & migrasi
    `0004` terpasang di DB dev. E2E browser: enroll → POST 200 (kode
    uppercase), hapus state lokal → reload → ruang belajar **pulih dari
    server**, aksi lesson → POST 200, konsol 0 error/warning. Lighthouse
    ulang: **mobile 91 / desktop 99** (a11y/BP/SEO 100/100, CLS 0).

21. **PROGRESS 4 — status permintaan user + warna situs disentralisasi ke
    token** (17 Okt 2026): audit tiga poin briefing terhadap kode.
    (a) Layanan/Profil/Inference/Hero "dihapus, nanti diedit via front" —
    **sudah** sejak refactor admin: 5 tab tersisa (Dokumentasi, Berita,
    Kursus, Tim, Testimoni), konten keempat bagian statis di `src/data/*`.
    (b) Prinsip "tambah baru tidak menambah jumlah tampilan" — **sudah**:
    testimoni slider tampil tepat 3 (counter `1 / 3`; item baru via admin
    otomatis #1, lama turun, tak dihapus), dokumentasi & berita di beranda
    selalu `slice(0, 3)` terbaru (urut tanggal desc; form admin default
    tanggal hari ini) — dikunci test (baris "testimoni tampil tepat 3",
    "highlight membatasi 3 card", "dokumentasi 3 terbaru").
    (c) Font & palet Figma — font + 10 warna daftar sudah 100% (butir 19);
    celah terakhir = 14 hex hardcoded di 6 file komponen. Dipindahkan ke
    token `@theme` tanpa mengubah tampilan: 3 hex = nilai persis token
    (gradien beranda kini `var(--color-navy/-sky/-soft)`), 5 hex turunan
    navy dideklarasikan (`--color-onnavy`, `-badge`, `-body`, `-soft`,
    `-dim`; nilai persis hex lama). Test baru: nilai tiap turunan persis
    + kontras ≥4.5:1 vs navy + **skener nol hex hardcoded** di seluruh
    `src/**/*.{ts,tsx}` (semua warna wajib lewat token). Gerbang: **FE 427
    test / 40 file**, **BE 255 test**, `tsc` bersih dua sisi, build OK;
    E2E beranda: keempat utilitas token warna terbaca persis di devtools,
    gradien `var()` ter-render, counter testimoni `1 / 3`, konsol 0
    error/warning. **Lighthouse ulang** (URL `/beranda`, metode sama dengan
    15 Okt): **mobile 90 / desktop 99** (a11y, best-practices, SEO
    100/100 di dua mode) — arsip
    `lighthouse_report/prod-2026-10-17-*.json` (lokal/gitignored).

22. **UX ruang belajar diperjelas (permintaan user setelah uji coba)**
    (17 Okt 2026): user bingung saat "Mulai modul" tidak bereaksi dan label
    "Lesson sudah ditonton" terlihat seperti tombol aktif. Tiga perbaikan
    (nilai/aturan belajar tidak berubah — hanya kejelasan):
    (a) kartu terkunci kini menjelaskan alasannya: "Tonton lesson + lulus
    checkpoint modul sebelumnya untuk membuka kunci." (3 kartu terkunci);
    (b) klik "Mulai modul"/"Ulangi" auto-scroll ke panel lesson
    (`#panel-lesson`; dulu panel terbuka diam-diam di bawah kartu);
    (c) setelah lesson ditonton, tombol berubah jadi penanda ✓ non-klik
    "Lesson ditonton" + pengarah "Selanjutnya: kerjakan checkpoint di bawah
    untuk menyelesaikan modul." (kunci i18n lama 'Lesson sudah ditonton'
    diganti; kamus ID→EN ikut diperbarui). Test baru: hint terkunci ×3,
    auto-scroll `Mulai modul` (spy `scrollIntoView`), status ✓ non-klik +
    pengarah. Gerbang: **FE 429 test / 40 file**, `tsc` bersih, build OK.
    E2E (state dibersihkan dulu — localStorage + baris cadangan backend
    di-overwrite state kosong): enroll → scroll ke ruang belajar → hint
    terkunci tampil 3× → "Mulai modul" → panel lesson pindah ke posisi
    atas layar (scrollY 1801 → 2561) → "Mulai lesson" → badge ✓ muncul,
    tombol aktif hilang, watched tersimpan → konsol 0 error/warning.

### Penanda TODO

- `TODO_BACKEND` — bagian yang sengaja belum diintegrasikan + alasannya (lihat "Status integrasi frontend").
- `TODO_ASSET` — file aset perlu diganti/ditambah di `public/`.
- `TODO_KONTEN` — teks menyusul dari dokumen resmi (mis. visi/misi dari PDF rapat 16 Sept 2026); untuk kursus: template isian data asli (level, instruktur, sertifikat, URL video + kuis per modul) ada di `docs/data-kursus.md`.
