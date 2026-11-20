/* Alur 2-4 dan 9 Langkah 6 lewat UI nyata: login Keycloak, siklus berita (tambah, ubah, Post, Retract, hapus, pulihkan,
   hapus permanen), dokumentasi multi-gambar, tim, testimoni, Log Aktivitas. Satu akun (KC_USER). */
'use strict';
const { test, expect, request: pwRequest } = require('@playwright/test');
const { BASE, KC_USER, KC_PASS, LABEL, PNG, klienAdmin, bersihkan, masukUI, tidur } = require('./bantu.cjs');

test.describe.configure({ mode: 'serial' });
test.skip(!KC_USER || !KC_PASS, 'KC_USER/KC_PASS (akun Keycloak ber-role) tidak di-set');

const JUDUL = `${LABEL} Berita UI`;
const JUDUL2 = `${LABEL} Berita UI diubah`;
const berkas = (nama = 'a.png') => ({ name: nama, mimeType: 'image/png', buffer: PNG });
let k;

/* Aksi UI tidak mengulang saat 429 (batas tulis 30/menit per akun): beri jeda satu jendela agar jatah segar. */
test.beforeAll(async () => { test.setTimeout(120_000); await tidur(Number(process.env.E2E_JEDA_MS ?? 61_000)); k = await klienAdmin(); await bersihkan(k); });
test.afterAll(async () => { await bersihkan(k); await k.tutup(); });

async function cariSlug(judul, sumber = 'berita') {
  const res = await k.panggil('GET', `/api/admin/${sumber}?status=semua`);
  return (await res.json()).items.find((b) => b.judul === judul);
}

test('login: tanpa sesi /admin/ dialihkan ke authorize Keycloak (PKCE S256); kredensial benar membuka admin', async ({ page }) => {
  const anon = await pwRequest.newContext({ baseURL: BASE });
  const res = await anon.get('/api/auth/login', { maxRedirects: 0 });
  expect(res.status()).toBe(302);
  const tujuan = new URL(res.headers().location);
  expect(tujuan.pathname).toMatch(/\/protocol\/openid-connect\/auth$/);
  const q = tujuan.searchParams;
  expect([q.get('response_type'), q.get('scope'), q.get('code_challenge_method')]).toEqual(['code', 'openid', 'S256']);
  expect(q.get('redirect_uri')).toBe(`${BASE}/api/auth/callback`);
  for (const p of ['state', 'nonce', 'code_challenge', 'client_id']) expect(q.get(p), p).toBeTruthy();
  expect(q.get('prompt')).toBeNull();
  const pilih = new URL((await anon.get('/api/auth/login?pilih=1', { maxRedirects: 0 })).headers().location);
  expect(pilih.searchParams.get('prompt')).toBe('login');
  expect((await anon.post('/api/auth/dev-login', { data: { password: 'apa-saja' }, headers: { Origin: BASE } })).status()).toBe(404);
  await anon.dispose();

  await page.goto('/admin/');
  await page.waitForURL(/\/protocol\/openid-connect\/auth\?/);
  await expect(page.locator('#username')).toBeVisible();
  await masukUI(page);
  await expect(page.locator('nav.samping')).toContainText('Berita');
  await expect(page.locator('nav.samping')).toContainText('Log Aktivitas');
  await expect(page.locator('.atas__nama')).toContainText(/pegawai|direktur/);
  await page.screenshot({ path: 'output/playwright/admin-daftar.png', fullPage: true });
});

test('galat callback: layar pesan tanpa pengalihan otomatis; parameter galat dibuang dari URL', async ({ browser }) => {
  const ctx = await browser.newContext();
  const page = await ctx.newPage();
  await page.goto('/admin/?galat=akses_ditolak');
  await expect(page.getByRole('heading', { name: 'Anda tidak punya akses' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Masuk dengan akun lain' })).toHaveAttribute('href', '/api/auth/login?pilih=1');
  await page.waitForTimeout(1500); // tidak boleh berpindah ke Keycloak sendiri
  expect(new URL(page.url()).pathname).toBe('/admin/');
  expect(new URL(page.url()).search).toBe('');
  for (const [kode, judul] of [['login_kedaluwarsa', 'Proses masuk kedaluwarsa'], ['login_gagal', 'Gagal masuk']]) {
    await page.goto(`/admin/?galat=${kode}`);
    await expect(page.getByRole('heading', { name: judul })).toBeVisible();
    await expect(page.locator('main.layar a')).toHaveAttribute('href', '/api/auth/login');
  }
  // callback tanpa cookie pengikat login: kembali ke admin dengan galat, bukan sesi
  const cb = await ctx.request.get('/api/auth/callback?code=palsu&state=palsu', { maxRedirects: 0 });
  expect(cb.status()).toBe(302);
  expect(cb.headers().location).toMatch(/\/admin\/\?galat=login_(kedaluwarsa|gagal)$/);
  expect((await ctx.request.get('/api/auth/me')).status()).toBe(401);
  await ctx.close();
});

test('akun tanpa role CMS ditolak (akses_ditolak)', async ({ browser }) => {
  const user = process.env.KC_USER_TANPA_ROLE, pass = process.env.KC_PASS_TANPA_ROLE;
  test.skip(!user || !pass, 'KC_USER_TANPA_ROLE/KC_PASS_TANPA_ROLE tidak di-set');
  const ctx = await browser.newContext();
  const page = await ctx.newPage();
  await page.goto('/admin/');
  await page.locator('#username').fill(user);
  await page.locator('#password').fill(pass);
  await page.locator('#kc-login').click();
  await expect(page.getByRole('heading', { name: 'Anda tidak punya akses' })).toBeVisible();
  expect((await ctx.request.get('/api/auth/me')).status()).toBe(401);
  await ctx.close();
});

test('kontrak API: tanggal DD-MM-YYYY, ISO ditolak, edit non-draft 400 bukan_draft, version basi 409 conflict', async () => {
  const { buatBerita, hariIni } = require('./bantu.cjs');
  const b = await buatBerita(k, `${LABEL} kontrak`);
  expect(b.tanggalPublikasi).toBe(hariIni());
  const badan = (tgl, v) => ({ judul: `${LABEL} kontrak`, isi: 'x', tanggalPublikasi: tgl, gambar: b.gambar.key, version: v });
  await k.panggil('POST', `/api/admin/berita/${b.id}/lock`);
  expect((await k.panggil('PUT', `/api/admin/berita/${b.id}`, { data: badan('2026-01-01', b.version) })).status()).toBe(400); // ISO tidak lagi diterima
  expect((await k.panggil('PUT', `/api/admin/berita/${b.id}`, { data: badan(hariIni(), 99) })).status()).toBe(409);
  await k.panggil('POST', `/api/admin/berita/${b.id}/post`, { data: { version: b.version } });
  const edit = await k.panggil('PUT', `/api/admin/berita/${b.id}`, { data: badan(hariIni(), b.version + 1) });
  expect(edit.status()).toBe(400);
  expect((await edit.json()).error.code).toBe('bukan_draft');
  const pub = await (await k.ctx.get(`/api/berita/${(await (await k.panggil('GET', `/api/admin/berita/${b.id}`)).json()).slug}`)).json();
  expect(pub.tanggalPublikasi).toBe(hariIni());
});

test('siklus berita: tambah (unggah gambar), ubah, Post, publik, Retract, trash, pulihkan, hapus permanen', async ({ page, request }) => {
  await masukUI(page);

  // tambah
  await page.goto('/admin/#/berita/baru');
  await page.locator('#f-gambar').setInputFiles(berkas());
  await page.locator('#f-judul').fill(JUDUL);
  await page.locator('#f-isi').fill('Paragraf satu.\n\nParagraf dua dari tes e2e.');
  await page.getByRole('button', { name: 'Simpan' }).click();
  await expect(page.getByText('Tersimpan sebagai draft.')).toBeVisible();
  const baris = () => page.locator('tr', { hasText: JUDUL });
  await expect(baris()).toBeVisible();
  await expect(baris().getByText('Draft', { exact: false })).toBeVisible();

  // ubah (judul)
  await baris().getByRole('link', { name: /^Edit:/ }).click();
  await page.locator('#f-judul').fill(JUDUL2);
  await page.getByRole('button', { name: 'Simpan' }).click();
  await expect(page.getByText('Perubahan tersimpan.')).toBeVisible();
  const b = await cariSlug(JUDUL2);
  expect(b, 'berita berjudul baru ada di API admin').toBeTruthy();
  expect(b.version).toBe(2);

  // Post -> tampil publik (SSR + JSON + HTML ber-SEO)
  const row = page.locator('tr', { hasText: JUDUL2 });
  await row.getByRole('button', { name: /^Post:/ }).click();
  await expect(page.getByText('Konten tampil di publik.')).toBeVisible();
  const html = await (await request.get(`/berita/${b.slug}`)).text();
  expect(html).toContain(`<title>${JUDUL2} |`);
  expect(html).toContain('application/ld+json');
  expect(html).toContain('id="data-awal"');
  await page.goto(`/berita/${b.slug}`);
  await expect(page.locator('article.artikel h1')).toContainText(JUDUL2);
  await expect(page.locator('[data-seo-konten]')).toHaveCount(0); // teks SEO polos diganti tampilan React
  await expect(page.locator('.artikel__gambar')).toHaveAttribute('src', /^\/media\/berita\//);
  await expect(page.locator('.artikel__isi')).toContainText('Paragraf dua dari tes e2e.');
  await page.screenshot({ path: 'output/playwright/berita-detail.png', fullPage: true });
  const daftarPublik = await (await request.get('/api/berita?q=E2E-PW')).json();
  expect(daftarPublik.items.some((x) => x.slug === b.slug)).toBe(true);

  // Retract -> langsung 404 di publik
  await page.goto('/admin/#/berita');
  await page.locator('tr', { hasText: JUDUL2 }).getByRole('button', { name: /^Retract:/ }).click();
  await page.getByRole('alertdialog').getByRole('button', { name: 'Retract', exact: true }).click();
  await expect(page.getByText('Dikembalikan ke draft.')).toBeVisible();
  expect((await request.get(`/berita/${b.slug}`)).status()).toBe(404);
  expect((await request.get(`/api/berita/${b.slug}`)).status()).toBe(404);

  // hapus ke trash -> pulihkan -> hapus lagi -> hapus permanen
  await page.locator('tr', { hasText: JUDUL2 }).getByRole('button', { name: /^Hapus:/ }).click();
  await page.getByRole('alertdialog').getByRole('button', { name: 'Pindahkan ke trash' }).click();
  await expect(page.getByText('Dipindahkan ke trash.')).toBeVisible();
  await page.goto('/admin/#/berita/trash');
  await page.locator('tr', { hasText: JUDUL2 }).getByRole('button', { name: /^Restore:/ }).click();
  await expect(page.getByText('Dipulihkan sebagai draft.')).toBeVisible();
  await page.goto('/admin/#/berita');
  await page.locator('tr', { hasText: JUDUL2 }).getByRole('button', { name: /^Hapus:/ }).click();
  await page.getByRole('alertdialog').getByRole('button', { name: 'Pindahkan ke trash' }).click();
  await page.goto('/admin/#/berita/trash');
  await page.locator('tr', { hasText: JUDUL2 }).getByRole('button', { name: /^Hapus Permanen:/ }).click();
  await page.getByRole('alertdialog').getByRole('button', { name: 'Hapus Permanen' }).click();
  await expect(page.getByText('Dihapus permanen.')).toBeVisible();
  expect(await cariSlug(JUDUL2)).toBeFalsy();
});

test('dokumentasi dengan dua gambar: tambah, Post, galeri publik', async ({ page, request }) => {
  const judul = `${LABEL} Dokumentasi UI`;
  await masukUI(page);
  await page.goto('/admin/#/dokumentasi/baru');
  await page.locator('#f-gambarDaftar').setInputFiles([berkas('1.png'), berkas('2.png')]);
  await expect(page.locator('ol.galeri-admin li')).toHaveCount(2);
  await page.locator('#f-judul').fill(judul);
  await page.getByRole('button', { name: 'Simpan' }).click();
  await expect(page.getByText('Tersimpan sebagai draft.')).toBeVisible();
  await page.locator('tr', { hasText: judul }).getByRole('button', { name: /^Post:/ }).click();
  await expect(page.getByText('Konten tampil di publik.')).toBeVisible();
  const d = await cariSlug(judul, 'dokumentasi');
  await page.goto(`/dokumentasi/${d.slug}`);
  await expect(page.locator('.galeri img')).toHaveCount(2);
  // kartu album menampilkan jumlah foto (desain prototipe)
  await page.goto('/dokumentasi');
  await expect(page.locator('article', { hasText: judul })).toContainText('2 Foto');
  const html = await (await request.get(`/dokumentasi/${d.slug}`)).text();
  expect(html).toContain('ImageGallery');
});

test('tim dan testimoni lewat form: tambah, tampil publik', async ({ page, request }) => {
  const nama = `${LABEL} Anggota`;
  await masukUI(page);
  await page.goto('/admin/#/tim/baru');
  await page.locator('#f-foto').setInputFiles(berkas('t.png'));
  await page.locator('#f-posisi').fill('Peneliti');
  await page.locator('#f-nama').fill(nama);
  await page.locator('#f-bidangRiset').fill('Computer Vision');
  await page.locator('#f-urutan').fill('99');
  await page.getByRole('button', { name: 'Simpan' }).click();
  await expect(page.getByText('Anggota tim ditambahkan.')).toBeVisible();
  await page.goto('/tim');
  await expect(page.locator('.kartu-tim', { hasText: nama })).toBeVisible();

  const ntes = `${LABEL} Pemberi`;
  await page.goto('/admin/#/testimoni/baru');
  await page.locator('#f-nama').fill(ntes);
  await page.locator('#f-teks').fill('Pelatihannya sangat membantu.');
  await page.locator('#f-posisi').fill('Dosen');
  await page.getByRole('button', { name: 'Simpan' }).click();
  await expect(page.getByText('Tersimpan sebagai draft.')).toBeVisible();
  await page.locator('tr', { hasText: ntes }).getByRole('button', { name: /^Post:/ }).click();
  await expect(page.getByText('Konten tampil di publik.')).toBeVisible();
  const pub = await (await request.get('/api/testimoni')).json();
  expect(pub.items.some((t) => t.nama === ntes)).toBe(true);
});

test('Log Aktivitas mencatat aksi di atas dengan username dari sesi', async ({ page }) => {
  await masukUI(page);
  await page.goto('/admin/#/log-aktivitas');
  const hasil = page.locator('#hasil-log');
  await expect(hasil).toContainText('Menambahkan testimoni');
  await expect(hasil).toContainText(LABEL);
  await expect(hasil).toContainText(KC_USER, { ignoreCase: true }); // preferred_username Keycloak (huruf kecil)
  await expect(hasil).toContainText('Mempublikasikan');
  await page.screenshot({ path: 'output/playwright/log-aktivitas.png', fullPage: true });
  // filter jenis
  await page.locator('#log-jenis').selectOption('hapus');
  await page.getByRole('button', { name: 'Terapkan' }).click();
  await expect(hasil).toContainText(/Memindahkan ke trash|Menghapus permanen/);
});

test('keluar: sesi lokal dan SSO Keycloak berakhir; /admin/ kembali meminta kredensial', async ({ page, context }) => {
  await masukUI(page);
  await page.getByRole('button', { name: 'Keluar' }).click();
  await page.getByRole('alertdialog').getByRole('button', { name: 'Keluar' }).click();
  await page.waitForURL(/\/protocol\/openid-connect\/auth\?/);
  await expect(page.locator('#username')).toBeVisible(); // SSO sudah diakhiri backend: tidak masuk otomatis
  expect((await context.request.get('/api/auth/me')).status()).toBe(401);
});
