/* Pembantu bersama untuk spesifikasi e2e. Login memakai Keycloak sungguhan lewat browser (tanpa jalan pintas password). */
'use strict';
const { chromium, request: pwRequest } = require('@playwright/test');

const BASE = process.env.BASE_URL ?? 'http://localhost:8080';
const KC_USER = process.env.KC_USER;
const KC_PASS = process.env.KC_PASS;
const KC_USER2 = process.env.KC_USER2; // akun kedua (berbeda) untuk tes kunci dua akun
const KC_PASS2 = process.env.KC_PASS2;
const LABEL = '[E2E-PW]'; // hanya baris berlabel ini yang pernah dibuat/dihapus oleh tes

/** PNG 1x1 sah (dikenali http.DetectContentType). */
const PNG = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==', 'base64');

const tidur = (ms) => new Promise((r) => setTimeout(r, ms));

/** Hari ini (WIB) dalam format tanggal API: DD-MM-YYYY. */
const hariIni = () => new Date().toLocaleDateString('en-GB', { timeZone: 'Asia/Jakarta' }).replaceAll('/', '-');

/** Klien API ber-cookie untuk satu akun. Login Keycloak lewat browser sekali, lalu cookie sesinya dipakai konteks request
    (Origin = BASE) dengan token CSRF dari /api/auth/me. Mengulang otomatis bila kena batas laju (429) dengan menunggu Retry-After. */
async function klienAdmin(user = KC_USER, pass = KC_PASS) {
  const browser = await chromium.launch();
  let state;
  try {
    const bctx = await browser.newContext({ baseURL: BASE });
    await masukUI(await bctx.newPage(), user, pass);
    state = await bctx.storageState();
  } finally {
    await browser.close();
  }
  const ctx = await pwRequest.newContext({ baseURL: BASE, storageState: state, extraHTTPHeaders: { Origin: BASE } });
  const me = await ctx.get('/api/auth/me');
  if (me.status() !== 200) throw new Error(`sesi tidak terbentuk setelah login (${me.status()}): periksa KC_USER/BASE_URL/ALLOWED_ORIGINS`);
  const sesi = await me.json();
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

/** Jalankan `aksi` yang memicu GET /api/auth/login. Bila dibalas 429 (RATE_LIMIT_LOGIN_PER_MINUTE per IP),
    tunggu Retry-After lalu buka /admin/ lagi; paling banyak 4 kali ulang. */
async function lewatiBatasLogin(page, aksi) {
  for (let i = 0; ; i++) {
    const login = page.waitForResponse((r) => new URL(r.url()).pathname === '/api/auth/login', { timeout: 15_000 }).catch(() => null);
    await aksi();
    const res = await login;
    if (res?.status() !== 429 || i >= 4) return;
    await tidur(((Number.parseInt(res.headers()['retry-after'], 10) || 30) + 1) * 1000);
    aksi = () => page.goto('/admin/');
  }
}

/** Login lewat Keycloak: buka /admin/ (dialihkan ke /api/auth/login lalu form Keycloak), isi form, tunggu admin tampil.
    Batas login per IP (RATE_LIMIT_LOGIN_PER_MINUTE) bisa terlampaui oleh banyak login berurutan: 429 ditunggu lalu diulang.
    Bila SSO Keycloak konteks ini masih aktif, form dilewati. */
async function masukUI(page, user = KC_USER, pass = KC_PASS) {
  if (!user || !pass) throw new Error('KC_USER/KC_PASS (atau KC_USER2/KC_PASS2) tidak di-set');
  await lewatiBatasLogin(page, () => page.goto('/admin/'));
  const tombol = page.locator('#kc-login');
  await page.locator('#kc-login, nav.samping').first().waitFor();
  if (await tombol.isVisible()) {
    await page.locator('#username').fill(user);
    await page.locator('#password').fill(pass);
    await tombol.click();
  }
  await page.waitForURL(/\/admin\/?(#.*)?$/);
  await page.locator('nav.samping').waitFor();
}

module.exports = { BASE, KC_USER, KC_PASS, KC_USER2, KC_PASS2, LABEL, PNG, klienAdmin, buatBerita, bersihkan, masukUI, lewatiBatasLogin, tidur, hariIni };
