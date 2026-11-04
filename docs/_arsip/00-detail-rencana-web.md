# **DISKUSI PLAN**

## **1\. Tujuan dan batas produk**

**Website ini buat apa?** Website ini memiliki 2 sisi. Sisi utama (publik) untuk menampilkan konten kepada calon customer sekaligus menjadi wadah untuk mempromosikan atau branding, perusahaan kami yaitu AI Center Universitas Surabaya (UBAYA), dimana kami memiliki fokus penyediaan layanan berbasis AI ('Pelatihan AI' dan 'Inference and Solutions'). Sisi kedua (internal) digunakan oleh tim internal perusahaan untuk input/edit data perubahan konten yang ditampilkan di sisi utama (publik). Fokus website: pengelolaan konten publlik

**Siapa pemilik dan siapa pemakai?** Pemiliknya adalah AI Center UBAYA. Penggunanya adalah Tim/staff AI Center UBAYA (yang bisa mengedit konten), user publik (hanya read only)

**Masalah apa yang diselesaikan?** Content Management System untuk keperluan AI Center UBAYA, konten yang akan di tampilkan di dalam website mencakup: detail layanan pelatihan, detail layanan solution inference, detail berita (memiliki object image), detail dokumentasi (memiliki object image), display susunan tim/staff (memiliki object image), section portofolio (berupa hardcode beberapa sampel produk sistem yg dibuat), dan terdapat kontak section (button mengarah ke WA ataupun email). Bahasa dibuat soft selling: menjelaskan manfaat, bukan menekan. Kalimat singkat, ramah, dan mudah dipahami. Tampilan menyesuaikan arah desain, tetap informatif tanpa membuat user kewalahan/overwhelm ketika mengakses website.

**Apa yang bukan produk / jasa yang ditawarkan ai center?** AI Center menawarkan adalah Layanan 'Pelatihan' dan AI 'Inference and Solutions'). Website ini bukan e-commerce, bukan LMS, bukan media sosial atau forum, bukan marketplace, bukan aplikasi SaaS, dan bukan sistem transaksi atau data pribadi skala besar.

**Apa yang termasuk scope?**

### **Scope User**

Sistem ini akan memiliki 2 sisi, yaitu publik dan internal. Dari sisi publik, sistem ini akan diakses oleh pengunjung anonim dan semua orang bisa baca. Dari sisi internal, staff internal bisa melakukan akses ke dashboard admin yang berfungsi untuk mengelola CRUD Dokumentasi, Berita, Tim, Testimoni, yang bisa juga melakukan unggah gambar (jpg/png/webp maks 2MB). Konten Pelatihan (Program Reguler, Program Khusus, Seminar / Workshop untuk Bisnis) hardcode di frontend pada tahap ini, hapus/logout akan selalu memakai konfirmasi. Kemudian, untuk pengelolaan login user akan dikelola oleh 1 tools sistem bernama keycloak. Dan memiliki 2 role yaitu direktur dan pegawai.

Website dari sisi user akan memiliki navbar dan footer. Kemudian akan berisi:

* fitur Beranda /beranda,  
* Berita /berita \+ /berita/:slug: daftar \+ pencarian judul (server-side lewat parameter `?q=`, dengan pagination), detail dengan gambar.  
* Dokumentasi /dokumentasi \+ /dokumentasi/:slug. Rute halaman ditangani frontend; data berdasarkan `:slug` diambil dari backend lewat API (`GET /api/dokumentasi/:slug`).  
* Tim /tim: susunan anggota \+ pencarian.  
* Tentang Kami /tentang-kami  
* Fitur konten yang akan berisi dokumentasi dan berita (masing-masing punya routing page sendiri)  
* Layanan: hanya dua rute tetap, /layanan/pelatihan dan /layanan/inference-solutions (tidak memakai /layanan/:slug). Isinya hardcode di frontend, tidak ada opsi menambah layanan, dan tidak ada endpoint backend untuk layanan.  
* List kursus /layanan/pelatihan: dibuatkan kartu untuk display pilihan pelatihan \[Ditunda: aktif setelah Pelatihan Reguler tidak lagi coming soon\]  
* List layanan dalam 'Inference and Solutions': dibuatkan kartu untuk display layanan konsultasi. Menampilkan dua kategori layanan utama yang dapat menjadi titik awal bagi pengunjung untuk mengenali kebutuhan konsultasi perusahaan. Dua kategori tersebut adalah IT Service Management dan Data Science  
* Detail kursus /layanan/pelatihan/:kode: fakta, hasil belajar, materi, instruktur. \[Ditunda: aktif setelah Pelatihan Reguler tidak lagi coming soon\]  
* Toggle bahasa ID/EN dihilangkan, jadi sistem hanya menggunakan bahasa indo, peta lokasi footer, halaman 404\.  
* List produk untuk menampilkan 3 contoh produk yang sudah digarap oleh tim AI center  
* Fitur umum dalam web termasuk: portofolio, list berita, list dokumentasi, paket layanan: Pelatihan dan 'Inference and Solutions', list tim AI Center  
* Kemudian tiap masuk ke dalam page lain selain beranda, akan menampilkan breadcrumb misalnya Beranda \> Pelatihan. Sehingga menampilkan lokasi yang seharusnya

Semua akan diawali dengan proses planning terlebih dahulu, lalu masuk ke proses prototyping yg dibuat menggunakan HTML, CSS, JS saja. Jangan melakukan implementasi langsung dengan menggunakan list stack fix. List stack fix hanya akan diimplementasikan ketika proses prototyping telah aku anggap selesai

### **Kriteria selesai per fase**

* Prototype (HTML, CSS, JS): semua halaman publik & admin bisa diklik; state kosong, error, konfirmasi, "sedang diedit oleh X", `409`, dan `429` tampil dengan data dummy; disetujui pemilik.  
* Frontend stack: semua halaman berjalan dengan data mock sesuai dokumen kontrak API.  
* Backend: semua endpoint di dokumen kontrak API berjalan dan lolos pengujian.  
* Integrasi: frontend memakai API asli (bukan mock) untuk semua halaman.  
* E2E: alur publik dan admin lolos pengujian end-to-end, termasuk:  
  * HTML awal berita tanpa JavaScript berisi title, OG, dan JSON-LD.  
  * Slug asing mengembalikan 404; berita yang di-Retract langsung 404.  
  * Request publik berulang mengembalikan 304/cache hit dan tidak menambah hitungan limit.  
  * Melewati batas mengembalikan 429 \+ Retry-After, dan UI menampilkan pesan yang wajar.  
  * Dua user admin dari IP yang sama tidak saling memengaruhi limit.  
  * Login gagal berulang pada satu akun dikunci oleh Keycloak; endpoint login backend dibatasi per IP.

Dokumen kontrak API: `plan/01-kontrak-api.md` memuat endpoint, skema JSON, format error, contoh mock, dan environment variable.

### **Detailing page yang ditampilkan ke public**

Navigation Bar menggunakan referensi dari tautan: [https\://symboolic.com/en](https://symboolic.com/en)

Animasi untuk interaksi button, card, pada page secara global: Fase prototype: CSS/JS murni. Fase stack: GSAP dan motion.dev (lihat List Stack).

#### **Landing Page (Home)**

Alur story telling pada page home adalah: hero, section jenis layanan, section dokumentasi, section testimoni klien (carousel), section berita, section kontak (tombol redirect ke wa ataupun email), footer yang berisi kontak (nomor whatsapp, email, akun instagram, dan almat resmi).

Pada page home, perpindahan dari satu section story telling ke section berikutnya menggunakan referensi animasi pada referensi [https\://brand.vettvangur.is/](https://brand.vettvangur.is/)

Section hero menggunakan referensi dari tautan: [https\://hub.orca.id/marketplace?\_gl=1\*63j1jh\*\_ga\*MTUyMjY2NzkxMy4xNzkzMzU4MTI2\*\_ga\_M905QM25JL\*czE3OTMzNTgxMjUkbzEkZzEkdDE3OTMzNTgxMzckajQ4JGwwJGgw](https://hub.orca.id/marketplace?_gl=1*63j1jh*_ga*MTUyMjY2NzkxMy4xNzkzMzU4MTI2*_ga_M905QM25JL*czE3OTMzNTgxMjUkbzEkZzEkdDE3OTMzNTgxMzckajQ4JGwwJGgw)

Section jenis layanan memuat dua card, yaitu 'Pelatihan' dan ''Inference and Solutions''. Tiap card terdiri dari: judul layanan, deskripsi singkat layanan, button detail layanan untuk mengakses page terkait.

Section dokumentasi akan menampilkan 3 unggahan dokumentasi terbaru (dari halaman detail dokumentasi), lalu ada button untuk menuju page daftar dokumentasi. Tiap preview item dokumentasi akan muncul dalam bentuk card yang berisi (foto headline, judul, tanggal publikasi dokumentasi terkait), lalu ada button untuk melihat detail dokumentasi yang dipilih.

Section testimoni klien (carousel otomatis, gerak per 10 detik): Referensi utama untuk carousel testimoni ada di: [https\://www\.embla-carousel.com/docs/examples/predefined](https://www.embla-carousel.com/docs/examples/predefined) (note: bagian autoplay). Per carousel akan tampilkan 1 testimoni.

Section berita menampilkan 3 unggahan berita terbaru (dari halaman detail berita), lalu ada button untuk menuju page daftar berita. Tampilan preview berita di home page akan menggunakan referensi [https\://www\.altoira.com/](https://www.altoira.com/) pada bagian 'Articles'. Tiap preview item berita akan muncul dalam bentuk card yang berisi (foto headline, judul, potongan berita dari paragraf pertama artikel terkait), lalu ada button untuk melihat detail berita yang dipilih.

Section footer yang berisi kontak berupa nomor whatsapp, email, akun instagram, alamat resmi, dan peta lokasi.

#### **Page pelatihan \-\> Detailing 3 Layanannya apa**

Referensi page pelatihan adalah [https\://aici-umg.com/](https://aici-umg.com/) ; akan tetapi dengan alur story telling: hero (deskripsi singkat soal 3 layanan, lalu tombol 'Pelatihan Reguler' yang akan scroll ke section pelatihan reguler & 'Konsultasi Gratis' yang akan scroll ke section pelatihan khusus), section pelatihan reguler (ditulis coming soon), section pelatihan khusus (urutan tulisan: alur konsultasi seperti referensi di [https\://lppm.ubaya.ac.id/aicenter/](https://lppm.ubaya.ac.id/aicenter/) lalu, tombol untuk 'Konsultasi dengan kami') , section Workshop atau Seminar untuk Bisnis (urutan tulisan: deskripsi singkat, lalu tombol untuk 'Hubungi Kami'

#### **Page 'Inference and Solutions' \-\> Inference & Solution**

Alur story telling page 'Inference and Solutions' adalah: Hero (memuat carousel pertanyaan untuk menarik minat user untuk berkonsultasi dengan AI Center; arahnya agar user mengetahui urgensi mereka untuk konsultasi solution inference dengan AI Center), jenis solution inference di AI Center, alur solution inference, 3 mock up portofolio (hardcode), dan section kontak kami.

#### **Page Dokumentasi**

Page ini memuat semua dokumentasi yang sudah di post dari sisi admin

Page ini menggunakan referensi dari [https\://www\.altoira.com/insights](https://www.altoira.com/insights). Alurnya adalah: header page, filter (filter bisa diisi huruf; by title), menampilkan 9 konten dokumentasi (jika lebih dari 9 maka akan tergeser ke pagination selanjutnya). Urutan dokumentasi teratas berdasarkan tanggal publikasi terbaru (waktu buat sebagai pemutus seri). pada card tiap item dokumentasi terdapat headline image dokumentasi, judul, kemudian tanggal publikasi dokumentasi. Tidak ada jenis tag untuk dokumentasi.

#### **Page Berita**

Page ini akan memuat semua berita yang sudah di post dari sisi admin.

Page ini menggunakan referensi display dari [https\://www\.altoira.com/insights](https://www.altoira.com/insights) . Alurnya adalah: header page, filter (filter bisa diisi huruf; by title), menampilkan 9 konten berita (jika lebih dari 9 maka akan tergeser ke pagination selanjutnya). Urutan berita teratas berdasarkan tanggal publikasi terbaru (waktu buat sebagai pemutus seri). pada card tiap item berita terdapat headline image berita, judul, deskripsi berita yang diambil dari potongan paragraf pertama, kemudian tanggal publikasi berita. Tidak ada jenis tag untuk berita.

#### **Page tim**

Page ini memuat daftar tim yang sudah di post dari sisi admin

Page ini menggunakan referensi display dari [https\://mohfarid99.github.io/ubaya-ai-center/tentang-ai-center.html\#tim](https://mohfarid99.github.io/ubaya-ai-center/tentang-ai-center.html#tim) . tiap card memuat: foto, jabatan, nama anggota, bidang penelitian.

#### **Page about**

Page ini menggunakan referensi pada [https\://chalk.ai/about](https://chalk.ai/about)

### **Detailing page yang ditampilkan untuk admin internal**

Page admin memuat data-data terkait Berita, Dokumentasi, Testimoni, Tim.

#### **Bagian Berita**

ada section yang mendisplay semua berita yang ada di database sistem (status draft & posted), section untuk edit berita (memuat image highlight, judul, tanggal publikasi, konten berita), section untuk Trash Berita (apapun data berita yang di page ini akan dihapus dalam waktu 30 hari setelah dipindahkan ke trash ini). Saat status berita adalah 'draft', button pada tiap item berita (pada section display semua berita) ada button untuk 'edit', 'hapus', 'post';

Saat status 'posted', button pada tiap item berita (pada display) ada button 'retract' untuk mengembalikan status posted ke draft.

Bila item berita pada list dihapus, maka item tersebut akan pindah ke section 'trash berita' sebagai mekanisme soft delete. Semua item yang berada di sini lebih dari 30 hari (setelah dipindahkan) akan terhapus secara permanen.

Di section trash tersedia tombol 'Pulihkan' (kembali ke draft) dan 'Hapus permanen' (dengan konfirmasi).

#### **Bagian Dokumentasi**

ada section yang mendisplay semua dokumentasi yang ada di database sistem (status draft & posted), section untuk edit dokumentasi (upload multi image maksimum jumlah \= 6, judul kegiatan, tanggal publikasi), section untuk Trash Dokumentasi (apapun data dokumentasi yang di page ini akan dihapus dalam waktu 30 hari setelah dipindahkan ke trash ini).

Saat status dokumentasi adalah 'draft', button pada tiap item dokumentasi (pada section display semua dokumentasi) ada button untuk 'edit', 'hapus', 'post'; Saat status 'posted', button pada tiap item dokumentasi (pada display) ada button 'retract' untuk mengembalikan status posted ke draft.

Bila item dokumentasi pada list dihapus, maka item tersebut akan pindah ke section 'trash dokumentasi' sebagai mekanisme soft delete. Semua item yang berada di sini lebih dari 30 hari (setelah dipindahkan) akan terhapus secara permanen.

Di section trash tersedia tombol 'Pulihkan' (kembali ke draft) dan 'Hapus permanen' (dengan konfirmasi).

#### **Bagian Testimoini**

ada section yang mendisplay semua testimoni yang ada di database sistem (status draft & posted), section untuk edit testimoni, section untuk Trash Testimoni (apapun data testimoni yang di page ini akan dihapus dalam waktu 30 hari setelah dipindahkan ke trash ini).

Untuk testimoni itu field inputnya: Foto (opsional; tanpa foto tampil inisial nama), Nama pemberi testimoni, Teks testimoni, Posisi (jabatan atau mungkin posisi lainnya) pemberi testimoni

Saat status testimoni adalah 'draft', button pada tiap testimoni (pada section display semua testimoni) ada button untuk 'edit', 'hapus', 'post'; Saat status 'posted', button pada tiap testimoni (pada display) ada button 'retract' untuk mengembalikan status posted ke draft.

Bila item testimoni pada list dihapus, maka item tersebut akan pindah ke section 'trash testimoni' sebagai mekanisme soft delete. Semua item yang berada di sini lebih dari 30 hari (setelah dipindahkan) akan terhapus secara permanen.

Di section trash tersedia tombol 'Pulihkan' (kembali ke draft) dan 'Hapus permanen' (dengan konfirmasi).

Referensi utama untuk carousel testimoni ada di: [https\://www\.embla-carousel.com/docs/examples/predefined](https://www.embla-carousel.com/docs/examples/predefined) (note: bagian auto play)

Per carousel akan tampilkan 1 testimoni.

#### **Bagian Tim**

ada section yang mendisplay semua data tim dan section untuk edit tim.

Untuk data tim itu field inputnya: Foto, posisi / jabatan, nama, & bidang riset

### **List Stack yang akan digunakan adalah sebagai berikut**

Referensi tambahan untuk section hero dapat diakses ke [https\://21st.dev/community/components/s/hero](https://21st.dev/community/components/s/hero),

**Frontend:**

* Astro \+ TypeScript: halaman publik. Mode render: statis; data diambil di browser lewat island (TanStack Query) dari API Go.  
* React \+ vite: dashboard admin (tempat penulisan dan pengelolaan konten).  
* Tailwind CSS versi terbaru: styling.  
* shadcn/ui: komponen UI.  
* GSAP dan motion.dev: animasi.  
* Embla-carousel: Library Carousel  
* TanStack Query: pengambilan data server dan cache di browser.  
* Bun: tooling dependency, build, dan pengujian frontend.

**Backend:**

* Go \+ Chi (framework Go): API (`/api/*`), logika bisnis, penyajian hasil build Astro (`/`) dan build admin React (`/admin`) dalam satu origin. \[RP-01\]\[RP-02\] Fallback ke index.html hanya untuk `/admin/*` (SPA admin). \[RP-07\] Rute publik tanpa fallback (lihat aturan /berita/{slug} dan /dokumentasi/{slug} di bawah).  
  Rute /berita/{slug} dan /dokumentasi/{slug}: Go menyisipkan title, meta description, canonical, OG/Twitter, JSON-LD, isi plain text (di-escape), dan data awal JSON ke kerangka Astro. Slug tidak ditemukan dibalas HTTP 404 \+ noindex. Path tanpa file statis dibalas 404.html dengan status 404. Hasil di-cache di Redis per slug dan di-invalidate saat Post dan Retract. SSR tidak dipakai (K-03 tetap). \[RP-12\] sitemap.xml dan robots.txt dilayani Go (slug valid saja).  
  SEO berita: gambar berita menjadi OG image; judul SEO dan deskripsi SEO opsional (default: judul, dan 150-160 karakter pertama isi). \[RP-11\]  
  SEO dokumentasi: gambar headline (urutan 1) menjadi OG image; judul SEO dan deskripsi SEO opsional (default: judul kegiatan, dan "Dokumentasi kegiatan {judul} — {tanggal publikasi}"). \[RP-11\]  
  Field SEO opsional dan tidak termasuk aturan "semua field wajib terisi"; form admin menyediakan kedua field ini.  
  Cara pembuatan slug: \[terbuka\] (usulan: otomatis dari judul, unik, dan tidak diubah setelah Post agar URL tidak rusak).  
  URL canonical dan OG memakai env `PUBLIC_BASE_URL`. \[RP-08\]  
* PostgreSQL: database utama.  
* RustFS (kompatibel S3): penyimpanan gambar dan media.  
* Redis: cache server dan rate limiter.  
* Cookie session HttpOnly \+ CSRF: autentikasi dan perlindungan permintaan admin.  
* E2E: \[tool, mis. Playwright\] untuk alur publik dan admin end-to-end.

**Status kode saat ini:** kode di repo (React+Vite, Bun+Elysia, PGlite, login password) bukan basis implementasi. Fase 3 dimulai dari proyek baru sesuai stack di atas; kode lama hanya referensi konten.

  ### **Rencana BACKEND**

Penggantian bahasa untuk logika backend menjadi Go, penambahan layer API ditengah-tengah frontend dan backend yang ditulis menggunakan Chi (framework Go), dan pengubahan struktur folder di /server agar sesuai dengan default struktur folder Go.

#### **Ketergantungan reverse proxy**

Tag \[RP-xx\] menandai bagian rencana yang bergantung pada reverse proxy atau infrastruktur di depan aplikasi Go, agar mudah dicari (grep `\[RP-`) dan dibahas bersama pihak infra. Reverse proxy hanya menentukan ke mana request dikirim berdasarkan prefix path; aplikasi yang menentukan isinya. RP-01, RP-02, dan RP-09 harus jelas sebelum backend dimulai; sisanya paling lambat terjawab sebelum fase integrasi.

| ID | Bagian | Isu | Status |
|---|---|---|---|
| RP-01 | List Stack: baris Go \+ Chi | Tabel routing prefix: /api/\*, /admin/\*, /media/\*, sitemap.xml, robots.txt, /berita/{slug}, /dokumentasi/{slug}, dan sisanya ke file statis Astro. | SEBAGIAN: ditentukan di Chi (Go); /media/\* bergantung pada RP-11. |
| RP-02 | List Stack: baris Go \+ Chi | Apakah backend menerima /api/x atau /x setelah diteruskan. | SELESAI: prefix /api dipertahankan. |
| RP-03 | Penerapan Redis: Rate limiter | IP asli klien: header yang diisi proxy (X-Forwarded-For / X-Real-IP) dan daftar IP proxy tepercaya, agar rate limiter membaca IP yang benar dan tidak bisa dipalsukan. | TERBUKA: bahas dengan infra. |
| RP-04 | Penerapan Redis: Rate limiter dan Cache server | Lokasi cache: lapisan cache yang dipakai (proxy, Redis aplikasi, CDN) dan path yang tidak boleh ter-cache (/admin, login). | TERBUKA: bahas dengan infra. |
| RP-05 | Penerapan Redis: Rate limiter | Rate limit proxy vs aplikasi: pembagian tugas limit kasar di proxy dan limit per user/akun di aplikasi, agar tidak dihitung dua kali. | TERBUKA: bahas dengan infra. |
| RP-06 | Penerapan Redis: Rate limiter; Sanitasi: Respons error | Respons 429 dan error: memastikan proxy tidak menimpa halaman error, status 429, dan header Retry-After. | TERBUKA: bahas dengan infra. |
| RP-07 | List Stack: baris Go \+ Chi | Fallback SPA: aturan fallback ke index.html agar refresh halaman tidak 404. Hanya relevan untuk SPA. | GUGUR: publik memakai Astro statis. Fallback hanya untuk /admin (ditangani Go). |
| RP-08 | List Stack: baris Go \+ Chi (`PUBLIC_BASE_URL`) | HTTPS dan header forward: siapa yang menghentikan TLS, dan bagaimana aplikasi tahu skema dan host asli (untuk canonical dan OG). | SEBAGIAN: pakai env PUBLIC_BASE_URL; detail TLS/header dibahas dengan infra. |
| RP-09 | Autentikasi dan Keamanan: Session | Origin, CORS, cookie: domain frontend, API, dan admin satu origin atau beda; memengaruhi CORS dan cookie session. | SELESAI: satu origin, tanpa CORS, cookie SameSite=Lax. Dev memakai proxy dev Vite/Astro ke Go. |
| RP-10 | Dokumentasi dan Berita: Data dan Validasi (batas 2 MB per file); timeout: \[belum ada bagian terkait\] | Upload dan timeout: batas ukuran file media dan lama request yang diizinkan (di proxy maupun Go). | TERBUKA: bahas dengan infra. |
| RP-11 | Penyimpanan Gambar; List Stack: baris Go \+ Chi (OG image) | File statis dan media: media/OG image dilayani lewat /media/\* di Go, atau bucket publik RustFS. OG image harus URL absolut yang bisa diakses crawler. | BELUM DIPUTUSKAN. |
| RP-12 | List Stack: baris Go \+ Chi (rute /berita/{slug} dan /dokumentasi/{slug}) | Injeksi meta SEO: penyisipan title, meta description, canonical, OG/Twitter, JSON-LD, dan data awal ke HTML kerangka; slug tidak ada dibalas 404. | DIPUTUSKAN: dikerjakan di backend Go (lihat N-03). |

#### **Race Condition (Soft Locking)**

Saat pengguna membuka form edit, backend mengunci konten tersebut (atas nama pengguna itu). Pengguna lain hanya bisa membaca dan melihat info siapa yang sedang mengedit.

Kunci punya TTL, diperpanjang otomatis (heartbeat), dan dilepas saat data disimpan atau edit dibatalkan.

**Penerapan kunci di Redis**

Key `lock:{jenis}:{id}` berisi pengguna pemegang kunci, dibuat hanya jika belum ada (`SET NX`). Hanya pemegang yang dapat memperpanjang (heartbeat) atau melepas. Kunci dilepas saat Simpan atau batal. TTL 5 menit, heartbeat tiap 1 menit. Heartbeat dihitung ke bucket tulis rate limiter (30 / menit per pengguna).

**Kapan kunci dipasang**

Edit hanya untuk data berstatus draft (konten yang sudah terbit di-Retract dulu). Saat form edit dibuka, data dikunci (soft lock). Aturan ini berlaku untuk Dokumentasi, Berita, dan Testimoni. Tim memakai soft lock tanpa aturan draft (lihat bagian Tim).

**Pengaman tambahan: optimistic locking**

Kolom `version` dicek saat Simpan. Semua perubahan diterapkan dalam satu request dan satu transaksi bersama pengecekan `version`, lalu `version` dinaikkan. Jika versi tidak cocok, backend mengembalikan `409 Conflict`.

### **Dokumentasi**

**Data Fetching dan Pagination**

\[Lihat dokumen kontrak API: `plan/01-kontrak-api.md`\]

**Data dan Validasi**

Semua field wajib terisi sejak penyimpanan pertama (hasilnya draft). Jumlah gambar minimal 1 dan maksimal 6\. Nama penulis: \[otomatis dari pengguna yang login\];

Upload divalidasi: tipe dicek dari isi file, ukuran maksimum 2 MB tiap file gambar. \[RP-10\] \[Lihat pada bagian sanitasi input dan output\]

**Penyimpanan Gambar**

\[Lihat dokumen kontrak API: `plan/01-kontrak-api.md`\]

**Tambah dan Edit**

Tambah: mengunggah 1-6 gambar; urutan upload menjadi `urutan`, gambar pertama menjadi headline.

Edit hanya untuk data berstatus draft (konten yang sudah terbit di-Retract dulu). \[Lihat pada bagian Race Condition (Soft Locking)\]

Operasi gambar saat edit:

* Replace: satu gambar diganti dan menempati `urutan` yang sama.  
* Hapus: satu atau beberapa gambar dihapus; `urutan` dirapikan menjadi 1..n (gambar berikutnya menjadi headline bila headline dihapus).  
* Tambah: gambar baru di urutan paling akhir.

Jumlah akhir harus tetap 1-6. Tanpa operasi gambar, gambar lama tetap dipakai.

Penguncian saat form edit dibuka dan pengecekan `version` saat Simpan (termasuk `409 Conflict`): \[Lihat pada bagian Race Condition (Soft Locking)\]

Urutan eksekusi: yang masuk, file dulu baru database; yang keluar, database dulu baru file.

1. Upload file baru ke RustFS.  
2. Transaksi database: hapus baris gambar lama, tambah baris baru, rapikan `urutan`, naikkan `version`.  
3. Setelah commit, hapus file lama di RustFS.

Jika gagal sebelum commit, file baru yang sudah terunggah dihapus kembali. File yatim dibersihkan job berkala (objek yang tidak ada di tabel gambar dan lebih tua dari 24 jam \[usulan; diputuskan pada fase backend\]).

Retract dan Hapus (soft delete) tidak menyentuh gambar. Hapus permanen (manual atau otomatis setelah 30 hari) menghapus baris dan semua file gambarnya.

**Index dan Cache**

\[Lihat dokumen kontrak API: `plan/01-kontrak-api.md`\]

Daftar dan detail di-cache di Redis (cache-aside \+ TTL 20 menit) dan di-invalidate saat Post dan Retract. Draft dan item terhapus tidak pernah masuk cache publik.

### **Berita**

**Data Fetching dan Pagination**

\[Lihat dokumen kontrak API: `plan/01-kontrak-api.md`\]

**Data dan Validasi**

Semua field wajib terisi sejak penyimpanan pertama (hasilnya draft), termasuk 1 gambar. Nama penulis: \[otomatis dari pengguna yang login\]; `created_by`/`updated_by` tetap dicatat dari session.

Upload divalidasi: tipe dicek dari isi file, ukuran maksimum 2 MB untuk gambar. \[RP-10\] Teks (`isi` disimpan sebagai teks biasa) \[Lihat pada bagian sanitasi input dan output\].

**Penyimpanan Gambar**

Satu berita memiliki tepat 1 gambar, disimpan sebagai nama objek di kolom `gambar` pada tabel `berita` (tanpa tabel tambahan dan tanpa `urutan`).

File di RustFS memakai nama unik yang tidak berubah: `berita/{id}/{uuid}.ext`. \[RP-11\]

**Tambah dan Edit**

Tambah: mengunggah 1 gambar.

Edit hanya untuk data berstatus draft. \[Lihat pada bagian Race Condition (Soft Locking)\] Satu-satunya operasi gambar adalah replace; gambar tidak dapat dihapus atau ditambah. Tanpa gambar baru, gambar lama tetap dipakai.

Penguncian saat form edit dibuka dan pengecekan `version` saat Simpan (termasuk `409 Conflict`): \[Lihat pada bagian Race Condition (Soft Locking)\]

Urutan eksekusi: yang masuk, file dulu baru database; yang keluar, database dulu baru file.

1. Upload gambar baru ke RustFS.  
2. Update database: `gambar` diganti nama objek baru, `version` dinaikkan, `updated_by` diisi.  
3. Setelah commit, hapus file lama di RustFS.

Jika gagal sebelum commit, file baru yang sudah terunggah dihapus kembali. File yatim dibersihkan oleh job berkala yang sama dengan dokumentasi.

Retract dan Hapus (soft delete) tidak menyentuh gambar. Hapus permanen (manual atau otomatis setelah 30 hari) menghapus baris dan file gambarnya.

**Index dan Cache**

\[Lihat dokumen kontrak API: `plan/01-kontrak-api.md`\]

### **Testimoni**

**Data dan Validasi**

Field wajib: nama pemberi testimoni, teks testimoni, posisi. Foto opsional (maksimal 1); tanpa foto tampil inisial nama. Upload divalidasi sama seperti berita (tipe dicek dari isi file, maksimal 2 MB). \[Lihat pada bagian sanitasi input dan output\]

**Penyimpanan Gambar**

Foto disimpan sebagai nama objek pada kolom `foto` (boleh kosong). File di RustFS: `testimoni/{id}/{uuid}.ext`. Urutan eksekusi, pembersihan file yatim, dan hapus permanen sama dengan berita.

**Status dan Edit**

Alur draft, post, retract, trash 30 hari sama dengan berita. Edit hanya untuk draft dengan soft lock dan pengecekan `version`. \[Lihat pada bagian Race Condition (Soft Locking)\]

**Cache**

Daftar testimoni posted di-cache di Redis (cache-aside \+ TTL 20 menit) dan di-invalidate saat Post dan Retract.

### **Tim**

**Data dan Validasi**

Field wajib: foto, posisi / jabatan, nama, bidang riset. Upload divalidasi sama seperti berita.

**Penyimpanan Gambar**

Foto disimpan sebagai nama objek pada kolom `foto`. File di RustFS: `tim/{id}/{uuid}.ext`. Urutan eksekusi sama dengan berita.

**Status, Edit, dan Hapus**

Tim tidak memiliki status draft/posted dan tidak memiliki trash. Saat form edit dibuka, data dikunci (soft lock) dan `version` dicek saat Simpan. Hapus langsung permanen (dengan konfirmasi) dan menghapus file fotonya.

**Cache**

Daftar tim di-cache di Redis (cache-aside \+ TTL 20 menit) dan di-invalidate saat tambah, edit, dan hapus.

### **Penerapan Redis**

Key dipisah dengan prefix: `session:`, `lock:`, `ratelimit:`, `cache:`. Semua TTL dan batas disimpan sebagai konfigurasi (environment variable)

**Session login:** Key `session:{id}` berisi `sub`, nama, token CSRF, dan waktu dibuat. Cookie hanya berisi ID acak (256 bit). Dibuat setelah login Keycloak berhasil, dihapus saat logout. Pengguna yang dinonaktifkan di Keycloak tetap punya session sampai kedaluwarsa \[Diputuskan pada fase backend: validasi ulang ke Keycloak berkala\]. Idle 1 jam (diperpanjang tiap request), batas mutlak 8 jam; cookie dibuat berumur sama dengan batas mutlak.

**Soft lock:** \[Lihat pada bagian Race Condition (Soft Locking)\]

**Rate limiter:** Penghitung per jendela 1 menit (`INCR` \+ `EXPIRE`) pada key `ratelimit:{bucket}:{id}`. Tiga bucket: baca publik (daftar, detail, pencarian), baca admin (halaman admin), dan tulis admin (tambah, edit, Post, Retract, hapus, restore, hapus permanen, dan heartbeat soft lock). Jika melebihi batas, backend membalas `429 Too Many Requests` dengan header `Retry-After`. Baca admin: 180 / menit per user ID. Tulis admin: 30 / menit per user ID. Baca publik: dilayani dari cache Redis dengan Cache-Control \+ ETag; limit per IP hanya pada cache miss, \[N\] / menit (env, nilai awal 300; 1.200 bila semua request dihitung). Endpoint mulai login dan callback: batas ketat per IP; per akun ditangani Keycloak. \[RP-05\] /admin dan login tidak di-cache. \[RP-04\] 429 menyertakan Retry-After; frontend menampilkan pesan jelas dan tidak retry agresif. \[RP-06\] IP klien dibaca dari X-Forwarded-For hanya dari proxy tepercaya. \[RP-03\]

**Cache server:** Pola cache-aside \+ TTL untuk daftar dan detail berita/dokumentasi; aturan invalidasi (Post dan Retract) \[Lihat pada bagian dokumentasi dan berita\]. \[RP-04\]

### **Autentikasi dan Keamanan**

Keycloak mengelola pengguna dan role; aplikasi tidak punya manajemen role sendiri. Semua pengguna yang berhasil login dapat menambah, mengubah, dan menghapus konten. Hanya pengguna dengan role `direktur` atau `pegawai` pada client CMS yang boleh login; lainnya `403`.

**Login:** browser diarahkan ke Keycloak (OIDC Authorization Code \+ PKCE). Setelah berhasil, backend menukar kode dengan token di sisi server dan memvalidasi `state`, penerbit (issuer), tanda tangan, dan masa berlaku.

**Session:** backend membuat session di Redis (detail pada bagian Penerapan Redis). Browser hanya menerima ID session acak lewat cookie `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/` (nama `__Host-session`). Nama cookie dan flag `Secure` diatur lewat environment variable: di production wajib `__Host-session` \+ `Secure` (server menolak start bila tidak); di development boleh nama tanpa prefix `__Host-` dan tanpa `Secure` agar bisa berjalan di http lokal. Mode ditentukan `APP_ENV`; override nama cookie dan `Secure` hanya berlaku bila `APP_ENV=development`; token Keycloak tidak pernah dikirim ke browser. ID session dibuat baru setiap login. Satu origin dengan frontend, sehingga tidak perlu konfigurasi CORS. \[RP-09\] Cookie berumur sama dengan batas mutlak session \[Lihat pada bagian penerapan redis\]. Session yang tidak ditemukan atau sudah kedaluwarsa menghasilkan `401`, lalu frontend mengarahkan pengguna ke login Keycloak.

**CSRF:** setiap session memiliki token CSRF sendiri (disimpan di session Redis). Frontend mengirimkannya di header `X-CSRF-Token` pada setiap request POST/PUT/DELETE; backend membandingkannya dengan token di session dan memeriksa header `Origin`. Tidak cocok menghasilkan `403`.

**Logout:** session dihapus dari Redis dan cookie dibersihkan. \[Lihat pada bagian penerapan redis\]

**Proteksi login:** password dimasukkan di Keycloak, sehingga pembatasan percobaan login (brute force detection) diatur di Keycloak. Pembatasan percobaan login per akun memakai Brute Force Detection di Keycloak \[nilai: diisi\]. Backend hanya membatasi endpoint mulai login dan callback per IP.

`created_by` / `updated_by` dicatat dari identitas (`sub`) di session sebagai jejak audit.

### **Sanitasi Input dan Output**

* Query database: selalu berparameter (tidak menyambung string SQL dari input user) untuk mencegah SQL injection.  
* Search box: (sarankan review security audit).  
* Parameter request dipastikan bertipe benar, `page` minimal 1, `limit` dibatasi maksimum, dan nilai yang tidak valid ditolak dengan `400`.  
* Konten teks (judul, `isi`): disimpan sebagai teks biasa. Paragraf dipisah baris kosong (`\n\n`); potongan pada kartu \= paragraf pertama, dipotong lalu diberi keterangan 'lihat detail'. Backend melakukan trim, membatasi panjang, dan membuang karakter kontrol. Pencegahan XSS dilakukan saat ditampilkan: frontend merender konten sebagai teks (bukan HTML mentah), dan tidak boleh memakai `innerHTML` atau `set:html` untuk konten ini.  
* Upload file ke RustFS: tipe file dicek dari isi file (bukan hanya ekstensi), ukuran dibatasi, dan nama file diganti dengan nama acak agar tidak terjadi path traversal atau tabrakan nama.  
* Respons error: tidak menampilkan detail internal (stack trace, query, struktur database) ke client; detailnya hanya dicatat di log. \[RP-06\]

Dokumen perencanaan berada di /plan; dokumentasi teknis lain (.md) berada di /docs, sehingga folder dapat lebih rapi

**Target user?** Sisi website yang tampil di publik ditujukan untuk publik. Sisi webside untuk admin ditujukan untuk internal.

## **2\. Pengguna, peran, izin**

Akun dan pemisahan role dibuat melalui keycloak.

Daftar role (Direktur dan pegawai).

Pada website cms role hanya berbeda display role di profil saja. Akses dan editing sama. Role ditampilkan di header admin (nama \+ role), tanpa halaman profil terpisah.

## **3\. Warna, Huruf**

* Mode terang saja  
* Base Color: Blue nuance  
* Arah Design Overall: Professional & Approachable(Profesional tetapi tetap ramah dan mudah dipahami publik); Modern Corporate (Profesional, tepercaya, clean, dan formal)  
* Arah Design khusus page/layanan pelatihan: Digital Learning Experience (Edukatif, interaktif, dan berorientasi pada pengalaman belajar)  
* Arah Design khusus page/layanan AI 'Inference and Solutions': Enterprise Tech (Teknologi modern dengan fokus pada kebutuhan perusahaan)  
* Main Font: General Sans, Cabinet Grotesk  
* Additional Font: Geist Mono  
* Primary colors: `#FFFFFF`; `#D8EBF8`; `#0674FD`; `#2547F4`; `#0A2240`  
* Warna overlay / teks sekunder: `#0000006B`  
* Secondary Color: `#3CFBC5`; `#F3D502`; `#0C7C6C`; `#FF6B2C`  
* Borders: adjust to needs.  
* Warna status: secondary color or Sesuaikan dengan arah desain.  
* Aturan: status wajib ada teks, bukan warna saja.  
* Kontras: Sesuaikan dengan arah desain.  
* Keluarga font: Sesuaikan dengan arah desain.  
* Skala ukuran: judul, isi, tabel, catatan kecil; Sesuaikan dengan arah desain.  
* Angka tabel rata.  
* Font: bebas lisensi.  
* Info penting (alasan tolak, error) gunakan kapitalisasi 'Sentence Case'.

  ## **4\. Tata letak**

* Kerangka sidebar, top bar, lebar isi; Sesuaikan dengan arah desain.  
* Jarak: sesuaikan dengan arah desain.  
* Titik pecah layar untuk desktop, tablet, HP: sesuaikan dengan kebutuhan desain.  
* Web responsive desktop display to mobile: regular resizing.  
* Teks panjang: dipotong lalu tambahkan keterangan 'lihat detail'.  
* Intinya dibuat lebih responsive untuk menjangkau banyak website

  ## **5\. Bentuk dan Layer**

* Sudut tumpul untuk tiap komponenn  
* Bayangan: sesuaikan dengan arah desain.  
* Motion: Sesuaikan dengan arah desain.  
* Ukuran target sentuh, sesuaikan dengan arah desain.

  ## **6\. Komponen: menyesuaikan arah desain**

  ## **7\. Peta halaman: menyesuaikan arah desain**

  ## **8\. Pola per layar: menyesuaikan arah desain**

  ## **Referensi lain yang bisa di akses**

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
1. 

