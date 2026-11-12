# Rencana: Menggabungkan frontend final (branch `Fullstack`) dengan backend Go (branch `backend`)

> Berkas ini dibuat agar bisa di-push dan dibaca ulang di branch `Fullstack` (sesi baru boleh langsung mengerjakannya).
> Lokasi akhir: `docs/backend/PLAN-MERGE-FULLSTACK.md` (permintaan eksplisit pengguna; ini pengecualian tambahan dari aturan "jangan menyunting `.md` di `docs/`", berlaku hanya untuk berkas ini).

> **Status revisi (2026-11-12):** Langkah 0 dan 1 sudah SELESAI di luar rencana ini (PR #9 menggabungkan `backend` ke `Fullstack`, `76533b1`; `backend/`, `CLAUDE.md`, `docs/backend/*` sudah ter-track; `01-kontrak-api.md` dan `Schema.md` sengaja di-`.gitignore`). Pekerjaan berikutnya mulai dari Langkah 2. Urutan kerja: 2 → 3 → 4 → 5 → 6 → 7.

## 1. Konteks

- **Backend** (branch `backend`, HEAD `921b05d`): Go + Chi di `backend/`. Selesai: API publik, CRUD admin (berita, dokumentasi, testimoni, tim), upload, cache Redis + ETag, rate limit, soft lock, SEO (HTML detail ber-meta, `sitemap.xml`, `robots.txt`), job berkala. Auth sementara: `POST /api/auth/dev-login` (hanya `APP_ENV=development`, dua akun dev). Keycloak belum.
- **Frontend** (branch `Fullstack`, `ec71707`): yang disebut "final" ternyata **prototipe HTML/CSS/JS murni** di `prototype/` (13 halaman publik, 1 halaman admin hash-route, data tiruan di `prototype/simulasi/api-simulasi.js`). Tidak ada proyek Astro atau React-admin baru. `src/` dan `server/` di branch itu adalah kode lama (React + Bun) yang hampir sama dengan branch `backend`.
- **Celah yang harus ditutup**: Go belum menyajikan file statis (`/`, `/admin/*`, 404 untuk path acak); prototipe memakai tautan `*.html`, path relatif `../assets/`, dan `?slug=`; lapisan data tiruan harus diganti `fetch` nyata; ada beberapa perbedaan kontrak (lihat langkah 5); halaman Log Aktivitas belum punya backend.

## 2. Keputusan pengguna (final)

1. **Bentuk frontend:** prototipe dipakai apa adanya (HTML/CSS/JS), disajikan oleh Go, satu origin, tanpa build frontend. Migrasi ke Astro/React = pekerjaan terpisah di kemudian hari. (Menyimpang dari `docs/stack.md`; itu keputusan pengguna.)
2. **Log Aktivitas:** gabungkan dan sesuaikan dulu; endpoint/tabel log dikerjakan SESUDAH penggabungan (langkah terpisah dengan `/grill-me`). Sementara itu menu Log Aktivitas disembunyikan.
3. **Login dev:** halaman `keycloak.html` dipertahankan sebagai pintu masuk login development, tetapi sesinya nyata (lihat langkah 4d).
4. **Git:** `git merge backend` di branch `Fullstack`. Kode lama (`src/`, `server/`, `package.json`, `vite.config.ts`, `bun.lock`) dipertahankan; penghapusannya langkah terpisah yang disetujui pengguna.
5. **Dokumen kerja ikut ke Fullstack:** `docs/backend/` dan `CLAUDE.md` dari branch `backend` dilacak git dan ikut ke `Fullstack`; berkas rencana ini disimpan di `docs/backend/`.

## 3. Aturan kerja (dari CLAUDE.md, tetap berlaku)

- Dilarang menyunting `.md` apa pun di `docs/` (kecuali `docs/backend/{perubahan-kode,req,schema-baru}.md` dan berkas rencana ini). Bila isi docs tampak keliru, laporkan ke pengguna. Setelah langkah 0, `docs/backend/` dan `CLAUDE.md` ikut ter-track dan ikut ke `Fullstack`.
- Docs yang belum menjawab (`[diisi]`, `[terbuka]`) → tanya pengguna, jangan menebak.
- Catat tiap langkah major di `docs/backend/perubahan-kode.md` (sesi yang sama), library baru di `req.md`, perubahan skema di `schema-baru.md`. Jangan klaim "selesai" sebelum verifikasi dijalankan.
- Jangan pernah mendaftar-lalu-menghapus isi bucket dev `RUSTFS_BUCKET`; tes hanya memakai bucket `ai-center-test`, DB `ai_center_test`, Redis DB 15.
- Setiap langkah berhenti untuk ditinjau pengguna. Jangan commit/push kecuali diminta.

## 4. Langkah

### Langkah 0 - SELESAI (arsip, jangan dikerjakan ulang)
Keadaan awal saat rencana ditulis (sudah tidak berlaku): saat ini `docs/backend/` dan `CLAUDE.md` ada di `.gitignore` (hanya ada di folder kerja), sehingga merge tidak akan membawanya. Permintaan pengguna: keduanya harus ikut ke `Fullstack`. Dikerjakan di branch `backend` SEBELUM merge:
1. `git status` bersih dan semua kerja (SEO, job) sudah ter-commit dan ter-push. Jalankan `govulncheck ./...` di `backend/`.
2. Salin berkas rencana ini ke `docs/backend/PLAN-MERGE-FULLSTACK.md`.
3. **Periksa isi sebelum dilacak** (baca saja): `ls docs/backend`, lalu cari rahasia di `docs/backend/*` dan `CLAUDE.md` (kata sandi, token, `ACCESS_KEY`, `SECRET`, URL koneksi berisi kredensial). `backend/.env` memuat kredensial asli dan tidak boleh ikut (tetap di-ignore); pastikan `.env` tidak ter-track.
4. Ubah `.gitignore`: hapus baris `docs/backend/` dan `CLAUDE.md` (kepemilikan `.gitignore` ada pada pengguna; ini dilakukan atas permintaan eksplisit). Berkas lain di `docs/backend/` (mis. `01-kontrak-api.md`, `Schema.md` bila ada) ikut dilacak setelah diperiksa pada butir 3; bila ada yang ragu, tanyakan.
5. Perbarui kalimat di `CLAUDE.md` yang menyebut "`docs/backend/` dan `CLAUDE.md` ada di `.gitignore`" agar sesuai keadaan baru. Tambahkan pengecualian aturan docs untuk `PLAN-MERGE-FULLSTACK.md` hanya bila pengguna menghendaki.
6. `git add .gitignore CLAUDE.md docs/backend` dan **tanyakan pengguna** sebelum `git commit` dan `git push` (tidak ada commit/push otomatis).
Dampak ke merge: `Fullstack` tidak punya `CLAUDE.md` maupun `docs/backend/`, jadi kedua jalur masuk sebagai berkas baru tanpa konflik; `.gitignore` tidak lagi berubah di sisi backend.

### Langkah 1 - SELESAI (merge sudah terjadi lewat PR #9); hanya butir 4 yang tetap dijalankan sebagai verifikasi awal
Hanya butir 4 (go vet, go test -p 1, tes Node) yang masih perlu dijalankan sebagai garis dasar sebelum Langkah 2. Butir berikut arsip: 1. `git switch Fullstack` (tree bersih; berkas ter-ignore seperti `backend/.env` dan `deploy/` tetap ada di folder kerja; `CLAUDE.md` dan `docs/backend/` kini ter-track dan akan datang lewat merge).
2. `git merge backend` (tanpa fast-forward). Konflik yang diprediksi: **hanya `docs/database.md`** (backend menghapus ±31 baris di awal berkas, Fullstack menambah log aktivitas/aturan tanggal/judul). Penyelesaian: ambil versi Fullstack utuh (`git checkout --ours docs/database.md`; "ours" = Fullstack), tanpa menyunting teks. **Laporkan ke pengguna** bahwa penghapusan di sisi backend tampak tidak disengaja.
3. Hasil yang diharapkan: `backend/`, `CLAUDE.md`, dan `docs/backend/*` ikut masuk; `docs/_arsip/*` terhapus; `prototype/` utuh. Tidak ada konflik lain (`package.json`, `vite.config.ts`, `.env.example` root, `.gitignore` tidak diubah kedua sisi). Pastikan `git status` tidak menampilkan `backend/.env`.
4. Verifikasi pasca-merge (sebelum lanjut): `cd backend && go vet ./... && go test -p 1 ./... -count=1 -timeout 15m` (butuh Docker: PostgreSQL, Redis, RustFS menyala; jalankan di latar belakang) dan `node prototype/tests/*.test.cjs`. Semua harus tetap lulus.

### Langkah 2 - Audit docs Fullstack terhadap backend (baca saja, lalu lapor)
Docs di Fullstack lebih baru (`design`, `requirement`, `security`, `stack`, `database`). Bandingkan dengan perilaku backend dan buat daftar selisih. Yang sudah diketahui:
- Judul berita/dokumentasi **maksimum 70 karakter** (backend: `BatasJudul = 200` di `internal/admin/admin.go:56`). **Keputusan pengguna (2026-11-12): server menolak SEMUA judul lebih dari 70 karakter** (400, termasuk judul lama yang tidak diubah); tidak meniru aturan prototipe "judul lama boleh tetap". Tes wajib: 70 lolos, 71 ditolak, di Buat dan Simpan. Catat sebagai penyimpangan/penjelasan docs di `perubahan-kode.md`. Konsekuensi: `admin.js` perlu membuat draft berjudul >70 (data lama) tetap terbuka tetapi gagal disimpan sampai dipendekkan; sesuaikan pesan galat dan aturan di UI.
- `tanggal_publikasi` **tidak boleh setelah hari ini** → 400 (`tanggal()` di `internal/admin/teks.go:60` baru memeriksa format; tentukan zona waktu hari ini, Asia/Jakarta, dan tanyakan pengguna bila docs belum menyebutnya).
- `GET /api/tim` tanpa parameter `q` (pencarian di browser).
- Rute demo statis `/layanan/inference-solutions/{jejak-alumni,perencanaan-implan-gigi,document-intelligence}` dan nav "Contoh Aplikasi" (tanpa endpoint).
- `GET /api/admin/log-aktivitas`, tabel `log_aktivitas`, `/me` perlu `username` → **DITUNDA** (keputusan 2).
- `security.md` §1: dev-login ditandai `[KONFLIK, belum diputuskan]` terhadap "tidak ada halaman login sendiri" → **tanyakan pengguna** bila perlu diputuskan sebelum production.
Terapkan selisih backend yang jelas (judul 70 tegas, tanggal ≤ hari ini) dengan tes; yang ambigu ditanyakan dulu.

### Langkah 3 - Buat `web/` (sumber frontend yang disajikan Go) dari `prototype/`
`prototype/` dibiarkan utuh sebagai acuan (tesnya tetap jalan). Salin ke `web/` lalu sesuaikan lewat satu skrip yang bisa diulang (`web/` + `scripts/`), bukan sunting tangan:
- Tata letak: `web/{index,berita,dokumentasi,tim,tentang-kami,pelatihan,inference,404,jejak-alumni,perencanaan-implan-gigi,document-intelligence}.html`, `web/admin/index.html`, `web/assets/{css,js,fonts,img}`, `web/berita/_kerangka/index.html`, `web/dokumentasi/_kerangka/index.html`.
- Path aset `../assets/...` → `/assets/...` (wajib: halaman detail berada di kedalaman 2).
- Tautan `*.html` → URL bersih, baik di HTML (termasuk `<noscript>`) maupun di JS (`public.js`, `components.js`): `index.html`→`/`, `berita.html`→`/berita`, `dokumentasi.html`→`/dokumentasi`, `pelatihan.html`→`/layanan/pelatihan`, `inference.html`→`/layanan/inference-solutions` (+`#portofolio`), `tim.html`→`/tim`, `tentang-kami.html`→`/tentang-kami`, tiga halaman demo → `/layanan/inference-solutions/<nama>`, `berita-detail.html?slug=X`→`/berita/X`, `dokumentasi-detail.html?slug=X`→`/dokumentasi/X`.
- Halaman detail: slug dibaca dari `location.pathname`; data pertama dari `<script id="data-awal">` (disisipkan Go) dengan cadangan `fetch`.
- `berita-detail.html` / `dokumentasi-detail.html` menjadi kerangka SEO: sisipkan tepat sekali `<!--seo:head-->` (di `<head>`, hapus `<title>`/meta description statis), `<!--seo:konten-->` (di `<main>`), `<!--seo:data-->` (sebelum skrip). **Risiko:** pastikan tidak ada minifier yang menghapus komentar HTML.
- Buang dari produksi: seluruh `simulasi/` kecuali `keycloak.html` (lihat 4d), `mock-data.js`, tag `panel.js`/`panel.css`/`api-simulasi.js`, `data-sim*`, label "Simulasi".
- `web/404.html`: dari `prototype/public/404.html` (backend menambah `noindex`).
- Skrip penulis ulang harus menangani juga string tautan di JS dinamis (`public.js` baris ±16–21, 47, 208–209, 223, 234, 248, 258; `admin.js` `URL_LOGIN` dan `location.href = 'index.html'`), bukan hanya HTML statis. `perencanaan-implan-gigi.html` (67 KB, inline) dan tiga halaman demo perlu dicek isinya terhadap aturan penulisan ulang.
- Tes statis (Node + JSDOM, pola `prototype/tests`): tidak ada `href`/`src` ber-`.html` atau `../` (di HTML DAN di string JS), tidak ada rujukan `simulasi`, penanda kerangka tepat sekali, semua target tautan internal punya rute.

### Langkah 4 - Backend: menyajikan `web/`
Berkas baru `backend/internal/httpapi/static.go` (+ tes), kabel di `router.go`, `cmd/api/seo.go`, `config.go`:
a. **Tabel rute eksplisit** (bukan menebak ekstensi) → berkas di `STATIC_DIR`: `/`, `/berita`, `/dokumentasi`, `/tim`, `/tentang-kami`, `/layanan/pelatihan`, `/layanan/inference-solutions`, tiga rute demo. `GET/HEAD` saja; `Cache-Control: no-cache` + ETag (`http.ServeContent`).
b. `/assets/*` → `STATIC_DIR/assets` (tanpa listing direktori, tolak `..`); `/admin` dan `/admin/` → `admin/index.html` dengan `Cache-Control: no-store` (RP-05). Admin memakai hash-route, jadi tidak perlu fallback SPA lebih luas.
c. **NotFound:** path di luar `/api` → `404.html` dengan status 404 + `noindex` (RP-07); di bawah `/api` tetap JSON.
d. **Login dev:** `keycloak.html` dipertahankan di `/admin/masuk`, **hanya dilayani saat `DevLoginAktif`** (production → 404). Tiga tombol akun tiruan diganti satu formulir password yang memanggil `POST /api/auth/dev-login`; dua akun dev dibedakan oleh password (`ADMIN_PASSWORD` / `ADMIN_PASSWORD_2`), jadi tidak ada password di berkas statis. Setelah sukses, arahkan ke `/admin/`. Perbedaan role Direktur/Pegawai/tanpa-role dan layar 403 tidak bisa ditiru backend (role selalu `pegawai`); kodenya dibiarkan untuk Keycloak.
e. `STATIC_DIR` default diubah dari `../web/dist` (`config.go:34`) menjadi `../web` (tanpa langkah build); perbarui `config_test.go` dan `.env.example`. Rute statis dipasang bersama blok `d.SEO != nil` di `router.go` (rute `/berita/{slug}` hanya ada bila SEO aktif); tentukan urutannya eksplisit dan uji bahwa `/berita` (daftar) dan `/berita/{slug}` tidak bentrok. Loader kerangka SEO yang ada (`cmd/api/seo.go`, `seo.BuatKerangka`) dipakai apa adanya dan harus menerima `web/berita/_kerangka/index.html`.
f. Header respons: `docs/security.md` hanya mewajibkan CSRF, Origin, cookie, rate limit, dan `no-store` untuk `/admin` dan login (RP-05); CSP/nosniff TIDAK diwajibkan. Jangan menambah CSP tanpa bertanya (prototipe memakai skrip/gaya inline).

### Langkah 5 - Lapisan API nyata di frontend
Buat `web/assets/js/api.js` dengan nama metode yang sama seperti objek `api` tiruan (`daftarBerita`, `detailBerita`, `adminDaftar`, `adminTambah`, `adminUbah`, `adminPost`, `kunciAmbil`, dst.), sehingga halaman hanya perlu perubahan kecil. Adaptor di dalamnya:
- **Bentuk data:** backend camelCase (`tanggalPublikasi`, `dihapusPada`, `gambar:{key,url}`) → bentuk yang dibaca halaman prototipe (snake_case, `gambar`/`gambar[0].key`). Dokumentasi daftar: `gambarUtama`; detail: `gambar:[{url,alt}]`.
- **Sesi:** `fetch` dengan `credentials: 'same-origin'`; muat sesi lewat `GET /api/auth/me`; simpan `csrfToken` dan kirim `X-CSRF-Token` pada POST/PUT/DELETE; 401 → layar "Masuk" yang mengarah ke `/admin/masuk`.
- **`version`:** (dikoreksi saat dikerjakan; backend sumber kebenaran) hanya `PUT`, `post`, dan `retract` yang wajib membawa `{version}`; `restore`, `DELETE` (soft delete) dan `DELETE .../permanen` tanpa body. Tim: `DELETE` langsung permanen. Field `urutan` tim hanya-baca (backend menolak field itu), jadi dihapus dari form tim. **Langkah 5 SELESAI** (lihat `perubahan-kode.md`).
- **Galat:** bungkus ke `{status, error:{code,message}}`; kode backend `version_conflict` (bukan `conflict`), `status_tidak_valid` (bukan `bukan_draft`), `invalid_request` (bukan `validation_error`); 429 membaca header `Retry-After` (bukan `retry_after` di body); lock: `423 locked` → `{milik:'lain', nama}`, `423 lock_hilang` (bukan `410 lock_lost`).
- **Daftar admin:** `status=semua|draft|posted|trash` + `page/limit` (default 20) menggantikan `trash:boolean` tanpa paginasi; `dikunci_oleh` per baris tidak tersedia (info lock hanya di detail).
- **Gambar:** unggah dulu lewat `POST /api/admin/upload` (multipart `sumber` + `gambar`, batas `UPLOAD_MAX_MB`, default 2) lalu kirim `key`; dokumentasi mengirim `gambar:[{key,alt?}]` (bukan `gambar_daftar`); tampilkan memakai `url` dari backend.
- **Lock:** satu path `/{id}/lock`: pasang `POST`, heartbeat `PUT`, lepas `DELETE`. Interval heartbeat memakai `LockHeartbeatDetik` dari respons backend (heartbeat dihitung di bucket tulis admin, 30/menit per user); jangan ditetapkan di klien. Karena `dikunci_oleh` per baris tidak ada, tombol Post/Restore/Hapus tidak bisa dinonaktifkan di muka: tangani 423 saat diklik.
- **Boot admin asinkron:** `admin.js:634` memanggil `api.masuk()` sinkron; ganti menjadi `await GET /api/auth/me` (tampilkan keadaan memuat), dan `akun.role` dibaca dari `/me` (tanpa `username`; Log Aktivitas ditunda).
- **Paginasi daftar admin:** daftar prototipe memuat semua item; tambahkan kontrol halaman (`page/limit`) untuk berita/dokumentasi. Daftar testimoni dan tim tanpa paginasi; `q` tim hanya di klien.
- Hapus ketergantungan `SIMULASI`/event `simulasi:*` dari `public.js` dan `admin.js`; sembunyikan menu dan rute Log Aktivitas dengan satu penanda fitur.
- Tes Node (`node:assert` + JSDOM, `fetch` palsu): pemetaan bentuk data, header CSRF, `version`, pemetaan galat/423/429, alur upload.

### Langkah 6 - Verifikasi menyeluruh
- Otomatis: `go vet ./...`, `gofmt -l .`, `go test -p 1 ./... -count=1` (latar belakang), tes Node `web/` dan `prototype/`.
- Tes Go baru: rute statis, 404 HTML vs JSON, `/admin/masuk` hanya di development, tidak ada traversal di `/assets`, kerangka dari `web/` lolos `BuatKerangka`, HTML detail berisi title/OG/JSON-LD tanpa JavaScript.
- Manual (browser, dev: `APP_ENV=development`, `JOBS_ENABLED=false`, `ADMIN_PASSWORD` dan `ADMIN_PASSWORD_2` diisi, dan `ALLOWED_ORIGINS=http://localhost:8080` (sesuaikan port) di `backend/.env`, bila tidak semua tulis admin ditolak `cekOrigin`): (1) semua halaman publik termuat dari DB kosong dan berisi; (2) login dua akun di dua browser; (3) tambah/ubah/Post/Retract/hapus/pulihkan berita; (4) dokumentasi dengan beberapa gambar; (5) dua akun saling mengunci (423) dan 409 bila versi basi; (6) 429 tampil wajar; (7) berita yang di-Retract langsung 404 di `/berita/{slug}`; (8) `/sitemap.xml`, `/robots.txt`.
- Jangan memakai server dengan job aktif terhadap bucket dev.

### Langkah 7 - Pencatatan dan sisa pekerjaan
- `docs/backend/perubahan-kode.md`: satu entri per langkah major (merge; penyesuaian backend dari docs baru; `web/` + skrip; penyajian statis; lapisan API), memuat rujukan docs, hal yang melengkapi docs, dan penyimpangan (bentuk frontend = prototipe, bukan Astro/React).
- Sisa (di luar rencana ini): endpoint + tabel Log Aktivitas (`/grill-me`), Keycloak, migrasi data lama, penghapusan `src/` + `server/` + berkas root lama, migrasi bertahap ke Astro/React, berkas deploy baru (yang ada di `deploy/` masih untuk stack lama).

## 5. Berkas kunci

- Dibuat: `web/**` (hasil skrip dari `prototype/`), `scripts/` (skrip penyalin/penulis ulang), `web/assets/js/api.js`, `backend/internal/httpapi/static.go` + tes.
- Diubah: `backend/internal/httpapi/router.go`, `backend/internal/config/config.go`, `backend/cmd/api/seo.go`, `backend/.env.example`, validasi judul/tanggal di `backend/internal/admin/*`.
- Dipakai ulang: `backend/internal/seo` (kerangka, meta), `backend/internal/httpapi/seo.go` (jalur cache + 404 `noindex`), `backend/internal/httpapi/publik_cache.go` (`kirimPublik`, ETag), pola tes JSDOM di `prototype/tests/*.cjs`.
- Diubah/ditambah di langkah 0: `.gitignore` (hapus dua baris ignore), `CLAUDE.md` (satu kalimat), `docs/backend/PLAN-MERGE-FULLSTACK.md` (baru), serta pelacakan `docs/backend/*` yang sudah ada.
- Tidak disentuh: `prototype/` (acuan), `src/`, `server/`, `docs/**/*.md` kecuali berkas di `docs/backend/` yang disebut di atas.

## 6. Asumsi yang perlu dikonfirmasi pengguna

1. Login dev memakai satu formulir password di halaman `keycloak.html` (bukan memilih akun tanpa password), karena backend butuh password agar sesi nyata. Perbedaan role Direktur/Pegawai/403 hilang di mode dev.
2. Sumber frontend yang disajikan = folder baru `web/` (salinan yang disesuaikan), bukan menyunting `prototype/` langsung, supaya `prototype/` dan tesnya tetap utuh sebagai acuan.
3. Aturan docs baru (judul 70 karakter, tanggal tidak boleh di masa depan) diterapkan ke backend pada langkah 2. **Terkonfirmasi pengguna:** judul >70 ditolak seluruhnya, termasuk data lama.
4. **Terkonfirmasi pengguna:** zona waktu "hari ini" untuk batas tanggal publikasi = Asia/Jakarta (WIB, UTC+7). Aturan lain (batas panjang, dll.) mengikuti frontend prototipe; bila frontend tidak menetapkan batas suatu field, batas backend yang ada dipertahankan.
