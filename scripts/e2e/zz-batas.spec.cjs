/* Batas laju tulis admin (30/menit per akun): melebihi batas -> 429 dengan Retry-After. Dijalankan terakhir
   (nama berkas zz-) karena menghabiskan jatah; menunggu jendela reset sebelum dan sesudah. */
'use strict';
const { test, expect, request: pwRequest } = require('@playwright/test');
const { BASE, KC_USER, KC_PASS, klienAdmin, bersihkan, tidur } = require('./bantu.cjs');

test.skip(!KC_USER || !KC_PASS, 'KC_USER/KC_PASS tidak di-set');
test.setTimeout(240_000);

test('tulis admin melewati batas -> 429 + Retry-After; baca tetap jalan; mulai login berulang -> 429', async () => {
  await tidur(62_000); // biarkan jendela tulis dari tes sebelumnya reset
  const k = await klienAdmin();
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

  // mulai login (/api/auth/login) berulang dari satu IP dibatasi RATE_LIMIT_LOGIN_PER_MINUTE
  const anon = await pwRequest.newContext({ baseURL: BASE });
  let kena = false;
  for (let i = 0; i < 40 && !kena; i++) {
    const r = await anon.get('/api/auth/login', { maxRedirects: 0 });
    if (r.status() === 429) { kena = true; expect(r.headers()['retry-after']).toBeTruthy(); }
    else expect(r.status()).toBe(302);
  }
  expect(kena, 'login harus dibatasi dalam 40 percobaan').toBe(true);
  await anon.dispose();

  await tidur((retry + 1) * 1000);
  await bersihkan(k);
  await k.tutup();
});
