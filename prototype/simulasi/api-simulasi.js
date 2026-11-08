/* TIRUAN API. Dibuang saat implementasi. Halaman hanya memanggil api.*; semua data di memori
   dan kembali ke awal saat refresh. Galat berbentuk { status, error: { code, message } }. */
const SIMULASI = { keadaan: 'normal', jedaMs: 300, paksa401: false, paksa409: false, heartbeatGagal: false, kunciLain: true };

const api = (() => {
  const db = JSON.parse(JSON.stringify(MOCK));
  const BERSTATUS = ['berita', 'dokumentasi', 'testimoni'];
  const BATAS_HALAMAN = 9;
  let akunAktif = null;
  let kunciSaya = null;

  const gagal = (status, code, message, tambahan) => Object.assign({ status, error: { code, message } }, tambahan);
  // Cermin validasi server: judul berita/dokumentasi maksimal 70 karakter (salinan MAKS_JUDUL di admin.js).
  const cekJudul = (sumber, data) => {
    if ((sumber === 'berita' || sumber === 'dokumentasi') && String(data.judul ?? '').trim().length > 70) throw gagal(400, 'validation_error', 'Field "judul" maksimal 70 karakter.');
  };
  const BUKAN = () => gagal(404, 'not_found', 'Data tidak ditemukan.');

  /** Bungkus jeda dan keadaan panel simulasi. `tulis` = permintaan admin yang mengubah data. */
  function jawab(fungsi, { admin = false } = {}) {
    return new Promise((selesai, tolak) => {
      if (SIMULASI.keadaan === 'loading') return; // sengaja tidak pernah selesai
      setTimeout(() => {
        try {
          if (admin && SIMULASI.paksa401) throw gagal(401, 'unauthorized', 'Sesi Anda sudah berakhir.');
          if (SIMULASI.keadaan === 'error') throw gagal(500, 'server_error', 'Terjadi kendala di server. Coba lagi beberapa saat lagi.');
          if (SIMULASI.keadaan === '429') throw gagal(429, 'too_many_requests', 'Terlalu banyak permintaan.', { retry_after: 30 });
          selesai(JSON.parse(JSON.stringify(fungsi())));
        } catch (e) {
          tolak(e);
        }
      }, SIMULASI.jedaMs);
    });
  }

  const hidup = (baris) => !baris.deleted_at;
  const terbit = (baris) => hidup(baris) && baris.status === 'posted';
  const kosongkan = (daftar) => (SIMULASI.keadaan === 'kosong' ? [] : daftar);
  const cocok = (teks, q) => teks.toLowerCase().includes(q.toLowerCase());
  const lengkapiGambar = (dok) => ({ ...dok, gambar: db.dokumentasi_gambar.filter((g) => g.dokumentasi_id === dok.id).sort((a, b) => a.urutan - b.urutan) });

  function halaman(daftar, { page = 1, limit = BATAS_HALAMAN }) {
    return { items: daftar.slice((page - 1) * limit, page * limit), page, limit, total: daftar.length };
  }

  /** Daftar publik: pencarian judul berlaku minimal 2 karakter. */
  function daftarPublik(tabel, { q = '', page, limit }) {
    let hasil = kosongkan(db[tabel].filter(terbit)).sort(Format.bandingkanTerbaru);
    if (q.trim().length >= 2) hasil = hasil.filter((b) => cocok(b.judul, q.trim()));
    if (tabel === 'dokumentasi') hasil = hasil.map(lengkapiGambar);
    return halaman(hasil, { page, limit });
  }

  function cari(sumber, id) {
    const baris = db[sumber].find((b) => b.id === id);
    if (!baris) throw BUKAN();
    return baris;
  }
  const kunciLain = (sumber, id) => (SIMULASI.kunciLain ? db.kunci.find((k) => k.jenis === sumber && k.id === id) : null);
  function cekTerkunci(sumber, id) {
    const k = kunciLain(sumber, id);
    if (k) throw gagal(423, 'locked', `Sedang diedit oleh ${k.nama}.`);
  }
  const bersih = (teks) => String(teks ?? '').replace(/[\u0000-\u0008\u000B\u000C\u000E-\u001F\u007F]/g, '').trim();
  function slugUnik(judul, tabel) {
    // Simulasi. Aturan slug final masih [terbuka] di database.md §3.3.
    const dasar = judul.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') || 'konten';
    let slug = dasar;
    for (let n = 2; db[tabel].some((b) => b.slug === slug); n++) slug = `${dasar}-${n}`;
    return slug;
  }
  const idBaru = (daftar) => Math.max(0, ...daftar.map((b) => b.id)) + 1;
  const pengguna = () => akunAktif?.sub ?? 'sub-tidak-dikenal';

  function simpanGambarDokumentasi(id, gambar) {
    db.dokumentasi_gambar = db.dokumentasi_gambar.filter((g) => g.dokumentasi_id !== id);
    gambar.forEach((g, i) => db.dokumentasi_gambar.push({ id: idBaru(db.dokumentasi_gambar), dokumentasi_id: id, key: g.key, urutan: i + 1 }));
  }

  function urutanTim(a, b) { return a.urutan - b.urutan || a.id - b.id; }

  return {
    // === Publik ===
    daftarBerita: (opsi = {}) => jawab(() => daftarPublik('berita', opsi)),
    daftarDokumentasi: (opsi = {}) => jawab(() => daftarPublik('dokumentasi', opsi)),
    detailBerita: (slug) => jawab(() => {
      const b = db.berita.find((x) => x.slug === slug && terbit(x));
      if (!b || SIMULASI.keadaan === 'kosong') throw BUKAN();
      return b;
    }),
    detailDokumentasi: (slug) => jawab(() => {
      const d = db.dokumentasi.find((x) => x.slug === slug && terbit(x));
      if (!d || SIMULASI.keadaan === 'kosong') throw BUKAN();
      return lengkapiGambar(d);
    }),
    daftarTestimoni: () => jawab(() => ({ items: kosongkan(db.testimoni.filter(terbit)).sort((a, b) => new Date(b.created_at) - new Date(a.created_at)) })),
    daftarTim: () => jawab(() => ({ items: kosongkan([...db.tim]).sort(urutanTim) })),

    // === Sesi ===
    masuk(idAkun) { akunAktif = db.akun.find((a) => a.id === idAkun) ?? null; return akunAktif; },

    // === Admin ===
    adminDaftar: (sumber, { trash = false } = {}) => jawab(() => {
      if (sumber === 'tim') return { items: kosongkan([...db.tim]).sort(urutanTim).map((b) => ({ ...b, dikunci_oleh: kunciLain(sumber, b.id)?.nama ?? null })) };
      let hasil = db[sumber].filter((b) => (trash ? !hidup(b) : hidup(b)));
      hasil = kosongkan(hasil).sort(sumber === 'testimoni' ? (a, b) => new Date(b.created_at) - new Date(a.created_at) : Format.bandingkanTerbaru);
      if (sumber === 'dokumentasi') hasil = hasil.map(lengkapiGambar);
      return { items: hasil.map((b) => ({ ...b, dikunci_oleh: kunciLain(sumber, b.id)?.nama ?? null })) };
    }, { admin: true }),

    adminDetail: (sumber, id) => jawab(() => {
      const b = cari(sumber, id);
      return sumber === 'dokumentasi' ? lengkapiGambar(b) : b;
    }, { admin: true }),

    adminTambah: (sumber, data) => jawab(() => {
      cekJudul(sumber, data);
      const id = idBaru(db[sumber]);
      const baris = { id, version: 1, created_at: new Date().toISOString(), created_by: pengguna(), updated_by: pengguna() };
      for (const [k, v] of Object.entries(data)) if (k !== 'gambar_daftar') baris[k] = typeof v === 'string' ? bersih(v) : v;
      if (BERSTATUS.includes(sumber)) { baris.status = 'draft'; baris.deleted_at = null; }
      if (sumber === 'berita' || sumber === 'dokumentasi') baris.slug = slugUnik(baris.judul, sumber);
      db[sumber].push(baris);
      if (sumber === 'dokumentasi') simpanGambarDokumentasi(id, data.gambar_daftar);
      return baris;
    }, { admin: true }),

    adminUbah: (sumber, id, data, version) => jawab(() => {
      const b = cari(sumber, id);
      if (BERSTATUS.includes(sumber) && b.status !== 'draft') throw gagal(400, 'bukan_draft', 'Hanya data berstatus draft yang bisa diedit.');
      if (data.judul !== b.judul) cekJudul(sumber, data); // judul lama >70 tetap bisa disimpan bila tidak diubah
      if (SIMULASI.paksa409) { b.version += 1; SIMULASI.paksa409 = false; }
      if (b.version !== version) throw gagal(409, 'conflict', 'Data sudah diubah oleh pengguna lain.');
      for (const [k, v] of Object.entries(data)) if (k !== 'gambar_daftar') b[k] = typeof v === 'string' ? bersih(v) : v;
      if (sumber === 'dokumentasi') simpanGambarDokumentasi(id, data.gambar_daftar);
      b.version += 1;
      b.updated_by = pengguna();
      kunciSaya = null;
      return b;
    }, { admin: true }),

    adminPost: (sumber, id) => jawab(() => { cekTerkunci(sumber, id); const b = cari(sumber, id); b.status = 'posted'; b.version += 1; return b; }, { admin: true }),
    adminRetract: (sumber, id) => jawab(() => { const b = cari(sumber, id); b.status = 'draft'; b.version += 1; return b; }, { admin: true }),
    adminHapus: (sumber, id) => jawab(() => {
      cekTerkunci(sumber, id);
      if (sumber === 'tim') { db.tim = db.tim.filter((b) => b.id !== id); return {}; }
      cari(sumber, id).deleted_at = new Date().toISOString();
      return {};
    }, { admin: true }),
    adminPulihkan: (sumber, id) => jawab(() => { const b = cari(sumber, id); b.deleted_at = null; b.status = 'draft'; b.version += 1; return b; }, { admin: true }),
    adminHapusPermanen: (sumber, id) => jawab(() => {
      db[sumber] = db[sumber].filter((b) => b.id !== id);
      if (sumber === 'dokumentasi') db.dokumentasi_gambar = db.dokumentasi_gambar.filter((g) => g.dokumentasi_id !== id);
      return {};
    }, { admin: true }),

    // === Soft lock ===
    kunciAmbil: (sumber, id) => jawab(() => {
      const k = kunciLain(sumber, id);
      if (k) return { milik: 'lain', nama: k.nama };
      kunciSaya = { sumber, id };
      SIMULASI.heartbeatGagal = false;
      return { milik: 'saya' };
    }, { admin: true }),
    kunciHeartbeat: () => jawab(() => {
      if (SIMULASI.heartbeatGagal || !kunciSaya) { kunciSaya = null; throw gagal(410, 'lock_lost', 'Sesi edit berakhir.'); }
      return { ok: true };
    }, { admin: true }),
    kunciLepas: () => { kunciSaya = null; },
  };
})();
