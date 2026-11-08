/* Demo JejakAlumni. Semua angka dan teks "AI" adalah data siap pakai di MOCK.konten.jejak_alumni;
   tidak ada inferensi. Persentase dan total selalu dihitung dari jumlah orang agar tidak saling bertentangan. */
const JejakAlumni = (() => {
  const D = MOCK.konten.jejak_alumni;
  const SEMUA = 'semua';
  const TAHUN = Object.keys(D.angkatan).sort().reverse();
  const JUMLAH = ['lulusan', 'responden', 'terserap', 'bekerja', 'kuliah', 'wirausaha', 'belum', 'sesuai'];
  const STATUS = [['bekerja', 'Bekerja'], ['kuliah', 'Kuliah'], ['wirausaha', 'Wirausaha'], ['belum', 'Belum bekerja']];
  const persen = (n, d) => Math.round((100 * n) / d);
  const bulan = new Intl.NumberFormat('id-ID', { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format;

  // Penjaga data simulasi: pelanggaran terlihat di konsol saat halaman dibuka.
  for (const [th, a] of Object.entries(D.angkatan)) {
    for (const [j, b] of Object.entries(a.per_jurusan)) {
      const status = STATUS.map(([k]) => persen(b[k], b.responden)).reduce((x, y) => x + y);
      console.assert(
        b.bekerja + b.kuliah + b.wirausaha + b.belum === b.responden && status === 100
          && b.responden <= b.lulusan && b.bekerja <= b.terserap && b.sesuai <= b.bekerja,
        `Data JejakAlumni tidak konsisten: ${th} ${j}`);
    }
  }

  function hitung(tahun, jurusan) {
    const a = D.angkatan[tahun];
    if (!a) return null;
    const baris = jurusan === SEMUA ? Object.values(a.per_jurusan) : [a.per_jurusan[jurusan]];
    const total = { median: jurusan === SEMUA ? a.median_semua : baris[0].median };
    for (const k of JUMLAH) total[k] = baris.reduce((s, b) => s + b[k], 0);
    return total;
  }

  const KPI = [
    { label: 'Tingkat respons kuesioner', nilai: (t) => persen(t.responden, t.lulusan), dasar: (t) => `${t.responden} dari ${t.lulusan} lulusan` },
    { label: 'Masa tunggu kerja (median)', nilai: (t) => t.median, dasar: (t) => `${t.terserap} responden terserap`, bulan: true },
    { label: 'Terserap dunia kerja', nilai: (t) => persen(t.terserap, t.responden), dasar: (t) => `${t.terserap} dari ${t.responden} responden` },
    { label: 'Kerja sesuai bidang', nilai: (t) => persen(t.sesuai, t.bekerja), dasar: (t) => `${t.sesuai} dari ${t.bekerja} yang bekerja` },
  ];

  /** Selisih dihitung dari nilai yang tampil. Masa tunggu lebih singkat berarti membaik. */
  function pembanding(k, ini, lalu, tahunLalu) {
    if (!lalu) return el('p.ja-kpi__banding', null, 'Tidak ada data angkatan sebelumnya');
    const selisih = Math.round((k.nilai(ini) - k.nilai(lalu)) * 10) / 10;
    if (!selisih) return el('p.ja-kpi__banding', null, `Sama dengan ${tahunLalu}`);
    const besar = Math.abs(selisih);
    const teks = k.bulan
      ? `${bulan(besar)} bln ${selisih < 0 ? 'lebih singkat' : 'lebih lama'} dari ${tahunLalu}`
      : `${selisih > 0 ? '+' : '−'}${besar} poin dari ${tahunLalu}`;
    const membaik = k.bulan ? selisih < 0 : selisih > 0;
    return el(`p.ja-kpi__banding.ja-kpi__banding--${membaik ? 'baik' : 'buruk'}`, null,
      el('span', { 'aria-hidden': 'true' }, selisih > 0 ? '▲ ' : '▼ '), teks);
  }

  function kartuKpi(k, ini, lalu, tahunLalu) {
    return el('div.ja-panel.ja-kpi__kartu', null,
      el('h3.ja-kpi__label', null, k.label),
      el('p.ja-kpi__nilai', null, k.bulan ? bulan(k.nilai(ini)) : `${k.nilai(ini)}%`, k.bulan && el('small', null, ' bln')),
      el('p.field__bantuan', null, k.dasar(ini)),
      pembanding(k, ini, lalu, tahunLalu));
  }

  function grafikPenyerapan(tahun, jurusan) {
    const a = D.angkatan[tahun].per_jurusan;
    return el('section.ja-panel', { 'aria-labelledby': 'ja-j-serap' },
      el('h3.ja-judul', { id: 'ja-j-serap' }, 'Terserap dunia kerja per jurusan'),
      el('p.field__bantuan', null, `Persentase responden yang bekerja dalam 6 bulan setelah lulus · angkatan ${tahun}`),
      el('ul.ja-batang', { 'data-pilih': jurusan === SEMUA ? null : 'ya' }, D.jurusan.map((j) => {
        const nilai = persen(a[j].terserap, a[j].responden);
        return el('li', { 'aria-current': j === jurusan ? 'true' : null },
          el('span.ja-batang__label', null, j),
          el('span.ja-batang__jalur', { 'aria-hidden': 'true' }, el('span.ja-batang__isi', { style: `width:${nilai}%` })),
          el('span.ja-batang__nilai', null, `${nilai}%`));
      })));
  }

  function grafikStatus(t, tahun, jurusan) {
    return el('section.ja-panel', { 'aria-labelledby': 'ja-j-status' },
      el('h3.ja-judul', { id: 'ja-j-status' }, 'Status alumni saat ini'),
      el('p.field__bantuan', null, `Saat mengisi kuesioner · angkatan ${tahun} · ${jurusan === SEMUA ? 'semua jurusan' : jurusan}`),
      el('div.ja-segmen', { 'aria-hidden': 'true' }, STATUS.map(([k]) => el(`span.ja-warna--${k}`, { style: `flex-grow:${t[k]}` }))),
      el('ul.ja-status', null, STATUS.map(([k, label]) => el('li', null,
        el(`span.ja-status__warna.ja-warna--${k}`, { 'aria-hidden': 'true' }),
        el('span', null, label),
        el('strong', null, `${persen(t[k], t.responden)}%`),
        el('span.ja-status__jumlah', null, `${t[k]} orang`)))));
  }

  function panelAi(tahun, jurusan) {
    const a = D.angkatan[tahun];
    const cakupan = `Cakupan: seluruh jurusan, angkatan ${tahun}${jurusan === SEMUA ? '' : ' (tidak mengikuti filter jurusan)'}.`;
    return [
      el('section.ja-panel', { 'aria-labelledby': 'ja-j-tema' },
        el('div.ja-panel__kepala', null,
          el('h3.ja-judul', { id: 'ja-j-tema' }, 'Tema jawaban terbuka'),
          el('span.badge.badge--simulasi', null, 'Simulasi Analisis AI')),
        el('p.field__bantuan', null, 'Contoh bagaimana AI dapat membantu mengelompokkan jawaban terbuka alumni menjadi tema yang lebih mudah ditinjau.'),
        el('ul.ja-tema', null, a.tema.map((t) => el('li', null,
          el(`span.badge.ja-lencana--${t.jenis}`, null, t.jenis === 'positif' ? 'Positif' : 'Keluhan'),
          el('span', null, t.teks),
          el('span.ja-tema__jumlah', null, `${t.sebutan} sebutan`)))),
        el('p.field__bantuan', null, cakupan)),
      el('section.ja-panel.ja-panel--ringkasan', { 'aria-labelledby': 'ja-j-ringkas' },
        el('div.ja-panel__kepala', null,
          el('h3.ja-judul', { id: 'ja-j-ringkas' }, 'Ringkasan untuk rapat kurikulum'),
          el('span.badge.badge--simulasi', null, 'Simulasi Ringkasan AI')),
        el('p', null, a.ringkasan),
        el('p.field__bantuan', null, cakupan)),
    ];
  }

  function bingkaiDasbor() {
    let tahun = TAHUN[0];
    let jurusan = SEMUA;
    const cakupan = el('p.ja-cakupan', { role: 'status' });
    const wadahKpi = el('div.ja-kpi');
    const wadahGrafik = el('div.ja-grafik');
    const wadahAi = el('div.ja-ai');

    function segarkan() {
      const ini = hitung(tahun, jurusan);
      const tahunLalu = String(Number(tahun) - 1);
      const lalu = hitung(tahunLalu, jurusan);
      cakupan.textContent = `Angkatan ${tahun} · ${jurusan === SEMUA ? 'Semua jurusan' : jurusan}: ${ini.responden} responden dari ${ini.lulusan} lulusan`;
      wadahKpi.replaceChildren(...KPI.map((k) => kartuKpi(k, ini, lalu, tahunLalu)));
      wadahGrafik.replaceChildren(grafikPenyerapan(tahun, jurusan), grafikStatus(ini, tahun, jurusan));
      wadahAi.replaceChildren(...panelAi(tahun, jurusan));
    }

    const pilihan = (id, label, opsi, ubah) => el('div.field', null,
      el('label', { for: id }, label),
      el('select.input', { id, onchange: (e) => { ubah(e.target.value); segarkan(); } },
        opsi.map(([nilai, teks]) => el('option', { value: nilai }, teks))));

    const dasbor = el('section.ja-bingkai', { 'aria-labelledby': 'ja-j-dasbor' },
      el('div.ja-panel__kepala', null,
        el('h2.ja-bingkai__judul', { id: 'ja-j-dasbor' }, 'Dasbor tracer study'),
        el('span.ja-demo', null, 'Demo'),
        el('span.badge.badge--simulasi', null, D.status)),
      el('ul.ja-tab', null,
        el('li', null, el('span.ja-tab__aktif', { 'aria-current': 'true' }, 'Ringkasan')),
        ['Responden', 'Mitra kerja'].map((t) => el('li', null,
          el('button.ja-tab__mati', { type: 'button', disabled: true }, t, el('small', null, 'Belum tersedia dalam demo'))))),
      el('div.ja-filter', null,
        pilihan('ja-angkatan', 'Angkatan lulus', TAHUN.map((t) => [t, t]), (v) => { tahun = v; }),
        pilihan('ja-jurusan', 'Jurusan', [[SEMUA, 'Semua jurusan'], ...D.jurusan.map((j) => [j, j])], (v) => { jurusan = v; })),
      cakupan,
      wadahKpi,
      el('details.ja-definisi', null,
        el('summary', null, 'Definisi indikator'),
        el('dl', null, D.definisi.map((d) => el('div', null, el('dt', null, d.istilah), el('dd', null, d.arti))))),
      wadahGrafik,
      wadahAi);
    segarkan();
    return dasbor;
  }

  function penjelasan() {
    const p = D.penjelasan;
    return el('aside.ja-penjelasan', { 'aria-labelledby': 'ja-j-tentang' },
      el('div', null,
        el('p.tentang-label', null, 'Data'),
        el('h2', { id: 'ja-j-tentang' }, 'Tentang JejakAlumni')),
      el('div', null, el('h3', null, 'Masalah nyata di sekolah'), el('p.teks-sekunder', null, p.masalah)),
      el('div', null,
        el('h3', null, 'Cara kerjanya'),
        el('ol.ja-langkah', null, p.langkah.map((l, i) => el('li', null, el('span.alur__nomor', { 'aria-hidden': 'true' }, String(i + 1)), el('span.teks-sekunder', null, l)))),
        el('p.field__bantuan', null, p.langkah_catatan)),
      el('div', null,
        el('h3', null, 'Data sekolah yang dipakai'),
        el('ul.ja-chip', null, p.data.map((d) => el('li', null, d))),
        el('p.field__bantuan', null, p.data_catatan)),
      el('div', null, el('h3', null, 'Yang berubah'), el('p.teks-sekunder', null, p.hasil)));
  }

  function halaman() {
    return el('div.ja', null,
      el('header.wadah.ja-kepala', null,
        el('a.tautan-panah.ja-kembali', { href: 'inference.html#portofolio' }, ikon('kiri', 18), 'Kembali ke Inference and Solutions'),
        el('div.ja-kepala__judul', null,
          el('h1', null, 'JejakAlumni'),
          el('span.badge.badge--simulasi', null, D.status)),
        el('p.ja-kepala__sub', null, 'Tracer Study Otomatis'),
        el('div.banner', null, el('div.banner__teks', null, ikon('peringatan', 18), el('span', null, D.disclaimer)))),
      el('div.wadah.ja-tata', null, bingkaiDasbor(), penjelasan()));
  }

  return { halaman };
})();
