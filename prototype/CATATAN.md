# Catatan Prototype UBAYA AI Center

## Penyempurnaan filter Log Aktivitas — 11 November 2026

Disetujui setelah audit baca-saja: Terapkan dan Reset filter berdampingan, sejajar dengan empat kontrol pada desktop mulai 1280 px. Kolom tombol mengikuti lebar isinya; pada layar lebih kecil kelompok tombol turun baris dan boleh membungkus bila diperlukan. Tinggi kontrol/tombol 44 px. Dua select log memakai padding kanan 44 px dan chevron CSS 16 px pada jarak 16 px dari tepi kanan; elemen select tetap native, panah native dipulihkan dalam forced colors. Tipografi, handler, label Simulasi, dan komponen admin lain dipertahankan.

Perubahan sumber hanya `prototype/assets/css/admin.css`, `docs/design.md`, dan catatan ini. Verifikasi memakai suite Log Aktivitas existing dengan tambahan pengukuran browser sementara di memori, tanpa mengubah file tes. API/UI regresi lulus; Chrome headless pada 1440, 1280, 1024, 820, 390, dan 320 px mengonfirmasi tinggi 44 px, tombol berdampingan dengan gap, padding/posisi panah, tanpa overflow halaman. Seluruh kontrol satu baris dan sejajar pada 1440/1280 px; filter/reset, pagination dan batal hapus tetap lulus pada enam ukuran. Screenshot desktop 1440 dan mobile 320 px diperiksa visual; bukti `prototype/tests/.browser-check/log-aktivitas-{1440,1280,1024,820,390,320}.png`. Pengujian terbatas pada Chrome headless, bukan perangkat fisik. Tidak melakukan operasi Git yang mengubah state; `/docs/_arsip/` diabaikan.

## Log Aktivitas — 11 November 2026

Disetujui setelah audit baca-saja. Menu admin `#/log-aktivitas` menampilkan tabel baca-saja Tanggal & Waktu/Aktivitas/Username, terbaru dahulu, filter username/jenis/rentang hari Jakarta, pagination 10 entri, loading/error/retry serta dua empty state. Label Simulasi dipertahankan dan batas penyimpanan dijelaskan langsung pada halaman. Tidak ada log contoh pada awal halaman.

Log dibuat di lapisan API tiruan setelah mutasi berhasil: tambah, ubah, Post, Retract, Restore, hapus ke trash dan hapus permanen pada Berita/Dokumentasi/Testimoni; tambah/ubah/hapus permanen Tim. Username dari akun aktif contoh, bukan payload form; ID dan snapshot label tetap terbaca setelah item dihapus. Admin API memeriksa akun/role, respons akun/log disalin, dan penghapusan item yang tidak ada ditolak tanpa entri log. Deskripsi tidak menyalin isi berita/testimoni atau payload autentikasi.

**Batas:** hanya memori instance halaman; refresh/logout yang memuat ulang halaman mengosongkan log dan mengembalikan data awal. Tidak dibagikan antar-tab atau pengguna. Identitas pilihan URL dan waktu perangkat tidak diverifikasi server. Ini simulasi perilaku, bukan audit log persisten yang aman. Tidak menggunakan localStorage. Rancangan backend transaksional, username sesi, serta hak akses log dicatat sebagai rencana di lima dokumen aktif, belum diterapkan pada backend.

Verifikasi: `node prototype/tests/activity-log.test.cjs --browser` menguji seluruh mutasi empat modul, atribusi username/sub, respons salinan, galat 400/401/403/404/409/423/429/500 tanpa tambahan log, urutan/pagination/filter gabungan, batas 17:00 UTC menjadi hari berikutnya di Jakarta, pembatalan hapus, teks menyerupai HTML yang tidak dieksekusi, empty/error/retry serta halaman baru kosong. JSDOM memakai shim dialog buka/tutup; Chrome headless juga memeriksa pembatalan dialog native, filter dan pagination pada 1440/820/390 px. Halaman tidak overflow horizontal; tabel 390 px digulir di dalam region sendiri. Tidak ada exception JavaScript yang terdeteksi pada alur browser tersebut. Bukti screenshot: `prototype/tests/.browser-check/log-aktivitas-{1440,820,390}.png`. Hasil ini terbatas pada Chrome headless, bukan pengujian perangkat fisik/screen reader.

Regresi existing lulus: `node prototype/tests/navbar.test.cjs` dan `node prototype/tests/audit-refinements.test.cjs`. Perubahan sumber terbatas pada `/prototype/` dan `/docs/`, tanpa dependency baru, perubahan backend/production atau operasi Git yang mengubah state. `/docs/_arsip/` diabaikan.

## HERO-03 A, SPHERE-03 A, CTA-03 A - Tinggi hero dan ukuran tombol

Disetujui pemilik: layar pertama mencakup navbar, margin dan hero; sphere desktop dibatasi referensi scene 642 px; tombol kontak diperbesar. Persetujuan tambahan SPHERE-03-FIX A mengizinkan pemisahan radius dari kanvas pada `assets/js/bola-neural.js`. File implementasi yang diubah adalah `assets/css/public.css` dan `assets/js/bola-neural.js`.

Chrome headless memeriksa Beranda dan delapan CTA pada lebar 375, 430, 768, 1280, dan 1440 px, dengan viewport setinggi 900 px (40 kombinasi CTA). Section berikutnya pada Beranda mulai tepat di 900 px. Hero terukur 800 px pada ketiga viewport mobile/tablet dan 788 px pada kedua desktop. Padding vertikal tetap 48/72 px dan H1 tetap 36–64 px. Konten hero tetap berada dalam container pada viewport pendek 375×500 serta 1440×500. Tidak terdeteksi overflow horizontal atau exception JavaScript.

Tombol WhatsApp/email memiliki tinggi sama: 52 px pada lebar ≤768 px, 56 px pada desktop; font 16 px, padding horizontal 24 px, ikon 20 px, gap ikon 10 px dan gap antartombol 16 px. Background terhitung identik di seluruh CTA. Kolom CTA Tentang Kami tetap sejajar di desktop. Tab/Enter dan klik mouse pada kedua tautan lulus pada delapan CTA di lima viewport; emulasi tap lulus pada Beranda/Tentang Kami di 375 dan 430 px. Aktivasi kontak dicegat tanpa membuka percakapan atau aplikasi email. Kedua tes existing navbar dan audit-refinements lulus tanpa perubahan file tes.

SPHERE-03-FIX A selesai: kanvas kembali memenuhi seluruh tinggi hero, sementara radius membaca batas CSS desktop 642 px. Validasi browser pada lima viewport mengonfirmasi tepi kanvas berimpit dengan hero dan section berikutnya tetap mulai pada 900 px. Radius dasar desktop terukur 385,2 px pada tinggi viewport 900 maupun 1080 px; mobile/tablet tetap memakai perhitungan tanpa batas desktop. Screenshot desktop serta tablet diperiksa: pemotongan titik pada batas scene internal sudah hilang. Clipping dekoratif yang memang mengikuti batas luar hero tetap sesuai desain sebelumnya.

Uji terkontrol lima skenario lulus: radius pada kanvas tinggi, adaptasi bidang kecil, serta fallback tanpa batas. Browser mengonfirmasi tarikan magnet dan kemunculan mint links tetap aktif, dengan maksimal garis di bawah batas delapan yang disetujui; reduced motion menonaktifkan gerak/interaksi. Pemeriksaan sintaks JS serta kedua tes existing dijalankan ulang dan lulus. Screenshot akhir: `prototype/tests/.browser-check/sphere-full-canvas-1440.png`, `sphere-full-canvas-768.png`, `sphere-full-canvas-375.png`, `cta-large-tentang-kami-1440.png`, dan `cta-large-tim-375.png`. Screenshot `hero-full-*` terdahulu menunjukkan tahap sementara sebelum pemisahan radius/kanvas.

Tidak menjalankan Git/GitHub, menambah dependency, atau mengubah kode production/backend. Emulasi touch bukan pengujian perangkat fisik.

## CTA-ABOUT-01 A dan TYPO-01 B - Layout serta tipografi terarah

Disetujui pemilik: CTA Tentang Kami horizontal pada desktop dan penyesuaian TYPO-01A–F. Teks CTA berada di kiri, kelompok tombol di kanan, dengan pusat vertikal sejajar mulai 1024 px. Tablet/mobile tetap vertikal. Gradient, radius, warna, ikon, dan tujuan kontak mengikuti komponen existing. File implementasi yang diubah hanya `assets/css/public.css` dan `assets/css/document-intelligence.css`; spesifikasi ada di `docs/design.md`.

| Komponen | Sebelum (375 / 768 / 1440 px) | Sesudah (375 / 768 / 1440 px) |
|---|---|---|
| H1 hero Beranda/Pelatihan/Inference | 40.5 / 57.79 / 80 px | 36 / 47.04 / 64 px |
| H1 Tim/Galeri/Tentang Kami | 40.5 / 57.79 / 80 px | 32 / 42.34 / 56 px |
| Tautan navbar | 16 / 16 / 14 px | 16 / 16 / 15 px |
| Portofolio: judul/deskripsi/badge pada tablet | 18 / 14 / 12 px, tiga kolom | 20 / 16 / 13 px, dua kolom |
| Dokumen sumber DocNext | 14 / 14 / 14 px | 16 / 15 / 15 px |
| H1 detail berita/dokumentasi | 32.15 / 42.37 / 56 px | 32 / 40.06 / 48 px |

Validasi Chrome headless memeriksa 13 halaman publik pada lebar 375, 430, 768, 1280, dan 1440 px (65 kombinasi). Ukuran terhitung memenuhi rentang yang disetujui, tanpa overflow horizontal atau exception JavaScript terdeteksi. Kolom CTA Tentang Kami sejajar di tengah pada desktop dan tidak bertabrakan; batas 1024 px juga diperiksa. Kartu portofolio memakai 1/2/3 kolom sesuai viewport tanpa tumpang tindih. Body 16 px, artikel 19 px, metadata 14 px dan footer 16 px terkonfirmasi tetap; tag kecil DocNext tetap 12 px.

Pemeriksaan tambahan: 13 halaman publik tidak mengalami overflow horizontal pada lebar 320 px. Navbar dengan teks 15 px tetap berganti ringkas/lebar dan kembali ke atas tanpa kehilangan fokus. Tab/Enter pada kedua kontak CTA Tentang Kami di 375 serta 1440 px mempertahankan indikator fokus dan tujuan WhatsApp/email; aktivasi dicegat sebelum membuka layanan eksternal. Kedua tes existing `node prototype/tests/navbar.test.cjs` dan `node prototype/tests/audit-refinements.test.cjs` lulus tanpa perubahan file tes.

Screenshot yang diperiksa: `prototype/tests/.browser-check/refine-about-1440.png`, `refine-about-375.png`, `refine-index-1440.png`, `refine-portfolio-768.png`, dan `refine-docnext-375.png`. Pengujian ini bukan sertifikasi WCAG. Zoom native 200%, pembesaran teks serta text spacing menyeluruh, screen reader, dan perangkat fisik belum dinyatakan terverifikasi. Uji 320 px membuktikan reflow halaman yang diperiksa, bukan seluruh perilaku zoom native. Tidak menjalankan Git/GitHub, menambah dependency, atau mengubah kode production/backend.

## UI-04 - Outline Portofolio pada navigasi lintas halaman

Disetujui pemilik: aturan lokal `#portofolio:focus { outline: none; }` menggantikan pengecualian yang hanya berlaku ketika penanda klik pointer tersedia. Sebelumnya, navigasi lintas halaman memfokuskan target hash tanpa membawa penanda tersebut, sehingga aturan global focus-visible menggambar kotak biru sekeliling section.

Validasi Chrome headless pada lebar 1440 dan 390 px: klik Contoh Aplikasi dari Beranda, Tim, dan Berita menuju Inference tidak menggambar outline section; target berada pada offset navbar yang benar dan menu Contoh Aplikasi aktif. Klik pada halaman Inference yang sama juga lulus. Tab dari section memfokuskan kartu pertama dengan indikator focus-visible solid yang tetap terlihat. Tidak terdeteksi exception JavaScript. Screenshot desktop diperiksa; bukti berada di `prototype/tests/.browser-check/portofolio-no-outline-1440.png` dan `portofolio-no-outline-390.png`. Tidak menjalankan Git/GitHub atau mengubah kode production/backend.

## Gradient dan kontak seluruh CTA - pilihan CTA-02 B

Disetujui pemilik melalui pilihan B: dua tautan layanan pada CTA Tentang Kami diganti WhatsApp/email. Seluruh delapan CTA kini memakai helper `aksiKontak()` dengan label “Hubungi lewat WhatsApp” dan “Kirim email”, URL dari konfigurasi existing, ikon/varian existing, serta fokus native. Email tidak memaksa tab baru. CTA Pelatihan dan Tim memperoleh tombol email; CTA lain diseragamkan labelnya. Tombol kontak pada navbar, hero, dan konten layanan bukan bagian dari perubahan ini.

CSS bersama menggantikan gradient linear lama dan pengecualian Beranda: navy radial ellipse 35% × 65% di kanan bawah, di atas dasar linear biru terang/medium. Token tetap `#2547F4` dan `#0A2240`. Footer tidak direvisi. Pada mobile, dua tombol mengisi lebar yang tersedia dan membungkus vertikal.

Validasi: kedua tes prototype lulus, termasuk regresi delapan CTA dengan nomor `62895634222240`, email `aicenter@unit.ubaya.ac.id`, parameter pesan/subjek, ikon, serta fokus kedua tautan. Chrome headless memeriksa delapan CTA pada lebar 1440, 820, dan 390 px (24 kombinasi): background terhitung identik, dua tombol terlihat dengan tinggi minimal 44 px, tanpa overflow horizontal atau exception JavaScript. Tab/Enter dan klik mouse masing-masing mengaktifkan kedua tautan dengan tujuan yang benar; event dicegat sebelum membuka WhatsApp atau handler email. Tidak mengirim pesan/email dan tidak mengklaim status akun atau pengiriman terverifikasi.

Screenshot Beranda desktop, Tim mobile, dan Tentang Kami desktop diperiksa. Sampling background di dekat empat sudut Beranda pada ketiga viewport mengonfirmasi kanan bawah paling gelap. Bukti berada di `prototype/tests/.browser-check/cta-shared-index-1440.png`, `cta-shared-index-820.png`, `cta-shared-index-390.png`, `cta-shared-tim-390.png`, serta `cta-shared-tentang-kami-1440.png`. Hasil terbatas pada browser headless yang diuji. Tidak menjalankan Git/GitHub, menambah dependency, atau mengubah kode production/backend.

## CTA-03 A - Titik tergelap pada kanan bawah Beranda

Disetujui pemilik: khusus CTA Beranda memakai arah `to bottom right` dengan titik biru 0%, campuran biru/navy 80/20 pada 40%, campuran 45/55 pada 72%, serta navy 100%. Token tetap `#2547F4` dan `#0A2240`. Gradient CTA lain dan footer tidak direvisi.

Chrome headless memeriksa lebar 1440, 820, dan 390 px. Perbandingan runtime sebelum/sesudah background mengonfirmasi geometri, padding, radius, font, warna teks, posisi tombol, dan background footer identik. Tidak terdeteksi overflow horizontal atau exception JavaScript. Screenshot desktop dan mobile diperiksa: transisi halus tanpa batas warna tajam yang terlihat; teks putih serta tombol tetap terbaca. Sampling background di dekat empat sudut mengonfirmasi kanan bawah paling gelap. Contoh mobile: kanan atas RGB 29/60/190, kanan bawah 12/39/83; kiri atas 36/70/238 dan kiri bawah 28/59/186. Screenshot mobile menggunakan viewport utuh agar koordinat sampel tidak terpengaruh pemotongan otomatis.

Bukti: `prototype/tests/.browser-check/cta-corner-1440.png`, `cta-corner-820.png`, dan `cta-corner-390.png`. Hasil ini terbatas pada browser headless yang diuji, bukan penilaian banding pada seluruh panel layar fisik. Tidak menjalankan Git/GitHub atau menambah dependency.

## UI-03 - Sudut membulat hanya di atas

Disetujui pemilik: permukaan section utama publik, termasuk hero, CTA, dan section demo aplikasi, memakai radius 24 px pada kiri atas dan kanan atas; kedua sudut bawah menjadi siku. Aturan bersama section/artikel dan aturan hero diselaraskan pada `public.css`; spesifikasi aktif dicatat di `docs/design.md`.

Validasi Chrome headless memeriksa 13 halaman pada lebar 1440 dan 390 px (26 kombinasi, 78 section/artikel). Seluruh elemen yang diperiksa memiliki radius terhitung 24/24/0/0 px, tanpa overflow horizontal atau exception JavaScript terdeteksi. Screenshot hero mobile dan CTA desktop diperiksa; bukti tersimpan di `prototype/tests/.browser-check/top-corners-hero-390.png` dan `top-corners-cta-1440.png`. Tidak menjalankan Git/GitHub atau menambah dependency.

## CTA-02 A - Gradient linear sesuai referensi

Disetujui pemilik: seluruh bidang CTA `.kontak` memakai `linear-gradient(135deg, var(--warna-biru), var(--warna-navy))`, dari biru `#2547F4` di kiri atas menuju navy `#0A2240` di kanan bawah. Ini menggantikan glow radial CTA sebelumnya. Spesifikasi aktif diselaraskan di `docs/design.md`.

Validasi Chrome headless pada Beranda, Tim, dan Tentang Kami di lebar 1440 serta 390 px (enam kombinasi) mengonfirmasi gradient terhitung sesuai token, tanpa overflow horizontal atau exception JavaScript terdeteksi. Screenshot desktop dan mobile diperiksa di `prototype/tests/.browser-check/cta-linear-1440.png` dan `cta-linear-390.png`. Tidak menjalankan Git/GitHub atau menambah dependency.

## Revisi UI/UX prototype - 10 November 2026

Disetujui pemilik: HERO-01 A, HERO-02, NAV-01/03, NAV-02 B, UI-01 A, UI-02, TEAM-01/02/03, CTA-01 A, FOOTER-01/02, NEWS-01, dan DOC-01 A. Implementasi dan spesifikasi terbaru ada di `docs/design.md` bagian 1 serta catatan khusus prototype pada `docs/requirement.md`. HERO-03 dikonfirmasi pemilik sebagai salah ketik; tidak ada revisi tambahan.

Root cause NAV-03 direproduksi pada kode lama dengan JSDOM: scroll 200, fokus menu, lalu scroll 0 tetap ringkas; blur memulihkan keadaan lebar. Penyebabnya guard fokus/dropdown keluar sebelum state scroll diperbarui. Perbaikan mengizinkan pemulihan di bawah 48 px tanpa memindahkan fokus. NAV-02 menghitung section dengan pasangan menu langsung memakai offset sticky header; section tanpa pasangan mempertahankan menu terakhir.

Verifikasi implementasi: `node prototype/tests/navbar.test.cjs` lulus untuk 13 halaman/fallback, enam menu, Kabar, Tim/Tentang Kami langsung, scrollspy dua arah dengan geometry tiruan, fokus/dropdown saat kembali ke atas, keyboard/pointer anchor, mobile, batas ruang, FLIP dan reduced motion. `node prototype/tests/audit-refinements.test.cjs` lulus untuk pagination/search/URL existing, enam profil tanpa pencarian, bidang riset, simulasi Tim kosong/error/pulih, alamat footer, penulis berita dan jumlah foto album. Uji canvas terkontrol 1600 frame menghasilkan kelompok 3/4/5 dot, maksimal delapan garis per frame dan 300 partikel. Uji ini bukan pengukuran performa perangkat nyata.

QA-01 disetujui pemilik. Chrome headless memeriksa 13 halaman pada lebar 390, 820, dan 1440 px (39 kombinasi) melalui file lokal prototype, tanpa overflow horizontal, gambar rusak, atau exception JavaScript yang terdeteksi. Screenshot pilihan diperiksa untuk hero, Tim, CTA/footer, galeri, menu mobile dan fokus Portofolio. Profil dan screenshot berada di `prototype/tests/.browser-check/`. Browser berhasil dijalankan di luar sandbox setelah izin eksekusi diberikan; inisialisasi renderer melalui CDP tidak merespons di dalam sandbox.

Pengujian interaksi browser menemukan bug ikon hamburger: penggantian SVG saat klik membuat target event lama terlepas dari DOM, lalu listener dokumen salah menganggap klik berasal dari luar navbar. Pemeriksaan menggunakan composedPath memperbaikinya. Regresi JSDOM kini mengklik path SVG; klik nyata pada lebar 390/820 px mengonfirmasi menu tetap terbuka, dropdown Kabar/Escape bekerja, dan Tim dapat dibuka dengan status halaman yang benar.

Lulus pada browser: scrollspy Beranda dua arah; navbar kembali lebar pada scroll 0 meskipun fokus/dropdown masih di navbar; Portofolio aktif setelah anchor mencapai offset sekitar 104 px; klik pointer tanpa outline section, Enter/Tab dengan fokus terlihat, serta Enter pada kartu membuka JejakAlumni. Galeri diuji dengan 1-6 clone gambar dalam DOM untuk memeriksa slot setara 4:3 dan contain pada tiga viewport; clone tidak ditulis ke sumber/data. Screenshot memperlihatkan foto utuh, kartu Tim ringkas, logo/alamat footer, serta gradient CTA/footer.

Instrumentasi canvas sementara di browser mencatat tarikan magnet sekitar 6,60 px dan p90 kerja 0,30 ms pada sampel desktop headless. Canvas berhenti menggambar saat keluar viewport dan menonaktifkan gerakan/interaksi pada reduced motion. Emulasi touch mengonfirmasi media tanpa hover/pointer halus. Ini bukan benchmark perangkat nyata atau bukti semua browser/perangkat. Konten iframe Google Maps, screen reader, dan touch fisik tidak dinilai.

Tidak menjalankan Git/GitHub, tidak menambah dependency, dan tidak mengubah kode aplikasi utama/backend. Dua tes prototype dijalankan ulang setelah perbaikan hamburger dan lulus.


## Implementasi keputusan audit — 10 November 2026

Selesai: NAV-01/02 (Aktivitas, konsistensi Galeri), UX-01/02/03 (fokus/status pagination, URL q/page dan Back/Forward, mailto), HOME-01 (scrim mobile), TRAIN-01 B (hanya paragraf layanan, semua judul tetap), INF-01 (tautan contoh penerapan), NEWS-01 (pengantar ringkas), TEAM-01 (portrait mobile 96×120 berdampingan dengan identitas), GAL-01 (gambar detail utuh), ALUMNI-01 (label lokal ringkas/tab dilipat), DOC-01 (tagline dihapus/jump link), UI-01 (h3 berita Beranda), MOTION-02 (perubahan reduced motion/hover), A11Y-01 (status skenario DocNext), ADMIN-01/02 (konfirmasi Retract, judul tabel mobile tidak tertutup aksi).

Dipertahankan: MOTION-01 dan CONTENT-01/02/03. Ditunda: A11Y-02. Simulasi tetap utuh. Tidak ada dependency baru, commit, atau push.

Verifikasi: `node prototype/tests/navbar.test.cjs` dan `node prototype/tests/audit-refinements.test.cjs` lulus. Edge headless memeriksa 26 halaman/state pada 1440/820/390 px (78 kombinasi), tanpa overflow halaman, gambar rusak, input tanpa label, atau exception yang terdeteksi. Fokus hasil, URL pagination, pencarian, selector Pelatihan, kontrol demo, status DocNext, validasi admin, dialog perubahan belum tersimpan, serta 429 diperiksa. Konfirmasi Retract diuji: Batal mempertahankan 15 item posted; konfirmasi mengubah menjadi 14 dalam memori sesi uji saja. Lock oleh pengguna lain menjadi readonly dan sesi 401 menampilkan pemulihan. Pemeriksaan visual kedua memperbaiki tinggi portrait mobile agar tidak mengikuti atribut tinggi lama.

QA-01 sebagian: browser/keyboard/viewport sudah diperiksa, tetapi screen reader, touch fisik dan zoom native 200–400% belum diverifikasi. QA-02 lulus untuk skenario prototype yang diperiksa: lock, validasi, unsaved, 401, 429, konfirmasi Retract, konflik 409 dan kehilangan heartbeat 410. Pada 409, dialog muncul dan isian tetap tersimpan saat ditutup; pada 410, form memberi instruksi pemulihan tanpa membuang isian. Integrasi backend nyata tetap memerlukan pengujian tersendiri. QA-03 tetap memerlukan perangkat nyata; hasil headless bukan klaim performa perangkat pengguna. QA-01/03 belum selesai menyeluruh. Kepatuhan WCAG menyeluruh belum dapat dinyatakan. Pengujian tidak mengubah data proyek atau backend.

Refinement Profil dan bola Beranda (9 November 2026): Tim/Tentang Kami kini berada di dropdown Profil, dengan status induk/anak dan ID submenu terpisah dari Layanan; 13 fallback navigasi tanpa JavaScript ikut diselaraskan. Scroll navbar tidak diubah. Skala bola turun 4% pada ketiga breakpoint; jeda aktivasi mint menjadi 1607–3750 ms (frekuensi sekitar 12% lebih tinggi), dengan envelope dan mode diam tetap. Rincian keputusan ada di `design.md` §1.

Verifikasi refinement: tes JSDOM lulus untuk 13 halaman, kedua submenu dan ID/ARIA, status Profil pada Tim/Tentang Kami, Escape/fokus, pergantian dropdown, navbar ringkas/mobile, serta fallback tanpa JavaScript. Edge headless mengonfirmasi pointer desktop pada keadaan lebar/ringkas, aktivasi kontrol mobile, navigasi ke Tim dan status aktifnya, skala bola pada 390/820/1440 px, serta konfigurasi jeda/envelope mint. Input tap sintetis tidak menghasilkan event klik pada sesi uji; touch pada perangkat fisik belum terverifikasi. Tidak ada galat runtime pada pemeriksaan browser yang selesai.

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
| §4.2 Pelatihan: hero konsultasi, pemilih jalur, Khusus, Workshop Bisnis, Reguler Coming Soon, penutup | `public/pelatihan.html` | Selesai; empat skenario radio berdasarkan hasil yang diinginkan, empat langkah konsultasi dikonfirmasi pemilik |
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
| T-16 | Konten | Enam bidang riset Tim di `assets/mock-data.js` ditetapkan pemilik untuk prototype, belum diverifikasi sebagai keahlian pribadi untuk produksi. Empat langkah konsultasi Pelatihan telah dikonfirmasi pemilik. Deskripsi IT Service Management serta Data Science masih data contoh. Handle Instagram masih placeholder di `src/data/kontak.ts`. |
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

Navbar (9 November 2026, dikoreksi): komponen global memiliki dua keadaan stabil. Keadaan lebar transparan memakai logo penuh di kiri, menu terpusat dengan kolom samping seimbang, dan CTA WhatsApp resmi di kanan; tidak ada kapsul pada keadaan ini. Scroll >96 px menyatukan ketiganya dalam pill terpusat (offset atas 12 px), dengan simbol dari aset logo yang sama, jarak menu merapat, permukaan putih, border, dan bayangan lembut; scroll <48 px mengembalikan keadaan lebar. FLIP native 360 ms menggerakkan kelompok tanpa meregangkan teks, dengan pembalikan dari posisi yang sedang terlihat. Reduced motion dan browser tanpa WAAPI mendapat keadaan statis yang berfungsi. Desktop mulai 1280 px dengan pemeriksaan ruang untuk menu dan kedua sisi; tablet/HP dan keadaan tidak muat memakai hamburger. Tiga belas HTML publik memiliki fallback `noscript`; Inference memiliki target Portofolio dan tiga tautan demo statis. `node prototype/tests/navbar.test.cjs` (jsdom) mencakup 13 halaman, CTA, hash/kunjungan langsung, hysteresis, freeze saat dropdown/fokus, Escape, menu mobile, fokus target, fallback tanpa JS, fallback bila isi tidak muat, FLIP dua arah/pembatalan, serta reduced motion. Canvas Beranda ditiru karena di luar lingkup tes navbar. Pemeriksaan terpisah di Edge headless mengukur posisi menu/kapsul, logo dan CTA di dalam kapsul, scroll dua arah dan pembalikan di tengah animasi, dropdown kedua keadaan, Escape, resize 390–1920 px tanpa overflow horizontal, anchor mobile, serta reduced motion. Belum menguji pembaca layar, perangkat fisik, atau seluruh browser. Splash serta sistem “Simulasi” tidak diubah.

Sapuan headline Inference (9 November 2026): satu span inline memakai gradient dari `#AAB7C7` ke putih selama 900 ms dan `box-decoration-break: slice`, tanpa gerakan/perubahan teks. Kontras warna awal terhadap navy dihitung 7,83:1. Uji jsdom lolos untuk teks aksesibel tunggal, pembersihan efek, tidak mengulang pada navigasi carousel, fallback reduced-motion/forced-colors/fitur tak didukung, penghentian saat preferensi berubah, serta struktur flow dan portofolio. Sintaks dan diff lolos. Jsdom tidak memverifikasi rendering gradient: urutan sapuan lintas baris, font wrapping, dan kelancaran desktop/HP masih perlu uji visual browser; host browser tidak tersedia. Deskripsi pemilik menjadi acuan motion, bukan pengamatan langsung situs Microdose. Splash dan sistem “Simulasi” tidak diubah.

Line-art Inference (9 November 2026): aset lokal `it-service-management-line.svg` dan `data-science-line.svg` dibuat sebagai ilustrasi konseptual 180 × 108 px, bukan dokumentasi kegiatan. Kartu memakai border tipis tanpa elevasi hover; seluruh flow berada dalam satu panel putih berbayangan ringan. Verifikasi jsdom lolos untuk validitas XML dan keberadaan dua SVG, dimensi/alt dekoratif, copy kedua area, satu panel berisi lima tahap yang tetap, pemicu visibilitas penuh, serta keberadaan portofolio/carousel/CTA. Sintaks dan diff lolos. Keseimbangan visual pada desktop/HP dan motion di browser belum diverifikasi karena host browser tidak tersedia. Label serta file sistem “Simulasi” tidak diubah.

Refinement Inference (9 November 2026): empat pertanyaan carousel diberi kutip editorial; heading menjadi “Jelajahi Peluang Penerapan AI”. Dua kartu area layanan sejajar pada ≥768 px dan bertumpuk pada HP. Lima tahap horizontal mulai 1280 px dengan penekanan sekali hanya ketika seluruh daftar terlihat; selain itu tetap statis vertikal. Uji jsdom dengan media/observer tiruan lolos untuk kutip, navigasi carousel, lima tahap tanpa perubahan isi, pemicu penuh/tidak berulang, fallback layar kecil/reduced-motion/tanpa observer, penghentian saat layar mengecil, tiga tautan portofolio, dan CTA. Sintaks serta diff lolos. Layout dan animasi visual desktop/HP belum diverifikasi di browser karena host tidak tersedia. Sistem “Simulasi” tidak disentuh.

Revisi flow Pelatihan (9 November 2026, Q1 A): eyebrow hero dihapus, status Reguler menjadi “Coming Soon”, dan eyebrow Workshop menjadi “Eksplorasi AI untuk Organisasi”. Workflow kini 4 kolom desktop dan vertikal pada tablet/HP, menggantikan susunan tablet 2 kolom. Nomor langkah terbaca pada daftar semantik. Observer dengan ambang seluruh daftar terlihat memicu penekanan sekali pada desktop; fallback selalu statis. Uji jsdom dengan observer/media tiruan lolos untuk pemicu penuh, tidak mengulang setelah selesai, fallback mobile/reduced-motion/tanpa observer, penghentian saat reduced motion berubah, empat pemetaan selector, dan CTA. Sintaks serta diff lolos. Geometry konektor, motion visual, keyboard native, dan pembaca layar belum diuji di browser; host browser tidak tersedia pada sesi ini. Sistem “Simulasi” tidak disentuh.

Soft Editorial Depth Pelatihan (9 November 2026): hero tetap tanpa media atau placeholder; pilihan mendapat state yang lebih jelas, panel hasil memakai fade opacity 200 ms saat motion diizinkan, dan empat langkah berada dalam satu panel workflow putih (4/2/1 kolom). Label “Simulasi” serta file sistem simulasi tidak diubah. Pemeriksaan jsdom lolos untuk empat pemetaan dan pergantian ulang, radio tunggal, fokus DOM, live region, tiga CTA resmi, serta struktur hero dan workflow. Sintaks dan diff diperiksa; tampilan browser, hover, keyboard native, pembaca layar, dan reduced motion secara visual belum diuji karena host browser tidak tersedia.

Refinement Pelatihan (9 November 2026): verifikasi jsdom mencakup empat pemetaan jalur dan pergantian ulang, satu radio aktif, alasan dan anchor hasil, fokus DOM tetap pada pilihan, urutan section, empat langkah, tiga CTA konsultasi resmi, status Reguler, serta penamaan institusi. Sintaks JavaScript dan `git diff --check` lolos. Uji browser desktop/HP, keyboard native, dan pembaca layar belum dilakukan pada revisi ini karena host browser/native tidak tersedia; hasil pemeriksaan lama di bawah tidak menggantikan uji visual revisi ini.

\[Ditandai\] Status bagian ini belum jelas dan tidak ikut dilebur ke `docs/` sampai diputuskan.

Diuji otomatis di Chrome headless lewat `file://`, semua lolos:

- **Pengujian umum (104 cek):** 10 halaman publik dan 10 rute admin pada 1440 px dan 390 px (tanpa galat konsol, tanpa scroll horizontal, gambar termuat), tautan lokal, pencarian, pagination, urutan tanggal, semua keadaan panel, validasi form dan tanggal, validasi file, Simpan nonaktif, teks berbahaya tampil sebagai teks, soft lock, 409, heartbeat, 401, 403, Post, Retract, trash, keluar.
- **Revisi admin (60 cek):** 7 tabel pada 1440 px dan 390 px (halaman tidak bergulir ke samping, kolom Aksi terjangkau, `caption`, `th scope`), urutan, aksi per status, foto dan placeholder termasuk gambar gagal dimuat, kalender dengan jam palsu 01.30 WIB (hari ini tetap 06-11-2026 walaupun UTC masih tanggal 5), tanggal masa depan ditolak lewat klik dan ketikan, keyboard, nilai default dan nilai tersimpan.
- **JejakAlumni:** lima lebar layar (1440, 1024, 768, 390, 360 px) tanpa scroll horizontal; 14 kombinasi filter cocok dengan data, status berjumlah 100%, tanpa nilai kosong; klik kartu, refresh, dan tautan kembali ke `inference.html#portofolio`; 11 halaman publik tanpa galat konsol.
- **Perencanaan Implan Gigi:** enam lebar layar (1440, 1280, 1024, 768, 390, 360 px) tanpa scroll horizontal di keempat tahap; semua toggle dan pilihan mengubah ilustrasi serta `aria-label`, dan bisa dikembalikan; Enter pada Berikutnya memindahkan fokus ke judul tahap; panah keyboard pada radio; klik kartu, refresh, dan tautan kembali; disclaimer tampil sebelum demo; tanpa input data; satu-satunya angka mm adalah MAE yang dilaporkan.

Belum diperiksa: pembaca layar, browser selain Chrome, perangkat sungguhan, dan isi peta Google (iframe ada, isinya tidak diverifikasi). Kontras dihitung manual untuk warna utama, bukan audit menyeluruh.
