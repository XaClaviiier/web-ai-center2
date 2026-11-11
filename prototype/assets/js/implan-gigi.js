/* Demo Perencanaan Implan Gigi. Ilustrasi konseptual statis dari <template> di halaman;
   kontrol hanya mengganti lapisan yang sudah disiapkan. Tidak ada inferensi atau input data medis. */
const ImplanGigi = (() => {
  const D = MOCK.konten.implan_gigi;

  const legenda = (t) => el('ul.ig-legenda', null, t.legenda.map((l) => el('li', null, el(`span.ig-sw.ig-sw--${l.kelas}`, { 'aria-hidden': 'true' }), l.teks)));

  /** Satu tahap: gambar, kontrol, dan penjelasan. Status kontrol tersimpan di elemen input masing-masing. */
  function tahap(t, i) {
    const svg = document.getElementById(`ig-gambar-${i + 1}`).content.firstElementChild.cloneNode(true);
    const keterangan = el('figcaption.ig-gambar__keterangan', { 'aria-live': t.kontrol.some((c) => c.tipe === 'pilihan') ? 'polite' : null }, t.keterangan ?? '');

    function terapkan() {
      const aktif = [];
      for (const c of t.kontrol) {
        if (c.tipe === 'toggle') {
          const nyala = kontrol.querySelector(`[data-kunci="${c.kunci}"]`).checked;
          svg.querySelectorAll(`[data-lapisan="${c.kunci}"]`).forEach((n) => n.classList.toggle('ig-mati', !nyala));
          if (nyala) aktif.push(c.label.replace('Tampilkan ', ''));
        } else {
          const nilai = kontrol.querySelector(`[data-kunci="${c.kunci}"]:checked`).value;
          svg.querySelectorAll(`[data-${c.kunci}]`).forEach((n) => n.classList.toggle('ig-mati', n.dataset[c.kunci] !== nilai));
          keterangan.textContent = c.opsi.find((o) => o.nilai === nilai).keterangan;
        }
      }
      svg.setAttribute('aria-label', `${t.ilustrasi} ${keterangan.textContent} Lapisan yang ditampilkan: ${aktif.join(', ') || 'tidak ada'}.`);
    }

    const kontrol = el('div.ig-kontrol', null, t.kontrol.map((c) => (c.tipe === 'toggle'
      ? el('label.ig-pilih', null, el('input', { type: 'checkbox', checked: true, 'data-kunci': c.kunci, onchange: terapkan }), el('span', null, c.label))
      : el('fieldset.ig-grup', null, el('legend', null, c.label), c.opsi.map((o) => el('label.ig-pilih', null,
        el('input', { type: 'radio', name: `ig-${i}-${c.kunci}`, value: o.nilai, checked: o.nilai === c.awal, 'data-kunci': c.kunci, onchange: terapkan }),
        el('span', null, o.label)))))),
    t.nilai && el('dl.ig-nilai', null, t.nilai.map((n) => el('div', null, el('dt', null, n.label), el('dd', null, n.nilai)))));

    svg.setAttribute('role', 'img');
    const judul = el('h3', { tabindex: '-1' }, t.judul);
    const panel = el('div.ig-panel', null,
      el('p.tentang-label', null, `Tahap ${i + 1} dari ${D.tahap.length}`),
      judul,
      el('p', null, t.inti),
      el('div', null, el('h4', null, 'Apa yang ditunjukkan'), el('p.teks-sekunder', null, t.ditunjukkan)),
      el('div', null, el('h4', null, 'Peran AI dalam penelitian'), el('p.teks-sekunder', null, t.peran_ai)),
      t.hasil && el('div', null,
        el('h4', null, 'Hasil penelitian'),
        el('dl.ig-nilai.ig-nilai--hasil', null, t.hasil.map((h) => el('div', null, el('dt', null, h.label), el('dd', null, h.nilai)))),
        el('p.field__bantuan', null, t.hasil_catatan)),
      el('div.ig-panel__batasan', null, el('h4', null, 'Batasan demo'), el('p', null, t.batasan)),
      t.teknis && el('details.ig-lipat', null,
        el('summary', null, `Detail teknis: ${t.teknis.judul}`),
        el('p', null, t.teknis.pengantar),
        el('ul.ig-daftar', null, t.teknis.butir.map((b) => el('li', null, b)))));

    const akar = el('div.ig-tahap', null,
      el('figure.ig-gambar', null,
        el('div.ig-gambar__kepala', null, el('span.badge.badge--simulasi', null, 'Ilustrasi Konseptual'), el('span.ig-gambar__nonklinis', null, 'Nonklinis, bukan citra pasien')),
        svg,
        keterangan,
        legenda(t)),
      kontrol,
      panel);
    terapkan();
    return { akar, judul };
  }

  function bagianDemo() {
    const semua = D.tahap.map(tahap);
    let aktif = 0;
    const akhir = semua.length - 1;

    function pilih(n, fokus) {
      aktif = n;
      semua.forEach((s, i) => { s.akar.hidden = i !== n; });
      langkah.forEach((b, i) => { if (i === n) b.setAttribute('aria-current', 'step'); else b.removeAttribute('aria-current'); });
      sebelum.disabled = n === 0;
      berikut.hidden = n === akhir;
      keHasil.hidden = n !== akhir;
      if (!fokus) return;
      semua[n].judul.focus({ preventScroll: true });
      if (stepper.getBoundingClientRect().top < 0) demo.scrollIntoView({ block: 'start' });
    }

    const langkah = D.tahap.map((t, i) => el('button.ig-langkah', { type: 'button', onclick: () => pilih(i, true) },
      el('span.ig-langkah__nomor', { 'aria-hidden': 'true' }, String(i + 1)),
      el('span', null, el('span.sr-only', null, `Tahap ${i + 1}: `), t.singkat)));
    const stepper = el('ol.ig-stepper', null, langkah.map((b) => el('li', null, b)));
    const sebelum = tombol('Sebelumnya', 'kedua', { onclick: () => pilih(aktif - 1, true) });
    const berikut = tombol('Berikutnya', 'utama', { onclick: () => pilih(aktif + 1, true) });
    const keHasil = tautanTombol('Lihat Hasil Penelitian', 'utama', '#hasil', null, 'panah');

    const demo = el('section.bagian.bagian--muda.ig-demo', { id: 'demo', 'aria-labelledby': 'ig-j-demo' }, el('div.wadah', null,
      el('div.ig-kepala-demo', null,
        el('h2', { id: 'ig-j-demo' }, 'Jelajahi empat tahap penelitian.'),
        el('p.bagian__sub', null, D.catatan_demo)),
      stepper,
      semua.map((s) => s.akar),
      el('div.ig-nav', null, sebelum, berikut, keHasil)));
    pilih(0, false);
    return demo;
  }

  function bagian() {
    return [
      el('header.wadah.ig-kepala', null,
        el('a.tautan-panah.ig-kembali', { href: 'inference.html#portofolio' }, ikon('kiri', 18), 'Kembali ke Inference and Solutions'),
        el('h1', null, D.judul),
        el('p.ig-kepala__sub', null, D.deskripsi),
        el('ul.ig-label', null, D.label.map((l) => el('li', null, el('span.badge.badge--simulasi', null, l)))),
        el('div.banner', null, el('div.banner__teks', null, ikon('peringatan', 18), el('span', null, D.disclaimer)))),
      el('section.bagian.wadah', { 'aria-labelledby': 'ig-j-konteks' },
        el('h2', { id: 'ig-j-konteks' }, 'Mengapa perlu analisis yang teliti?'),
        el('ul.alur', null, D.konteks.map((k) => el('li', null, el('h3', null, k.judul), el('p.teks-sekunder', null, k.teks))))),
      bagianDemo(),
      el('section.bagian.wadah', { 'aria-labelledby': 'ig-j-cara' },
        el('h2', { id: 'ig-j-cara' }, 'Cara kerja AI dalam penelitian.'),
        el('div.ig-daftar-lipat', null, D.cara_kerja.map((c) => el('details.ig-lipat', null, el('summary', null, c.judul), el('p', null, c.teks))))),
      el('section.bagian.bagian--muda', { id: 'hasil', 'aria-labelledby': 'ig-j-hasil' }, el('div.wadah', null,
        el('h2', { id: 'ig-j-hasil' }, 'Hasil penelitian yang dilaporkan.'),
        el('p.bagian__sub', null, D.hasil_pengantar),
        el('ul.ig-metrik', null, D.hasil.map((h) => el('li', null,
          el('h3', null, h.judul),
          el('p.ig-metrik__nilai', null, h.nilai),
          el('p', null, h.metrik),
          el('p.field__bantuan', null, h.rincian)))),
        el('ul.ig-daftar', null, D.hasil_batasan.map((b) => el('li', null, b))),
        el('details.ig-lipat.ig-lipat--lebar', null,
          el('summary', null, 'Catatan statistik'),
          el('ul.ig-daftar', null, D.statistik.map((s) => el('li', null, s)))))),
      el('section.bagian.wadah', { 'aria-labelledby': 'ig-j-batasan' },
        el('h2', { id: 'ig-j-batasan' }, 'Batasan demo.'),
        el('ul.ig-daftar', null, D.batasan.map((b) => el('li', null, b)))),
    ];
  }

  return { bagian };
})();
