# Design — AI Center UBAYA CMS

Dokumen aktif. Riwayat perencanaan awal (`00-detail-rencana-web.md`, `01-kontrak-api.md`) tersedia di riwayat Git, commit `5928f05` (tag `arsip-docs-v1`). Alur konten halaman: \[Lihat requirement.md §4 dan §5\].

## 1. Navbar dan animasi global

Navigation Bar menggunakan referensi dari tautan: [https\://symboolic.com/en](https://symboolic.com/en)

Animasi untuk interaksi button, card, pada page secara global: Fase prototype: CSS/JS murni. Fase stack: GSAP dan motion.dev \[Lihat stack.md §2\]. Pengecualian: animasi bola neural pada hero Beranda memakai canvas 2D dengan loop `requestAnimationFrame` sendiri (bukan GSAP/motion.dev); acuannya `prototype/assets/js/bola-neural.js`.

Hero Beranda berupa bidang navy dengan "bola neural" (canvas 2D): titik berbentuk bola yang berputar pelan dan "bernapas", sesekali satu node menyala mint beserta garis tipis ke tetangganya, dan titik di dekat kursor tertarik sedikit (tidak aktif di atas tautan atau CTA). Bola adalah lapisan latar yang tidak menerima klik, dengan scrim statis agar teks tetap terbaca. Perangkat tanpa hover atau pointer halus memakai titik lebih sedikit, frame rate lebih rendah, dan tanpa interaksi. `prefers-reduced-motion`: bola diam. Bila waktu kerja per frame terlalu tinggi, titik dikurangi, lalu bola diam. Animasi berhenti saat hero di luar layar atau tab tersembunyi. Angka rinci ada di `KONFIG` pada `bola-neural.js`; kinerja belum diukur di perangkat nyata.

Tidak ada tombol jeda untuk animasi hero, dan carousel (testimoni Beranda, pertanyaan Inference) juga tanpa tombol jeda. Ini penyimpangan disengaja dari WCAG 2.2.2 (Pause, Stop, Hide) atas keputusan pemilik. Untuk membatasinya, autoplay carousel hanya aktif di perangkat ber-hover dan berhenti saat disorot atau difokus; keringanan lain hanya `prefers-reduced-motion`.

## 2. Referensi visual per halaman

**Landing Page (Home)**

Copy Beranda memakai arah **Warm Editorial Storytelling**. Hero memperkenalkan manfaat pengembangan keterampilan dan penerapan AI bagi individu, institusi, serta bisnis. Kartu Pelatihan menekankan manfaat belajar bagi individu dan kelompok; kartu Inference and Solutions menekankan penerapan model dalam sistem atau proses kerja organisasi, termasuk integrasi dan pendampingan teknis. Kedua kartu tidak harus memberi penekanan yang sama pada setiap audiens.

CTA utama hero “Diskusikan Kebutuhan AI” membuka WhatsApp resmi, sedangkan “Lihat Layanan” menuju bagian layanan. Kartu layanan memakai “Lihat Pelatihan” dan “Lihat Solusi AI”; bagian Galeri Dokumentasi menaut ke daftar album. Judul testimoni “Cerita dari mereka yang belajar dan bekerja bersama kami” dipertahankan. Penutup menampung pertanyaan belajar maupun penerapan AI melalui kanal WhatsApp dan email yang sudah ada. Penyuntingan copy mempertahankan alur section, cakupan layanan, dan pola visual Beranda.

Pada page home, perpindahan dari satu section story telling ke section berikutnya menggunakan referensi animasi pada referensi [https\://brand.vettvangur.is/](https://brand.vettvangur.is/)

Section hero menggunakan referensi dari tautan: [https\://hub.orca.id/marketplace?\_gl=1\*63j1jh\*\_ga\*MTUyMjY2NzkxMy4xNzkzMzU4MTI2\*\_ga\_M905QM25JL\*czE3OTMzNTgxMjUkbzEkZzEkdDE3OTMzNTgxMzckajQ4JGwwJGgw](https://hub.orca.id/marketplace?_gl=1*63j1jh*_ga*MTUyMjY2NzkxMy4xNzkzMzU4MTI2*_ga_M905QM25JL*czE3OTMzNTgxMjUkbzEkZzEkdDE3OTMzNTgxMzckajQ4JGwwJGgw)

Referensi tambahan untuk section hero dapat diakses ke [https\://21st.dev/community/components/s/hero](https://21st.dev/community/components/s/hero),

Referensi utama untuk carousel testimoni ada di: [https\://www\.embla-carousel.com/docs/examples/predefined](https://www.embla-carousel.com/docs/examples/predefined) (note: bagian autoplay).

Tampilan preview berita di home page akan menggunakan referensi [https\://www\.altoira.com/](https://www.altoira.com/) pada bagian 'Articles'.

**Page Pelatihan:** referensi ada di dalam alur halaman \[Lihat requirement.md §4.2\].

Pada fase prototype, pengantar Pelatihan menjelaskan tiga pilihan: Pelatihan Reguler yang sedang disiapkan, Pelatihan Khusus untuk kebutuhan kelompok, serta Workshop atau Seminar untuk Bisnis. Tombol “Konsultasi Gratis” tetap menuju bagian Pelatihan Khusus sesuai `requirement.md` §4.2; CTA pada bagian Khusus dan Bisnis membuka kanal WhatsApp resmi dengan label sesuai topik. Empat langkah Pelatihan Khusus telah dikonfirmasi pemilik sebagai proses layanan, sehingga penanda data contoh pada alur dihapus. Perubahan ini menyunting bahasa dan hierarki informasi, bukan menambah jenis layanan atau jalur pendaftaran.

**Page Inference and Solutions**

Pada fase prototype, pengantar menjelaskan inference sebagai penggunaan model AI untuk menghasilkan prediksi pada data baru, lalu menghubungkannya dengan kebutuhan aplikasi atau sistem. IT Service Management dan Data Science tetap menjadi dua area awal diskusi; deskripsinya menggunakan bahasa yang tidak menjanjikan hasil tertentu. Lima tahap penerapan model AI telah dikonfirmasi pemilik sebagai proses layanan dan dipertahankan. Penutup mengajak pengunjung mendiskusikan kebutuhan AI melalui kanal kontak yang sudah ada. Nama resmi layanan dan tujuan CTA tidak berubah.

Bagian Portofolio memakai eyebrow “PORTOFOLIO”, judul “Jelajahi Portofolio Kami.”, dan grid kartu yang sudah ada: tiga kolom di desktop, mengikuti lebar di tablet, satu kolom di HP. Kartu JejakAlumni menaut penuh ke demo dan memuat badge “Prototipe · Data Simulasi” serta CTA “Coba Prototipe”. Kartu tanpa demo tidak memiliki tautan, CTA, atau efek angkat saat hover.

**Page demo JejakAlumni**

Halaman berarah dasbor analitik ringan: latar `#F1F7FC`, panel putih bergaris tipis, dan bayangan minimal. Desktop (≥64rem) memakai dasbor dan penjelasan 65:35; di bawahnya satu kolom dengan dasbor lebih dulu. Jumlah kolom KPI dan grafik mengikuti lebar bingkai dasbor lewat container query, bukan lebar layar. Grafik penyerapan memakai satu warna `#2547F4`; saat satu jurusan dipilih, keenam batang tetap tampil dan jurusan itu ditonjolkan. Status alumni memakai batang tersegmen tipis dan daftar berlabel: Bekerja `#2547F4`, Kuliah `#FF6B2C`, Wirausaha `#0C8C78`, Belum bekerja `#D4A900`. Dua warna terakhir adalah turunan khusus grafik karena `#0C7C6C` dan `#F3D502` tidak lolos validator palet kategori; dua warna dengan kontras di bawah 3:1 terhadap latar selalu disertai label teks. Nilai grafik selalu terlihat, tanpa animasi loading yang mengesankan inferensi langsung. Implementasi prototype memakai CSS dan JavaScript murni, tanpa dependensi baru.

**Page Galeri Dokumentasi**

Pada fase prototype, halaman **Galeri Dokumentasi** memakai arah **Clean Editorial Gallery**: pengantar singkat, pencarian album, grid foto, dan pagination yang sudah ada. Setiap kartu hanya memuat sampul, tanggal publikasi dan badge jumlah foto pada satu baris metadata, lalu judul album. Jumlah foto diambil dari koleksi gambar album; gambar pertama menjadi sampul. Seluruh kartu menaut ke detail album, tanpa deskripsi atau tag kategori.

Daftar tetap menampilkan maksimal **9 album per halaman**, sesuai `requirement.md` §4.4. Grid memakai tiga kolom pada desktop, dua pada tablet, dan satu pada HP; jumlah album lebih sedikit tidak diisi kartu tambahan. Foto memakai rasio tampilan konsisten, dengan hover ringan pada perangkat yang mendukung hover, fokus keyboard yang terlihat, dan `prefers-reduced-motion` yang dihormati. Data album dan gambar contoh tetap dipakai selama prototyping; konten produksi harus memakai dokumentasi kegiatan yang telah diverifikasi. Halaman detail serta rute teknis yang ada tidak berubah.

**Page Berita**

Halaman Berita memakai arah **Minimal Editorial News**, dengan [AI Singapore News](https://aisingapore.org/news/) sebagai acuan struktur pengantar dan keterbacaan daftar artikel. Warna, huruf, komponen, navbar, dan footer tetap mengikuti identitas AI Center UBAYA; tata letak situs referensi tidak disalin.

Pada fase prototype, halaman dibuka dengan pengantar “Kabar Terbaru dari AI Center UBAYA.” lalu langsung menampilkan daftar berita. Kartu berisi gambar, tanggal, judul, ringkasan dari paragraf pertama, dan satu tautan ke detail; tidak ada tag kategori atau bagian berita unggulan. Grid memakai dua kolom pada layar lebar dan satu kolom pada HP, dengan lebar kartu dibatasi ketika hanya ada satu hasil. Pencarian judul minimal dua karakter dan pagination sembilan berita per halaman tetap mengikuti `requirement.md` §4.5; pagination hanya tampil bila diperlukan.

Gambar dan isi berita yang dipublikasikan harus sesuai peristiwa atau jelas berfungsi sebagai ilustrasi. Data berita contoh dalam `prototype/assets/mock-data.js` tetap dipakai selama fase prototyping dan tidak dianggap artikel terverifikasi untuk produksi. Interaksi kartu memakai umpan balik hover ringan pada perangkat yang mendukung hover, fokus keyboard yang terlihat, serta penghormatan pada `prefers-reduced-motion`; tidak ada dependensi animasi baru. Halaman detail tetap memakai struktur yang ada: judul, tanggal, gambar, isi, dan tautan kembali. Penyempurnaan detail dan metadata produksi mengikuti `stack.md` §3 dan ditinjau terpisah.

**Page tim**

Page ini menggunakan referensi display dari [https\://mohfarid99.github.io/ubaya-ai-center/tentang-ai-center.html\#tim](https://mohfarid99.github.io/ubaya-ai-center/tentang-ai-center.html#tim) .

Pada fase prototype, halaman Tim memakai arah **Editorial Team Showcase**: pengantar singkat, pencarian nama/peran/bidang, enam profil sesuai urutan organisasi, lalu CTA kontak. Kartu menampilkan foto, nama dan gelar, peran, serta bidang riset dalam grid tiga kolom di desktop, dua di tablet, dan satu di HP. Bidang riset pada data prototype ditetapkan pemilik tetapi masih perlu verifikasi sebagai keahlian pribadi sebelum produksi.

Penutup mengajak pengunjung berdiskusi melalui satu tombol “Hubungi via WhatsApp” ke kanal resmi AI Center UBAYA yang sudah dipakai halaman layanan. CTA ditempatkan setelah seluruh profil, bukan di tengah direktori. Umpan balik tombol memakai pola yang ada; pencarian memperbarui hasil saat mengetik tanpa animasi grid atau dependensi baru.

**Page about**

Page ini menggunakan referensi pada [https\://chalk.ai/about](https://chalk.ai/about)

Pada fase prototype, halaman Tentang Kami memakai arah **Warm Editorial Storytelling** dengan alur ringkas: pengantar AI Center UBAYA, visi dan misi, fasilitas, lalu pilihan layanan. Hero memakai judul “Mendekatkan Inovasi AI dengan Kebutuhan Nyata.” Visi resmi dan lima poin misi dipertahankan maknanya. Fasilitas menampilkan foto ruang pelatihan dan ruang diskusi; foto ruang demo diganti placeholder “Foto Ruangan Coming Soon”. Bagian penutup berjudul “Langkah Berikutnya Dimulai di Sini.” Dua kartu Pelatihan dan Inference and Solutions menjadi CTA utama; masing-masing memuat deskripsi singkat dan langsung menuju halaman layanan terkait. Tombol umum “Lihat Layanan Kami” dihapus karena mengulang pilihan yang sudah tersedia.

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
