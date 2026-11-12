/* Demo DocNext. Dokumen sintetis di KONTEN.document_intelligence.
   Ekstraksi dan keterkaitan deskripsi disiapkan (tidak ada model AI); hanya perbandingan yang dihitung di sini. */
const DocNext = (() => {
  const D = KONTEN.document_intelligence;
  const NAMA_DOK = { po: 'Purchase Order', penerimaan: 'Dokumen Penerimaan', invoice: 'Invoice' };
  const NAMA_FIELD = { refPo: 'Referensi PO', pemasok: 'Pemasok', kode: 'Kode barang', deskripsi: 'Deskripsi barang', qty: 'Kuantitas', harga: 'Harga satuan', subtotal: 'Subtotal' };
  const STATUS = { sesuai: ['✓', 'Sesuai'], ditinjau: ['!', 'Perlu Ditinjau'], verifikasi: ['?', 'Perlu Verifikasi'] };
  const rp = (n) => `Rp${n.toLocaleString('id-ID')}`;

  /** Nilai field seperti tertulis di dokumen; pada PO, referensinya adalah nomor PO itu sendiri. */
  function nilai(dok, jenis, kunci) {
    if (kunci === 'refPo') return jenis === 'po' ? dok.nomor : dok.refPo;
    if (dok[kunci] === undefined) return null;
    if (kunci === 'qty') return `${dok.qty} ${dok.satuan}`;
    if (kunci === 'harga' || kunci === 'subtotal') return rp(dok[kunci]);
    return dok[kunci];
  }

  /** Aturan bisnis deterministik. Hanya baris identitas yang memakai interpretasi makna. */
  function periksa(s) {
    const { po, penerimaan: gr, invoice: inv } = s;
    const sama = (...a) => a.every((x) => x === a[0]);
    const nilaiPer = (kunci) => ['po', 'penerimaan', 'invoice'].map((j) => nilai(s[j], j, kunci));
    const baris = [
      { kunci: 'refPo', oleh: 'aturan', status: sama(po.nomor, gr.refPo, inv.refPo) ? 'sesuai' : 'ditinjau' },
      { kunci: 'pemasok', oleh: 'aturan', status: sama(po.pemasok, gr.pemasok, inv.pemasok) ? 'sesuai' : 'ditinjau' },
      { kunci: 'deskripsi', oleh: 'ai', status: sama(po.kode, gr.kode, inv.kode) && sama(po.deskripsi, gr.deskripsi, inv.deskripsi) ? 'sesuai' : 'verifikasi' },
      { kunci: 'qty', oleh: 'aturan', status: sama(po.qty, gr.qty, inv.qty) ? 'sesuai' : 'ditinjau' },
      { kunci: 'harga', oleh: 'aturan', status: po.harga === inv.harga ? 'sesuai' : 'ditinjau' },
      { kunci: 'subtotal', oleh: 'aturan', status: inv.subtotal === gr.qty * inv.harga ? 'sesuai' : 'ditinjau',
        catatan: inv.subtotal === gr.qty * inv.harga ? null : `Subtotal invoice berbeda dari jumlah diterima dikali harga satuan (${rp(gr.qty * inv.harga)}).` },
    ].map((b) => ({ ...b, nilai: nilaiPer(b.kunci) }));
    const status = baris.some((b) => b.status === 'ditinjau') ? 'ditinjau' : baris.some((b) => b.status === 'verifikasi') ? 'verifikasi' : 'sesuai';
    return { baris, status, selisihUnit: inv.qty - gr.qty, selisihNilai: (inv.qty - gr.qty) * inv.harga };
  }

  // Penjaga data sintetis: pelanggaran terlihat di konsol saat halaman dibuka.
  D.skenario.forEach((s) => {
    console.assert(s.po.qty * s.po.harga === s.po.subtotal && s.invoice.qty * s.invoice.harga === s.invoice.subtotal, `Subtotal tidak konsisten: ${s.label}`);
    console.assert(periksa(s).status === s.harapan, `Status tidak sesuai harapan: ${s.label}`);
  });

  const lencana = (status) => el(`span.badge.dn-status.dn-status--${status}`, null, el('span', { 'aria-hidden': 'true' }, `${STATUS[status][0]} `), STATUS[status][1]);
  const tagOleh = (oleh) => (oleh === 'ai' ? el('span.dn-tag.dn-tag--ai', null, 'Simulasi AI') : el('span.dn-tag', null, 'Aturan Bisnis'));
  const f = (kunci, teks) => el('span.dn-f', { 'data-field': kunci }, teks);
  const judulDok = (jenis, s, teks) => el('header.dn-dok__kepala', null, el('div', null, el('strong', null, jenis === 'invoice' ? f('pemasok', s.pemasok) : D.pembeli), el('p.field__bantuan', null, 'Dokumen sintetis')), el('h4', null, teks));
  const metaDok = (pasangan) => el('dl.dn-dok__meta', null, pasangan.map(([l, isi]) => el('div', null, el('dt', null, l), el('dd', null, isi))));

  function tabelBarang(kolom) {
    return el('table.dn-tabel', null,
      el('thead', null, el('tr', null, kolom.map(([k]) => el('th', { scope: 'col' }, k)))),
      el('tbody', null, el('tr', null, kolom.map(([k, isi]) => el('td', { 'data-kolom': k }, isi)))));
  }

  /* Tata letak tiap dokumen sengaja berbeda, seperti dokumen dari pihak yang berbeda. */
  const TEMPLAT = {
    po: (d) => { const L = D.label_dokumen.po; return [
      judulDok('po', d, 'PURCHASE ORDER'),
      metaDok([[L.refPo, f('refPo', d.nomor)], ['Tanggal PO', d.tanggal], [L.pemasok, f('pemasok', d.pemasok)], ['Pembeli', D.pembeli]]),
      tabelBarang([[L.kode, f('kode', d.kode)], [L.deskripsi, f('deskripsi', d.deskripsi)], [L.qty, f('qty', `${d.qty} ${d.satuan}`)], [L.harga, f('harga', rp(d.harga))], [L.subtotal, f('subtotal', rp(d.subtotal))]]),
    ]; },
    penerimaan: (d) => { const L = D.label_dokumen.penerimaan; return [
      judulDok('penerimaan', d, 'BUKTI PENERIMAAN BARANG'),
      metaDok([['No. Penerimaan', d.nomor], ['Tgl. Terima', d.tanggal], [L.refPo, f('refPo', d.refPo)], [L.pemasok, f('pemasok', d.pemasok)], ['Lokasi', D.lokasi]]),
      el('div.dn-dok__blok', null, metaDok([[L.kode, f('kode', d.kode)], [L.deskripsi, f('deskripsi', d.deskripsi)], [L.qty, f('qty', `${d.qty} ${d.satuan}`)]])),
    ]; },
    invoice: (d) => { const L = D.label_dokumen.invoice; return [
      judulDok('invoice', d, 'INVOICE'),
      metaDok([['Invoice No.', d.nomor], ['Tanggal', d.tanggal], [L.refPo, f('refPo', d.refPo)], ['Ditagihkan kepada', D.pembeli]]),
      tabelBarang([[L.kode, f('kode', d.kode)], [L.deskripsi, f('deskripsi', d.deskripsi)], [L.qty, f('qty', `${d.qty} ${d.satuan}`)], [L.harga, f('harga', rp(d.harga))]]),
      el('p.dn-dok__total', null, `${L.subtotal}: `, f('subtotal', rp(d.subtotal))),
      el('p.field__bantuan', null, 'Pajak tidak dihitung dalam demo.'),
    ]; },
  };

  function ruangKerja() {
    let s = D.skenario[0];
    let jenis = 'po';
    let field = null;

    const banner = el('div.dn-banner', null);
    const ringkasanHasil = el('p.sr-only', { role: 'status', 'aria-atomic': 'true' });
    const sumber = el('p.dn-sumber', { role: 'status' });
    const pratinjau = el('article.dn-dok', { 'aria-label': 'Pratinjau dokumen sintetis' });
    const daftarField = el('ul.dn-field');
    const daftarBanding = el('ul.dn-banding');
    const daftarTemuan = el('div.dn-temuan');

    const pilihDok = el('fieldset.dn-pilih', null, el('legend', null, 'Dokumen'), Object.entries(NAMA_DOK).map(([j, nama]) => el('label.dn-chip', null,
      el('input', { type: 'radio', name: 'dn-dokumen', value: j, checked: j === jenis, onchange: () => { jenis = j; gambarDokumen(); } }), el('span', null, nama))));

    function sorot() {
      pratinjau.querySelectorAll('.dn-f').forEach((n) => { n.classList.toggle('dn-f--aktif', n.dataset.field === field); });
      daftarField.querySelectorAll('button').forEach((b) => b.setAttribute('aria-pressed', String(b.dataset.field === field)));
      const nilaiAktif = field && nilai(s[jenis], jenis, field);
      if (!field) sumber.textContent = `Menampilkan ${NAMA_DOK[jenis]}. Pilih hasil ekstraksi untuk melihat sumbernya.`;
      else if (nilaiAktif === null) sumber.textContent = `${NAMA_FIELD[field]} tidak tercantum pada ${NAMA_DOK[jenis]}.`;
      else sumber.textContent = `Sumber: ${NAMA_DOK[jenis]}, label "${D.label_dokumen[jenis][field]}", nilai ${nilaiAktif}.`;
    }

    function gambarDokumen() {
      pratinjau.replaceChildren(...TEMPLAT[jenis](s[jenis]));
      daftarField.replaceChildren(...Object.keys(D.label_dokumen[jenis]).map((k) => el('li', null,
        el('button.dn-field__tombol', { type: 'button', 'data-field': k, 'aria-pressed': 'false', onclick: () => { field = k; sorot(); } },
          el('span.dn-field__nama', null, NAMA_FIELD[k]),
          el('span.dn-field__nilai', null, nilai(s[jenis], jenis, k)),
          el('span.field__bantuan', null, `Label di dokumen: ${D.label_dokumen[jenis][k]}`)))));
      pilihDok.querySelector(`input[value="${jenis}"]`).checked = true;
      sorot();
    }

    function keSumber(j, k) {
      jenis = j;
      field = k;
      gambarDokumen();
      if (pratinjau.getBoundingClientRect().top < 0) pratinjau.scrollIntoView({ block: 'center' });
    }

    function gambarSkenario() {
      const h = periksa(s);
      jenis = 'po';
      field = null;
      banner.replaceChildren(lencana(h.status), el('div', null, el('p', null, s.penjelasan), el('p.field__bantuan', null, s.catatan)));
      ringkasanHasil.textContent = `${s.label}. ${banner.querySelector('.dn-badge')?.textContent || banner.firstElementChild.textContent}. ${s.penjelasan}`;
      daftarBanding.replaceChildren(...h.baris.map((b) => el('li', null,
        el('div.dn-banding__kepala', null, el('strong', null, b.kunci === 'deskripsi' ? 'Identitas barang' : NAMA_FIELD[b.kunci]), lencana(b.status)),
        el('dl.dn-banding__nilai', null, Object.values(NAMA_DOK).map((nama, i) => el('div', null, el('dt', null, nama), el('dd', null, b.nilai[i] ?? 'Tidak tercantum')))),
        b.oleh === 'ai' && s.keterkaitan && el('p.dn-keterkaitan', null, s.keterkaitan),
        b.catatan && el('p.dn-keterkaitan', null, b.catatan),
        tagOleh(b.oleh))));
      daftarTemuan.replaceChildren(...(s.temuan.length ? s.temuan.map((t) => el('article.dn-temuan__kartu', null,
        el('h4', null, t.judul),
        el('dl.dn-banding__nilai', null, t.dokumen.map((j) => el('div', null, el('dt', null, NAMA_DOK[j]), el('dd', null, nilai(s[j], j, t.field))))),
        t.field === 'qty' && el('p', null, `Selisih ${h.selisihUnit} ${s.invoice.satuan}, senilai ${rp(h.selisihNilai)} sebelum pajak.`),
        el('p', null, t.mengapa),
        el('p.dn-temuan__label', null, 'Yang perlu diperiksa peninjau'),
        el('ul.dn-daftar', null, t.periksa.map((p) => el('li', null, p))),
        el('div.dn-temuan__aksi', null, t.dokumen.map((j) => tombol(`Lihat di ${NAMA_DOK[j]}`, 'kedua', { class: 'tombol tombol--kedua tombol--kecil', onclick: () => keSumber(j, t.field) })))))
        : [el('p.teks-sekunder', null, 'Tidak ada temuan berdasarkan aturan pemeriksaan yang disimulasikan.')]));
      gambarDokumen();
    }

    const pilihSkenario = el('fieldset.dn-pilih', null, el('legend', null, 'Skenario transaksi'), D.skenario.map((x, i) => el('label.dn-chip', null,
      el('input', { type: 'radio', name: 'dn-skenario', value: String(i), checked: i === 0, onchange: () => { s = x; gambarSkenario(); } }), el('span', null, `${i + 1}. ${x.label}`))));

    gambarSkenario();
    return el('div.dn-kerja', null,
      pilihSkenario,
      ringkasanHasil,
      banner,
      el('div.dn-kerja__kolom', null,
        el('div.dn-kerja__dokumen', null, pilihDok, sumber, pratinjau),
        el('div.dn-kerja__analisis', null,
          el('section.dn-panel', { 'aria-labelledby': 'dn-j-ekstraksi' },
            el('div.dn-panel__kepala', null, el('h3', { id: 'dn-j-ekstraksi' }, 'Hasil ekstraksi'), tagOleh('ai')),
            daftarField,
            el('p.field__bantuan', null, D.catatan_ekstraksi)),
          el('section.dn-panel', { 'aria-labelledby': 'dn-j-banding' },
            el('h3', { id: 'dn-j-banding' }, 'Perbandingan'),
            daftarBanding,
            el('details.dn-lipat', null,
              el('summary', null, 'Apa bedanya AI dan aturan bisnis?'),
              el('p', null, 'Simulasi AI menandai informasi yang dalam penerapan nyata dapat dibantu model, seperti membaca dokumen dan menafsirkan deskripsi. Aturan Bisnis adalah perhitungan biasa yang dijalankan browser pada demo ini.'))),
          el('section.dn-panel', { 'aria-labelledby': 'dn-j-temuan' },
            el('h3', { id: 'dn-j-temuan' }, 'Temuan untuk peninjau'),
            daftarTemuan))));
  }

  function bagian() {
    const p = D.deployment;
    return [
      el('header.wadah.dn-kepala', null,
        el('a.tautan-panah.dn-kembali', { href: '/layanan/inference-solutions#portofolio' }, ikon('kiri', 18), 'Kembali ke Inference and Solutions'),
        el('h1', null, D.judul),
        el('p.dn-kepala__sub', null, D.subjudul),
        el('p.teks-sekunder.dn-kepala__teks', null, D.deskripsi),
        el('a.tautan-panah', { href: '#demo' }, 'Langsung ke Demo', ikon('panah', 18)),
        el('span.badge.badge--simulasi', null, D.status),
        el('div.banner', null, el('div.banner__teks', null, ikon('peringatan', 18), el('span', null, D.disclaimer)))),
      el('section.bagian.wadah', { 'aria-labelledby': 'dn-j-masalah' },
        el('h2', { id: 'dn-j-masalah' }, 'Satu transaksi, banyak dokumen untuk diperiksa.'),
        el('ul.alur', null, D.masalah.map((m) => el('li', null, el('h3', null, m.judul), el('p.teks-sekunder', null, m.teks))))),
      el('section.bagian.bagian--muda.dn-demo', { id: 'demo', 'aria-labelledby': 'dn-j-demo' }, el('div.wadah', null,
        el('h2', { id: 'dn-j-demo' }, 'Telusuri tiga contoh transaksi.'),
        ruangKerja())),
      el('section.bagian.wadah', { 'aria-labelledby': 'dn-j-lapisan' },
        el('h2', { id: 'dn-j-lapisan' }, 'Peran AI, Aturan Bisnis, dan Peninjau.'),
        D.lapisan.pengantar.map((t) => el('p.bagian__sub', null, t)),
        el('div.dn-tiga', null, D.lapisan.kelompok.map((k) => el('article.jenis', null,
          el('p.tentang-label', null, k.label),
          el('h3', null, k.judul),
          el('ul.dn-daftar', null, k.butir.map((b) => el('li', null, b))))))),
      el('section.bagian.bagian--muda', { 'aria-labelledby': 'dn-j-penerapan' }, el('div.wadah', null,
        el('h2', { id: 'dn-j-penerapan' }, 'Satu Pendekatan, Berbagai Kebutuhan Dokumen.'),
        el('p.bagian__sub', null, D.penerapan.teks),
        el('ul.dn-contoh', null, D.penerapan.contoh.map((c) => el('li', null, c))),
        el('p.field__bantuan', null, D.penerapan.catatan))),
      el('section.bagian.wadah', { 'aria-labelledby': 'dn-j-deploy' },
        el('h2', { id: 'dn-j-deploy' }, 'AI yang Dapat Dirancang untuk Infrastruktur Internal Perusahaan.'),
        el('p.bagian__sub', null, p.teks),
        el('figure.dn-alur', null,
          el('ol', null, p.alur.map((a) => el('li', null, a))),
          el('figcaption.field__bantuan', null, p.alur_label)),
        el('div.jenis-grid', null, p.opsi.map((o) => el('article.jenis', null, el('h3', null, o.judul), el('p.teks-sekunder', null, o.teks)))),
        el('details.dn-lipat', null,
          el('summary', null, 'Hal yang perlu dipertimbangkan'),
          el('ul.dn-daftar', null, p.prasyarat.map((x) => el('li', null, x)))),
        el('p.teks-sekunder', null, p.catatan)),
      el('section.bagian.bagian--muda', { 'aria-labelledby': 'dn-j-batasan' }, el('div.wadah', null,
        el('h2', { id: 'dn-j-batasan' }, 'Batasan demo.'),
        el('ul.dn-daftar', null, D.batasan.map((b) => el('li', null, b))))),
    ];
  }

  return { bagian };
})();
