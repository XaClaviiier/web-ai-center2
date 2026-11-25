// Alur terpenting: admin menulis dan menerbitkan berita lewat UI, pengunjung melihatnya, lalu admin menariknya ke trash.
// Memakai langkah agen Claude (lihat e2e.config.ts). Satu tujuan atomik per agent.act. Data uji berlabel [E2E-AGENT].
import { test } from '@e2e-dev/web';
import { expect, unique } from 'e2e';

test('direktur menerbitkan berita dan pengunjung dapat membacanya', { session: 'direktur', timeout: 600_000 }, async ({ app, agent, browser }) => {
  const judul = `[E2E-AGENT] Berita uji ${Date.now()}`;
  await app.open('/admin/#/berita/baru');
  await agent.act('isi form berita: judul {judul}, satu paragraf singkat tentang kegiatan AI Center sebagai konten, dan unggah sebuah gambar', {
    params: { judul: unique(judul) },
  });
  await agent.act('simpan berita ini');
  await agent.act('terbitkan (Post) berita berjudul {judul}', { params: { judul: unique(judul) } });
  await agent.assert('berita berjudul {judul} tampil di daftar berita admin dengan status terbit', { params: { judul: unique(judul) } });

  await app.open('/berita');
  await expect(browser.locator('main')).toContainText(judul);

  await app.open('/admin/');
  await agent.act('tarik kembali (Retract) berita berjudul {judul}', { params: { judul: unique(judul) } });
  await agent.act('pindahkan berita berjudul {judul} ke trash', { params: { judul: unique(judul) } });
  await app.open('/berita');
  await expect(browser.locator('main')).not.toContainText(judul);
});

test('pengunjung di layar ponsel dapat membuka halaman Tim lewat menu', async ({ app, agent, browser }) => {
  await browser.setViewport({ width: 390, height: 844 });
  await app.open('/');
  await agent.act('buka halaman Tim lewat menu navigasi situs');
  await expect(browser).toHaveURL('/tim');
  // Isi tim bergantung data; yang diuji di sini tata letak ponsel.
  await agent.assert('halaman Tim tampil utuh di layar ponsel: judul terlihat, dan tidak ada teks atau tombol yang terpotong keluar layar');
});
