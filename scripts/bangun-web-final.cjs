#!/usr/bin/env node
/* Membangun `web/` (frontend yang disajikan Go) dari `frontend-final/`. Dapat dijalankan ulang kapan saja:
     node scripts/bangun-web-final.cjs [--tanpa-install]
   Langkah:
     1. `bun install --frozen-lockfile` (kecuali --tanpa-install) dan `bun --bun run build` langsung di `frontend-final/`
        (astro check + tsc + astro build + vite admin). `dist/` dan `.astro/` di sana ter-ignore git.
     2. `web/` dihapus lalu diisi `dist/public`; kerangka detail dipindah ke `web/{berita,dokumentasi}/_kerangka/index.html`;
        halaman masuk development (`frontend-final/admin-masuk/`) ditambahkan ke `web/admin/`; hash CSP skrip inline ditulis.
   Jangan menyunting `web/` dengan tangan. Alat: Bun (variabel BUN_BIN untuk path khusus). */
'use strict';
const fs = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const crypto = require('node:crypto');

const akar = path.resolve(__dirname, '..');
const SUMBER = path.join(akar, 'frontend-final');
const WEB = path.join(akar, 'web');
const BUN = process.env.BUN_BIN || 'bun';

const gagal = (pesan) => { console.error('GAGAL: ' + pesan); process.exit(1); };
const lf = (s) => s.replace(/\r\n/g, '\n');

function jalankan(perintah, args) {
  const r = spawnSync(perintah, args, { cwd: SUMBER, stdio: 'inherit', shell: process.platform === 'win32' });
  if (r.status !== 0) gagal(`${perintah} ${args.join(' ')} berakhir dengan kode ${r.status}`);
}

// 1. Build di frontend-final/ (dist/ dan .astro ter-ignore).
if (!process.argv.includes('--tanpa-install')) jalankan(BUN, ['install', '--frozen-lockfile']);
jalankan(BUN, ['--bun', 'run', 'build']); // --bun: Astro 7 menolak Node < 22.23.4; Bun melaporkan versi Node yang memenuhi

// 2. Tulis web/.
const DIST = path.join(SUMBER, 'dist', 'public');
if (!fs.existsSync(path.join(DIST, 'index.html'))) gagal('hasil build tidak ditemukan: ' + DIST);
if (path.basename(WEB) !== 'web' || path.dirname(WEB) !== akar) gagal('lokasi web/ tidak sesuai harapan: ' + WEB);
fs.rmSync(WEB, { recursive: true, force: true });
fs.cpSync(DIST, WEB, { recursive: true });

for (const sumber of ['berita', 'dokumentasi']) {
  const dari = path.join(WEB, sumber, 'kerangka-seo');
  const ke = path.join(WEB, sumber, '_kerangka');
  if (!fs.existsSync(path.join(dari, 'index.html'))) gagal(`kerangka ${sumber} tidak ada di hasil build`);
  fs.renameSync(dari, ke);
  const isi = fs.readFileSync(path.join(ke, 'index.html'), 'utf8');
  for (const penanda of ['<!--seo:head-->', '<!--seo:konten-->', '<!--seo:data-->']) {
    const n = isi.split(penanda).length - 1;
    if (n !== 1) gagal(`kerangka ${sumber}: penanda ${penanda} muncul ${n} kali (harus 1)`);
  }
}

for (const nama of ['masuk.html', 'masuk.css', 'masuk.js']) fs.copyFileSync(path.join(SUMBER, 'admin-masuk', nama), path.join(WEB, 'admin', nama));
// CSP: Astro memasang skrip inline kecil (hidrasi island). Hash SHA-256-nya ditulis ke web/csp-hash-skrip.txt dan dibaca Go
// untuk script-src, sehingga 'unsafe-inline' tidak dipakai. Skrip bertipe data (JSON) bukan skrip yang dieksekusi.
const hash = new Set();
(function telusuri(dir) {
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) telusuri(p);
    else if (e.name.endsWith('.html')) {
      for (const m of fs.readFileSync(p, 'utf8').matchAll(/<script([^>]*)>([\s\S]*?)<\/script>/g)) {
        if (/src=/.test(m[1]) || /type="application\/(ld\+)?json"/.test(m[1])) continue;
        hash.add('sha256-' + crypto.createHash('sha256').update(m[2]).digest('base64'));
      }
    }
  }
})(WEB);
fs.writeFileSync(path.join(WEB, 'csp-hash-skrip.txt'), [...hash].sort().join('\n') + '\n');
console.log(`${hash.size} hash skrip inline ditulis ke web/csp-hash-skrip.txt`);
console.log('web/ selesai dibangun dari frontend-final/.');
