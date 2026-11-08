# Requirement — AI Center UBAYA CMS

Dokumen aktif. Riwayat perencanaan awal (`00-detail-rencana-web.md`, `01-kontrak-api.md`) tersedia di riwayat Git, commit `5928f05` (tag `arsip-docs-v1`). File planning lain: `stack.md`, `design.md`, `database.md`, `security.md`.

## 1. Tujuan dan batas produk

**Website ini buat apa?** Website ini memiliki 2 sisi. Sisi utama (publik) untuk menampilkan konten kepada calon customer sekaligus menjadi wadah untuk mempromosikan atau branding, perusahaan kami yaitu AI Center Universitas Surabaya (UBAYA), dimana kami memiliki fokus penyediaan layanan berbasis AI ('Pelatihan AI' dan 'Inference and Solutions'). Sisi kedua (internal) digunakan oleh tim internal perusahaan untuk input/edit data perubahan konten yang ditampilkan di sisi utama (publik). Fokus website: pengelolaan konten publlik

**Siapa pemilik dan siapa pemakai?** Pemiliknya adalah AI Center UBAYA. Penggunanya adalah Tim/staff AI Center UBAYA (yang bisa mengedit konten), user publik (hanya read only)

**Masalah apa yang diselesaikan?** Content Management System untuk keperluan AI Center UBAYA, konten yang akan di tampilkan di dalam website mencakup: detail layanan pelatihan, detail layanan solution inference, detail berita (memiliki object image), detail dokumentasi (memiliki object image), display susunan tim/staff (memiliki object image), section portofolio (berupa hardcode beberapa sampel produk sistem yg dibuat), dan terdapat kontak section (button mengarah ke WA ataupun email). Bahasa dibuat soft selling: menjelaskan manfaat, bukan menekan. Kalimat singkat, ramah, dan mudah dipahami. Tampilan menyesuaikan arah desain, tetap informatif tanpa membuat user kewalahan/overwhelm ketika mengakses website.

**Apa yang bukan produk / jasa yang ditawarkan ai center?** AI Center menawarkan adalah Layanan 'Pelatihan' dan AI 'Inference and Solutions'). Website ini bukan e-commerce, bukan LMS, bukan media sosial atau forum, bukan marketplace, bukan aplikasi SaaS, dan bukan sistem transaksi atau data pribadi skala besar.

**Target user?** Sisi website yang tampil di publik ditujukan untuk publik. Sisi webside untuk admin ditujukan untuk internal.

## 2. Scope

**Apa yang termasuk scope?**

### Scope User

Sistem ini akan memiliki 2 sisi, yaitu publik dan internal. Dari sisi publik, sistem ini akan diakses oleh pengunjung anonim dan semua orang bisa baca. Dari sisi internal, staff internal bisa melakukan akses ke dashboard admin yang berfungsi untuk mengelola CRUD Dokumentasi, Berita, Tim, Testimoni, yang bisa juga melakukan unggah gambar (jpg/png/webp maks 2MB). Konten Pelatihan (Program Reguler, Program Khusus, Seminar / Workshop untuk Bisnis) hardcode di frontend pada tahap ini, hapus/logout akan selalu memakai konfirmasi. Kemudian, untuk pengelolaan login user akan dikelola oleh 1 tools sistem bernama keycloak. Dan memiliki 2 role yaitu direktur dan pegawai. \[menunggu konfirmasi Keycloak\]

Validasi upload: \[Lihat security.md §7\] dan \[Lihat database.md §3\].

## 3. Rute publik

Website dari sisi user akan memiliki navbar dan footer. Kemudian akan berisi:

* fitur Beranda /beranda,  
* Berita /berita \+ /berita/:slug: daftar \+ pencarian judul (server-side lewat parameter `?q=`, minimal 2 karakter, dengan pagination), detail dengan gambar.  
* Dokumentasi /dokumentasi \+ /dokumentasi/:slug. HTML detail disajikan Go dengan meta SEO dan data awal \[Lihat stack.md §3\]; data juga tersedia lewat `GET /api/dokumentasi/:slug`.  
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

## 4. Alur halaman publik

Referensi visual dan animasi tiap halaman: \[Lihat design.md §1 dan §2\].

### 4.1 Landing Page (Home)

Alur story telling pada page home adalah: hero, section jenis layanan, section dokumentasi, section testimoni klien (carousel), section berita, section kontak (tombol redirect ke wa ataupun email), footer yang berisi kontak (nomor whatsapp, email, akun instagram, alamat resmi, dan peta lokasi).

Section jenis layanan memuat dua card, yaitu 'Pelatihan' dan ''Inference and Solutions''. Tiap card terdiri dari: judul layanan, deskripsi singkat layanan, button detail layanan untuk mengakses page terkait.

Section dokumentasi akan menampilkan 3 unggahan dokumentasi terbaru (dari halaman detail dokumentasi), lalu ada button untuk menuju page daftar dokumentasi. Tiap preview item dokumentasi akan muncul dalam bentuk card yang berisi (foto headline, judul, tanggal publikasi dokumentasi terkait), lalu ada button untuk melihat detail dokumentasi yang dipilih.

Section testimoni klien (carousel otomatis, gerak per 10 detik). Per carousel akan tampilkan 1 testimoni.

Section berita menampilkan 3 unggahan berita terbaru (dari halaman detail berita), lalu ada button untuk menuju page daftar berita. Tiap preview item berita akan muncul dalam bentuk card yang berisi (foto headline, judul, potongan berita dari paragraf pertama artikel terkait), lalu ada button untuk melihat detail berita yang dipilih.

Section footer yang berisi kontak berupa nomor whatsapp, email, akun instagram, alamat resmi, dan peta lokasi.

### 4.2 Page pelatihan \-\> Detailing 3 Layanannya apa

Referensi page pelatihan adalah [https\://aici-umg.com/](https://aici-umg.com/) ; akan tetapi dengan alur story telling: hero (deskripsi singkat soal 3 layanan, lalu tombol 'Pelatihan Reguler' yang akan scroll ke section pelatihan reguler & 'Konsultasi Gratis' yang akan scroll ke section pelatihan khusus), section pelatihan reguler (ditulis coming soon), section pelatihan khusus (urutan tulisan: alur konsultasi seperti referensi di [https\://lppm.ubaya.ac.id/aicenter/](https://lppm.ubaya.ac.id/aicenter/) lalu, tombol untuk 'Konsultasi dengan kami') , section Workshop atau Seminar untuk Bisnis (urutan tulisan: deskripsi singkat, lalu tombol untuk 'Hubungi Kami'

### 4.3 Page 'Inference and Solutions' \-\> Inference & Solution

Alur story telling page 'Inference and Solutions' adalah: Hero (memuat carousel pertanyaan untuk menarik minat user untuk berkonsultasi dengan AI Center; arahnya agar user mengetahui urgensi mereka untuk konsultasi solution inference dengan AI Center), jenis solution inference di AI Center, alur solution inference, 3 mock up portofolio (hardcode), dan section kontak kami.

### 4.4 Page Dokumentasi

Page ini memuat semua dokumentasi yang sudah di post dari sisi admin

Alurnya adalah: header page, filter (filter bisa diisi huruf; by title), menampilkan 9 konten dokumentasi (jika lebih dari 9 maka akan tergeser ke pagination selanjutnya). Urutan dokumentasi teratas berdasarkan tanggal publikasi terbaru (waktu buat sebagai pemutus seri). pada card tiap item dokumentasi terdapat headline image dokumentasi, judul, kemudian tanggal publikasi dokumentasi. Tidak ada jenis tag untuk dokumentasi.

### 4.5 Page Berita

Page ini akan memuat semua berita yang sudah di post dari sisi admin.

Alurnya adalah: header page, filter (filter bisa diisi huruf; by title), menampilkan 9 konten berita (jika lebih dari 9 maka akan tergeser ke pagination selanjutnya). Urutan berita teratas berdasarkan tanggal publikasi terbaru (waktu buat sebagai pemutus seri). pada card tiap item berita terdapat headline image berita, judul, deskripsi berita yang diambil dari potongan paragraf pertama, kemudian tanggal publikasi berita. Tidak ada jenis tag untuk berita.

### 4.6 Page tim

Page ini memuat daftar tim yang dikelola dari sisi admin

tiap card memuat: foto, jabatan, nama anggota, bidang penelitian.

### 4.7 Page about

Rute /tentang-kami. Referensi visual: \[Lihat design.md §2\].

## 5. Detailing page yang ditampilkan untuk admin internal

Page admin memuat data-data terkait Berita, Dokumentasi, Testimoni, Tim.

### 5.1 Bagian Berita

ada section yang mendisplay semua berita yang ada di database sistem (status draft & posted), section untuk edit berita (memuat image highlight, judul, tanggal publikasi, konten berita), section untuk Trash Berita (apapun data berita yang di page ini akan dihapus dalam waktu 30 hari setelah dipindahkan ke trash ini). Saat status berita adalah 'draft', button pada tiap item berita (pada section display semua berita) ada button untuk 'edit', 'hapus', 'post';

Saat status 'posted', button pada tiap item berita (pada display) ada button 'retract' untuk mengembalikan status posted ke draft.

Bila item berita pada list dihapus, maka item tersebut akan pindah ke section 'trash berita' sebagai mekanisme soft delete. Semua item yang berada di sini lebih dari 30 hari (setelah dipindahkan) akan terhapus secara permanen.

Di section trash tersedia tombol 'Pulihkan' (kembali ke draft) dan 'Hapus permanen' (dengan konfirmasi).

### 5.2 Bagian Dokumentasi

ada section yang mendisplay semua dokumentasi yang ada di database sistem (status draft & posted), section untuk edit dokumentasi (upload multi image maksimum jumlah \= 6, judul kegiatan, tanggal publikasi), section untuk Trash Dokumentasi (apapun data dokumentasi yang di page ini akan dihapus dalam waktu 30 hari setelah dipindahkan ke trash ini).

Saat status dokumentasi adalah 'draft', button pada tiap item dokumentasi (pada section display semua dokumentasi) ada button untuk 'edit', 'hapus', 'post'; Saat status 'posted', button pada tiap item dokumentasi (pada display) ada button 'retract' untuk mengembalikan status posted ke draft.

Bila item dokumentasi pada list dihapus, maka item tersebut akan pindah ke section 'trash dokumentasi' sebagai mekanisme soft delete. Semua item yang berada di sini lebih dari 30 hari (setelah dipindahkan) akan terhapus secara permanen.

Di section trash tersedia tombol 'Pulihkan' (kembali ke draft) dan 'Hapus permanen' (dengan konfirmasi).

### 5.3 Bagian Testimoini

ada section yang mendisplay semua testimoni yang ada di database sistem (status draft & posted), section untuk edit testimoni, section untuk Trash Testimoni (apapun data testimoni yang di page ini akan dihapus dalam waktu 30 hari setelah dipindahkan ke trash ini).

Untuk testimoni itu field inputnya: Foto (opsional; tanpa foto tampil inisial nama), Nama pemberi testimoni, Teks testimoni, Posisi (jabatan atau mungkin posisi lainnya) pemberi testimoni

Saat status testimoni adalah 'draft', button pada tiap testimoni (pada section display semua testimoni) ada button untuk 'edit', 'hapus', 'post'; Saat status 'posted', button pada tiap testimoni (pada display) ada button 'retract' untuk mengembalikan status posted ke draft.

Bila item testimoni pada list dihapus, maka item tersebut akan pindah ke section 'trash testimoni' sebagai mekanisme soft delete. Semua item yang berada di sini lebih dari 30 hari (setelah dipindahkan) akan terhapus secara permanen.

Di section trash tersedia tombol 'Pulihkan' (kembali ke draft) dan 'Hapus permanen' (dengan konfirmasi).

Tampilan carousel testimoni: \[Lihat §4.1\] dan \[Lihat design.md §2\].

### 5.4 Bagian Tim

ada section yang mendisplay semua data tim dan section untuk edit tim.

Untuk data tim itu field inputnya: Foto, posisi / jabatan, nama, & bidang riset

## 6. Pengguna, peran, izin

Akun dan pemisahan role dibuat melalui keycloak. \[menunggu konfirmasi Keycloak\]

Daftar role (Direktur dan pegawai).

Pada website cms role hanya berbeda display role di profil saja. Akses dan editing sama. Role ditampilkan di header admin (nama \+ role), tanpa halaman profil terpisah.

Penegakan akses saat login: \[Lihat security.md §1\].

## 7. File planning terkait

* Stack, alur fase, routing, reverse proxy, kontrak API: `stack.md`  
* Referensi visual, warna, huruf, tata letak: `design.md`  
* Entitas, field, status, soft lock, gambar, cache: `database.md`  
* Autentikasi, session, CSRF, rate limit, sanitasi: `security.md`

## 8. Kriteria selesai per fase

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

Dokumen kontrak API: \[Lihat stack.md §9\] memuat endpoint, skema JSON, format error, contoh mock, dan environment variable.

Alur fase kerja: \[Lihat stack.md §1\].
