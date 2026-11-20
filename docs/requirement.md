# Requirement — UBAYA AI Center CMS

## Log Aktivitas — disetujui 11 November 2026

Bagian admin baru `#/log-aktivitas` menampilkan Tanggal & Waktu, Aktivitas, dan Username, terbaru dahulu. Baca-saja, tanpa tambah/edit/hapus log. Mendukung filter username, jenis aktivitas dan rentang hari Asia/Jakarta; pagination 10 entri per halaman khusus log (daftar CRUD tetap mengikuti aturan sebelumnya).

Prototype mencatat otomatis setelah mutasi API tiruan berhasil: tambah, ubah, hapus ke trash, hapus permanen, Post, Retract, dan Restore pada Berita/Dokumentasi/Testimoni; tambah/ubah/hapus permanen pada Tim. Post/Retract/Restore termasuk jenis Ubah Data. Perubahan foto mengikuti penyimpanan item induk. Batal, pembacaan, heartbeat, dan perubahan gagal tidak dicatat. Tidak ada entri riwayat palsu pada awal halaman.

Status saat ini **simulasi frontend**, bukan audit log server-side: log tersimpan di memori instance halaman, hilang setelah refresh/logout yang memuat ulang halaman, tidak dibagikan antar-tab/pengguna/perangkat. Username berasal dari akun aktif contoh, bukan isian form; akun ini belum terautentikasi nyata. Waktu memakai jam perangkat dan ditampilkan dalam WIB. Label Simulasi dipertahankan dan keterbatasan dijelaskan pada halaman.

Audit log produksi memerlukan identitas sesi terverifikasi, waktu server, penyimpanan PostgreSQL transaksional dan pembatasan akses log. Belum diimplementasikan; lihat `database.md`, `security.md`, dan `stack.md` bagian Log Aktivitas.

Dokumen aktif. Riwayat perencanaan awal (`00-detail-rencana-web.md`, `01-kontrak-api.md`) tersedia di riwayat Git, commit `5928f05` (tag `arsip-docs-v1`). File planning lain: `stack.md`, `design.md`, `database.md`, `security.md`.

## Keputusan khusus prototype - 10 November 2026

Ketentuan berikut berlaku hanya untuk simulasi `prototype/`, mendahului catatan prototype lama yang bertentangan pada bagian selanjutnya. Kebutuhan dan implementasi produksi tidak direvisi pada pekerjaan ini.

- Menu utama: Beranda, Layanan, Contoh Aplikasi, Kabar, Tim, Tentang Kami. Kabar berisi Berita/Galeri Dokumentasi; tidak ada menu Profil.
- Status navigasi mengikuti halaman dan section yang memiliki pasangan menu langsung (NAV-02 B). Section lainnya mempertahankan pasangan terakhir; scroll kembali ke atas memulihkan status halaman. Navbar harus kembali lebar walaupun fokus/dropdown masih berada di navbar.
- Halaman Tim tidak menyediakan pencarian. Enam anggota dan urutan existing tetap; kartu mengutamakan bidang penelitian dengan portrait ringkas.
- Detail berita menampilkan tanggal diikuti "oleh Tim Media Ubaya AI Center". Detail dokumentasi menampilkan seluruh foto setara, utuh, tanpa foto headline.
- Penyelarasan radius section publik, CTA/footer, logo/alamat footer, serta hero neural mengikuti keputusan pada `design.md` bagian 1. Tidak memerlukan layanan atau dependency baru.


## 1. Tujuan dan batas produk

**Website ini buat apa?** Website ini memiliki 2 sisi. Sisi utama (publik) untuk menampilkan konten kepada calon customer sekaligus menjadi wadah untuk mempromosikan atau branding, perusahaan kami yaitu AI Center Universitas Surabaya (UBAYA), dimana kami memiliki fokus penyediaan layanan berbasis AI ('Pelatihan AI' dan 'Inference and Solutions'). Sisi kedua (internal) digunakan oleh tim internal perusahaan untuk input/edit data perubahan konten yang ditampilkan di sisi utama (publik). Fokus website: pengelolaan konten publlik

**Siapa pemilik dan siapa pemakai?** Pemiliknya adalah UBAYA AI Center. Penggunanya adalah Tim/staff UBAYA AI Center (yang bisa mengedit konten), user publik (hanya read only)

**Masalah apa yang diselesaikan?** Content Management System untuk keperluan UBAYA AI Center, konten yang akan di tampilkan di dalam website mencakup: detail layanan pelatihan, detail layanan solution inference, detail berita (memiliki object image), detail dokumentasi (memiliki object image), display susunan tim/staff (memiliki object image), section portofolio (berupa hardcode beberapa sampel produk sistem yg dibuat), dan terdapat kontak section (button mengarah ke WA ataupun email). Bahasa dibuat soft selling: menjelaskan manfaat, bukan menekan. Kalimat singkat, ramah, dan mudah dipahami. Tampilan menyesuaikan arah desain, tetap informatif tanpa membuat user kewalahan/overwhelm ketika mengakses website.

**Apa yang bukan produk / jasa yang ditawarkan ai center?** AI Center menawarkan adalah Layanan 'Pelatihan' dan AI 'Inference and Solutions'). Website ini bukan e-commerce, bukan LMS, bukan media sosial atau forum, bukan marketplace, bukan aplikasi SaaS, dan bukan sistem transaksi atau data pribadi skala besar.

**Target user?** Sisi website yang tampil di publik ditujukan untuk publik. Sisi webside untuk admin ditujukan untuk internal.

## 2. Scope

**Apa yang termasuk scope?**

### Scope User

Sistem ini akan memiliki 2 sisi, yaitu publik dan internal. Dari sisi publik, sistem ini akan diakses oleh pengunjung anonim dan semua orang bisa baca. Dari sisi internal, staff internal bisa melakukan akses ke dashboard admin yang berfungsi untuk mengelola CRUD Dokumentasi, Berita, Tim, Testimoni, yang bisa juga melakukan unggah gambar (jpg/png/webp maks 2MB). Konten Pelatihan (Program Reguler, Program Khusus, Seminar / Workshop untuk Bisnis) hardcode di frontend pada tahap ini, hapus/logout akan selalu memakai konfirmasi. \[Direncanakan: Retract juga memakai konfirmasi, dan form yang isinya berubah menampilkan peringatan sebelum ditinggalkan; keduanya belum ada di prototype.\] Kemudian, untuk pengelolaan login user akan dikelola oleh 1 tools sistem bernama keycloak. Dan memiliki 2 role yaitu direktur dan pegawai. \[menunggu konfirmasi Keycloak\]

Validasi upload: \[Lihat security.md §7\] dan \[Lihat database.md §3\].

## 3. Rute publik

Website dari sisi user akan memiliki navbar dan footer. Kemudian akan berisi:

Menu publik “Profil” mengelompokkan Tim dan Tentang Kami sebagai dua tautan anak; kedua halaman dan tujuannya tetap. Dropdown tersedia pada desktop dan panel navigasi mobile, dengan padanan native `details` pada fallback tanpa JavaScript. Pengelompokan ini tidak menambah halaman Profil baru.

Navbar publik menambahkan “Contoh Aplikasi” setelah menu Layanan, menuju bagian Portofolio di Inference and Solutions (prototype `inference.html#portofolio`, produksi `/layanan/inference-solutions#portofolio`), bukan halaman baru. Seluruh menu, dropdown Layanan, serta tujuan yang ada dipertahankan. CTA “Diskusikan Kebutuhan AI” membuka WhatsApp resmi. Penanganan anchor mendukung klik lintas halaman, klik pada halaman yang sama, refresh langsung, dan Back/Forward dengan offset header. Status Contoh Aplikasi aktif pada hash Portofolio tanpa penekanan utama Layanan bersamaan. HTML publik menyediakan fallback navigasi tanpa JavaScript; Inference menyediakan target Portofolio dan tautan demo statis. Detail visual serta perilaku responsive navbar: [Lihat design.md §1].

* fitur Beranda /beranda,  
* Berita /berita \+ /berita/:slug: daftar \+ pencarian judul (server-side lewat parameter `?q=`, minimal 2 karakter, dengan pagination), detail dengan gambar.  
* Dokumentasi /dokumentasi \+ /dokumentasi/:slug. HTML detail disajikan Go dengan meta SEO dan data awal \[Lihat stack.md §3\]; data juga tersedia lewat `GET /api/dokumentasi/:slug`.  
* Tim /tim: susunan anggota \+ pencarian.  
* Tentang Kami /tentang-kami  
* Fitur konten yang akan berisi dokumentasi dan berita (masing-masing punya routing page sendiri)  
* Layanan: hanya dua rute tetap, /layanan/pelatihan dan /layanan/inference-solutions (tidak memakai /layanan/:slug). Isinya hardcode di frontend, tidak ada opsi menambah layanan, dan tidak ada endpoint backend untuk layanan. Pengecualian: demo statis /layanan/inference-solutions/jejak-alumni, /layanan/inference-solutions/perencanaan-implan-gigi, dan /layanan/inference-solutions/document-intelligence \[Lihat §4.3\].  
* List kursus /layanan/pelatihan: dibuatkan kartu untuk display pilihan pelatihan \[Ditunda: aktif setelah Pelatihan Reguler tidak lagi coming soon\]  
* List layanan dalam 'Inference and Solutions': dibuatkan kartu untuk display layanan konsultasi. Menampilkan dua kategori layanan utama yang dapat menjadi titik awal bagi pengunjung untuk mengenali kebutuhan konsultasi perusahaan. Dua kategori tersebut adalah IT Service Management dan Data Science  
* Detail kursus /layanan/pelatihan/:kode: fakta, hasil belajar, materi, instruktur. \[Ditunda: aktif setelah Pelatihan Reguler tidak lagi coming soon\]  
* Toggle bahasa ID/EN dihilangkan, jadi sistem hanya menggunakan bahasa indo, peta lokasi footer, halaman 404\.  
* List portofolio untuk menampilkan 3 contoh penerapan AI, dengan status tiap entri ditampilkan apa adanya \[Lihat §4.3\]  
* Fitur umum dalam web termasuk: portofolio, list berita, list dokumentasi, paket layanan: Pelatihan dan 'Inference and Solutions', list tim AI Center  
* Kemudian tiap masuk ke dalam page lain selain beranda, akan menampilkan breadcrumb misalnya Beranda \> Pelatihan. Sehingga menampilkan lokasi yang seharusnya

## 4. Alur halaman publik

Referensi visual dan animasi tiap halaman: \[Lihat design.md §1 dan §2\].

### 4.1 Landing Page (Home)

Alur story telling pada page home adalah: hero, section jenis layanan, section dokumentasi, section testimoni klien (carousel), section berita, section kontak (tombol redirect ke wa ataupun email), footer yang berisi kontak (nomor whatsapp, email, akun instagram, alamat resmi, dan peta lokasi).

Section jenis layanan memuat dua card, yaitu 'Pelatihan' dan ''Inference and Solutions''. Tiap card terdiri dari: judul layanan, deskripsi singkat layanan, button detail layanan untuk mengakses page terkait.

Section dokumentasi akan menampilkan 3 unggahan dokumentasi terbaru (dari halaman detail dokumentasi), lalu ada button untuk menuju page daftar dokumentasi. Tiap preview item dokumentasi akan muncul dalam bentuk card yang berisi (foto headline, judul, tanggal publikasi dokumentasi terkait), lalu ada button untuk melihat detail dokumentasi yang dipilih.

Section testimoni klien (carousel otomatis, gerak per 5 detik; sebelumnya 10 detik, direvisi di frontend-final 13 November 2026; carousel pertanyaan Inference tetap 6 detik). Per carousel akan tampilkan 1 testimoni. Autoplay hanya aktif di perangkat dengan hover dan berhenti saat disorot atau difokus; di layar sentuh dan saat `prefers-reduced-motion`, carousel digerakkan manual lewat panah. Tidak ada tombol jeda \[Lihat design.md §1\].

Section berita menampilkan 3 unggahan berita terbaru (dari halaman detail berita), lalu ada button untuk menuju page daftar berita. Tiap preview item berita akan muncul dalam bentuk card yang berisi (foto headline, judul, potongan berita dari paragraf pertama artikel terkait), lalu ada button untuk melihat detail berita yang dipilih.

Section footer yang berisi kontak berupa nomor whatsapp, email, akun instagram, alamat resmi, dan peta lokasi.

### 4.2 Page Pelatihan

Halaman layanan Pelatihan UBAYA AI Center bertujuan memulai konsultasi. Alurnya: hero teks, pemilih jalur berdasarkan tujuan, Pelatihan Khusus, Workshop & Seminar untuk Bisnis, teaser Pelatihan Reguler, lalu penutup konsultasi. Tombol utama hero dan penutup berlabel tepat “Konsultasi Gratis” menuju WhatsApp resmi; “Lihat Pilihan Pelatihan” menggulir ke pemilih jalur. Halaman tidak menyediakan pendaftaran atau transaksi.

Pelatihan Khusus melayani individu, kelompok, dan organisasi (B2B/B2C), dengan tujuan pengembangan pemahaman serta kompetensi AI melalui materi dan pendekatan yang disesuaikan. Format dan cakupan ditentukan saat konsultasi; tidak menjanjikan kelas privat satu peserta untuk setiap permintaan. Empat langkah konsultasi yang dikonfirmasi pemilik dipertahankan. Workshop & Seminar untuk Bisnis melayani perusahaan (B2B), dengan tujuan membangun pemahaman bersama dan mengeksplorasi peluang, manfaat, keterbatasan, serta tantangan AI dalam proses bisnis. CTA-nya “Tanya Workshop Bisnis”. Pembeda kedua layanan adalah hasil yang diinginkan, bukan jumlah peserta, durasi, atau ada/tidaknya latihan praktik.

Pemilih jalur bersifat opsional dan memakai radio native, tanpa akun, data pribadi, backend, atau model AI. Skenario 15 karyawan belajar penerapan AI dan individu mencari kelas mengarah ke Pelatihan Khusus; skenario manajemen memahami peluang/risiko dan tim operasional mengeksplorasi ide AI mengarah ke Workshop & Seminar Bisnis. Pilihan langsung memperbarui rekomendasi, alasan, serta tautan ke `#khusus` atau `#workshop`, tanpa pemuatan halaman atau perpindahan fokus otomatis. Tidak ada pilihan awal. Kebutuhan yang tumpang tindih dapat dibahas melalui konsultasi.

Pelatihan Reguler tetap terlihat secara sekunder dengan status “Coming Soon”. Jadwal, harga, instruktur, kurikulum, tanggal peluncuran, dan pendaftaran tidak ditambahkan sebelum informasi resminya tersedia.

### 4.3 Page 'Inference and Solutions' \-\> Inference & Solution

Alur story telling page 'Inference and Solutions' adalah: Hero (memuat carousel pertanyaan untuk menarik minat user untuk berkonsultasi dengan AI Center; arahnya agar user mengetahui urgensi mereka untuk konsultasi solution inference dengan AI Center), jenis solution inference di AI Center, alur solution inference, 3 mock up portofolio (hardcode), dan section kontak kami.

Portofolio menampilkan tiga kartu hardcode: JejakAlumni, Perencanaan Implan Gigi Berbasis AI, dan DocNext. DocNext menggantikan kartu Algae Finder; album dokumentasi Algae Finder tetap ada. Setiap kartu membuka demo masing-masing. Copy memakai istilah "contoh penerapan". Ketiga demo adalah demonstrasi interaktif, bukan Proof of Concept yang tervalidasi.

**Demo JejakAlumni** adalah contoh penerapan untuk tracer study sekolah, berstatus "Demo Konsep · Data Simulasi". Demo ini hardcode di frontend: tanpa backend, database, data alumni sebenarnya, maupun inferensi AI langsung. Rute prototype `public/jejak-alumni.html`; rute produksi direncanakan /layanan/inference-solutions/jejak-alumni. Halaman memakai `noindex`.

* Isi: filter angkatan (2025, 2024) dan jurusan, empat KPI, grafik penyerapan per jurusan, status alumni, tema jawaban terbuka, ringkasan untuk rapat kurikulum, dan penjelasan aplikasi. Tab Ringkasan berfungsi; Responden dan Mitra kerja nonaktif dengan teks "Belum tersedia dalam demo".
* Definisi indikator: Terserap dunia kerja adalah responden yang memperoleh pekerjaan dalam 6 bulan setelah lulus, tidak termasuk wirausaha. Bekerja adalah status saat mengisi kuesioner. Kerja sesuai bidang adalah bagian dari responden berstatus bekerja. Masa tunggu adalah median bulan hingga pekerjaan pertama pada responden yang terserap.
* Statistik tidak disebut hasil AI. Hanya tema dan ringkasan yang berlabel simulasi AI, dan keduanya berlaku per angkatan untuk seluruh jurusan, tidak mengikuti filter jurusan.
* Pembanding hanya tampil bila data angkatan sebelumnya ada; angkatan 2024 tidak punya pembanding.
* Data simulasi: jumlah orang per jurusan adalah sumber kebenaran, sedangkan persentase dan total dihitung saat tampil agar tidak saling bertentangan. Jumlah per jurusan, median per jurusan, tema 2024, dan ringkasan 2024 adalah tambahan simulasi, bukan data sekolah. Median per jurusan tidak dapat dibuktikan secara matematis. Filter tidak disimpan di URL.

**Demo Perencanaan Implan Gigi Berbasis AI** adalah demo penelitian, berstatus "Demo Penelitian · Visualisasi Konseptual · Nonklinis". Demo ini hardcode di frontend: tanpa backend, unggah CBCT, data pasien, inferensi AI, maupun rekomendasi ukuran. Rute prototype `public/perencanaan-implan-gigi.html`; rute produksi direncanakan /layanan/inference-solutions/perencanaan-implan-gigi. Halaman memakai `noindex`.

* Isi: konteks, empat tahap terpandu (Identifikasi Area Edentulous, Segmentasi Anatomi Rahang, Rekomendasi Dimensi Implan, Visualisasi 3D), cara kerja AI, hasil penelitian, batasan, dan kontak.
* Semua gambar adalah ilustrasi konseptual buatan sendiri. Kontrol hanya mengganti tampilan yang sudah disiapkan; tidak ada proses atau loading tiruan.
* Dimensi implan ditampilkan sebagai "Ilustratif / Tidak Ditentukan". Tidak ada angka ukuran implan untuk pasien tertentu.
* Metrik ditampilkan sebagai hasil evaluasi yang dilaporkan dalam penelitian yang menjadi dasar demonstrasi ini, selalu bersama tugas yang diukur, dan terpisah dari ilustrasi. Klaim kesetaraan dengan dokter gigi, kesiapan klinis, atau keamanan ukuran tidak boleh dipakai.
* Disclaimer nonklinis wajib terlihat sebelum area demo.
* Untuk saat ini halaman tidak mencantumkan nama penulis, judul publikasi, maupun tahun terbit (keputusan pemilik). Asal dataset tidak disebut di halaman publik. Copy tidak boleh mengklaim UBAYA AI Center memiliki dataset, mengembangkan model, melakukan studi, atau bermitra secara institusional tanpa verifikasi, dan tidak boleh mengesankan demo memproses dataset penelitian atau data pasien. Bila publikasi atau lisensi data mewajibkan kredit publik, tambahkan bagian referensi penelitian.
* Penampang melintang yang bergerak sepanjang lengkung rahang pada Tahap 1 dan 2 adalah penalaran perancang, bukan fakta dari abstrak penelitian.
* \[Ditandai\] Tinjauan ahli kedokteran gigi untuk ilustrasi, terutama arah slice, posisi kanal mandibula, bentuk tulang alveolar, dan penempatan implan.

**Demo DocNext** (Document Intelligence untuk Pemeriksaan Dokumen Pengadaan) ditujukan bagi calon klien korporat, berstatus "Demo Konsep · Dokumen Sintetis". Demo ini hardcode di frontend: tanpa backend, OCR, inferensi AI, integrasi ERP, unggah dokumen, login, maupun pengumpulan data pengunjung. Rute prototype `public/document-intelligence.html`; rute produksi direncanakan /layanan/inference-solutions/document-intelligence. Halaman memakai `noindex`.

* Isi: masalah bisnis, ruang kerja dokumen, peran AI dan aturan bisnis, penerapan lain, opsi deployment, batasan, dan kontak.
* Ada tiga skenario sintetis: Dokumen Sesuai (Sesuai), Selisih Kuantitas (Perlu Ditinjau), dan Deskripsi Barang Berbeda (Perlu Verifikasi). Tiap skenario memuat Purchase Order, dokumen penerimaan, dan invoice.
* Memilih hasil ekstraksi menyorot sumbernya pada dokumen. Tiap temuan menyebut dokumen, field, nilai, alasan, dan hal yang perlu diperiksa peninjau.
* Hasil ekstraksi dan keterkaitan deskripsi disiapkan sebelumnya dan berlabel "Simulasi AI". Perbandingan angka dan referensi dihitung di browser dan berlabel "Aturan Bisnis". Keputusan tetap di tangan peninjau.
* Tidak boleh ada klaim penipuan, persetujuan pembayaran otomatis, persentase keyakinan, akurasi, penghematan, atau ROI. Kemiripan deskripsi bukan bukti barang identik.
* Opsi on-premises dan private cloud disebut sebagai potensi yang "dapat dirancang", bergantung pada asesmen teknis. Tidak boleh ada jaminan keamanan absolut atau klaim bahwa demo publik memakai inferensi lokal.
* Penerapan pada kontrak, formulir, laporan, dan dokumen operasional disebut sebagai arah pengembangan, bukan kemampuan demo.
* Nilai transaksi, nama barang, dan nama pihak adalah buatan perancang. Pada Skenario 3 kode barang di invoice sengaja berbeda dari Purchase Order agar keraguan identitas barang nyata.
* Nama "DocNext" adalah judul demo. Dokumen, antarmuka, dan copy dibuat sendiri, tanpa menyalin atau mengesankan afiliasi dengan vendor lain. Pengecekan web awal bukan clearance merek. \[terbuka\]

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

**Aturan umum admin** (berlaku untuk §5.1 sampai §5.4):

* Edit, Hapus, dan Post hanya muncul untuk item berstatus draft. Item posted hanya punya Retract.  
* Item yang sedang dikunci pengguna lain: Hapus dan Post nonaktif, sedangkan Edit berganti menjadi "Lihat" (mode baca saja) \[Lihat database.md §4\].  
* Tabel Trash tidak memiliki kolom Status, karena Restore selalu mengembalikan item ke draft.  
* Tanggal publikasi (Berita dan Dokumentasi): input teks `DD-MM-YYYY` dengan kalender (bahasa Indonesia, minggu mulai Senin). Form tambah mengisinya dengan hari ini. Tanggal setelah hari ini ditolak, sehingga tidak ada posting terjadwal \[Lihat database.md §2\]. Testimoni dan Tim tidak memiliki tanggal publikasi.  
* Tidak ada dashboard, pencarian, atau pagination di daftar admin, dan tanggal masuk trash tidak ditampilkan \[terbuka\].  
* \[Terbuka, perlu didiskusikan sebelum backend karena berdampak ke API\] Konten posted harus di-Retract sebelum diedit, sehingga memperbaiki satu typo membuat halaman publiknya 404 sementara, termasuk bagi tautan yang sudah dibagikan.  
* Konfirmasi Retract dan peringatan perubahan belum tersimpan: \[Direncanakan\] \[Lihat §2\].

### 5.1 Bagian Berita

ada section yang mendisplay semua berita yang ada di database sistem (status draft & posted), section untuk edit berita (memuat image highlight, judul, tanggal publikasi, konten berita), section untuk Trash Berita (apapun data berita yang di page ini akan dihapus dalam waktu 30 hari setelah dipindahkan ke trash ini). Saat status berita adalah 'draft', button pada tiap item berita (pada section display semua berita) ada button untuk 'edit', 'hapus', 'post';

Saat status 'posted', button pada tiap item berita (pada display) ada button 'retract' untuk mengembalikan status posted ke draft.

Bila item berita pada list dihapus, maka item tersebut akan pindah ke section 'trash berita' sebagai mekanisme soft delete. Semua item yang berada di sini lebih dari 30 hari (setelah dipindahkan) akan terhapus secara permanen.

Di section trash tersedia tombol 'Restore' (kembali ke draft) dan 'Hapus Permanen' (dengan konfirmasi).

### 5.2 Bagian Dokumentasi

ada section yang mendisplay semua dokumentasi yang ada di database sistem (status draft & posted), section untuk edit dokumentasi (upload multi image maksimum jumlah \= 6, judul kegiatan, tanggal publikasi), section untuk Trash Dokumentasi (apapun data dokumentasi yang di page ini akan dihapus dalam waktu 30 hari setelah dipindahkan ke trash ini).

Saat status dokumentasi adalah 'draft', button pada tiap item dokumentasi (pada section display semua dokumentasi) ada button untuk 'edit', 'hapus', 'post'; Saat status 'posted', button pada tiap item dokumentasi (pada display) ada button 'retract' untuk mengembalikan status posted ke draft.

Bila item dokumentasi pada list dihapus, maka item tersebut akan pindah ke section 'trash dokumentasi' sebagai mekanisme soft delete. Semua item yang berada di sini lebih dari 30 hari (setelah dipindahkan) akan terhapus secara permanen.

Di section trash tersedia tombol 'Restore' (kembali ke draft) dan 'Hapus Permanen' (dengan konfirmasi).

### 5.3 Bagian Testimoini

ada section yang mendisplay semua testimoni yang ada di database sistem (status draft & posted), section untuk edit testimoni, section untuk Trash Testimoni (apapun data testimoni yang di page ini akan dihapus dalam waktu 30 hari setelah dipindahkan ke trash ini).

Untuk testimoni itu field inputnya: Foto (opsional; tanpa foto tampil inisial nama), Nama pemberi testimoni, Teks testimoni, Posisi (jabatan atau mungkin posisi lainnya) pemberi testimoni

Saat status testimoni adalah 'draft', button pada tiap testimoni (pada section display semua testimoni) ada button untuk 'edit', 'hapus', 'post'; Saat status 'posted', button pada tiap testimoni (pada display) ada button 'retract' untuk mengembalikan status posted ke draft.

Bila item testimoni pada list dihapus, maka item tersebut akan pindah ke section 'trash testimoni' sebagai mekanisme soft delete. Semua item yang berada di sini lebih dari 30 hari (setelah dipindahkan) akan terhapus secara permanen.

Di section trash tersedia tombol 'Restore' (kembali ke draft) dan 'Hapus Permanen' (dengan konfirmasi).

Tampilan carousel testimoni: \[Lihat §4.1\] dan \[Lihat design.md §2\].

### 5.4 Bagian Tim

ada section yang mendisplay semua data tim dan section untuk edit tim.

Untuk data tim itu field inputnya: Foto, posisi / jabatan, nama, bidang riset, & urutan (angka bulat mulai dari 1; angka kecil tampil lebih dulu). Tabel Tim tetap memakai kolom Foto.

## 6. Pengguna, peran, izin

Pembaruan prototype 10 November 2026: konfirmasi Retract kini diimplementasikan sebelum pemanggilan API simulasi; membatalkan dialog tidak mengubah status konten. Ini menyelesaikan rencana konfirmasi Retract yang disebut pada bagian sebelumnya, tanpa mengubah aturan izin. Navbar publik mengelompokkan Berita/Galeri Dokumentasi di bawah Aktivitas; kedua rute tetap. Pencarian dan pagination publik memulihkan state URL `q`/`page` dan memberi status hasil serta pemulihan fokus.

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

**Catatan status prototype** (hal yang belum atau baru sebagian; cara membuka dan simulasi: \[Lihat stack.md §10\]):

* List dan detail kursus belum dibuat \[Ditunda\].
* HTML awal berisi title, OG, dan JSON-LD tidak bisa dibuat di HTML statis; dikerjakan Go \[Lihat stack.md §3, RP-12\].
* Tentang Kami: isi diambil dari data situs lama, dan dokumen ini belum merincinya.
* Inference: lima tahap alur sudah dikonfirmasi pemilik, tetapi deskripsi dua jenis layanan (IT Service Management dan Data Science) masih data contoh \[terbuka\].
* Pelatihan Khusus: empat langkah alur sudah dikonfirmasi pemilik. Belum ada sumber tertulis untuk alur konsultasi.
* Tim: enam bidang riset ditetapkan pemilik untuk prototype dan perlu diverifikasi sebagai keahlian pribadi sebelum produksi \[terbuka\].
* Handle Instagram di footer belum terverifikasi \[terbuka\].

Dokumen kontrak API: \[Lihat stack.md §9\] memuat endpoint, skema JSON, format error, contoh mock, dan environment variable.

Alur fase kerja: \[Lihat stack.md §1\].
