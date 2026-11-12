/* Tes statis dan render untuk web/ (hasil scripts/bangun-web.cjs). Jalankan: node scripts/tests/web.test.cjs
   Membangun ulang web/ lebih dulu agar yang diuji selalu hasil skrip terbaru. */
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const { JSDOM, VirtualConsole } = require('jsdom');

const akar = path.resolve(__dirname, '..', '..');
const W = path.join(akar, 'web');
execFileSync(process.execPath, [path.join(akar, 'scripts', 'bangun-web.cjs')], { stdio: 'pipe' });

function semuaBerkas(dir, hasil = []) {
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) semuaBerkas(p, hasil);
    else hasil.push(p);
  }
  return hasil;
}
const rel = (p) => path.relative(W, p).split(path.sep).join('/');
const berkas = semuaBerkas(W);
const html = berkas.filter((f) => f.endsWith('.html'));
const js = berkas.filter((f) => f.endsWith('.js'));
const teks = (f) => fs.readFileSync(f, 'utf8');

let jumlah = 0;
const tes = (nama, fn) => { fn(); jumlah++; console.log('ok  ' + nama); };

// Berkas tulisan tangan (scripts/tambahan-web) disalin skrip pembangun, jadi harus ada.
const MENUNGGU = new Set();

tes('tidak ada rujukan .html lokal (HTML dan string JS)', () => {
  for (const f of [...html, ...js]) {
    const salah = teks(f).split('\n').map((b, i) => [b, i + 1]).filter(([b]) => /(?<![\w./-])[\w-]+\.html\b/.test(b.replace(/https?:\/\/[^\s'"`)]+/g, '')));
    assert.deepEqual(salah.map(([b, i]) => `${rel(f)}:${i}: ${b.trim().slice(0, 100)}`), []);
  }
});

tes('tidak ada path relatif ../ pada href/src/url', () => {
  for (const f of [...html, ...js]) {
    const salah = teks(f).split('\n').map((b, i) => [b, i + 1]).filter(([b]) => /(href|src|url)\s*[=:(]\s*['"`(]?\.\.\//.test(b) || /['"`]\.\.\/assets/.test(b));
    assert.deepEqual(salah.map(([b, i]) => `${rel(f)}:${i}`), []);
  }
});

tes('tidak ada sisa simulasi, mock, atau panel', () => {
  const pola = /api\.masuk\(|simulasi\/|SIMULASI|['"`]simulasi:|api-simulasi|mock-data|panel\.(js|css)|\bMOCK\b|data-sim/;
  for (const f of [...html, ...js]) {
    const salah = teks(f).split('\n').map((b, i) => [b, i + 1]).filter(([b]) => pola.test(b));
    assert.deepEqual(salah.map(([b, i]) => `${rel(f)}:${i}: ${b.trim().slice(0, 100)}`), []);
  }
});

tes('kerangka SEO: tiga penanda tepat sekali, tanpa title/description statis', () => {
  for (const nama of ['berita', 'dokumentasi']) {
    const isi = teks(path.join(W, nama, '_kerangka', 'index.html'));
    for (const p of ['<!--seo:head-->', '<!--seo:konten-->', '<!--seo:data-->']) assert.equal(isi.split(p).length - 1, 1, `${nama}: ${p}`);
    assert(!/<title>/i.test(isi), `${nama}: <title> statis`);
    assert(!/name="description"/i.test(isi), `${nama}: description statis`);
    assert(isi.indexOf('<!--seo:head-->') < isi.indexOf('</head>'), `${nama}: head di dalam <head>`);
    assert(isi.indexOf('<!--seo:konten-->') > isi.indexOf('<main'), `${nama}: konten di dalam <main>`);
    assert(isi.indexOf('<!--seo:data-->') < isi.indexOf('/assets/js/api.js'), `${nama}: data sebelum skrip`);
  }
});

tes('halaman 404 noindex', () => {
  assert.match(teks(path.join(W, '404.html')), /<meta name="robots" content="noindex">/);
});

tes('semua href/src internal menunjuk berkas yang ada', () => {
  for (const f of html) {
    const dom = new JSDOM(teks(f));
    for (const el of dom.window.document.querySelectorAll('[src],link[href]')) {
      const u = el.getAttribute('src') ?? el.getAttribute('href');
      if (!u.startsWith('/') || MENUNGGU.has(u)) continue;
      assert(fs.existsSync(path.join(W, u)), `${rel(f)}: ${u} tidak ada`);
    }
    dom.window.close();
  }
});

// Rute yang akan dilayani Go (langkah 4a) ditambah pola detail.
const RUTE = new Set(['/', '/berita', '/dokumentasi', '/tim', '/tentang-kami', '/layanan/pelatihan', '/layanan/inference-solutions',
  '/layanan/inference-solutions/jejak-alumni', '/layanan/inference-solutions/perencanaan-implan-gigi', '/layanan/inference-solutions/document-intelligence',
  '/admin', '/admin/', '/admin/masuk']);
const rutaSah = (u) => {
  const p = u.split(/[?#]/)[0];
  return RUTE.has(p) || /^\/(berita|dokumentasi)\/[^/]*$/.test(p) || p.startsWith('/assets/') || p.startsWith('/media/') || p.startsWith('/api/');
};

tes('semua tautan internal (HTML dan string JS) punya rute', () => {
  for (const f of html) {
    const dom = new JSDOM(teks(f));
    for (const a of dom.window.document.querySelectorAll('a[href^="/"]')) assert(rutaSah(a.getAttribute('href')), `${rel(f)}: ${a.getAttribute('href')}`);
    dom.window.close();
  }
  for (const f of js) {
    for (const m of teks(f).matchAll(/(?:href\s*[:=]\s*|tautan\([^,]+,\s*|tautanTombol\([^,]+,[^,]+,\s*)['"`](\/[^'"`$]*)/g)) assert(rutaSah(m[1]), `${rel(f)}: ${m[1]}`);
  }
});

tes('masuk.html hanya berisi formulir password (tanpa tautan ke akun tiruan)', () => {
  const isi = teks(path.join(W, 'admin', 'masuk.html'));
  assert(isi.includes('id="form-masuk"') && isi.includes('type="password"') && isi.includes('/assets/js/masuk.js'));
  assert(!/akun=/.test(isi));
});

// === Render (JSDOM) dengan api tiruan: memastikan KONTEN, tautan bersih, dan slug dari path bekerja ===

function render(file, url, apiPalsu) {
  const galat = [];
  const vc = new VirtualConsole();
  vc.on('jsdomError', (e) => galat.push(e.message));
  const dom = new JSDOM(teks(path.join(W, file)), { url: 'http://localhost' + url, runScripts: 'outside-only', pretendToBeVisual: true, virtualConsole: vc });
  const w = dom.window;
  w.matchMedia = (q) => ({ matches: false, addEventListener() {}, removeEventListener() {} });
  w.CSS = { supports: () => false };
  w.HTMLElement.prototype.scrollIntoView = function () {};
  w.BolaNeural = { pasang() {} };
  const pecah = { items: [], page: 1, limit: 9, total: 0 };
  w.__api = apiPalsu ?? { daftarBerita: async () => pecah, daftarDokumentasi: async () => pecah, daftarTestimoni: async () => ({ items: [] }), daftarTim: async () => ({ items: [] }) };
  const skrip = [...w.document.querySelectorAll('script[src]')].map((s) => s.getAttribute('src')).filter((s) => !s.endsWith('api.js') && !s.endsWith('bola-neural.js'));
  const kode = skrip.map((s) => teks(path.join(W, s))).join('\n');
  w.eval('const api = window.__api; const BolaNeural = { pasang() {} };\n' + kode);
  return { w, d: w.document, galat };
}

const HALAMAN_URL = [
  ['index.html', '/'], ['berita.html', '/berita'], ['dokumentasi.html', '/dokumentasi'], ['tim.html', '/tim'], ['tentang-kami.html', '/tentang-kami'],
  ['pelatihan.html', '/layanan/pelatihan'], ['inference.html', '/layanan/inference-solutions'], ['404.html', '/tidak-ada'],
  ['jejak-alumni.html', '/layanan/inference-solutions/jejak-alumni'], ['perencanaan-implan-gigi.html', '/layanan/inference-solutions/perencanaan-implan-gigi'],
  ['document-intelligence.html', '/layanan/inference-solutions/document-intelligence'],
];

(async () => {
  for (const [file, url] of HALAMAN_URL) {
    const t = render(file, url);
    await new Promise((r) => setTimeout(r, 20));
    assert.equal(t.galat.length, 0, `${file}: ${t.galat.join('; ')}`);
    assert(t.d.querySelector('header.nav'), `${file}: navbar`);
    assert(t.d.querySelector('#utama').children.length > 0, `${file}: #utama kosong`);
    t.w.close();
  }
  jumlah++; console.log('ok  semua halaman publik dirender tanpa galat (api tiruan)');

  // Navigasi memakai URL bersih; submenu aktif dikenali dari path.
  {
    const t = render('berita.html', '/berita');
    await new Promise((r) => setTimeout(r, 20)); // biarkan fetch tiruan selesai sebelum jendela ditutup
    const href = [...t.d.querySelectorAll('header.nav a[href]')].map((a) => a.getAttribute('href'));
    for (const h of ['/', '/layanan/pelatihan', '/layanan/inference-solutions', '/layanan/inference-solutions#portofolio', '/berita', '/dokumentasi', '/tim', '/tentang-kami']) assert(href.includes(h), `nav tanpa ${h}`);
    assert.equal(t.d.querySelector('header.nav a[href="/berita"]').getAttribute('aria-current'), 'page');
    t.w.close();
    jumlah++; console.log('ok  navbar memakai URL bersih dan menandai halaman aktif');
  }

  // Detail: slug dari path, isi awal dari server digantikan (tidak ditumpuk).
  for (const [file, url, fn, kelas] of [['berita/_kerangka/index.html', '/berita/pelatihan-ai%20dasar', 'detailBerita'], ['dokumentasi/_kerangka/index.html', '/dokumentasi/kegiatan-satu', 'detailDokumentasi']]) {
    const dipanggil = [];
    const api = { [fn]: async (slug) => { dipanggil.push(slug); throw { status: 404 }; } };
    const asli = teks(path.join(W, file)).replace('<!--seo:konten-->', '<article id="konten-awal"><h1>dari server</h1></article>').replace('<!--seo:head-->', '').replace('<!--seo:data-->', '');
    const tmp = path.join(W, '__tes_detail.html');
    fs.writeFileSync(tmp, asli);
    const t = render('__tes_detail.html', url, api);
    await new Promise((r) => setTimeout(r, 20));
    fs.rmSync(tmp);
    assert.deepEqual(dipanggil, [decodeURIComponent(url.split('/').pop())], `${file}: slug`);
    assert(!t.d.querySelector('#konten-awal'), `${file}: isi awal server masih menumpuk`);
    assert.equal(t.galat.length, 0, t.galat.join('; '));
    t.w.close();
  }
  jumlah++; console.log('ok  halaman detail membaca slug dari path dan menggantikan isi awal server');

  console.log(`\n${jumlah} kelompok tes lulus`);
})().catch((e) => { console.error(e); process.exit(1); });
