/**
 * Implementasi repository dengan Drizzle ORM (PostgreSQL).
 * Semua fungsi tipis: query + mapper — tanpa logika bisnis.
 * Disuntikkan ke route lewat `Repositories`, jadi test route tidak menyentuh file ini.
 */
import { and, asc, desc, eq, ilike, isNotNull, or, sql } from 'drizzle-orm';
import type { Db } from '../db/client';
import { berita, dokumentasi, heroSlides, inference, klien, kursus, layanan, profil, progressPembelajaran, tim, testimoni } from '../db/schema';
import { toLikePattern } from '../lib/query';
import {
  toAnggotaTim,
  toBeritaItem,
  toDokumentasiItem,
  toHeroSlide,
  toKlien,
  toKontenInference,
  toKursus,
  toLayanan,
  toProfil,
  toProgresPeserta,
  toTestimoni,
} from './mappers';
import type {
  DokumentasiRepository,
  HeroRepository,
  InferenceRepository,
  KlienRepository,
  KursusRepository,
  LayananRepository,
  ProfilRepository,
  ProgressRepository,
  Repositories,
  TestimoniRepository,
  TimRepository,
  BeritaRepository,
} from './types';

/** Terapkan LIMIT/OFFSET bila pagination aktif. */
function applyPagination<T extends { limit(n: number): unknown; offset(n: number): unknown }>(
  query: T,
  pagination: { limit: number | null; offset: number | null },
): T {
  if (pagination.limit === null || pagination.offset === null) return query;
  // Kedua method mengembalikan rantai query baru pada Drizzle; panggil berurutan.
  const withLimit = query.limit(pagination.limit) as T;
  return withLimit.offset(pagination.offset) as T;
}

export function createRepositories(db: Db): Repositories {
  const beritaRepo: BeritaRepository = {
    async list({ q, pagination, order }) {
      const search =
        q === null
          ? undefined
          : or(
              ilike(berita.judul, toLikePattern(q)),
              ilike(berita.ringkasan, toLikePattern(q)),
              ilike(berita.penulis, toLikePattern(q)),
            );
      const base = db.select().from(berita).where(search);
      const ordered =
        order === 'asc'
          ? base.orderBy(asc(berita.tanggal), desc(berita.id))
          : base.orderBy(desc(berita.tanggal), desc(berita.id));
      const rows = await applyPagination(ordered, pagination);
      return rows.map(toBeritaItem);
    },
    async findBySlug(slug) {
      if (slug === '') return null;
      const rows = await db.select().from(berita).where(eq(berita.slug, slug)).limit(1);
      return rows[0] !== undefined ? toBeritaItem(rows[0]) : null;
    },
    async create(slug, data) {
      const rows = await db
        .insert(berita)
        .values({
          slug,
          judul: data.judul,
          ringkasan: data.ringkasan,
          isi: data.isi,
          tanggal: data.tanggal,
          penulis: data.penulis,
          gambar: data.gambar,
        })
        .returning();
      return toBeritaItem(rows[0]);
    },
    async update(slug, data) {
      const rows = await db
        .update(berita)
        .set({
          judul: data.judul,
          ringkasan: data.ringkasan,
          isi: data.isi,
          tanggal: data.tanggal,
          penulis: data.penulis,
          gambar: data.gambar,
        })
        .where(eq(berita.slug, slug))
        .returning();
      return rows[0] !== undefined ? toBeritaItem(rows[0]) : null;
    },
    async remove(slug) {
      const rows = await db
        .delete(berita)
        .where(eq(berita.slug, slug))
        .returning({ id: berita.id });
      return rows.length > 0;
    },
  };

  const dokumentasiRepo: DokumentasiRepository = {
    async list({ q, pagination, order, kategori }) {
      const conditions = [];
      if (q !== null) {
        conditions.push(
          or(
            ilike(dokumentasi.judul, toLikePattern(q)),
            ilike(dokumentasi.deskripsi, toLikePattern(q)),
            ilike(dokumentasi.kategori, toLikePattern(q)),
          ),
        );
      }
      if (kategori !== null) conditions.push(eq(dokumentasi.kategori, kategori));
      const base = db
        .select()
        .from(dokumentasi)
        .where(conditions.length > 0 ? and(...conditions.filter((c) => c !== undefined)) : undefined);
      const ordered =
        order === 'asc'
          ? base.orderBy(asc(dokumentasi.tanggal), desc(dokumentasi.id))
          : base.orderBy(desc(dokumentasi.tanggal), desc(dokumentasi.id));
      const rows = await applyPagination(ordered, pagination);
      return rows.map(toDokumentasiItem);
    },
    async findBySlug(slug) {
      if (slug === '') return null;
      const rows = await db
        .select()
        .from(dokumentasi)
        .where(eq(dokumentasi.slug, slug))
        .limit(1);
      return rows[0] !== undefined ? toDokumentasiItem(rows[0]) : null;
    },
    async create(slug, data) {
      const rows = await db
        .insert(dokumentasi)
        .values({
          slug,
          judul: data.judul,
          deskripsi: data.deskripsi,
          tanggal: data.tanggal,
          kategori: data.kategori,
          gambar: data.gambar,
        })
        .returning();
      return toDokumentasiItem(rows[0]);
    },
    async update(slug, data) {
      const rows = await db
        .update(dokumentasi)
        .set({
          judul: data.judul,
          deskripsi: data.deskripsi,
          tanggal: data.tanggal,
          kategori: data.kategori,
          gambar: data.gambar,
        })
        .where(eq(dokumentasi.slug, slug))
        .returning();
      return rows[0] !== undefined ? toDokumentasiItem(rows[0]) : null;
    },
    async remove(slug) {
      const rows = await db
        .delete(dokumentasi)
        .where(eq(dokumentasi.slug, slug))
        .returning({ id: dokumentasi.id });
      return rows.length > 0;
    },
  };

  const timRepo: TimRepository = {
    async list({ q, pagination }) {
      const search =
        q === null
          ? undefined
          : or(
              ilike(tim.nama, toLikePattern(q)),
              ilike(tim.peran, toLikePattern(q)),
              ilike(tim.kredensial, toLikePattern(q)),
            );
      const base = db.select().from(tim).where(search).orderBy(asc(tim.urutan), asc(tim.id));
      const rows = await applyPagination(base, pagination);
      return rows.map(toAnggotaTim);
    },
    async createAtas(data) {
      return db.transaction(async (tx) => {
        await tx.update(tim).set({ urutan: sql`${tim.urutan} + 1` });
        const rows = await tx
          .insert(tim)
          .values({
            nama: data.nama,
            peran: data.peran,
            kredensial: data.kredensial,
            foto: data.foto,
            urutan: data.urutan,
          })
          .returning();
        return toAnggotaTim(rows[0]);
      });
    },
    async update(id, data) {
      const rows = await db
        .update(tim)
        .set({
          nama: data.nama,
          peran: data.peran,
          kredensial: data.kredensial,
          foto: data.foto,
          urutan: data.urutan,
        })
        .where(eq(tim.id, id))
        .returning();
      return rows[0] !== undefined ? toAnggotaTim(rows[0]) : null;
    },
    async remove(id) {
      const rows = await db.delete(tim).where(eq(tim.id, id)).returning({ id: tim.id });
      return rows.length > 0;
    },
  };

  const layananRepo: LayananRepository = {
    async list() {
      const rows = await db.select().from(layanan).orderBy(asc(layanan.id));
      return rows.map(toLayanan);
    },
    async findBySlug(slug) {
      if (slug === '') return null;
      const rows = await db.select().from(layanan).where(eq(layanan.slug, slug)).limit(1);
      return rows[0] !== undefined ? toLayanan(rows[0]) : null;
    },
    async create(slug, data) {
      const rows = await db
        .insert(layanan)
        .values({
          slug,
          nama: data.nama,
          tagline: data.tagline,
          deskripsi: data.deskripsi,
          fitur: data.fitur,
        })
        .returning();
      return toLayanan(rows[0]);
    },
    async update(slug, data) {
      const rows = await db
        .update(layanan)
        .set({
          nama: data.nama,
          tagline: data.tagline,
          deskripsi: data.deskripsi,
          fitur: data.fitur,
        })
        .where(eq(layanan.slug, slug))
        .returning();
      return rows[0] !== undefined ? toLayanan(rows[0]) : null;
    },
    async remove(slug) {
      const rows = await db
        .delete(layanan)
        .where(eq(layanan.slug, slug))
        .returning({ id: layanan.id });
      return rows.length > 0;
    },
  };

  const heroRepo: HeroRepository = {
    async list() {
      const rows = await db.select().from(heroSlides).orderBy(asc(heroSlides.urutan), asc(heroSlides.id));
      return rows.map(toHeroSlide);
    },
    async create(data) {
      const rows = await db
        .insert(heroSlides)
        .values({
          urutan: data.urutan,
          eyebrow: data.eyebrow,
          judul: data.judul,
          judulAksen: data.judulAksen,
          sub: data.sub,
          ctaPrimer: data.ctaPrimer,
          ctaSekunder: data.ctaSekunder,
          badgeJudul: data.badgeJudul,
          badgeSub: data.badgeSub,
          image: data.image,
          srcSet: data.srcSet,
          sizes: data.sizes,
          layout: data.layout,
        })
        .returning();
      return toHeroSlide(rows[0]);
    },
    async update(id, data) {
      const rows = await db
        .update(heroSlides)
        .set({
          urutan: data.urutan,
          eyebrow: data.eyebrow,
          judul: data.judul,
          judulAksen: data.judulAksen,
          sub: data.sub,
          ctaPrimer: data.ctaPrimer,
          ctaSekunder: data.ctaSekunder,
          badgeJudul: data.badgeJudul,
          badgeSub: data.badgeSub,
          image: data.image,
          srcSet: data.srcSet,
          sizes: data.sizes,
          layout: data.layout,
        })
        .where(eq(heroSlides.id, id))
        .returning();
      return rows[0] !== undefined ? toHeroSlide(rows[0]) : null;
    },
    async remove(id) {
      const rows = await db.delete(heroSlides).where(eq(heroSlides.id, id)).returning({ id: heroSlides.id });
      return rows.length > 0;
    },
  };

  const klienRepo: KlienRepository = {
    async list() {
      const rows = await db.select().from(klien).orderBy(asc(klien.urutan), asc(klien.id));
      return rows.map(toKlien);
    },
  };

  const testimoniRepo: TestimoniRepository = {
    async list() {
      const rows = await db
        .select()
        .from(testimoni)
        .orderBy(asc(testimoni.urutan), asc(testimoni.id));
      return rows.map(toTestimoni);
    },
    async createAtas(data) {
      return db.transaction(async (tx) => {
        await tx.update(testimoni).set({ urutan: sql`${testimoni.urutan} + 1` });
        const rows = await tx
          .insert(testimoni)
          .values({ nama: data.nama, peran: data.peran, kutipan: data.kutipan, urutan: data.urutan })
          .returning();
        return toTestimoni(rows[0]);
      });
    },
    async update(id, data) {
      const rows = await db
        .update(testimoni)
        .set({ nama: data.nama, peran: data.peran, kutipan: data.kutipan, urutan: data.urutan })
        .where(eq(testimoni.id, id))
        .returning();
      return rows[0] !== undefined ? toTestimoni(rows[0]) : null;
    },
    async remove(id) {
      const rows = await db.delete(testimoni).where(eq(testimoni.id, id)).returning({ id: testimoni.id });
      return rows.length > 0;
    },
  };

  const profilRepo: ProfilRepository = {
    async get() {
      const rows = await db.select().from(profil).where(eq(profil.id, 1)).limit(1);
      return rows[0] !== undefined ? toProfil(rows[0]) : null;
    },
    async update(data) {
      // `statistik` tidak disertakan — kolom jsonb dipertahankan apa adanya.
      const rows = await db
        .update(profil)
        .set({
          nama: data.nama,
          tagline: data.tagline,
          ringkasan: data.ringkasan,
          alamat: data.alamat,
          email: data.email,
          telepon: data.telepon,
          visi: data.visi,
          misi: data.misi,
        })
        .where(eq(profil.id, 1))
        .returning();
      return rows[0] !== undefined ? toProfil(rows[0]) : null;
    },
  };

  const inferenceRepo: InferenceRepository = {
    async get() {
      const rows = await db.select().from(inference).where(eq(inference.id, 1)).limit(1);
      return rows[0] !== undefined ? toKontenInference(rows[0]) : null;
    },
    async update(data) {
      const rows = await db
        .update(inference)
        .set({
          judulApaItu: data.judulApaItu,
          deskripsiApaItu: data.deskripsiApaItu,
          kebutuhan: data.kebutuhan,
          alur: data.alur,
          contohIntro: data.contohIntro,
          contoh: data.contoh,
        })
        .where(eq(inference.id, 1))
        .returning();
      return rows[0] !== undefined ? toKontenInference(rows[0]) : null;
    },
  };

  const kursusRepo: KursusRepository = {
    async list({ q, pagination }) {
      const search =
        q === null
          ? undefined
          : or(
              ilike(kursus.judul, toLikePattern(q)),
              ilike(kursus.deskripsi, toLikePattern(q)),
              ilike(kursus.tentang, toLikePattern(q)),
              ilike(kursus.instruktur, toLikePattern(q)),
            );
      const base = db
        .select()
        .from(kursus)
        .where(search)
        .orderBy(asc(kursus.urutan), asc(kursus.id));
      const rows = await applyPagination(base, pagination);
      return rows.map(toKursus);
    },
    async findByKode(kode) {
      const bersih = kode.trim().toUpperCase();
      if (bersih === '') return null;
      // eq dengan nilai sudah di-uppercase = pencocokan case-insensitive
      // untuk kode tersimpan 'R01' (mirror frontend kursusDariKode).
      const rows = await db.select().from(kursus).where(eq(kursus.kode, bersih)).limit(1);
      return rows[0] !== undefined ? toKursus(rows[0]) : null;
    },
    async create(data) {
      const rows = await db
        .insert(kursus)
        .values({
          kode: data.kode,
          judul: data.judul,
          deskripsi: data.deskripsi,
          tentang: data.tentang,
          durasi: data.durasi,
          level: data.level,
          format: data.format,
          instruktur: data.instruktur,
          peran: data.peran,
          inisial: data.inisial,
          target: data.target,
          hasil: data.hasil,
          modul: data.modul,
        })
        .returning();
      return toKursus(rows[0]);
    },
    async update(kode, data) {
      const bersih = kode.trim().toUpperCase();
      if (bersih === '') return null;
      const rows = await db
        .update(kursus)
        .set({
          judul: data.judul,
          deskripsi: data.deskripsi,
          tentang: data.tentang,
          durasi: data.durasi,
          level: data.level,
          format: data.format,
          instruktur: data.instruktur,
          peran: data.peran,
          inisial: data.inisial,
          target: data.target,
          hasil: data.hasil,
          modul: data.modul,
        })
        .where(eq(kursus.kode, bersih))
        .returning();
      return rows[0] !== undefined ? toKursus(rows[0]) : null;
    },
    async remove(kode) {
      const bersih = kode.trim().toUpperCase();
      if (bersih === '') return false;
      const rows = await db
        .delete(kursus)
        .where(eq(kursus.kode, bersih))
        .returning({ id: kursus.id });
      return rows.length > 0;
    },
  };

  const progressRepo: ProgressRepository = {
    async list(pesertaId, kode) {
      const kondisi =
        kode === null
          ? eq(progressPembelajaran.pesertaId, pesertaId)
          : and(
              eq(progressPembelajaran.pesertaId, pesertaId),
              eq(progressPembelajaran.kode, kode),
            );
      const rows = await db
        .select()
        .from(progressPembelajaran)
        .where(kondisi)
        .orderBy(asc(progressPembelajaran.kode));
      return rows.map(toProgresPeserta);
    },
    async simpan(data) {
      // UPSERT per (peserta, kode): state lama diganti total + stempel waktu baru.
      const rows = await db
        .insert(progressPembelajaran)
        .values({ pesertaId: data.pesertaId, kode: data.kode, state: data.state })
        .onConflictDoUpdate({
          target: [progressPembelajaran.pesertaId, progressPembelajaran.kode],
          set: { state: data.state, diperbaruiPada: new Date() },
        })
        .returning();
      return toProgresPeserta(rows[0]);
    },
  };

  return {
    ping: async () => {
      // Check actual schema, not only network connectivity.
      await db.select({ id: berita.id }).from(berita).limit(1);
    },
    referensiGambar: async () => {
      // Empat select kecil (bukan union SQL) — tetap portabel di kedua driver
      // (PGlite dev & node-postgres produksi) dan tipenya aman.
      const [b, d, t, h] = await Promise.all([
        db.select({ v: berita.gambar }).from(berita).where(isNotNull(berita.gambar)),
        db.select({ v: dokumentasi.gambar }).from(dokumentasi).where(isNotNull(dokumentasi.gambar)),
        db.select({ v: tim.foto }).from(tim).where(isNotNull(tim.foto)),
        db.select({ image: heroSlides.image, srcSet: heroSlides.srcSet }).from(heroSlides),
      ]);
      const langsung = [...b, ...d, ...t]
        .map((r) => r.v)
        .filter((v): v is string => v !== null);
      const imageHero = h.map((r) => r.image).filter((v): v is string => v !== null);
      // Kolom `srcSet` berisi "url lebar, url lebar, …" — pecah per entri lalu
      // buang deskriptornya ("800w") supaya varian gambar hero tidak dianggap
      // file tidak terpakai lalu terhapus oleh pembersih.
      const srcSetHero = h.flatMap((r) =>
        (r.srcSet ?? '')
          .split(',')
          .map((entri) => entri.trim().split(/\s+/)[0] ?? '')
          .filter((url) => url !== ''),
      );
      return [...langsung, ...imageHero, ...srcSetHero];
    },
    berita: beritaRepo,
    dokumentasi: dokumentasiRepo,
    tim: timRepo,
    layanan: layananRepo,
    hero: heroRepo,
    klien: klienRepo,
    testimoni: testimoniRepo,
    profil: profilRepo,
    inference: inferenceRepo,
    kursus: kursusRepo,
    progress: progressRepo,
  };
}
