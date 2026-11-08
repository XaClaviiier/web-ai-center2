/* PANEL SIMULASI. Hanya untuk tinjauan, dibuang saat implementasi.
   Mengubah SIMULASI lalu mengirim event "simulasi:ubah" agar halaman memuat ulang datanya. */
(() => {
  const sisi = document.body.dataset.sisi;
  const buat = (tag, attr = {}, ...anak) => {
    const n = document.createElement(tag);
    for (const [k, v] of Object.entries(attr)) n[k] = v;
    n.append(...anak);
    return n;
  };
  const ubah = () => document.dispatchEvent(new CustomEvent('simulasi:ubah'));

  const KEADAAN = [['normal', 'Normal'], ['loading', 'Loading'], ['kosong', 'Kosong'], ['error', 'Error'], ['429', '429']];
  const tombolKeadaan = KEADAAN.map(([nilai, label]) => {
    const t = buat('button', { type: 'button', textContent: label, onclick: () => { SIMULASI.keadaan = nilai; segarkan(); ubah(); } });
    t.dataset.nilai = nilai;
    return t;
  });
  const grupKeadaan = buat('div', { className: 'sim-grup' }, ...tombolKeadaan);
  function segarkan() {
    tombolKeadaan.forEach((t) => t.setAttribute('aria-pressed', String(t.dataset.nilai === SIMULASI.keadaan)));
    if (kunci) kunci.setAttribute('aria-pressed', String(SIMULASI.kunciLain));
  }

  const isi = buat('div', { className: 'sim-isi', hidden: true },
    buat('strong', { textContent: 'Simulasi, bukan bagian aplikasi' }),
    buat('p', { textContent: 'Keadaan data' }), grupKeadaan);

  let kunci = null;
  if (sisi === 'admin') {
    const aksi = (label, fn) => buat('button', { type: 'button', textContent: label, onclick: fn });
    kunci = aksi('Item dikunci pengguna lain', () => { SIMULASI.kunciLain = !SIMULASI.kunciLain; segarkan(); ubah(); });
    isi.append(buat('p', { textContent: 'Admin' }), buat('div', { className: 'sim-grup sim-grup--tegak' },
      kunci,
      aksi('409 pada simpan berikutnya', () => { SIMULASI.paksa409 = true; document.dispatchEvent(new CustomEvent('simulasi:catat', { detail: '409 aktif untuk simpan berikutnya.' })); }),
      aksi('Heartbeat gagal sekarang', () => { SIMULASI.heartbeatGagal = true; document.dispatchEvent(new CustomEvent('simulasi:heartbeat')); }),
      aksi('Sesi berakhir (401)', () => { SIMULASI.paksa401 = true; ubah(); })));
  }

  const buka = buat('button', { type: 'button', className: 'sim-buka', textContent: 'Simulasi' });
  buka.setAttribute('aria-expanded', 'false');
  buka.onclick = () => {
    isi.hidden = !isi.hidden;
    buka.setAttribute('aria-expanded', String(!isi.hidden));
  };
  document.body.append(buat('aside', { className: 'sim-panel', ariaLabel: 'Panel simulasi' }, isi, buka));
  segarkan();
})();
