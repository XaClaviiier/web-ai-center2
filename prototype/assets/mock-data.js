/* DATA DUMMY. Dibuang saat implementasi. Kolom mengikuti docs/database.md (snake_case).
   Tanggal publikasi: string DD-MM-YYYY. `gambar`, `foto`, `key` = nama file di assets/img. */
const MOCK = (() => {
  const audit = (i) => ({ created_at: `2026-10-${String((i % 27) + 1).padStart(2, '0')}T02:00:00Z`, created_by: 'sub-direktur', updated_by: 'sub-direktur' });
  const fotoContoh = ['ruang-pelatihan.jpg', 'ruang-diskusi.jpg', 'ruang-demo.jpg', 'implan-gigi.jpg', 'algae-finder.jpg', 'cacat-las.jpg', 'gesture-translator.jpg', 'xray-pneumonia.jpg', 'kesegaran-ikan.jpg'];
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
    isi: `${judul} menjadi salah satu kegiatan yang disiapkan AI Center UBAYA untuk membantu peserta memahami manfaat kecerdasan artifisial dalam pekerjaan sehari-hari.\n\nPeserta mempelajari konsep dasar, mencoba latihan singkat, lalu berdiskusi bersama tim tentang penerapan yang paling sesuai dengan kebutuhan mereka.\n\n${catatanContoh}`,
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
    judul: 'AI Center Ubaya Siap Dukung Layanan Kesehatan di Jawa Timur',
    isi: 'Pengembangan AI Center dibahas dalam audiensi Rektor Ubaya, Benny Lianto, dengan Gubernur Jawa Timur, Khofifah Indar Parawansa, di Gedung Negara Grahadi, Surabaya, Jumat (22/5/2026). Gubernur Khofifah menyambut positif langkah Ubaya mengembangkan teknologi AI untuk mendukung sektor kesehatan dan peningkatan kualitas sumber daya manusia.\n\nRektor Benny Lianto menyebut AI Center diproyeksikan menjadi pusat riset unggulan di Jawa Timur yang menghasilkan inovasi berbasis AI yang aplikatif, antara lain teknologi implan gigi dan aplikasi AI untuk membantu diagnosis penyakit.\n\nPengembangan AI kesehatan di Ubaya didukung ekosistem yang lengkap, mulai dari Fakultas Kedokteran, rumah sakit milik kampus, hingga kolaborasi dengan institusi kesehatan lain termasuk RSUD Dr. Soetomo. Artikel ini disadur dari siaran pers Dinas Komunikasi dan Informatika Provinsi Jawa Timur (23 Mei 2026).',
    tanggal_publikasi: '22-05-2026',
    gambar: 'implan-gigi.jpg',
    status: 'posted',
    version: 1,
    deleted_at: null,
    ...audit(20),
  });

  const karya = [
    ['Rekomendasi Implan Gigi Otomatis', '20-08-2026', 'implan-gigi.jpg'],
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
    { id: 'direktur', sub: 'sub-direktur', nama: 'Dian Direktur', role: 'Direktur' },
    { id: 'pegawai', sub: 'sub-pegawai', nama: 'Pandu Pegawai', role: 'Pegawai' },
    { id: 'tanpa-role', sub: 'sub-lain', nama: 'Tamu Tanpa Role', role: null },
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
      email_url: 'https://mail.google.com/mail/?view=cm&fs=1&to=' + encodeURIComponent('aicenter@unit.ubaya.ac.id') + '&su=' + encodeURIComponent('Kolaborasi dengan AI Center Ubaya'),
      instagram: '@aicenter_ubaya',
      instagram_url: 'https://instagram.com/aicenter_ubaya',
      alamat: ['Gedung Fakultas Teknik · TA 1.2', 'Jalan Raya Kalirungkut, Tenggilis, Surabaya'],
      peta_embed: 'https://maps.google.com/maps?q=-7.321919965577327,112.76790932621508&z=16&output=embed',
      peta_url: 'https://www.google.com/maps/search/?api=1&query=-7.321919965577327,112.76790932621508',
      deskripsi_footer: 'Pusat riset dan layanan kecerdasan artifisial Universitas Surabaya.',
    },
    hero: {
      judul: 'Pusat Riset & Layanan Kecerdasan Artifisial',
      sub: 'Kami membantu individu, institusi, dan bisnis mengembangkan keterampilan serta menerapkan AI sesuai kebutuhan.',
      gambar: 'hero-1-1600.webp',
    },
    layanan: [
      { nama: 'Pelatihan', deskripsi: 'Kembangkan pengetahuan dan keterampilan AI melalui pelatihan yang dirancang untuk individu, komunitas, institusi, hingga kebutuhan tim di dunia kerja.', href: 'pelatihan.html' },
      { nama: 'Inference and Solutions', deskripsi: 'Bantu organisasi Anda menerapkan model AI ke dalam sistem atau proses kerja, dengan dukungan integrasi dan pendampingan teknis sesuai kebutuhan.', href: 'inference.html' },
    ],
    pelatihan: {
      hero_judul: 'Pelatihan AI untuk kebutuhan belajar dan kerja.',
      hero_sub: 'Kenali Pelatihan Reguler yang sedang disiapkan, Pelatihan Khusus sesuai kebutuhan kelompok, serta Workshop atau Seminar untuk Bisnis.',
      reguler: 'Program pelatihan reguler sedang kami siapkan.',
      khusus_sub: 'Ceritakan tujuan belajar dan kebutuhan peserta. AI Center UBAYA dapat membahas format pelatihan untuk sekolah, kampus, komunitas, atau organisasi.',
      // Empat langkah berikut dikonfirmasi sebagai proses layanan oleh pemilik.
      alur_konsultasi: [
        { judul: 'Ceritakan kebutuhan', deskripsi: 'Hubungi kami dan sampaikan tujuan serta peserta pelatihan.' },
        { judul: 'Diskusi bersama tim', deskripsi: 'Kami memetakan materi, durasi, dan format yang paling sesuai.' },
        { judul: 'Terima usulan program', deskripsi: 'Anda menerima usulan program untuk ditinjau bersama.' },
        { judul: 'Pelatihan berjalan', deskripsi: 'Program dijalankan dan didampingi oleh tim AI Center.' },
      ],
      workshop_sub: 'Bantu tim dan pimpinan bisnis memahami penerapan AI yang relevan dengan pekerjaan mereka melalui workshop atau seminar.',
    },
    inference: {
      hero_judul: 'Bawa model AI ke penggunaan nyata.',
      hero_sub: 'Inference adalah proses menggunakan model AI untuk menghasilkan prediksi pada data baru. AI Center UBAYA membantu membahas bagaimana hasil tersebut dapat digunakan oleh aplikasi atau sistem sesuai kebutuhan Anda.',
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
        { judul: 'Rekomendasi Implan Gigi Otomatis', bidang: 'Kesehatan', deskripsi: 'Menganalisis citra CBCT 3D untuk merekomendasikan ukuran, posisi, dan sudut implan gigi secara otomatis.', gambar: 'implan-gigi.jpg' },
        { judul: 'Algae Finder', bidang: 'Lingkungan', deskripsi: 'Mendeteksi dan menghitung jenis mikroalga di perairan dari citra mikroskop.', gambar: 'algae-finder.jpg' },
        { judul: 'Deteksi Cacat Las', bidang: 'Industri', deskripsi: 'Menilai kualitas hasil pengelasan langsung dari foto sambungan las.', gambar: 'cacat-las.jpg' },
      ],
    },
    tentang: {
      hero_judul: 'Mendekatkan Inovasi AI dengan Kebutuhan Nyata.',
      deskripsi: 'Melalui riset, kolaborasi, dan keahlian teknologi, AI Center UBAYA menghadirkan pengembangan solusi serta pembelajaran AI yang relevan bagi dunia akademik, industri, dan masyarakat.',
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
