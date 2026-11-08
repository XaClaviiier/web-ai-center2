# Catatan Prototype UBAYA AI Center

Dilebur ke `docs/` setelah tahap prototyping selesai. Rencana yang disetujui: cara membuka, struktur, dan simulasi ke `stack.md`; status halaman ke `requirement.md` §8; temuan ke dokumen sesuai topik dengan penanda `[terbuka]`; kode T-xx dibuang; butir yang sudah ada di `docs/` tidak disalin ulang; angka rinci T-29, T-30, T-32, dan T-33 diringkas di `design.md`.

## Cara membuka

Tanpa server. Klik dua kali:

- Publik: `prototype/public/index.html`
- Admin: `prototype/admin/index.html`, lalu tombol **Masuk**

Tombol kuning **Simulasi** di pojok kanan bawah memicu keadaan loading, kosong, error, 429, 401, 409, heartbeat gagal, dan item dikunci. Peta di footer butuh internet.

## Struktur

| Lokasi | Isi | Saat implementasi |
|---|---|---|
| `public/` | 13 halaman publik | Acuan visual dan alur |
| `admin/index.html` | Admin, rute hash | Acuan visual dan alur |
| `assets/css/` | `tokens`, `base`, `components`, `kalender`, `public`, `jejak-alumni`, `implan-gigi`, `document-intelligence`, `admin` | Acuan token dan komponen |
| `assets/js/` | `format`, `components`, `kalender`, `bola-neural`, `jejak-alumni`, `implan-gigi`, `document-intelligence`, `public`, `admin` | Acuan perilaku. `bola-neural.js` mandiri (hanya dimuat di `public/index.html`); angka waktu, ukuran, dan warna ada di `KONFIG` di bagian atasnya. `jejak-alumni.js` hanya dimuat di `public/jejak-alumni.html`, `implan-gigi.js` hanya di `public/perencanaan-implan-gigi.html`, dan `document-intelligence.js` hanya di `public/document-intelligence.html` |
| `assets/mock-data.js` | Semua data dummy | Dibuang, kecuali isi `konten` (lihat T-06) |
| `assets/fonts`, `assets/img` | Salinan dari `public/` | Foto contoh dibuang |
| `simulasi/` | Tiruan API, panel, layar Keycloak, `buat-ilustrasi-implan.cjs` (penulis HTML dan SVG demo implan; tidak dimuat halaman) | Dibuang |

## Kebutuhan terhadap halaman

| Kebutuhan (`requirement.md`) | Halaman prototype | Status |
|---|---|---|
| §3 Navbar, footer, breadcrumb, peta di footer | Semua halaman publik | Selesai |
| §3, §4.1 Beranda: hero, layanan, 3 dokumentasi, testimoni, 3 berita, kontak | `public/index.html` | Selesai |
| §4.1 Carousel testimoni otomatis 10 detik, 1 per slide | `public/index.html` | Selesai (dibuat sendiri, bukan Embla); autoplay hanya di perangkat ber-hover (T-33) |
| §4.2 Pelatihan: hero, Reguler coming soon, Khusus dengan alur, Workshop | `public/pelatihan.html` | Selesai; empat langkah alur konsultasi dikonfirmasi pemilik |
| §4.3 Inference: carousel pertanyaan, jenis, alur, 3 portofolio, kontak | `public/inference.html` | Sebagian: lima tahap alur dikonfirmasi pemilik; deskripsi dua jenis layanan masih data contoh |
| §4.3 Demo JejakAlumni: filter, KPI, grafik, simulasi AI, penjelasan | `public/jejak-alumni.html` | Selesai dengan data simulasi (T-38) |
| §4.3 Demo Perencanaan Implan Gigi: empat tahap, hasil, batasan | `public/perencanaan-implan-gigi.html` | Sebagian: ilustrasi belum ditinjau ahli (T-40) |
| §4.3 Demo DocNext: tiga skenario, sorotan sumber, perbandingan, deployment | `public/document-intelligence.html` | Selesai dengan dokumen sintetis (T-41) |
| §4.4 Dokumentasi: cari judul, 9 per halaman, urut terbaru | `public/dokumentasi.html`, `dokumentasi-detail.html` | Selesai |
| §4.5 Berita: cari judul, 9 per halaman, urut terbaru | `public/berita.html`, `berita-detail.html` | Selesai |
| §3, §4.6 Tim dengan pencarian | `public/tim.html` | Sebagian: enam bidang riset ditetapkan pemilik untuk prototype, perlu verifikasi sebelum produksi |
| §4.7 Tentang Kami | `public/tentang-kami.html` | Sebagian: isi diambil dari `src/data`, dokumen tidak merincinya |
| §3 Halaman 404 | `public/404.html`, juga untuk slug tidak dikenal | Selesai di sisi client; status HTTP 404 tidak bisa ditiru |
| §3 List dan detail kursus | Tidak dibuat | Belum: ditandai `[Ditunda]` |
| §5.1 Admin Berita: daftar, edit, trash, Post, Retract, Pulihkan, Hapus permanen | `admin/#/berita` | Selesai |
| §5.2 Admin Dokumentasi: 1–6 gambar, ganti, hapus, tambah | `admin/#/dokumentasi` | Selesai |
| §5.3 Admin Testimoni: foto opsional, status, trash | `admin/#/testimoni` | Selesai |
| §5.4 Admin Tim: tambah, edit, hapus permanen | `admin/#/tim` | Selesai |
| §6 Nama dan role di header | Admin | Selesai |
| §8 Keadaan kosong, error, konfirmasi, "sedang diedit oleh X", 409, 429 | Publik dan admin, lewat panel | Selesai |
| §8 HTML awal berisi title, OG, JSON-LD | Tidak bisa di HTML statis | Belum: dikerjakan Go (RP-12) |

## Yang disimulasikan

- **Login:** `simulasi/keycloak.html` menggantikan redirect Keycloak. Akun dikirim lewat `?akun=`. Tidak ada cookie, session, atau CSRF sungguhan.
- **API:** `simulasi/api-simulasi.js` memberi jeda 300 ms dan galat tiruan.
- **Penyimpanan:** data hanya di memori. Refresh mengembalikan data awal. Perubahan di admin tidak terlihat di publik.
- **Lock:** kunci pengguna lain berasal dari data awal `kunci`, bukan Redis. Kunci tidak bisa berpindah karena hanya ada satu browser.
- **Heartbeat:** berjalan tiap 60 detik sungguhan, tetapi hanya memanggil tiruan. Kegagalan dipicu dari panel.
- **Konflik 409:** dipicu dari panel dengan menaikkan `version` seolah pengguna lain menyimpan.
- **Slug:** dibuat dari judul hanya untuk simulasi. Aturannya masih `[terbuka]`.
- **Upload:** gambar tidak dikirim ke mana pun; pratinjau memakai URL lokal browser. Validasi isi file dan 2 MB benar-benar berjalan di browser.
- **Validasi server:** batas judul 70 karakter ditiru di `simulasi/api-simulasi.js` (400 `validation_error`), bukan server sungguhan (T-34).
- **Rate limit:** 429 dipicu dari panel, tidak menghitung request.
- **Tidak disimulasikan:** penghapusan otomatis trash setelah 30 hari, pembersihan file yatim, cache, ETag.

## Temuan rancangan

Keputusan produk di bawah ini sudah dipindahkan ke `docs/`: T-01, T-04, T-07 sampai T-12, T-15 (rencana), T-18 sampai T-26, T-28 sampai T-31, T-33 butir 1, T-34, T-35, T-37 (rencana), T-38, dan T-40. Baris-barisnya tetap dicatat di sini sebagai riwayat. T-03 ditandai sebagai konflik di `docs/security.md` dan `docs/stack.md`. Sisanya adalah catatan prototype atau menunggu keputusan.

| ID | Dokumen | Temuan |
|---|---|---|
| T-01 | `design.md` | Hanya ada nilai warna dan font, tanpa nama token. Jarak, radius, bayangan, skala teks hanya "sesuaikan". Nama dan nilainya dibuat di `tokens.css`. |
| T-02 | `design.md` §2 | Referensi visual tidak bisa ditiru. Dari 4 situs yang dicoba: `symboolic.com` hanya terbaca struktur menu, `hub.orca.id` hanya teks judul, `brand.vettvangur.is` hanya palet tanpa perilaku animasi, `lppm.ubaya.ac.id/aicenter` gagal (HTTP 468). Situs lain tidak dicoba. Hero dan pergantian section Beranda adalah tafsiran sendiri. |
| T-03 | `security.md` §1 | Masih menyebut login password development, sedangkan keputusan terbaru menyatakan aplikasi tidak punya halaman login sendiri. Tidak dibuat di prototype. |
| T-04 | `database.md` §2, §3 | `tanggal_publikasi` masih bertipe `date`. Keputusan terbaru: backend menyimpan dan menampilkan `DD-MM-YYYY`. Tipe kolom perlu disesuaikan; pengurutan di database perlu mengubah string menjadi tanggal. |
| T-05 | `stack.md` §3 | Prototype memakai `?slug=`, rute akhir `/berita/:slug`. Batas `file://`, hanya di prototype. |
| T-06 | Aturan data mock | `konten` ikut di `mock-data.js` sesuai aturan satu file, padahal di implementasi konten hardcode tetap dipakai dan tidak dibuang. |
| T-07 | `design.md` §3 | `#0000006B` sebagai teks sekunder kontrasnya sekitar 3:1. Dipakai hanya untuk overlay; teks sekunder memakai `#0A2240` opasitas 72%. (Q8) |
| T-08 | `design.md` §3 | `#0674FD` di atas putih kontrasnya sekitar 4,3:1, sehingga tidak dipakai untuk teks biasa. `#FF6B2C` juga kurang untuk teks, sehingga pesan error memakai warna turunan `#B23F0D`, yang tidak ada di `design.md`. |
| T-09 | `database.md` §4 | Tidak mengatur kapan "sesi edit berakhir" muncul, batas pengguna diam, atau nasib isian saat kunci hilang. Di prototype: muncul saat heartbeat gagal, Simpan nonaktif, isian tetap ada. (Q9) |
| T-10 | `database.md` §4 | Tidak mengatur Post dan Hapus pada item yang dikunci. Di prototype keduanya nonaktif. (Q10) |
| T-11 | `stack.md` §9.4 | `GET /api/tim` tidak punya `q`, padahal `requirement.md` §3 meminta pencarian. Di prototype dilakukan di browser, pada nama, jabatan, dan bidang. (Q11) |
| T-12 | `requirement.md` §5.4 | Field form Tim tidak menyebut `urutan`, padahal `database.md` §8 memakainya. Ditambah sebagai kolom angka. (Q12) |
| T-13 | `requirement.md` §5 | Tidak ada dashboard, pencarian, atau pagination di daftar admin, jadi daftar bisa panjang. Tidak dibuat. (Q13) |
| T-14 | `requirement.md` §5 | Tanggal masuk trash dan sisa hari tidak ditampilkan, walaupun item terhapus setelah 30 hari. Tidak dibuat. (Q14) |
| T-15 | `requirement.md` §4, §5 | Dokumen tidak merinci konfirmasi Post dan Retract. Mengikuti aturan yang ada, hanya hapus dan logout yang memakai konfirmasi, padahal Retract langsung menurunkan konten dari publik. Diputuskan pemilik: Retract memakai konfirmasi (direncanakan, belum dikerjakan di prototype). |
| T-16 | Konten | Enam bidang riset Tim di `assets/mock-data.js` ditetapkan pemilik untuk prototype, belum diverifikasi sebagai keahlian pribadi untuk produksi. Belum ada sumber untuk alur konsultasi dan deskripsi IT Service Management serta Data Science. Handle Instagram masih placeholder di `src/data/kontak.ts`. |
| T-17 | `requirement.md` §4.1 | Tidak menyebut isi tombol hero. Dipakai "Diskusikan Kebutuhan AI" (WhatsApp) dan "Lihat Layanan"; sudah dicatat di `design.md` §2. |
| T-18 | `database.md` §5.1 | Pola nama objek RustFS untuk Dokumentasi belum ada. Tidak memengaruhi prototype. |
| T-19 | Keputusan Q16 | **Q16 diganti.** Input tanggal kini kolom teks `DD-MM-YYYY` ditambah kalender, bukan kolom teks saja. |
| T-20 | `stack.md` §2 | Stack akhir belum menyebut komponen pemilih tanggal. Perlu dipastikan apakah komponen kalender shadcn/ui bisa memenuhi semua aturan ini (format `DD-MM-YYYY`, bahasa Indonesia, mulai Senin, masa depan nonaktif) atau `assets/js/kalender.js` diterjemahkan langsung. |
| T-21 | `requirement.md`, `security.md` §7 | Aturan "tanggal publikasi tidak boleh setelah hari ini" dan nilai awal hari ini belum tertulis di dokumen. Di prototype hanya diperiksa di browser. Backend juga perlu menolaknya, dan perlu menetapkan zona waktu acuan: tanggal lokal perangkat bisa berbeda dari tanggal server. Rancangan juga jadi tidak mendukung posting terjadwal. |
| T-22 | `database.md` §3 | Sesuai dugaan, `tanggal_publikasi` hanya ada di Berita dan Dokumentasi. Testimoni dan Tim tidak mendapat kalender. |
| T-23 | `requirement.md` §5.1–§5.3 | Dokumen menulis tombol "Pulihkan" dan "Hapus permanen". Revisi admin meminta "Restore" dan "Hapus Permanen", dan prototype mengikuti revisi. Dokumen perlu diselaraskan. |
| T-24 | `requirement.md` §5 | Revisi admin menyebut aksi Edit, Hapus, Post. Sesuai keputusan K1, itu hanya untuk Draft; item Posted tetap hanya punya Retract. |
| T-25 | Revisi admin | Sesuai keputusan K2, saat item dikunci Hapus dan Post nonaktif, tetapi Edit berganti "Lihat" yang tetap aktif untuk mode baca saja. Tabel Trash tanpa kolom Status (K3), karena Restore selalu mengembalikan ke draft. |
| T-26 | Data lama | Item yang sudah tersimpan dengan tanggal masa depan tetap tampil apa adanya di form edit, tetapi tidak bisa disimpan sampai tanggalnya diperbaiki. Tiga tanggal draft di mock diubah ke tanggal lampau (K4). Perlu aturan untuk data serupa saat migrasi. |
| T-27 | Konten | Foto testimoni di mock adalah avatar SVG netral, bukan foto orang (K5). Avatar pada testimoni Posted ikut tampil di carousel Beranda lewat kode publik yang sudah ada. Foto asli belum tersedia. |
| T-28 | `requirement.md` §5.4 | Tabel Tim tetap memakai kolom Foto (K6), walaupun dokumen tidak merinci kolom tabel. |
| T-29 | `design.md` §1 | Hero Beranda kini bidang navy dengan "bola neural" (canvas 2D): 300 titik berbentuk bola, berputar pelan (75 dtk per putaran), bernapas, sesekali satu node menyala mint beserta garis tipis ke 3 tetangganya (maks. 2 bersamaan), dan titik dekat kursor tertarik sedikit (tidak aktif di atas tautan/CTA). Bola adalah lapisan latar absolut; tinggi hero mengikuti teks (±642 px di 1440, ±546 px di 768, ±513 px di 390). Posisi dan skala bola diatur di CSS lewat `--bola-x/-y`, `--bola-skala`, `--bola-regang-y`, `--bola-magnet` (dibaca JS saat ukuran berubah). ≥1024px: kanvas 62% kanan, bola besar diregangkan vertikal 1,3× sehingga terpotong atas–bawah, teks 55%. 768–1023px: kanvas 85% kanan, opacity 0,8, teks 70%. <768px: kanvas seluruh hero, bola di tengah di belakang headline, opacity 0,6. Scrim statis `.hero::after` (gradien navy: vertikal di mobile; dari kiri + radial kiri-bawah di ≥768px) meredam partikel di belakang teks. Teks di depan; canvas `pointer-events: none`. Opacity efektif partikel terukur: paragraf ≤0,18 dan CTA ≤0,12 di semua lebar; h1 ≤0,37 di ≥768px dan 0,56 di mobile (batas 0,55, dari node mint yang sesekali lewat di belakang headline; selisih 0,01 diterima pemilik). Terbuka: uji CPU 6× di headless memicu fallback tingkat 1 (p90 ±4,8 ms); menunggu pengukuran di perangkat nyata. Foto `hero-1-1600.webp` dan `hero.gambar` di mock tidak dipakai lagi. Acuan orca belum terbaca. Animasi "forward pass" sebelumnya dihapus. |
| T-30 | `design.md` §1 | Perangkat tanpa hover/pointer halus: 200 titik, 30 fps, DPR maks. 1,5, tanpa interaksi. Backing store kanvas dibatasi 1,5 juta piksel (DPR turun otomatis). `prefers-reduced-motion`: bola diam, tanpa magnet, hanya kedipan opacity yang jarang; kanvas tidak menggambar di antaranya. Jika p90 waktu kerja per frame >6 ms selama 120 frame, titik dikurangi separuh, lalu bola jadi diam. Animasi berhenti saat hero di luar layar atau tab tersembunyi. Ukuran script 11 KB (4,3 KB gzip). Di produksi cukup dibungkus island Astro; warna bisa diberikan lewat `opsi.warna` jika nama token berubah. Belum diukur di perangkat nyata. |
| T-31 | WCAG 2.2.2 | Atas keputusan pemilik, tidak ada tombol jeda untuk animasi hero yang berjalan lebih dari 5 detik. Ini penyimpangan dari WCAG 2.2.2 (Pause, Stop, Hide) yang disengaja; satu-satunya keringanan adalah `prefers-reduced-motion`. |
| T-32 | QC hero | Gutter samping hero Beranda, Pelatihan, dan Inference dinaikkan menjadi `clamp(--jarak-5, 5vw, --jarak-8)` per sisi (24 px di 375–390, ±41 px di 820, ±56 px di 1120; tidak berubah di ≥1240) lewat `.hero__isi, .sampul .wadah` di `public.css`. `.wadah` global, navbar, breadcrumb, dan section lain tidak diubah, sehingga di bawah ±1240 px teks hero Pelatihan/Inference menjorok 4–36 px lebih dalam dari breadcrumb dan section di bawahnya (diterima pemilik). Jumlah baris headline dan paragraf tidak berubah. |
| T-33 | QC Beranda & footer | (1) Tombol jeda dihapus dari komponen `carousel()` (`components.js`), sehingga hilang di testimoni Beranda dan pertanyaan Inference. Agar tetap memenuhi WCAG 2.2.2, autoplay hanya aktif di perangkat ber-hover (`(hover: hover)`) dan tetap berhenti saat disorot/difokus; layar sentuh dan `prefers-reduced-motion` manual lewat panah. (2) Padding `.lembar` (section Beranda) menjadi `clamp(--jarak-7, 6vw, --jarak-8)`: jarak antar-konten section 200 → 120 px di desktop, ±99 di 1024, ±74 di 820, 72 di mobile; kartu hero ke judul Layanan 124 → 84 / 60 px. (3) Tautan WhatsApp, email, dan Instagram di footer (`.footer__kontak`) tanpa garis bawah, putih, bobot 500, dengan ikon (`pesan`, `surat`, `kamera` baru); garis bawah muncul saat hover/fokus. Tautan footer lain tetap bergaris bawah. Teks yang tampil hanya nomor, email, dan `@aicenter_ubaya`; kata "WhatsApp" dan "Instagram" tetap ada untuk pembaca layar (`.sr-only`) supaya nama tautan tidak hilang. Ikon `jeda`/`putar` dihapus karena tidak dipakai lagi. |
| T-34 | Batas judul | Judul Berita dan Dokumentasi dibatasi 70 karakter di prototype saja (`MAKS_JUDUL` di `assets/js/admin.js`, cermin di `simulasi/api-simulasi.js`). Penghitung `n/70` tampil di bawah input; lebih dari 70 memberi `aria-invalid` dan galat, Simpan tertahan. Dihitung dengan `.length` pada nilai yang di-trim (spasi dihitung, emoji = 2), sama dengan `maxlength` HTML dan `bacaTeks` di server. `maxlength` hanya dipasang bila nilai awal ≤70, sehingga judul lama tidak terpotong. Paste yang melebihi batas terpotong oleh `maxlength` dan menampilkan pesan jumlah karakter yang dipotong (keputusan pemilik, opsi B); teks yang diseret (drop) belum memberi pesan. Judul lama >70 tetap bisa disimpan selama tidak diubah, baik di form maupun di simulasi server. Bukan batas untuk judul modul, hero, inference, nama layanan, atau nama tim. Belum diterapkan di `src/pages/Admin.tsx` dan `server/src/lib/tulis.ts` (batas keduanya masih 200 untuk berita, dokumentasi, dan kursus), dan `database.md` §2–§3 belum menyebut batas. Sebelum diterapkan di produksi, periksa judul live yang >70 karakter (seed dan mock maksimal 59). Judul kursus belum ada di prototype. |
| T-35 | `requirement.md` §5 | **Perlu didiskusikan sebelum backend karena berdampak ke API.** Konten Posted harus di-Retract sebelum diedit (dipertahankan, keputusan pemilik), sehingga memperbaiki satu typo membuat halaman 404 sementara, termasuk bagi tautan yang sudah dibagikan. |
| T-36 | Kontak | Nomor WhatsApp resmi `0895634222240` dikonfirmasi pemilik; tautan `whatsapp_url` diperbaiki ke `62895634222240`. Komentar `src/data/kontak.ts:5` masih menyebut `6289563422240`, tetapi berada di luar prototype. |
| T-37 | Form admin | Meninggalkan form yang isinya berubah memunculkan dialog "Tinggalkan halaman ini?" (sidebar, logo, Batal, tombol Back); Keluar menambahkan peringatan di dialognya; tutup tab dan refresh memakai dialog bawaan browser (keputusan pemilik). Perubahan dihitung dengan membandingkan nilai form terhadap nilai awal. Meninggalkan form tanpa perubahan, atau menekan Simpan pada form edit tanpa perubahan, menampilkan toast "Tidak ada perubahan yang disimpan." dan Simpan tidak memanggil API, sehingga `version` tidak naik. Simpan berhasil menampilkan "Tersimpan sebagai draft.", "Anggota tim ditambahkan.", atau "Perubahan tersimpan.". Di prototype pembatalan mengembalikan hash dan menambah satu entri riwayat; di produksi gunakan blocker router. Panel simulasi tetap merender ulang tanpa dialog. |
| T-38 | `requirement.md` §4.3 | Portofolio kini JejakAlumni, Perencanaan Implan Gigi, dan DocNext, semuanya menaut ke demo masing-masing (keputusan pemilik). Kartu Cacat Las keluar lebih dulu, lalu kartu Algae Finder digantikan DocNext (T-41); album dokumentasi Algae Finder tetap ada. Status JejakAlumni: "Demo Konsep · Data Simulasi". Data JejakAlumni ada di `konten.jejak_alumni` agar ikut terbawa ke implementasi (T-06). Jumlah orang adalah sumber kebenaran, persentase dihitung saat render; jumlah per jurusan, median per jurusan, tema 2024, dan ringkasan 2024 adalah simulasi tambahan, bukan dari spesifikasi awal. Median per jurusan tidak bisa dibuktikan secara matematis. Halaman demo tidak memuat `api-simulasi.js` dan panel kuning karena tidak memakai API dan bisa rancu dengan label "Data Simulasi". Filter tidak disimpan di URL: refresh kembali ke 2025 dan Semua jurusan. `jejak-alumni-preview.webp` adalah tangkapan layar dasbor yang dipadatkan (tanpa tab, filter, dan panel AI) dengan badge status tertanam, sehingga perlu diambil ulang bila dasbor atau label berubah. |
| T-39 | Aset implan | Gambar implan lama dihapus dari prototype. Kartu dan album Implan Gigi memakai `implan-gigi-konseptual.svg`, ilustrasi buatan sendiri; berita contoh yang memakainya beralih ke `ruang-diskusi.jpg`. Rincian disimpan di luar repo. |
| T-40 | `requirement.md` §4.3 | Demo Perencanaan Implan Gigi. Kartunya kini tautan; status memakai "Demo Penelitian · Visualisasi Konseptual · Nonklinis" (keputusan pemilik). Penulis tidak dicantumkan atas keputusan pemilik. Rincian penelitian dan asal dataset disimpan di luar repo. Penampang melintang yang bergerak sepanjang lengkung adalah penalaran, bukan fakta dari abstrak. \[Ditandai\] Tinjauan ahli: arah slice, posisi kanal, bentuk tulang alveolar, penempatan implan, dan kesan jarak. HTML dan SVG halaman ditulis oleh `simulasi/buat-ilustrasi-implan.cjs`, jadi perubahan ilustrasi dilakukan di sana lalu dijalankan ulang; `assets/img/implan-gigi-konseptual.svg` terpisah dan disunting langsung. |
| T-41 | `requirement.md` §4.3 | Demo DocNext menggantikan kartu Algae Finder. Album dokumentasi dan `algae-finder.jpg` dipertahankan, begitu juga rujukan di luar prototype. Nama kerja ProcureMatch AI dan AI Document Intelligence ditarik karena berisiko bentrok; "Document Intelligence" tetap dipakai sebagai istilah. Pengecekan web "DocNext" bukan clearance merek. Di Skenario 3, kode pada invoice sengaja berbeda dari PO agar keraguan identitas barang nyata. Nilai 50 unit dan Rp85.000 di Skenario 3 serta nama barang Skenario 1 dan 2 adalah tambahan prototype. Label JejakAlumni menjadi "Demo Konsep · Data Simulasi" dan `jejak-alumni-preview.webp` diambil ulang. |

## Pemeriksaan

\[Ditandai\] Status bagian ini belum jelas dan tidak ikut dilebur ke `docs/` sampai diputuskan.

Diuji otomatis di Chrome headless lewat `file://`, semua lolos:

- **Pengujian umum (104 cek):** 10 halaman publik dan 10 rute admin pada 1440 px dan 390 px (tanpa galat konsol, tanpa scroll horizontal, gambar termuat), tautan lokal, pencarian, pagination, urutan tanggal, semua keadaan panel, validasi form dan tanggal, validasi file, Simpan nonaktif, teks berbahaya tampil sebagai teks, soft lock, 409, heartbeat, 401, 403, Post, Retract, trash, keluar.
- **Revisi admin (60 cek):** 7 tabel pada 1440 px dan 390 px (halaman tidak bergulir ke samping, kolom Aksi terjangkau, `caption`, `th scope`), urutan, aksi per status, foto dan placeholder termasuk gambar gagal dimuat, kalender dengan jam palsu 01.30 WIB (hari ini tetap 06-11-2026 walaupun UTC masih tanggal 5), tanggal masa depan ditolak lewat klik dan ketikan, keyboard, nilai default dan nilai tersimpan.
- **JejakAlumni:** lima lebar layar (1440, 1024, 768, 390, 360 px) tanpa scroll horizontal; 14 kombinasi filter cocok dengan data, status berjumlah 100%, tanpa nilai kosong; klik kartu, refresh, dan tautan kembali ke `inference.html#portofolio`; 11 halaman publik tanpa galat konsol.
- **Perencanaan Implan Gigi:** enam lebar layar (1440, 1280, 1024, 768, 390, 360 px) tanpa scroll horizontal di keempat tahap; semua toggle dan pilihan mengubah ilustrasi serta `aria-label`, dan bisa dikembalikan; Enter pada Berikutnya memindahkan fokus ke judul tahap; panah keyboard pada radio; klik kartu, refresh, dan tautan kembali; disclaimer tampil sebelum demo; tanpa input data; satu-satunya angka mm adalah MAE yang dilaporkan.

Belum diperiksa: pembaca layar, browser selain Chrome, perangkat sungguhan, dan isi peta Google (iframe ada, isinya tidak diverifikasi). Kontras dihitung manual untuk warna utama, bukan audit menyeluruh.
