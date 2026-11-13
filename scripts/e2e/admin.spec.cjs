/* Alur 2-4 dan 9 Langkah 6 lewat UI nyata: login, siklus berita (tambah, ubah, Post, Retract, hapus, pulihkan,
   hapus permanen), dokumentasi multi-gambar, tim, testimoni, Log Aktivitas. Satu akun (PW). */
'use strict';
const { test, expect } = require('@playwright/test');
const { PW, LABEL, PNG, klienAdmin, bersihkan, masukUI, tidur } = require('./bantu.cjs');

test.describe.configure({ mode: 'serial' });
test.skip(!PW, 'PW (ADMIN_PASSWORD) tidak di-set');

const JUDUL = `${LABEL} Berita UI`;
const JUDUL2 = `${LABEL} Berita UI diubah`;
const berkas = (nama = 'a.png') => ({ name: nama, mimeType: 'image/png', buffer: PNG });
let k;

/* Aksi UI tidak mengulang saat 429 (batas tulis 30/menit per akun): beri jeda satu jendela agar jatah segar. */
test.beforeAll(async () => { test.setTimeout(120_000); await tidur(Number(process.env.E2E_JEDA_MS ?? 61_000)); k = await klienAdmin(PW); await bersihkan(k); });
test.afterAll(async () => { await bersihkan(k); await k.tutup(); });

async function cariSlug(judul, sumber = 'berita') {
  const res = await k.panggil('GET', `/api/admin/${sumber}?status=semua`);
  return (await res.json()).items.find((b) => b.judul === judul);
}

test('login: password salah ditolak, password benar membuka admin', async ({ page }) => {
  await page.goto('/admin/masuk');
  await page.locator('#password').fill('salah-sekali');
  await page.getByRole('button', { name: 'Masuk' }).click();
  await expect(page.locator('#galat-masuk')).toContainText('Password salah');
  await expect(page).toHaveURL(/\/admin\/masuk/);
  await masukUI(page);
  await expect(page.locator('nav.samping')).toContainText('Berita');
  await expect(page.locator('nav.samping')).toContainText('Log Aktivitas');
  await page.screenshot({ path: 'output/playwright/admin-daftar.png', fullPage: true });
});

test('admin tanpa login menampilkan layar Masuk', async ({ browser }) => {
  const ctx = await browser.newContext();
  const page = await ctx.newPage();
  await page.goto('/admin/');
  await expect(page.getByRole('heading', { name: /Masuk ke admin/ })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Masuk' })).toHaveAttribute('href', '/admin/masuk');
  await ctx.close();
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
  await expect(page.getByRole('heading', { level: 1 })).toContainText(JUDUL2);
  await expect(page.locator('.artikel__isi')).toContainText('Paragraf dua dari tes e2e.');
  await page.screenshot({ path: 'output/playwright/berita-detail.png', fullPage: true });
  const daftarPublik = await (await request.get('/api/berita?q=E2E-PW')).json();
  expect(daftarPublik.items.some((x) => x.slug === b.slug)).toBe(true);

  // Retract -> langsung 404 di publik
  await page.goto('/admin/#/berita');
  await page.locator('tr', { hasText: JUDUL2 }).getByRole('button', { name: /^Retract:/ }).click();
  await page.locator('dialog').getByRole('button', { name: 'Retract', exact: true }).click();
  await expect(page.getByText('Dikembalikan ke draft.')).toBeVisible();
  expect((await request.get(`/berita/${b.slug}`)).status()).toBe(404);
  expect((await request.get(`/api/berita/${b.slug}`)).status()).toBe(404);

  // hapus ke trash -> pulihkan -> hapus lagi -> hapus permanen
  await page.locator('tr', { hasText: JUDUL2 }).getByRole('button', { name: /^Hapus:/ }).click();
  await page.locator('dialog').getByRole('button', { name: 'Pindahkan ke trash' }).click();
  await expect(page.getByText('Dipindahkan ke trash.')).toBeVisible();
  await page.goto('/admin/#/berita/trash');
  await page.locator('tr', { hasText: JUDUL2 }).getByRole('button', { name: /^Restore:/ }).click();
  await expect(page.getByText('Dipulihkan sebagai draft.')).toBeVisible();
  await page.goto('/admin/#/berita');
  await page.locator('tr', { hasText: JUDUL2 }).getByRole('button', { name: /^Hapus:/ }).click();
  await page.locator('dialog').getByRole('button', { name: 'Pindahkan ke trash' }).click();
  await page.goto('/admin/#/berita/trash');
  await page.locator('tr', { hasText: JUDUL2 }).getByRole('button', { name: /^Hapus Permanen:/ }).click();
  await page.locator('dialog').getByRole('button', { name: 'Hapus Permanen' }).click();
  await expect(page.getByText('Dihapus permanen.')).toBeVisible();
  expect(await cariSlug(JUDUL2)).toBeFalsy();
});

test('dokumentasi dengan dua gambar: tambah, Post, galeri publik', async ({ page, request }) => {
  const judul = `${LABEL} Dokumentasi UI`;
  await masukUI(page);
  await page.goto('/admin/#/dokumentasi/baru');
  await page.locator('#f-gambar_daftar').setInputFiles([berkas('1.png'), berkas('2.png')]);
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
  await page.locator('#f-bidang_riset').fill('Computer Vision');
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
  await expect(hasil).toContainText('admin.dev');
  await expect(hasil).toContainText('Mempublikasikan');
  await page.screenshot({ path: 'output/playwright/log-aktivitas.png', fullPage: true });
  // filter jenis
  await page.locator('#log-jenis').selectOption('hapus');
  await page.getByRole('button', { name: 'Terapkan' }).click();
  await expect(hasil).toContainText(/Memindahkan ke trash|Menghapus permanen/);
});
