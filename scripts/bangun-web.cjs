#!/usr/bin/env node
/* Membangun `web/` (sumber frontend yang disajikan Go) dari `prototype/`. Dapat dijalankan ulang kapan saja:
     node scripts/bangun-web.cjs
   `web/` ditulis ulang SELURUHNYA (dihapus lalu dibuat). Jangan menyunting `web/` dengan tangan.
   - Perubahan atas hasil salinan: dilakukan di skrip ini.
   - Berkas tulisan tangan yang harus ada di `web/` (mis. assets/js/api.js): taruh di `scripts/tambahan-web/`
     dengan struktur yang sama; isinya disalin ke `web/` setelah transformasi.
   - `prototype/` tidak pernah diubah.
   Setiap penggantian teks di JS wajib cocok tepat sebanyak yang diharapkan; bila prototipe berubah dan tidak
   cocok lagi, skrip berhenti dengan pesan jelas (bukan diam-diam menghasilkan berkas yang salah). */
'use strict';
const fs = require('node:fs');
const path = require('node:path');

const akar = path.resolve(__dirname, '..');
const P = path.join(akar, 'prototype');
const W = path.join(akar, 'web');
const TAMBAHAN = path.join(__dirname, 'tambahan-web');

// === Pemetaan ===

// Berkas prototipe (public/) -> berkas keluaran di web/. Rute Go (langkah 4) mengacu ke nama-nama ini.
const HALAMAN = {
  'index.html': 'index.html',
  'berita.html': 'berita.html',
  'dokumentasi.html': 'dokumentasi.html',
  'tim.html': 'tim.html',
  'tentang-kami.html': 'tentang-kami.html',
  'pelatihan.html': 'pelatihan.html',
  'inference.html': 'inference.html',
  '404.html': '404.html',
  'jejak-alumni.html': 'jejak-alumni.html',
  'perencanaan-implan-gigi.html': 'perencanaan-implan-gigi.html',
  'document-intelligence.html': 'document-intelligence.html',
  'berita-detail.html': 'berita/_kerangka/index.html',
  'dokumentasi-detail.html': 'dokumentasi/_kerangka/index.html',
};

// Nama halaman prototipe -> URL bersih.
const URL_BERSIH = {
  index: '/',
  berita: '/berita',
  dokumentasi: '/dokumentasi',
  tim: '/tim',
  'tentang-kami': '/tentang-kami',
  pelatihan: '/layanan/pelatihan',
  inference: '/layanan/inference-solutions',
  'jejak-alumni': '/layanan/inference-solutions/jejak-alumni',
  'perencanaan-implan-gigi': '/layanan/inference-solutions/perencanaan-implan-gigi',
  'document-intelligence': '/layanan/inference-solutions/document-intelligence',
};

// === Pembantu ===

const lf = (s) => s.replace(/\r\n/g, '\n');
const baca = (...p) => lf(fs.readFileSync(path.join(...p), 'utf8'));
function tulis(relatif, isi) {
  const tujuan = path.join(W, relatif);
  fs.mkdirSync(path.dirname(tujuan), { recursive: true });
  fs.writeFileSync(tujuan, isi);
}
function gagal(pesan) {
  console.error('GAGAL: ' + pesan);
  process.exit(1);
}

/** Ganti teks persis `dari` -> `ke`, wajib tepat `jumlah` kali. */
function ganti(isi, dari, ke, jumlah, konteks) {
  const n = isi.split(dari).length - 1;
  if (n !== jumlah) gagal(`${konteks}: "${dari.slice(0, 70)}" ditemukan ${n}x, diharapkan ${jumlah}x`);
  return isi.split(dari).join(ke);
}

/** Tautan `*.html` -> URL bersih (di HTML maupun string JS). */
function tulisUlangTautan(isi) {
  isi = isi.replace(/(?<![\w./-])berita-detail\.html\?slug=/g, '/berita/');
  isi = isi.replace(/(?<![\w./-])dokumentasi-detail\.html\?slug=/g, '/dokumentasi/');
  const nama = Object.keys(URL_BERSIH).sort((a, b) => b.length - a.length).join('|');
  return isi.replace(new RegExp(`(?<![\\w./-])(${nama})\\.html`, 'g'), (_, n) => URL_BERSIH[n]);
}

const tulisUlangAset = (isi) => isi.replaceAll('../assets/', '/assets/');

// === Halaman publik ===

function ubahHtml(isi, berkas, { detail = null } = {}) {
  isi = isi.replace(/^[ \t]*<link[^>]*simulasi\/panel\.css[^>]*>\n/gm, '');
  isi = isi.replace(/^[ \t]*<script src="\.\.\/simulasi\/panel\.js"><\/script>\n/gm, '');
  isi = isi.replace('<script src="../assets/mock-data.js"></script>', '<script src="/assets/js/konten.js"></script>');
  isi = isi.replace('<script src="../simulasi/api-simulasi.js"></script>', '<script src="/assets/js/api.js"></script>');
  isi = tulisUlangAset(isi);
  isi = tulisUlangTautan(isi);

  if (berkas === '404.html') {
    // Halaman 404 tidak boleh diindeks (RP-07); prototipe memberi "index".
    isi = ganti(isi, '<meta name="robots" content="index">', '<meta name="robots" content="noindex">', 1, berkas);
  }
  if (detail) {
    // Kerangka SEO: Go mengganti tiga penanda (backend/internal/seo). Title dan description statis dibuang.
    isi = ganti(isi, /<title>[^<]*<\/title>\n/.exec(isi)?.[0] ?? '<title>', '<!--seo:head-->\n', 1, berkas);
    const desc = /[ \t]*<meta name="description"[^>]*>\n/.exec(isi);
    if (!desc) gagal(`${berkas}: meta description tidak ditemukan`);
    isi = isi.replace(desc[0], '');
    isi = ganti(isi, '<main id="utama" tabindex="-1"></main>', '<main id="utama" tabindex="-1"><!--seo:konten--></main>', 1, berkas);
    isi = ganti(isi, '  <script src="/assets/js/konten.js"></script>', '  <!--seo:data-->\n  <script src="/assets/js/konten.js"></script>', 1, berkas);
  }
  return isi;
}

// === JS ===

function ubahPublikJs(isi) {
  const k = 'public.js';
  isi = ganti(isi, 'const K = MOCK.konten;', 'const K = KONTEN;', 1, k);
  isi = ganti(isi, 'const gambar = (nama) => `../assets/img/${nama}`;',
    'const gambar = (nama) => (/^(\\/|https?:)/.test(nama) ? nama : `/assets/img/${nama}`); // URL dari API (/media/...) dipakai apa adanya',
    1, k);
  // Slug dari path (/berita/{slug}), bukan ?slug=.
  isi = ganti(isi, "const parameter = new URLSearchParams(location.search);",
    "const parameter = new URLSearchParams(location.search);\n" +
    "const slugDariUrl = () => { try { return decodeURIComponent(location.pathname.split('/').filter(Boolean).pop() ?? ''); } catch { return ''; } };",
    1, k);
  isi = ganti(isi, "const slug = parameter.get('slug');", 'const slug = slugDariUrl();', 2, k);
  // Isi awal dari server (<!--seo:konten-->) diganti utuh oleh render JS, bukan ditumpuk.
  isi = ganti(isi, "const wadah = el('div');\n    utama.append(wadah);\n    muatDetail(", "const wadah = el('div');\n    utama.replaceChildren(wadah);\n    muatDetail(", 2, k);
  // Penanda halaman aktif di submenu: path bersih, bukan akhiran nama berkas.
  isi = ganti(isi, "location.pathname.endsWith(s.href)", "location.pathname.replace(/\\/$/, '') === s.href", 1, k);
  // Panel simulasi dibuang.
  isi = ganti(isi,
    "  // Hanya panggilan terakhir per wadah yang mendengar panel simulasi.\n  wadah._lepasPanel?.();\n  const dengar = () => jalan();\n  document.addEventListener('simulasi:ubah', dengar);\n  wadah._lepasPanel = () => document.removeEventListener('simulasi:ubah', dengar);\n",
    '', 1, k);
  isi = ganti(isi,
    "  // Panel simulasi: muat ulang detail, juga setelah layar 404 menggantikan wadah.\n  document.addEventListener('simulasi:ubah', () => { utama.replaceChildren(wadah); jalan(); });\n",
    '', 1, k);
  return isi;
}

function ubahAdminJs(isi) {
  const k = 'admin.js';
  isi = ganti(isi, "const URL_LOGIN = '../simulasi/keycloak.html';", "const URL_LOGIN = '/admin/masuk';", 1, k);
  isi = ganti(isi, ' Salinan di simulasi/api-simulasi.js.', '', 1, k);

  // Gambar: key dari backend disajikan di /media/{key}; blob: (belum diunggah) dan URL penuh dipakai apa adanya.
  isi = ganti(isi, "const urlGambar = (key) => (String(key).startsWith('blob:') ? key : `../assets/img/${key}`);",
    "const urlGambar = (key) => (/^(blob:|\\/|https?:)/.test(String(key)) ? key : `/media/${key}`);", 1, k);
  isi = ganti(isi, 'key = URL.createObjectURL(berkas);', 'key = api.blob(berkas);', 1, k);
  isi = ganti(isi, 'daftar.push(URL.createObjectURL(b));', 'daftar.push(api.blob(b));', 1, k);
  isi = ganti(isi, 'daftar[i] = URL.createObjectURL(b);', 'daftar[i] = api.blob(b);', 1, k);


  // Halaman Log Aktivitas memakai data server nyata (bukan simulasi): hanya TEKS yang menyatakan "simulasi" diluruskan,
  // tata letak dan perilakunya tetap seperti prototipe.
  isi = ganti(isi, "el('strong', null, 'Simulasi. '), 'Aktivitas simulasi hanya tersimpan selama halaman ini terbuka dan kembali kosong setelah refresh. Identitas akun dan waktu perangkat belum diverifikasi server.'",
    "el('strong', null, 'Riwayat. '), 'Aktivitas tersimpan di server bersama perubahan datanya dan tidak dapat diubah dari halaman ini. Waktu memakai jam server dan ditampilkan dalam WIB.'", 1, k);
  isi = ganti(isi, "'Riwayat aktivitas simulasi, terbaru dahulu'", "'Riwayat aktivitas, terbaru dahulu'", 1, k);
  isi = ganti(isi, "'Perubahan data yang berhasil pada simulasi ini akan tercatat di sini.'", "'Perubahan data yang berhasil akan tercatat di sini.'", 1, k);

  // Keluar: akhiri sesi di server, lalu kembali ke /admin/ (layar Masuk).
  isi = ganti(isi, "if (ya) { hentikanForm(); location.href = 'index.html'; }", "if (ya) { hentikanForm(); await api.keluar(); location.href = '/admin/'; }", 1, k);

  // Post/Retract wajib membawa version (optimistic locking, 409 version_conflict bila basi).
  isi = ganti(isi, 'api.adminPost(sumber, b.id)', 'api.adminPost(sumber, b.id, b.version)', 1, k);
  isi = ganti(isi, 'api.adminRetract(sumber, b.id)', 'api.adminRetract(sumber, b.id, b.version)', 1, k);

  // Soft lock: interval dari backend (heartbeatDetik); kunci hilang = 423 lock_hilang (atau locked bila diambil orang lain).
  isi = ganti(isi, 'timer = setInterval(detak, 60000);', 'timer = setInterval(detak, (kunci.heartbeatDetik || 60) * 1000);', 1, k);
  isi = ganti(isi, 'if (g?.status !== 410) return; // galat sementara: coba lagi pada detak berikutnya',
    "if (g?.error?.code !== 'lock_hilang' && g?.error?.code !== 'locked') return; // galat lain sementara: coba lagi pada detak berikutnya", 1, k);
  isi = ganti(isi, "    document.addEventListener('simulasi:heartbeat', detak);\n", '', 1, k);
  isi = ganti(isi, "      document.removeEventListener('simulasi:heartbeat', detak);\n", '', 1, k);

  // Boot: sesi nyata dari GET /api/auth/me (asinkron), bukan api.masuk() tiruan.
  isi = ganti(isi,
    "document.addEventListener('simulasi:ubah', render);\ndocument.addEventListener('simulasi:catat', (e) => toast(e.detail));\nakun = api.masuk(new URLSearchParams(location.search).get('akun'));\nrender();",
    '/* Boot: sesi dari server. Belum masuk -> layar Masuk; server bermasalah -> layar galat dengan Coba lagi. */\n' +
    "async function mulai() {\n  aplikasi.replaceChildren(el('main.layar', null, kerangkaBaris(3)));\n  try {\n    akun = await api.muatSesi();\n  } catch (g) {\n" +
    "    layar({ judul: 'Admin belum bisa dibuka', teks: pesanGalat(g), galat: true, aksi: tombol('Coba lagi', 'kedua', { onclick: mulai }, 'ulang') });\n    return;\n  }\n  render();\n}\nmulai();", 1, k);
  return isi;
}

// === Halaman masuk dev (dari simulasi/keycloak.html) ===

function bangunMasuk() {
  let isi = baca(P, 'simulasi', 'keycloak.html');
  isi = tulisUlangAset(isi);
  isi = ganti(isi, '<title>Simulasi Keycloak</title>', '<title>Masuk (development) | UBAYA AI Center</title>', 1, 'keycloak.html');
  const awal = isi.indexOf('<main>');
  const akhir = isi.indexOf('</main>') + '</main>'.length;
  if (awal < 0 || akhir < awal) gagal('keycloak.html: <main> tidak ditemukan');
  const main = `<main>
    <p class="label-sim">Hanya untuk development</p>
    <h1>Masuk ke admin</h1>
    <p class="teks-sekunder">Halaman ini hanya tersedia saat server berjalan dalam mode development. Di produksi, langkah ini adalah login Keycloak.</p>
    <form id="form-masuk" novalidate>
      <div class="field">
        <label for="password">Password</label>
        <input id="password" name="password" type="password" autocomplete="current-password" required>
      </div>
      <p id="galat-masuk" role="alert" hidden></p>
      <button class="tombol tombol--utama" type="submit">Masuk</button>
    </form>
    <a href="/admin/">Batal</a>
  </main>`;
  isi = isi.slice(0, awal) + main + isi.slice(akhir);
  isi = ganti(isi, '</body>', '  <script src="/assets/js/masuk.js"></script>\n</body>', 1, 'keycloak.html');
  return isi;
}

// === Utama ===

function salinPohon(dari, ke, saring = () => true) {
  fs.mkdirSync(ke, { recursive: true });
  for (const e of fs.readdirSync(dari, { withFileTypes: true })) {
    const d = path.join(dari, e.name);
    const t = path.join(ke, e.name);
    if (!saring(d)) continue;
    if (e.isDirectory()) salinPohon(d, t, saring);
    else fs.copyFileSync(d, t);
  }
}

function jalankan() {
  if (path.basename(W) !== 'web' || path.dirname(W) !== akar) gagal('lokasi web/ tidak sesuai harapan: ' + W);
  fs.rmSync(W, { recursive: true, force: true });

  // 1. Aset statis: css, font, gambar disalin apa adanya (css memakai ../fonts/ relatif, tetap benar).
  for (const sub of ['css', 'fonts', 'img']) salinPohon(path.join(P, 'assets', sub), path.join(W, 'assets', sub));

  // 2. JS. Berkas demo, format, komponen, kalender, bola-neural disalin dengan penulisan ulang tautan/aset.
  const jsDir = path.join(P, 'assets', 'js');
  for (const nama of fs.readdirSync(jsDir)) {
    let isi = baca(jsDir, nama);
    if (nama === 'public.js') isi = ubahPublikJs(isi);
    if (nama === 'admin.js') isi = ubahAdminJs(isi);
    isi = tulisUlangAset(tulisUlangTautan(isi)).replaceAll('MOCK.konten', 'KONTEN');
    tulis(path.join('assets', 'js', nama), isi);
  }

  // 3. Konten hardcode (Layanan, Pelatihan, kontak, demo) dari mock-data.js: hanya objek `konten`.
  //    Data dinamis (berita, dokumentasi, testimoni, tim, akun, kunci) tidak ikut: itu dari API.
  const mock = baca(P, 'assets', 'mock-data.js');
  const mulai = mock.indexOf('  const konten = {');
  const akhirKonten = mock.lastIndexOf('  return { berita');
  if (mulai < 0 || akhirKonten < mulai) gagal('mock-data.js: objek konten tidak ditemukan');
  let konten = mock.slice(mulai, akhirKonten).trimEnd();
  konten = konten.replace('  const konten = {', 'const KONTEN = {').replace(/;$/, ';');
  konten = konten.split('\n').map((b) => (b.startsWith('  ') ? b.slice(2) : b)).join('\n');
  konten = tulisUlangAset(tulisUlangTautan(konten));
  tulis(path.join('assets', 'js', 'konten.js'),
    '/* Konten hardcode (Layanan, Pelatihan, kontak, demo). Dibangun oleh scripts/bangun-web.cjs dari objek konten di prototipe. */\n' + konten + '\n');

  // 4. HTML publik.
  for (const [asal, keluar] of Object.entries(HALAMAN)) {
    const detail = asal.endsWith('-detail.html');
    let html = ubahHtml(baca(P, 'public', asal), asal, { detail });
    // Komentar rujukan ke skrip simulasi tidak relevan di web/.
    html = html.replace('Dibuat oleh simulasi/buat-ilustrasi-implan.cjs. ', '');
    tulis(keluar, html);
  }

  // 5. Admin dan halaman masuk dev.
  let admin = baca(P, 'admin', 'index.html');
  admin = admin.replace(/^[ \t]*<link[^>]*simulasi\/panel\.css[^>]*>\n/gm, '');
  admin = admin.replace(/^[ \t]*<script src="\.\.\/simulasi\/panel\.js"><\/script>\n/gm, '');
  admin = admin.replace(/^[ \t]*<script src="\.\.\/assets\/mock-data\.js"><\/script>\n/gm, '');
  admin = admin.replace('<script src="../simulasi/api-simulasi.js"></script>', '<script src="/assets/js/api.js"></script>');
  tulis(path.join('admin', 'index.html'), tulisUlangAset(admin));
  tulis(path.join('admin', 'masuk.html'), bangunMasuk());

  // 6. Berkas tulisan tangan (api.js dst., langkah 5).
  if (fs.existsSync(TAMBAHAN)) salinPohon(TAMBAHAN, W);

  console.log('web/ selesai dibangun.');
}

jalankan();
