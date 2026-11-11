/* DATA DUMMY. Dibuang saat implementasi. Kolom mengikuti docs/database.md (snake_case).
   Tanggal publikasi: string DD-MM-YYYY. `gambar`, `foto`, `key` = nama file di assets/img. */
const MOCK = (() => {
  const audit = (i) => ({ created_at: `2026-10-${String((i % 27) + 1).padStart(2, '0')}T02:00:00Z`, created_by: 'sub-direktur', updated_by: 'sub-direktur' });
  const fotoContoh = ['ruang-pelatihan.jpg', 'ruang-diskusi.jpg', 'ruang-demo.jpg', 'algae-finder.jpg', 'cacat-las.jpg', 'gesture-translator.jpg', 'xray-pneumonia.jpg', 'kesegaran-ikan.jpg'];
  const catatanContoh = 'Teks ini adalah data contoh untuk meninjau tampilan prototype dan bukan berita sungguhan.';

  // [judul, tanggal, status, dihapus]
  const daftarBerita = [
    ['Pelatihan pengenalan machine learning untuk guru', '28-10-2026', 'posted'],
    ['Diskusi terbuka pemanfaatan AI di sekolah', '21-10-2026', 'posted'],
    ['Lokakarya menyusun prompt untuk dosen', '14-10-2026', 'posted'],
    ['Kunjungan mitra industri ke ruang demo AI', '02-10-2026', 'posted'],
    ['Sesi berbagi riset computer vision', '24-09-2026', 'posted'],
    ['Bootcamp data science bagi mahasiswa', '17-09-2026', 'posted'],
    ['Pendampingan deployment model untuk mitra', '05-09-2026', 'posted'],
    ['Seminar etika kecerdasan artifisial', '27-08-2026', 'posted'],
    ['Pelatihan analisis data untuk tenaga kesehatan', '12-08-2026', 'posted'],
    ['Peluncuran materi belajar AI untuk umum', '30-07-2026', 'posted'],
    ['Kolaborasi riset bersama fakultas', '16-07-2026', 'posted'],
    ['Workshop computer vision tingkat dasar', '01-07-2026', 'posted'],
    ['Klinik konsultasi AI untuk UMKM', '18-06-2026', 'posted'],
    ['Pameran karya mahasiswa bidang AI', '03-06-2026', 'posted'],
    ['Rencana kegiatan semester depan', '30-10-2026', 'draft'],
    ['Agenda seminar AI untuk bisnis', '05-11-2026', 'draft'],
    ['Pengumuman kegiatan yang dibatalkan', '11-05-2026', 'posted', '2026-10-30T03:00:00Z'],
    ['Berita ganda yang akan dihapus', '20-04-2026', 'draft', '2026-10-25T03:00:00Z'],
  ];

  const berita = daftarBerita.map(([judul, tanggal_publikasi, status, deleted_at], i) => ({
    id: i + 2,
    slug: judul.toLowerCase().replace(/[^a-z0-9]+/g, '-'),
    judul,
    isi: `${judul} menjadi salah satu kegiatan yang disiapkan UBAYA AI Center untuk membantu peserta memahami manfaat kecerdasan artifisial dalam pekerjaan sehari-hari.\n\nPeserta mempelajari konsep dasar, mencoba latihan singkat, lalu berdiskusi bersama tim tentang penerapan yang paling sesuai dengan kebutuhan mereka.\n\n${catatanContoh}`,
    tanggal_publikasi,
    gambar: fotoContoh[i % fotoContoh.length],
    status,
    version: 1,
    deleted_at: deleted_at ?? null,
    ...audit(i),
  }));
  // Satu-satunya berita asli dari src/data/berita.ts
  berita.unshift({
    id: 1,
    slug: 'ai-center-dukung-layanan-kesehatan-jatim',
    judul: 'UBAYA AI Center Siap Dukung Layanan Kesehatan di Jawa Timur',
    isi: 'Pengembangan AI Center dibahas dalam audiensi Rektor Ubaya, Benny Lianto, dengan Gubernur Jawa Timur, Khofifah Indar Parawansa, di Gedung Negara Grahadi, Surabaya, Jumat (22/5/2026). Gubernur Khofifah menyambut positif langkah Ubaya mengembangkan teknologi AI untuk mendukung sektor kesehatan dan peningkatan kualitas sumber daya manusia.\n\nRektor Benny Lianto menyebut AI Center diproyeksikan menjadi pusat riset unggulan di Jawa Timur yang menghasilkan inovasi berbasis AI yang aplikatif, antara lain teknologi implan gigi dan aplikasi AI untuk membantu diagnosis penyakit.\n\nPengembangan AI kesehatan di Ubaya didukung ekosistem yang lengkap, mulai dari Fakultas Kedokteran, rumah sakit milik kampus, hingga kolaborasi dengan institusi kesehatan lain termasuk RSUD Dr. Soetomo. Artikel ini disadur dari siaran pers Dinas Komunikasi dan Informatika Provinsi Jawa Timur (23 Mei 2026).',
    tanggal_publikasi: '22-05-2026',
    gambar: 'ruang-diskusi.jpg',
    status: 'posted',
    version: 1,
    deleted_at: null,
    ...audit(20),
  });

  const karya = [
    ['Rekomendasi Implan Gigi Otomatis', '20-08-2026', 'implan-gigi-konseptual.svg'],
    ['Algae Finder', '28-07-2026', 'algae-finder.jpg'],
    ['Translator Bahasa Isyarat', '02-07-2026', 'gesture-translator.jpg'],
    ['Deteksi Cacat Las', '28-05-2026', 'cacat-las.jpg'],
    ['Klasifikasi X-Ray Pneumonia', '15-09-2026', 'xray-pneumonia.jpg'],
    ['Deteksi Kesegaran Ikan', '15-04-2026', 'kesegaran-ikan.jpg'],
  ];
  const kegiatan = [
    ['Dokumentasi pelatihan AI untuk guru', '29-10-2026', 'posted'],
    ['Dokumentasi diskusi bersama mitra industri', '22-10-2026', 'posted'],
    ['Dokumentasi lokakarya prompt', '13-10-2026', 'posted'],
    ['Dokumentasi kunjungan ruang demo', '01-10-2026', 'posted'],
    ['Dokumentasi bootcamp data science', '18-09-2026', 'posted'],
    ['Dokumentasi seminar etika AI', '26-08-2026', 'posted'],
    ['Dokumentasi klinik konsultasi UMKM', '19-06-2026', 'posted'],
    ['Dokumentasi pameran karya mahasiswa', '04-06-2026', 'posted'],
    ['Dokumentasi rapat persiapan pelatihan', '03-05-2026', 'posted'],
    ['Dokumentasi kegiatan bulan depan', '31-10-2026', 'draft'],
    ['Dokumentasi yang masih disusun', '03-11-2026', 'draft'],
    ['Dokumentasi unggahan ganda', '09-03-2026', 'draft', '2026-10-28T03:00:00Z'],
    ['Dokumentasi kegiatan yang dibatalkan', '11-02-2026', 'posted', '2026-10-20T03:00:00Z'],
  ];
  const semuaDok = [...karya.map(([j, t, g]) => [j, t, 'posted', null, g]), ...kegiatan];
  const dokumentasi = semuaDok.map(([judul, tanggal_publikasi, status, deleted_at], i) => ({
    id: i + 1,
    slug: judul.toLowerCase().replace(/[^a-z0-9]+/g, '-'),
    judul,
    tanggal_publikasi,
    status,
    version: 1,
    deleted_at: deleted_at ?? null,
    ...audit(i),
  }));
  let idGambar = 1;
  const dokumentasi_gambar = semuaDok.flatMap(([, , , , utama], i) => {
    const jumlah = (i % 6) + 1; // 1 sampai 6 gambar
    return Array.from({ length: jumlah }, (_, n) => ({
      id: idGambar++,
      dokumentasi_id: i + 1,
      key: n === 0 && utama ? utama : fotoContoh[(i + n) % fotoContoh.length],
      urutan: n + 1,
    }));
  });

  const testimoni = [
    { nama: 'Peserta Pelatihan ML', foto: 'testimoni/contoh-1.svg', posisi: 'Mahasiswa', teks: 'Materi pelatihan runtut dan langsung praktik. Akses GPU lab membuat eksperimen tugas akhir jauh lebih cepat.', status: 'posted' },
    { nama: 'Mitra Industri', posisi: 'Pengguna Inference Solution', teks: 'Proses deployment model didampingi sampai jalan di infrastruktur kami. Komunikasi tim responsif dan dokumentasinya jelas.', status: 'posted' },
    { nama: 'Dosen Peneliti', posisi: 'Peneliti', teks: 'Pendampingan AI Center sangat membantu riset. Tim responsif dan memberi dukungan teknis saat ada kendala.', status: 'posted' },
    { nama: 'Data Contoh Draft', foto: 'testimoni/contoh-2.svg', posisi: 'Guru', teks: 'Testimoni draft ini adalah data contoh untuk meninjau alur admin.', status: 'draft' },
    { nama: 'Data Contoh Trash', posisi: 'Karyawan', teks: 'Testimoni di trash ini adalah data contoh untuk meninjau alur pulihkan.', status: 'draft', deleted_at: '2026-10-29T03:00:00Z' },
  ].map((t, i) => ({ id: i + 1, foto: null, version: 1, deleted_at: null, ...t, ...audit(i) }));

  // Nama, posisi, dan foto dari src/data/tim.ts. Bidang riset ditetapkan pemilik untuk prototype; verifikasi sebelum produksi.
  const tim = [
    ['Dr. Mohammad Farid Naufal', 'Ketua', 'farid-naufal.jpg', 'Computer Vision'],
    ['Dr. Monica Widiasri', 'Koordinator Riset & Edukasi', 'monica-widiasri.jpg', 'Computer Vision'],
    ['Prof. Joko Siswantoro', 'Tim Riset', 'joko-siswantoro.jpg', 'Deep Learning & Natural Language Processing'],
    ['Marco Ariano Kristyanto, M.M., M.Kom.', 'Tim Hardware', 'marco-kristyanto.jpg', 'Jaringan Komputer & Infrastruktur'],
    ['Fikri Baharuddin, M.Kom.', 'Tim Software', 'fikri-baharuddin.jpg', 'Natural Language Processing'],
    ['Jabesh Nehemiah Wijaya, S.Kom., M.Kom.', 'Tim Software', 'jabesh-wijaya.jpg', 'Natural Language Processing'],
  ].map(([nama, posisi, foto, bidang_riset], i) => ({ id: i + 1, nama, posisi, bidang_riset, foto, urutan: i + 1, version: 1, ...audit(i) }));

  const akun = [
    { id: 'direktur', sub: 'sub-direktur', username: 'dian.direktur', nama: 'Dian Direktur', role: 'Direktur' },
    { id: 'pegawai', sub: 'sub-pegawai', username: 'pandu.pegawai', nama: 'Pandu Pegawai', role: 'Pegawai' },
    { id: 'tanpa-role', sub: 'sub-lain', username: 'tamu', nama: 'Tamu Tanpa Role', role: null },
  ];

  // Kunci awal oleh pengguna lain. Di implementasi ini ada di Redis, bukan tabel.
  const kunci = [
    { jenis: 'berita', id: 16, nama: 'Rani Pegawai' },
    { jenis: 'dokumentasi', id: 16, nama: 'Rani Pegawai' },
    { jenis: 'testimoni', id: 4, nama: 'Rani Pegawai' },
    { jenis: 'tim', id: 3, nama: 'Rani Pegawai' },
  ];

  // Konten hardcode. Sumber src/data kecuali yang ditandai data contoh.
  const konten = {
    kontak: {
      email: 'aicenter@unit.ubaya.ac.id',
      whatsapp: '0895-6342-22240',
      whatsapp_url: 'https://wa.me/62895634222240?text=' + encodeURIComponent('Halo Ubaya AI Center, saya ingin berdiskusi mengenai layanan/kerja sama AI.'),
      email_url: 'mailto:aicenter@unit.ubaya.ac.id?subject=' + encodeURIComponent('Kolaborasi dengan UBAYA AI Center'),
      instagram: '@aicenter_ubaya',
      instagram_url: 'https://instagram.com/aicenter_ubaya',
      alamat: ['Gedung Fakultas Teknik TA', 'Jalan Raya Kalirungkut, Tenggilis, Surabaya'],
      peta_embed: 'https://maps.google.com/maps?q=-7.321919965577327,112.76790932621508&z=16&output=embed',
      peta_url: 'https://www.google.com/maps/search/?api=1&query=-7.321919965577327,112.76790932621508',
      deskripsi_footer: 'Pusat riset dan layanan kecerdasan artifisial Universitas Surabaya.',
    },
    hero: {
      judul: 'Pusat Riset & Layanan Kecerdasan Artifisial',
      sub: 'Kami membantu individu, institusi, dan organisasi mengembangkan keterampilan serta menerapkan AI sesuai kebutuhan.',
      gambar: 'hero-1-1600.webp',
    },
    layanan: [
      { nama: 'Pelatihan', deskripsi: 'Kembangkan pengetahuan dan keterampilan AI melalui pelatihan yang dirancang untuk individu, komunitas, institusi, hingga kebutuhan tim di dunia kerja.', href: 'pelatihan.html' },
      { nama: 'Inference and Solutions', deskripsi: 'Bantu organisasi Anda menerapkan model AI ke dalam sistem atau proses kerja, dengan dukungan integrasi dan pendampingan teknis sesuai kebutuhan.', href: 'inference.html' },
    ],
    pelatihan: {
      hero_judul: 'Pelatihan AI untuk Kebutuhan Belajar dan Kerja',
      hero_sub: 'Kembangkan kompetensi AI melalui kelas yang disesuaikan dengan kebutuhan peserta atau organisasi. Untuk perusahaan, tersedia pula workshop dan seminar untuk mengeksplorasi peluang penerapan AI dalam proses bisnis.',
      reguler: 'Program pelatihan AI dengan topik dan jadwal yang telah ditentukan sedang kami persiapkan. Informasi program akan diumumkan setelah tersedia.',
      khusus_isi: [
        'Pelatihan untuk individu, kelompok, dan organisasi yang ingin mengembangkan pemahaman serta keterampilan AI. Materi dan pendekatan disesuaikan dengan tujuan, latar belakang peserta, dan konteks penerapan yang ingin dipelajari.',
      ],
      khusus_fokus: ['Memahami konsep dan kemampuan AI.', 'Mempelajari contoh penerapan AI sesuai kebutuhan.', 'Mengembangkan keterampilan melalui pendekatan pembelajaran yang disepakati.'],
      // Empat langkah berikut dikonfirmasi sebagai proses layanan oleh pemilik.
      alur_konsultasi: [
        { judul: 'Ceritakan kebutuhan', deskripsi: 'Hubungi kami dan sampaikan tujuan serta peserta pelatihan.' },
        { judul: 'Diskusi bersama tim', deskripsi: 'Kami memetakan materi, durasi, dan format yang paling sesuai.' },
        { judul: 'Terima usulan program', deskripsi: 'Anda menerima usulan program untuk ditinjau bersama.' },
        { judul: 'Pelatihan berjalan', deskripsi: 'Program dijalankan dan didampingi oleh tim UBAYA AI Center.' },
      ],
      workshop_isi: [
        'Workshop dan seminar bagi tim serta pimpinan perusahaan untuk membangun pemahaman bersama tentang AI dan mengeksplorasi penerapannya. Pembahasan manfaat, keterbatasan, dan tantangan disesuaikan dengan kebutuhan organisasi.',
      ],
      workshop_fokus: ['Memahami peluang dan keterbatasan penerapan AI dalam bisnis.', 'Mengeksplorasi contoh penggunaan AI pada proses kerja.', 'Membangun landasan diskusi untuk langkah eksplorasi berikutnya.'],
      jalur: [
        { kebutuhan: 'Perusahaan ingin 15 karyawan mempelajari penerapan AI untuk pekerjaan sehari-hari.', layanan: 'Pelatihan Khusus', href: '#khusus', tautan: 'Lihat Pelatihan Khusus', alasan: 'Tujuan utamanya adalah mengembangkan kompetensi peserta, meskipun pelatihan dipesan oleh perusahaan.' },
        { kebutuhan: 'Manajemen ingin memahami peluang dan risiko AI dalam proses bisnis.', layanan: 'Workshop & Seminar Bisnis', href: '#workshop', tautan: 'Lihat Workshop Bisnis', alasan: 'Tujuan utamanya adalah membangun pemahaman bersama dan membuka diskusi mengenai peluang penerapan AI di organisasi.' },
        { kebutuhan: 'Seorang peserta ingin mengikuti kelas AI yang sesuai dengan kebutuhan belajarnya.', layanan: 'Pelatihan Khusus', href: '#khusus', tautan: 'Lihat Pelatihan Khusus', alasan: 'Pelatihan Khusus juga melayani kebutuhan individu. Format dan cakupan pembelajaran dapat dibicarakan saat konsultasi.' },
        { kebutuhan: 'Tim operasional ingin mengikuti workshop untuk mengeksplorasi ide penggunaan AI.', layanan: 'Workshop & Seminar Bisnis', href: '#workshop', tautan: 'Lihat Workshop Bisnis', alasan: 'Fokusnya adalah mengeksplorasi peluang AI bersama tim. Jika diperlukan pembelajaran lebih mendalam, kebutuhan tersebut dapat dilanjutkan melalui Pelatihan Khusus.' },
      ],
    },
    inference: {
      hero_judul: 'Dari Model AI Menuju Penerapan Nyata.',
      hero_sub: 'Inference adalah proses menjalankan model AI untuk menghasilkan prediksi atau keluaran berdasarkan data yang diterima. UBAYA AI Center membantu Anda mengeksplorasi bagaimana hasil tersebut dapat dimanfaatkan dalam aplikasi atau sistem sesuai kebutuhan.',
      pertanyaan: [
        'Model riset atau prototipe sudah jadi, tetapi belum bisa dipakai tim lain?',
        'Butuh fitur AI di aplikasi atau website tanpa membangun infrastruktur ML sendiri?',
        'Perlu prediksi otomatis yang konsisten untuk gambar, teks, atau data?',
        'Ingin model tetap andal saat dipakai banyak pengguna?',
      ],
      // Deskripsi dua kategori adalah data contoh; nama kategori dari requirement.md.
      jenis: [
        { nama: 'IT Service Management', deskripsi: 'Diskusikan pengelolaan layanan TI dan bagian yang mungkin terbantu oleh penerapan AI.' },
        { nama: 'Data Science', deskripsi: 'Diskusikan kebutuhan pengolahan data dan kemungkinan penggunaannya untuk analisis atau prediksi.' },
      ],
      // Lima tahap berikut dikonfirmasi sebagai proses layanan oleh pemilik.
      alur: [
        { judul: 'Konsultasi & audit kebutuhan', deskripsi: 'Kami memetakan use case, data, dan model yang sudah atau perlu disiapkan bersama tim Anda.' },
        { judul: 'Persiapan model', deskripsi: 'Model dioptimalkan dan dibungkus agar siap dijalankan di server.' },
        { judul: 'Deployment sebagai API', deskripsi: 'Model dijalankan sebagai endpoint layanan yang stabil, lengkap dengan dokumentasi pemakaian.' },
        { judul: 'Integrasi & pengujian', deskripsi: 'API disambungkan ke aplikasi Anda lalu diuji pada data nyata.' },
        { judul: 'Operasional & purna implementasi', deskripsi: 'Monitoring, pembaruan model, dan dukungan teknis setelah layanan berjalan.' },
      ],
      portofolio: [
        { judul: 'JejakAlumni', bidang: 'Pendidikan', deskripsi: 'Lihat bagaimana data alumni dapat diolah menjadi informasi tentang penyerapan kerja, tren lulusan, dan ringkasan masukan alumni berbantuan AI.', gambar: 'jejak-alumni-konseptual.svg', href: 'jejak-alumni.html', status: 'Demo Konsep · Data Simulasi', cta: 'Jelajahi Demo' },
        { judul: 'Perencanaan Implan Gigi Berbasis AI', bidang: 'Riset AI Medis', deskripsi: 'Kenali pendekatan AI untuk mengidentifikasi area kehilangan gigi, memetakan struktur anatomi rahang, dan memvisualisasikan tahapan perencanaan implan.', gambar: 'implan-gigi-konseptual.svg', href: 'perencanaan-implan-gigi.html', status: 'Demo Penelitian · Visualisasi Konseptual · Nonklinis', cta: 'Jelajahi Demo' },
        { judul: 'DocNext', bidang: 'AI untuk Bisnis', deskripsi: 'Pelajari bagaimana AI dapat membantu membaca dokumen bisnis, mengekstraksi informasi penting, dan menemukan ketidaksesuaian antar dokumen untuk ditinjau lebih lanjut.', gambar: 'document-intelligence-konseptual.svg', href: 'document-intelligence.html', status: 'Demo Konsep · Dokumen Sintetis', cta: 'Jelajahi Demo' },
      ],
    },
    // Seluruh angka dan teks berikut adalah simulasi, bukan data sekolah dan bukan hasil model AI.
    // Sumber kebenaran: jumlah orang per jurusan; persentase dan total dihitung di assets/js/jejak-alumni.js.
    // Median disimpan langsung karena tidak bisa dihitung dari median per jurusan.
    jejak_alumni: {
      status: 'Demo Konsep · Data Simulasi',
      disclaimer: 'Seluruh data dan contoh hasil analisis dalam demo ini bersifat simulasi. Tidak ada pemrosesan AI secara langsung atau penggunaan data alumni sebenarnya.',
      jurusan: ['TKJ', 'TKR', 'DKV', 'AKL', 'Kuliner', 'Busana'],
      angkatan: {
        2025: {
          median_semua: 2.4,
          per_jurusan: {
            TKJ: { lulusan: 108, responden: 86, terserap: 65, bekerja: 60, kuliah: 14, wirausaha: 5, belum: 7, sesuai: 46, median: 1.8 },
            TKR: { lulusan: 92, responden: 72, terserap: 49, bekerja: 44, kuliah: 11, wirausaha: 7, belum: 10, sesuai: 34, median: 2.1 },
            DKV: { lulusan: 60, responden: 46, terserap: 28, bekerja: 25, kuliah: 9, wirausaha: 7, belum: 5, sesuai: 17, median: 2.6 },
            AKL: { lulusan: 56, responden: 43, terserap: 25, bekerja: 24, kuliah: 10, wirausaha: 2, belum: 7, sesuai: 15, median: 2.7 },
            Kuliner: { lulusan: 46, responden: 35, terserap: 19, bekerja: 16, kuliah: 5, wirausaha: 7, belum: 7, sesuai: 10, median: 3.0 },
            Busana: { lulusan: 38, responden: 30, terserap: 14, bekerja: 12, kuliah: 4, wirausaha: 6, belum: 8, sesuai: 7, median: 3.4 },
          },
          tema: [
            { jenis: 'positif', teks: 'Praktik di bengkel/lab dianggap paling berguna.', sebutan: 84 },
            { jenis: 'positif', teks: 'Pembimbing PKL responsif.', sebutan: 51 },
            { jenis: 'keluhan', teks: 'Informasi lowongan terlambat sampai.', sebutan: 63 },
            { jenis: 'keluhan', teks: 'Sertifikasi kompetensi belum terbit saat lulus.', sebutan: 47 },
          ],
          ringkasan: 'Busana (47%) dan Kuliner (54%) menunjukkan penyerapan kerja lebih rendah dibandingkan TKJ (76%). Keluhan yang paling sering muncul di seluruh jurusan adalah informasi lowongan yang terlambat sampai (63 sebutan). Temuan ini dapat menjadi bahan diskusi, tetapi belum menunjukkan hubungan sebab-akibat.',
        },
        2024: {
          median_semua: 2.9,
          per_jurusan: {
            TKJ: { lulusan: 98, responden: 70, terserap: 49, bekerja: 45, kuliah: 13, wirausaha: 4, belum: 8, sesuai: 33, median: 2.3 },
            TKR: { lulusan: 86, responden: 61, terserap: 39, bekerja: 36, kuliah: 11, wirausaha: 5, belum: 9, sesuai: 25, median: 2.6 },
            DKV: { lulusan: 62, responden: 44, terserap: 24, bekerja: 22, kuliah: 7, wirausaha: 6, belum: 9, sesuai: 14, median: 3.1 },
            AKL: { lulusan: 56, responden: 40, terserap: 21, bekerja: 18, kuliah: 10, wirausaha: 2, belum: 10, sesuai: 12, median: 3.2 },
            Kuliner: { lulusan: 40, responden: 28, terserap: 14, bekerja: 14, kuliah: 4, wirausaha: 3, belum: 7, sesuai: 7, median: 3.4 },
            Busana: { lulusan: 36, responden: 25, terserap: 11, bekerja: 10, kuliah: 3, wirausaha: 4, belum: 8, sesuai: 5, median: 3.8 },
          },
          tema: [
            { jenis: 'positif', teks: 'Praktik di bengkel/lab dianggap paling berguna.', sebutan: 66 },
            { jenis: 'positif', teks: 'Pembimbing PKL responsif.', sebutan: 38 },
            { jenis: 'keluhan', teks: 'Informasi lowongan terlambat sampai.', sebutan: 61 },
            { jenis: 'keluhan', teks: 'Sertifikasi kompetensi belum terbit saat lulus.', sebutan: 52 },
          ],
          ringkasan: 'Busana (44%) dan Kuliner (50%) menunjukkan penyerapan kerja lebih rendah dibandingkan TKJ (70%). Keluhan yang paling sering muncul di seluruh jurusan adalah informasi lowongan yang terlambat sampai (61 sebutan). Temuan ini dapat menjadi bahan diskusi, tetapi belum menunjukkan hubungan sebab-akibat.',
        },
      },
      definisi: [
        { istilah: 'Terserap dunia kerja', arti: 'Responden yang memperoleh pekerjaan dalam 6 bulan setelah lulus, tidak termasuk wirausaha.' },
        { istilah: 'Bekerja', arti: 'Responden yang berstatus bekerja saat mengisi kuesioner.' },
        { istilah: 'Kerja sesuai bidang', arti: 'Bagian dari responden berstatus bekerja yang pekerjaannya sesuai bidang jurusan.' },
        { istilah: 'Masa tunggu kerja', arti: 'Median bulan hingga pekerjaan pertama pada responden yang terserap dunia kerja.' },
      ],
      penjelasan: {
        masalah: 'Tracer study sering dikerjakan menjelang akreditasi, sementara respons alumni terbatas dan jawaban terbuka membutuhkan waktu untuk ditinjau.',
        langkah: [
          'Kuesioner tracer study dikumpulkan dari alumni secara berkala.',
          'Indikator lulusan disajikan berdasarkan angkatan dan jurusan.',
          'Jawaban terbuka dapat dibantu pengelompokannya menjadi tema dan ringkasan evaluasi.',
        ],
        langkah_catatan: 'Ini gambaran konsep. Demo ini tidak menyebarkan kuesioner atau mengumpulkan jawaban.',
        data: ['Data lulusan', 'Hasil kuesioner tracer', 'Data mitra industri'],
        data_catatan: 'Demo hanya menampilkan data lulusan dan kuesioner simulasi. Data mitra industri belum ditampilkan.',
        hasil: 'Informasi mengenai kondisi lulusan menjadi lebih mudah ditinjau untuk kebutuhan evaluasi, pelaporan, dan diskusi pengembangan pembelajaran.',
      },
    },
    // Dokumen dan pihak sepenuhnya sintetis. Ekstraksi dan keterkaitan deskripsi disiapkan, tidak ada model AI.
    // Perbandingan angka dihitung sungguhan di assets/js/document-intelligence.js.
    document_intelligence: {
      judul: 'DocNext',
      subjudul: 'Document Intelligence untuk Pemeriksaan Dokumen Pengadaan',
      tagline: 'Dari Dokumen Menjadi Informasi yang Dapat Ditindaklanjuti.',
      deskripsi: 'Demonstrasi interaktif ini memperlihatkan contoh alur pemeriksaan Purchase Order, dokumen penerimaan barang, dan invoice menggunakan dokumen sintetis dan hasil analisis simulasi.',
      status: 'Demo Konsep · Dokumen Sintetis',
      disclaimer: 'Demo ini merupakan contoh konsep dengan dokumen sintetis dan hasil analisis simulasi. Tidak ada pemrosesan AI secara langsung, integrasi ERP, atau penggunaan dokumen perusahaan sebenarnya.',
      catatan_ekstraksi: 'Contoh hasil ekstraksi informasi yang telah disiapkan untuk demonstrasi. Tidak ada model AI yang dijalankan secara langsung.',
      masalah: [
        { judul: 'Satu transaksi, beberapa dokumen', teks: 'Satu pembelian dapat melibatkan Purchase Order, dokumen penerimaan barang, dan invoice yang dibuat oleh pihak berbeda.' },
        { judul: 'Format dan istilah berbeda', teks: 'Tata letak, istilah, dan cara menuliskan barang dapat berbeda antardokumen, walaupun membahas transaksi yang sama.' },
        { judul: 'Pemeriksaan manual', teks: 'Tim perlu membuka beberapa berkas, mencari informasi penting, membandingkannya, lalu menelusuri perbedaan yang ditemukan.' },
      ],
      label_dokumen: {
        po: { refPo: 'No. PO', pemasok: 'Kepada', kode: 'Kode Barang', deskripsi: 'Uraian', qty: 'Jml Pesan', harga: 'Harga Satuan', subtotal: 'Jumlah' },
        penerimaan: { refPo: 'Ref. Pesanan', pemasok: 'Diterima dari', kode: 'Kode Internal', deskripsi: 'Nama Barang', qty: 'Jumlah Diterima' },
        invoice: { refPo: 'Order Ref', pemasok: 'Diterbitkan oleh', kode: 'Item Code', deskripsi: 'Description', qty: 'Qty', harga: 'Unit Price', subtotal: 'Subtotal (sebelum pajak)' },
      },
      pembeli: 'PT Mitra Contoh Manufaktur',
      lokasi: 'Gudang Contoh A',
      skenario: [
        {
          label: 'Dokumen Sesuai',
          harapan: 'sesuai',
          penjelasan: 'Informasi utama pada Purchase Order, dokumen penerimaan, dan invoice sesuai berdasarkan aturan pemeriksaan yang disimulasikan.',
          catatan: 'Status ini bukan persetujuan pembayaran. Keputusan tetap mengikuti prosedur perusahaan.',
          po: { nomor: 'SIM-PO-0101', tanggal: '05-10-2026', pemasok: 'CV Pemasok Contoh Satu', kode: 'SIM-FO-220', deskripsi: 'FILTER OLI MESIN FO-220', qty: 100, satuan: 'unit', harga: 125000, subtotal: 12500000 },
          penerimaan: { nomor: 'SIM-GR-0101', tanggal: '12-10-2026', refPo: 'SIM-PO-0101', pemasok: 'CV Pemasok Contoh Satu', kode: 'SIM-FO-220', deskripsi: 'FILTER OLI MESIN FO-220', qty: 100, satuan: 'unit' },
          invoice: { nomor: 'SIM-INV-0101', tanggal: '14-10-2026', refPo: 'SIM-PO-0101', pemasok: 'CV Pemasok Contoh Satu', kode: 'SIM-FO-220', deskripsi: 'FILTER OLI MESIN FO-220', qty: 100, satuan: 'unit', harga: 125000, subtotal: 12500000 },
          temuan: [],
        },
        {
          label: 'Selisih Kuantitas',
          harapan: 'ditinjau',
          penjelasan: 'Invoice mencantumkan 100 unit, sementara dokumen penerimaan mencatat 96 unit. Selisih ini memerlukan pemeriksaan lebih lanjut.',
          catatan: 'Selisih belum tentu berarti kesalahan. Invoice tidak ditolak otomatis.',
          po: { nomor: 'SIM-PO-0102', tanggal: '06-10-2026', pemasok: 'CV Pemasok Contoh Dua', kode: 'SIM-GL-104', deskripsi: 'SARUNG TANGAN KERJA GL-104', qty: 100, satuan: 'unit', harga: 125000, subtotal: 12500000 },
          penerimaan: { nomor: 'SIM-GR-0102', tanggal: '13-10-2026', refPo: 'SIM-PO-0102', pemasok: 'CV Pemasok Contoh Dua', kode: 'SIM-GL-104', deskripsi: 'SARUNG TANGAN KERJA GL-104', qty: 96, satuan: 'unit' },
          invoice: { nomor: 'SIM-INV-0102', tanggal: '15-10-2026', refPo: 'SIM-PO-0102', pemasok: 'CV Pemasok Contoh Dua', kode: 'SIM-GL-104', deskripsi: 'SARUNG TANGAN KERJA GL-104', qty: 100, satuan: 'unit', harga: 125000, subtotal: 12500000 },
          temuan: [{
            judul: 'Selisih kuantitas antara invoice dan penerimaan',
            field: 'qty',
            dokumen: ['penerimaan', 'invoice'],
            mengapa: 'Jumlah yang ditagihkan lebih besar daripada jumlah yang tercatat diterima, sehingga perlu ditinjau sebelum transaksi diproses lebih lanjut.',
            periksa: ['Apakah pengiriman dilakukan bertahap?', 'Apakah ada dokumen penerimaan yang belum tercatat?', 'Apakah pesanan berubah setelah Purchase Order terbit?', 'Apakah ada perbedaan administrasi antardokumen?'],
          }],
        },
        {
          label: 'Deskripsi Barang Berbeda',
          harapan: 'verifikasi',
          penjelasan: 'Deskripsi barang memiliki kemiripan, tetapi kesesuaian identitas produk tetap perlu diperiksa berdasarkan SKU atau referensi barang yang berlaku.',
          catatan: 'Kemiripan deskripsi bukan bukti bahwa kedua barang identik.',
          keterkaitan: 'Kemungkinan deskripsi terkait — memerlukan verifikasi.',
          po: { nomor: 'SIM-PO-0103', tanggal: '07-10-2026', pemasok: 'CV Pemasok Contoh Tiga', kode: 'SIM-BR-6205', deskripsi: 'BEARING 6205-2RS', qty: 50, satuan: 'unit', harga: 85000, subtotal: 4250000 },
          penerimaan: { nomor: 'SIM-GR-0103', tanggal: '14-10-2026', refPo: 'SIM-PO-0103', pemasok: 'CV Pemasok Contoh Tiga', kode: 'SIM-BR-6205', deskripsi: 'BEARING 6205-2RS', qty: 50, satuan: 'unit' },
          invoice: { nomor: 'SIM-INV-0103', tanggal: '16-10-2026', refPo: 'SIM-PO-0103', pemasok: 'CV Pemasok Contoh Tiga', kode: 'VND-88213', deskripsi: 'BALL BEARING 6205 2RS', qty: 50, satuan: 'unit', harga: 85000, subtotal: 4250000 },
          temuan: [{
            judul: 'Identitas barang pada invoice belum pasti',
            field: 'deskripsi',
            dokumen: ['po', 'invoice'],
            mengapa: 'Kode dan deskripsi pada invoice berbeda dari Purchase Order. Interpretasi makna dapat menunjukkan kemungkinan keterkaitan, tetapi tidak dapat memastikan identitas barang.',
            periksa: ['Apakah kode pemasok terhubung dengan kode barang internal?', 'Apakah spesifikasi barang yang diterima sesuai pesanan?'],
          }],
        },
      ],
      lapisan: {
        pengantar: ['AI berpotensi membantu mengenali struktur dokumen, mengekstraksi informasi penting, dan memahami variasi penulisan antar dokumen.', 'Informasi tersebut kemudian dapat diperiksa menggunakan aturan bisnis untuk menemukan perbedaan jumlah, harga, atau identitas barang yang perlu ditinjau karyawan.'],
        kelompok: [
          { judul: 'Peran AI', label: 'Disimulasikan dalam demo', butir: ['Mengenali tata letak dan jenis dokumen.', 'Mengekstraksi informasi seperti kode, jumlah, dan harga.', 'Memetakan label berbeda, misalnya No. PO, Ref. Pesanan, dan Order Ref.', 'Menunjukkan kemungkinan keterkaitan deskripsi barang.'] },
          { judul: 'Aturan bisnis', label: 'Dihitung di browser', butir: ['Membandingkan kuantitas dan harga.', 'Menghitung selisih dan subtotal.', 'Memeriksa referensi Purchase Order.', 'Menandai hal yang tidak memenuhi aturan.'] },
          { judul: 'Peninjau', label: 'Tetap oleh manusia', butir: ['Memverifikasi kecocokan yang belum pasti.', 'Menilai konteks selisih operasional.', 'Mengikuti prosedur persetujuan.', 'Memutuskan apakah invoice dapat diproses.'] },
        ],
      },
      penerapan: {
        teks: 'Selain dokumen pengadaan, pendekatan Document Intelligence dapat dikembangkan untuk mendukung pengolahan kontrak, formulir, laporan, dan dokumen operasional lainnya sesuai kebutuhan perusahaan.',
        contoh: ['Kontrak', 'Formulir', 'Laporan', 'Dokumen operasional'],
        catatan: 'Demo ini hanya memperagakan dokumen pengadaan. Jenis dokumen lain memerlukan penyesuaian domain, ekstraksi khusus, konfigurasi aturan, integrasi, dan evaluasi teknis.',
      },
      deployment: {
        teks: 'Solusi AI dapat dirancang untuk berjalan pada server internal (on-premises) atau lingkungan private cloud, sehingga perusahaan memiliki kendali lebih besar atas pemrosesan dan pengelolaan data sensitif. Kelayakan implementasi disesuaikan dengan kebutuhan model, infrastruktur, integrasi, serta kebijakan keamanan perusahaan.',
        alur: ['Dokumen Internal', 'Pemrosesan AI Lokal', 'Hasil Pemeriksaan', 'Reviewer Perusahaan'],
        alur_label: 'Contoh arsitektur potensial, bukan arsitektur demo ini.',
        opsi: [
          { judul: 'On-premises', teks: 'Model dan komponen pemrosesan AI dapat dirancang untuk berjalan pada server internal perusahaan, bergantung pada kebutuhan dan kelayakan teknis.' },
          { judul: 'Private cloud', teks: 'Solusi dapat dirancang untuk lingkungan cloud privat dengan pengaturan jaringan, akses, dan pengelolaan data sesuai kebutuhan organisasi.' },
        ],
        prasyarat: ['Model, OCR, dan layanan pencocokan makna yang dapat dijalankan secara lokal.', 'Tidak ada panggilan inferensi eksternal yang wajib, termasuk pembaruan atau telemetri yang perlu diperiksa.', 'Perangkat keras yang sesuai.', 'Pembatasan jaringan, kontrol akses, dan audit trail.', 'Kebijakan log dan retensi, karena log dapat memuat isi dokumen.', 'Lisensi perangkat lunak dan model yang sesuai.', 'Tinjauan keamanan dan asesmen teknis.'],
        catatan: 'Pemasangan lokal mengurangi kebutuhan mengirim dokumen ke layanan eksternal, tetapi tidak dengan sendirinya menjamin kerahasiaan data.',
      },
      batasan: [
        'Semua dokumen, perusahaan, nomor, dan nilai transaksi bersifat sintetis.',
        'Hasil ekstraksi dan keterkaitan deskripsi telah disiapkan. Tidak ada OCR atau model AI yang berjalan.',
        'Aturan perbandingan dalam demo disederhanakan: satu dokumen penerimaan, satu baris barang, dan tanpa toleransi selisih. Pajak tidak dihitung.',
        'Tidak ada unggah dokumen, integrasi ERP, proses persetujuan, atau pengumpulan data pengunjung.',
        'Demo ini bukan Proof of Concept yang tervalidasi. Proof of Concept memerlukan model yang berjalan, dataset, kriteria evaluasi, dan hasil terukur.',
      ],
    },
    // Metrik berasal dari abstrak penelitian, bukan hasil demo. Ilustrasi konseptual, tidak ada inferensi.
    // Jangan menambah angka ukuran implan: abstrak tidak memberi nilai untuk pasien tertentu.
    implan_gigi: {
      judul: 'Perencanaan Implan Gigi Berbasis AI',
      deskripsi: 'Demonstrasi interaktif yang memperkenalkan tahapan penelitian AI berbasis citra CBCT, mulai dari identifikasi area edentulous, segmentasi anatomi rahang, hingga rekomendasi dimensi dan visualisasi 3D.',
      label: ['Demo Penelitian', 'Visualisasi Konseptual', 'Nonklinis'],
      disclaimer: 'Demo ini merupakan visualisasi konseptual berdasarkan alur penelitian. Tidak digunakan untuk diagnosis, penentuan tindakan klinis, atau perencanaan implan pada pasien nyata.',
      catatan_demo: 'Tidak ada citra pasien, pemrosesan CBCT, atau inferensi AI yang dijalankan dalam demonstrasi ini.',
      konteks: [
        { judul: 'Kondisi tulang berbeda', teks: 'Implan gigi menggantikan akar gigi yang hilang dan menjadi dasar gigi tiruan. Perencanaannya perlu menilai anatomi rahang dengan teliti karena kondisi tulang berbeda pada tiap orang.' },
        { judul: 'Peran CBCT', teks: 'Dokter gigi memakai Cone Beam Computed Tomography (CBCT) untuk melihat anatomi tulang rahang dan merencanakan dimensi implan.' },
        { judul: 'Hubungan AB dan MC', teks: 'Tulang alveolar (Alveolar Bone, AB) menopang gigi dan berkaitan dengan penempatan implan. Hubungan implan dengan kanal mandibula (Mandibular Canal, MC) perlu diperhatikan dengan cermat.' },
        { judul: 'Bagian yang dibantu AI', teks: 'Penelitian yang menjadi dasar demonstrasi ini mengkaji empat langkah pada implan molar rahang bawah: identifikasi area edentulous, segmentasi anatomi, rekomendasi dimensi implan, dan rekonstruksi 3D.' },
      ],
      tahap: [
        {
          singkat: 'Area Edentulous',
          judul: 'Identifikasi Area Edentulous',
          inti: 'Area edentulous adalah bagian rahang tempat gigi sudah tidak ada. Menemukan area ini menjadi langkah awal sebelum membahas perencanaan implan.',
          ilustrasi: 'Ilustrasi konseptual lengkung rahang bawah dari atas, dengan satu area molar tanpa gigi.',
          keterangan: 'Lima posisi slice melintasi lengkung rahang. Penanda pada posisi 3 menunjukkan area edentulous.',
          ditunjukkan: 'Lengkung rahang bawah dari atas, satu area molar tanpa gigi, dan lima posisi slice yang sudah diberi penanda.',
          peran_ai: 'Penelitian memakai model klasifikasi berbasis transfer learning untuk mengenali slice yang memuat area edentulous pada rahang bawah.',
          hasil: [{ label: 'F1-score MobileNetV2', nilai: '99,97%' }],
          hasil_catatan: 'MobileNetV2 memperoleh F1-score 99,97% pada evaluasi klasifikasi slice edentulous yang dilaporkan dalam penelitian. Angka ini tidak berlaku untuk keseluruhan alur perencanaan implan.',
          batasan: 'Penanda area pada ilustrasi ini telah disiapkan sebelumnya dan bukan keluaran model yang dijalankan secara langsung.',
          legenda: [
            { kelas: 'gigi', teks: 'Gigi' },
            { kelas: 'target', teks: 'Area edentulous' },
            { kelas: 'bergigi', teks: 'Penanda slice bergigi' },
            { kelas: 'edentulous', teks: 'Penanda slice edentulous' },
          ],
          kontrol: [
            { tipe: 'toggle', kunci: 'area', label: 'Tampilkan Area Edentulous' },
            { tipe: 'toggle', kunci: 'penanda', label: 'Tampilkan Penanda Klasifikasi' },
          ],
        },
        {
          singkat: 'Segmentasi Anatomi',
          judul: 'Segmentasi Anatomi Rahang',
          inti: 'Segmentasi berarti menandai area setiap struktur pada citra. Di sini ada dua struktur: tulang alveolar (AB) yang menopang gigi, dan kanal mandibula (MC) yang hubungannya dengan implan perlu diperhatikan.',
          ilustrasi: 'Ilustrasi konseptual penampang rahang pada satu posisi sepanjang lengkung rahang.',
          ditunjukkan: 'Penampang konseptual di lima posisi sepanjang lengkung rahang. Peta kecil menunjukkan posisi yang sedang dilihat.',
          peran_ai: 'DIR-YOLO (Dental Implant Recommender-YOLO) dikembangkan untuk segmentasi AB dan MC pada slice yang mengikuti lengkung rahang, bukan hanya tampilan koronal atau aksial biasa.',
          hasil: [
            { label: 'DSC AB', nilai: '97,65%' },
            { label: 'DSC MC', nilai: '86,59%' },
            { label: 'mDSC', nilai: '92,12%' },
          ],
          hasil_catatan: 'Hasil evaluasi yang dilaporkan dalam penelitian, bukan keluaran demo ini. Dice Similarity Coefficient (DSC) mengukur kemiripan area hasil segmentasi dengan area acuan.',
          batasan: 'Overlay anatomi ini merupakan ilustrasi konseptual, bukan hasil segmentasi dari citra CBCT pasien.',
          teknis: {
            judul: 'DIR-YOLO',
            pengantar: 'Berbasis modifikasi arsitektur YOLOv8n:',
            butir: ['Blok C2f diganti dengan C3Ghost.', 'Ghost Convolution mengurangi parameter model dan meningkatkan efisiensi komputasi.', 'Segmentation head dikurangi dari tiga skala menjadi dua skala, rendah dan tinggi.', 'Arsitektur disesuaikan dengan karakteristik AB dan MC.'],
          },
          legenda: [
            { kelas: 'tulang', teks: 'Tulang rahang' },
            { kelas: 'gigi', teks: 'Gigi' },
            { kelas: 'ab', teks: 'Alveolar Bone (AB), arsir' },
            { kelas: 'mc', teks: 'Mandibular Canal (MC), lingkaran' },
          ],
          kontrol: [
            { tipe: 'toggle', kunci: 'ab', label: 'Tampilkan Alveolar Bone' },
            { tipe: 'toggle', kunci: 'mc', label: 'Tampilkan Mandibular Canal' },
            { tipe: 'pilihan', kunci: 'slice', label: 'Posisi Slice', awal: '3', opsi: [
              { nilai: '1', label: '1', keterangan: 'Posisi 1 dari 5: slice bergigi di area premolar.' },
              { nilai: '2', label: '2', keterangan: 'Posisi 2 dari 5: slice bergigi di area premolar.' },
              { nilai: '3', label: '3', keterangan: 'Posisi 3 dari 5: area edentulous, tidak ada gigi di atas tulang.' },
              { nilai: '4', label: '4', keterangan: 'Posisi 4 dari 5: slice bergigi di area molar.' },
              { nilai: '5', label: '5', keterangan: 'Posisi 5 dari 5: slice bergigi di area molar.' },
            ] },
          ],
        },
        {
          singkat: 'Dimensi Implan',
          judul: 'Rekomendasi Dimensi Implan',
          inti: 'Diameter dan tinggi implan perlu disesuaikan dengan kondisi tulang. Penelitian tersebut mengkaji rekomendasi dimensi implan untuk molar rahang bawah berdasarkan informasi anatomi.',
          ilustrasi: 'Ilustrasi konseptual penampang rahang dengan implan skematis dan panah ukuran, tanpa nilai dan tanpa skala.',
          keterangan: 'Gambar tidak berskala. Jarak antarbagian tidak menunjukkan posisi yang aman.',
          ditunjukkan: 'Penampang skematis dengan implan, tulang, kanal, dan panah yang menjelaskan arti diameter dan tinggi implan.',
          peran_ai: 'Penelitian mengembangkan pendekatan rekomendasi dimensi implan secara otomatis. Rekomendasi berupa angka membutuhkan masukan anatomi nyata yang tervalidasi dan keluaran model, sehingga demo ini tidak menampilkan angka ukuran.',
          nilai: [
            { label: 'Diameter Implan', nilai: 'Ilustratif / Tidak Ditentukan' },
            { label: 'Tinggi Implan', nilai: 'Ilustratif / Tidak Ditentukan' },
          ],
          hasil: [
            { label: 'MAE diameter', nilai: '0,17 mm' },
            { label: 'MAE tinggi implan', nilai: '0,10 mm' },
          ],
          hasil_catatan: 'Evaluasi pada 30 pasien dalam penelitian yang dilaporkan, bukan ukuran dari demo ini. Mean Absolute Error (MAE) adalah rata-rata selisih absolut.',
          batasan: 'Ilustrasi ini menjelaskan konsep rekomendasi dimensi implan dan tidak menghasilkan ukuran yang dapat digunakan untuk tindakan klinis.',
          legenda: [
            { kelas: 'tulang', teks: 'Tulang rahang' },
            { kelas: 'ab', teks: 'Alveolar Bone (AB), arsir' },
            { kelas: 'mc', teks: 'Mandibular Canal (MC), lingkaran' },
            { kelas: 'implan', teks: 'Implan skematis' },
          ],
          kontrol: [
            { tipe: 'toggle', kunci: 'ukur', label: 'Tampilkan Pengukuran' },
            { tipe: 'toggle', kunci: 'anatomi', label: 'Tampilkan Struktur Anatomi' },
          ],
        },
        {
          singkat: 'Visualisasi 3D',
          judul: 'Visualisasi 3D Konseptual',
          inti: 'Penelitian juga merekonstruksi struktur anatomi hasil segmentasi agar penempatan implan dapat dilihat bersama dalam tiga dimensi.',
          ilustrasi: 'Simulasi visualisasi 3D konseptual rahang bawah, alveolar bone, kanal mandibula, dan implan skematis.',
          ditunjukkan: 'Simulasi visualisasi 3D konseptual dengan tiga sudut pandang yang sudah disiapkan. Struktur di dalam rahang digambar seolah tembus pandang.',
          peran_ai: 'Rekonstruksi dan visualisasi 3D diintegrasikan ke dalam aplikasi penelitian untuk perencanaan implan berbasis CBCT.',
          batasan: 'Visualisasi ini bersifat konseptual dan tidak merepresentasikan rekonstruksi 3D pasien atau keluaran aktual dari model penelitian.',
          legenda: [
            { kelas: 'tulang', teks: 'Rahang dan gigi' },
            { kelas: 'ab', teks: 'Alveolar Bone (AB)' },
            { kelas: 'mc', teks: 'Mandibular Canal (MC)' },
            { kelas: 'implan', teks: 'Implan skematis' },
          ],
          kontrol: [
            { tipe: 'toggle', kunci: 'rahang', label: 'Tampilkan Rahang' },
            { tipe: 'toggle', kunci: 'ab', label: 'Tampilkan Alveolar Bone' },
            { tipe: 'toggle', kunci: 'mc', label: 'Tampilkan Mandibular Canal' },
            { tipe: 'toggle', kunci: 'implan', label: 'Tampilkan Implan' },
            { tipe: 'pilihan', kunci: 'sudut', label: 'Sudut pandang', awal: 'perspektif', opsi: [
              { nilai: 'perspektif', label: 'Perspektif', keterangan: 'Sudut pandang perspektif, dari depan atas.' },
              { nilai: 'samping', label: 'Samping', keterangan: 'Sudut pandang samping, dari sisi area edentulous.' },
              { nilai: 'depan', label: 'Depan', keterangan: 'Sudut pandang depan.' },
            ] },
          ],
        },
      ],
      cara_kerja: [
        { judul: 'Transfer learning', teks: 'Model yang sudah dilatih untuk tugas lain disesuaikan untuk tugas baru. Penelitian tersebut memakai pendekatan tersebut pada klasifikasi slice edentulous.' },
        { judul: 'MobileNetV2', teks: 'Arsitektur yang memperoleh kinerja tertinggi pada evaluasi klasifikasi slice edentulous dalam penelitian tersebut.' },
        { judul: 'DIR-YOLO', teks: 'Model segmentasi AB dan MC hasil modifikasi YOLOv8n. Ghost Convolution dan C3Ghost dipakai untuk mengurangi parameter model, dan segmentation head memakai dua skala.' },
        { judul: 'Segmentasi anatomi', teks: 'Proses menandai area AB dan MC pada slice yang mengikuti lengkung rahang, sehingga kedua struktur dapat ditinjau secara terpisah.' },
        { judul: 'Rekomendasi dimensi implan', teks: 'Pendekatan otomatis dalam penelitian untuk merekomendasikan diameter dan tinggi implan. Demo ini hanya menjelaskan konsepnya dan tidak menghasilkan ukuran.' },
      ],
      hasil: [
        { judul: 'Klasifikasi area edentulous', nilai: '99,97%', metrik: 'F1-score MobileNetV2', rincian: 'Berlaku untuk klasifikasi slice edentulous.' },
        { judul: 'Segmentasi anatomi', nilai: '92,12%', metrik: 'mDSC DIR-YOLO', rincian: 'AB 97,65% dan MC 86,59%.' },
        { judul: 'Evaluasi diameter', nilai: '0,17 mm', metrik: 'Mean Absolute Error', rincian: 'Evaluasi pada 30 pasien.' },
        { judul: 'Evaluasi tinggi implan', nilai: '0,10 mm', metrik: 'Mean Absolute Error', rincian: 'Evaluasi pada 30 pasien.' },
      ],
      hasil_pengantar: 'Metrik yang ditampilkan merupakan hasil evaluasi yang dilaporkan dalam penelitian yang menjadi dasar demonstrasi ini, bukan hasil pengujian langsung aplikasi demo.',
      hasil_batasan: [
        'Angka ini tidak menunjukkan kesiapan klinis untuk penggunaan umum dan bukan akurasi diagnostik.',
        'Hasil evaluasi 30 pasien tidak dapat digeneralisasi ke semua populasi.',
      ],
      statistik: [
        "Cohen's Weighted Kappa: signifikansi p < 0,001. Nilai koefisien kesepakatan tidak tersedia pada ringkasan ini, sehingga tingkat kesepakatan belum dapat disimpulkan.",
        'Kruskal–Wallis: signifikansi asimptotik 0,658 untuk diameter dan 0,982 untuk tinggi implan. Hasil ini berarti tidak ditemukan perbedaan signifikan, bukan bukti kesetaraan klinis.',
      ],
      batasan: [
        'Seluruh gambar adalah ilustrasi konseptual buatan khusus untuk demo ini.',
        'Tidak ada dataset penelitian, citra CBCT, atau data pasien dalam demo ini.',
        'Tidak ada inferensi AI yang berjalan. Kontrol hanya mengganti tampilan yang sudah disiapkan.',
        'Demo ini bukan alat bantu keputusan klinis dan tidak menghasilkan rencana implan.',
      ],
    },
    tentang: {
      hero_judul: 'Mendekatkan Inovasi AI dengan Kebutuhan Nyata.',
      deskripsi: 'Melalui riset, kolaborasi, dan keahlian teknologi, UBAYA AI Center menghadirkan pengembangan solusi serta pembelajaran AI yang relevan bagi dunia akademik, industri, dan masyarakat.',
      visi_pengantar: 'Visi dan misi menjadi pijakan kami dalam mengembangkan pengetahuan, membangun kolaborasi, dan menghadirkan manfaat AI yang lebih luas.',
      fasilitas_pengantar: 'Kenali fasilitas yang mendukung kegiatan pembelajaran, diskusi, dan eksplorasi teknologi AI.',
      visi_judul: 'Menjadi AI Solution Factory terdepan',
      visi: 'Menghasilkan produk, menjadi pusat riset, serta meningkatkan kapasitas sumber daya manusia di bidang AI yang memberikan dampak nyata bagi akademik, industri, dan masyarakat.',
      misi: [
        'Mendorong riset AI yang inovatif dan aplikatif.',
        'Menghasilkan produk dan solusi AI yang siap digunakan.',
        'Mengembangkan talenta AI yang kompeten.',
        'Membangun kolaborasi strategis dengan pemerintah dan industri.',
        'Menciptakan ekosistem inovasi dan startup berbasis AI.',
      ],
      fasilitas: [
        { nama: 'Ruang pelatihan', gambar: 'ruang-pelatihan.jpg' },
        { nama: 'Ruang diskusi', gambar: 'ruang-diskusi.jpg' },
        { nama: 'Ruang demo' },
      ],
    },
  };

  return { berita, dokumentasi, dokumentasi_gambar, testimoni, tim, akun, kunci, konten };
})();
