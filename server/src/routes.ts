/**
 * Endpoint API CMS (dokumentasi, berita, kursus, tim, testimoni), progres,
 * auth admin, dan upload. Semua route memakai repository yang disuntikkan.
 *
 * Route tulis memakai token hasil POST /api/admin/login. Login dibatasi
 * rate-limit 5 gagal / 10 menit per IP. Semua field teks punya batas panjang.
 *
 * Kontrak daftar: { items: T[] }; detail tidak ada → 404 { error }.
 */
import type { Elysia } from 'elysia';
import { buatTokenAdmin, passwordCocok, tokenAdminValid, tokenDariHeader } from './auth';
import { buatRateLimitLogin, kunciIpDariRequest } from './lib/rateLimit';
import { normalizeSearch, parseOrder, parsePagination, trimOrNull } from './lib/query';
import {
  kodeFilterQuery,
  pesertaIdSah,
  slugDariJudul,
  validasiBerita,
  validasiDokumentasi,
  validasiKursus,
  validasiProgress,
  validasiTestimoni,
  validasiTim,
} from './lib/tulis';
import {
  createFilesystemImageStorage,
  namaFileAman,
  tipeKontenGambar,
  urlFileUnggahan,
  validasiFileGambar,
  type ImageStorage,
} from './lib/upload';
import { findOrphanUploads } from './lib/orphanUpload';
import type { Repositories } from './repositories/types';

/**
 * Rantai `.get()` mengubah generic Routes Elysia, sehingga helper komposisi
 * tidak bisa memakai tipe default-nya. Boundary ini memakai `any` disengaja
 * agar generic monster tidak menyebar ke seluruh kode.
 */
export type AnyElysia = Elysia<any, any, any, any, any, any, any>;

const ORDER_OPTIONS = ['asc', 'desc'] as const;

/** Pesan penolakan standar route tulis (401). */
const PESAN_BELUM_LOGIN = 'Belum login — ambil token dari POST /api/admin/login';

export function registerRoutes(
  app: AnyElysia,
  repos: Repositories,
  /** Password admin dari env ADMIN_PASSWORD — kunci verifikasi token tulis. */
  adminPassword: string,
  /** Override storage filesystem (khusus test); runtime utama memakai RustFS. */
  opsi?: { uploadDir?: string; storage?: ImageStorage },
): AnyElysia {
  /** True bila header Authorization membawa token valid untuk password ini. */
  const terautentikasi = (request: Request): boolean => {
    const token = tokenDariHeader(request.headers.get('Authorization'));
    return token !== null && tokenAdminValid(token, adminPassword);
  };
  const storage = opsi?.storage ?? createFilesystemImageStorage(opsi?.uploadDir);

  /** Path `:id` tim → integer positif; `null` bila tidak sah (→ 400). */
  const idDariParams = (mentah: string): number | null => {
    const id = Number(mentah);
    return Number.isInteger(id) && id > 0 ? id : null;
  };

  /** Batas brute-force login admin: 5 gagal / 10 menit per IP (fixed-window). */
  const limiterLogin = buatRateLimitLogin({ batasPercobaan: 5, jendelaMs: 10 * 60_000 });
  /**
   * Rate-limit tulis progres publik (POST /api/progress): 60 percobaan/menit
   * per IP. Pabriknya sama dengan limiter login (fixed-window in-memory) —
   * `catat` menghitung SETIAP percobaan yang diterima, jadi pertumbuhan baris
   * DB dari satu IP tetap terbatas meski endpoint ini tanpa token.
   */
  const limiterProgress = buatRateLimitLogin({ batasPercobaan: 60, jendelaMs: 60_000 });

  return app
    .get('/api/health', async ({ set }) => {
      try {
        await repos.ping();
      } catch {
        // Sengaja tidak di-log: endpoint ini bisa di-poll berkala oleh pemantau,
        // jadi saat DB mati log tidak ikut membanjiri. Sinyalnya = status 503.
        set.status = 503;
        return { error: 'Database tidak terjangkau' };
      }
      return { status: 'ok' as const, db: 'ok' as const };
    })

    // — Berita —————————————————————————————————————————————————
    .get('/api/berita', async ({ query }) => ({
      items: await repos.berita.list({
        q: normalizeSearch(query.q),
        pagination: parsePagination(query.page, query.limit),
        order: parseOrder(query.order, ORDER_OPTIONS, 'desc'),
      }),
    }))
    .get('/api/berita/:slug', async ({ params, set }) => {
      const item = await repos.berita.findBySlug(params.slug);
      if (item === null) {
        set.status = 404;
        return { error: 'Berita tidak ditemukan' };
      }
      return item;
    })

    // — Dokumentasi ————————————————————————————————————————————
    .get('/api/dokumentasi', async ({ query }) => ({
      items: await repos.dokumentasi.list({
        q: normalizeSearch(query.q),
        pagination: parsePagination(query.page, query.limit),
        order: parseOrder(query.order, ORDER_OPTIONS, 'desc'),
        kategori: trimOrNull(query.kategori),
      }),
    }))
    .get('/api/dokumentasi/:slug', async ({ params, set }) => {
      const item = await repos.dokumentasi.findBySlug(params.slug);
      if (item === null) {
        set.status = 404;
        return { error: 'Dokumentasi tidak ditemukan' };
      }
      return item;
    })

    // — Tim ———————————————————————————————————————————————————
    .get('/api/tim', async ({ query }) => ({
      items: await repos.tim.list({
        q: normalizeSearch(query.q),
        pagination: parsePagination(query.page, query.limit),
      }),
    }))


    // — Kursus (Pelatihan) —————————————————————————————————————
    .get('/api/kursus', async ({ query }) => ({
      items: await repos.kursus.list({
        q: normalizeSearch(query.q),
        pagination: parsePagination(query.page, query.limit),
      }),
    }))
    .get('/api/kursus/:kode', async ({ params, set }) => {
      const item = await repos.kursus.findByKode(params.kode);
      if (item === null) {
        set.status = 404;
        return { error: 'Kursus tidak ditemukan' };
      }
      return item;
    })

    // — Progres belajar peserta (LMS; PUBLIK tanpa akun) ————————————
    // Baca & simpan progres per (pesertaId browser, kode kursus). Tanpa token
    // karena peserta tidak punya akun — keamanan = pola ID ketat + batas ukuran
    // state (lib/tulis) + rate-limit tulis per IP.
    .get('/api/progress', async ({ query, set }) => {
      const pesertaId = pesertaIdSah(query.pesertaId);
      if (pesertaId === null) {
        set.status = 400;
        return { error: 'Query "pesertaId" wajib: 8-64 karakter (huruf/angka/-)' };
      }
      const kode = kodeFilterQuery(query.kode);
      if ('error' in kode) {
        set.status = 400;
        return { error: kode.error };
      }
      return { items: await repos.progress.list(pesertaId, kode.kode) };
    })
    .post('/api/progress', async ({ body, request, set }) => {
      // Rate-limit dulu: endpoint publik, jadi tiap percobaan dihitung per IP.
      const kunciIp = kunciIpDariRequest(request);
      const keputusan = limiterProgress.periksa(kunciIp);
      if (!keputusan.diizinkan) {
        set.status = 429;
        set.headers['Retry-After'] = String(keputusan.detikTersisa);
        return {
          error: `Terlalu banyak permintaan simpan progres — coba lagi dalam ${keputusan.detikTersisa} detik`,
        };
      }
      limiterProgress.catat(kunciIp);
      const hasil = validasiProgress(body);
      if (!hasil.ok) {
        set.status = 400;
        return { error: hasil.error };
      }
      // UPSERT per (peserta, kode) — entri tersimpan dikirim balik sebagai
      // konfirmasi (stempel waktu dari DB).
      return await repos.progress.simpan(hasil.data);
    })

    // — Testimoni CMS ————————————————————————————————————————————
    .get('/api/testimoni', async () => ({ items: await repos.testimoni.list() }))


    // — Admin: login sederhana (1 akun; password dari ADMIN_PASSWORD) ————
    .post('/api/admin/login', ({ body, request, set }) => {
      // Rate-limit dulu sebelum verifikasi: blokir tanpa membuka oracle password.
      const kunciIp = kunciIpDariRequest(request);
      const keputusan = limiterLogin.periksa(kunciIp);
      if (!keputusan.diizinkan) {
        set.status = 429;
        set.headers['Retry-After'] = String(keputusan.detikTersisa);
        return {
          error: `Terlalu banyak percobaan login — coba lagi dalam ${keputusan.detikTersisa} detik`,
        };
      }
      const input = (body ?? {}) as { password?: unknown };
      const password = typeof input.password === 'string' ? input.password : '';
      if (!passwordCocok(password, adminPassword)) {
        limiterLogin.catatGagal(kunciIp);
        set.status = 401;
        return { error: 'Password salah' };
      }
      limiterLogin.reset(kunciIp);
      return { token: buatTokenAdmin(adminPassword) };
    })

    // — Tulis dokumentasi (CRUD admin; butuh Authorization: Bearer) ——————
    .post('/api/dokumentasi', async ({ body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const hasil = validasiDokumentasi(body);
      if (!hasil.ok) {
        set.status = 400;
        return { error: hasil.error };
      }
      const slug = slugDariJudul(hasil.data.judul);
      if (slug === '') {
        set.status = 400;
        return { error: 'Judul tidak menghasilkan slug — tambahkan huruf/angka' };
      }
      if ((await repos.dokumentasi.findBySlug(slug)) !== null) {
        set.status = 409;
        return { error: 'Slug sudah dipakai — judul bentrok dengan konten lain' };
      }
      set.status = 201;
      return await repos.dokumentasi.create(slug, hasil.data);
    })
    .put('/api/dokumentasi/:slug', async ({ params, body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const hasil = validasiDokumentasi(body);
      if (!hasil.ok) {
        set.status = 400;
        return { error: hasil.error };
      }
      // Slug TIDAK diubah saat edit — tautan lama tetap hidup walau judul berubah.
      const item = await repos.dokumentasi.update(params.slug, hasil.data);
      if (item === null) {
        set.status = 404;
        return { error: 'Dokumentasi tidak ditemukan' };
      }
      return item;
    })
    .delete('/api/dokumentasi/:slug', async ({ params, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const terhapus = await repos.dokumentasi.remove(params.slug);
      if (!terhapus) {
        set.status = 404;
        return { error: 'Dokumentasi tidak ditemukan' };
      }
      return { ok: true as const };
    })

    // — Tulis berita (CRUD admin; butuh Authorization: Bearer) ——————————
    .post('/api/berita', async ({ body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const hasil = validasiBerita(body);
      if (!hasil.ok) {
        set.status = 400;
        return { error: hasil.error };
      }
      const slug = slugDariJudul(hasil.data.judul);
      if (slug === '') {
        set.status = 400;
        return { error: 'Judul tidak menghasilkan slug — tambahkan huruf/angka' };
      }
      if ((await repos.berita.findBySlug(slug)) !== null) {
        set.status = 409;
        return { error: 'Slug sudah dipakai — judul bentrok dengan konten lain' };
      }
      set.status = 201;
      return await repos.berita.create(slug, hasil.data);
    })
    .put('/api/berita/:slug', async ({ params, body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const hasil = validasiBerita(body);
      if (!hasil.ok) {
        set.status = 400;
        return { error: hasil.error };
      }
      const item = await repos.berita.update(params.slug, hasil.data);
      if (item === null) {
        set.status = 404;
        return { error: 'Berita tidak ditemukan' };
      }
      return item;
    })
    .delete('/api/berita/:slug', async ({ params, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const terhapus = await repos.berita.remove(params.slug);
      if (!terhapus) {
        set.status = 404;
        return { error: 'Berita tidak ditemukan' };
      }
      return { ok: true as const };
    })

    // — Tulis kursus (CRUD admin; kunci = kode, mis. 'R01') ————————————
    .post('/api/kursus', async ({ body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const hasil = validasiKursus(body);
      if (!hasil.ok) {
        set.status = 400;
        return { error: hasil.error };
      }
      if ((await repos.kursus.findByKode(hasil.data.kode)) !== null) {
        set.status = 409;
        return { error: 'Kode sudah dipakai — gunakan kode lain' };
      }
      set.status = 201;
      return await repos.kursus.create(hasil.data);
    })
    .put('/api/kursus/:kode', async ({ params, body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const hasil = validasiKursus(body);
      if (!hasil.ok) {
        set.status = 400;
        return { error: hasil.error };
      }
      // Kode kunci tidak boleh diganti saat edit — samakan dengan path.
      if (hasil.data.kode !== params.kode.trim().toUpperCase()) {
        set.status = 400;
        return { error: 'Kode tidak boleh diganti — samakan dengan kode di URL' };
      }
      const item = await repos.kursus.update(params.kode, hasil.data);
      if (item === null) {
        set.status = 404;
        return { error: 'Kursus tidak ditemukan' };
      }
      return item;
    })
    .delete('/api/kursus/:kode', async ({ params, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const terhapus = await repos.kursus.remove(params.kode);
      if (!terhapus) {
        set.status = 404;
        return { error: 'Kursus tidak ditemukan' };
      }
      return { ok: true as const };
    })


    // — Tulis tim (CRUD admin; kunci = id numerik; Prioritas 2) ————————————
    .post('/api/tim', async ({ body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const hasil = validasiTim(body);
      if (!hasil.ok) {
        set.status = 400;
        return { error: hasil.error };
      }
      // Item baru selalu paling atas; geser + insert berjalan atomik.
      set.status = 201;
      return await repos.tim.createAtas({ ...hasil.data, urutan: 0 });
    })
    .put('/api/tim/:id', async ({ params, body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const id = idDariParams(params.id);
      if (id === null) {
        set.status = 400;
        return { error: 'Id tim tidak valid — isi bilangan bulat positif' };
      }
      const hasil = validasiTim(body);
      if (!hasil.ok) {
        set.status = 400;
        return { error: hasil.error };
      }
      const anggota = await repos.tim.update(id, hasil.data);
      if (anggota === null) {
        set.status = 404;
        return { error: 'Anggota tim tidak ditemukan' };
      }
      return anggota;
    })
    .delete('/api/tim/:id', async ({ params, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const id = idDariParams(params.id);
      if (id === null) {
        set.status = 400;
        return { error: 'Id tim tidak valid — isi bilangan bulat positif' };
      }
      const terhapus = await repos.tim.remove(id);
      if (!terhapus) {
        set.status = 404;
        return { error: 'Anggota tim tidak ditemukan' };
      }
      return { ok: true as const };
    })

    // — Tulis testimoni (CRUD admin; kunci = id numerik) ——————————————————
    .post('/api/testimoni', async ({ body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const hasil = validasiTestimoni(body);
      if (!hasil.ok) {
        set.status = 400;
        return { error: hasil.error };
      }
      // Geser + insert berada dalam satu transaksi agar dua admin tidak kehilangan urutan.
      set.status = 201;
      return await repos.testimoni.createAtas({ ...hasil.data, urutan: 0 });
    })
    .put('/api/testimoni/:id', async ({ params, body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const id = idDariParams(params.id);
      if (id === null) {
        set.status = 400;
        return { error: 'Id testimoni tidak valid — isi bilangan bulat positif' };
      }
      const hasil = validasiTestimoni(body);
      if (!hasil.ok) {
        set.status = 400;
        return { error: hasil.error };
      }
      const item = await repos.testimoni.update(id, hasil.data);
      if (item === null) {
        set.status = 404;
        return { error: 'Testimoni tidak ditemukan' };
      }
      return item;
    })
    .delete('/api/testimoni/:id', async ({ params, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const id = idDariParams(params.id);
      if (id === null) {
        set.status = 400;
        return { error: 'Id testimoni tidak valid — isi bilangan bulat positif' };
      }
      const terhapus = await repos.testimoni.remove(id);
      if (!terhapus) {
        set.status = 404;
        return { error: 'Testimoni tidak ditemukan' };
      }
      return { ok: true as const };
    })


    // — Bersihkan file unggahan tidak terpakai (Prioritas 2; butuh Bearer) ——
    // Body `{ kering: true }` = mode uji: hanya melaporkan kandidat tanpa
    // menghapus apa pun. Tanpa body / `kering: false` = hapus betulan;
    // `items` = daftar file (kandidat saat kering, terhapus saat nyata).
    .post('/api/admin/uploads/bersihkan', async ({ body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const kering = (body as { kering?: unknown } | null)?.kering === true;
      const daftar = await storage.list();
      const referensi = await repos.referensiGambar();
      const takTerpakai = findOrphanUploads(daftar, referensi);
      if (!kering) {
        for (const nama of takTerpakai) {
          await storage.delete(nama);
        }
      }
      return { kering, items: takTerpakai };
    })

    // — Unggah gambar admin (multipart `gambar`; butuh Bearer) —————————
    .post('/api/admin/upload', async ({ body, request, set }) => {
      if (!terautentikasi(request)) {
        set.status = 401;
        return { error: PESAN_BELUM_LOGIN };
      }
      const file = (body as { gambar?: unknown } | null)?.gambar;
      const valid = validasiFileGambar(file);
      if (!valid.ok) {
        set.status = 400;
        return { error: valid.error };
      }
      const nama = await storage.save(file as File, valid.ext);
      set.status = 201;
      return { url: urlFileUnggahan(nama) };
    })

    // — Sajikan file unggahan (publik, tanpa auth) —————————————————————
    .get('/uploads/:nama', async ({ params, set }) => {
      const aman = namaFileAman(params.nama);
      if (aman === null) {
        set.status = 404;
        return { error: 'File tidak ditemukan' };
      }
      const gambar = await storage.get(aman);
      if (gambar === null) {
        set.status = 404;
        return { error: 'File tidak ditemukan' };
      }
      return new Response(gambar.body, {
        headers: {
          'Content-Type': tipeKontenGambar(aman.slice(aman.lastIndexOf('.') + 1)),
          'Cache-Control': 'public, max-age=31536000, immutable',
        },
      });
    });
}
