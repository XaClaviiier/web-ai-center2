# Rencana: Menggabungkan `frontend-final/` (branch `frontend`) dengan backend Go (branch `Fullstack`)

> Berkas ini disimpan di `docs/backend/PLAN-MERGE-FULLSTACK.md` atas permintaan eksplisit pengguna (pengecualian dari aturan "jangan menyunting `.md` di `docs/`", berlaku hanya untuk berkas ini bersama `perubahan-kode.md`, `req.md`, `schema-baru.md`). Isi lama (merge prototipe HTML/CSS/JS ke backend) digantikan oleh rencana ini; riwayat langkah lama ada di `docs/backend/perubahan-kode.md` dan git.
> **Status (2026-11-13): perencanaan. Belum ada kode yang diubah untuk tahap ini.**

## 1. Konteks

- **Backend** (`backend/`): Go + Chi, PostgreSQL, Redis, RustFS. API publik dan admin (berita, dokumentasi, testimoni, tim), upload, cache + ETag, rate limit, soft lock, SEO detail, job berkala (termasuk rolling Log Aktivitas), header keamanan. Auth sementara `POST /api/auth/dev-login`; Keycloak belum.
- **Frontend baru** (`frontend-final/`, branch `origin/frontend`, dikerjakan tim frontend): Astro 7 + TypeScript (halaman publik statis), React 19 + Vite (admin di `/admin`, base `/admin/`), Tailwind v4 tanpa preflight, shadcn, Embla, GSAP/motion, TanStack Query, Bun. Keluaran build: `dist/public` (+ `dist/public/admin`). Data masih **API tiruan** di `src/simulation/` (`api.ts`, `state.ts`, `Panel.tsx`) dan pemilihan akun pengganti Keycloak (`/admin/?simulation=login`).
- **Frontend lama** (`prototype/` → `web/` lewat `scripts/bangun-web.cjs`): digantikan. `web/` lama direname menjadi `web-deprecated/`.

## 2. Keputusan pengguna

1. Pengguna sendiri yang melakukan merge `origin/frontend` ke `Fullstack`. Agen tidak merge, commit, atau push.
2. **`frontend-final/` tidak boleh disentuh** (tidak diedit, tidak ada klien API ditambahkan di dalamnya).
3. `web/` saat ini direname menjadi `web-deprecated/`; `web/` baru adalah keluaran build `frontend-final/` yang disalin oleh skrip (bukan disunting tangan).
4. **Backend menyesuaikan kontrak yang diharapkan frontend.** Kontrak ditentukan dari objek `api` di `frontend-final/src/simulation/api.ts` dan tipe di `src/shared/types.ts`; backend yang disesuaikan, bukan sebaliknya.
5. Rencana ini menggantikan isi lama berkas ini.

## 3. Aturan kerja (dari CLAUDE.md, tetap berlaku)

- Dilarang menyunting `.md` di `docs/` selain `docs/backend/{perubahan-kode,req,schema-baru}.md` dan berkas ini. Bila docs (termasuk perubahan yang dibawa branch `frontend`) tampak bertentangan dengan backend, laporkan ke pengguna.
- Docs yang belum menjawab (`[diisi]`, `[terbuka]`) → tanya pengguna, jangan menebak.
- Catat langkah major di `perubahan-kode.md`, library/alat baru di `req.md`, perubahan skema di `schema-baru.md`. Jangan klaim selesai sebelum verifikasi dijalankan. Setiap penyimpangan dari `docs/` dicatat beserta alasan dan pemutusnya.
- Tes memakai DB `ai_center_test`, Redis DB 15, bucket `ai-center-test`; jangan mendaftar-lalu-menghapus isi bucket dev; jangan jalankan server dengan job aktif terhadap bucket dev; jangan menghentikan proses server milik pengguna.

## 4. Kontrak frontend vs backend (hasil identifikasi, baca-saja)

| Hal | Frontend mengharapkan | Backend sekarang | Keputusan |
|---|---|---|---|
| Tanggal publikasi | string `DD-MM-YYYY` (seed, form, tampilan) | `YYYY-MM-DD` (docs/database.md, stack.md §9) | DIPUTUSKAN |
| Versi bentrok | 409, kode `conflict` (UI hanya memeriksa status 409) | 409 `version_conflict` | cocok di status; kode DIPUTUSKAN |
| Edit item bukan draft | 400 `bukan_draft` | 409 `status_tidak_valid` | DIPUTUSKAN |
| Kunci dipegang orang lain | 423 `locked` (`Sedang diedit oleh {nama}`) | server menolak tulis bila dipegang orang lain; kode galat dan status perlu dicocokkan | DIPUTUSKAN |
| Kunci hilang saat heartbeat | 410 `lock_lost` (UI hanya memeriksa status 410) | rute lock perlu dicocokkan: acquire `{milik:'saya'\|'lain', nama?}`, heartbeat `{ok:true}`, lepas | DIPUTUSKAN |
| Gambar dokumentasi | payload `gambarDaftar:[{key}]`; respons `gambar:[{id,key,urutan,dokumentasiId}]` | payload `gambar:[{key}]` (keadaan akhir, 1-6 item) | DIPUTUSKAN |
| Daftar admin | `adminDaftar(sumber,{trash})` → `{items}` | `?status=semua\|draft\|posted\|trash` → `{items}` | cocok (klien memetakan `trash`) |
| Log aktivitas | `{items,page,limit,total,usernames}` | sama | cocok |
| Daftar/detail publik | `{items,page,limit,total}`; testimoni dan tim `{items}`; detail 404 bila tidak ada | sama | cocok |
| Sesi, CSRF, unggah | tidak ada di tiruan (akun lewat `?akun=`) | cookie `HttpOnly`, `X-CSRF-Token`, `/api/admin/upload`, `Origin` ber-allowlist | klien nyata wajib menambah; lokasi kode tergantung §5 |

### 4b. Keputusan pengguna (2026-11-13) dan selisih tambahan yang ditemukan

Diputuskan: tanggal API mengikuti frontend (`DD-MM-YYYY`); kode galat disamakan dengan frontend (`conflict`, `bukan_draft`, `locked` 423, `lock_lost` 410); payload dokumentasi memakai `gambarDaftar`; jembatan klien berupa skrip pascabuild di `scripts/`. Semuanya menyimpang dari `docs/` (stack.md §9, database.md) dan dicatat sebagai penyimpangan di `perubahan-kode.md`. Parameter log `dari`/`sampai` tetap `YYYY-MM-DD` (frontend membandingkan dengan `hariJakarta`).

Selisih tambahan (belum diputuskan; pemetaan di jembatan atau perubahan backend):

- Nama field metadata: frontend `createdAt`/`createdBy`/`updatedBy`/`deletedAt`; backend `dibuatPada`/`dibuatOleh`/`diubahOleh`/`dihapusPada`.
- Dokumentasi publik dan admin: frontend membaca `gambar:[{key,urutan}]` di daftar (`gambar[0].key`, `gambar.length`); backend daftar memberi `gambarUtama`+`jumlahGambar`, detail `gambar:[{url}]` (publik) / `[{key,url}]` (admin).
- Gambar: frontend memanggil `image(key)`; backend mengirim URL `/media/...` pada berita/testimoni/tim.
- Lock: acquire frontend `{milik:'saya'|'lain', nama}` (200); backend 423 `locked` bila dipegang orang lain, 200 `{pemegang,milikSaya,...}` bila berhasil. Heartbeat `{ok:true}` vs objek lock.
- Login: frontend memilih akun lewat `?akun=`; backend `dev-login` dengan password.

## 5. Pertanyaan struktural yang menunggu keputusan

1. **Tempat klien API nyata.** Karena `frontend-final/` tidak boleh disentuh, jembatan ke `/api/*` (cookie, CSRF, unggah, lock, pemetaan tanggal) tidak dapat ditaruh di dalamnya. Opsi: (a) skrip pascabuild di `scripts/` yang mengganti modul simulasi pada bundel; (b) tim frontend menyediakan titik sambung (mis. `api` yang dapat diganti lewat env/alias build) — diurus tim frontend; (c) backend meniru bentuk tiruan sepenuhnya sehingga hanya "sumber data" yang ditukar. Perlu dipilih setelah hasil build dilihat.
2. **Alat build.** `frontend-final` memakai Bun; repo ini memakai npm/Node. Siapa yang menjalankan build dan di mana artefaknya disimpan.
3. **Hasil build di `web/`.** Struktur Astro (`/_astro/`, `404.html`, detail statis hanya untuk slug seed) berbeda dari `web/` lama. Go harus menyajikan `dist/public` dan tetap menyisipkan SEO ke kerangka detail generik (stack.md §3).

## 6. Langkah

| # | Langkah | Pelaksana | Status |
|---|---|---|---|
| 1 | Rename `web/` → `web-deprecated/`; merge `origin/frontend` ke `Fullstack` | pengguna | menunggu |
| 2 | Putuskan item **terbuka** di §4 dan §5 | pengguna | menunggu |
| 3 | Sesuaikan backend dengan kontrak yang diputuskan (handler, kode galat, format tanggal, payload gambar, rute lock), lengkap dengan tes | agen | menunggu #2 |
| 4 | Skrip penyalin build ke `web/` tanpa menyentuh `frontend-final/`; sesuaikan `static.go` untuk struktur Astro dan fallback `/admin/*`, `404.html` berstatus 404 | agen | menunggu #1, #2 |
| 5 | Sisip SEO Go pada kerangka detail Astro generik; slug tidak ada → 404 + noindex | agen | menunggu #4 |
| 6 | Sesuaikan CSP (`middleware.go`) agar Astro/React berjalan tanpa pelanggaran, tanpa melonggarkan berlebihan | agen | menunggu #4 |
| 7 | Perbarui `scripts/e2e/*` (selektor DOM berubah) dan jalankan ulang | agen | menunggu #5, #6 |
| 8 | Hapus jalur lama (`bangun-web.cjs`, `scripts/tambahan-web/`, `prototype/` bila tak dipakai) | atas keputusan pengguna | menunggu |

## 7. Verifikasi

`go vet ./...`, `gofmt -l .`, `go test -p 1 ./... -count=1`, `govulncheck ./...`; build `frontend-final` hijau; e2e Playwright terhadap server uji di port bebas; konsol browser tanpa pelanggaran CSP; audit dependensi frontend baru dicatat di `req.md`.
