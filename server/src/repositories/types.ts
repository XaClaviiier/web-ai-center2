/**
 * Interface repository — kontrak antara route (Elysia) dan lapisan DB.
 * Route hanya tahu interface ini; implementasi Drizzle disuntikkan di app.ts
 * sehingga unit test route bisa memakai fake repository tanpa DB.
 */
import type {
  AnggotaTim,
  BeritaItem,
  DokumentasiItem,
  HeroSlide,
  Klien,
  Kursus,
  KontenInference,
  Layanan,
  Profil,
  ProgresPeserta,
  Testimoni,
} from '../api/types';
import type { Pagination } from '../lib/query';
// Input tulis admin (PROGRESS 2 + Prioritas 2) hidup di lib/tulis — satu
// sumber kontrak validasi body, dipakai route & repository.
import type {
  InputBerita,
  InputDokumentasi,
  InputHero,
  InputInference,
  InputKursus,
  InputLayanan,
  InputProfil,
  InputProgress,
  InputTestimoni,
  InputTim,
} from '../lib/tulis';

/** Alias agar interface repository tidak perlu import langsung dari lib/tulis. */
export type BeritaInput = InputBerita;
export type DokumentasiInput = InputDokumentasi;
export type KursusInput = InputKursus;
export type LayananInput = InputLayanan;
export type ProfilInput = InputProfil;
export type TimInput = InputTim;
export type InferenceInput = InputInference;
export type TestimoniInput = InputTestimoni;
export type HeroInput = InputHero;
export type ProgressInput = InputProgress;

export interface ListParams {
  /** Kata kunci pencarian; `null` = tanpa filter. */
  q: string | null;
  /** Hasil `parsePagination`. */
  pagination: Pagination;
  /** Urutan tanggal; hanya untuk berita/dokumentasi. */
  order?: 'asc' | 'desc';
}

export interface DokumentasiListParams extends ListParams {
  /** Filter kategori; `null` = semua. */
  kategori: string | null;
}

export interface BeritaRepository {
  list(params: ListParams): Promise<BeritaItem[]>;
  findBySlug(slug: string): Promise<BeritaItem | null>;
  /** Buat item baru (slug unik sudah dipastikan route); balikan item tersimpan. */
  create(slug: string, data: BeritaInput): Promise<BeritaItem>;
  /** Perbarui item berdasarkan slug; `null` bila slug tidak ditemukan. */
  update(slug: string, data: BeritaInput): Promise<BeritaItem | null>;
  /** Hapus item berdasarkan slug; `true` bila ada baris yang terhapus. */
  remove(slug: string): Promise<boolean>;
}

export interface DokumentasiRepository {
  list(params: DokumentasiListParams): Promise<DokumentasiItem[]>;
  findBySlug(slug: string): Promise<DokumentasiItem | null>;
  /** Buat item baru (slug unik sudah dipastikan route); balikan item tersimpan. */
  create(slug: string, data: DokumentasiInput): Promise<DokumentasiItem>;
  /** Perbarui item berdasarkan slug; `null` bila slug tidak ditemukan. */
  update(slug: string, data: DokumentasiInput): Promise<DokumentasiItem | null>;
  /** Hapus item berdasarkan slug; `true` bila ada baris yang terhapus. */
  remove(slug: string): Promise<boolean>;
}

export interface TimRepository {
  list(params: ListParams): Promise<AnggotaTim[]>;
  /** Geser urutan dan simpan anggota baru sebagai item paling atas atomik. */
  createAtas(data: TimInput): Promise<AnggotaTim>;
  /** Perbarui anggota berdasarkan id; `null` bila id tidak ditemukan. */
  update(id: number, data: TimInput): Promise<AnggotaTim | null>;
  /** Hapus anggota berdasarkan id; `true` bila ada baris yang terhapus. */
  remove(id: number): Promise<boolean>;
}

export interface LayananRepository {
  list(): Promise<Layanan[]>;
  findBySlug(slug: string): Promise<Layanan | null>;
  /** Buat layanan baru (slug unik sudah dipastikan route); balikan item tersimpan. */
  create(slug: string, data: LayananInput): Promise<Layanan>;
  /** Perbarui layanan berdasarkan slug; `null` bila slug tidak ditemukan. */
  update(slug: string, data: LayananInput): Promise<Layanan | null>;
  /** Hapus layanan berdasarkan slug; `true` bila ada baris yang terhapus. */
  remove(slug: string): Promise<boolean>;
}

export interface HeroRepository {
  list(): Promise<HeroSlide[]>;
  /** Buat slide baru; balikan slide tersimpan (id & urutan dari DB). */
  create(data: HeroInput): Promise<HeroSlide>;
  /** Perbarui slide berdasarkan id; `null` bila id tidak ditemukan. */
  update(id: number, data: HeroInput): Promise<HeroSlide | null>;
  /** Hapus slide berdasarkan id; `true` bila ada baris yang terhapus. */
  remove(id: number): Promise<boolean>;
}

export interface KlienRepository {
  list(): Promise<Klien[]>;
}

export interface TestimoniRepository {
  list(): Promise<Testimoni[]>;
  /** Geser urutan dan simpan testimoni baru sebagai item paling atas atomik. */
  createAtas(data: TestimoniInput): Promise<Testimoni>;
  /** Perbarui testimoni berdasarkan id; `null` bila id tidak ditemukan. */
  update(id: number, data: TestimoniInput): Promise<Testimoni | null>;
  /** Hapus testimoni berdasarkan id; `true` bila ada baris yang terhapus. */
  remove(id: number): Promise<boolean>;
}

export interface ProfilRepository {
  /** Profil baris tunggal (id = 1); `null` bila belum di-seed. */
  get(): Promise<Profil | null>;
  /**
   * Perbarui profil baris tunggal; `null` bila baris belum ada (belum di-seed).
   * Kolom `statistik` TIDAK disentuh (tidak ada input edit untuknya).
   */
  update(data: ProfilInput): Promise<Profil | null>;
}

export interface InferenceRepository {
  /** Konten halaman inference (baris tunggal id = 1); `null` bila belum di-seed. */
  get(): Promise<KontenInference | null>;
  /** Perbarui konten inference; `null` bila baris belum ada (belum di-seed). */
  update(data: InferenceInput): Promise<KontenInference | null>;
}

export interface KursusRepository {
  list(params: ListParams): Promise<Kursus[]>;
  /** Detail kursus; kode case-insensitive (URL lama memakai 'r01'). */
  findByKode(kode: string): Promise<Kursus | null>;
  /** Buat kursus (kode unik sudah dipastikan route). */
  create(data: KursusInput): Promise<Kursus>;
  /** Perbarui kursus berdasarkan kode; `null` bila tidak ditemukan. */
  update(kode: string, data: KursusInput): Promise<Kursus | null>;
  /** Hapus kursus berdasarkan kode; `true` bila ada baris yang terhapus. */
  remove(kode: string): Promise<boolean>;
}

/**
 * Progres belajar peserta (LMS tanpa akun): satu baris per (peserta, kursus).
 * `state` disimpan apa adanya (blob JSON frontend) — normalisasi saat baca
 * terjadi di `bacaStateBelajar` (src/lib/pembelajaran.ts).
 */
export interface ProgressRepository {
  /** Daftar progres milik satu peserta; `kode` null = semua kursusnya. */
  list(pesertaId: string, kode: string | null): Promise<ProgresPeserta[]>;
  /** Simpan (UPSERT per peserta+kode); balik entri tersimpan + stempel waktu. */
  simpan(data: ProgressInput): Promise<ProgresPeserta>;
}

/** Kumpulan semua repository yang disuntikkan ke route. */
export interface Repositories {
  berita: BeritaRepository;
  dokumentasi: DokumentasiRepository;
  tim: TimRepository;
  layanan: LayananRepository;
  hero: HeroRepository;
  klien: KlienRepository;
  testimoni: TestimoniRepository;
  profil: ProfilRepository;
  inference: InferenceRepository;
  kursus: KursusRepository;
  progress: ProgressRepository;
  /** Cek koneksi DB untuk `/api/health`; lempar bila DB tidak terjangkau. */
  ping: () => Promise<void>;
  /**
   * Semua nilai URL gambar/foto yang dirujuk konten (berita.gambar,
   * dokumentasi.gambar, tim.foto, hero.image/srcSet) — pembanding file
   * unggahan tidak terpakai (lihat lib/orphanUpload.ts + route
   * POST /api/admin/uploads/bersihkan).
   */
  referensiGambar: () => Promise<string[]>;
}
