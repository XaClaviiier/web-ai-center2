/**
 * Skema database (Drizzle ORM) — target PostgreSQL.
 * Nama kolom snake_case; nama properti TS camelCase mengikuti interface frontend.
 */
import { date, integer, jsonb, pgTable, primaryKey, serial, text, timestamp } from 'drizzle-orm/pg-core';

/** CTA hero: { label, to } — mengikuti interface HeroSlide di frontend. */
export interface HeroCta {
  label: string;
  to: string;
}

/** Soal kuis per modul — ↔ src/data/pelatihan.ts Soal (`kunci` = index opsi benar). */
export interface SoalKursus {
  pertanyaan: string;
  opsi: string[];
  kunci: number;
}

/**
 * Elemen modul kursus (jsonb) — ↔ src/data/pelatihan.ts Modul.
 * `video` & `quiz` opsional (baris lama dari seed tidak memilikinya);
 * tulis via admin selalu menyimpan keduanya (null / [] bila kosong).
 */
export interface ModulKursus {
  judul: string;
  deskripsi: string;
  meta: string;
  /** URL video lesson (opsional); tampilan frontend meng-embed bila didukung. */
  video?: string | null;
  /** Kuis per modul (opsional); kosong → frontend memakai checkpoint dummy. */
  quiz?: SoalKursus[];
}

/** Item statistik "Tentang Kami": { label, value }. */
export interface StatistikItem {
  label: string;
  value: string;
}

/** Berita — interface frontend: BeritaItem. */
export const berita = pgTable('berita', {
  id: serial('id').primaryKey(),
  slug: text('slug').notNull().unique(),
  judul: text('judul').notNull(),
  ringkasan: text('ringkasan').notNull(),
  isi: text('isi').notNull(),
  tanggal: date('tanggal', { mode: 'string' }).notNull(),
  gambar: text('gambar'),
  penulis: text('penulis').notNull(),
  createdAt: timestamp('created_at', { mode: 'date' }).notNull().defaultNow(),
});

/** Dokumentasi — interface frontend: DokumentasiItem. */
export const dokumentasi = pgTable('dokumentasi', {
  id: serial('id').primaryKey(),
  slug: text('slug').notNull().unique(),
  judul: text('judul').notNull(),
  deskripsi: text('deskripsi').notNull(),
  tanggal: date('tanggal', { mode: 'string' }).notNull(),
  kategori: text('kategori').notNull(),
  gambar: text('gambar'),
  createdAt: timestamp('created_at', { mode: 'date' }).notNull().defaultNow(),
});

/** Anggota tim — interface frontend: AnggotaTim (+ urutan untuk CMS). */
export const tim = pgTable('tim', {
  id: serial('id').primaryKey(),
  nama: text('nama').notNull(),
  peran: text('peran').notNull(),
  kredensial: text('kredensial'),
  foto: text('foto'),
  urutan: integer('urutan').notNull().default(0),
});

/** Layanan — interface frontend: Layanan (fitur = string[]). */
export const layanan = pgTable('layanan', {
  id: serial('id').primaryKey(),
  slug: text('slug').notNull().unique(),
  nama: text('nama').notNull(),
  tagline: text('tagline').notNull(),
  deskripsi: text('deskripsi').notNull(),
  fitur: jsonb('fitur').$type<string[]>().notNull(),
});

/** Slide hero — interface frontend: HeroSlide (CTA sebagai JSON). */
export const heroSlides = pgTable('hero_slides', {
  id: serial('id').primaryKey(),
  urutan: integer('urutan').notNull().default(0),
  eyebrow: text('eyebrow').notNull(),
  judul: text('judul').notNull(),
  judulAksen: text('judul_aksen').notNull(),
  sub: text('sub').notNull(),
  ctaPrimer: jsonb('cta_primer').$type<HeroCta>().notNull(),
  ctaSekunder: jsonb('cta_sekunder').$type<HeroCta>().notNull(),
  badgeJudul: text('badge_judul').notNull(),
  badgeSub: text('badge_sub').notNull(),
  image: text('image'),
  srcSet: text('src_set'),
  sizes: text('sizes'),
  /** Layout tampil: 'default' (bg penuh) / 'image-left' (gambar kiri, teks kanan) — ↔ hero.ts */
  layout: text('layout').$type<'default' | 'image-left' | 'teks-kanan'>(),
});

/** Klien — interface frontend: Klien. */
export const klien = pgTable('klien', {
  id: serial('id').primaryKey(),
  nama: text('nama').notNull(),
  bidang: text('bidang').notNull(),
  urutan: integer('urutan').notNull().default(0),
});

/** Testimoni — interface frontend: Testimoni. */
export const testimoni = pgTable('testimoni', {
  id: serial('id').primaryKey(),
  nama: text('nama').notNull(),
  peran: text('peran').notNull(),
  kutipan: text('kutipan').notNull(),
  urutan: integer('urutan').notNull().default(0),
});

/**
 * Profil "Tentang Kami" + kontak footer — baris tunggal (id = 1).
 * Nama, alamat, email, telepon, tagline footer, ringkasan/visi/misi/statistik.
 */
export const profil = pgTable('profil', {
  id: integer('id').primaryKey(),
  nama: text('nama').notNull(),
  tagline: text('tagline').notNull(),
  ringkasan: text('ringkasan').notNull(),
  alamat: text('alamat').notNull(),
  email: text('email').notNull(),
  telepon: text('telepon').notNull(),
  visi: text('visi'),
  misi: text('misi'),
  statistik: jsonb('statistik').$type<StatistikItem[]>(),
});

/** Satu langkah alur inference (jsonb) — ↔ src/data/inference.ts LangkahInference. */
export interface LangkahInference {
  nomor: string;
  judul: string;
  deskripsi: string;
}

/** Contoh penerapan inference (jsonb) — ↔ src/data/inference.ts contohInference. */
export interface ContohPenerapan {
  slug: string;
  judul: string;
}

/**
 * Legacy table retained for old data compatibility.
 * Static inference content now lives in `src/data/inference.ts`;
 * no route or seed writes this table.
 */
export const inference = pgTable('inference', {
  id: integer('id').primaryKey(),
  judulApaItu: text('judul_apa_itu').notNull(),
  deskripsiApaItu: text('deskripsi_apa_itu').notNull(),
  kebutuhan: jsonb('kebutuhan').$type<string[]>().notNull(),
  alur: jsonb('alur').$type<LangkahInference[]>().notNull(),
  contohIntro: text('contoh_intro').notNull(),
  contoh: jsonb('contoh').$type<ContohPenerapan[]>().notNull(),
});

/**
 * Kursus (Pelatihan) — interface frontend: Kursus.
 * `target`, `hasil`, `modul` bertipe jsonb (array); `urutan` menentukan
 * urutan tampil katalog, bukan bagian kontrak API.
 */
export const kursus = pgTable('kursus', {
  id: serial('id').primaryKey(),
  kode: text('kode').notNull().unique(),
  target: jsonb('target').$type<string[]>().notNull(),
  judul: text('judul').notNull(),
  deskripsi: text('deskripsi').notNull(),
  tentang: text('tentang').notNull(),
  durasi: text('durasi').notNull(),
  level: text('level').notNull(),
  format: text('format').notNull(),
  instruktur: text('instruktur').notNull(),
  peran: text('peran').notNull(),
  inisial: text('inisial').notNull(),
  hasil: jsonb('hasil').$type<string[]>().notNull(),
  modul: jsonb('modul').$type<ModulKursus[]>().notNull(),
  urutan: integer('urutan').notNull().default(0),
});

/**
 * Progres belajar peserta (LMS) TANPA akun — identitas = `pesertaId`
 * (UUID acak yang disimpan browser di localStorage). Satu baris per
 * (peserta, kursus) → PRIMARY KEY gabungan. `state` berisi JSON mentah
 * StateBelajar dari frontend (dinormalisasi ulang saat dibaca — karena itu
 * tidak divalidasi field-per-field, cukup objek + batas ukuran di lib/tulis).
 */
export const progressPembelajaran = pgTable(
  'progress_pembelajaran',
  {
    pesertaId: text('peserta_id').notNull(),
    kode: text('kode').notNull(),
    state: jsonb('state').$type<Record<string, unknown>>().notNull(),
    diperbaruiPada: timestamp('diperbarui_pada', { mode: 'date' }).notNull().defaultNow(),
  },
  // PRIMARY KEY gabungan: progres tersimpan sekali per (peserta, kursus) —
  // upsert route POST /api/progress memakai konflik dua kolom ini.
  (table) => ({ pk: primaryKey({ columns: [table.pesertaId, table.kode] }) }),
);
