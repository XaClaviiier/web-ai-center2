/* Dua admin saling mengunci (soft lock) dan optimistic locking. Dilewati bila PW2 (ADMIN_PASSWORD_2) kosong. */
'use strict';
const { test, expect } = require('@playwright/test');
const { PW, PW2, LABEL, klienAdmin, buatBerita, bersihkan, masukUI } = require('./bantu.cjs');

test.describe.configure({ mode: 'serial' });
test.skip(!PW || !PW2, 'PW dan PW2 harus di-set');

let a, b;
test.beforeAll(async () => { a = await klienAdmin(PW); b = await klienAdmin(PW2); await bersihkan(a); });
test.afterAll(async () => { await bersihkan(a); await a.tutup(); await b.tutup(); });

test('kunci: akun B ditolak 423 saat A mengedit; setelah dilepas B boleh; PUT tanpa kunci 423', async () => {
  const item = await buatBerita(a, `${LABEL} kunci dua akun`);
  const url = `/api/admin/berita/${item.id}`;
  expect((await a.panggil('POST', `${url}/lock`)).status()).toBe(200);

  const bPasang = await b.panggil('POST', `${url}/lock`);
  expect(bPasang.status()).toBe(423);
  expect((await bPasang.json()).error.code).toBe('locked');
  const bUbah = await b.panggil('PUT', url, { data: { judul: `${LABEL} dibajak`, isi: 'x', tanggalPublikasi: '2020-01-01', gambar: item.gambar.key, version: item.version } });
  expect(bUbah.status()).toBe(423);
  expect((await b.panggil('DELETE', url)).status()).toBe(423);
  await b.panggil('DELETE', `${url}/lock`); // B mencoba melepas kunci A (boleh no-op), kunci A harus tetap berlaku
  expect((await b.panggil('POST', `${url}/lock`)).status()).toBe(423);

  // daftar menandai pemegang kunci untuk B, bukan untuk A
  const daftarB = (await (await b.panggil('GET', '/api/admin/berita?status=semua')).json()).items.find((x) => x.id === item.id);
  expect(daftarB.dikunciOleh).toBeTruthy();
  const daftarA = (await (await a.panggil('GET', '/api/admin/berita?status=semua')).json()).items.find((x) => x.id === item.id);
  expect(daftarA.dikunciOleh ?? null).toBeNull();

  expect((await a.panggil('DELETE', `${url}/lock`)).status()).toBe(204);
  expect((await b.panggil('POST', `${url}/lock`)).status()).toBe(200);
  expect((await b.panggil('DELETE', `${url}/lock`)).status()).toBe(204);
});

test('optimistic locking: version basi -> 409 version_conflict', async () => {
  const item = await buatBerita(a, `${LABEL} versi basi`);
  const url = `/api/admin/berita/${item.id}`;
  const badan = (v) => ({ judul: `${LABEL} versi basi`, isi: 'Isi baru.', tanggalPublikasi: item.tanggalPublikasi, gambar: item.gambar.key, version: v });
  await a.panggil('POST', `${url}/lock`);
  const ok = await a.panggil('PUT', url, { data: badan(item.version) });
  expect(ok.status()).toBe(200);
  await a.panggil('POST', `${url}/lock`);
  const basi = await a.panggil('PUT', url, { data: badan(item.version) });
  expect(basi.status()).toBe(409);
  expect((await basi.json()).error.code).toBe('version_conflict');
});

test('UI: B melihat banner baca-saja saat A memegang kunci', async ({ browser }) => {
  const item = await buatBerita(a, `${LABEL} banner kunci`);
  const url = `/api/admin/berita/${item.id}`;
  expect((await a.panggil('POST', `${url}/lock`)).status()).toBe(200);
  const ctx = await browser.newContext();
  const page = await ctx.newPage();
  await masukUI(page, PW2);
  await page.goto('/admin/#/berita');
  await expect(page.locator('tr', { hasText: item.judul })).toContainText('Sedang diedit oleh');
  await page.goto(`/admin/#/berita/${item.id}/edit`);
  await expect(page.locator('.banner__teks')).toContainText('Sedang diedit oleh');
  await expect(page.getByRole('button', { name: 'Simpan' })).toHaveCount(0);
  await page.screenshot({ path: 'output/playwright/kunci-banner.png', fullPage: true });
  await ctx.close();
  await a.panggil('DELETE', `${url}/lock`);
});
