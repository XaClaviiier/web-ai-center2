/* Uji MANUAL api.js terhadap server Go nyata (bukan bagian suite otomatis).
   Prasyarat: Docker (PostgreSQL, Redis, RustFS) menyala dan server berjalan di http://localhost:8080 dengan
   APP_ENV=development dan JOBS_ENABLED=false, mis.  cd backend; $env:JOBS_ENABLED='false'; go run ./cmd/api
   Jalankan dari akar repo (password dibaca dari backend/.env, tidak dicetak):
     PowerShell: $env:PW=(sls '^ADMIN_PASSWORD=(.*)$' backend/.env).Matches[0].Groups[1].Value; node scripts/tests/e2e-server.cjs
   Membuat data uji bertanda [E2E] di database dan bucket DEV, lalu menghapus permanen semuanya di akhir (juga bila
   gagal di tengah). Hanya key yang dibuat uji ini yang dihapus. Uji dua akun hanya jalan bila PW2 (ADMIN_PASSWORD_2) diisi. */
'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const root = path.resolve(__dirname, '..', '..');
const { JSDOM } = require(path.join(root, 'node_modules', 'jsdom'));
const BASE = process.env.BASE_URL ?? 'http://localhost:8080'; // BASE_URL untuk server uji di port lain
const PNG = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==', 'base64');

async function masuk(pw) {
  const r = await fetch(BASE + '/api/auth/dev-login', { method: 'POST', headers: { 'Content-Type': 'application/json', Origin: BASE }, body: JSON.stringify({ password: pw }) });
  assert.equal(r.status, 200, 'dev-login');
  return r.headers.getSetCookie().map((c) => c.split(';')[0]).join('; ');
}

function buatKlien(cookie) {
  const dom = new JSDOM('<body></body>', { url: BASE + '/', runScripts: 'outside-only' });
  const w = dom.window;
  // Batas tulis admin 30/menit per akun: uji ini menulis lebih banyak, jadi 429 ditunggu (Retry-After) lalu diulang.
  w.fetch = async (u, o = {}) => {
    for (let i = 0; ; i++) {
      const res = await fetch(BASE + u, { ...o, headers: { ...(o.headers ?? {}), Cookie: cookie, Origin: BASE } });
      if (res.status !== 429 || i >= 3) return res;
      await new Promise((r) => setTimeout(r, (Number.parseInt(res.headers.get('retry-after'), 10) || 5) * 1000 + 500));
    }
  };
  w.FormData = FormData; w.File = File;
  let n = 0;
  w.URL.createObjectURL = () => `blob:e2e/${++n}`;
  w.eval(fs.readFileSync(path.join(root, 'web/assets/js/api.js'), 'utf8') + '\nwindow.__api = api;');
  return w.__api;
}
const berkas = (nama) => new File([PNG], nama, { type: 'image/png' });
const hasil = [];
const ok = (m) => { hasil.push(m); console.log('ok  ' + m); };

(async () => {
  const pw = process.env.PW; const pw2 = process.env.PW2;
  const api = buatKlien(await masuk(pw));
  const dibuat = []; // { sumber, id } untuk pembersihan
  try {
    const s = await api.muatSesi();
    assert.equal(s.role, 'pegawai'); ok(`muatSesi: ${s.nama} (${s.role})`);

    // --- Berita: siklus penuh
    const b = await api.adminTambah('berita', { judul: '[E2E] Berita uji api.js', isi: 'Paragraf satu.\n\nParagraf dua.', tanggal_publikasi: '10-11-2026', gambar: api.blob(berkas('b.png')) });
    dibuat.push({ sumber: 'berita', id: b.id });
    assert.equal(b.status, 'draft'); assert.equal(b.tanggal_publikasi, '10-11-2026'); assert.match(b.gambar, /^berita\/.+\.png$/); assert.equal(b.version, 1);
    ok(`berita dibuat (id ${b.id}, slug ${b.slug}, key ${b.gambar})`);

    const daftar = await api.adminDaftar('berita', {});
    assert(daftar.items.some((x) => x.id === b.id)); ok('adminDaftar berita (status=semua, tanpa page/limit = semua item)');
    const d = await api.adminDetail('berita', b.id);
    assert.equal(d.isi, 'Paragraf satu.\n\nParagraf dua.'); ok('adminDetail berita');

    const k = await api.kunciAmbil('berita', b.id);
    assert.equal(k.milik, 'saya'); assert(k.heartbeatDetik > 0); ok(`kunciAmbil (heartbeat ${k.heartbeatDetik}s)`);
    await api.kunciHeartbeat(); ok('kunciHeartbeat (PUT /lock)');

    const u = await api.adminUbah('berita', b.id, { judul: '[E2E] Berita uji diubah', isi: 'Isi baru.', tanggal_publikasi: '10-11-2026', gambar: b.gambar }, d.version);
    assert.equal(u.version, 2); assert.equal(u.judul, '[E2E] Berita uji diubah'); ok('adminUbah berita (version 1 -> 2, gambar lama tanpa unggah ulang)');

    const basi = await api.adminUbah('berita', b.id, { judul: 'x', isi: 'y', tanggal_publikasi: '10-11-2026', gambar: b.gambar }, 1).catch((g) => g);
    assert.equal(basi.status, 409); assert.equal(basi.error.code, 'version_conflict'); ok('version basi -> 409 version_conflict');

    const panjang = await api.adminUbah('berita', b.id, { judul: 'J'.repeat(71), isi: 'y', tanggal_publikasi: '10-11-2026', gambar: b.gambar }, 2).catch((g) => g);
    assert.equal(panjang.status, 400); ok(`judul 71 karakter -> 400 (${panjang.error.message})`);
    const depan = await api.adminUbah('berita', b.id, { judul: 'x', isi: 'y', tanggal_publikasi: '01-01-2999', gambar: b.gambar }, 2).catch((g) => g);
    assert.equal(depan.status, 400); ok(`tanggal masa depan -> 400 (${depan.error.message})`);

    const p = await api.adminPost('berita', b.id, 2);
    assert.equal(p.status, 'posted'); assert.equal(p.version, 3); ok('adminPost (version 2 -> 3)');
    const pub = await api.detailBerita(b.slug);
    assert.equal(pub.judul, '[E2E] Berita uji diubah'); assert.equal(pub.tanggal_publikasi, '10-11-2026'); assert(pub.gambar.startsWith('/media/berita/'));
    ok('detailBerita publik memetakan data (gambar = /media/...)');
    const lp = await api.daftarBerita({ q: 'E2E', limit: 9 });
    assert(lp.items.some((x) => x.id === b.id)); ok('daftarBerita publik (q=E2E)');
    const html = await (await fetch(`${BASE}/berita/${b.slug}`)).text();
    assert(html.includes('<title>[E2E] Berita uji diubah |') && html.includes('id="data-awal"') && html.includes('id="konten-awal"')); ok('/berita/{slug}: HTML ber-SEO + data-awal');
    const img = await fetch(BASE + pub.gambar); assert.equal(img.status, 200); assert.match(img.headers.get('content-type'), /image\/png/); ok('gambar tersaji di /media');

    const gagalHapus = await api.adminHapus('berita', b.id).catch((g) => g);
    assert.equal(gagalHapus.status, 409); assert.equal(gagalHapus.error.code, 'status_tidak_valid'); ok('hapus berita posted -> 409 status_tidak_valid');
    const r = await api.adminRetract('berita', b.id, 3);
    assert.equal(r.status, 'draft'); ok('adminRetract');
    assert.equal((await fetch(`${BASE}/berita/${b.slug}`)).status, 404); ok('berita di-retract -> /berita/{slug} 404');
    await api.adminHapus('berita', b.id); ok('adminHapus (ke trash)');
    const tr = await api.adminDaftar('berita', { trash: true });
    assert(tr.items.find((x) => x.id === b.id)?.deleted_at); ok('adminDaftar trash memuat dihapusPada -> deleted_at');
    const pl = await api.adminPulihkan('berita', b.id); assert.equal(pl.status, 'draft'); ok('adminPulihkan (tanpa body)');
    await api.adminHapus('berita', b.id);
    await api.adminHapusPermanen('berita', b.id); dibuat.pop(); ok('adminHapusPermanen');
    assert.equal((await fetch(BASE + pub.gambar)).status, 404); ok('file gambar ikut terhapus');

    // --- Dokumentasi: dua gambar, urutan dibalik
    const dk = await api.adminTambah('dokumentasi', { judul: '[E2E] Dokumentasi', tanggal_publikasi: '10-11-2026', gambar_daftar: [{ key: api.blob(berkas('1.png')) }, { key: api.blob(berkas('2.png')) }] });
    dibuat.push({ sumber: 'dokumentasi', id: dk.id });
    assert.equal(dk.gambar.length, 2); ok(`dokumentasi dibuat (id ${dk.id}, 2 gambar)`);
    const dl = await api.adminDaftar('dokumentasi', {});
    const baris = dl.items.find((x) => x.id === dk.id);
    assert.equal(baris.gambar.length, 2); assert(baris.gambar[0].key); ok('daftar dokumentasi admin: gambar[0] = utama, panjang = jumlah');
    const dd = await api.adminDetail('dokumentasi', dk.id);
    const ud = await api.adminUbah('dokumentasi', dk.id, { judul: '[E2E] Dokumentasi 2', tanggal_publikasi: '10-11-2026', gambar_daftar: dd.gambar.map((g) => ({ key: g.key })).reverse() }, dd.version);
    assert.equal(ud.gambar[0].key, dd.gambar[1].key); assert.equal(ud.gambar[1].key, dd.gambar[0].key); ok('ubah dokumentasi: urutan dibalik');
    await api.adminPost('dokumentasi', dk.id, ud.version);
    const dp = await api.detailDokumentasi(ud.slug);
    assert.equal(dp.gambar.length, 2); assert(dp.gambar[0].key.startsWith('/media/dokumentasi/')); ok('detailDokumentasi publik (gambar url)');
    const dlp = await api.daftarDokumentasi({});
    const kartu = dlp.items.find((x) => x.id === dk.id);
    assert(kartu.gambar[0].key.startsWith('/media/')); assert.equal(kartu.gambar.length, 2); ok('daftarDokumentasi publik (gambarUtama + jumlah: kartu "2 Foto")');
    await api.adminRetract('dokumentasi', dk.id, (await api.adminDetail('dokumentasi', dk.id)).version);
    await api.adminHapus('dokumentasi', dk.id); await api.adminHapusPermanen('dokumentasi', dk.id); dibuat.pop(); ok('dokumentasi dibersihkan');

    // --- Testimoni tanpa foto
    const t = await api.adminTambah('testimoni', { nama: '[E2E] Penguji', teks: 'Bagus', posisi: 'QA', foto: null });
    dibuat.push({ sumber: 'testimoni', id: t.id });
    assert.equal(t.foto, null); ok(`testimoni tanpa foto (id ${t.id})`);
    const tl = await api.adminDaftar('testimoni', {}); assert(tl.items.some((x) => x.id === t.id)); ok('adminDaftar testimoni (tanpa paginasi)');
    await api.adminHapus('testimoni', t.id); await api.adminHapusPermanen('testimoni', t.id); dibuat.pop(); ok('testimoni dibersihkan');

    // --- Tim: urutan dari klien, hapus langsung permanen
    const tm = await api.adminTambah('tim', { nama: '[E2E] Anggota', posisi: 'Peneliti', bidang_riset: 'CV', foto: api.blob(berkas('t.png')), urutan: 99 });
    dibuat.push({ sumber: 'tim', id: tm.id });
    assert.equal(tm.urutan, 99); ok(`tim dibuat (id ${tm.id}, urutan ${tm.urutan} dari klien)`);
    const tmDetail = await api.adminDetail('tim', tm.id);
    await api.adminUbah('tim', tm.id, { nama: '[E2E] Anggota 2', posisi: 'Peneliti', bidang_riset: 'NLP', foto: tmDetail.foto, urutan: 5 }, tmDetail.version);
    assert.equal((await api.adminDetail('tim', tm.id)).urutan, 5); ok('ubah tim (urutan 99 -> 5)');
    const pt = await api.daftarTim(); assert(pt.items.some((x) => x.id === tm.id && x.bidang_riset === 'NLP')); ok('daftarTim publik (bidang_riset)');
    await api.adminHapus('tim', tm.id); dibuat.pop(); ok('hapus tim (permanen)');

    // --- Kunci oleh akun lain (hanya bila ADMIN_PASSWORD_2 diisi)
    if (pw2) {
      const lain = buatKlien(await masuk(pw2)); await lain.muatSesi();
      const x = await api.adminTambah('testimoni', { nama: '[E2E] Kunci', teks: 'T', posisi: 'P', foto: null });
      dibuat.push({ sumber: 'testimoni', id: x.id });
      assert.equal((await api.kunciAmbil('testimoni', x.id)).milik, 'saya');
      const milik = await lain.kunciAmbil('testimoni', x.id);
      assert.equal(milik.milik, 'lain'); assert(milik.nama);
      const barisKunci = (await lain.adminDaftar('testimoni', {})).items.find((i) => i.id === x.id);
      assert(barisKunci.dikunci_oleh, 'daftar admin akun lain menandai item dikunci'); ok(`daftar admin: dikunci_oleh = ${barisKunci.dikunci_oleh}`);
      const tolak = await lain.adminUbah('testimoni', x.id, { nama: 'x', teks: 'x', posisi: 'x', foto: null }, x.version).catch((g) => g);
      assert.equal(tolak.status, 423); assert.equal(tolak.error.code, 'locked'); ok(`akun lain: kunciAmbil -> milik lain (${milik.nama}); PUT -> 423 locked`);
      api.kunciLepas(); await new Promise((r) => setTimeout(r, 300));
      await api.adminHapus('testimoni', x.id); await api.adminHapusPermanen('testimoni', x.id); dibuat.pop();
    } else console.log('--  ADMIN_PASSWORD_2 kosong: uji dua akun dilewati');

    // --- Log aktivitas: aksi di atas tercatat di server dengan username dari sesi
    // halaman log tetap 10 entri (desain); telusuri semua halaman
    let log = await api.adminLog({ username: 'admin.dev' });
    assert.equal(log.limit, 10);
    const semua = [...log.items];
    for (let p = 2; p <= Math.ceil(log.total / log.limit); p++) semua.push(...(await api.adminLog({ username: 'admin.dev', page: p })).items);
    const e2e = semua.filter((e) => String(e.label_item).startsWith('[E2E]'));
    assert(e2e.length >= 20, `log memuat aksi uji (${e2e.length})`);
    const aksi = new Set(e2e.map((e) => e.aksi));
    for (const a of ['tambah', 'ubah', 'post', 'retract', 'hapus', 'restore', 'hapus_permanen']) assert(aksi.has(a), `aksi ${a} tercatat`);
    assert(e2e.every((e) => e.username === 'admin.dev' && e.pengguna_sub === 'dev-admin' && e.terjadi_pada && e.item_id > 0));
    assert(log.usernames.includes('admin.dev')); ok(`log aktivitas: ${e2e.length} entri [E2E], aksi lengkap, username dari sesi`);

    // --- Keluar
    await api.keluar();
    const sesudah = await api.muatSesi().catch((g) => g);
    assert(sesudah === null || sesudah.status === 401); ok('keluar -> sesi tidak berlaku lagi');
  } finally {
    // Pembersihan: hanya baris uji yang tercatat di sini.
    for (const { sumber, id } of dibuat) {
      try { await api.adminHapus(sumber, id); } catch { /* mungkin sudah di trash/posted */ }
      try { await api.adminHapusPermanen(sumber, id); } catch { /* sumber tim: hapus pertama sudah permanen */ }
    }
  }
  console.log(`\n${hasil.length} pemeriksaan lulus terhadap server nyata`);
  process.exit(0);
})().catch((e) => { console.error('GAGAL:', e); process.exit(1); });
