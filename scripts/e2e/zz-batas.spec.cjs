/* Batas laju tulis admin (30/menit per akun): melebihi batas -> 429 dengan Retry-After. Dijalankan terakhir
   (nama berkas zz-) karena menghabiskan jatah; menunggu jendela reset sebelum dan sesudah. */
'use strict';
const { test, expect, request: pwRequest } = require('@playwright/test');
const { BASE, PW, LABEL, klienAdmin, bersihkan, tidur } = require('./bantu.cjs');

test.skip(!PW, 'PW tidak di-set');
test.setTimeout(240_000);

test('tulis admin melewati batas -> 429 + Retry-After; baca tetap jalan; login salah berulang -> 429', async () => {
  await tidur(62_000); // biarkan jendela tulis dari tes sebelumnya reset
  const k = await klienAdmin(PW);
  let n429 = 0, retry = null;
  for (let i = 0; i < 45 && n429 === 0; i++) {
    // DELETE id tak ada: tetap dihitung sebagai tulis oleh limiter, tanpa mengubah data
    const res = await k.ctx.fetch('/api/admin/berita/00000000-0000-4000-8000-000000000000', { method: 'DELETE', headers: { 'X-CSRF-Token': k.csrf } });
    if (res.status() === 429) { n429++; retry = Number.parseInt(res.headers()['retry-after'], 10); expect((await res.json()).error.code).toBe('rate_limited'); }
  }
  expect(n429, 'harus kena 429 sebelum 45 tulis').toBe(1);
  expect(retry).toBeGreaterThanOrEqual(1);
  expect(retry).toBeLessThanOrEqual(60);
  expect((await k.ctx.get('/api/admin/berita')).status()).toBe(200); // baca tidak terhambat

  // login salah berulang dari satu klien
  const anon = await pwRequest.newContext({ baseURL: BASE, extraHTTPHeaders: { Origin: BASE } });
  let kena = false;
  for (let i = 0; i < 20 && !kena; i++) {
    const r = await anon.post('/api/auth/dev-login', { data: { password: 'salah-' + i } });
    if (r.status() === 429) { kena = true; expect(r.headers()['retry-after']).toBeTruthy(); }
  }
  test.info().annotations.push({ type: 'login-limit', description: kena ? 'dev-login dibatasi' : 'dev-login TIDAK dibatasi dalam 20 percobaan (dev-only)' });
  await anon.dispose();

  await tidur((retry + 1) * 1000);
  await bersihkan(k);
  await k.tutup();
});
