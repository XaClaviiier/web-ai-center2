/**
 * Unit test route via app.handle(Request) dengan fake repository
 * (tanpa DB, tanpa listen jaringan) — input query → status + body benar.
 */
import { describe, expect, test } from 'bun:test';
import { createApp } from './app';
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
} from './api/types';
import type { ListParams, Repositories, DokumentasiListParams } from './repositories/types';
import type { Pagination } from './lib/query';
import type { InputBerita as BeritaInput, InputDokumentasi as DokumentasiInput, InputHero as HeroInput, InputInference as InferenceInput, InputKursus as KursusInput, InputLayanan as LayananInput, InputProfil as ProfilInput, InputProgress as ProgressInput, InputTestimoni as TestimoniInput, InputTim as TimInput } from './lib/tulis';
import { buatTokenAdmin } from './auth';

const NO_PAGINATION: Pagination = { page: null, limit: null, offset: null };

const sampleBerita: BeritaItem = {
  slug: 'berita-1',
  judul: 'Judul',
  ringkasan: 'Ringkasan',
  isi: 'Isi',
  tanggal: '2026-08-01',
  penulis: 'Humas',
};
const sampleDok: DokumentasiItem = {
  slug: 'dok-1',
  judul: 'Workshop',
  deskripsi: 'Deskripsi',
  tanggal: '2026-07-15',
  kategori: 'Workshop',
};
const sampleTim: AnggotaTim = { id: 1, nama: 'Nama', peran: 'Kepala', urutan: 0 };
const sampleLayanan: Layanan = {
  slug: 'pelatihan',
  nama: 'Pelatihan',
  tagline: 'Tagline',
  deskripsi: 'Deskripsi',
  fitur: ['a'],
};
const sampleHero: HeroSlide = {
  id: 10,
  urutan: 0,
  eyebrow: 'E',
  judul: 'J',
  judulAksen: 'A',
  sub: 'S',
  ctaPrimer: { label: 'L', to: '/x' },
  ctaSekunder: { label: 'L2', to: '/y' },
  badgeJudul: 'B',
  badgeSub: 'BS',
};
const sampleKlien: Klien = { nama: 'PT X', bidang: 'Teknologi' };
const sampleTestimoni: Testimoni = { id: 7, urutan: 0, nama: 'A', peran: 'B', kutipan: 'C' };
/** Konversi InputHero (boleh null) → HeroSlide respons; null di-omitted (meniru mapper). */
function heroDariInput(data: HeroInput, id: number): HeroSlide {
  return {
    id,
    urutan: data.urutan,
    eyebrow: data.eyebrow,
    judul: data.judul,
    judulAksen: data.judulAksen,
    sub: data.sub,
    ctaPrimer: data.ctaPrimer,
    ctaSekunder: data.ctaSekunder,
    badgeJudul: data.badgeJudul,
    badgeSub: data.badgeSub,
    ...(data.image !== null ? { image: data.image } : {}),
    ...(data.srcSet !== null ? { srcSet: data.srcSet } : {}),
    ...(data.sizes !== null ? { sizes: data.sizes } : {}),
    ...(data.layout !== null ? { layout: data.layout } : {}),
  };
}
const sampleProfil: Profil = {
  nama: 'AI Center Ubaya',
  tagline: 'Tagline',
  ringkasan: 'Ringkasan',
  alamat: 'Jl. Contoh',
  email: 'info@example.test',
  telepon: '+62312981000',
};
/** Fixture konten halaman inference (baris tunggal — GET/PUT /api/inference). */
const sampleInference: KontenInference = {
  judulApaItu: 'Apa itu Inference Solution?',
  deskripsiApaItu: 'Deskripsi inference.',
  kebutuhan: ['Kebutuhan 1', 'Kebutuhan 2'],
  alur: [
    { nomor: '01', judul: 'Konsultasi', deskripsi: 'Memetakan kebutuhan.' },
    { nomor: '02', judul: 'Deployment', deskripsi: 'Menjalankan model sebagai API.' },
  ],
  contohIntro: 'Contoh penerapan:',
  contoh: [{ slug: 'algae-finder', judul: 'Algae Finder' }],
};

/** Body valid untuk PUT /api/inference (bentuk InputInference). */
const BODY_INFERENCE: InferenceInput = {
  judulApaItu: 'Apa itu Inference Solution?',
  deskripsiApaItu: 'Deskripsi inference.',
  kebutuhan: ['Kebutuhan 1', 'Kebutuhan 2'],
  alur: [{ nomor: '01', judul: 'Konsultasi', deskripsi: 'Memetakan kebutuhan.' }],
  contohIntro: 'Contoh penerapan:',
  contoh: [{ slug: 'algae-finder', judul: 'Algae Finder' }],
};

const sampleKursus: Kursus = {
  kode: 'R01',
  target: ['Mahasiswa'],
  judul: 'Judul Kursus',
  deskripsi: 'Deskripsi',
  tentang: 'Tentang',
  durasi: '4 sesi',
  level: 'Pemula',
  format: 'Online mandiri',
  instruktur: 'Tim',
  peran: 'Pengajar',
  inisial: 'RA',
  hasil: ['Hasil'],
  modul: [{ judul: 'Modul', deskripsi: 'Desc', meta: '4 video · 35 menit' }],
};

/** Fixture progres belajar (GET/POST /api/progress). */
const sampleProgres: ProgresPeserta = {
  kode: 'R01',
  state: { enrolled: true, completed: [true, false] },
  diperbaruiPada: '2026-10-15T08:00:00.000Z',
};

/** Fake repository: mencatat argumen tiap panggilan, balikan di-hardcode. */
interface FakeHandle {
  repos: Repositories;
  calls: Array<{ method: string; args: unknown[] }>;
}

function createFakeRepos(): FakeHandle {
  const calls: Array<{ method: string; args: unknown[] }> = [];
  const rec = <A extends unknown[], R>(method: string, impl: (...args: A) => R) =>
    (...args: A): R => {
      calls.push({ method, args });
      return impl(...args);
    };

  const repos: Repositories = {
    ping: rec('ping', async (): Promise<void> => undefined),
    referensiGambar: rec('referensiGambar', async (): Promise<string[]> => ['/uploads/pakai.jpg']),
    berita: {
      list: rec('berita.list', async (_params: ListParams): Promise<BeritaItem[]> => [sampleBerita]),
      findBySlug: rec('berita.findBySlug', async (slug: string): Promise<BeritaItem | null> =>
        slug === sampleBerita.slug ? sampleBerita : null),
      create: rec('berita.create', async (slug: string, data: BeritaInput): Promise<BeritaItem> =>
        ({ slug, ...data, gambar: data.gambar ?? undefined })),
      update: rec('berita.update', async (slug: string, data: BeritaInput): Promise<BeritaItem | null> =>
        slug === sampleBerita.slug ? { slug, ...data, gambar: data.gambar ?? undefined } : null),
      remove: rec('berita.remove', async (slug: string): Promise<boolean> =>
        slug === sampleBerita.slug),
    },
    dokumentasi: {
      list: rec('dokumentasi.list', async (_params: DokumentasiListParams): Promise<DokumentasiItem[]> => [sampleDok]),
      findBySlug: rec('dokumentasi.findBySlug', async (slug: string): Promise<DokumentasiItem | null> =>
        slug === sampleDok.slug ? sampleDok : null),
      create: rec('dokumentasi.create', async (slug: string, data: DokumentasiInput): Promise<DokumentasiItem> =>
        ({ slug, ...data, gambar: data.gambar ?? undefined })),
      update: rec('dokumentasi.update', async (slug: string, data: DokumentasiInput): Promise<DokumentasiItem | null> =>
        slug === sampleDok.slug ? { slug, ...data, gambar: data.gambar ?? undefined } : null),
      remove: rec('dokumentasi.remove', async (slug: string): Promise<boolean> =>
        slug === sampleDok.slug),
    },
    tim: {
      list: rec('tim.list', async (_params: ListParams): Promise<AnggotaTim[]> => [sampleTim]),
      naikkanSemuaUrutan: rec('tim.naikkanSemuaUrutan', async (): Promise<void> => undefined),
      create: rec('tim.create', async (data: TimInput): Promise<AnggotaTim> => ({
        id: 9,
        nama: data.nama,
        peran: data.peran,
        urutan: data.urutan,
        ...(data.kredensial !== null ? { kredensial: data.kredensial } : {}),
        ...(data.foto !== null ? { foto: data.foto } : {}),
      })),
      update: rec('tim.update', async (id: number, data: TimInput): Promise<AnggotaTim | null> =>
        id === sampleTim.id
          ? {
              id,
              nama: data.nama,
              peran: data.peran,
              urutan: data.urutan,
              ...(data.kredensial !== null ? { kredensial: data.kredensial } : {}),
              ...(data.foto !== null ? { foto: data.foto } : {}),
            }
          : null),
      remove: rec('tim.remove', async (id: number): Promise<boolean> => id === sampleTim.id),
    },
    layanan: {
      list: rec('layanan.list', async (): Promise<Layanan[]> => [sampleLayanan]),
      findBySlug: rec('layanan.findBySlug', async (slug: string): Promise<Layanan | null> =>
        slug === sampleLayanan.slug ? sampleLayanan : null),
      create: rec('layanan.create', async (slug: string, data: LayananInput): Promise<Layanan> =>
        ({ slug, ...data })),
      update: rec('layanan.update', async (slug: string, data: LayananInput): Promise<Layanan | null> =>
        slug === sampleLayanan.slug ? { slug, ...data } : null),
      remove: rec('layanan.remove', async (slug: string): Promise<boolean> =>
        slug === sampleLayanan.slug),
    },
    hero: {
      list: rec('hero.list', async (): Promise<HeroSlide[]> => [sampleHero]),
      create: rec('hero.create', async (data: HeroInput): Promise<HeroSlide> =>
        heroDariInput(data, 99)),
      update: rec('hero.update', async (id: number, data: HeroInput): Promise<HeroSlide | null> =>
        id === sampleHero.id ? heroDariInput(data, id) : null),
      remove: rec('hero.remove', async (id: number): Promise<boolean> => id === sampleHero.id),
    },
    klien: { list: rec('klien.list', async (): Promise<Klien[]> => [sampleKlien]) },
    testimoni: {
      list: rec('testimoni.list', async (): Promise<Testimoni[]> => [sampleTestimoni]),
      naikkanSemuaUrutan: rec('testimoni.naikkanSemuaUrutan', async (): Promise<void> => undefined),
      create: rec('testimoni.create', async (data: TestimoniInput): Promise<Testimoni> => ({
        ...sampleTestimoni,
        id: 99,
        ...data,
      })),
      update: rec(
        'testimoni.update',
        async (id: number, data: TestimoniInput): Promise<Testimoni | null> =>
          id === sampleTestimoni.id ? { ...sampleTestimoni, ...data } : null,
      ),
      remove: rec('testimoni.remove', async (id: number): Promise<boolean> => id === sampleTestimoni.id),
    },
    profil: {
      get: rec('profil.get', async (): Promise<Profil | null> => sampleProfil),
      update: rec('profil.update', async (data: ProfilInput): Promise<Profil | null> => ({
        nama: data.nama,
        tagline: data.tagline,
        ringkasan: data.ringkasan,
        alamat: data.alamat,
        email: data.email,
        telepon: data.telepon,
        ...(data.visi !== null ? { visi: data.visi } : {}),
        ...(data.misi !== null ? { misi: data.misi } : {}),
      })),
    },
    inference: {
      get: rec('inference.get', async (): Promise<KontenInference | null> => sampleInference),
      update: rec('inference.update', async (data: InferenceInput): Promise<KontenInference | null> => data),
    },
    kursus: {
      list: rec('kursus.list', async (_params: ListParams): Promise<Kursus[]> => [sampleKursus]),
      findByKode: rec('kursus.findByKode', async (kode: string): Promise<Kursus | null> =>
        kode.toUpperCase() === sampleKursus.kode ? sampleKursus : null),
      create: rec('kursus.create', async (data: KursusInput): Promise<Kursus> =>
        ({ ...sampleKursus, ...data })),
      update: rec('kursus.update', async (kode: string, data: KursusInput): Promise<Kursus | null> =>
        kode.toUpperCase() === sampleKursus.kode ? { ...sampleKursus, ...data } : null),
      remove: rec('kursus.remove', async (kode: string): Promise<boolean> =>
        kode.toUpperCase() === sampleKursus.kode),
    },
    progress: {
      list: rec('progress.list', async (_pesertaId: string, kode: string | null): Promise<ProgresPeserta[]> =>
        // Fake: hanya punya entri R01 — filter kode lain → kosong.
        kode === null || kode === 'R01' ? [sampleProgres] : []),
      simpan: rec('progress.simpan', async (data: ProgressInput): Promise<ProgresPeserta> => ({
        kode: data.kode,
        state: data.state,
        diperbaruiPada: sampleProgres.diperbaruiPada,
      })),
    },
  };
  return { repos, calls };
}

const CORS_ORIGINS = ['http://localhost:5173'];
/** Password admin khusus test — token di-buat langsung dengan `buatTokenAdmin`. */
const ADMIN_PASSWORD_TEST = 'rahasia-admin-test';
const TOKEN_ADMIN = buatTokenAdmin(ADMIN_PASSWORD_TEST);

function createTestApp(handle = createFakeRepos()) {
  const app = createApp({
    repos: handle.repos,
    corsOrigins: CORS_ORIGINS,
    adminPassword: ADMIN_PASSWORD_TEST,
  });
  return { app, calls: handle.calls };
}

const getJson = async (app: ReturnType<typeof createApp>, path: string) => {
  const res = await app.handle(new Request(`http://localhost${path}`));
  return { status: res.status, body: (await res.json()) as unknown, res };
};

describe('GET /api/health', () => {
  test('200 { status: "ok", db: "ok" } — ping DB sukses', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await getJson(app, '/api/health');
    expect(status).toBe(200);
    expect(body).toEqual({ status: 'ok', db: 'ok' });
    expect(calls.some((c) => c.method === 'ping')).toBe(true);
  });

  test('503 { error } — ping DB gagal (DB tidak terjangkau)', async () => {
    const handle = createFakeRepos();
    handle.repos.ping = async () => {
      throw new Error('koneksi putus');
    };
    const { app } = createTestApp(handle);
    const { status, body } = await getJson(app, '/api/health');
    expect(status).toBe(503);
    expect(body).toEqual({ error: 'Database tidak terjangkau' });
  });
});

describe('GET /api/berita', () => {
  test('tanpa query → { items } dan params default (q null, tanpa pagination)', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await getJson(app, '/api/berita');
    expect(status).toBe(200);
    expect(body).toEqual({ items: [sampleBerita] });
    expect(calls).toEqual([
      {
        method: 'berita.list',
        args: [{ q: null, pagination: NO_PAGINATION, order: 'desc' }],
      },
    ]);
  });

  test('query lengkap dinormalisasi & diteruskan ke repository', async () => {
    const { app, calls } = createTestApp();
    const { body } = await getJson(
      app,
      '/api/berita?q=' + encodeURIComponent('  gpu   rental ') + '&page=2&limit=3&order=asc',
    );
    expect(body).toEqual({ items: [sampleBerita] });
    expect(calls[0].args[0]).toEqual({
      q: 'gpu rental',
      pagination: { page: 2, limit: 3, offset: 3 },
      order: 'asc',
    });
  });
});

describe('GET /api/berita/:slug', () => {
  test('ada → 200 body item', async () => {
    const { app } = createTestApp();
    const { status, body } = await getJson(app, `/api/berita/${sampleBerita.slug}`);
    expect(status).toBe(200);
    expect(body).toEqual(sampleBerita);
  });

  test('tidak ada → 404 { error }', async () => {
    const { app } = createTestApp();
    const { status, body } = await getJson(app, '/api/berita/tidak-ada');
    expect(status).toBe(404);
    expect(body).toEqual({ error: 'Berita tidak ditemukan' });
  });
});

describe('GET /api/dokumentasi', () => {
  test('q & kategori diteruskan; kategori kosong → null', async () => {
    const { app, calls } = createTestApp();
    await getJson(app, '/api/dokumentasi?q=workshop&kategori=%20%20');
    expect(calls[0].args[0]).toMatchObject({ q: 'workshop', kategori: null });
    await getJson(app, '/api/dokumentasi?kategori=Workshop');
    expect(calls[1].args[0]).toMatchObject({ kategori: 'Workshop' });
  });

  test('detail tidak ada → 404', async () => {
    const { app } = createTestApp();
    const { status } = await getJson(app, '/api/dokumentasi/tidak-ada');
    expect(status).toBe(404);
  });
});

describe('GET /api/tim', () => {
  test('q di-normalize, pagination nonaktif tanpa query', async () => {
    const { app, calls } = createTestApp();
    const { body } = await getJson(app, '/api/tim?q=' + encodeURIComponent('  koordinator '));
    expect(body).toEqual({ items: [sampleTim] });
    expect(calls[0].args[0]).toEqual({ q: 'koordinator', pagination: NO_PAGINATION });
  });
});

describe('endpoint konten beranda & layanan', () => {
  test('layanan, hero-slides, klien, testimoni → { items }', async () => {
    const { app } = createTestApp();
    expect((await getJson(app, '/api/layanan')).body).toEqual({ items: [sampleLayanan] });
    expect((await getJson(app, '/api/hero-slides')).body).toEqual({ items: [sampleHero] });
    expect((await getJson(app, '/api/klien')).body).toEqual({ items: [sampleKlien] });
    expect((await getJson(app, '/api/testimoni')).body).toEqual({ items: [sampleTestimoni] });
  });

  test('detail layanan ada/tidak ada → 200/404', async () => {
    const { app } = createTestApp();
    expect((await getJson(app, '/api/layanan/pelatihan')).status).toBe(200);
    expect((await getJson(app, '/api/layanan/tidak-ada')).status).toBe(404);
  });
});

describe('GET /api/kursus', () => {
  test('tanpa query → { items } + pagination nonaktif (semua kursus)', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await getJson(app, '/api/kursus');
    expect(status).toBe(200);
    expect(body).toEqual({ items: [sampleKursus] });
    expect(calls).toEqual([
      { method: 'kursus.list', args: [{ q: null, pagination: NO_PAGINATION }] },
    ]);
  });

  test('q & limit dinormalisasi & diteruskan ke repository', async () => {
    const { app, calls } = createTestApp();
    await getJson(app, '/api/kursus?q=' + encodeURIComponent('  jurnal ') + '&limit=5');
    expect(calls[0].args[0]).toEqual({
      q: 'jurnal',
      pagination: { page: 1, limit: 5, offset: 0 },
    });
  });
});

describe('GET /api/kursus/:kode', () => {
  test('kode kecil ("r01") tetap diterima (URL lama) → 200', async () => {
    const { app } = createTestApp();
    const { status, body } = await getJson(app, '/api/kursus/r01');
    expect(status).toBe(200);
    expect(body).toEqual(sampleKursus);
  });

  test('tidak ada → 404 { error }', async () => {
    const { app } = createTestApp();
    const { status, body } = await getJson(app, '/api/kursus/Z99');
    expect(status).toBe(404);
    expect(body).toEqual({ error: 'Kursus tidak ditemukan' });
  });
});

describe('GET /api/progress (publik tanpa akun)', () => {
  test('pesertaId sah → 200 { items } tanpa token; kode difilter & uppercase', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await getJson(
      app,
      '/api/progress?pesertaId=abc12345-uuid&kode=r01',
    );
    expect(status).toBe(200);
    expect(body).toEqual({ items: [sampleProgres] });
    expect(calls).toEqual([
      { method: 'progress.list', args: ['abc12345-uuid', 'R01'] },
    ]);
  });

  test('tanpa kode → filter null (semua progres peserta)', async () => {
    const { app, calls } = createTestApp();
    await getJson(app, '/api/progress?pesertaId=abc12345-uuid');
    expect(calls[0].args).toEqual(['abc12345-uuid', null]);
  });

  test('pesertaId kosong/salah pola → 400 { error } tanpa memanggil repo', async () => {
    const { app, calls } = createTestApp();
    const kosong = await getJson(app, '/api/progress');
    expect(kosong.status).toBe(400);
    expect(kosong.body).toEqual({
      error: 'Query "pesertaId" wajib: 8-64 karakter (huruf/angka/-)',
    });

    const pendek = await getJson(app, '/api/progress?pesertaId=abc');
    expect(pendek.status).toBe(400);

    const liar = await getJson(app, '/api/progress?pesertaId=' + encodeURIComponent('id dengan spasi!'));
    expect(liar.status).toBe(400);
    expect(calls).toEqual([]);
  });

  test('kode query tidak sah → 400 { error }', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await getJson(
      app,
      '/api/progress?pesertaId=abc12345-uuid&kode=' + encodeURIComponent('R01;DROP'),
    );
    expect(status).toBe(400);
    expect(body).toEqual({ error: 'Query "kode" hanya huruf/angka/- (maks 12)' });
    expect(calls).toEqual([]);
  });
});

describe('POST /api/progress (publik tanpa akun)', () => {
  const BODY_PROGRESS = {
    pesertaId: 'abc12345-uuid',
    kode: 'r01',
    state: { enrolled: true, completed: [true, false], activeModule: 1 },
  };

  test('body valid → 200 entri tersimpan (kode uppercase) + repo dipanggil', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/progress', BODY_PROGRESS, null);
    expect(status).toBe(200);
    expect(body).toEqual({
      kode: 'R01',
      state: BODY_PROGRESS.state,
      diperbaruiPada: sampleProgres.diperbaruiPada,
    });
    expect(calls).toEqual([
      {
        method: 'progress.simpan',
        args: [{ pesertaId: 'abc12345-uuid', kode: 'R01', state: BODY_PROGRESS.state }],
      },
    ]);
  });

  test('state terlalu besar (>32KB) / bukan objek / pesertaId pendek → 400 tanpa repo', async () => {
    const { app, calls } = createTestApp();
    const raksasa = await kirimTulis(app, 'POST', '/api/progress', {
      ...BODY_PROGRESS,
      state: { review: 'x'.repeat(40_000) },
    }, null);
    expect(raksasa.status).toBe(400);
    expect((raksasa.body as { error: string }).error).toContain('state');

    const array = await kirimTulis(app, 'POST', '/api/progress', { ...BODY_PROGRESS, state: [] }, null);
    expect(array.status).toBe(400);

    const idPendek = await kirimTulis(app, 'POST', '/api/progress', { ...BODY_PROGRESS, pesertaId: 'abc' }, null);
    expect(idPendek.status).toBe(400);
    expect(calls).toEqual([]);
  });

  test('61 percobaan berturut dari 1 IP → 429 + Retry-After (batas 60/menit)', async () => {
    const { app } = createTestApp();
    const kirim = (ip: string) =>
      app.handle(
        new Request('http://localhost/api/progress', {
          method: 'POST',
          headers: { 'content-type': 'application/json', 'x-forwarded-for': ip },
          body: JSON.stringify(BODY_PROGRESS),
        }),
      );
    // 60 percobaan pertama diterima (200), percobaan ke-61 diblokir.
    for (let i = 0; i < 60; i++) {
      expect((await kirim('203.0.113.7')).status).toBe(200);
    }
    const terkunci = await kirim('203.0.113.7');
    expect(terkunci.status).toBe(429);
    expect(terkunci.headers.get('Retry-After')).not.toBeNull();
    const pesan = (await terkunci.json()) as { error: string };
    expect(pesan.error).toContain('simpan progres');

    // IP lain tidak terdampak.
    expect((await kirim('203.0.113.8')).status).toBe(200);
  });
});

describe('GET /api/profil', () => {
  test('ada → 200 body profil', async () => {
    const { app } = createTestApp();
    const { status, body } = await getJson(app, '/api/profil');
    expect(status).toBe(200);
    expect(body).toEqual(sampleProfil);
  });

  test('edge: belum di-seed (null) → 404', async () => {
    const handle = createFakeRepos();
    handle.repos.profil = { get: async () => null, update: async () => null };
    const { app } = createTestApp(handle);
    const { status, body } = await getJson(app, '/api/profil');
    expect(status).toBe(404);
    expect(body).toEqual({ error: 'Profil belum diisi' });
  });
});

describe('GET /api/inference', () => {
  test('ada → 200 body konten inference', async () => {
    const { app } = createTestApp();
    const { status, body } = await getJson(app, '/api/inference');
    expect(status).toBe(200);
    expect(body).toEqual(sampleInference);
  });

  test('edge: belum di-seed (null) → 404', async () => {
    const handle = createFakeRepos();
    handle.repos.inference = { get: async () => null, update: async () => null };
    const { app } = createTestApp(handle);
    const { status, body } = await getJson(app, '/api/inference');
    expect(status).toBe(404);
    expect(body).toEqual({ error: 'Konten inference belum diisi' });
  });
});

describe('error handling', () => {
  test('rute tidak dikenal → 404 JSON { error }', async () => {
    const { app } = createTestApp();
    const { status, body } = await getJson(app, '/api/tidak-ada');
    expect(status).toBe(404);
    expect(body).toEqual({ error: 'Not Found' });
  });

  test('repository melempar error → 500 JSON, bukan bocor stack', async () => {
    const handle = createFakeRepos();
    handle.repos.berita.list = async () => {
      throw new Error('db down');
    };
    const { app } = createTestApp(handle);
    const { status, body } = await getJson(app, '/api/berita');
    expect(status).toBe(500);
    expect(body).toEqual({ error: 'Internal Server Error' });
  });

  test('body bukan JSON valid → 400 eksplisit, bukan 500 (salah ketik = klien)', async () => {
    const { app } = createTestApp();
    const res = await app.handle(
      new Request('http://localhost/api/profil', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: '{rusak',
      }),
    );
    expect(res.status).toBe(400);
    expect(await res.json()).toEqual({ error: 'Body bukan JSON valid' });
  });
});

describe('CORS', () => {
  test('origin whitelist → di-echo + Vary: Origin', async () => {
    const { app } = createTestApp();
    const res = await app.handle(
      new Request('http://localhost/api/health', {
        headers: { Origin: 'http://localhost:5173' },
      }),
    );
    expect(res.headers.get('Access-Control-Allow-Origin')).toBe('http://localhost:5173');
    expect(res.headers.get('Vary')).toBe('Origin');
  });

  test('origin di luar whitelist → header tidak ada', async () => {
    const { app } = createTestApp();
    const res = await app.handle(
      new Request('http://localhost/api/health', {
        headers: { Origin: 'https://evil.test' },
      }),
    );
    expect(res.headers.get('Access-Control-Allow-Origin')).toBeNull();
  });

  test('preflight OPTIONS → 204 + allow methods/headers', async () => {
    const { app } = createTestApp();
    const res = await app.handle(
      new Request('http://localhost/api/berita', {
        method: 'OPTIONS',
        headers: {
          Origin: 'http://localhost:5173',
          'Access-Control-Request-Method': 'GET',
          'Access-Control-Request-Headers': 'content-type,authorization',
        },
      }),
    );
    expect(res.status).toBe(204);
    expect(res.headers.get('Access-Control-Allow-Methods')).toBe('GET, POST, PUT, DELETE, OPTIONS');
    expect(res.headers.get('Access-Control-Allow-Headers')).toBe('content-type, authorization');
    expect(res.headers.get('Access-Control-Allow-Origin')).toBe('http://localhost:5173');
  });
});

/** Kirim request tulis JSON (POST/PUT/DELETE) — opsi token: `Bearer <TOKEN_ADMIN>`. */
const kirimTulis = async (
  app: ReturnType<typeof createApp>,
  method: 'POST' | 'PUT' | 'DELETE',
  path: string,
  body: unknown,
  token: string | null = TOKEN_ADMIN,
) => {
  const res = await app.handle(
    new Request(`http://localhost${path}`, {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...(token === null ? {} : { Authorization: `Bearer ${token}` }),
      },
      body: JSON.stringify(body),
    }),
  );
  return { status: res.status, body: (await res.json()) as unknown };
};

const BODY_DOK = {
  judul: 'Dokumentasi Baru',
  deskripsi: 'Deskripsi baru.',
  tanggal: '2026-10-01',
  kategori: 'Workshop',
};

const BODY_BERITA = {
  judul: 'Berita Baru',
  ringkasan: 'Ringkasan.',
  isi: 'Isi lengkap.',
  tanggal: '2026-10-01',
  penulis: 'Admin',
};

describe('POST /api/admin/login', () => {
  test('password benar → 200 { token } yang valid', async () => {
    const { app } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/admin/login', {
      password: ADMIN_PASSWORD_TEST,
    });
    expect(status).toBe(200);
    const token = (body as { token: string }).token;
    expect(typeof token).toBe('string');
    // Token yang diterbitkan diterima guard route tulis.
    const cek = await kirimTulis(app, 'DELETE', `/api/dokumentasi/${sampleDok.slug}`, null, token);
    expect(cek.status).toBe(200);
  });

  test('password salah / body bukan objek → 401', async () => {
    const { app } = createTestApp();
    const salah = await kirimTulis(app, 'POST', '/api/admin/login', { password: 'salah' });
    expect(salah.status).toBe(401);
    expect(salah.body).toEqual({ error: 'Password salah' });

    const bukanObjek = await kirimTulis(app, 'POST', '/api/admin/login', 'halo');
    expect(bukanObjek.status).toBe(401);
  });

  test('gagal beruntun dibatasi 429 + Retry-After; login benar me-reset hitungan', async () => {
    const { app } = createTestApp();
    const kirim = (password: string, ip: string) =>
      app.handle(
        new Request('http://localhost/api/admin/login', {
          method: 'POST',
          headers: { 'content-type': 'application/json', 'x-forwarded-for': ip },
          body: JSON.stringify({ password }),
        }),
      );

    // IP A: 5 gagal pertama tetap 401 (kuota habis), percobaan ke-6 diblokir 429.
    for (let i = 0; i < 5; i++) {
      expect((await kirim('salah', '198.51.100.1')).status).toBe(401);
    }
    const terkunci = await kirim('salah', '198.51.100.1');
    expect(terkunci.status).toBe(429);
    expect(terkunci.headers.get('Retry-After')).not.toBeNull();
    const pesan = (await terkunci.json()) as { error: string };
    expect(pesan.error).toContain('percobaan login');

    // IP A tetap terkunci walau password benar — dicek sebelum verifikasi.
    expect((await kirim(ADMIN_PASSWORD_TEST, '198.51.100.1')).status).toBe(429);

    // IP lain tidak terdampak lock.
    expect((await kirim(ADMIN_PASSWORD_TEST, '198.51.100.2')).status).toBe(200);

    // Login benar mereset hitungan: 4 gagal + sukses, lalu 5 gagal baru
    // terkunci pada percobaan keenam sesudah reset.
    for (let i = 0; i < 4; i++) {
      expect((await kirim('salah', '198.51.100.3')).status).toBe(401);
    }
    expect((await kirim(ADMIN_PASSWORD_TEST, '198.51.100.3')).status).toBe(200);
    for (let i = 0; i < 4; i++) {
      expect((await kirim('salah', '198.51.100.3')).status).toBe(401);
    }
    expect((await kirim('salah', '198.51.100.3')).status).toBe(401); // gagal ke-5 setelah reset
    expect((await kirim('salah', '198.51.100.3')).status).toBe(429); // baru terkunci
  });
});

describe('guard route tulis (dokumentasi & berita)', () => {
  test('tanpa Authorization → 401; token salah → 401', async () => {
    const { app } = createTestApp();
    const tanpaToken = await kirimTulis(app, 'POST', '/api/dokumentasi', BODY_DOK, null);
    expect(tanpaToken.status).toBe(401);

    const tokenSalah = await kirimTulis(
      app,
      'POST',
      '/api/dokumentasi',
      BODY_DOK,
      '1700000000000.abc',
    );
    expect(tokenSalah.status).toBe(401);
  });
});

describe('POST /api/dokumentasi', () => {
  test('token valid + body valid → 201, slug dibuat dari judul, repo.create dipanggil', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/dokumentasi', BODY_DOK);
    expect(status).toBe(201);
    expect((body as { slug: string }).slug).toBe('dokumentasi-baru');

    const create = calls.find((c) => c.method === 'dokumentasi.create');
    expect(create).toBeDefined();
    expect(create?.args[0]).toBe('dokumentasi-baru');
  });

  test('slug bentrok dengan konten ada → 409', async () => {
    const { app } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/dokumentasi', {
      ...BODY_DOK,
      judul: 'dok-1', // slugDariJudul → 'dok-1' = sampleDok.slug yang sudah ada
    });
    expect(status).toBe(409);
    expect(body).toEqual({ error: 'Slug sudah dipakai — judul bentrok dengan konten lain' });
  });

  test('field wajib hilang → 400 dengan pesan field', async () => {
    const { app } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/dokumentasi', {
      ...BODY_DOK,
      kategori: '',
    });
    expect(status).toBe(400);
    expect(body).toEqual({ error: 'Field "kategori" wajib diisi' });
  });

  test('slug hasil judul kosong (tanda baca semua) → 400', async () => {
    const { app } = createTestApp();
    const { status } = await kirimTulis(app, 'POST', '/api/dokumentasi', {
      ...BODY_DOK,
      judul: '!!! ???',
    });
    expect(status).toBe(400);
  });
});

describe('PUT /api/dokumentasi/:slug', () => {
  test('slug ada + body valid → 200, repo.update dipanggil dengan slug yang sama', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(
      app,
      'PUT',
      `/api/dokumentasi/${sampleDok.slug}`,
      BODY_DOK,
    );
    expect(status).toBe(200);
    expect((body as { slug: string }).slug).toBe(sampleDok.slug);

    const update = calls.find((c) => c.method === 'dokumentasi.update');
    expect(update?.args[0]).toBe(sampleDok.slug);
  });

  test('slug tidak ada → 404', async () => {
    const { app } = createTestApp();
    const { status, body } = await kirimTulis(app, 'PUT', '/api/dokumentasi/tidak-ada', BODY_DOK);
    expect(status).toBe(404);
    expect(body).toEqual({ error: 'Dokumentasi tidak ditemukan' });
  });

  test('body tidak valid → 400 tanpa menyentuh repository', async () => {
    const { app, calls } = createTestApp();
    const { status } = await kirimTulis(app, 'PUT', `/api/dokumentasi/${sampleDok.slug}`, {
      ...BODY_DOK,
      tanggal: 'besok',
    });
    expect(status).toBe(400);
    expect(calls.some((c) => c.method === 'dokumentasi.update')).toBe(false);
  });
});

describe('DELETE /api/dokumentasi/:slug', () => {
  test('slug ada → 200 { ok: true }', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(
      app,
      'DELETE',
      `/api/dokumentasi/${sampleDok.slug}`,
      null,
    );
    expect(status).toBe(200);
    expect(body).toEqual({ ok: true });
    expect(calls.some((c) => c.method === 'dokumentasi.remove')).toBe(true);
  });

  test('slug tidak ada → 404', async () => {
    const { app } = createTestApp();
    const { status, body } = await kirimTulis(app, 'DELETE', '/api/dokumentasi/hilang', null);
    expect(status).toBe(404);
    expect(body).toEqual({ error: 'Dokumentasi tidak ditemukan' });
  });
});

describe('POST /api/berita (mirror dokumentasi)', () => {
  test('token valid + body valid → 201 + slug dari judul', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/berita', BODY_BERITA);
    expect(status).toBe(201);
    expect((body as { slug: string }).slug).toBe('berita-baru');
    expect(calls.some((c) => c.method === 'berita.create')).toBe(true);
  });

  test('tanpa token → 401; field hilang → 400', async () => {
    const { app } = createTestApp();
    const tanpaToken = await kirimTulis(app, 'POST', '/api/berita', BODY_BERITA, null);
    expect(tanpaToken.status).toBe(401);

    const tanpaPenulis = await kirimTulis(app, 'POST', '/api/berita', {
      ...BODY_BERITA,
      penulis: '  ',
    });
    expect(tanpaPenulis.status).toBe(400);
    expect(tanpaPenulis.body).toEqual({ error: 'Field "penulis" wajib diisi' });
  });

  test('slug bentrok → 409', async () => {
    const { app } = createTestApp();
    const { status } = await kirimTulis(app, 'POST', '/api/berita', {
      ...BODY_BERITA,
      judul: 'berita-1', // slugDariJudul → 'berita-1' = sampleBerita.slug
    });
    expect(status).toBe(409);
  });
});

describe('PUT & DELETE /api/berita', () => {
  test('PUT slug ada → 200; tidak ada → 404', async () => {
    const { app } = createTestApp();
    const ada = await kirimTulis(app, 'PUT', `/api/berita/${sampleBerita.slug}`, BODY_BERITA);
    expect(ada.status).toBe(200);
    expect((ada.body as { slug: string }).slug).toBe(sampleBerita.slug);

    const tidakAda = await kirimTulis(app, 'PUT', '/api/berita/tidak-ada', BODY_BERITA);
    expect(tidakAda.status).toBe(404);
    expect(tidakAda.body).toEqual({ error: 'Berita tidak ditemukan' });
  });

  test('DELETE ada → 200; tidak ada → 404; tanpa token → 401', async () => {
    const { app } = createTestApp();
    const ada = await kirimTulis(app, 'DELETE', `/api/berita/${sampleBerita.slug}`, null);
    expect(ada.status).toBe(200);
    expect(ada.body).toEqual({ ok: true });

    const tidakAda = await kirimTulis(app, 'DELETE', '/api/berita/hilang', null);
    expect(tidakAda.status).toBe(404);

    const tanpaToken = await kirimTulis(app, 'DELETE', `/api/berita/${sampleBerita.slug}`, null, null);
    expect(tanpaToken.status).toBe(401);
  });
});

const BODY_LAYANAN = {
  nama: 'Layanan Baru',
  tagline: 'Tagline baru.',
  deskripsi: 'Deskripsi layanan.',
  fitur: ['Poin satu', 'Poin dua'],
};

describe('POST /api/layanan (mirror berita)', () => {
  test('token valid + body valid → 201 + slug dari nama + repo.create dipanggil', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/layanan', BODY_LAYANAN);
    expect(status).toBe(201);
    expect((body as { slug: string }).slug).toBe('layanan-baru');

    const create = calls.find((c) => c.method === 'layanan.create');
    expect(create).toBeDefined();
    expect(create?.args[0]).toBe('layanan-baru');
    expect((create?.args[1] as LayananInput).fitur).toEqual(['Poin satu', 'Poin dua']);
  });

  test('slug bentrok dengan layanan ada → 409', async () => {
    const { app } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/layanan', {
      ...BODY_LAYANAN,
      nama: 'pelatihan', // slugDariJudul → 'pelatihan' = sampleLayanan.slug
    });
    expect(status).toBe(409);
    expect(body).toEqual({ error: 'Slug sudah dipakai — nama bentrok dengan layanan lain' });
  });

  test('fitur kosong → 400; tanpa token → 401', async () => {
    const { app, calls } = createTestApp();
    const kosong = await kirimTulis(app, 'POST', '/api/layanan', { ...BODY_LAYANAN, fitur: [] });
    expect(kosong.status).toBe(400);
    expect(kosong.body).toEqual({ error: 'Field "fitur" minimal 1 item' });
    expect(calls.some((c) => c.method === 'layanan.create')).toBe(false);

    const tanpaToken = await kirimTulis(app, 'POST', '/api/layanan', BODY_LAYANAN, null);
    expect(tanpaToken.status).toBe(401);
  });

  test('nama tidak menghasilkan slug → 400', async () => {
    const { app } = createTestApp();
    const { status } = await kirimTulis(app, 'POST', '/api/layanan', { ...BODY_LAYANAN, nama: '!!!' });
    expect(status).toBe(400);
  });
});

describe('PUT & DELETE /api/layanan', () => {
  test('PUT slug ada → 200 tanpa mengubah slug; tidak ada → 404; body invalid → 400', async () => {
    const { app, calls } = createTestApp();
    const ada = await kirimTulis(app, 'PUT', `/api/layanan/${sampleLayanan.slug}`, BODY_LAYANAN);
    expect(ada.status).toBe(200);
    expect((ada.body as { slug: string }).slug).toBe(sampleLayanan.slug);
    expect(calls.find((c) => c.method === 'layanan.update')?.args[0]).toBe(sampleLayanan.slug);

    const tidakAda = await kirimTulis(app, 'PUT', '/api/layanan/tidak-ada', BODY_LAYANAN);
    expect(tidakAda.status).toBe(404);
    expect(tidakAda.body).toEqual({ error: 'Layanan tidak ditemukan' });

    const sebelumInvalid = calls.filter((c) => c.method === 'layanan.update').length;
    const invalid = await kirimTulis(app, 'PUT', `/api/layanan/${sampleLayanan.slug}`, {
      ...BODY_LAYANAN,
      tagline: '',
    });
    expect(invalid.status).toBe(400);
    // Body invalid ditolak SEBELUM repository disentuh.
    expect(calls.filter((c) => c.method === 'layanan.update').length).toBe(sebelumInvalid);
  });

  test('DELETE ada → 200 { ok: true }; tidak ada → 404; tanpa token → 401', async () => {
    const { app, calls } = createTestApp();
    const ada = await kirimTulis(app, 'DELETE', `/api/layanan/${sampleLayanan.slug}`, null);
    expect(ada.status).toBe(200);
    expect(ada.body).toEqual({ ok: true });
    expect(calls.some((c) => c.method === 'layanan.remove')).toBe(true);

    const tidakAda = await kirimTulis(app, 'DELETE', '/api/layanan/hilang', null);
    expect(tidakAda.status).toBe(404);

    const tanpaToken = await kirimTulis(app, 'DELETE', `/api/layanan/${sampleLayanan.slug}`, null, null);
    expect(tanpaToken.status).toBe(401);
  });
});

const BODY_KURSUS = {
  kode: 'p02',
  judul: 'Kursus Baru',
  deskripsi: 'Deskripsi.',
  tentang: 'Tentang.',
  durasi: '2 sesi',
  level: 'Pemula',
  format: 'Online',
  instruktur: 'Tim',
  peran: 'Pengajar',
  inisial: 'T',
  target: ['Mahasiswa'],
  hasil: ['Hasil 1'],
  modul: [{ judul: 'M1', deskripsi: 'D1', meta: '2 video' }],
};

describe('CRUD /api/kursus (admin)', () => {
  test('POST valid → 201 + kode di-uppercase; duplikat → 409; invalid → 400; tanpa token → 401', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/kursus', BODY_KURSUS);
    expect(status).toBe(201);
    expect((body as { kode: string }).kode).toBe('P02');
    expect(calls.find((c) => c.method === 'kursus.create')).toBeDefined();

    const duplikat = await kirimTulis(app, 'POST', '/api/kursus', { ...BODY_KURSUS, kode: 'R01' });
    expect(duplikat.status).toBe(409);

    const invalid = await kirimTulis(app, 'POST', '/api/kursus', { ...BODY_KURSUS, modul: [] });
    expect(invalid.status).toBe(400);

    const tanpaToken = await kirimTulis(app, 'POST', '/api/kursus', BODY_KURSUS, null);
    expect(tanpaToken.status).toBe(401);
  });

  test('POST dengan video + kuis per modul → 201, data modul terpetak ke repo.create', async () => {
    const { app, calls } = createTestApp();
    const soal = { pertanyaan: 'Pertanyaan uji?', opsi: ['Benar', 'Salah'], kunci: 0 };
    const { status } = await kirimTulis(app, 'POST', '/api/kursus', {
      ...BODY_KURSUS,
      modul: [{ judul: 'M1', deskripsi: 'D1', meta: '2 video', video: 'https://youtu.be/abc123', quiz: [soal] }],
    });
    expect(status).toBe(201);
    const create = calls.find((c) => c.method === 'kursus.create');
    expect(create).toBeDefined();
    const modul = (create?.args[0] as { modul: Array<Record<string, unknown>> }).modul;
    expect(modul[0].video).toBe('https://youtu.be/abc123');
    expect(modul[0].quiz).toEqual([soal]);
  });

  test('PUT cocok → 200; kode beda dengan path → 400; tidak ada → 404', async () => {
    const { app } = createTestApp();
    const { status } = await kirimTulis(app, 'PUT', '/api/kursus/R01', { ...BODY_KURSUS, kode: 'r01' });
    expect(status).toBe(200);

    const beda = await kirimTulis(app, 'PUT', '/api/kursus/R01', BODY_KURSUS);
    expect(beda.status).toBe(400);

    const hilang = await kirimTulis(app, 'PUT', '/api/kursus/Z99', { ...BODY_KURSUS, kode: 'Z99' });
    expect(hilang.status).toBe(404);
  });

  test('DELETE ada → 200; tidak ada → 404; tanpa token → 401', async () => {
    const { app } = createTestApp();
    expect((await kirimTulis(app, 'DELETE', '/api/kursus/R01', null)).status).toBe(200);
    expect((await kirimTulis(app, 'DELETE', '/api/kursus/Z99', null)).status).toBe(404);
    expect((await kirimTulis(app, 'DELETE', '/api/kursus/R01', null, null)).status).toBe(401);
  });
});

describe('POST /api/admin/upload + GET /uploads/:nama', () => {
  const filePng = () => new File([new Uint8Array([137, 80, 78, 71])], 'foto.png', { type: 'image/png' });

  async function kirimFile(
    app: ReturnType<typeof createApp>,
    file: unknown,
    token: string | null = TOKEN_ADMIN,
  ) {
    const form = new FormData();
    form.append('gambar', file as Blob);
    const res = await app.handle(
      new Request('http://localhost/api/admin/upload', {
        method: 'POST',
        headers: token === null ? {} : { Authorization: `Bearer ${token}` },
        body: form,
      }),
    );
    return { status: res.status, body: (await res.json()) as unknown };
  }

  test('tanpa token → 401; file bukan gambar → 400', async () => {
    const { mkdtemp, rm } = await import('node:fs/promises');
    const { tmpdir } = await import('node:os');
    const { join } = await import('node:path');
    const dir = await mkdtemp(join(tmpdir(), 'rute-unggah-'));
    const lama = process.env.UPLOAD_DIR;
    process.env.UPLOAD_DIR = dir;
    try {
      const { app } = createTestApp();
      const tanpaToken = await kirimFile(app, filePng(), null);
      expect(tanpaToken.status).toBe(401);

      const form = new FormData();
      form.append('gambar', new File(['halo'], 'a.txt', { type: 'text/plain' }));
      const res = await app.handle(
        new Request('http://localhost/api/admin/upload', {
          method: 'POST',
          headers: { Authorization: `Bearer ${TOKEN_ADMIN}` },
          body: form,
        }),
      );
      expect(res.status).toBe(400);
    } finally {
      if (lama === undefined) delete process.env.UPLOAD_DIR;
      else process.env.UPLOAD_DIR = lama;
      await rm(dir, { recursive: true, force: true });
    }
  });

  test('file valid → 201 { url } + GET menyajikan file; traversal → 404', async () => {
    const { mkdtemp, rm } = await import('node:fs/promises');
    const { tmpdir } = await import('node:os');
    const { join } = await import('node:path');
    const dir = await mkdtemp(join(tmpdir(), 'rute-unggah-'));
    const lama = process.env.UPLOAD_DIR;
    process.env.UPLOAD_DIR = dir;
    try {
      const { app } = createTestApp();
      const { status, body } = await kirimFile(app, filePng());
      expect(status).toBe(201);
      const url = (body as { url: string }).url;
      expect(url).toMatch(/^\/uploads\/[A-Za-z0-9_-]+\.png$/);

      const ambil = await app.handle(new Request(`http://localhost${url}`));
      expect(ambil.status).toBe(200);
      expect(ambil.headers.get('Content-Type')).toBe('image/png');

      const hilang = await app.handle(new Request('http://localhost/uploads/tidak-ada.png'));
      expect(hilang.status).toBe(404);
      const jahat = await app.handle(new Request('http://localhost/uploads/..%2Fapp.ts'));
      expect(jahat.status).toBe(404);
    } finally {
      if (lama === undefined) delete process.env.UPLOAD_DIR;
      else process.env.UPLOAD_DIR = lama;
      await rm(dir, { recursive: true, force: true });
    }
  });
});

// — Prioritas 2: PUT /api/profil ————————————————————————————————————————

const BODY_PROFIL = {
  nama: 'AI Center Universitas Surabaya',
  tagline: 'Pusat riset AI Ubaya.',
  ringkasan: 'Ringkasan profil.',
  alamat: 'Gedung Perpustakaan LT.4',
  email: 'aicenter@unit.ubaya.ac.id',
  telepon: '0895-6342-22240',
  visi: 'Menjadi yang terdepan — Deskripsi visi.',
  misi: 'Mendorong riset.\nMenghasilkan produk.',
};

describe('PUT /api/profil (Prioritas 2)', () => {
  test('token valid + body valid → 200, repo.profil.update dipanggil dengan data bersih', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'PUT', '/api/profil', {
      ...BODY_PROFIL,
      nama: '  AI Center  ',
    });
    expect(status).toBe(200);
    expect((body as Profil).nama).toBe('AI Center');
    const update = calls.find((c) => c.method === 'profil.update');
    expect(update).toBeDefined();
    expect(update?.args[0]).toEqual({ ...BODY_PROFIL, nama: 'AI Center' });
  });

  test('tanpa token → 401; field wajib kosong → 400 menyebut field', async () => {
    const { app } = createTestApp();
    expect((await kirimTulis(app, 'PUT', '/api/profil', BODY_PROFIL, null)).status).toBe(401);
    const { status, body } = await kirimTulis(app, 'PUT', '/api/profil', {
      ...BODY_PROFIL,
      telepon: '',
    });
    expect(status).toBe(400);
    expect(body).toEqual({ error: 'Field "telepon" wajib diisi' });
  });

  test('baris profil belum di-seed (update → null) → 404 eksplisit', async () => {
    const handle = createFakeRepos();
    handle.repos.profil.update = async () => null;
    const { app } = createTestApp(handle);
    const { status, body } = await kirimTulis(app, 'PUT', '/api/profil', BODY_PROFIL);
    expect(status).toBe(404);
    expect(body).toEqual({ error: 'Profil tidak ditemukan — jalankan seed dulu' });
  });
});

// — Konten halaman inference: GET sudah di atas, ini PUT —————————————————

describe('PUT /api/inference', () => {
  test('token valid + body valid → 200, repo.inference.update dipanggil dengan data bersih', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'PUT', '/api/inference', {
      ...BODY_INFERENCE,
      judulApaItu: '  Apa itu Inference Solution?  ',
      kebutuhan: ['  Kebutuhan 1  ', 'Kebutuhan 2', ''],
    });
    expect(status).toBe(200);
    expect((body as KontenInference).judulApaItu).toBe('Apa itu Inference Solution?');
    const update = calls.find((c) => c.method === 'inference.update');
    expect(update).toBeDefined();
    // Item kosong dibuang (bacaDaftarTeks) — data masuk repository sudah bersih.
    expect(update?.args[0]).toEqual({ ...BODY_INFERENCE, kebutuhan: ['Kebutuhan 1', 'Kebutuhan 2'] });
  });

  test('tanpa token → 401; field wajib kosong → 400 menyebut field', async () => {
    const { app } = createTestApp();
    expect((await kirimTulis(app, 'PUT', '/api/inference', BODY_INFERENCE, null)).status).toBe(401);
    const { status, body } = await kirimTulis(app, 'PUT', '/api/inference', {
      ...BODY_INFERENCE,
      deskripsiApaItu: '',
    });
    expect(status).toBe(400);
    expect(body).toEqual({ error: 'Field "deskripsiApaItu" wajib diisi' });
  });

  test('daftar kosong → 400 menyebut minimal item', async () => {
    const { app } = createTestApp();
    const { status, body } = await kirimTulis(app, 'PUT', '/api/inference', {
      ...BODY_INFERENCE,
      alur: [],
    });
    expect(status).toBe(400);
    expect(body).toEqual({ error: 'Field "alur" minimal 1 item' });
  });

  test('baris belum di-seed (update → null) → 404 eksplisit', async () => {
    const handle = createFakeRepos();
    handle.repos.inference.update = async () => null;
    const { app } = createTestApp(handle);
    const { status, body } = await kirimTulis(app, 'PUT', '/api/inference', BODY_INFERENCE);
    expect(status).toBe(404);
    expect(body).toEqual({ error: 'Konten inference tidak ditemukan — jalankan seed dulu' });
  });
});

// — Prioritas 2: CRUD /api/tim ————————————————————————————————————————

const BODY_TIM = {
  nama: 'Dr. Contoh',
  peran: 'Ketua',
  kredensial: 'Ph.D.',
  foto: '/tim/contoh.jpg',
  urutan: 7,
};

describe('CRUD /api/tim (Prioritas 2)', () => {
  test('POST valid → 201; urutan lama digeser +1 lalu item baru disimpan di urutan 0 (paling atas)', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/tim', BODY_TIM);
    expect(status).toBe(201);
    expect((body as AnggotaTim).id).toBe(9);
    expect((body as AnggotaTim).urutan).toBe(0);
    expect(calls.find((c) => c.method === 'tim.naikkanSemuaUrutan')).toBeDefined();
    // Urutan dari body diabaikan saat tambah — selalu 0.
    expect(calls.find((c) => c.method === 'tim.create')?.args[0]).toEqual({ ...BODY_TIM, urutan: 0 });
    // Geser terjadi sebelum insert supaya baris lama tidak menimpa posisi baru.
    const idxGeser = calls.findIndex((c) => c.method === 'tim.naikkanSemuaUrutan');
    const idxBuat = calls.findIndex((c) => c.method === 'tim.create');
    expect(idxGeser).toBeGreaterThanOrEqual(0);
    expect(idxGeser).toBeLessThan(idxBuat);
    expect((await kirimTulis(app, 'POST', '/api/tim', BODY_TIM, null)).status).toBe(401);
  });

  test('POST body tidak valid → 400 tanpa menyentuh repository', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/tim', {
      ...BODY_TIM,
      urutan: 'abc',
    });
    expect(status).toBe(400);
    expect(body).toEqual({ error: 'Field "urutan" harus bilangan bulat 0-9999' });
    expect(calls.some((c) => c.method === 'tim.create')).toBe(false);
  });

  test('PUT id valid → 200; id tidak sah → 400; tidak ada → 404', async () => {
    const { app } = createTestApp();
    const ok = await kirimTulis(app, 'PUT', '/api/tim/1', BODY_TIM);
    expect(ok.status).toBe(200);
    expect((ok.body as AnggotaTim).nama).toBe('Dr. Contoh');

    const salah = await kirimTulis(app, 'PUT', '/api/tim/abc', BODY_TIM);
    expect(salah.status).toBe(400);
    expect(salah.body).toEqual({ error: 'Id tim tidak valid — isi bilangan bulat positif' });

    const hilang = await kirimTulis(app, 'PUT', '/api/tim/99', BODY_TIM);
    expect(hilang.status).toBe(404);
    expect(hilang.body).toEqual({ error: 'Anggota tim tidak ditemukan' });
  });

  test('DELETE id valid → 200 { ok }; tidak ada → 404; id tidak sah → 400; tanpa token → 401', async () => {
    const { app } = createTestApp();
    expect((await kirimTulis(app, 'DELETE', '/api/tim/1', null)).status).toBe(200);
    expect((await kirimTulis(app, 'DELETE', '/api/tim/99', null)).status).toBe(404);
    expect((await kirimTulis(app, 'DELETE', '/api/tim/nol', null)).status).toBe(400);
    expect((await kirimTulis(app, 'DELETE', '/api/tim/1', null, null)).status).toBe(401);
  });
});

// — Lengkapi CRUD admin: testimoni & slide hero ————————————————————————

const BODY_TESTIMONI = { nama: 'Peserta', peran: 'Mahasiswa', kutipan: 'Materinya runtut.', urutan: 5 };

const BODY_HERO = {
  urutan: 3,
  eyebrow: 'AI Center',
  judul: 'Judul slide',
  judulAksen: 'Aksen',
  sub: 'Sub judul slide.',
  ctaPrimer: { label: 'Mulai', to: '/pelatihan' },
  ctaSekunder: { label: 'Kontak', to: 'https://example.test/kontak' },
  badgeJudul: 'Badge',
  badgeSub: 'Sub badge',
  image: '/uploads/hero-baru.jpg',
  srcSet: null,
  sizes: null,
  layout: 'image-left',
};

describe('CRUD /api/testimoni (lengkapi CRUD admin)', () => {
  test('POST valid → 201; urutan lama digeser +1 lalu testimoni baru jadi #1 (urutan 0)', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/testimoni', BODY_TESTIMONI);
    expect(status).toBe(201);
    expect((body as Testimoni).id).toBe(99);
    expect((body as Testimoni).urutan).toBe(0);
    expect(calls.find((c) => c.method === 'testimoni.naikkanSemuaUrutan')).toBeDefined();
    // Urutan dari body diabaikan saat tambah — selalu 0 (paling atas).
    expect(calls.find((c) => c.method === 'testimoni.create')?.args[0]).toEqual({ ...BODY_TESTIMONI, urutan: 0 });
    expect((await kirimTulis(app, 'POST', '/api/testimoni', BODY_TESTIMONI, null)).status).toBe(401);
  });

  test('POST kutipan kelewat batas → 400 eksplisit tanpa menyentuh repository', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/testimoni', {
      ...BODY_TESTIMONI,
      kutipan: 'x'.repeat(1001),
    });
    expect(status).toBe(400);
    expect(body).toEqual({ error: 'Field "kutipan" maksimal 1000 karakter' });
    expect(calls.some((c) => c.method === 'testimoni.create')).toBe(false);
  });

  test('PUT id valid → 200; id tidak sah → 400; tidak ada → 404; DELETE alur sama', async () => {
    const { app } = createTestApp();
    const ok = await kirimTulis(app, 'PUT', '/api/testimoni/7', BODY_TESTIMONI);
    expect(ok.status).toBe(200);
    expect((ok.body as Testimoni).nama).toBe('Peserta');

    expect((await kirimTulis(app, 'PUT', '/api/testimoni/abc', BODY_TESTIMONI)).status).toBe(400);
    expect((await kirimTulis(app, 'PUT', '/api/testimoni/99', BODY_TESTIMONI)).status).toBe(404);

    expect((await kirimTulis(app, 'DELETE', '/api/testimoni/7', null)).status).toBe(200);
    expect((await kirimTulis(app, 'DELETE', '/api/testimoni/99', null)).status).toBe(404);
    expect((await kirimTulis(app, 'DELETE', '/api/testimoni/7', null, null)).status).toBe(401);
  });
});

describe('CRUD /api/hero-slides (lengkapi CRUD admin)', () => {
  test('POST valid → 201 + repo.create menerima body penuh', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/hero-slides', BODY_HERO);
    expect(status).toBe(201);
    expect((body as HeroSlide).id).toBe(99);
    expect(calls.find((c) => c.method === 'hero.create')?.args[0]).toEqual(BODY_HERO);
    expect((await kirimTulis(app, 'POST', '/api/hero-slides', BODY_HERO, null)).status).toBe(401);
  });

  test('layout di luar enum → 400 eksplisit tanpa menyentuh repository', async () => {
    const { app, calls } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/hero-slides', {
      ...BODY_HERO,
      layout: 'tengah',
    });
    expect(status).toBe(400);
    expect(body).toEqual({
      error: 'Field "layout" harus salah satu dari: default, image-left, teks-kanan',
    });
    expect(calls.some((c) => c.method === 'hero.create')).toBe(false);
  });

  test('CTA bukan objek → 400 yang menyebut nama fieldnya', async () => {
    const { app } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/hero-slides', {
      ...BODY_HERO,
      ctaSekunder: 'masih diisi string lama',
    });
    expect(status).toBe(400);
    expect(body).toEqual({ error: 'Field "ctaSekunder" harus objek { label, to }' });
  });

  test('PUT id valid → 200 (field null di-omitted); tidak ada → 404; DELETE alur sama', async () => {
    const { app } = createTestApp();
    const ok = await kirimTulis(app, 'PUT', '/api/hero-slides/10', {
      ...BODY_HERO,
      image: null,
      srcSet: null,
      sizes: null,
      layout: null,
    });
    expect(ok.status).toBe(200);
    const json = JSON.stringify(ok.body);
    expect(json).not.toContain('image');
    expect(json).not.toContain('layout');

    expect((await kirimTulis(app, 'PUT', '/api/hero-slides/99', BODY_HERO)).status).toBe(404);
    expect((await kirimTulis(app, 'DELETE', '/api/hero-slides/10', null)).status).toBe(200);
    expect((await kirimTulis(app, 'DELETE', '/api/hero-slides/abc', null)).status).toBe(400);
  });
});

// — Prioritas 2: batas panjang terekspos di route ————————————————————————

describe('batas panjang field teks di route (Prioritas 2)', () => {
  test('POST /api/berita judul 201 karakter → 400 eksplisit', async () => {
    const { app } = createTestApp();
    const { status, body } = await kirimTulis(app, 'POST', '/api/berita', {
      ...BODY_BERITA,
      judul: 'x'.repeat(201),
    });
    expect(status).toBe(400);
    expect(body).toEqual({ error: 'Field "judul" maksimal 200 karakter' });
  });

  test('PUT /api/dokumentasi/:slug deskripsi 10001 karakter → 400 eksplisit', async () => {
    const { app } = createTestApp();
    const { status, body } = await kirimTulis(
      app,
      'PUT',
      `/api/dokumentasi/${sampleDok.slug}`,
      { ...BODY_DOK, deskripsi: 'x'.repeat(10_001) },
    );
    expect(status).toBe(400);
    expect(body).toEqual({ error: 'Field "deskripsi" maksimal 10000 karakter' });
  });
});

// — Prioritas 2: pembersih file gambar tidak terpakai ——————————————————————

describe('POST /api/admin/uploads/bersihkan (Prioritas 2)', () => {
  test('mode kering → lapor kandidat tanpa menghapus; nyata → file terhapus', async () => {
    const { mkdtemp, rm, writeFile, stat } = await import('node:fs/promises');
    const { tmpdir } = await import('node:os');
    const { join } = await import('node:path');
    const dir = await mkdtemp(join(tmpdir(), 'rute-orphan-'));
    const lama = process.env.UPLOAD_DIR;
    process.env.UPLOAD_DIR = dir;
    try {
      // Dua file: 'pakai.jpg' dirujuk fake referensiGambar, 'tidak-terpakai.jpg' tidak.
      await writeFile(join(dir, 'pakai.jpg'), 'x');
      await writeFile(join(dir, 'tidak-terpakai.jpg'), 'x');
      const { app } = createTestApp();

      // Kering: hanya lapor — kedua file tetap ada.
      const kering = await kirimTulis(app, 'POST', '/api/admin/uploads/bersihkan', {
        kering: true,
      });
      expect(kering.status).toBe(200);
      expect(kering.body).toEqual({ kering: true, items: ['tidak-terpakai.jpg'] });
      await stat(join(dir, 'tidak-terpakai.jpg'));

      // Nyata: file tidak terpakai terhapus, file terpakai tetap aman.
      const nyata = await kirimTulis(app, 'POST', '/api/admin/uploads/bersihkan', {});
      expect(nyata.status).toBe(200);
      expect(nyata.body).toEqual({ kering: false, items: ['tidak-terpakai.jpg'] });
      await expect(stat(join(dir, 'tidak-terpakai.jpg'))).rejects.toThrow();
      await stat(join(dir, 'pakai.jpg'));
    } finally {
      if (lama === undefined) delete process.env.UPLOAD_DIR;
      else process.env.UPLOAD_DIR = lama;
      await rm(dir, { recursive: true, force: true });
    }
  });

  test('tanpa token → 401', async () => {
    const { app } = createTestApp();
    const { status } = await kirimTulis(app, 'POST', '/api/admin/uploads/bersihkan', {}, null);
    expect(status).toBe(401);
  });
});
