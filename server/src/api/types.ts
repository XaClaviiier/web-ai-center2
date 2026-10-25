/**
 * Tipe respons API CMS dan progres.
 * Tipe static-content di bawah tetap dipakai persistence legacy/seed,
 * tetapi tidak punya endpoint publik atau admin.
 */

/** ↔ src/data/berita.ts BeritaItem */
export interface BeritaItem {
  slug: string;
  judul: string;
  ringkasan: string;
  isi: string;
  /** Tanggal ISO (YYYY-MM-DD). */
  tanggal: string;
  gambar?: string;
  penulis: string;
}

/** ↔ src/data/dokumentasi.ts DokumentasiItem */
export interface DokumentasiItem {
  slug: string;
  judul: string;
  deskripsi: string;
  /** Tanggal ISO (YYYY-MM-DD). */
  tanggal: string;
  kategori: string;
  gambar?: string;
}

/**
 * ↔ src/data/tim.ts AnggotaTim. `id` & `urutan` hanya dipakai admin
 * (CRUD tab "Tim" di src/pages/Admin.tsx) — halaman publik mengabaikannya;
 * data dummy frontend tidak memilikinya (karena itu opsional di sana).
 */
export interface AnggotaTim {
  id: number;
  nama: string;
  peran: string;
  kredensial?: string;
  foto?: string;
  urutan: number;
}

/** ↔ src/data/layanan.ts Layanan */
export interface Layanan {
  slug: string;
  nama: string;
  tagline: string;
  deskripsi: string;
  fitur: string[];
}

/** ↔ src/data/hero.ts HeroSlide CTA */
export interface HeroCta {
  label: string;
  to: string;
}

/**
 * ↔ src/data/hero.ts HeroSlide. `id` & `urutan` hanya dipakai admin
 * (CRUD tab "Hero" di src/pages/Admin.tsx) — halaman publik mengabaikannya;
 * data dummy frontend tidak memilikinya (karena itu opsional di sana).
 */
export interface HeroSlide {
  id: number;
  /** Urutan tampil (asc) — slide pertama = 0. */
  urutan: number;
  eyebrow: string;
  judul: string;
  judulAksen: string;
  sub: string;
  ctaPrimer: HeroCta;
  ctaSekunder: HeroCta;
  badgeJudul: string;
  badgeSub: string;
  image?: string;
  srcSet?: string;
  sizes?: string;
  /** Layout tampil slide — ↔ src/data/hero.ts HeroSlide.layout. */
  layout?: 'default' | 'image-left' | 'teks-kanan';
}

/** ↔ src/data/klien.ts Klien */
export interface Klien {
  nama: string;
  bidang: string;
}

/**
 * ↔ src/data/testimoni.ts Testimoni. `id` & `urutan` hanya dipakai admin
 * (CRUD tab "Testimoni" di src/pages/Admin.tsx) — idem AnggotaTim/HeroSlide.
 */
export interface Testimoni {
  id: number;
  /** Urutan tampil (asc) — testimoni pertama = 0. */
  urutan: number;
  nama: string;
  peran: string;
  kutipan: string;
}

/** ↔ TentangKami + kontak footer SiteLayout (baris tunggal, id = 1). */
export interface Profil {
  nama: string;
  tagline: string;
  ringkasan: string;
  alamat: string;
  email: string;
  telepon: string;
  visi?: string;
  misi?: string;
  statistik?: Array<{ label: string; value: string }>;
}

/** Soal kuis per modul — ↔ src/data/pelatihan.ts Soal (`kunci` = index opsi benar). */
export interface SoalKursus {
  pertanyaan: string;
  opsi: string[];
  kunci: number;
}

/** ↔ src/data/pelatihan.ts Modul (kursus) — `video` & `quiz` opsional (baris lama). */
export interface ModulKursus {
  judul: string;
  deskripsi: string;
  meta: string;
  /** URL video lesson (opsional); tampilan frontend meng-embed bila didukung. */
  video?: string | null;
  /** Kuis per modul (opsional); kosong → checkpoint dummy di frontend. */
  quiz?: SoalKursus[];
}

/** ↔ src/data/pelatihan.ts Kursus */
export interface Kursus {
  /** Kode kursus, mis. 'R01' (unik, disimpan huruf besar). */
  kode: string;
  /** Label audiens untuk filter katalog, mis. ['Mahasiswa', 'Dosen']. */
  target: string[];
  judul: string;
  deskripsi: string;
  tentang: string;
  durasi: string;
  level: string;
  format: string;
  instruktur: string;
  peran: string;
  inisial: string;
  hasil: string[];
  modul: ModulKursus[];
}

/** ↔ src/data/inference.ts LangkahInference (alur kerja 5 langkah). */
export interface LangkahInference {
  nomor: string;
  judul: string;
  deskripsi: string;
}

/** ↔ src/data/inference.ts contohInference (chip "Contoh penerapan"). */
export interface ContohPenerapan {
  slug: string;
  judul: string;
}

/**
 * ↔ src/data/inference.ts KontenInference — konten halaman Inference Solution
 * (baris tunggal id = 1, dikelola lewat tab "Inference" di /admin).
 */
export interface KontenInference {
  judulApaItu: string;
  deskripsiApaItu: string;
  kebutuhan: string[];
  alur: LangkahInference[];
  contohIntro: string;
  contoh: ContohPenerapan[];
}

/** Bentuk standar daftar — { items: [...] } sesuai komentar TODO_BACKEND. */
export interface ListResponse<T> {
  items: T[];
}

/**
 * Progres belajar satu kursus milik satu peserta (GET/POST /api/progress).
 * `state` = JSON mentah StateBelajar frontend; `diperbaruiPada` ISO string.
 */
export interface ProgresPeserta {
  kode: string;
  state: Record<string, unknown>;
  diperbaruiPada: string;
}
