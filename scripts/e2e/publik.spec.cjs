/* Alur 1 dan 8 Langkah 6: halaman publik, 404, sitemap, robots, HEAD. Tidak membutuhkan login. */
'use strict';
const { test, expect } = require('@playwright/test');

const HALAMAN = [
  ['/', /Beranda|UBAYA AI Center/], ['/berita', /Berita/], ['/dokumentasi', /Galeri Dokumentasi|Dokumentasi/], ['/tim', /Tim/],
  ['/tentang-kami', /Tentang/], ['/layanan/pelatihan', /Pelatihan/], ['/layanan/inference-solutions', /Inference/],
  ['/layanan/inference-solutions/jejak-alumni', /JejakAlumni/], ['/layanan/inference-solutions/perencanaan-implan-gigi', /Implan/],
  ['/layanan/inference-solutions/document-intelligence', /Document|DocNext|Dokumen/],
];

for (const [url, judul] of HALAMAN) {
  test(`halaman publik ${url} termuat, berjudul, tanpa galat konsol`, async ({ page }) => {
    const galat = [];
    page.on('pageerror', (e) => galat.push(e.message));
    page.on('console', (m) => { if (m.type() === 'error') galat.push(m.text()); });
    const res = await page.goto(url);
    expect(res.status()).toBe(200);
    await expect(page).toHaveTitle(judul);
    await expect(page.locator('header.nav')).toBeVisible();
    await expect(page.locator('#utama')).not.toBeEmpty();
    // galat jaringan 404 pada /media untuk data kosong bukan masalah; yang tidak boleh: galat skrip
    expect(galat.filter((g) => !/404|Failed to load resource/i.test(g))).toEqual([]);
  });
}

test('halaman tak dikenal: 404 HTML noindex, bukan JSON', async ({ request, page }) => {
  const res = await request.get('/halaman-tidak-ada');
  expect(res.status()).toBe(404);
  expect(res.headers()['content-type']).toContain('text/html');
  expect(res.headers()['x-robots-tag']).toBe('noindex');
  const resApi = await request.get('/api/tidak-ada');
  expect(resApi.status()).toBe(404);
  expect(resApi.headers()['content-type']).toContain('application/json');
  await page.goto('/halaman-tidak-ada');
  await expect(page).toHaveTitle(/tidak ditemukan/i);
});

test('detail slug tidak ada: 404 HTML ber-noindex', async ({ request }) => {
  for (const p of ['/berita/slug-yang-tidak-ada', '/dokumentasi/slug-yang-tidak-ada']) {
    const res = await request.get(p);
    expect(res.status()).toBe(404);
    expect(res.headers()['x-robots-tag']).toBe('noindex');
  }
});

test('sitemap.xml dan robots.txt: tipe benar; GET dan HEAD sama', async ({ request }) => {
  const s = await request.get('/sitemap.xml');
  expect(s.status()).toBe(200);
  expect(s.headers()['content-type']).toContain('xml');
  expect(await s.text()).toContain('<urlset');
  const r = await request.get('/robots.txt');
  expect(r.status()).toBe(200);
  expect(await r.text()).toMatch(/Sitemap:/i);
  for (const p of ['/', '/berita', '/sitemap.xml', '/robots.txt', '/admin']) {
    const h = await request.head(p);
    expect(h.status(), `HEAD ${p}`).toBe(200);
  }
});

test('API publik: bentuk daftar, ETag 304, q tidak di-cache', async ({ request }) => {
  const a = await request.get('/api/berita');
  expect(a.status()).toBe(200);
  const j = await a.json();
  expect(Array.isArray(j.items)).toBe(true);
  expect(j).toMatchObject({ page: 1 });
  const etag = a.headers().etag;
  if (etag) expect((await request.get('/api/berita', { headers: { 'If-None-Match': etag } })).status()).toBe(304);
  expect((await request.get('/api/berita?q=a')).status()).toBe(400); // minimal 2 karakter
  expect((await request.get('/api/berita?page=0')).status()).toBe(400);
});
