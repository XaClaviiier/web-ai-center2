/* Tes lapisan API (web/assets/js/api.js) dan alur admin terhadap kontrak backend, dengan fetch palsu.
   Jalankan: node scripts/tests/api.test.cjs   (membangun ulang web/ lebih dulu) */
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const { deepEqual: lemah } = require('node:assert'); // objek dari jendela JSDOM berbeda realm: bandingkan struktur, bukan prototipe
const { execFileSync } = require('node:child_process');
const { JSDOM, VirtualConsole } = require('jsdom');

const akar = path.resolve(__dirname, '..', '..');
const W = path.join(akar, 'web');
execFileSync(process.execPath, [path.join(akar, 'scripts', 'bangun-web.cjs')], { stdio: 'pipe' });
const baca = (rel) => fs.readFileSync(path.join(W, rel), 'utf8');

/** Respons fetch palsu. */
const balas = (status, isi, header = {}) => ({
  ok: status >= 200 && status < 300, status,
  headers: { get: (k) => header[k] ?? header[k.toLowerCase()] ?? null },
  json: async () => { if (isi === undefined) throw new Error('kosong'); return isi; },
});

/** Jendela JSDOM dengan fetch palsu yang mencatat permintaan. `penangan(req)` mengembalikan respons. */
function siapkan({ html = '<body></body>', url = 'http://localhost:8080/', skrip = [], penangan }) {
  const galat = [];
  const vc = new VirtualConsole();
  vc.on('jsdomError', (e) => galat.push(e.message));
  const dom = new JSDOM(html, { url, runScripts: 'outside-only', pretendToBeVisual: true, virtualConsole: vc });
  const w = dom.window;
  const log = [];
  w.fetch = async (u, opts = {}) => {
    const req = { url: u, metode: opts.method ?? 'GET', headers: opts.headers ?? {}, body: opts.body, keepalive: Boolean(opts.keepalive) };
    if (typeof req.body === 'string') { try { req.json = JSON.parse(req.body); } catch { /* bukan JSON */ } }
    log.push(req);
    return penangan(req);
  };
  w.matchMedia = () => ({ matches: false, addEventListener() {}, removeEventListener() {} });
  w.CSS = { supports: () => false };
  w.HTMLElement.prototype.scrollIntoView = function () {};
  let n = 0;
  w.URL.createObjectURL = () => `blob:test/${++n}`;
  w.eval(skrip.map(baca).join('\n') + '\nwindow.__api = api;');
  return { w, d: w.document, log, galat, api: w.__api };
}
const ts = (ms = 20) => new Promise((r) => setTimeout(r, ms));
const API = ['assets/js/api.js'];

let jumlah = 0;
const tes = async (nama, fn) => { await fn(); jumlah++; console.log('ok  ' + nama); };
const gagalDengan = async (janji) => { try { await janji; } catch (g) { return g; } assert.fail('seharusnya melempar galat'); };

(async () => {
  /* === Publik === */
  await tes('daftar berita: query, pemetaan camelCase -> halaman (tanggal DD-MM-YYYY, isi = ringkasan)', async () => {
    const t = siapkan({ skrip: API, penangan: () => balas(200, { items: [{ id: 1, slug: 'a-b', judul: 'A B', ringkasan: 'Ringkas.', gambar: '/media/berita/x.png', tanggalPublikasi: '2026-11-09' }], page: 2, limit: 9, total: 12 }) });
    const d = await t.api.daftarBerita({ q: 'ai', page: 2, limit: 9 });
    assert.equal(t.log[0].url, '/api/berita?q=ai&page=2&limit=9');
    lemah(d.items[0], { id: 1, slug: 'a-b', judul: 'A B', isi: 'Ringkas.', tanggal_publikasi: '09-11-2026', gambar: '/media/berita/x.png' });
    assert.equal(d.page, 2); assert.equal(d.total, 12);
    assert.equal(t.log[0].headers['X-CSRF-Token'], undefined, 'GET tanpa CSRF');
  });

  await tes('detail berita: data-awal dipakai sekali (slug cocok), lalu fetch', async () => {
    const awal = { id: 1, slug: 'a-b', judul: 'A B', ringkasan: 'R', isi: 'Isi', gambar: '/media/x.png', tanggalPublikasi: '2026-11-09' };
    const t = siapkan({ html: `<body><script id="data-awal" type="application/json">${JSON.stringify(awal)}</script></body>`, skrip: API, penangan: () => balas(200, awal) });
    const a = await t.api.detailBerita('a-b');
    assert.equal(t.log.length, 0, 'data-awal tidak memicu fetch');
    assert.equal(a.isi, 'Isi'); assert.equal(a.tanggal_publikasi, '09-11-2026');
    await t.api.detailBerita('a-b');
    assert.equal(t.log.length, 1, 'kedua kali fetch');
    assert.equal(t.log[0].url, '/api/berita/a-b');
    const lain = siapkan({ html: `<body><script id="data-awal" type="application/json">${JSON.stringify(awal)}</script></body>`, skrip: API, penangan: () => balas(200, awal) });
    await lain.api.detailBerita('slug-lain');
    assert.equal(lain.log.length, 1, 'slug berbeda tidak memakai data-awal');
  });

  await tes('dokumentasi: daftar (gambarUtama + jumlahGambar) dan detail (url)', async () => {
    const t = siapkan({ skrip: API, penangan: (r) => (r.url.startsWith('/api/dokumentasi/')
      ? balas(200, { id: 2, slug: 'k', judul: 'K', tanggalPublikasi: '2026-01-02', gambar: [{ url: '/media/a.png' }, { url: '/media/b.png' }] })
      : balas(200, { items: [{ id: 2, slug: 'k', judul: 'K', tanggalPublikasi: '2026-01-02', gambarUtama: '/media/a.png', jumlahGambar: 3 }, { id: 3, slug: 'z', judul: 'Z', tanggalPublikasi: '2026-01-03' }], page: 1, limit: 9, total: 2 })) });
    const l = await t.api.daftarDokumentasi({});
    // kartu memakai gambar[0].key (utama) dan gambar.length (jumlah foto)
    assert.equal(l.items[0].gambar.length, 3); assert.equal(l.items[0].gambar[0].key, '/media/a.png');
    lemah(l.items[1].gambar, []);
    const d = await t.api.detailDokumentasi('k');
    assert.equal(d.gambar.length, 2); assert.equal(d.gambar[0].key, '/media/a.png'); assert.equal(d.tanggal_publikasi, '02-01-2026');
  });

  await tes('testimoni dan tim publik dipetakan', async () => {
    const t = siapkan({ skrip: API, penangan: (r) => (r.url === '/api/tim'
      ? balas(200, { items: [{ id: 1, nama: 'Ani', posisi: 'Dosen', bidangRiset: 'CV', foto: '/media/tim/a.png' }] })
      : balas(200, { items: [{ id: 5, nama: 'Budi', teks: 'Bagus', posisi: 'Mhs' }] })) });
    assert.equal((await t.api.daftarTim()).items[0].bidang_riset, 'CV');
    assert.equal((await t.api.daftarTestimoni()).items[0].foto, null);
  });

  /* === Galat === */
  await tes('pemetaan galat: 404, 429 + Retry-After, 423 locked + pemegang, jaringan, 500', async () => {
    const jawaban = {
      '/api/a': balas(404, { error: { code: 'not_found', message: 'Sumber tidak ditemukan' } }),
      '/api/b': balas(429, { error: { code: 'rate_limited', message: 'Terlalu banyak permintaan' } }, { 'Retry-After': '42' }),
      '/api/c': balas(500, undefined),
    };
    const t = siapkan({ skrip: API, penangan: (r) => jawaban[r.url] });
    const uji = async (url) => { t.w.fetch = async (u) => jawaban[url]; return gagalDengan(t.api.detailBerita('x')); };
    let g = await uji('/api/a');
    assert.equal(g.status, 404); assert.equal(g.error.code, 'not_found');
    g = await uji('/api/b');
    assert.equal(g.status, 429); assert.equal(g.retry_after, 42); assert.equal(g.error.code, 'rate_limited');
    g = await uji('/api/c');
    assert.equal(g.status, 500); assert.match(g.error.message, /kendala/i);
    t.w.fetch = async () => { throw new TypeError('gagal'); };
    g = await gagalDengan(t.api.detailBerita('x'));
    assert.equal(g.status, 0); assert.equal(g.error.code, 'jaringan');
  });

  /* === Sesi dan CSRF === */
  await tes('sesi: 401 -> null; 200 -> token CSRF dikirim pada tulis, bukan pada baca', async () => {
    let masuk = false;
    const t = siapkan({ skrip: API, penangan: (r) => {
      if (r.url === '/api/auth/me') return masuk ? balas(200, { nama: 'Admin Dev', username: 'admin.dev', role: 'pegawai', csrfToken: 'tok-123' }) : balas(401, { error: { code: 'unauthenticated', message: 'Belum login' } });
      return balas(204, undefined);
    } });
    assert.equal(await t.api.muatSesi(), null);
    masuk = true;
    lemah(await t.api.muatSesi(), { nama: 'Admin Dev', role: 'pegawai', username: 'admin.dev' });
    await t.api.adminHapus('tim', 3);
    const hapus = t.log.at(-1);
    assert.equal(hapus.metode, 'DELETE'); assert.equal(hapus.url, '/api/admin/tim/3'); assert.equal(hapus.headers['X-CSRF-Token'], 'tok-123');
    assert.equal(t.log[1].headers['X-CSRF-Token'], undefined);
    await t.api.keluar();
    assert.equal(t.log.at(-1).url, '/api/auth/logout'); assert.equal(t.log.at(-1).headers['X-CSRF-Token'], 'tok-123');
  });

  /* === Admin === */
  const sesiAdmin = (tambah) => (r) => (r.url === '/api/auth/me' ? balas(200, { nama: 'Admin', role: 'pegawai', csrfToken: 'tk' }) : tambah(r));
  const beritaAdmin = (o = {}) => ({ id: 1, slug: 's', judul: 'Judul', ringkasan: 'R', isi: 'Isi', tanggalPublikasi: '2026-11-09', gambar: { key: 'berita/k.png', url: '/media/berita/k.png' }, status: 'draft', version: 3, dibuatPada: '2026-11-09T00:00:00Z', dibuatOleh: 'u', diubahOleh: 'u', ...o });

  await tes('adminDaftar: tanpa page/limit (daftar lengkap), status di query, dikunci_oleh dipetakan', async () => {
    const t = siapkan({ skrip: API, penangan: sesiAdmin((r) => (r.url.startsWith('/api/admin/berita')
      ? balas(200, { items: [beritaAdmin({ dikunciOleh: 'Rani' }), beritaAdmin({ id: 2, dihapusPada: '2026-11-10T00:00:00Z' })], page: 1, limit: 100000, total: 2 })
      : r.url.startsWith('/api/admin/tim') ? balas(200, { items: [{ id: 1, nama: 'Ani', posisi: 'D', bidangRiset: 'CV', foto: { key: 'tim/a.png', url: '/media/tim/a.png' }, urutan: 1, version: 1 }] })
        : balas(200, { items: [{ id: 4, nama: 'B', teks: 'T', posisi: 'P', status: 'posted', version: 2 }] }))) });
    await t.api.muatSesi();
    let d = await t.api.adminDaftar('berita', {});
    assert.equal(t.log.at(-1).url, '/api/admin/berita?status=semua', 'tanpa page/limit agar server mengembalikan semua');
    assert.equal(d.items[0].gambar, 'berita/k.png'); assert.equal(d.items[0].tanggal_publikasi, '09-11-2026'); assert.equal(d.items[0].version, 3);
    assert.equal(d.items[1].deleted_at, '2026-11-10T00:00:00Z'); assert.equal(d.items[0].dikunci_oleh, 'Rani'); assert.equal(d.items[1].dikunci_oleh, null);
    await t.api.adminDaftar('berita', { trash: true });
    assert.equal(t.log.at(-1).url, '/api/admin/berita?status=trash');
    d = await t.api.adminDaftar('tim');
    assert.equal(t.log.at(-1).url, '/api/admin/tim'); assert.equal(d.items[0].foto, 'tim/a.png'); assert.equal(d.items[0].urutan, 1); assert.equal(d.total, 1);
    d = await t.api.adminDaftar('testimoni', { trash: true });
    assert.equal(t.log.at(-1).url, '/api/admin/testimoni?status=trash'); assert.equal(d.items[0].foto, null);
  });

  await tes('adminTambah berita: unggah blob dulu (multipart sumber+gambar), lalu POST camelCase tanpa version', async () => {
    const t = siapkan({ skrip: API, penangan: sesiAdmin((r) => (r.url === '/api/admin/upload' ? balas(201, { key: 'berita/baru.png', url: '/media/berita/baru.png' }) : balas(201, beritaAdmin({ id: 9 })))) });
    await t.api.muatSesi();
    const berkas = new t.w.File(['x'], 'foto.png', { type: 'image/png' });
    const blob = t.api.blob(berkas);
    const hasil = await t.api.adminTambah('berita', { judul: 'Judul Baru', isi: 'Isi', tanggal_publikasi: '09-11-2026', gambar: blob });
    const up = t.log.find((r) => r.url === '/api/admin/upload');
    assert.equal(up.metode, 'POST'); assert.equal(up.body.get('sumber'), 'berita'); assert(up.body.get('gambar') instanceof t.w.File);
    assert.equal(up.headers['X-CSRF-Token'], 'tk'); assert.equal(up.headers['Content-Type'], undefined, 'multipart: biarkan browser mengisi boundary');
    const buat = t.log.at(-1);
    assert.equal(buat.url, '/api/admin/berita'); assert.equal(buat.metode, 'POST');
    lemah(buat.json, { judul: 'Judul Baru', isi: 'Isi', tanggalPublikasi: '2026-11-09', gambar: 'berita/baru.png' });
    assert.equal(hasil.id, 9);
  });

  await tes('adminUbah: version dikirim; key lama tidak diunggah ulang; blob yang sama tidak diunggah dua kali saat coba lagi', async () => {
    let tolak = true;
    const t = siapkan({ skrip: API, penangan: sesiAdmin((r) => {
      if (r.url === '/api/admin/upload') return balas(201, { key: 'dokumentasi/n.png', url: '/media/dokumentasi/n.png' });
      if (tolak) { tolak = false; return balas(409, { error: { code: 'version_conflict', message: 'Data sudah diubah oleh pengguna lain' } }); }
      return balas(200, { id: 5, slug: 'd', judul: 'D', tanggalPublikasi: '2026-11-09', gambar: [], status: 'draft', version: 4 });
    }) });
    await t.api.muatSesi();
    const blob = t.api.blob(new t.w.File(['x'], 'b.png', { type: 'image/png' }));
    const data = { judul: 'D', tanggal_publikasi: '09-11-2026', gambar_daftar: [{ key: 'dokumentasi/lama.png' }, { key: blob }] };
    const g = await gagalDengan(t.api.adminUbah('dokumentasi', 5, data, 3));
    assert.equal(g.status, 409); assert.equal(g.error.code, 'version_conflict');
    await t.api.adminUbah('dokumentasi', 5, data, 3);
    assert.equal(t.log.filter((r) => r.url === '/api/admin/upload').length, 1, 'unggah sekali');
    const put = t.log.at(-1);
    assert.equal(put.metode, 'PUT'); assert.equal(put.url, '/api/admin/dokumentasi/5');
    lemah(put.json, { judul: 'D', tanggalPublikasi: '2026-11-09', gambar: [{ key: 'dokumentasi/lama.png' }, { key: 'dokumentasi/n.png' }], version: 3 });
  });

  await tes('testimoni tanpa foto mengirim foto:null; tim memakai bidangRiset dan mengirim urutan', async () => {
    const t = siapkan({ skrip: API, penangan: sesiAdmin((r) => balas(201, r.url.endsWith('/tim') ? { id: 2, nama: 'A', posisi: 'P', bidangRiset: 'CV', foto: { key: 'tim/a.png', url: '/media/tim/a.png' }, urutan: 3, version: 1 } : { id: 1, nama: 'A', teks: 'T', posisi: 'P', status: 'draft', version: 1 })) });
    await t.api.muatSesi();
    await t.api.adminTambah('testimoni', { nama: 'A', teks: 'T', posisi: 'P', foto: null });
    lemah(t.log.at(-1).json, { nama: 'A', teks: 'T', posisi: 'P', foto: null });
    await t.api.adminTambah('tim', { nama: 'A', posisi: 'P', bidang_riset: 'CV', foto: 'tim/a.png', urutan: 7 });
    lemah(t.log.at(-1).json, { nama: 'A', posisi: 'P', bidangRiset: 'CV', foto: 'tim/a.png', urutan: 7 });
  });

  await tes('post/retract membawa {version}; restore, hapus, hapus permanen tanpa body', async () => {
    const t = siapkan({ skrip: API, penangan: sesiAdmin((r) => (r.metode === 'DELETE' ? balas(204, undefined) : balas(200, beritaAdmin()))) });
    await t.api.muatSesi();
    await t.api.adminPost('berita', 1, 3);
    lemah([t.log.at(-1).metode, t.log.at(-1).url, t.log.at(-1).json], ['POST', '/api/admin/berita/1/post', { version: 3 }]);
    await t.api.adminRetract('testimoni', 4, 2);
    lemah([t.log.at(-1).url, t.log.at(-1).json], ['/api/admin/testimoni/4/retract', { version: 2 }]);
    await t.api.adminPulihkan('berita', 1);
    lemah([t.log.at(-1).metode, t.log.at(-1).url, t.log.at(-1).body], ['POST', '/api/admin/berita/1/restore', undefined]);
    await t.api.adminHapus('berita', 1);
    lemah([t.log.at(-1).metode, t.log.at(-1).url], ['DELETE', '/api/admin/berita/1']);
    await t.api.adminHapusPermanen('berita', 1);
    lemah([t.log.at(-1).metode, t.log.at(-1).url], ['DELETE', '/api/admin/berita/1/permanen']);
  });

  await tes('soft lock: pasang (POST), heartbeat (PUT), lepas (DELETE keepalive), 423 locked -> milik lain, lock_hilang', async () => {
    let dipegangLain = false;
    const t = siapkan({ skrip: API, penangan: sesiAdmin((r) => {
      if (r.url.endsWith('/lock') && r.metode === 'POST') return dipegangLain ? balas(423, { error: { code: 'locked', message: 'Konten sedang diedit oleh Rani', pemegang: { nama: 'Rani' } } }) : balas(200, { pemegang: { nama: 'Admin' }, milikSaya: true, heartbeatDetik: 45 });
      if (r.url.endsWith('/lock') && r.metode === 'PUT') return balas(423, { error: { code: 'lock_hilang', message: 'Kunci edit sudah berakhir' } });
      return balas(204, undefined);
    }) });
    await t.api.muatSesi();
    lemah(await gagalDengan(t.api.kunciHeartbeat()).then((g) => [g.status, g.error.code]), [423, 'lock_hilang'], 'tanpa kunci');
    lemah(await t.api.kunciAmbil('berita', 7), { milik: 'saya', heartbeatDetik: 45 });
    assert.equal(t.log.at(-1).url, '/api/admin/berita/7/lock');
    const g = await gagalDengan(t.api.kunciHeartbeat());
    assert.equal(t.log.at(-1).metode, 'PUT'); assert.equal(g.error.code, 'lock_hilang');
    t.api.kunciLepas();
    await ts();
    assert.equal(t.log.at(-1).metode, 'DELETE'); assert.equal(t.log.at(-1).url, '/api/admin/berita/7/lock'); assert.equal(t.log.at(-1).keepalive, true);
    dipegangLain = true;
    lemah(await t.api.kunciAmbil('berita', 7), { milik: 'lain', nama: 'Rani' });
    const n = t.log.length; t.api.kunciLepas(); await ts();
    assert.equal(t.log.length, n, 'tidak ada kunci milik sendiri: tidak melepas');
  });

  /* === Alur admin di halaman (JSDOM + fetch palsu) === */
  const HTML_ADMIN = baca('admin/index.html').replace(/<script src="[^"]+"><\/script>/g, '');
  const SKRIP_ADMIN = ['assets/js/format.js', 'assets/js/api.js', 'assets/js/components.js', 'assets/js/kalender.js', 'assets/js/admin.js'];

  await tes('admin: belum masuk -> layar Masuk mengarah ke /admin/masuk', async () => {
    const t = siapkan({ html: HTML_ADMIN, url: 'http://localhost:8080/admin/', skrip: SKRIP_ADMIN, penangan: () => balas(401, { error: { code: 'unauthenticated', message: 'Belum login' } }) });
    await ts();
    assert.match(t.d.body.textContent, /Masuk ke admin/);
    assert.equal(t.d.querySelector('a[href="/admin/masuk"]')?.textContent.trim(), 'Masuk');
    assert.equal(t.galat.length, 0, t.galat.join('; '));
  });

  await tes('admin: daftar berita, menu Log Aktivitas tampil, Post membawa version, halaman Log memakai data server', async () => {
    let post409 = false;
    const t = siapkan({ html: HTML_ADMIN, url: 'http://localhost:8080/admin/#/berita', skrip: SKRIP_ADMIN, penangan: sesiAdmin((r) => {
      if (r.url.startsWith('/api/admin/berita?')) return balas(200, { items: [beritaAdmin({ judul: 'Berita Satu', version: 3, dikunciOleh: 'Rani' })], page: 1, limit: 100000, total: 1 });
      if (r.url.startsWith('/api/admin/log-aktivitas')) return balas(200, { items: [{ id: 5, terjadiPada: '2026-11-12T03:04:05Z', penggunaSub: 'dev-admin', username: 'admin.dev', jenis: 'ubah', aksi: 'post', modul: 'berita', itemId: 1, labelItem: 'Berita Satu' }], page: 1, limit: 10, total: 1, usernames: ['admin.dev'] });
      if (r.url === '/api/admin/berita/1/post') return post409 ? balas(409, { error: { code: 'version_conflict', message: 'Data sudah diubah oleh pengguna lain' } }) : balas(200, beritaAdmin({ status: 'posted', version: 4 }));
      return balas(404, { error: { code: 'not_found', message: 'x' } });
    }) });
    await ts(60);
    assert.match(t.d.body.textContent, /Berita Satu/);
    assert([...t.d.querySelectorAll('nav.samping a')].some((a) => /Log Aktivitas/.test(a.textContent)), 'menu Log Aktivitas tampil seperti desain');
    assert.equal(t.d.querySelector('td img.thumb').closest('tr').querySelectorAll('.tabel__kunci').length, 1, 'baris dikunci orang lain menampilkan "Sedang diedit oleh"');
    assert.match(t.d.querySelector('.tabel__kunci').textContent, /Sedang diedit oleh Rani/);
    assert([...t.d.querySelectorAll('button')].find((b) => b.textContent.trim() === 'Post').disabled, 'Post nonaktif saat dikunci orang lain');
    assert.equal(t.d.querySelector('td img.thumb').getAttribute('src'), '/media/berita/k.png');
    // Post: tombol aktif bila tidak dikunci; paksa aktif untuk memeriksa permintaan
    const tombolPost = [...t.d.querySelectorAll('button')].find((b) => b.textContent.trim() === 'Post');
    tombolPost.disabled = false; tombolPost.click(); await ts(60);
    const post = t.log.find((r) => r.url === '/api/admin/berita/1/post');
    lemah(post.json, { version: 3 }); assert.equal(post.headers['X-CSRF-Token'], 'tk');
    // Halaman Log Aktivitas: data dari server, tata letak desain.
    t.w.location.hash = '#/log-aktivitas'; await ts(80);
    assert.match(t.d.querySelector('#hasil-log').textContent, /Mempublikasikan berita "Berita Satu" \(ID 1\)/);
    assert.match(t.d.body.textContent, /admin\.dev/);
    assert(t.log.some((r) => r.url.startsWith('/api/admin/log-aktivitas')), 'log dibaca dari server');
    assert.equal(t.galat.length, 0, t.galat.join('; '));
    post409 = true;
  });

  await tes('admin: form edit memasang kunci; judul lama >70 boleh tersimpan bila tidak diubah, diubah jadi >70 ditolak', async () => {
    const panjang = 'J'.repeat(71);
    const t = siapkan({ html: HTML_ADMIN, url: 'http://localhost:8080/admin/#/berita/1/edit', skrip: SKRIP_ADMIN, penangan: sesiAdmin((r) => {
      if (r.url === '/api/admin/berita/1') return balas(200, { ...beritaAdmin({ judul: panjang }), lock: null });
      if (r.url === '/api/admin/berita/1/lock' && r.metode === 'POST') return balas(200, { pemegang: { nama: 'Admin' }, milikSaya: true, heartbeatDetik: 60 });
      if (r.metode === 'PUT') return balas(200, beritaAdmin({ judul: panjang, version: 4 }));
      return balas(204, undefined);
    }) });
    await ts(80);
    assert.equal(t.d.querySelector('#f-judul').value, panjang);
    assert.equal(t.d.querySelector('#f-tanggal_publikasi').value, '09-11-2026');
    assert(t.log.some((r) => r.url === '/api/admin/berita/1/lock' && r.metode === 'POST'), 'kunci dipasang');
    const kirim = () => t.d.querySelector('form.form').dispatchEvent(new t.w.Event('submit', { cancelable: true, bubbles: true }));
    // judul diubah menjadi >70: ditolak di klien, tidak ada PUT
    t.d.querySelector('#f-judul').value = 'K'.repeat(80);
    kirim(); await ts(40);
    assert.match(t.d.querySelector('#f-judul-galat').textContent, /maksimal 70 karakter/);
    assert(!t.log.some((r) => r.metode === 'PUT'), 'judul diubah >70 tidak boleh dikirim');
    // judul lama dikembalikan (tidak diubah), isi diubah: lolos ke server (validasi klien tidak menolak)
    t.d.querySelector('#f-judul').value = panjang;
    t.d.querySelector('#f-isi').value = 'Isi diperbarui.';
    kirim(); await ts(60);
    const put = t.log.find((r) => r.metode === 'PUT');
    assert(put, 'judul lama yang tidak diubah harus bisa disimpan (PUT terkirim)');
    assert.equal(put.json.judul, panjang); assert.equal(put.json.isi, 'Isi diperbarui.'); assert.equal(put.json.version, 3);
    assert.equal(t.galat.length, 0, t.galat.join('; '));
  });

  await tes('admin: form edit dikunci orang lain -> hanya baca dengan nama pemegang', async () => {
    const t = siapkan({ html: HTML_ADMIN, url: 'http://localhost:8080/admin/#/berita/1/edit', skrip: SKRIP_ADMIN, penangan: sesiAdmin((r) => {
      if (r.url === '/api/admin/berita/1') return balas(200, { ...beritaAdmin(), lock: { pemegang: { nama: 'Rani' }, milikSaya: false } });
      if (r.url === '/api/admin/berita/1/lock') return balas(423, { error: { code: 'locked', message: 'Konten sedang diedit oleh Rani', pemegang: { nama: 'Rani' } } });
      return balas(204, undefined);
    }) });
    await ts(80);
    assert.match(t.d.querySelector('.banner--kunci').textContent, /Sedang diedit oleh Rani/);
    assert.equal(t.d.querySelector('#f-judul').disabled, true);
  });

  console.log(`\n${jumlah} tes lulus`);
})().catch((e) => { console.error(e); process.exit(1); });
