/* Tes klien API nyata (frontend-final/src/api/api.ts) dengan fetch palsu.
   Jalankan: bun test scripts/tests/api.test.ts   (atau npm run test:api)
   api.ts mengimpor ../shared/types, jadi keduanya disalin ke folder sementara dengan tata letak src/ yang sama. */
import { beforeAll, beforeEach, describe, expect, test } from 'bun:test';
import { cpSync, mkdirSync, mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';

const akar = resolve(import.meta.dir, '..', '..');
let api: any;
let ApiError: any;

type Req = { url: string; metode: string; headers: Record<string, string>; body?: any; json?: any; keepalive: boolean };
let log: Req[] = [];
let penangan: (r: Req) => Response | Promise<Response> = () => balas(204);

const balas = (status: number, isi?: unknown, header: Record<string, string> = {}): Response =>
  new Response(isi === undefined ? null : JSON.stringify(isi), { status, headers: { 'Content-Type': 'application/json', ...header } });

beforeAll(async () => {
  const tmp = mkdtempSync(join(tmpdir(), 'api-klien-'));
  mkdirSync(join(tmp, 'src', 'api'), { recursive: true });
  mkdirSync(join(tmp, 'src', 'shared'), { recursive: true });
  cpSync(join(akar, 'frontend-final', 'src', 'api', 'api.ts'), join(tmp, 'src', 'api', 'api.ts'));
  cpSync(join(akar, 'frontend-final', 'src', 'shared', 'types.ts'), join(tmp, 'src', 'shared', 'types.ts'));
  ({ api } = await import(join(tmp, 'src', 'api', 'api.ts')));
  ({ ApiError } = await import(join(tmp, 'src', 'shared', 'types.ts')));
  globalThis.fetch = (async (u: any, opts: any = {}) => {
    const req: Req = { url: String(u), metode: opts.method ?? 'GET', headers: opts.headers ?? {}, body: opts.body, keepalive: Boolean(opts.keepalive) } as Req;
    if (typeof req.body === 'string') { try { req.json = JSON.parse(req.body); } catch { /* bukan JSON */ } }
    log.push(req);
    return penangan(req);
  }) as typeof fetch;
});
beforeEach(() => { log = []; penangan = () => balas(204); });

const sesiAdmin = (tambah: (r: Req) => Response) => (r: Req) => (r.url === '/api/auth/me' ? balas(200, { nama: 'Admin', username: 'admin.dev', role: 'pegawai', csrfToken: 'tk' }) : tambah(r));
const beritaAdmin = (o: object = {}) => ({ id: 1, slug: 's', judul: 'Judul', ringkasan: 'R', isi: 'Isi', tanggalPublikasi: '09-11-2026', gambar: { key: 'berita/k.png', url: '/media/berita/k.png' }, status: 'draft', version: 3, dibuatPada: '2026-11-09T00:00:00Z', dibuatOleh: 'u', diubahOleh: 'u', ...o });
const gagalDengan = async (p: Promise<unknown>) => { try { await p; } catch (g) { return g as any; } throw new Error('seharusnya melempar galat'); };

describe('publik', () => {
  test('daftar berita: query, tanggal DD-MM-YYYY apa adanya, isi = ringkasan, gambar jadi key', async () => {
    penangan = () => balas(200, { items: [{ id: 1, slug: 'a-b', judul: 'A B', ringkasan: 'Ringkas.', gambar: '/media/berita/x.png', tanggalPublikasi: '09-11-2026' }], page: 2, limit: 9, total: 12 });
    const d = await api.daftarBerita({ q: 'ai', page: 2, limit: 9 });
    expect(log[0].url).toBe('/api/berita?q=ai&page=2&limit=9');
    expect(d.items[0]).toMatchObject({ id: 1, slug: 'a-b', judul: 'A B', isi: 'Ringkas.', tanggalPublikasi: '09-11-2026', gambar: 'berita/x.png' });
    expect([d.page, d.total]).toEqual([2, 12]);
    expect(log[0].headers['X-CSRF-Token']).toBeUndefined();
  });

  test('dokumentasi: daftar (gambarUtama + jumlahGambar) dan detail (url -> key)', async () => {
    penangan = (r) => (r.url.startsWith('/api/dokumentasi/')
      ? balas(200, { id: 2, slug: 'k', judul: 'K', tanggalPublikasi: '02-01-2026', gambar: [{ url: '/media/dokumentasi/a.png' }, { url: '/media/dokumentasi/b.png' }] })
      : balas(200, { items: [{ id: 2, slug: 'k', judul: 'K', tanggalPublikasi: '02-01-2026', gambarUtama: '/media/dokumentasi/a.png', jumlahGambar: 3 }, { id: 3, slug: 'z', judul: 'Z', tanggalPublikasi: '03-01-2026', jumlahGambar: 0 }], page: 1, limit: 9, total: 2 }));
    const l = await api.daftarDokumentasi({});
    expect(l.items[0].gambar.length).toBe(3);
    expect(l.items[0].gambar[0].key).toBe('dokumentasi/a.png');
    expect(l.items[1].gambar).toEqual([]);
    const d = await api.detailDokumentasi('k');
    expect(d.gambar.map((g: any) => g.key)).toEqual(['dokumentasi/a.png', 'dokumentasi/b.png']);
    expect(log.at(-1)!.url).toBe('/api/dokumentasi/k');
  });

  test('slug di-encode; testimoni dan tim dipetakan', async () => {
    penangan = (r) => (r.url === '/api/tim'
      ? balas(200, { items: [{ id: 1, nama: 'Ani', posisi: 'Dosen', bidangRiset: 'CV', foto: '/media/tim/a.png', urutan: 4 }] })
      : r.url === '/api/testimoni' ? balas(200, { items: [{ id: 5, nama: 'Budi', teks: 'Bagus', posisi: 'Mhs' }] })
        : balas(200, { id: 1, slug: 'a b', judul: 'J', isi: 'I', tanggalPublikasi: '01-01-2026', gambar: '/media/berita/x.png' }));
    const tim = await api.daftarTim();
    expect(tim.items[0]).toMatchObject({ bidangRiset: 'CV', foto: 'tim/a.png', urutan: 4 });
    expect((await api.daftarTestimoni()).items[0].foto).toBeNull();
    await api.detailBerita('a b');
    expect(log.at(-1)!.url).toBe('/api/berita/a%20b');
  });
});

describe('galat', () => {
  test('404, 429 + Retry-After, 500 tanpa badan, jaringan', async () => {
    penangan = () => balas(404, { error: { code: 'not_found', message: 'Sumber tidak ditemukan' } });
    let g = await gagalDengan(api.detailBerita('x'));
    expect(g).toBeInstanceOf(ApiError);
    expect([g.status, g.error.code]).toEqual([404, 'not_found']);
    penangan = () => balas(429, { error: { code: 'rate_limited', message: 'Terlalu banyak permintaan' } }, { 'Retry-After': '42' });
    g = await gagalDengan(api.detailBerita('x'));
    expect([g.status, g.retryAfter, g.error.code]).toEqual([429, 42, 'rate_limited']);
    penangan = () => balas(500);
    g = await gagalDengan(api.detailBerita('x'));
    expect(g.status).toBe(500);
    expect(g.error.message).toMatch(/kendala/i);
    penangan = () => { throw new TypeError('gagal'); };
    g = await gagalDengan(api.detailBerita('x'));
    expect([g.status, g.error.code]).toEqual([0, 'jaringan']);
  });
});

describe('sesi dan CSRF', () => {
  test('401 -> null; 200 -> token CSRF dikirim pada tulis, bukan pada baca; keluar menghapus token', async () => {
    penangan = () => balas(401, { error: { code: 'unauthenticated', message: 'Belum login' } });
    expect(await api.siapkan()).toBeNull();
    expect(api.akun()).toBeNull();
    penangan = sesiAdmin(() => balas(204));
    expect(await api.siapkan()).toMatchObject({ nama: 'Admin', username: 'admin.dev', role: 'pegawai' });
    await api.adminHapus('tim', 3);
    const hapus = log.at(-1)!;
    expect([hapus.metode, hapus.url, hapus.headers['X-CSRF-Token']]).toEqual(['DELETE', '/api/admin/tim/3', 'tk']);
    expect(log[0].headers['X-CSRF-Token']).toBeUndefined();
    await api.keluar();
    expect([log.at(-1)!.url, log.at(-1)!.headers['X-CSRF-Token']]).toEqual(['/api/auth/logout', 'tk']);
    expect(api.akun()).toBeNull();
    await api.adminHapus('tim', 4).catch(() => {});
    expect(log.at(-1)!.headers['X-CSRF-Token']).toBeUndefined();
  });

  test('galat selain 401 pada /me dilempar ulang', async () => {
    penangan = () => balas(500);
    expect((await gagalDengan(api.siapkan())).status).toBe(500);
  });
});

describe('admin', () => {
  test('adminDaftar: status di query; tim tanpa query; gambar jadi key; dikunciOleh dan deletedAt dipetakan', async () => {
    penangan = sesiAdmin((r) => (r.url.startsWith('/api/admin/berita')
      ? balas(200, { items: [beritaAdmin({ dikunciOleh: 'Rani' }), beritaAdmin({ id: 2, dihapusPada: '2026-11-10T00:00:00Z' })], page: 1, limit: 100000, total: 2 })
      : r.url.startsWith('/api/admin/tim') ? balas(200, { items: [{ id: 1, nama: 'Ani', posisi: 'D', bidangRiset: 'CV', foto: { key: 'tim/a.png', url: '/media/tim/a.png' }, urutan: 1, version: 1 }] })
        : balas(200, { items: [{ id: 4, nama: 'B', teks: 'T', posisi: 'P', foto: null, status: 'posted', version: 2 }] })));
    await api.siapkan();
    let d = await api.adminDaftar('berita');
    expect(log.at(-1)!.url).toBe('/api/admin/berita?status=semua');
    expect(d.items[0]).toMatchObject({ gambar: 'berita/k.png', tanggalPublikasi: '09-11-2026', version: 3, dikunciOleh: 'Rani', deletedAt: null });
    expect(d.items[1]).toMatchObject({ deletedAt: '2026-11-10T00:00:00Z', dikunciOleh: null });
    await api.adminDaftar('berita', { trash: true });
    expect(log.at(-1)!.url).toBe('/api/admin/berita?status=trash');
    d = await api.adminDaftar('tim');
    expect(log.at(-1)!.url).toBe('/api/admin/tim');
    expect(d.items[0]).toMatchObject({ foto: 'tim/a.png', urutan: 1, bidangRiset: 'CV' });
    d = await api.adminDaftar('testimoni', { trash: true });
    expect(log.at(-1)!.url).toBe('/api/admin/testimoni?status=trash');
    expect(d.items[0].foto).toBeNull();
  });

  test('adminDaftar dokumentasi: ringkasan (gambarUtama + jumlahGambar) jadi daftar gambar', async () => {
    penangan = sesiAdmin(() => balas(200, { items: [{ id: 7, slug: 'd', judul: 'D', tanggalPublikasi: '01-01-2026', gambarUtama: { key: 'dokumentasi/a.png', url: '/media/dokumentasi/a.png' }, jumlahGambar: 2, status: 'draft', version: 1 }] }));
    await api.siapkan();
    const d = await api.adminDaftar('dokumentasi');
    expect(d.items[0].gambar).toHaveLength(2);
    expect(d.items[0].gambar[0].key).toBe('dokumentasi/a.png');
  });

  test('adminTambah berita: unggah blob dulu (multipart sumber+gambar), lalu POST tanpa version dengan key hasil unggah', async () => {
    const blobUrl = URL.createObjectURL(new Blob(['x'], { type: 'image/png' }));
    const asli = globalThis.fetch;
    penangan = sesiAdmin((r) => (r.url === '/api/admin/upload' ? balas(201, { key: 'berita/baru.png', url: '/media/berita/baru.png' }) : balas(201, beritaAdmin({ id: 9 }))));
    await api.siapkan();
    // fetch(blob:) untuk membaca berkas lokal: dijawab oleh fetch palsu dengan Blob
    globalThis.fetch = (async (u: any, o: any) => (String(u).startsWith('blob:') ? new Response(new Blob(['x'], { type: 'image/png' })) : asli(u, o))) as typeof fetch;
    try {
      const hasil = await api.adminTambah('berita', { judul: 'Judul Baru', isi: 'Isi', tanggalPublikasi: '09-11-2026', gambar: blobUrl });
      const up = log.find((r) => r.url === '/api/admin/upload')!;
      expect(up.metode).toBe('POST');
      expect(up.body.get('sumber')).toBe('berita');
      expect(up.body.get('gambar')).toBeInstanceOf(Blob);
      expect(up.headers['X-CSRF-Token']).toBe('tk');
      expect(up.headers['Content-Type']).toBeUndefined(); // multipart: boundary diisi fetch
      const buat = log.at(-1)!;
      expect([buat.metode, buat.url]).toEqual(['POST', '/api/admin/berita']);
      expect(buat.json).toEqual({ judul: 'Judul Baru', isi: 'Isi', tanggalPublikasi: '09-11-2026', gambar: 'berita/baru.png' });
      expect(hasil.id).toBe(9);
    } finally { globalThis.fetch = asli; }
  });

  test('adminUbah dokumentasi: kirim gambarDaftar + version; key lama tidak diunggah; blob yang sama tidak diunggah dua kali saat coba ulang', async () => {
    const blobUrl = URL.createObjectURL(new Blob(['x'], { type: 'image/png' }));
    const asli = globalThis.fetch;
    let tolak = true;
    penangan = sesiAdmin((r) => {
      if (r.url === '/api/admin/upload') return balas(201, { key: 'dokumentasi/n.png', url: '/media/dokumentasi/n.png' });
      if (tolak) { tolak = false; return balas(409, { error: { code: 'conflict', message: 'Data sudah diubah oleh pengguna lain' } }); }
      return balas(200, { id: 5, slug: 'd', judul: 'D', tanggalPublikasi: '09-11-2026', gambar: [], status: 'draft', version: 4 });
    });
    await api.siapkan();
    globalThis.fetch = (async (u: any, o: any) => (String(u).startsWith('blob:') ? new Response(new Blob(['x'], { type: 'image/png' })) : asli(u, o))) as typeof fetch;
    try {
      const data = { judul: 'D', tanggalPublikasi: '09-11-2026', gambarDaftar: [{ key: 'dokumentasi/lama.png' }, { key: blobUrl }] };
      const g = await gagalDengan(api.adminUbah('dokumentasi', 5, data, 3));
      expect([g.status, g.error.code]).toEqual([409, 'conflict']);
      await api.adminUbah('dokumentasi', 5, data, 3);
      expect(log.filter((r) => r.url === '/api/admin/upload')).toHaveLength(1);
      const put = log.at(-1)!;
      expect([put.metode, put.url]).toEqual(['PUT', '/api/admin/dokumentasi/5']);
      expect(put.json).toEqual({ judul: 'D', tanggalPublikasi: '09-11-2026', gambarDaftar: [{ key: 'dokumentasi/lama.png' }, { key: 'dokumentasi/n.png' }], version: 3 });
    } finally { globalThis.fetch = asli; }
  });

  test('testimoni tanpa foto mengirim foto:null; tim memakai bidangRiset dan mengirim urutan angka', async () => {
    penangan = sesiAdmin((r) => balas(201, r.url.endsWith('/tim') ? { id: 2, nama: 'A', posisi: 'P', bidangRiset: 'CV', foto: { key: 'tim/a.png', url: '/media/tim/a.png' }, urutan: 3, version: 1 } : { id: 1, nama: 'A', teks: 'T', posisi: 'P', status: 'draft', version: 1 }));
    await api.siapkan();
    await api.adminTambah('testimoni', { nama: 'A', teks: 'T', posisi: 'P', foto: null });
    expect(log.at(-1)!.json).toEqual({ nama: 'A', teks: 'T', posisi: 'P', foto: null });
    await api.adminTambah('tim', { nama: 'A', posisi: 'P', bidangRiset: 'CV', foto: 'tim/a.png', urutan: '7' });
    expect(log.at(-1)!.json).toEqual({ nama: 'A', posisi: 'P', bidangRiset: 'CV', foto: 'tim/a.png', urutan: 7 });
  });

  test('post/retract membawa {version} terakhir yang dilihat (atau diambil dari server); restore, hapus, hapus permanen tanpa body', async () => {
    penangan = sesiAdmin((r) => (r.metode === 'DELETE' ? balas(204) : balas(200, beritaAdmin({ version: 3 }))));
    await api.siapkan();
    await api.adminDetail('berita', 1); // mengingat version 3
    await api.adminPost('berita', 1);
    expect([log.at(-1)!.metode, log.at(-1)!.url, log.at(-1)!.json]).toEqual(['POST', '/api/admin/berita/1/post', { version: 3 }]);
    const n = log.length;
    await api.adminRetract('testimoni', 44); // belum pernah dilihat: GET dulu
    expect(log[n].url).toBe('/api/admin/testimoni/44');
    expect([log.at(-1)!.url, log.at(-1)!.json]).toEqual(['/api/admin/testimoni/44/retract', { version: 3 }]);
    await api.adminPulihkan('berita', 1);
    expect([log.at(-1)!.metode, log.at(-1)!.url, log.at(-1)!.body]).toEqual(['POST', '/api/admin/berita/1/restore', null]);
    await api.adminHapus('berita', 1);
    expect([log.at(-1)!.metode, log.at(-1)!.url]).toEqual(['DELETE', '/api/admin/berita/1']);
    await api.adminHapusPermanen('berita', 1);
    expect([log.at(-1)!.metode, log.at(-1)!.url]).toEqual(['DELETE', '/api/admin/berita/1/permanen']);
  });

  test('sumber tidak dikenal -> 404 tanpa memanggil server', async () => {
    const g = await gagalDengan(api.adminDaftar('rahasia'));
    expect(g.status).toBe(404);
    expect(log).toHaveLength(0);
  });

  test('log aktivitas: query diteruskan, usernames dibawa', async () => {
    penangan = sesiAdmin(() => balas(200, { items: [{ id: 5, jenis: 'ubah' }], page: 2, limit: 10, total: 11, usernames: ['admin.dev'] }));
    await api.siapkan();
    const d = await api.adminLog({ page: 2, username: 'admin.dev', jenis: 'ubah', dari: '2026-11-01' });
    expect(log.at(-1)!.url).toBe('/api/admin/log-aktivitas?page=2&username=admin.dev&jenis=ubah&dari=2026-11-01');
    expect([d.page, d.total, d.usernames]).toEqual([2, 11, ['admin.dev']]);
  });
});

describe('soft lock', () => {
  test('pasang (POST), heartbeat (PUT), lepas (DELETE keepalive); 423 locked -> milik lain; tanpa kunci -> 410 lock_lost', async () => {
    let dipegangLain = false;
    penangan = sesiAdmin((r) => {
      if (r.url.endsWith('/lock') && r.metode === 'POST') return dipegangLain ? balas(423, { error: { code: 'locked', message: 'Konten sedang diedit oleh Rani', pemegang: { nama: 'Rani' } } }) : balas(200, { pemegang: { nama: 'Admin' }, milikSaya: true, heartbeatDetik: 45 });
      return balas(204);
    });
    await api.siapkan();
    const hilang = await gagalDengan(api.kunciHeartbeat());
    expect([hilang.status, hilang.error.code]).toEqual([410, 'lock_lost']);
    expect(await api.kunciAmbil('berita', 7)).toEqual({ milik: 'saya' });
    expect(log.at(-1)!.url).toBe('/api/admin/berita/7/lock');
    await api.kunciHeartbeat();
    expect([log.at(-1)!.metode, log.at(-1)!.url]).toEqual(['PUT', '/api/admin/berita/7/lock']);
    api.kunciLepas();
    await Bun.sleep(20);
    expect([log.at(-1)!.metode, log.at(-1)!.url, log.at(-1)!.keepalive]).toEqual(['DELETE', '/api/admin/berita/7/lock', true]);
    dipegangLain = true;
    expect(await api.kunciAmbil('berita', 7)).toEqual({ milik: 'lain', nama: 'Rani' });
    const n = log.length;
    api.kunciLepas();
    await Bun.sleep(20);
    expect(log).toHaveLength(n); // tidak memegang kunci: tidak melepas
  });

  test('galat selain 423 locked dilempar ulang', async () => {
    penangan = sesiAdmin(() => balas(500));
    await api.siapkan();
    expect((await gagalDengan(api.kunciAmbil('berita', 1))).status).toBe(500);
  });
});
