/* Lapisan API nyata (satu origin, tanpa CORS). Halaman memanggil api.* dengan nama yang sama seperti prototipe;
   backend Go adalah sumber kebenaran, jadi bentuk data, kode galat, version, CSRF, dan soft lock mengikuti backend.
   Bentuk data ke halaman (snake_case, tanggal DD-MM-YYYY) dipetakan di sini; backend memakai camelCase dan YYYY-MM-DD.

   Galat selalu dilempar sebagai { status, error: { code, message }, retry_after?, pemegang? }:
   - 401 unauthenticated, 403 csrf_invalid/origin_invalid, 404 not_found, 409 version_conflict/status_tidak_valid,
     400 invalid_request/query_tidak_valid, 423 locked (pemegang) / lock_hilang, 429 rate_limited (+ Retry-After). */
const BERKAS_BARU = new Map(); // URL blob: -> File yang dipilih admin, belum diunggah
const KEY_TERUNGGAH = new Map(); // URL blob: -> key hasil unggah (agar percobaan ulang tidak mengunggah dua kali)

const api = (() => {
  let csrfToken = null;
  let kunciSaya = null; // { sumber, id } yang sedang dipegang form terbuka
  let dataAwal; // isi <script id="data-awal"> (HTML detail ber-SEO), dipakai sekali

  /* === Galat dan fetch dasar === */
  const galat = (status, code, message, tambahan) => Object.assign({ status, error: { code, message } }, tambahan);

  async function minta(metode, url, { json, form, tulis = metode !== 'GET' && metode !== 'HEAD', keepalive = false } = {}) {
    const headers = { Accept: 'application/json' };
    let body;
    if (json !== undefined) { headers['Content-Type'] = 'application/json'; body = JSON.stringify(json); }
    if (form) body = form;
    if (tulis && csrfToken) headers['X-CSRF-Token'] = csrfToken;
    let res;
    try {
      res = await fetch(url, { method: metode, headers, body, credentials: 'same-origin', keepalive });
    } catch {
      throw galat(0, 'jaringan', 'Tidak dapat menghubungi server. Periksa koneksi Anda lalu coba lagi.');
    }
    if (res.status === 204) return null;
    let isi = null;
    try { isi = await res.json(); } catch { /* badan bukan JSON */ }
    if (res.ok) return isi;
    const e = isi?.error ?? {};
    const tambahan = {};
    if (res.status === 429) tambahan.retry_after = Number.parseInt(res.headers.get('Retry-After'), 10) || 30;
    if (e.pemegang) tambahan.pemegang = e.pemegang;
    throw galat(res.status, e.code ?? 'galat_server', e.message ?? 'Terjadi kendala di server. Coba lagi beberapa saat lagi.', tambahan);
  }

  /* === Pemetaan tanggal dan gambar === */
  const keDepan = (iso) => (/^\d{4}-\d{2}-\d{2}$/.test(iso ?? '') ? `${iso.slice(8, 10)}-${iso.slice(5, 7)}-${iso.slice(0, 4)}` : iso);
  const keBelakang = (dmy) => {
    const m = /^(\d{2})-(\d{2})-(\d{4})$/.exec(String(dmy ?? '').trim());
    return m ? `${m[3]}-${m[2]}-${m[1]}` : dmy;
  };
  const kueri = (opsi) => {
    const p = new URLSearchParams();
    for (const [k, v] of Object.entries(opsi)) if (v !== undefined && v !== null && v !== '') p.set(k, String(v));
    const s = p.toString();
    return s ? `?${s}` : '';
  };
  const halaman = (d, items) => ({ items, page: d.page, limit: d.limit, total: d.total });

  /* === Publik === */
  const beritaPublik = (b) => ({ id: b.id, slug: b.slug, judul: b.judul, isi: b.isi ?? b.ringkasan ?? '', tanggal_publikasi: keDepan(b.tanggalPublikasi), gambar: b.gambar });
  // Gambar dokumentasi: url dari backend dipakai sebagai "key" (halaman memakai gambar(key), yang meneruskan URL apa adanya).
  // Daftar publik hanya membawa gambar utama + jumlahGambar; kartu memakai gambar[0] dan gambar.length, jadi daftar diisi sepanjang jumlah.
  const gambarDaftar = (d) => Array.from({ length: d.jumlahGambar ?? (d.gambarUtama ? 1 : 0) }, (_, i) => ({ key: i === 0 ? d.gambarUtama : null }));
  const dokPublik = (d) => ({
    id: d.id, slug: d.slug, judul: d.judul, tanggal_publikasi: keDepan(d.tanggalPublikasi),
    gambar: Array.isArray(d.gambar) ? d.gambar.map((g) => ({ key: g.url })) : gambarDaftar(d),
  });

  /** Isi awal dari HTML ber-SEO (sekali pakai, hanya bila slug cocok); null bila tidak ada. */
  function ambilDataAwal(slug) {
    if (dataAwal === undefined) {
      try { dataAwal = JSON.parse(document.getElementById('data-awal')?.textContent ?? 'null'); } catch { dataAwal = null; }
    }
    if (dataAwal && dataAwal.slug === slug) { const d = dataAwal; dataAwal = null; return d; }
    return null;
  }

  const publik = {
    daftarBerita: async ({ q, page, limit } = {}) => { const d = await minta('GET', `/api/berita${kueri({ q, page, limit })}`); return halaman(d, d.items.map(beritaPublik)); },
    daftarDokumentasi: async ({ q, page, limit } = {}) => { const d = await minta('GET', `/api/dokumentasi${kueri({ q, page, limit })}`); return halaman(d, d.items.map(dokPublik)); },
    detailBerita: async (slug) => beritaPublik(ambilDataAwal(slug) ?? await minta('GET', `/api/berita/${encodeURIComponent(slug)}`)),
    detailDokumentasi: async (slug) => dokPublik(ambilDataAwal(slug) ?? await minta('GET', `/api/dokumentasi/${encodeURIComponent(slug)}`)),
    daftarTestimoni: async () => { const d = await minta('GET', '/api/testimoni'); return { items: d.items.map((t) => ({ id: t.id, nama: t.nama, teks: t.teks, posisi: t.posisi, foto: t.foto || null })) }; },
    daftarTim: async () => { const d = await minta('GET', '/api/tim'); return { items: d.items.map((t) => ({ id: t.id, nama: t.nama, posisi: t.posisi, bidang_riset: t.bidangRiset, foto: t.foto })) }; },
  };

  /* === Sesi === */
  const sesi = (s) => { csrfToken = s.csrfToken; return { nama: s.nama, role: s.role, username: s.username }; };

  /* === Admin: pemetaan baris === */
  const SUMBER = ['berita', 'dokumentasi', 'testimoni', 'tim'];
  const adaStatus = (sumber) => sumber !== 'tim';
  const dasar = (b) => ({ id: b.id, version: b.version, dibuat_oleh: b.dibuatOleh, diubah_oleh: b.diubahOleh, dikunci_oleh: b.dikunciOleh ?? null }); // dikunciOleh: nama pemegang kunci orang lain
  const berstatus = (b) => ({ status: b.status, deleted_at: b.dihapusPada ?? null });

  function dariAdmin(sumber, b) {
    if (sumber === 'berita') return { ...dasar(b), ...berstatus(b), slug: b.slug, judul: b.judul, ringkasan: b.ringkasan, isi: b.isi, tanggal_publikasi: keDepan(b.tanggalPublikasi), gambar: b.gambar.key };
    if (sumber === 'dokumentasi') {
      const out = { ...dasar(b), ...berstatus(b), slug: b.slug, judul: b.judul, tanggal_publikasi: keDepan(b.tanggalPublikasi) };
      if (Array.isArray(b.gambar)) out.gambar = b.gambar.map((g) => ({ key: g.key })); // detail
      // daftar: hanya gambar utama + jumlah; kolom tabel memakai gambar[0].key dan gambar.length, jadi diisi sepanjang jumlah
      else out.gambar = Array.from({ length: b.jumlahGambar }, (_, i) => ({ key: i === 0 ? (b.gambarUtama?.key ?? null) : null }));
      return out;
    }
    if (sumber === 'testimoni') return { ...dasar(b), ...berstatus(b), nama: b.nama, teks: b.teks, posisi: b.posisi, foto: b.foto?.key ?? null };
    return { ...dasar(b), nama: b.nama, posisi: b.posisi, bidang_riset: b.bidangRiset, foto: b.foto.key, urutan: b.urutan };
  }

  /** Unggah gambar baru (URL blob:) dan kembalikan key; key yang sudah ada dikembalikan apa adanya. */
  async function pastikanKey(sumber, nilai) {
    if (nilai === null || nilai === undefined || !String(nilai).startsWith('blob:')) return nilai ?? null;
    if (KEY_TERUNGGAH.has(nilai)) return KEY_TERUNGGAH.get(nilai);
    const berkas = BERKAS_BARU.get(nilai);
    if (!berkas) throw galat(400, 'invalid_request', 'Gambar tidak ditemukan. Pilih ulang gambar tersebut.');
    const form = new FormData();
    form.append('sumber', sumber);
    form.append('gambar', berkas);
    const hasil = await minta('POST', '/api/admin/upload', { form });
    KEY_TERUNGGAH.set(nilai, hasil.key);
    return hasil.key;
  }

  /** Badan permintaan buat/ubah menurut kontrak backend (tanpa version; version ditambahkan pemanggil). */
  async function keAdmin(sumber, d) {
    if (sumber === 'berita') return { judul: d.judul, isi: d.isi, tanggalPublikasi: keBelakang(d.tanggal_publikasi), gambar: await pastikanKey(sumber, d.gambar) };
    if (sumber === 'dokumentasi') {
      const gambar = [];
      for (const g of d.gambar_daftar) {
        gambar.push({ key: await pastikanKey(sumber, g.key) });
      }
      return { judul: d.judul, tanggalPublikasi: keBelakang(d.tanggal_publikasi), gambar };
    }
    if (sumber === 'testimoni') return { nama: d.nama, teks: d.teks, posisi: d.posisi, foto: await pastikanKey(sumber, d.foto) };
    return { nama: d.nama, posisi: d.posisi, bidangRiset: d.bidang_riset, foto: await pastikanKey(sumber, d.foto), urutan: Number(d.urutan) };
  }

  const dasarUrl = (sumber) => {
    if (!SUMBER.includes(sumber)) throw galat(404, 'not_found', 'Data tidak ditemukan.');
    return `/api/admin/${sumber}`;
  };

  const admin = {
    /** Sesi saat ini; null bila belum masuk. Mengisi token CSRF. */
    async muatSesi() {
      try { return sesi(await minta('GET', '/api/auth/me')); } catch (g) { if (g.status === 401) return null; throw g; }
    },
    async keluar() { try { await minta('POST', '/api/auth/logout'); } catch { /* sudah keluar atau sesi hilang: tujuan tercapai */ } csrfToken = null; },

    /** Daftar lengkap tanpa halaman (backend: tanpa page/limit = semua item); baris memuat dikunci_oleh bila dikunci orang lain. */
    async adminDaftar(sumber, { trash = false } = {}) {
      const q = adaStatus(sumber) ? { status: trash ? 'trash' : 'semua' } : {};
      const d = await minta('GET', `${dasarUrl(sumber)}${kueri(q)}`);
      const items = d.items.map((b) => dariAdmin(sumber, b));
      return d.total === undefined ? { items, page: 1, limit: items.length || 1, total: items.length } : halaman(d, items);
    },
    /** Log aktivitas (baca-saja): { page, username, jenis, dari, sampai } -> { items, page, limit, total, usernames }. */
    async adminLog({ page, username, jenis, dari, sampai } = {}) {
      const d = await minta('GET', `/api/admin/log-aktivitas${kueri({ page, username, jenis, dari, sampai })}`);
      const items = d.items.map((e) => ({ id: e.id, terjadi_pada: e.terjadiPada, pengguna_sub: e.penggunaSub, username: e.username, jenis: e.jenis, aksi: e.aksi, modul: e.modul, item_id: e.itemId, label_item: e.labelItem }));
      return { items, page: d.page, limit: d.limit, total: d.total, usernames: d.usernames };
    },
    async adminDetail(sumber, id) { return dariAdmin(sumber, await minta('GET', `${dasarUrl(sumber)}/${id}`)); },
    async adminTambah(sumber, data) { return dariAdmin(sumber, await minta('POST', dasarUrl(sumber), { json: await keAdmin(sumber, data) })); },
    async adminUbah(sumber, id, data, version) { return dariAdmin(sumber, await minta('PUT', `${dasarUrl(sumber)}/${id}`, { json: { ...await keAdmin(sumber, data), version } })); },
    async adminPost(sumber, id, version) { return dariAdmin(sumber, await minta('POST', `${dasarUrl(sumber)}/${id}/post`, { json: { version } })); },
    async adminRetract(sumber, id, version) { return dariAdmin(sumber, await minta('POST', `${dasarUrl(sumber)}/${id}/retract`, { json: { version } })); },
    async adminHapus(sumber, id) { await minta('DELETE', `${dasarUrl(sumber)}/${id}`); return {}; }, // berita/dokumentasi/testimoni: trash; tim: permanen
    async adminPulihkan(sumber, id) { return dariAdmin(sumber, await minta('POST', `${dasarUrl(sumber)}/${id}/restore`)); },
    async adminHapusPermanen(sumber, id) { await minta('DELETE', `${dasarUrl(sumber)}/${id}/permanen`); return {}; },

    /* Soft lock: POST pasang, PUT heartbeat, DELETE lepas. 423 locked = dipegang orang lain. */
    async kunciAmbil(sumber, id) {
      try {
        const k = await minta('POST', `${dasarUrl(sumber)}/${id}/lock`);
        kunciSaya = { sumber, id };
        return { milik: 'saya', heartbeatDetik: k.heartbeatDetik };
      } catch (g) {
        if (g.status === 423 && g.error.code === 'locked') return { milik: 'lain', nama: g.pemegang?.nama ?? 'pengguna lain' };
        throw g;
      }
    },
    async kunciHeartbeat() {
      if (!kunciSaya) throw galat(423, 'lock_hilang', 'Kunci edit sudah berakhir; buka kembali form edit.');
      return minta('PUT', `${dasarUrl(kunciSaya.sumber)}/${kunciSaya.id}/lock`);
    },
    /** Lepas kunci yang dipegang (best effort, tanpa menunggu). */
    kunciLepas() {
      if (!kunciSaya) return;
      const { sumber, id } = kunciSaya;
      kunciSaya = null;
      minta('DELETE', `${dasarUrl(sumber)}/${id}/lock`, { keepalive: true }).catch(() => {});
    },

    /** Daftarkan File pilihan admin; kembalikan URL blob: sebagai nilai sementara field gambar. */
    blob(berkas) { const u = URL.createObjectURL(berkas); BERKAS_BARU.set(u, berkas); return u; },
  };

  return { ...publik, ...admin };
})();
