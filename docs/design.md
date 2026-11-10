# Design — UBAYA AI Center CMS

Dokumen aktif. Riwayat perencanaan awal (`00-detail-rencana-web.md`, `01-kontrak-api.md`) tersedia di riwayat Git, commit `5928f05` (tag `arsip-docs-v1`). Alur konten halaman: \[Lihat requirement.md §4 dan §5\].

## 1. Navbar dan animasi global

### Hero viewport dan tombol kontak - HERO-03 A, SPHERE-03 A, CTA-03 A

Persetujuan terbaru: navbar, margin hero 12 px di setiap sisi vertikal, dan hero bersama-sama memenuhi satu layar pertama. `--hero-tinggi-navbar` lokal Beranda bernilai 88 px atau 76 px ketika navbar memakai keadaan mobile; `:has(> .nav--mobile)` mengikuti fallback menu berdasarkan ruang. Hero memakai minimum `100svh - tinggi navbar - 24px`, dengan fallback `100vh`. Isi tetap dapat menambah tinggi jika diperlukan. Padding vertikal 48 px mobile/72 px mulai tablet, gutter responsif, gap 24 px, serta H1 36–64 px dipertahankan. Komposisi teks berada di tengah vertikal ruang hero.

SPHERE-03-FIX A disetujui dan menggantikan pembatasan tinggi kanvas sebelumnya. Kanvas memenuhi tinggi hero; CSS desktop menetapkan `--bola-tinggi-acuan: 642` sebagai angka dalam CSS px. `bola-neural.js` membaca acuan opsional hanya ketika ukuran berubah, lalu menghitung radius dari skala dikali sisi terkecil antara lebar, tinggi kanvas, dan tinggi acuan. Tanpa acuan, perhitungan mengikuti ukuran kanvas seperti sebelumnya. Radius dasar desktop tetap 385,2 px pada hero yang lebih tinggi, dengan pusat/skala/regangan terdokumentasi. Tidak ada batas scene internal yang memotong titik; clipping dekoratif di batas luar hero sesuai skala yang ada tetap berlaku. Parameter magnet, aktivasi mint dan konektivitas tidak direvisi.

Delapan kelompok `.kontak__aksi` memakai tombol dengan minimum tinggi 56 px pada layar di atas 48rem, serta 52 px pada lebar ≤48rem. Padding .75rem/1.5rem (12/24 px), font 16 px, ikon CSS 20 px, gap ikon 10 px, gap tombol 16 px. Pada mobile, grid satu kolom dengan baris setinggi sama menjaga konsistensi bila label membungkus. Seluruh permukaan tautan tetap menjadi target klik; varian primary putih dan secondary outline mengikuti style existing. Dimensi tombol di navbar/hero serta gradient CTA/footer tidak direvisi.

### Revisi UI prototype yang disetujui - 10 November 2026

Keputusan berikut berlaku hanya pada direktori `prototype/` dan menggantikan rincian prototype terdahulu yang bertentangan di bawahnya. Implementasi produksi tidak diubah.

- Navbar berisi Beranda, Layanan, Contoh Aplikasi, Kabar, Tim, dan Tentang Kami. Kabar mengelompokkan Berita/Galeri Dokumentasi; Profil dihapus. Semua 13 fallback tanpa JavaScript mengikuti susunan ini. Listener klik luar memakai jalur event (composedPath) agar penggantian ikon hamburger tidak langsung menutup menu yang baru dibuka.
- NAV-02 pilihan B: hanya section dengan pasangan menu langsung yang mengubah menu aktif. Di Beranda: hero menjadi Beranda, Layanan menjadi Layanan, Galeri Dokumentasi/Berita menjadi Kabar. Testimoni, kontak, dan footer mempertahankan pasangan terakhir. Inference memakai Layanan sampai Portofolio menjadi aktif, lalu Contoh Aplikasi. Scroll kembali ke atas memulihkan menu halaman; halaman terpisah mengikuti konteksnya. Pembacaan posisi dibatasi requestAnimationFrame dan memakai offset navbar 92/104 px; resize, font siap, dan perubahan tinggi konten mengevaluasi ulang posisi.
- Navbar mempertahankan threshold 96/48 px dan FLIP 360 ms. Fokus/dropdown masih menahan transformasi di tengah halaman, tetapi tidak menghalangi pemulihan keadaan lebar pada scroll <48 px. Fokus pengguna tidak dipindahkan.
- HERO-01 A: radius magnet 150 px, kekuatan 0,32, tau 180 ms; ukuran/posisi bola tidak diubah. HERO-02: jeda aktivasi mint 1200-2800 ms, kelompok 3-5 dot total dari kandidat tetangga terdekat yang diacak; maksimal dua kelompok dan delapan garis per frame. Batas DPR, partikel, reduced motion, serta penghentian di luar viewport/tab tersembunyi dipertahankan.
- UI-03: permukaan section utama publik, termasuk hero, CTA, dan demo aplikasi, memakai token radius besar 24 px hanya pada kiri atas dan kanan atas. Kedua sudut bawah siku (`border-radius: var(--radius-besar) var(--radius-besar) 0 0`). Tidak memakai clipping global yang memotong fokus/dropdown; pola lembar bertumpuk tetap.
- UI-04: container `#portofolio` tidak menggambar outline saat menerima fokus anchor, termasuk navigasi lintas halaman. Fokus tetap dapat dipindahkan ke section sebagai titik awal navigasi. Border dan outline keyboard pada tautan kartu, tombol, serta navbar tetap terlihat melalui aturan focus-visible existing; Tab dari section menuju kartu pertama. Aturan fokus global dan skip link tetap.
- Tim tanpa pencarian; foto 96x120 px berdampingan dengan identitas. Bidang penelitian ditampilkan pada panel berpenekanan visual, tanpa mengubah isi keahlian. Grid 3/2/1 kolom; padding atas hero 32 px.
- Seluruh delapan CTA mengikuti revisi gradient/kontak bersama di bawah. Footer memakai glow radial biru terang di kanan atas (28%). Logo footer dibatasi lebar 198 px dengan rasio intrinsik; tinggi sekitar 64 px termasuk padding. Alamat: Gedung Fakultas Teknik TA.
- Metadata berita: [tanggal existing] oleh Tim Media Ubaya AI Center. Foto detail album setara dalam bingkai 4:3 dengan object-fit: contain; dua kolom desktop/tablet, satu kolom mobile. Tidak ada perlakuan headline untuk foto pertama. Foto berbeda rasio menyisakan ruang latar.

Revisi gradient/kontak seluruh CTA, dengan pilihan CTA-02 B, menggantikan aturan linear CTA-02 A dan aturan khusus Beranda CTA-03 A. Delapan section `.kontak` (Beranda, Pelatihan, Inference, JejakAlumni, Perencanaan Implan Gigi, DocNext, Tim, Tentang Kami) memakai CSS bersama: radial ellipse 35% × 65% berpusat di kanan bawah, navy penuh di pusat, navy beralfa 65% pada 30%, transparan pada 100%. Dasarnya linear 110° dari biru ke campuran biru/navy 80/20. Biru terang/medium mendominasi; navy pekat terbatas di sudut. Token tetap `#2547F4` dan `#0A2240`; gradient footer tidak direvisi.

Helper `aksiKontak()` menghasilkan dua tautan native dengan ikon existing: “Hubungi lewat WhatsApp” (`https://wa.me/62895634222240` beserta pesan existing, tab baru dengan noopener) serta “Kirim email” (`mailto:aicenter@unit.ubaya.ac.id` beserta subjek existing, tanpa memaksa tab baru). Keduanya memakai sumber `MOCK.konten.kontak`, yang sesuai konfigurasi `src/data/kontak.ts`. Seluruh CTA memakai label, varian, gap, hover dan fokus yang sama; pada lebar ≤48rem kedua tombol memakai lebar penuh dan membungkus vertikal. Pelatihan/Tim memperoleh email. Pilihan B mengganti dua tautan layanan pada CTA Tentang Kami dengan pasangan kontak. Tidak menambah section CTA pada halaman yang sebelumnya tidak memilikinya. Pengujian tidak mengirim pesan atau email sungguhan.

HERO-03 dikonfirmasi sebagai salah ketik. QA-01 disetujui dan validasi Chrome headless selesai; rincian bukti dan batas pengujian ada di `prototype/CATATAN.md`.

### CTA Tentang Kami dan penyesuaian tipografi terarah

Disetujui pemilik: CTA-ABOUT-01 A dan TYPO-01 B (TYPO-01A–F). Ketentuan ini menggantikan rincian prototype terdahulu yang bertentangan; token global, font, body, nama/jabatan/bidang riset Tim, metadata, footer dan tipografi admin tetap. TYPO-01G–I dipertahankan; TYPO-01J ditunda.

CTA Tentang Kami memakai dua kolom `minmax(0, 1fr) auto` mulai 64rem (1024 px), dengan `align-items: center` dan gap existing 3rem (48 px). Judul/deskripsi berada di kiri, pasangan tombol horizontal di kanan. Di bawah breakpoint tersebut, grid satu kolom tetap. Aturan dibatasi pada `.kontak--tentang .kontak__dalam`; gradient, radius, ikon, tautan kontak dan komponen tombol mengikuti spesifikasi bersama.

| Kode | Komponen | Ukuran dan perilaku aktif |
|---|---|---|
| TYPO-01A | Hero Beranda, Pelatihan, Inference | `clamp(2.25rem, 1.5rem + 3vw, 4rem)`: 36–64 px pada root 16 px |
| TYPO-01B | H1 Tim, Galeri Dokumentasi, Tentang Kami | `clamp(2rem, 1.35rem + 2.7vw, 3.5rem)`: 32–56 px |
| TYPO-01C | Tautan menu navbar | Desktop .9375rem (15 px), mobile tetap 1rem (16 px); pengukuran ruang/hamburger existing tetap |
| TYPO-01D | Kartu portofolio tablet | Pada 48–63.99rem: dua kolom, judul 1.25rem (20 px), deskripsi 1rem/1.6 (16/25.6 px), badge .8125rem (13 px). Desktop tiga kolom dan mobile satu kolom tetap; subgrid menyelaraskan isi kartu |
| TYPO-01E | Dokumen sumber DocNext | .9375rem (15 px) pada tablet/desktop; 1rem (16 px) pada mobile. Tag kecil tetap .75rem (12 px) sesuai penundaan TYPO-01J |
| TYPO-01F | H1 detail berita/dokumentasi | `clamp(2rem, 1.4rem + 2.3vw, 3rem)`: 32–48 px; lebar baca dan isi artikel tetap |

Implementasi memakai selector khusus halaman/komponen dalam `public.css` serta `document-intelligence.css`, sehingga `--judul-1`, `--judul-2`, dan `--teks-kecil` tidak diubah secara global. Font Cabinet Grotesk, General Sans, serta Geist Mono dipertahankan. Hasil computed style, pemeriksaan layout, dan batas validasi aksesibilitas dicatat di `prototype/CATATAN.md`.


### Keputusan audit yang diterapkan — 10 November 2026

Catatan ini memperbarui rincian implementasi sebelumnya tanpa mengubah identitas visual atau motion scroll navbar. Menu utama kini lima: Beranda, Layanan, Contoh Aplikasi, Aktivitas, dan Profil. Aktivitas mengelompokkan Berita serta Galeri Dokumentasi; label galeri juga diselaraskan pada footer, breadcrumb, tautan balik, dan fallback tanpa JavaScript. Kontak email memakai `mailto:` ke alamat resmi.

Pencarian dan pagination Berita/Galeri menyimpan `q` dan `page` di URL serta memulihkannya pada refresh/Back/Forward. Status hasil memakai live region ringkas; aktivasi pagination memulihkan fokus ke hasil bila fokus sebelumnya terlepas saat render, dengan offset header. Nomor halaman di luar jangkauan dikoreksi ke halaman terakhir.

Refinement visual: scrim hero Beranda mobile diperkuat tanpa mengubah ukuran titik; pengantar Berita memakai skala judul lebih tenang; kartu Tim mobile menyandingkan portrait ringkas dengan nama/peran; gambar detail album tampil utuh pada proporsi aslinya. Hero Inference menambah “Lihat Contoh Penerapan” ke Portofolio dan ruang pertanyaan mobile diringkas. JejakAlumni mengurangi label demo lokal berulang dan melipat tab belum tersedia. Pembuka DocNext menghapus tagline berulang, menyediakan tautan langsung ke demo, dan mengumumkan ringkasan hasil skenario. Judul kartu Berita pada Beranda memakai h3.

TRAIN-01 disetujui dengan revisi: hanya paragraf penjelas dua layanan yang dipadatkan; semua judul, eyebrow, skenario, fokus, workflow, dan CTA dipertahankan. Carousel mengikuti perubahan preferensi reduced motion dan kemampuan hover, tanpa menambah tombol jeda. Tabel admin mobile tidak menempelkan kolom aksi di atas judul; tersedia petunjuk scroll dan judul dapat membungkus. Retract memerlukan konfirmasi sebelum konten ditarik.

Keputusan dipertahankan: MOTION-01 serta CONTENT-01/02/03. A11Y-02 ditunda. Label sistem Simulasi tetap. QA-01/02/03 adalah cakupan verifikasi; hasil dan keterbatasan pengujian dicatat di `prototype/CATATAN.md`, bukan dianggap seluruhnya lulus hanya karena implementasi selesai.

Refinement Profil dan hero Beranda (9 November 2026): Tim dan Tentang Kami dikelompokkan di bawah tombol “Profil”, setelah Dokumentasi, dengan tujuan `tim.html` dan `tentang-kami.html` tetap. Profil memakai pola dropdown Layanan dengan ID submenu unik (`menu-profil` dan `menu-layanan`), dukungan keyboard/touch, serta penekanan induk ketika halaman anak aktif; tautan anak tetap memakai `aria-current="page"`. Fallback tanpa JavaScript memakai `details` untuk kedua dropdown. Logika scroll dan timing navbar tidak berubah; lebar kapsul menyesuaikan jumlah menu yang kini enam. Pada hero Beranda, skala bola dikurangi 4% di setiap breakpoint: 1.056 mobile, 1.008 tablet, 1.20 desktop. Pusat dan regangan yang ada dipertahankan. Frekuensi aktivasi mint dinaikkan sekitar 12% melalui jeda 1607–3750 ms, sementara fase muncul/bertahan/meredup tetap 500/900/1100 ms. Jeda mode diam, reduced motion, rotasi, dan napas tidak diubah.

Navigation Bar awalnya menggunakan referensi dari tautan: [https\://symboolic.com/en](https://symboolic.com/en). Koreksi berikut mengikuti spesifikasi dua keadaan dari pemilik; tidak mengasumsikan perilaku tambahan dari referensi.

Navbar prototype memakai satu komponen dengan dua keadaan stabil. Di atas halaman, logo penuh berada di kiri, navigasi tepat di tengah area viewport yang tersedia, dan CTA navy “Diskusikan Kebutuhan AI” ke WhatsApp resmi berada di kanan. Dua kolom samping seimbang menjaga posisi tengah menu; gutter responsif mengikuti lebar layar, bukan batas wadah konten 1200 px. Keadaan lebar transparan tanpa kapsul, border, atau bayangan. Setelah scroll >96 px, ketiga kelompok berkumpul dalam satu pill terpusat dengan lebar mengikuti isi, offset atas 12 px, permukaan putih sedikit transparan, border halus, serta bayangan kecil. Jarak antarlink merapat. Keadaan lebar kembali saat scroll <48 px untuk menghindari jitter; navbar tidak terus menyusut sepanjang halaman. Logo simbol memakai bagian simbol utuh dari aset `AI-Center_Logo.png` melalui viewport CSS; rasio gambar tidak diubah dan tautan logo memiliki satu nama aksesibel.

Transisi posisi memakai FLIP melalui Web Animations API native selama 360 ms dengan `cubic-bezier(0.77, 0, 0.175, 1)`. Posisi logo, tiap item menu, dan CTA dibaca sebelum/sesudah perubahan keadaan, lalu dianimasikan dengan `transform` tanpa meregangkan teks; permukaan kapsul memakai transform dan opacity terpisah. Tidak ada animasi lebar, padding, atau tinggi setiap frame. Pembalikan scroll membaca posisi yang sedang terlihat sebelum membatalkan animasi sebelumnya. Pergantian logo memakai opacity 160 ms dengan `--gerak-keluar`; saat kembali melebar, pergantian ditunda 160 ms agar wordmark tidak bertabrakan dengan menu. Pengukuran ulang saat resize/font siap tidak memutar ulang efek ini. Ruang header tetap (88 px desktop, 76 px mobile) agar scroll tidak menggeser konten. Transformasi ditahan selama dropdown terbuka atau fokus berada di navbar, lalu dievaluasi kembali saat interaksi selesai; fokus yang masuk menghentikan gerakan ke keadaan tujuan. Reduced motion meniadakan perpindahan dan transisi logo. Browser tanpa Web Animations API tetap memperoleh kedua keadaan tanpa animasi FLIP.

Navigasi desktop digunakan mulai 1280 px hanya bila menu dan dua kolom samping seimbang muat; jika tidak, menu hamburger digunakan. Pada tablet/HP, logo penuh dan hamburger tetap, dengan panel menu yang dapat digulir dan CTA di dalamnya. Dropdown desktop berposisi absolut sehingga tidak mengubah ukuran kapsul. Dropdown memakai tombol native, `aria-expanded`/`aria-controls`, klik luar, Escape, dan penutupan saat fokus keluar. Escape dari menu mobile mengembalikan fokus ke hamburger. Tautan tetap native, termasuk dukungan klik dengan modifier.

“Contoh Aplikasi” berada setelah Layanan dan menaut ke `inference.html#portofolio`. Pada halaman Inference dengan hash tersebut, hanya Contoh Aplikasi yang mendapat penekanan aktif utama (`aria-current="location"`); pada Inference biasa, Layanan aktif. Hashchange dan Back/Forward memperbarui keadaan. Menu mobile menutup sebelum navigasi. Offset anchor mengikuti header (104 px desktop, 92 px mobile); kunjungan langsung/refresh menyelaraskan target setelah render dan font siap, kecuali pengunjung sudah berinteraksi. Struktur Portofolio, kartu, serta CTA demo tidak diubah.

Ketiga belas HTML publik memiliki fallback `noscript` dengan navigasi native dan dropdown Layanan melalui `details`. Halaman Inference menyediakan target Portofolio statis beserta tautan ke tiga demo; demo interaktif tetap memerlukan JavaScript. Fallback tidak tampil saat JavaScript aktif. Splash, footer, dan label sistem “Simulasi” tidak termasuk refinement navbar.

Catatan koreksi navbar (9 November 2026): komposisi lebar dan transisi diperbaiki agar logo, menu, dan CTA benar-benar menyatu dalam satu kapsul. Pemeriksaan Edge headless mencakup posisi kedua keadaan, scroll dua arah, pembalikan saat animasi berjalan, dropdown, Escape, resize 390–1920 px, anchor mobile, dan reduced motion. Tes JSDOM tetap mencakup navigasi 13 halaman dan fallback tanpa JavaScript. Pengukuran ini tidak membuktikan kinerja pada perangkat fisik atau kompatibilitas seluruh browser.

Animasi untuk interaksi button, card, pada page secara global: Fase prototype: CSS/JS murni. Fase stack: GSAP dan motion.dev \[Lihat stack.md §2\]. Pengecualian: animasi bola neural pada hero Beranda memakai canvas 2D dengan loop `requestAnimationFrame` sendiri (bukan GSAP/motion.dev); acuannya `prototype/assets/js/bola-neural.js`.

Hero Beranda berupa bidang navy dengan "bola neural" (canvas 2D): titik berbentuk bola yang berputar pelan dan "bernapas", sesekali satu node menyala mint beserta garis tipis ke tetangganya, dan titik di dekat kursor tertarik sedikit (tidak aktif di atas tautan atau CTA). Bola adalah lapisan latar yang tidak menerima klik, dengan scrim statis agar teks tetap terbaca. Perangkat tanpa hover atau pointer halus memakai titik lebih sedikit, frame rate lebih rendah, dan tanpa interaksi. `prefers-reduced-motion`: bola diam. Bila waktu kerja per frame terlalu tinggi, titik dikurangi, lalu bola diam. Animasi berhenti saat hero di luar layar atau tab tersembunyi. Angka rinci ada di `KONFIG` pada `bola-neural.js`; kinerja belum diukur di perangkat nyata.

Tidak ada tombol jeda untuk animasi hero, dan carousel (testimoni Beranda, pertanyaan Inference) juga tanpa tombol jeda. Ini penyimpangan disengaja dari WCAG 2.2.2 (Pause, Stop, Hide) atas keputusan pemilik. Untuk membatasinya, autoplay carousel hanya aktif di perangkat ber-hover dan berhenti saat disorot atau difokus; keringanan lain hanya `prefers-reduced-motion`.

## 2. Referensi visual per halaman

Tidak semua situs referensi dapat dibaca penuh: dari empat yang dicoba, satu hanya terbaca struktur menunya, satu hanya teks judul, satu hanya palet tanpa perilaku animasi, dan satu gagal diakses. Hero dan pergantian section Beranda karena itu merupakan tafsiran sendiri, bukan salinan referensi.

**Landing Page (Home)**

Copy Beranda memakai arah **Warm Editorial Storytelling**. Hero memperkenalkan manfaat pengembangan keterampilan dan penerapan AI bagi individu, institusi, serta bisnis. Kartu Pelatihan menekankan manfaat belajar bagi individu dan kelompok; kartu Inference and Solutions menekankan penerapan model dalam sistem atau proses kerja organisasi, termasuk integrasi dan pendampingan teknis. Kedua kartu tidak harus memberi penekanan yang sama pada setiap audiens.

CTA utama hero “Diskusikan Kebutuhan AI” membuka WhatsApp resmi, sedangkan “Lihat Layanan” menuju bagian layanan. Kartu layanan memakai “Lihat Pelatihan” dan “Lihat Solusi AI”; bagian Galeri Dokumentasi menaut ke daftar album. Judul testimoni “Cerita dari mereka yang belajar dan bekerja bersama kami” dipertahankan. Penutup menampung pertanyaan belajar maupun penerapan AI melalui kanal WhatsApp dan email yang sudah ada. Penyuntingan copy mempertahankan alur section, cakupan layanan, dan pola visual Beranda.

Pada page home, perpindahan dari satu section story telling ke section berikutnya menggunakan referensi animasi pada referensi [https\://brand.vettvangur.is/](https://brand.vettvangur.is/)

Section hero menggunakan referensi dari tautan: [https\://hub.orca.id/marketplace?\_gl=1\*63j1jh\*\_ga\*MTUyMjY2NzkxMy4xNzkzMzU4MTI2\*\_ga\_M905QM25JL\*czE3OTMzNTgxMjUkbzEkZzEkdDE3OTMzNTgxMzckajQ4JGwwJGgw](https://hub.orca.id/marketplace?_gl=1*63j1jh*_ga*MTUyMjY2NzkxMy4xNzkzMzU4MTI2*_ga_M905QM25JL*czE3OTMzNTgxMjUkbzEkZzEkdDE3OTMzNTgxMzckajQ4JGwwJGgw)

Referensi tambahan untuk section hero dapat diakses ke [https\://21st.dev/community/components/s/hero](https://21st.dev/community/components/s/hero),

Referensi utama untuk carousel testimoni ada di: [https\://www\.embla-carousel.com/docs/examples/predefined](https://www.embla-carousel.com/docs/examples/predefined) (note: bagian autoplay).

Tampilan preview berita di home page akan menggunakan referensi [https\://www\.altoira.com/](https://www.altoira.com/) pada bagian 'Articles'.

**Page Pelatihan:** referensi ada di dalam alur halaman \[Lihat requirement.md §4.2\].

Pada fase prototype, Pelatihan memakai **Warm Editorial Storytelling**: hero teks berlatar biru muda dimulai langsung dengan headline tanpa eyebrow pengganti, lalu pemilih jalur, Pelatihan Khusus, Workshop & Seminar untuk Bisnis, teaser Reguler, dan penutup. “Konsultasi Gratis” pada hero, Khusus, dan penutup langsung membuka WhatsApp resmi; tombol kedua hero “Lihat Pilihan Pelatihan” menuju pemilih. Nama institusi dalam copy halaman adalah UBAYA AI Center.

Pelatihan Khusus menekankan pengembangan kompetensi untuk individu maupun organisasi; Workshop & Seminar untuk Bisnis menekankan pemahaman bersama serta eksplorasi AI untuk perusahaan. Masing-masing memakai eyebrow, judul, subheading, dua paragraf, fokus singkat, dan CTA sesuai tujuan. Eyebrow Workshop adalah “Eksplorasi AI untuk Organisasi”. Empat langkah konsultasi tetap berada di bagian Khusus. Reguler tampil sebagai blok ringkas “Coming Soon”, tanpa CTA pendaftaran. Penutup menyediakan satu CTA konsultasi.

Pemilih jalur bersifat opsional: empat skenario berbentuk radio, dua kolom pada layar lebar dan satu pada HP, diikuti panel hasil berisi layanan, alasan, dan tautan kontekstual. Skenario karyawan belajar dan individu mengarah ke Khusus; manajemen serta tim operasional yang mengeksplorasi AI mengarah ke Workshop Bisnis. Tidak ada pilihan awal, animasi pemrosesan, atau keputusan berbasis AI. Fokus tetap pada radio saat hasil diperbarui; hasil diumumkan melalui live region. Token warna, radius, fokus, dan jarak yang ada digunakan kembali tanpa dependensi baru.

Kartu Pelatihan Reguler memakai outline dashed 2 px abu-abu (`#A1A8AF`) dan badge statis “Coming Soon” dengan latar `#F3F4F6`, teks `#4B5563`, serta border `#D1D5DB`. Badge, judul, dan deskripsi rata kiri dengan jarak 12 px; padding kartu menyesuaikan 24–32 px. Kartu tetap tanpa bayangan, animasi, atau CTA pendaftaran agar sekunder terhadap dua layanan aktif. Gaya ini khusus Reguler dan tidak mengubah eyebrow layanan lain.

Refinement visual memakai **Soft Editorial Depth** secara selektif. Hero tetap berbasis teks, dengan headline dibatasi lebarnya dan gradasi biru muda yang halus. Dokumentasi pelatihan autentik akan menjadi visual hero setelah aset tersedia; untuk sementara tidak ada gambar, ilustrasi, placeholder, atau kolom visual kosong. Hierarki visual tetap Beranda paling ekspresif, halaman layanan lebih tenang, dan halaman pendukung/detail lebih sederhana.

Pilihan skenario memakai border yang lebih jelas, radio native sebagai tanda pilihan selain warna, serta ring inset pada keadaan terpilih. Hover pada pointer halus dapat naik 1 px selama 170 ms; transisi permukaan dan border 180 ms. Panel rekomendasi memakai latar muda, border, dan bayangan kecil; isi hasil muncul dengan fade 200 ms tanpa menunda pembaruan hasil atau memindahkan fokus. Judul dan empat langkah konsultasi berada dalam satu panel workflow putih, bukan kartu terpisah. Workshop mempertahankan transisi section putih; Reguler tetap sekunder tanpa elevasi. Motion hanya aktif saat `prefers-reduced-motion: no-preference`, tanpa animasi tinggi, loop, atau scroll reveal baru. Label “Simulasi” milik sistem prototype dikecualikan dari refinement dan tetap utuh.

Workflow konsultasi memakai nomor langkah yang terbaca dan konektor statis. Desktop (≥1024 px) menampilkan empat langkah horizontal; tablet dan HP memakai susunan vertikal. Bila seluruh daftar terlihat pada desktop dan motion diizinkan, penekanan halus berjalan sekali dari langkah 1–4: durasi 420 ms per indikator dengan jeda awal 160 ms (total 900 ms), menggunakan opacity dan transform pada lapisan dekoratif. Semua teks langsung terbaca; tampilan akhir kembali statis dan tidak menyatakan proses sedang berjalan atau selesai. Observer dilepas setelah pemicu. Reduced motion, browser tanpa observer, atau layar lebih kecil memakai konektor statis; perubahan ke layar kecil/reduced motion menghentikan animasi.

**Page Inference and Solutions**

Headline hero memakai sapuan warna progresif sekali selama 900 ms (linear), dari abu-abu terang `#AAB7C7` menuju putih pada latar navy. Efek memakai satu span inline, gradient yang di-clip pada teks, dan `box-decoration-break: slice` untuk mempertahankan urutan fragmen baris; teks tidak dipecah per huruf, diduplikasi, atau digerakkan. Setelah animasi selesai/dibatalkan, warna kembali ke pewarisan normal. Reduced motion, forced colors, atau browser tanpa fitur yang diperlukan memakai headline biasa. Pendukung dan kontrol langsung tersedia; carousel tetap memiliki transisi slide sendiri dan tidak mengulang sapuan headline. Splash tidak diubah atau dijadikan pemicu. Ini adaptasi berdasarkan deskripsi motion pemilik; motion Microdose belum diamati langsung. Perilaku lintas baris dan kelancaran masih memerlukan verifikasi visual browser.

Kartu IT Service Management dan Data Science memakai SVG line-art dekoratif sebelum judul, masing-masing 180 × 108 px, dengan garis navy, aksen biru, dan latar transparan. Visual menggambarkan alur layanan sistem serta tabel/grafik analisis; tidak mewakili hasil layanan atau data nyata. Kedua kartu memakai border tipis dengan bayangan hover yang muncul halus selama 180 ms pada perangkat ber-hover dan pointer halus. Efek hanya menganimasikan opacity lapisan bayangan, tanpa perpindahan kartu atau perubahan cursor; reduced motion meniadakan transisinya. Dimensi gambar ditetapkan dan `alt` kosong karena judul serta deskripsi sudah menyampaikan topiknya. Lima tahap berada dalam satu panel putih dengan border dan bayangan kecil di atas latar section muda; padding responsif menjaga lebar baca. Portofolio dan penutup mempertahankan perlakuan visual yang ada, tanpa tambahan motion atau bayangan.

Carousel hero memakai tanda kutip tipografis “…” pada empat pertanyaan hipotetis, tanpa atribusi testimoni. Perilaku carousel yang ada tetap. Heading dua area layanan adalah “Jelajahi Peluang Penerapan AI”; kartu IT Service Management dan Data Science sejajar dua kolom pada desktop/tablet (≥768 px), bertumpuk pada HP, dan mengikuti tinggi baris yang sama.

Lima tahap penerapan model AI tetap dalam urutan dan isi semula, dengan nomor yang terbaca serta konektor dekoratif. Pada ≥1280 px, flow horizontal mendapat penekanan sekali hanya bila seluruh daftar masuk viewport dan motion diizinkan: 420 ms per indikator dengan jeda awal 160 ms (total 1060 ms). Pengamat dilepas setelah pemicu dan animasi berhenti pada keadaan statis. Bila seluruh daftar tidak pernah terlihat bersamaan, flow tetap statis. Di bawah 1280 px, susunan vertikal memakai konektor statis; reduced motion dan browser tanpa observer juga selalu memiliki tampilan lengkap. Perubahan ukuran ke layar kecil atau preferensi reduced motion menghentikan penekanan. Motion tidak mengisyaratkan proses konsultasi sedang berlangsung. Portofolio, CTA, dan sistem label “Simulasi” tidak diubah.

Pada fase prototype, pengantar menjelaskan inference sebagai penggunaan model AI untuk menghasilkan prediksi pada data baru, lalu menghubungkannya dengan kebutuhan aplikasi atau sistem. IT Service Management dan Data Science tetap menjadi dua area awal diskusi; deskripsinya menggunakan bahasa yang tidak menjanjikan hasil tertentu. Lima tahap penerapan model AI telah dikonfirmasi pemilik sebagai proses layanan dan dipertahankan. Penutup mengajak pengunjung mendiskusikan kebutuhan AI melalui kanal kontak yang sudah ada. Nama resmi layanan dan tujuan CTA tidak berubah.

Bagian Portofolio memakai eyebrow “PORTOFOLIO”, judul “Jelajahi Portofolio Kami.”, dan grid kartu yang sudah ada: tiga kolom di desktop, mengikuti lebar di tablet, satu kolom di HP. Setiap kartu menaut penuh ke demonya dan memuat badge status serta CTA “Jelajahi Demo”. Status JejakAlumni adalah “Demo Konsep · Data Simulasi”.

Sampul JejakAlumni memakai `jejak-alumni-konseptual.svg`, menggambarkan catatan alumni menuju grafik tren dan ringkasan masukan tanpa identitas, angka, atau hasil nyata. Kanvas 960 × 600, latar muda, panel putih, garis navy/biru, serta label “Ilustrasi konseptual” mengikuti gaya sampul Implan Gigi dan DocNext. Sampul ini menggantikan tangkapan layar pada kartu saja; halaman demo dan data simulasinya tetap.

**Page demo DocNext**

Bagian editorial mengikuti pola situs, sedangkan ruang kerja dokumen memakai gaya aplikasi yang lebih fungsional: latar `#F1F7FC`, panel putih bergaris tipis, dan angka rata. Pada ≥64rem, dokumen dan analisis memakai rasio 60:40; di bawahnya satu kolom dengan urutan skenario, dokumen, hasil ekstraksi, perbandingan, dan temuan. Tiga dokumen sintetis tidak diperkecil berdampingan, melainkan dipilih satu per satu, dan tiap jenis dokumen memiliki tata letak sendiri. Sorotan sumber memakai outline dan penanda “▸”, bukan warna saja. Status selalu memakai ikon dan teks: ✓ Sesuai (`#0C7C6C`), ! Perlu Ditinjau (`#FF6B2C`), ? Perlu Verifikasi (`#F3D502`). Tidak ada progres OCR atau animasi yang mengesankan pemrosesan.

**Page demo Perencanaan Implan Gigi**

Halaman berarah Hybrid Editorial Research Showcase: pengantar dan penjelasan lebih dominan daripada kontrol. Area demo memuat stepper empat tahap, ilustrasi dan kontrol di kolom kiri, serta penjelasan di kolom kanan pada ≥64rem (65:35); di bawahnya satu kolom dengan urutan stepper, ilustrasi, kontrol, penjelasan, dan navigasi. Stepper menjadi 2×2 di bawah 48rem. Detail teknis memakai `<details>`.

Ilustrasi berupa SVG statis berlatar terang agar tidak menyerupai citra CBCT. Warna tetap di semua tahap: Alveolar Bone `#2547F4` dengan arsir, Mandibular Canal `#0C8C78` dengan garis putus-putus di tengah, area target `#FF6B2C` dengan tepi putus-putus `#B23F0D`, implan `#0A2240`, dan struktur lain abu-abu. Tiga warna pertama lolos validator palet, dan oranye selalu berlabel karena kontrasnya di bawah 3:1. Keterangan panjang ada di legenda HTML karena teks di dalam SVG mengecil pada HP. Setiap ilustrasi memuat label “Ilustrasi Konseptual”. Tahap 4 memakai tiga sudut pandang yang disiapkan dari satu geometri, tanpa WebGL atau dependensi baru, dan disebut simulasi visualisasi 3D konseptual. Gerak hanya berupa perubahan opasitas singkat dan mengikuti `prefers-reduced-motion`.

**Page demo JejakAlumni**

Halaman berarah dasbor analitik ringan: latar `#F1F7FC`, panel putih bergaris tipis, dan bayangan minimal. Desktop (≥64rem) memakai dasbor dan penjelasan 65:35; di bawahnya satu kolom dengan dasbor lebih dulu. Jumlah kolom KPI dan grafik mengikuti lebar bingkai dasbor lewat container query, bukan lebar layar. Grafik penyerapan memakai satu warna `#2547F4`; saat satu jurusan dipilih, keenam batang tetap tampil dan jurusan itu ditonjolkan. Status alumni memakai batang tersegmen tipis dan daftar berlabel: Bekerja `#2547F4`, Kuliah `#FF6B2C`, Wirausaha `#0C8C78`, Belum bekerja `#D4A900`. Dua warna terakhir adalah turunan khusus grafik karena `#0C7C6C` dan `#F3D502` tidak lolos validator palet kategori; dua warna dengan kontras di bawah 3:1 terhadap latar selalu disertai label teks. Nilai grafik selalu terlihat, tanpa animasi loading yang mengesankan inferensi langsung. Implementasi prototype memakai CSS dan JavaScript murni, tanpa dependensi baru.

**Page Galeri Dokumentasi**

Pada fase prototype, halaman **Galeri Dokumentasi** memakai arah **Clean Editorial Gallery**: pengantar singkat, pencarian album, grid foto, dan pagination yang sudah ada. Setiap kartu hanya memuat sampul, tanggal publikasi dan badge jumlah foto pada satu baris metadata, lalu judul album. Jumlah foto diambil dari koleksi gambar album; gambar pertama menjadi sampul. Seluruh kartu menaut ke detail album, tanpa deskripsi atau tag kategori.

Daftar tetap menampilkan maksimal **9 album per halaman**, sesuai `requirement.md` §4.4. Grid memakai tiga kolom pada desktop, dua pada tablet, dan satu pada HP; jumlah album lebih sedikit tidak diisi kartu tambahan. Foto memakai rasio tampilan konsisten, dengan hover ringan pada perangkat yang mendukung hover, fokus keyboard yang terlihat, dan `prefers-reduced-motion` yang dihormati. Data album dan gambar contoh tetap dipakai selama prototyping; konten produksi harus memakai dokumentasi kegiatan yang telah diverifikasi. Halaman detail serta rute teknis yang ada tidak berubah.

**Page Berita**

Halaman Berita memakai arah **Minimal Editorial News**, dengan [AI Singapore News](https://aisingapore.org/news/) sebagai acuan struktur pengantar dan keterbacaan daftar artikel. Warna, huruf, komponen, navbar, dan footer tetap mengikuti identitas UBAYA AI Center; tata letak situs referensi tidak disalin.

Pada fase prototype, halaman dibuka dengan pengantar “Kabar Terbaru dari UBAYA AI Center.” lalu langsung menampilkan daftar berita. Kartu berisi gambar, tanggal, judul, ringkasan dari paragraf pertama, dan satu tautan ke detail; tidak ada tag kategori atau bagian berita unggulan. Grid memakai dua kolom pada layar lebar dan satu kolom pada HP, dengan lebar kartu dibatasi ketika hanya ada satu hasil. Pencarian judul minimal dua karakter dan pagination sembilan berita per halaman tetap mengikuti `requirement.md` §4.5; pagination hanya tampil bila diperlukan.

Gambar dan isi berita yang dipublikasikan harus sesuai peristiwa atau jelas berfungsi sebagai ilustrasi. Data berita contoh dalam `prototype/assets/mock-data.js` tetap dipakai selama fase prototyping dan tidak dianggap artikel terverifikasi untuk produksi. Interaksi kartu memakai umpan balik hover ringan pada perangkat yang mendukung hover, fokus keyboard yang terlihat, serta penghormatan pada `prefers-reduced-motion`; tidak ada dependensi animasi baru. Halaman detail tetap memakai struktur yang ada: judul, tanggal, gambar, isi, dan tautan kembali. Penyempurnaan detail dan metadata produksi mengikuti `stack.md` §3 dan ditinjau terpisah.

**Page tim**

Page ini menggunakan referensi display dari [https\://mohfarid99.github.io/ubaya-ai-center/tentang-ai-center.html\#tim](https://mohfarid99.github.io/ubaya-ai-center/tentang-ai-center.html#tim) .

Pada fase prototype, halaman Tim memakai arah **Editorial Team Showcase**: pengantar singkat, pencarian nama/peran/bidang, enam profil sesuai urutan organisasi, lalu CTA kontak. Kartu menampilkan foto, nama dan gelar, peran, serta bidang riset dalam grid tiga kolom di desktop, dua di tablet, dan satu di HP. Bidang riset pada data prototype ditetapkan pemilik tetapi masih perlu verifikasi sebagai keahlian pribadi sebelum produksi.

Penutup mengajak pengunjung berdiskusi melalui satu tombol “Hubungi via WhatsApp” ke kanal resmi UBAYA AI Center yang sudah dipakai halaman layanan. CTA ditempatkan setelah seluruh profil, bukan di tengah direktori. Umpan balik tombol memakai pola yang ada; pencarian memperbarui hasil saat mengetik tanpa animasi grid atau dependensi baru.

**Page about**

Page ini menggunakan referensi pada [https\://chalk.ai/about](https://chalk.ai/about)

Pada fase prototype, halaman Tentang Kami memakai arah **Warm Editorial Storytelling** dengan alur ringkas: pengantar UBAYA AI Center, visi dan misi, fasilitas, lalu pilihan layanan. Hero memakai judul “Mendekatkan Inovasi AI dengan Kebutuhan Nyata.” Visi resmi dan lima poin misi dipertahankan maknanya. Fasilitas menampilkan foto ruang pelatihan dan ruang diskusi; foto ruang demo diganti placeholder “Foto Ruangan Coming Soon”. Bagian penutup berjudul “Langkah Berikutnya Dimulai di Sini.” Dua kartu Pelatihan dan Inference and Solutions menjadi CTA utama; masing-masing memuat deskripsi singkat dan langsung menuju halaman layanan terkait. Tombol umum “Lihat Layanan Kami” dihapus karena mengulang pilihan yang sudah tersedia.

Komposisi menggunakan hierarki tipografi dan susunan fasilitas yang lebih editorial; kolom ditumpuk pada layar kecil agar urutan baca tetap jelas. Kartu layanan tampil dua kolom pada layar lebar dan bertumpuk pada layar kecil, dengan keadaan hover dan fokus yang jelas. Motion khusus halaman dibatasi pada masuknya teks hero, reveal bagian sekali saat terlihat, serta hover foto dan panah kartu layanan pada perangkat yang mendukung hover. `prefers-reduced-motion` dihormati, dan konten tetap terlihat tanpa efek scroll. Implementasi prototype menggunakan CSS dan JavaScript yang sudah ada, tanpa dependensi animasi tambahan.

## 3. Warna, Huruf

* Mode terang saja  
* Base Color: Blue nuance  
* Arah Design Overall: Professional & Approachable(Profesional tetapi tetap ramah dan mudah dipahami publik); Modern Corporate (Profesional, tepercaya, clean, dan formal)  
* Arah Design khusus page/layanan pelatihan: Digital Learning Experience (Edukatif, interaktif, dan berorientasi pada pengalaman belajar)  
* Arah Design khusus page/layanan AI 'Inference and Solutions': Enterprise Tech (Teknologi modern dengan fokus pada kebutuhan perusahaan)  
* Main Font: General Sans, Cabinet Grotesk  
* Additional Font: Geist Mono  
* Primary colors: `#FFFFFF`; `#D8EBF8`; `#0674FD`; `#2547F4`; `#0A2240`  
* Warna overlay: `#0000006B` (kontrasnya sekitar 3:1, jadi tidak dipakai sebagai teks sekunder)  
* Teks sekunder (turunan prototype): `#0A2240` dengan opasitas 72%  
* Secondary Color: `#3CFBC5`; `#F3D502`; `#0C7C6C`; `#FF6B2C`  
* Borders: adjust to needs.  
* Warna status: secondary color or Sesuaikan dengan arah desain.  
* Aturan: status wajib ada teks, bukan warna saja.  
* Kontras: Sesuaikan dengan arah desain. `#0674FD` di atas putih sekitar 4,3:1 sehingga tidak dipakai untuk teks biasa; `#FF6B2C` juga kurang untuk teks, sehingga pesan galat memakai turunan `#B23F0D`. Nama dan nilai token turunan (jarak, radius, bayangan, skala teks) sementara didefinisikan di `prototype/assets/css/tokens.css`.  
* Keluarga font: Sesuaikan dengan arah desain.  
* Skala ukuran: judul, isi, tabel, catatan kecil; Sesuaikan dengan arah desain.  
* Angka tabel rata.  
* Font: bebas lisensi.  
* Info penting (alasan tolak, error) gunakan kapitalisasi 'Sentence Case'.  
* Format tanggal yang ditampilkan: `DD-MM-YYYY` \[Lihat database.md §2\].

## 4. Tata letak

* Kerangka sidebar, top bar, lebar isi; Sesuaikan dengan arah desain.  
* Jarak: sesuaikan dengan arah desain.  
* Titik pecah layar untuk desktop, tablet, HP: sesuaikan dengan kebutuhan desain.  
* Web responsive desktop display to mobile: regular resizing.  
* Teks panjang: dipotong lalu tambahkan keterangan 'lihat detail'.  
* Intinya dibuat lebih responsive untuk menjangkau banyak website  
* Gutter samping hero (Beranda, Pelatihan, Inference) dibuat lebih lega daripada `.wadah` biasa. Di bawah ±1240 px teks hero tampak sedikit lebih menjorok ke dalam dibanding breadcrumb dan section di bawahnya; pemilik menerima perbedaan ini.  
* Tautan kontak di footer (WhatsApp, email, Instagram) tampil tanpa garis bawah, dengan ikon, dan garis bawah muncul saat hover atau fokus. Kata "WhatsApp" dan "Instagram" tetap ada untuk pembaca layar agar nama tautan tidak hilang.

## 5. Bentuk dan Layer

* Sudut tumpul untuk tiap komponenn  
* Bayangan: sesuaikan dengan arah desain.  
* Motion: Sesuaikan dengan arah desain.  
* Ukuran target sentuh, sesuaikan dengan arah desain.

## 6. Komponen: menyesuaikan arah desain

## 7. Peta halaman: menyesuaikan arah desain

## 8. Pola per layar: menyesuaikan arah desain

## 9. Referensi lain yang bisa di akses

* [https\://dribbble.com/shots/27457563-Venture-Capital-Landing-Page](https://dribbble.com/shots/27457563-Venture-Capital-Landing-Page)  
* [https\://dribbble.com/shots/26983214-Health-Website-Design](https://dribbble.com/shots/26983214-Health-Website-Design)  
* [https\://www\.altoira.com/](https://www.altoira.com/)  
* [https\://www\.aptedge.io/](https://www.aptedge.io/)  
* [https\://chalk.ai/](https://chalk.ai/)  
* [https\://www\.ruangguru.com/rea](https://www.ruangguru.com/rea)  
* [https\://lppm.ubaya.ac.id/aicenter/](https://lppm.ubaya.ac.id/aicenter/)  
* [https\://hub.orca.id/marketplace?\_gl=1\*63j1jh\*\_ga\*MTUyMjY2NzkxMy4xNzkzMzU4MTI2\*\_ga\_M905QM25JL\*czE3OTMzNTgxMjUkbzEkZzEkdDE3OTMzNTgxMzckajQ4JGwwJGgw](https://hub.orca.id/marketplace?_gl=1*63j1jh*_ga*MTUyMjY2NzkxMy4xNzkzMzU4MTI2*_ga_M905QM25JL*czE3OTMzNTgxMjUkbzEkZzEkdDE3OTMzNTgxMzckajQ4JGwwJGgw) (FIX HERO)  
* [https\://aici-umg.com/](https://aici-umg.com/) referensi utk konten pelatihan (penjelasan" terkait kelas dan kegiatan" dari aicen)  
* [https\://symboolic.com/en](https://symboolic.com/en) (navbar fix)
