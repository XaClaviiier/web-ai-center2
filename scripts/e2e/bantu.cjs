/* Pembantu bersama untuk spesifikasi e2e. */
'use strict';
const { request: pwRequest } = require('@playwright/test');

const BASE = process.env.BASE_URL ?? 'http://localhost:8080';
const PW = process.env.PW;
const PW2 = process.env.PW2;
const LABEL = '[E2E-PW]'; // hanya baris berlabel ini yang pernah dibuat/dihapus oleh tes

/** PNG 1x1 sah (dikenali http.DetectContentType). */
const PNG = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==', 'base64');

const tidur = (ms) => new Promise((r) => setTimeout(r, ms));

/** Hari ini (WIB) dalam format tanggal API: DD-MM-YYYY. */
const hariIni = () => new Date().toLocaleDateString('en-GB', { timeZone: 'Asia/Jakarta' }).replaceAll('/', '-');

/** Klien API ber-cookie untuk satu akun. Mengulang otomatis bila kena batas laju (429) dengan menunggu Retry-After. */
async function klienAdmin(password, { origin = BASE } = {}) {
  const ctx = await pwRequest.newContext({ baseURL: BASE, extraHTTPHeaders: { Origin: origin } });
  const login = await ctx.post('/api/auth/dev-login', { data: { password } });
  if (login.status() !== 200) throw new Error(`login gagal (${login.status()}): periksa PW/BASE_URL/ALLOWED_ORIGINS`);
  const sesi = await login.json();
  const csrf = sesi.csrfToken;

  async function panggil(metode, url, { data, multipart, tanpaCsrf = false, header = {}, coba = 6 } = {}) {
    for (let i = 0; ; i++) {
      const opsi = { headers: { ...(tanpaCsrf ? {} : { 'X-CSRF-Token': csrf }), ...header } };
      if (data !== undefined) opsi.data = data;
      if (multipart) opsi.multipart = multipart;
      const res = await ctx.fetch(url, { method: metode, ...opsi });
      if (res.status() !== 429 || i >= coba) return res;
      const tunggu = Number.parseInt(res.headers()['retry-after'], 10) || 5;
      await tidur(tunggu * 1000 + 300);
    }
  }
  return { ctx, csrf, sesi, panggil, tutup: () => ctx.dispose() };
}

/** Buat berita draft lewat API; kembalikan objek admin lengkap. */
async function buatBerita(k, judul, isi = 'Paragraf satu.\n\nParagraf dua.') {
  const up = await k.panggil('POST', '/api/admin/upload', { multipart: { sumber: 'berita', gambar: { name: 'e2e.png', mimeType: 'image/png', buffer: PNG } } });
  if (up.status() !== 201) throw new Error('upload gagal ' + up.status());
  const { key } = await up.json();
  const res = await k.panggil('POST', '/api/admin/berita', { data: { judul, isi, tanggalPublikasi: hariIni(), gambar: key } });
  if (res.status() !== 201) throw new Error('buat berita gagal ' + res.status() + ' ' + (await res.text()));
  return res.json();
}

/** Hapus permanen semua baris berlabel [E2E-PW] pada keempat sumber (draft, posted, trash). */
async function bersihkan(k) {
  let n = 0;
  for (const sumber of ['berita', 'dokumentasi', 'testimoni', 'tim']) {
    for (const status of sumber === 'tim' ? [''] : ['semua', 'trash']) {
      const res = await k.panggil('GET', `/api/admin/${sumber}${status ? `?status=${status}` : ''}`);
      if (res.status() !== 200) continue;
      for (const b of (await res.json()).items) {
        if (!String(b.judul ?? b.nama ?? '').startsWith(LABEL)) continue;
        if (b.status === 'posted') await k.panggil('POST', `/api/admin/${sumber}/${b.id}/retract`, { data: { version: b.version } });
        if (sumber === 'tim') await k.panggil('DELETE', `/api/admin/tim/${b.id}`);
        else {
          if (status !== 'trash') await k.panggil('DELETE', `/api/admin/${sumber}/${b.id}`);
          await k.panggil('DELETE', `/api/admin/${sumber}/${b.id}/permanen`);
        }
        n++;
      }
    }
  }
  return n;
}

/** Login lewat UI /admin/masuk; menunggu daftar admin tampil. */
async function masukUI(page, password = PW) {
  await page.goto('/admin/masuk');
  await page.locator('#password').fill(password);
  await page.getByRole('button', { name: 'Masuk' }).click();
  await page.waitForURL(/\/admin\/?(#.*)?$/);
  await page.locator('nav.samping').waitFor();
}

module.exports = { BASE, PW, PW2, LABEL, PNG, klienAdmin, buatBerita, bersihkan, masukUI, tidur, hariIni };
