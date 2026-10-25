// Konten halaman Inference Solution — disusun dari pemahaman layanan inference:
// tahap menjalankan model ML (prediksi data baru) yang di-deploy menjadi layanan siap pakai
// (API) untuk mitra, dari integrasi sampai purna implementasi — konsisten dengan hero slide 2
// (hero.ts) dan deskripsi layanan di layanan.ts. Situs lama tidak punya section ini sendiri
// (pillar lama: Kolaborasi Penelitian, Pelatihan & Talenta, Komputasi Performa Tinggi).
// KONTEN STATIS frontend (sumber satu-satunya) — diedit langsung di file ini
// (tidak lewat backend/admin); LayananDetail memakai kontenInferenceDummy langsung.
import { buildWaLink, kontakDummy } from './kontak.ts';
import { portofolioDummy } from './portofolio.ts';

/** Panel penjelasan "Apa itu Inference Solution?" (ditampilkan di bawah deskripsi layanan). */
export const apaItuInference = {
  judul: 'Apa itu Inference Solution?',
  deskripsi:
    'Inference adalah tahap menjalankan model machine learning untuk memprediksi data baru — bagian ' +
    'yang membuat model benar-benar dipakai, bukan sekadar dilatih. Inference Solution adalah layanan ' +
    'Ubaya AI Center yang mengubah model Anda menjadi layanan siap pakai — biasanya berupa API — ' +
    'sehingga aplikasi, website, atau sistem internal Anda dapat memanggil prediksi kapan pun ' +
    'dibutuhkan, diinfrastruktur yang terkelola dan didampingi tim kami.',
};

/** Tanda kebutuhan layanan inference — kebutuhan umum mitra. */
export const kebutuhanInference: string[] = [
  'Model riset atau prototipe sudah jadi, tetapi belum bisa dipakai tim lain',
  'Butuh fitur AI di aplikasi atau website tanpa membangun infrastruktur ML sendiri',
  'Perlu prediksi otomatis yang konsisten untuk gambar, teks, atau data',
  'Ingin hasil model tetap andal performanya saat dipakai banyak pengguna',
];

/** Satu langkah pada alur kerja layanan inference. */
export interface LangkahInference {
  /** Nomor langkah tampil, mis. '01'. */
  nomor: string;
  judul: string;
  deskripsi: string;
}

/** Alur kerja layanan — 5 langkah dari konsultasi sampai purna implementasi. */
export const alurInference: LangkahInference[] = [
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
];

/** Subjudul section contoh penerapan. */
export const contohInferenceIntro =
  'Beberapa produk AI Center yang menjalankan prediksi model secara langsung:';

/**
 * Contoh penerapan — judul karya portofolio yang berbasis prediksi model
 * (disaring dari portofolioDummy agar tidak ada duplikasi teks).
 */
export const contohInference = portofolioDummy.map((k) => ({
  slug: k.slug,
  judul: k.judul,
}));

/**
 * Konten halaman inference statis frontend.
 * `LayananDetail` memakai data ini langsung; backend tidak mengelolanya.
 */
export interface KontenInference {
  /** Panel "Apa itu Inference Solution?" */
  judulApaItu: string;
  deskripsiApaItu: string;
  /** Tanda kebutuhan ("Kapan Anda membutuhkannya?"). */
  kebutuhan: string[];
  /** Alur kerja 5 langkah. */
  alur: LangkahInference[];
  contohIntro: string;
  /** Chip "Contoh penerapan" (slug dipakai sebagai key React). */
  contoh: Array<{ slug: string; judul: string }>;
}

/** Konten inference frontend — bukan fallback API. */
export const kontenInferenceDummy: KontenInference = {
  judulApaItu: apaItuInference.judul,
  deskripsiApaItu: apaItuInference.deskripsi,
  kebutuhan: kebutuhanInference,
  alur: alurInference,
  contohIntro: contohInferenceIntro,
  contoh: contohInference,
};

/** Link WhatsApp diskusi kebutuhan inference solution. */
export function waDiskusiInference(): string {
  return buildWaLink(
    kontakDummy.whatsappNumber,
    `${kontakDummy.whatsappGreeting} Saya ingin membahas kebutuhan inference solution ` +
      '(deployment model) untuk kebutuhan kami.',
  );
}
