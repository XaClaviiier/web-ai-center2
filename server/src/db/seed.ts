/**
 * Seed data development untuk lima konten CMS:
 * dokumentasi, berita, kursus, tim, dan testimoni.
 *
 * Reset bersifat destruktif dan wajib memakai confirmation eksplisit.
 * Tabel konten static legacy tidak disentuh.
 */
import type { Db } from './client';
import { createDb, resolveDbTarget } from './client';
import { resolveConfig } from '../config';
import { berita, dokumentasi, heroSlides, inference, klien, kursus, layanan, profil, tim, testimoni } from './schema';

type BeritaSeed = Omit<typeof berita.$inferInsert, 'id' | 'createdAt'>;
type DokumentasiSeed = Omit<typeof dokumentasi.$inferInsert, 'id' | 'createdAt'>;
type TimSeed = Omit<typeof tim.$inferInsert, 'id'>;
type LayananSeed = Omit<typeof layanan.$inferInsert, 'id'>;
type HeroSeed = Omit<typeof heroSlides.$inferInsert, 'id'>;
type KlienSeed = Omit<typeof klien.$inferInsert, 'id'>;
type TestimoniSeed = Omit<typeof testimoni.$inferInsert, 'id'>;
type ProfilSeed = typeof profil.$inferInsert;
type InferenceSeed = typeof inference.$inferInsert;
type KursusSeed = Omit<typeof kursus.$inferInsert, 'id'>;

// Kebijakan konten (Okt 2026): hanya berita/kegiatan yang bisa diverifikasi dari
// sumber resmi — TIDAK ada entri karangan. Berita = saduran siaran pers Kominfo
// Jatim 23 Mei 2026; dokumentasi = 6 karya portofolio asli.
export const SEED_BERITA: BeritaSeed[] = [
  {
    slug: 'ai-center-dukung-layanan-kesehatan-jatim',
    judul: 'AI Center Ubaya Siap Dukung Layanan Kesehatan di Jawa Timur',
    ringkasan:
      'Rektor Ubaya bertemu Gubernur Jawa Timur di Grahadi; AI Center disiapkan sebagai pusat riset unggulan yang menjawab persoalan nyata, terutama di bidang kesehatan.',
    isi: 'Pengembangan AI Center dibahas dalam audiensi Rektor Ubaya, Benny Lianto, dengan Gubernur Jawa Timur, Khofifah Indar Parawansa, di Gedung Negara Grahadi, Surabaya, Jumat (22/5/2026). Gubernur Khofifah menyambut positif langkah Ubaya mengembangkan teknologi AI untuk mendukung sektor kesehatan dan peningkatan kualitas sumber daya manusia. Rektor Benny Lianto menyebut AI Center diproyeksikan menjadi pusat riset unggulan di Jawa Timur yang menghasilkan inovasi berbasis AI yang aplikatif, antara lain teknologi implan gigi dan aplikasi AI untuk membantu diagnosis penyakit. Pengembangan AI kesehatan di Ubaya didukung ekosistem yang lengkap, mulai dari Fakultas Kedokteran, rumah sakit milik kampus, hingga kolaborasi dengan institusi kesehatan lain termasuk RSUD Dr. Soetomo. Artikel ini disadur dari siaran pers Dinas Komunikasi dan Informatika Provinsi Jawa Timur (23 Mei 2026).',
    tanggal: '2026-05-22',
    penulis: 'AI Center Ubaya',
    gambar: null,
  },
];

export const SEED_DOKUMENTASI: DokumentasiSeed[] = [
  // Konsul PROGRESS 2: karya portofolio jadi entri dokumentasi (tanggal = terbit karya).
  {
    slug: 'klasifikasi-xray-pneumonia',
    judul: 'Klasifikasi X-Ray Pneumonia',
    deskripsi: 'Mengklasifikasi citra rontgen dada untuk indikasi pneumonia menggunakan ResNet152V2 + SVM.',
    tanggal: '2026-09-15',
    kategori: 'Kesehatan',
    gambar: '/portofolio/xray-pneumonia.jpg',
  },
  {
    slug: 'implan-gigi-otomatis',
    judul: 'Rekomendasi Implan Gigi Otomatis',
    deskripsi:
      'Menganalisis citra CBCT 3D untuk merekomendasikan ukuran, posisi, dan sudut implan gigi secara otomatis.',
    tanggal: '2026-08-20',
    kategori: 'Kesehatan',
    gambar: '/portofolio/implan-gigi.jpg',
  },
  {
    slug: 'algae-finder',
    judul: 'Algae Finder',
    deskripsi:
      'Mendeteksi dan menghitung jenis mikroalga di perairan (mis. Thalassiosira & Nannochloropsis) dari citra mikroskop.',
    tanggal: '2026-07-28',
    kategori: 'Lingkungan',
    gambar: '/portofolio/algae-finder.jpg',
  },
  {
    slug: 'translator-bahasa-isyarat',
    judul: 'Translator Bahasa Isyarat',
    deskripsi:
      'Menerjemahkan gerakan bahasa isyarat menjadi teks secara real-time melalui kamera, lengkap dengan skor keyakinan.',
    tanggal: '2026-07-02',
    kategori: 'Aksesibilitas',
    gambar: '/portofolio/gesture-translator.jpg',
  },
  {
    slug: 'deteksi-cacat-las',
    judul: 'Deteksi Cacat Las (Welding)',
    deskripsi:
      'Menilai kualitas hasil pengelasan — Good Weld, Bad Weld, atau Defect — langsung dari foto sambungan las.',
    tanggal: '2026-05-28',
    kategori: 'Industri',
    gambar: '/portofolio/cacat-las.jpg',
  },
  {
    slug: 'deteksi-kesegaran-ikan',
    judul: 'Deteksi Kesegaran Ikan',
    deskripsi:
      'Mengenali jenis ikan dan tingkat kesegarannya dari foto, langsung dari ponsel, beserta informasi nutrisi.',
    tanggal: '2026-04-15',
    kategori: 'Pangan',
    gambar: '/portofolio/kesegaran-ikan.jpg',
  },
];

export const SEED_TIM: TimSeed[] = [
  { nama: 'Dr. Mohammad Farid Naufal', peran: 'Ketua', kredensial: null, foto: '/tim/farid-naufal.jpg', urutan: 0 },
  {
    nama: 'Dr. Monica Widiasri',
    peran: 'Koordinator Riset & Edukasi',
    kredensial: null,
    foto: '/tim/monica-widiasri.jpg',
    urutan: 1,
  },
  { nama: 'Prof. Joko Siswantoro', peran: 'Tim Riset', kredensial: null, foto: '/tim/joko-siswantoro.jpg', urutan: 2 },
  {
    nama: 'Marco Ariano Kristyanto',
    peran: 'Tim Hardware',
    kredensial: 'M.M., M.Kom.',
    foto: '/tim/marco-kristyanto.jpg',
    urutan: 3,
  },
  {
    nama: 'Fikri Baharuddin',
    peran: 'Tim Software',
    kredensial: 'M.Kom.',
    foto: '/tim/fikri-baharuddin.jpg',
    urutan: 4,
  },
  {
    nama: 'Jabesh Nehemiah Wijaya',
    peran: 'Tim Software',
    kredensial: 'S.Kom., M.Kom.',
    foto: '/tim/jabesh-wijaya.jpg',
    urutan: 5,
  },
];

export const SEED_LAYANAN: LayananSeed[] = [
  {
    slug: 'pelatihan',
    nama: 'Pelatihan',
    // Copy disinkronkan dengan src/data/layanan.ts (konsul PROGRESS 2 — gaya LPPM).
    tagline: 'Pelatihan AI yang benar-benar dipakai sehari-hari.',
    deskripsi:
      'Ubaya AI Center mendampingi mahasiswa, dosen, guru, profesional, dan masyarakat umum ' +
      'belajar AI — dari pengenalan dasar, praktik dengan studi kasus, sampai penerapan di kerja nyata. ' +
      'Etis, aplikatif, dan berdampak.',
    fitur: [
      'Workshop & bootcamp terjadwal',
      'Pelatihan kustom sesuai kebutuhan',
      'Materi terstruktur dan mudah diikuti',
      'Studi kasus akademik dan profesional',
      'Sertifikat setelah menyelesaikan program',
    ],
  },
  {
    slug: 'inference-solution',
    nama: 'Inference Solution',
    tagline: 'Solusi deployment model untuk kebutuhan industri.',
    deskripsi:
      'Layanan deployment model machine learning menjadi solusi inference siap pakai untuk mitra industri, ' +
      'didampingi tim AI Center dari integrasi sampai berjalan di infrastruktur Anda.',
    fitur: [
      'Deployment model menjadi API layanan',
      'Pendampingan integrasi infrastruktur',
      'Dokumentasi teknis yang jelas',
      'Dukungan purna implementasi',
    ],
  },
];

export const SEED_HERO: HeroSeed[] = [
  {
    urutan: 0,
    eyebrow: 'AI Center · Universitas Surabaya',
    judul: 'Pusat Riset & Layanan',
    judulAksen: 'Kecerdasan Artifisial',
    sub: 'Kursus AI/ML dan inference solution untuk sivitas akademika serta mitra industri.',
    // CTA disinkronkan dengan src/data/hero.ts (frontend). Link konsultasi WA
    // memakai nomor + pesan statis situs lama (lihat src/data/kontak.ts).
    ctaPrimer: {
      label: 'Jadwalkan Konsultasi AI',
      to: 'https://wa.me/62895634222240?text=Halo%20Ubaya%20AI%20Center%2C%20saya%20ingin%20berdiskusi%20mengenai%20layanan%2Fkerja%20sama%20AI.',
    },
    ctaSekunder: { label: 'Lihat Berita', to: '/berita' },
    badgeJudul: 'AI Center Ubaya',
    badgeSub: 'Riset · Pelatihan · Layanan',
    image: '/hero-1-1600.webp',
    srcSet: '/hero-1-800.webp 800w, /hero-1-1600.webp 1600w',
    sizes: '100vw',
  },
  {
    urutan: 1,
    eyebrow: 'Inference Solution',
    judul: 'Akselerasi Riset dan Produk',
    judulAksen: 'untuk Kebutuhan Nyata',
    sub: 'Kembangkan dan deploy model Anda sebagai layanan inference dengan pendampingan AI Center.',
    ctaPrimer: { label: 'Mulai Transformasi Bisnis Anda', to: '/layanan/inference-solution' },
    ctaSekunder: { label: 'Tim Kami', to: '/tim' },
    badgeJudul: 'Inference Solution',
    badgeSub: 'Integrasi model + dukungan teknis',
    image: '/hero-2-1600.webp',
    srcSet: '/hero-2-800.webp 800w, /hero-2-1600.webp 1600w',
    // Konsul 2 Okt: teks slide Inference dipindah ke samping kanan (layout
    // 'teks-kanan', kolom `layout` di DB dipakai lagi) supaya subjek foto di
    // kiri tidak tertutup teks. Slide 1 tetap default (null).
    layout: 'teks-kanan',
    sizes: '100vw',
  },
];

// Sinkron dengan src/data/klien.ts (frontend) — aturan seed = data frontend,
// supaya tampilan "Klien Kami" sama persis dengan atau tanpa backend.
// TODO_KONTEN: nama FTB/CAW/Ubaya masih placeholder — konfirmasi nama resmi
// + bidang ke user sebelum rilis.
export const SEED_KLIEN: KlienSeed[] = [
  { nama: 'FTB', bidang: 'Mitra Fakultas (placeholder)', urutan: 0 },
  { nama: 'CAW', bidang: 'Mitra (placeholder)', urutan: 1 },
  { nama: 'Ubaya', bidang: 'Universitas Surabaya', urutan: 2 },
];

export const SEED_TESTIMONI: TestimoniSeed[] = [
  {
    nama: 'Peserta Pelatihan ML',
    peran: 'Mahasiswa',
    kutipan:
      'Materi pelatihan runtut dan langsung praktik. Akses GPU lab membuat eksperimen tugas akhir jauh lebih cepat.',
  },
  {
    nama: 'Mitra Industri',
    peran: 'Pengguna Inference Solution',
    kutipan:
      'Proses deployment model didampingi sampai jalan di infrastruktur kami. Komunikasi tim responsif dan dokumentasinya jelas.',
  },
  {
    nama: 'Dosen Peneliti',
    peran: 'Peneliti',
    kutipan:
      'Pendampingan AI Center sangat membantu riset. Tim responsif dan memberi dukungan teknis saat ada kendala.',
  },
];

export const SEED_PROFIL: ProfilSeed = {
  id: 1,
  nama: 'AI Center Universitas Surabaya',
  tagline: 'Pusat riset dan layanan kecerdasan artifisial dalam ekosistem LPPM Universitas Surabaya.',
  ringkasan:
    'AI Center Universitas Surabaya adalah pusat riset dan layanan kecerdasan artifisial yang mendukung ' +
    'pendidikan, penelitian, dan pengabdian masyarakat dalam ekosistem LPPM Universitas Surabaya. ' +
    'Layanan utama: pelatihan AI/ML dan solusi inference untuk industri.',
  alamat: 'Gedung Perpustakaan LT.4, Jalan Raya Kalirungkut, Tenggilis, Surabaya',
  email: 'aicenter@unit.ubaya.ac.id',
  telepon: '0895-6342-22240',
  // Visi resmi: judul + deskripsi digabung satu baris teks.
  visi:
    'Menjadi AI Solution Factory terdepan — Menghasilkan produk, menjadi pusat riset, serta ' +
    'meningkatkan kapasitas sumber daya manusia di bidang AI yang memberikan dampak nyata bagi ' +
    'akademik, industri, dan masyarakat.',
  // Misi resmi: satu poin per baris (frontend memecah dengan split('\n')).
  misi: [
    'Mendorong riset AI yang inovatif dan aplikatif.',
    'Menghasilkan produk dan solusi AI yang siap digunakan.',
    'Mengembangkan talenta AI yang kompeten.',
    'Membangun kolaborasi strategis dengan pemerintah dan industri.',
    'Menciptakan ekosistem inovasi dan startup berbasis AI.',
  ].join('\n'),
  statistik: null,
};

/**
 * Konten halaman Inference Solution — disalin dari src/data/inference.ts
 * (`apaItuInference`, `kebutuhanInference`, `alurInference`,
 * `contohInferenceIntro`, `contohInference`). Guard di sinkron.test.ts.
 */
export const SEED_INFERENCE: InferenceSeed = {
  id: 1,
  judulApaItu: 'Apa itu Inference Solution?',
  deskripsiApaItu:
    'Inference adalah tahap menjalankan model machine learning untuk memprediksi data baru — bagian ' +
    'yang membuat model benar-benar dipakai, bukan sekadar dilatih. Inference Solution adalah layanan ' +
    'Ubaya AI Center yang mengubah model Anda menjadi layanan siap pakai — biasanya berupa API — ' +
    'sehingga aplikasi, website, atau sistem internal Anda dapat memanggil prediksi kapan pun ' +
    'dibutuhkan, diinfrastruktur yang terkelola dan didampingi tim kami.',
  kebutuhan: [
    'Model riset atau prototipe sudah jadi, tetapi belum bisa dipakai tim lain',
    'Butuh fitur AI di aplikasi atau website tanpa membangun infrastruktur ML sendiri',
    'Perlu prediksi otomatis yang konsisten untuk gambar, teks, atau data',
    'Ingin hasil model tetap andal performanya saat dipakai banyak pengguna',
  ],
  alur: [
    {
      nomor: '01',
      judul: 'Konsultasi & audit kebutuhan',
      deskripsi:
        'Kami memetakan use case, data, dan model yang sudah — atau yang perlu disiapkan — bersama tim Anda.',
    },
    {
      nomor: '02',
      judul: 'Persiapan model',
      deskripsi:
        'Model dioptimalkan dan dibungkus agar siap dijalankan di server, termasuk preprocessing dan versi yang terkendali.',
    },
    {
      nomor: '03',
      judul: 'Deployment sebagai API',
      deskripsi:
        'Model dijalankan sebagai endpoint layanan (API) yang stabil, lengkap dengan dokumentasi pemakaian.',
    },
    {
      nomor: '04',
      judul: 'Integrasi & pengujian',
      deskripsi:
        'API disambungkan ke aplikasi Anda lalu diuji pada data nyata: akurasi, latensi, dan perilakunya.',
    },
    {
      nomor: '05',
      judul: 'Operasional & purna implementasi',
      deskripsi:
        'Monitoring, pembaruan model, dan dukungan teknis agar layanan tetap andal setelah berjalan.',
    },
  ],
  contohIntro: 'Beberapa produk AI Center yang menjalankan prediksi model secara langsung:',
  contoh: [
    { slug: 'implan-gigi-otomatis', judul: 'Rekomendasi Implan Gigi Otomatis' },
    { slug: 'algae-finder', judul: 'Algae Finder' },
    { slug: 'translator-bahasa-isyarat', judul: 'Translator Bahasa Isyarat' },
    { slug: 'deteksi-cacat-las', judul: 'Deteksi Cacat Las (Welding)' },
    { slug: 'klasifikasi-xray-pneumonia', judul: 'Klasifikasi X-Ray Pneumonia' },
    { slug: 'deteksi-kesegaran-ikan', judul: 'Deteksi Kesegaran Ikan' },
  ],
};

/** Data kursus — disalin dari `kursusDummy` frontend (src/data/pelatihan.ts). */
export const SEED_KURSUS: KursusSeed[] = [
  {
    urutan: 0,
    kode: 'R01',
    target: ['Mahasiswa', 'Dosen', 'Masyarakat umum'],
    judul: 'AI untuk Mencari dan Mengelola Referensi Jurnal',
    deskripsi:
      'Bangun alur kerja riset yang lebih terarah dengan bantuan AI, mulai dari menemukan jurnal hingga ' +
      'mengelola sitasi secara bertanggung jawab.',
    tentang:
      'Kursus praktis untuk mahasiswa, dosen, dan masyarakat umum yang ingin memakai AI sebagai alat bantu ' +
      'riset tanpa mengabaikan akurasi, etika, dan integritas akademik.',
    durasi: '4 sesi',
    level: 'Pemula',
    format: 'Online mandiri',
    instruktur: 'Tim Riset Ubaya AI Center',
    peran: 'Pengajar dan praktisi riset AI',
    inisial: 'RA',
    hasil: [
      'Merumuskan kata kunci dan query akademik',
      'Menguji relevansi sumber dengan tools AI',
      'Mengenali halusinasi dan referensi palsu',
      'Mengelola sitasi dengan rapi',
    ],
    modul: [
      {
        judul: 'Merumuskan pertanyaan dan kata kunci riset',
        deskripsi: 'Ubah topik menjadi pertanyaan riset dan query akademik yang terarah.',
        meta: '4 video · 35 menit',
      },
      {
        judul: 'Mencari jurnal ilmiah dengan Semantic Scholar',
        deskripsi: 'Temukan sumber primer, telusuri sitasi, dan buat daftar bacaan awal.',
        meta: '5 video · 48 menit',
      },
      {
        judul: 'Menguji relevansi dengan Consensus dan Elicit',
        deskripsi: 'Bandingkan temuan AI dengan isi jurnal dan cek kualitas buktinya.',
        meta: '4 video · 42 menit',
      },
      {
        judul: 'Meringkas jurnal dan mengelola sitasi',
        deskripsi: 'Buat ringkasan berbasis konteks dan rapikan pustaka dengan Zotero atau Mendeley.',
        meta: '5 video · 55 menit',
      },
    ],
  },
  {
    urutan: 1,
    kode: 'E01',
    target: ['Guru', 'Dosen', 'Masyarakat umum'],
    judul: 'Merancang Pembelajaran dengan AI',
    deskripsi:
      'Rancang aktivitas kelas, asesmen, dan umpan balik yang lebih personal dengan AI yang kritis dan ' +
      'bertanggung jawab.',
    tentang:
      'Program untuk pendidik yang ingin mengintegrasikan AI ke dalam perencanaan dan praktik pembelajaran ' +
      'secara efektif.',
    durasi: '3 sesi',
    level: 'Pemula',
    format: 'Online mandiri',
    instruktur: 'Tim Edukasi Ubaya AI Center',
    peran: 'Pengajar dan fasilitator pendidikan',
    inisial: 'ED',
    hasil: [
      'Menyusun ide pembelajaran dengan AI',
      'Membuat asesmen yang lebih adaptif',
      'Menguji kualitas output AI',
      'Menjaga etika dan privasi peserta didik',
    ],
    modul: [
      {
        judul: 'Memetakan kebutuhan belajar dan tujuan kelas',
        deskripsi: 'Tentukan tujuan belajar dan konteks peserta didik sebelum memakai AI.',
        meta: '4 video · 35 menit',
      },
      {
        judul: 'Menyusun materi serta aktivitas pembelajaran',
        deskripsi: 'Kembangkan bahan ajar, aktivitas, dan contoh yang relevan.',
        meta: '5 video · 45 menit',
      },
      {
        judul: 'Membuat asesmen dan rubrik dengan AI',
        deskripsi: 'Rancang asesmen, rubrik, dan umpan balik yang tetap dikontrol pendidik.',
        meta: '4 video · 40 menit',
      },
    ],
  },
  {
    urutan: 2,
    kode: 'P01',
    target: ['Mahasiswa', 'Masyarakat umum'],
    judul: 'Produktivitas Akademik dengan AI',
    deskripsi:
      'Bangun alur kerja untuk brainstorming, menulis, menganalisis data, dan mempresentasikan ide tanpa ' +
      'mengorbankan integritas akademik.',
    tentang:
      'Kursus fundamental untuk membangun kebiasaan kerja yang lebih produktif dengan AI, dari ide awal ' +
      'hingga presentasi yang terstruktur.',
    durasi: '3 sesi',
    level: 'Pemula',
    format: 'Online mandiri',
    instruktur: 'Tim Talenta Ubaya AI Center',
    peran: 'Pengajar dan mentor produktivitas AI',
    inisial: 'TA',
    hasil: [
      'Membuat workflow kerja dengan AI',
      'Mengolah ide dan data secara terstruktur',
      'Memeriksa kualitas tulisan dan analisis',
      'Menggunakan AI secara jujur dan bertanggung jawab',
    ],
    modul: [
      {
        judul: 'Brainstorming dan perencanaan tugas',
        deskripsi: 'Pecah tugas besar menjadi langkah kerja yang jelas dan realistis.',
        meta: '4 video · 32 menit',
      },
      {
        judul: 'Menulis, menyunting, dan memeriksa ide',
        deskripsi: 'Gunakan AI sebagai partner berpikir tanpa menggantikan suara penulis.',
        meta: '5 video · 46 menit',
      },
      {
        judul: 'Menganalisis data dan menyajikan temuan',
        deskripsi: 'Ubah data dan ide menjadi kesimpulan serta presentasi yang mudah dipahami.',
        meta: '4 video · 43 menit',
      },
    ],
  },
];

/** Confirmation required before destructive seed reset. */
export const SEED_CONFIRMATION = 'RESET';

/** Isi ulang semua tabel dummy (hapus dulu → idempoten & deterministik). */
export async function runSeed(db: Db, options: { allowDestructive?: boolean } = {}): Promise<void> {
  if (options.allowDestructive !== true) {
    throw new Error(`Seed destruktif diblokir — set SEED_CONFIRM=${SEED_CONFIRMATION} untuk reset eksplisit`);
  }

  await db.delete(berita);
  await db.delete(dokumentasi);
  await db.delete(tim);
  await db.delete(testimoni);
  await db.delete(kursus);

  if (SEED_BERITA.length > 0) await db.insert(berita).values(SEED_BERITA);
  if (SEED_DOKUMENTASI.length > 0) await db.insert(dokumentasi).values(SEED_DOKUMENTASI);
  if (SEED_TIM.length > 0) await db.insert(tim).values(SEED_TIM);
  if (SEED_TESTIMONI.length > 0) await db.insert(testimoni).values(SEED_TESTIMONI);
  if (SEED_KURSUS.length > 0) await db.insert(kursus).values(SEED_KURSUS);
}

/** CLI: `SEED_CONFIRM=RESET bun run db:seed` — exit code 1 bila gagal. */
if (import.meta.main) {
  if (process.env.SEED_CONFIRM !== SEED_CONFIRMATION) {
    console.error(`Seed diblokir — jalankan ulang dengan SEED_CONFIRM=${SEED_CONFIRMATION}`);
    process.exit(1);
  }

  const config = resolveConfig(process.env);
  const target = resolveDbTarget(config.databaseUrl);
  try {
    const handle = await createDb(target);
    try {
      await runSeed(handle.db, { allowDestructive: true });
      console.log(`Seed selesai (db: ${target.kind}).`);
    } finally {
      await handle.close();
    }
  } catch (error) {
    console.error('Seed gagal:', error);
    process.exit(1);
  }
}
