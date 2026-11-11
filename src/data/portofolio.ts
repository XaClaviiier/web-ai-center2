// Konten section Portofolio — migrasi dari index.html situs lama (section `#portofolio`):
// 6 karya AI Center per bidang; foto diekstrak dari base64 di halaman lama ke public/portofolio/.
export interface Karya {
  /** Slug unik untuk key React. */
  slug: string;
  /** Bidang aplikasi, mis. 'Kesehatan'. */
  kategori: string;
  /** Teknologi pada chip, mis. 'Computer Vision 3D'. */
  teknologi: string;
  judul: string;
  deskripsi: string;
  /** Path gambar di `public/portofolio/`. */
  gambar: string;
  alt: string;
  /** Dimensi asli gambar (atribut width/height anti-CLS). */
  lebar: number;
  tinggi: number;
}

export const portofolioDummy: Karya[] = [
  {
    slug: 'implan-gigi-otomatis',
    kategori: 'Kesehatan',
    teknologi: 'Computer Vision 3D',
    judul: 'Rekomendasi Implan Gigi Otomatis',
    deskripsi:
      'Menganalisis citra CBCT 3D untuk merekomendasikan ukuran, posisi, dan sudut implan gigi secara otomatis.',
    gambar: '/portofolio/implan-gigi-konseptual.svg',
    alt: 'Aplikasi rekomendasi implan gigi otomatis',
    lebar: 960,
    tinggi: 600,
  },
  {
    slug: 'algae-finder',
    kategori: 'Lingkungan',
    teknologi: 'Klasifikasi Citra',
    judul: 'Algae Finder',
    deskripsi:
      'Mendeteksi dan menghitung jenis mikroalga di perairan (mis. Thalassiosira & Nannochloropsis) dari citra mikroskop.',
    gambar: '/portofolio/algae-finder.jpg',
    alt: 'Algae Finder — deteksi alga perairan',
    lebar: 900,
    tinggi: 621,
  },
  {
    slug: 'translator-bahasa-isyarat',
    kategori: 'Aksesibilitas',
    teknologi: 'Gesture AI Translator',
    judul: 'Translator Bahasa Isyarat',
    deskripsi:
      'Menerjemahkan gerakan bahasa isyarat menjadi teks secara real-time melalui kamera, lengkap dengan skor keyakinan.',
    gambar: '/portofolio/gesture-translator.jpg',
    alt: 'Translator bahasa isyarat',
    lebar: 900,
    tinggi: 740,
  },
  {
    slug: 'deteksi-cacat-las',
    kategori: 'Industri',
    teknologi: 'Object Detection',
    judul: 'Deteksi Cacat Las (Welding)',
    deskripsi:
      'Menilai kualitas hasil pengelasan — Good Weld, Bad Weld, atau Defect — langsung dari foto sambungan las.',
    gambar: '/portofolio/cacat-las.jpg',
    alt: 'Deteksi cacat las',
    lebar: 900,
    tinggi: 792,
  },
  {
    slug: 'klasifikasi-xray-pneumonia',
    kategori: 'Kesehatan',
    teknologi: 'Medical Imaging',
    judul: 'Klasifikasi X-Ray Pneumonia',
    deskripsi:
      'Mengklasifikasi citra rontgen dada untuk indikasi pneumonia menggunakan ResNet152V2 + SVM.',
    gambar: '/portofolio/xray-pneumonia.jpg',
    alt: 'Klasifikasi X-Ray pneumonia',
    lebar: 900,
    tinggi: 963,
  },
  {
    slug: 'deteksi-kesegaran-ikan',
    kategori: 'Pangan',
    teknologi: 'Mobile AI',
    judul: 'Deteksi Kesegaran Ikan',
    deskripsi:
      'Mengenali jenis ikan dan tingkat kesegarannya dari foto, langsung dari ponsel, beserta informasi nutrisi.',
    gambar: '/portofolio/kesegaran-ikan.jpg',
    alt: 'Deteksi kesegaran ikan',
    lebar: 620,
    tinggi: 1230,
  },
];
