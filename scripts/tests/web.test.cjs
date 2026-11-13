/* Tes statis untuk web/ (hasil scripts/bangun-web-final.cjs dari frontend-final/). Jalankan: node scripts/tests/web.test.cjs
   Menguji web/ yang ada. Untuk membangun ulang lebih dulu: REBUILD=1 node scripts/tests/web.test.cjs (butuh Bun; lama). */
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const { JSDOM } = require('jsdom');

const akar = path.resolve(__dirname, '..', '..');
const W = path.join(akar, 'web');
if (process.env.REBUILD) execFileSync(process.execPath, [path.join(akar, 'scripts', 'bangun-web-final.cjs')], { stdio: 'inherit' });

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
const kerangka = html.filter((f) => rel(f).includes('/_kerangka/'));
const halaman = html.filter((f) => !kerangka.includes(f) && rel(f) !== 'admin/masuk.html');
const teks = (f) => fs.readFileSync(f, 'utf8');

let jumlah = 0;
const tes = (nama, fn) => { fn(); jumlah++; console.log('ok  ' + nama); };

tes('keluaran build lengkap: halaman publik, admin, kerangka detail, hash CSP', () => {
  for (const p of ['index.html', '404.html', 'tim/index.html', 'berita/index.html', 'dokumentasi/index.html', 'tentang-kami/index.html',
    'layanan/pelatihan/index.html', 'layanan/inference-solutions/index.html', 'admin/index.html', 'admin/masuk.html', 'admin/masuk.js', 'admin/masuk.css',
    'berita/_kerangka/index.html', 'dokumentasi/_kerangka/index.html', 'csp-hash-skrip.txt']) assert(fs.existsSync(path.join(W, p)), `${p} tidak ada`);
  assert(berkas.some((f) => rel(f).startsWith('_astro/')), 'aset ber-hash /_astro/ tidak ada');
  assert(berkas.some((f) => rel(f).startsWith('admin/assets/')), 'aset admin tidak ada');
});

tes('web/ bukan sisa web lama (tanpa assets/js dan assets/css buatan tangan)', () => {
  assert(!fs.existsSync(path.join(W, 'assets', 'js')) && !fs.existsSync(path.join(W, 'assets', 'css')));
  assert(!fs.existsSync(path.join(W, 'berita.html')) && !fs.existsSync(path.join(W, 'tim.html')));
});

tes('kerangka SEO: tiga penanda tepat sekali, tanpa title/description statis', () => {
  assert.equal(kerangka.length, 2);
  for (const nama of ['berita', 'dokumentasi']) {
    const isi = teks(path.join(W, nama, '_kerangka', 'index.html'));
    for (const p of ['<!--seo:head-->', '<!--seo:konten-->', '<!--seo:data-->']) assert.equal(isi.split(p).length - 1, 1, `${nama}: ${p}`);
    assert(!/<title>/i.test(isi), `${nama}: <title> statis`);
    assert(!/name="description"/i.test(isi), `${nama}: description statis`);
    assert(isi.indexOf('<!--seo:head-->') < isi.indexOf('</head>'), `${nama}: head di dalam <head>`);
    assert(isi.indexOf('<!--seo:konten-->') > isi.indexOf('<main'), `${nama}: konten di dalam <main>`);
    assert(isi.indexOf('<!--seo:konten-->') < isi.indexOf('</main>'), `${nama}: konten sebelum </main>`);
  }
});

tes('halaman 404 noindex', () => {
  assert.match(teks(path.join(W, '404.html')), /<meta name="robots" content="noindex">/);
});

tes('setiap halaman publik punya navbar, #utama berisi, dan satu <h1>', () => {
  for (const f of halaman.filter((x) => !rel(x).startsWith('admin/'))) {
    const dom = new JSDOM(teks(f));
    const d = dom.window.document;
    assert(d.querySelector('header.nav'), `${rel(f)}: navbar`);
    assert(d.querySelector('#utama')?.children.length > 0, `${rel(f)}: #utama kosong`);
    assert(d.querySelector('#utama h1'), `${rel(f)}: h1`);
    assert(d.title.trim().length > 0, `${rel(f)}: title`);
    dom.window.close();
  }
});

tes('semua href/src internal menunjuk berkas yang ada (selain /api dan /media)', () => {
  for (const f of html) {
    const dom = new JSDOM(teks(f));
    for (const el of dom.window.document.querySelectorAll('[src],link[href]')) {
      const u = el.getAttribute('src') ?? el.getAttribute('href');
      if (!u.startsWith('/') || u.startsWith('/api/') || u.startsWith('/media/')) continue;
      assert(fs.existsSync(path.join(W, u.split(/[?#]/)[0])), `${rel(f)}: ${u} tidak ada`);
    }
    dom.window.close();
  }
});

// Rute yang dilayani Go: halaman statis (berkas index.html), detail berita/dokumentasi, admin, media, api.
const rutaSah = (u) => {
  const p = u.split(/[?#]/)[0];
  if (p === '/' || p === '/admin' || p === '/admin/' || p === '/admin/masuk') return true;
  if (/^\/(berita|dokumentasi)\/[^/]+$/.test(p) && p !== '/berita/_kerangka' && p !== '/dokumentasi/_kerangka') return true;
  if (p.startsWith('/media/') || p.startsWith('/api/')) return true;
  return fs.existsSync(path.join(W, p, 'index.html')) || fs.existsSync(path.join(W, p));
};

tes('semua tautan internal <a href="/..."> punya rute, tanpa rujukan .html', () => {
  for (const f of html) {
    const dom = new JSDOM(teks(f));
    for (const a of dom.window.document.querySelectorAll('a[href^="/"]')) {
      const h = a.getAttribute('href');
      assert(rutaSah(h), `${rel(f)}: ${h}`);
      assert(!/\.html([?#]|$)/.test(h), `${rel(f)}: ${h} merujuk .html`);
    }
    dom.window.close();
  }
});

tes('CSP: setiap skrip inline eksekusi punya hash di csp-hash-skrip.txt; tidak ada handler inline', () => {
  const hash = new Set(teks(path.join(W, 'csp-hash-skrip.txt')).split(/\s+/).filter(Boolean));
  assert(hash.size > 0);
  let dipakai = 0;
  for (const f of html) {
    const dom = new JSDOM(teks(f));
    for (const s of dom.window.document.querySelectorAll('script:not([src])')) {
      const tipe = (s.getAttribute('type') || '').toLowerCase();
      if (tipe === 'application/ld+json' || tipe === 'application/json') continue; // data, bukan eksekusi
      const h = 'sha256-' + crypto.createHash('sha256').update(s.textContent).digest('base64');
      assert(hash.has(h), `${rel(f)}: skrip inline tanpa hash CSP (${h})`);
      dipakai++;
    }
    assert(!dom.window.document.querySelector('[onclick],[onload],[onerror],[onsubmit]'), `${rel(f)}: handler inline`);
    dom.window.close();
  }
  assert(dipakai > 0, 'tidak ada skrip inline sama sekali: hash CSP seharusnya terpakai');
});

tes('admin: index.html memuat bundel modul; masuk.html hanya formulir password tanpa akun tiruan', () => {
  const idx = teks(path.join(W, 'admin', 'index.html'));
  assert.match(idx, /<script type="module"[^>]*src="\/admin\/assets\/[^"]+\.js"/);
  assert.match(idx, /id="root"|id="app"/);
  const masuk = teks(path.join(W, 'admin', 'masuk.html'));
  assert(masuk.includes('type="password"') && masuk.includes('id="password"') && masuk.includes('/admin/masuk.js'));
  assert(!/akun=/.test(masuk));
});

tes('bundel admin memakai klien nyata, bukan simulasi (rujuk /api/auth/me, /api/admin, CSRF)', () => {
  const js = berkas.filter((f) => rel(f).startsWith('admin/assets/') && f.endsWith('.js')).map(teks).join('\n');
  for (const s of ['/api/auth/me', '/api/admin/', 'X-CSRF-Token', '/api/admin/upload']) assert(js.includes(s), `bundel admin tanpa ${s}`);
  assert(!/[?&]akun=/.test(js), 'pemilih akun tiruan masih ada');
  assert(js.includes('/admin/masuk'), 'pengalihan ke halaman masuk tidak ada');
});

console.log(`\n${jumlah} tes lulus`);
