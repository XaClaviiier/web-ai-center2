/* Validasi runtime atas lead audit keamanan (run-1) dan kontrol sumber: CSRF/Origin, cookie sesi, XSS tersimpan,
   upload, traversal, header. Berjalan terhadap server dev lokal dengan data uji berlabel [E2E-PW]. */
'use strict';
const { test, expect, request: pwRequest } = require('@playwright/test');
const { BASE, PW, LABEL, PNG, klienAdmin, buatBerita, bersihkan, masukUI } = require('./bantu.cjs');

test.describe.configure({ mode: 'serial' });
test.skip(!PW, 'PW (ADMIN_PASSWORD) tidak di-set');

let k;
test.beforeAll(async () => { k = await klienAdmin(PW); await bersihkan(k); });
test.afterAll(async () => { await bersihkan(k); await k.tutup(); });

test('tanpa login: API admin 401; tulis tanpa Origin 403', async () => {
  const anon = await pwRequest.newContext({ baseURL: BASE });
  expect((await anon.get('/api/admin/berita')).status()).toBe(401);
  expect((await anon.get('/api/auth/me')).status()).toBe(401);
  expect((await anon.get('/api/admin/log-aktivitas')).status()).toBe(401);
  // dev-login tanpa Origin (CSRF login) ditolak
  const tanpaOrigin = await anon.post('/api/auth/dev-login', { data: { password: PW } });
  expect(tanpaOrigin.status()).toBe(403);
  const asing = await anon.post('/api/auth/dev-login', { data: { password: PW }, headers: { Origin: 'http://evil.example' } });
  expect(asing.status()).toBe(403);
  await anon.dispose();
});

test('CSRF: tulis admin tanpa token atau Origin asing ditolak; dengan keduanya lolos', async () => {
  const body = { judul: `${LABEL} csrf`, isi: 'x', tanggalPublikasi: '2020-01-01', gambar: 'berita/x.png' };
  const tanpaToken = await k.panggil('POST', '/api/admin/berita', { data: body, tanpaCsrf: true });
  expect(tanpaToken.status()).toBe(403);
  const tokenSalah = await k.panggil('POST', '/api/admin/berita', { data: body, header: { 'X-CSRF-Token': 'salah' } });
  expect(tokenSalah.status()).toBe(403);
  const originAsing = await k.panggil('POST', '/api/admin/berita', { data: body, header: { Origin: 'http://evil.example' } });
  expect(originAsing.status()).toBe(403);
  const tanpaOrigin = await k.ctx.fetch('/api/admin/berita', { method: 'POST', data: body, headers: { 'X-CSRF-Token': k.csrf, Origin: '' } });
  expect(tanpaOrigin.status()).toBe(403);
  // logout juga wajib CSRF
  expect((await k.panggil('POST', '/api/auth/logout', { tanpaCsrf: true })).status()).toBe(403);
});

test('cookie sesi: HttpOnly, SameSite=Lax, tidak terbaca JavaScript', async ({ page, context }) => {
  await masukUI(page);
  const c = (await context.cookies()).find((x) => /session/.test(x.name));
  expect(c, 'cookie sesi ada').toBeTruthy();
  expect(c.httpOnly).toBe(true);
  expect(c.sameSite).toBe('Lax');
  expect(c.path).toBe('/');
  test.info().annotations.push({ type: 'dev', description: `nama=${c.name} secure=${c.secure} (production wajib __Host-session + Secure; tidak diuji di mode development)` });
  expect(await page.evaluate(() => document.cookie)).not.toMatch(/session/);
  // token CSRF tidak tersimpan di storage
  expect(await page.evaluate(() => [localStorage.length, sessionStorage.length])).toEqual([0, 0]);
});

test('XSS tersimpan: payload di judul/isi/nama tampil sebagai teks di publik, SSR, dan admin', async ({ page, request }) => {
  const judul = `${LABEL} <img src=x onerror=window.__xss=1>`;
  const b = await buatBerita(k, judul, '<script>window.__xss=3</script><b>tebal</b>\n\n<iframe src=javascript:window.__xss=5>');
  await k.panggil('POST', `/api/admin/berita/${b.id}/post`, { data: { version: b.version } });
  const nama = `${LABEL} <svg onload=window.__xss=4>`;
  const t = await (await k.panggil('POST', '/api/admin/testimoni', { data: { nama, teks: '<script>window.__xss=6</script>', posisi: '"><b>x', foto: null } })).json();
  await k.panggil('POST', `/api/admin/testimoni/${t.id}/post`, { data: { version: t.version } });

  // HTML mentah dari SSR: payload ter-escape, tidak ada elemen/atribut berbahaya
  const html = await (await request.get(`/berita/${b.slug}`)).text();
  expect(html).not.toContain('<img src=x onerror');
  expect(html).not.toContain('<script>window.__xss');
  expect(html).toContain('&lt;img src=x onerror=window.__xss=1&gt;');
  expect(html).not.toContain('</script><script');

  const galat = [];
  page.on('dialog', (d) => { galat.push('dialog:' + d.message()); d.dismiss(); });
  for (const url of [`/berita/${b.slug}`, '/berita?q=E2E-PW', '/', '/berita']) {
    await page.goto(url);
    await page.waitForLoadState('networkidle');
    expect(await page.evaluate(() => window.__xss), `window.__xss di ${url}`).toBeUndefined();
  }
  await page.goto(`/berita/${b.slug}`);
  await expect(page.getByRole('heading', { level: 1 })).toContainText('<img src=x onerror=window.__xss=1>');
  await expect(page.locator('.artikel__isi')).toContainText('<script>window.__xss=3</script>');
  expect(await page.locator('.artikel__isi script, .artikel__isi iframe, h1 img').count()).toBe(0);

  await masukUI(page);
  for (const hash of ['#/berita', '#/testimoni', '#/log-aktivitas']) {
    await page.goto('/admin/' + hash);
    await page.waitForLoadState('networkidle');
    expect(await page.evaluate(() => window.__xss), `window.__xss di admin ${hash}`).toBeUndefined();
  }
  await page.goto('/admin/#/log-aktivitas');
  await expect(page.locator('#hasil-log')).toContainText('<img src=x onerror=window.__xss=1>'); // label log tampil sebagai teks
  expect(galat).toEqual([]);
});

test('upload: non-gambar, terlalu besar, sumber salah, dan tanpa file ditolak; gambar sah diterima', async () => {
  const kirim = (multipart) => k.panggil('POST', '/api/admin/upload', { multipart });
  const teks = await kirim({ sumber: 'berita', gambar: { name: 'x.png', mimeType: 'image/png', buffer: Buffer.from('<html><script>alert(1)</script>') } });
  expect([400, 415]).toContain(teks.status());
  const svg = await kirim({ sumber: 'berita', gambar: { name: 'x.svg', mimeType: 'image/svg+xml', buffer: Buffer.from('<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"/>') } });
  expect([400, 415]).toContain(svg.status());
  const besar = await kirim({ sumber: 'berita', gambar: { name: 'b.png', mimeType: 'image/png', buffer: Buffer.concat([PNG, Buffer.alloc(3 * 1024 * 1024)]) } });
  expect([400, 413]).toContain(besar.status());
  expect((await kirim({ sumber: 'rahasia', gambar: { name: 'a.png', mimeType: 'image/png', buffer: PNG } })).status()).toBe(400);
  expect((await kirim({ sumber: 'berita' })).status()).toBe(400);
  const ok = await kirim({ sumber: 'berita', gambar: { name: '../../../evil.png', mimeType: 'text/plain', buffer: PNG } });
  expect(ok.status()).toBe(201);
  const { key, url } = await ok.json();
  expect(key).toMatch(/^berita\/[0-9a-f-]{36}\.png$/); // nama berkas dari klien diabaikan
  const media = await k.ctx.get(url);
  expect(media.status()).toBe(200);
  expect(media.headers()['content-type']).toBe('image/png');
  expect(media.headers()['x-content-type-options']).toBe('nosniff');
  expect(media.headers()['content-security-policy']).toContain("default-src 'none'");
});

test('traversal dan akses langsung: /assets, /media, berkas konfigurasi tidak terbaca', async () => {
  const anon = await pwRequest.newContext({ baseURL: BASE });
  const jalur = ['/assets/', '/assets/css', '/assets/%2e%2e/%2e%2e/backend/.env', '/assets/..%2f..%2fbackend%2f.env', '/assets/css/..%5c..%5c..%5cbackend%5c.env',
    '/media/../backend/.env', '/media/berita/../../x', '/media/berita/00000000-0000-4000-8000-000000000000.png', '/media/%2e%2e/%2e%2e/etc/passwd',
    '/backend/.env', '/.env', '/web/admin/index.html', '/index.html', '/berita.html', '/admin/index.html', '/api/admin/../auth/me'];
  for (const p of jalur) {
    const res = await anon.get(p, { maxRedirects: 0 });
    const t = await res.text();
    expect([404, 401, 400], `GET ${p} -> ${res.status()}`).toContain(res.status());
    expect(t, `GET ${p}`).not.toMatch(/DATABASE_URL|ADMIN_PASSWORD|RUSTFS_SECRET|root:x:/);
  }
  await anon.dispose();
});

test('kontrol kunci: heartbeat dan lepas tidak bisa dipakai tanpa memegang kunci', async () => {
  const b = await buatBerita(k, `${LABEL} kunci`);
  expect((await k.panggil('PUT', `/api/admin/berita/${b.id}/lock`)).status()).toBe(423); // belum memegang kunci
  const pasang = await k.panggil('POST', `/api/admin/berita/${b.id}/lock`);
  expect(pasang.status()).toBe(200);
  expect((await pasang.json()).heartbeatDetik).toBeGreaterThan(0);
  expect((await k.panggil('PUT', `/api/admin/berita/${b.id}/lock`)).status()).toBe(200);
  expect((await k.panggil('DELETE', `/api/admin/berita/${b.id}/lock`)).status()).toBe(204);
});

test('header keamanan: nosniff, X-Frame-Options, CSP same-origin, Referrer-Policy pada HTML, API, admin; /media tetap CSP ketat', async () => {
  const anon = await pwRequest.newContext({ baseURL: BASE });
  for (const p of ['/', '/admin', '/admin/masuk', '/berita', '/api/health', '/api/berita', '/halaman-tidak-ada', '/sitemap.xml']) {
    const h = (await anon.get(p)).headers();
    expect(h['x-content-type-options'], `nosniff ${p}`).toBe('nosniff');
    expect(h['x-frame-options'], `XFO ${p}`).toBe('DENY');
    expect(h['referrer-policy'], `referrer ${p}`).toBe('strict-origin-when-cross-origin');
    expect(h['content-security-policy'], `CSP ${p}`).toContain("frame-ancestors 'none'");
    expect(h['content-security-policy'], `CSP ${p}`).toContain("script-src 'self'");
    expect(h['content-security-policy'], `CSP ${p}`).not.toMatch(/script-src[^;]*unsafe/);
    expect(h['strict-transport-security'], `HSTS tidak aktif di development ${p}`).toBeUndefined();
  }
  expect((await anon.get('/admin')).headers()['cache-control']).toBe('no-store'); // RP-05
  await anon.dispose();
});

test('CSP tidak merusak halaman: tanpa pelanggaran CSP di beranda, berita, admin, dan login', async ({ page }) => {
  const pelanggaran = [];
  page.on('console', (m) => { if (/Content Security Policy|Refused to/i.test(m.text())) pelanggaran.push(m.text()); });
  page.on('pageerror', (e) => pelanggaran.push(e.message));
  for (const u of ['/', '/berita', '/dokumentasi', '/tim', '/layanan/inference-solutions/perencanaan-implan-gigi', '/admin/masuk']) {
    await page.goto(u);
    await page.waitForLoadState('networkidle');
  }
  await masukUI(page);
  for (const h of ['#/berita', '#/dokumentasi', '#/testimoni', '#/tim', '#/log-aktivitas']) {
    await page.goto('/admin/' + h);
    await page.waitForLoadState('networkidle');
  }
  expect(pelanggaran).toEqual([]);
});

test('URL gambar: nilai //host dan http(s) tidak dipakai sebagai src (hardening frontend)', async ({ page }) => {
  await page.goto('/');
  const hasil = await page.evaluate(() => [gambar('//evil.example/x.png'), gambar('https://evil.example/x.png'), gambar('/media/berita/a.png'), gambar('logo.png')]);
  expect(hasil).toEqual(['/assets/img///evil.example/x.png', '/assets/img/https://evil.example/x.png', '/media/berita/a.png', '/assets/img/logo.png']);
});

test('karakter tersembunyi (bidi, lebar nol) ditolak 400 pada judul dan isi; ZWJ emoji tetap diterima', async () => {
  const sisip = (kode) => `${LABEL} a${String.fromCharCode(kode)}b`;
  // validasi gagal menghapus gambar yang baru diunggah, jadi setiap percobaan mengunggah gambarnya sendiri
  const kirim = async (judul, isi = 'Isi biasa.') => {
    const up = await k.panggil('POST', '/api/admin/upload', { multipart: { sumber: 'berita', gambar: { name: 'a.png', mimeType: 'image/png', buffer: PNG } } });
    const { key } = await up.json();
    return k.panggil('POST', '/api/admin/berita', { data: { judul, isi, tanggalPublikasi: '2026-01-01', gambar: key } });
  };
  for (const kode of [0x202e, 0x200b, 0x2066, 0xfeff]) {
    for (const res of [await kirim(sisip(kode)), await kirim(`${LABEL} bersih`, `isi${String.fromCharCode(kode)}x`)]) {
      expect(res.status(), `U+${kode.toString(16)}`).toBe(400);
      expect((await res.json()).error.message, `U+${kode.toString(16)}`).toMatch(/karakter yang tidak diizinkan/);
    }
  }
  const ok = await kirim(`${LABEL} emoji ${String.fromCodePoint(0x1f468, 0x200d, 0x1f469)}`);
  expect(ok.status()).toBe(201);
});
