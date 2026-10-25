import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ErrorApi, ambilDaftar, ambilJson, bangunUrlApi, harusFetch, kirimFileAdmin, kirimJsonAdmin, resolveBaseUrl } from './api.ts';

const BASE = 'http://api.test';

/** Stub fetch yang mengembalikan respons JSON sukses dengan status tertentu. */
function stubFetch(body: unknown, status = 200): ReturnType<typeof vi.fn> {
  const fn = vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  });
  vi.stubGlobal('fetch', fn);
  return fn;
}

describe('resolveBaseUrl', () => {
  it('undefined → /api; kosong → ""; whitespace & slash akhir dibersihkan', () => {
    expect(resolveBaseUrl(undefined)).toBe('/api');
    expect(resolveBaseUrl('')).toBe('');
    expect(resolveBaseUrl('   ')).toBe('');
    expect(resolveBaseUrl('  http://localhost:3000/api/  ')).toBe('http://localhost:3000/api');
    expect(resolveBaseUrl('http://localhost:3000/api///')).toBe('http://localhost:3000/api');
  });
});

describe('harusFetch', () => {
  it('hanya fetch bila base terisi dan bukan mode test', () => {
    expect(harusFetch('', 'development')).toBe(false);
    expect(harusFetch(BASE, 'test')).toBe(false);
    expect(harusFetch(BASE, 'development')).toBe(true);
    expect(harusFetch(BASE, 'production')).toBe(true);
  });
});

describe('bangunUrlApi', () => {
  it('menggabungkan base + path', () => {
    expect(bangunUrlApi(BASE, '/berita')).toBe('http://api.test/berita');
    expect(bangunUrlApi('http://api.test/api', '/kursus/R01')).toBe(
      'http://api.test/api/kursus/R01',
    );
  });

  it('edge: base kosong atau path tanpa "/" → error eksplisit', () => {
    expect(() => bangunUrlApi('', '/berita')).toThrow('base URL kosong');
    expect(() => bangunUrlApi(BASE, 'berita')).toThrow('diawali');
  });
});

describe('ambilJson', () => {
  beforeEach(() => {
    vi.spyOn(console, 'warn').mockImplementation(() => undefined);
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('HTTP 200 objek → data dikembalikan; URL terpasang benar', async () => {
    const fetchMock = stubFetch({ slug: 'berita-1' });
    const hasil = await ambilJson('/berita/berita-1', { slug: 'fallback' }, BASE);
    expect(hasil).toEqual({ slug: 'berita-1' });
    expect(fetchMock).toHaveBeenCalledWith('http://api.test/berita/berita-1');
  });

  it('HTTP non-2xx → fallback + peringatan dev (tanpa console.error)', async () => {
    stubFetch({ error: 'x' }, 500);
    const hasil = await ambilJson('/kursus/R01', { kode: 'fallback' }, BASE);
    expect(hasil).toEqual({ kode: 'fallback' });
    expect(console.warn).toHaveBeenCalled();
  });

  it('respons non-objek (null/angka) → dianggap tidak valid → fallback', async () => {
    stubFetch(null);
    expect(await ambilJson('/profil', { nama: 'fb' }, BASE)).toEqual({ nama: 'fb' });
    stubFetch(42);
    expect(await ambilJson('/profil', { nama: 'fb' }, BASE)).toEqual({ nama: 'fb' });
  });

  it('edge: fetch melempar (network error) → fallback, tidak melempar ke pemanggil', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('offline')));
    const hasil = await ambilJson('/profil', { nama: 'fb' }, BASE);
    expect(hasil).toEqual({ nama: 'fb' });
  });
});

describe('ambilDaftar', () => {
  beforeEach(() => {
    vi.spyOn(console, 'warn').mockImplementation(() => undefined);
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('respons { items: [...] } → array item', async () => {
    stubFetch({ items: [{ slug: 'a' }, { slug: 'b' }] });
    const hasil = await ambilDaftar('/berita', [{ slug: 'fallback' }], BASE);
    expect(hasil).toEqual([{ slug: 'a' }, { slug: 'b' }]);
  });

  it('bentuk respons salah (tanpa items) → fallback + peringatan dev', async () => {
    stubFetch({ data: 'bukan-items' });
    const hasil = await ambilDaftar('/berita', [{ slug: 'fallback' }], BASE);
    expect(hasil).toEqual([{ slug: 'fallback' }]);
    expect(console.warn).toHaveBeenCalled();
  });

  it('items bukan array → fallback (validasi bentuk eksplisit)', async () => {
    stubFetch({ items: 'bukan-array' });
    expect(await ambilDaftar('/berita', [{ slug: 'fb' }], BASE)).toEqual([{ slug: 'fb' }]);
  });

  it('edge: base kosong → fallback + peringatan tercatat (tanpa request)', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const hasil = await ambilDaftar('/berita', [{ slug: 'fb' }], '');
    expect(hasil).toEqual([{ slug: 'fb' }]);
    expect(fetchMock).not.toHaveBeenCalled();
    expect(console.warn).toHaveBeenCalled();
  });
});

describe('kirimJsonAdmin', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  const WRITE_OPTS = { method: 'POST' as const, baseUrl: BASE };

  it('sukses → data dikembalikan; URL/method/body/token terpasang benar', async () => {
    const fetchMock = stubFetch({ slug: 'baru' }, 201);
    const hasil = await kirimJsonAdmin('/dokumentasi', {
      ...WRITE_OPTS,
      body: { judul: 'Baru' },
      token: 'tok123',
    });
    expect(hasil).toEqual({ slug: 'baru' });
    expect(fetchMock).toHaveBeenCalledWith('http://api.test/dokumentasi', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer tok123',
      },
      body: JSON.stringify({ judul: 'Baru' }),
    });
  });

  it('tanpa token → tanpa header Authorization', async () => {
    const fetchMock = stubFetch({ ok: true });
    await kirimJsonAdmin('/dokumentasi/x/hapus', { method: 'DELETE', baseUrl: BASE });
    const init = fetchMock.mock.calls[0][1] as { headers: Record<string, string> };
    expect(init.headers.Authorization).toBeUndefined();
    expect(fetchMock.mock.calls[0][1]).toMatchObject({ method: 'DELETE', body: undefined });
  });

  it('HTTP non-2xx → melempar ErrorApi dengan pesan field `error` server + status', async () => {
    stubFetch({ error: 'Password salah' }, 401);
    const menunggu = kirimJsonAdmin('/admin/login', { ...WRITE_OPTS, body: {} });
    await expect(menunggu).rejects.toBeInstanceOf(ErrorApi);
    await expect(
      kirimJsonAdmin('/admin/login', { ...WRITE_OPTS, body: {} }),
    ).rejects.toMatchObject({ message: 'Password salah', status: 401 });
  });

  it('HTTP non-2xx tanpa field error → pesan `HTTP <status>`', async () => {
    stubFetch(null, 500);
    await expect(kirimJsonAdmin('/dokumentasi', WRITE_OPTS)).rejects.toMatchObject({
      message: 'HTTP 500',
      status: 500,
    });
  });

  it('fetch melempar (network) → ErrorApi status 0, bukan error liar', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('offline')));
    await expect(kirimJsonAdmin('/dokumentasi', WRITE_OPTS)).rejects.toMatchObject({
      status: 0,
    });
  });

  it('edge: base kosong → ErrorApi status 0 tanpa request', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    await expect(
      kirimJsonAdmin('/admin/login', { method: 'POST', baseUrl: '' }),
    ).rejects.toMatchObject({ status: 0, message: expect.stringContaining('VITE_API_BASE_URL') });
    expect(fetchMock).not.toHaveBeenCalled();
  });
});

describe('kirimFileAdmin', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  const file = () => new File([new Uint8Array([1, 2, 3])], 'foto.png', { type: 'image/png' });

  it('FormData + token terkirim; balikan url dipakai', async () => {
    const fetchMock = stubFetch({ url: '/uploads/123-ab.png' }, 201);
    const hasil = await kirimFileAdmin('/admin/upload', file(), { token: 'tok123', baseUrl: BASE });
    expect(hasil).toEqual({ url: '/uploads/123-ab.png' });
    expect(fetchMock).toHaveBeenCalledWith(
      'http://api.test/admin/upload',
      expect.objectContaining({
        method: 'POST',
        headers: { Authorization: 'Bearer tok123' },
      }),
    );
    const body = fetchMock.mock.calls[0][1].body as FormData;
    expect(body).toBeInstanceOf(FormData);
    expect((body.get('gambar') as File).name).toBe('foto.png');
  });

  it('server menolak → ErrorApi; respons tanpa url → ErrorApi', async () => {
    stubFetch({ error: 'Tipe file harus JPEG, PNG, atau WebP' }, 400);
    await expect(kirimFileAdmin('/admin/upload', file(), { baseUrl: BASE })).rejects.toMatchObject({
      message: 'Tipe file harus JPEG, PNG, atau WebP',
      status: 400,
    });
    stubFetch({ ok: true }, 201);
    await expect(kirimFileAdmin('/admin/upload', file(), { baseUrl: BASE })).rejects.toMatchObject({
      message: expect.stringContaining('tanpa url'),
    });
  });

  it('edge: base kosong → ErrorApi tanpa request', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    await expect(kirimFileAdmin('/admin/upload', file(), { baseUrl: '' })).rejects.toMatchObject({
      status: 0,
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
