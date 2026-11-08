/* Halaman publik. Tiap HTML menentukan <body data-halaman>; isi dirender di sini. */
const K = MOCK.konten;
const utama = document.getElementById('utama');
const gambar = (nama) => `../assets/img/${nama}`;
const judulSitus = 'AI Center UBAYA';
const parameter = new URLSearchParams(location.search);

function setJudul(judul, deskripsi) {
  document.title = `${judul} | ${judulSitus}`;
  if (deskripsi) document.querySelector('meta[name="description"]')?.setAttribute('content', deskripsi);
}

/* === Kerangka === */
function navbar(aktif) {
  const MENU = [
    { kunci: 'index', label: 'Beranda', href: 'index.html' },
    { kunci: 'layanan', label: 'Layanan', sub: [{ label: 'Pelatihan', href: 'pelatihan.html' }, { label: 'Inference and Solutions', href: 'inference.html' }] },
    { kunci: 'berita', label: 'Berita', href: 'berita.html' },
    { kunci: 'dokumentasi', label: 'Dokumentasi', href: 'dokumentasi.html' },
    { kunci: 'tim', label: 'Tim', href: 'tim.html' },
    { kunci: 'tentang-kami', label: 'Tentang Kami', href: 'tentang-kami.html' },
  ];
  const menu = el('ul.nav__menu', { id: 'menu-utama' }, MENU.map((m) => {
    if (!m.sub) return el('li', null, el('a.nav__tautan', { href: m.href, 'aria-current': m.kunci === aktif ? 'page' : null }, m.label));
    const sub = el('ul.nav__sub', { hidden: true }, m.sub.map((s) => el('li', null, el('a', { href: s.href, 'aria-current': location.pathname.endsWith(s.href) ? 'page' : null }, s.label))));
    const btn = el('button.nav__tautan', { type: 'button', 'aria-expanded': 'false', 'data-aktif': m.kunci === aktif ? 'ya' : null,
      onclick: () => { const buka = sub.hidden; sub.hidden = !buka; btn.setAttribute('aria-expanded', String(buka)); } }, m.label, ikon('bawah', 16));
    const item = el('li.nav__punya-sub', null, btn, sub);
    document.addEventListener('click', (e) => { if (!item.contains(e.target)) { sub.hidden = true; btn.setAttribute('aria-expanded', 'false'); } });
    item.addEventListener('keydown', (e) => { if (e.key === 'Escape' && !sub.hidden) { e.stopPropagation(); sub.hidden = true; btn.setAttribute('aria-expanded', 'false'); btn.focus(); } });
    return item;
  }));
  const alihkanMenu = () => {
    const buka = menu.classList.toggle('terbuka');
    burger.setAttribute('aria-expanded', String(buka));
    burger.setAttribute('aria-label', buka ? 'Tutup menu' : 'Buka menu');
    burger.replaceChildren(ikon(buka ? 'tutup' : 'menu', 24));
  };
  const burger = el('button.nav__burger', { type: 'button', 'aria-label': 'Buka menu', 'aria-expanded': 'false', 'aria-controls': 'menu-utama', onclick: alihkanMenu }, ikon('menu', 24));
  return el('header.nav', { onkeydown: (e) => { if (e.key === 'Escape' && menu.classList.contains('terbuka')) { alihkanMenu(); burger.focus(); } } },
    el('div.wadah.nav__dalam', null,
      el('a.nav__logo', { href: 'index.html', 'aria-label': 'AI Center UBAYA, ke Beranda' }, el('img', { src: gambar('AI-Center_Logo.png'), alt: '', height: 40 })),
      burger,
      el('nav', { 'aria-label': 'Utama' }, menu)));
}

function footer() {
  const k = K.kontak;
  const tautan = (label, href) => el('li', null, el('a', { href }, label));
  return el('footer.footer', null,
    el('div.wadah.footer__grid', null,
      el('div.footer__profil', null,
        el('img', { src: gambar('AI-Center_Logo.png'), alt: 'AI Center UBAYA', height: 44, class: 'footer__logo' }),
        el('p', null, k.deskripsi_footer)),
      el('div', null, el('h2', null, 'Jelajahi'), el('ul', null,
        tautan('Pelatihan', 'pelatihan.html'), tautan('Inference and Solutions', 'inference.html'), tautan('Berita', 'berita.html'),
        tautan('Dokumentasi', 'dokumentasi.html'), tautan('Tim', 'tim.html'), tautan('Tentang Kami', 'tentang-kami.html'))),
      el('address', null, el('h2', null, 'Kontak'), el('ul', null,
        el('li', null, el('a.footer__kontak', { href: k.whatsapp_url, target: '_blank', rel: 'noopener' }, ikon('pesan', 16), el('span.sr-only', null, 'WhatsApp '), k.whatsapp)),
        el('li', null, el('a.footer__kontak', { href: k.email_url, target: '_blank', rel: 'noopener' }, ikon('surat', 16), k.email)),
        el('li', null, el('a.footer__kontak', { href: k.instagram_url, target: '_blank', rel: 'noopener' }, ikon('kamera', 16), el('span.sr-only', null, 'Instagram '), k.instagram)),
        k.alamat.map((b) => el('li', null, b)))),
      el('div.footer__peta', null,
        el('h2', null, 'Peta lokasi'),
        el('iframe', { src: k.peta_embed, title: 'Peta lokasi AI Center UBAYA', loading: 'lazy', referrerpolicy: 'no-referrer-when-downgrade' }),
        el('a', { href: k.peta_url, target: '_blank', rel: 'noopener' }, 'Buka di Google Maps'))),
    el('div.wadah.footer__bawah', null, el('small', null, `© 2026 AI Center Universitas Surabaya`)));
}

function breadcrumb(item) {
  const semua = [{ label: 'Beranda', href: 'index.html' }, ...item];
  return el('nav.remah', { 'aria-label': 'Breadcrumb' }, el('ol.wadah', null, semua.map((c, i) => el('li', null,
    i === semua.length - 1 ? el('span', { 'aria-current': 'page' }, c.label) : el('a', { href: c.href }, c.label)))));
}

function kepalaHalaman(judul, sub) {
  return el('div.kepala-halaman.wadah', null, el('h1', null, judul), sub && el('p', null, sub));
}

/* === Kartu === */
function kartuBerita(b) {
  const href = `berita-detail.html?slug=${encodeURIComponent(b.slug)}`;
  const tanggal = Format.parseTanggal(b.tanggal_publikasi);
  return el('article.kartu.kartu-berita', null,
    el('a.kartu-berita__tautan', { href },
      el('span.kartu-berita__media', null,
        el('img.kartu__gambar', { src: gambar(b.gambar), alt: '', loading: 'lazy' })),
      el('div.kartu__isi', null,
        el('time.tanggal', { datetime: tanggal?.toISOString().slice(0, 10) }, b.tanggal_publikasi),
        el('h2', null, b.judul),
        el('p.teks-sekunder', null, Format.potongan(b.isi)),
        el('span.tautan-panah', null, 'Lihat detail', ikon('panah', 18)))));
}

function kartuDokumentasi(d) {
  const href = `dokumentasi-detail.html?slug=${encodeURIComponent(d.slug)}`;
  return el('article.kartu', null,
    el('img.kartu__gambar', { src: gambar(d.gambar[0]?.key), alt: '', loading: 'lazy' }),
    el('div.kartu__isi', null,
      el('time.tanggal', null, d.tanggal_publikasi),
      el('h3', null, el('a.tautan-kartu', { href }, d.judul)),
      el('a.tautan-panah', { href, 'aria-label': `Lihat detail: ${d.judul}` }, 'Lihat detail', ikon('panah', 18))));
}

function kartuAlbum(d) {
  const href = `dokumentasi-detail.html?slug=${encodeURIComponent(d.slug)}`;
  const tanggal = Format.parseTanggal(d.tanggal_publikasi);
  return el('article.kartu.kartu-album', null,
    el('a.kartu-album__tautan', { href },
      el('span.kartu-album__media', null,
        el('img.kartu__gambar', { src: gambar(d.gambar[0]?.key), alt: '', loading: 'lazy' })),
      el('div.kartu__isi', null,
        el('div.kartu-album__meta', null,
          el('time.tanggal', { datetime: tanggal?.toISOString().slice(0, 10) }, d.tanggal_publikasi),
          el('span.kartu-album__jumlah', null, ikon('galeri', 16), `${d.gambar.length} Foto`)),
        el('h2', null, d.judul))));
}

function kartuTim(t) {
  return el('article.kartu-tim', null,
    el('img', { src: gambar(t.foto), alt: `Foto ${t.nama}`, loading: 'lazy', width: 520, height: 650 }),
    el('h2', null, t.nama),
    el('p.kartu-tim__jabatan', null, t.posisi),
    el('p.teks-sekunder', null, `Bidang penelitian: ${t.bidang_riset}`));
}

/** Hanya entri ber-href yang menjadi tautan; entri lain belum punya demo. */
function kartuPortofolio(p) {
  const isi = [
    el('img.kartu__gambar', { src: gambar(p.gambar), alt: '', loading: 'lazy' }),
    el('div.kartu__isi', null,
      el('p.kartu-tim__jabatan', null, p.bidang),
      el('h3', null, p.judul),
      el('p.teks-sekunder', null, p.deskripsi),
      p.status && el('p', null, el('span.badge.badge--simulasi', null, p.status)),
      p.cta && el('span.tautan-panah', null, p.cta, ikon('panah', 18))),
  ];
  return p.href
    ? el('article.kartu.kartu-album', null, el('a.kartu-album__tautan', { href: p.href }, isi))
    : el('article.kartu.porto__statis', null, isi);
}

/* === Bagian bersama === */
function bagianKontak(judul, sub, cta = {}) {
  const k = K.kontak;
  return el('section.kontak', { 'aria-labelledby': 'judul-kontak' }, el('div.wadah.kontak__dalam', null,
    el('div', null, el('h2', { id: 'judul-kontak' }, judul), el('p', null, sub)),
    el('div.kontak__aksi', null,
      tautanTombol(cta.wa ?? 'Hubungi lewat WhatsApp', 'terang', k.whatsapp_url, { target: '_blank', rel: 'noopener' }, 'pesan'),
      tautanTombol(cta.email ?? 'Kirim email', 'garis-terang', k.email_url, { target: '_blank', rel: 'noopener' }, 'surat'))));
}

/** Muat data ke wadah, lengkap dengan loading, galat, dan kosong. Dipanggil ulang saat panel berubah. */
function muatKe(wadah, { kerangka, ambil, gambarkan, kosong }) {
  wadah.dataset.token = String((Number(wadah.dataset.token) || 0) + 1);
  const token = wadah.dataset.token;
  wadah.replaceChildren(kerangka());
  const jalan = () => muatKe(wadah, { kerangka, ambil, gambarkan, kosong });
  ambil().then((hasil) => {
    if (wadah.dataset.token !== token) return;
    const daftar = hasil.items ?? hasil;
    if (Array.isArray(daftar) && daftar.length === 0) wadah.replaceChildren(kosong(hasil));
    else wadah.replaceChildren(...[gambarkan(hasil)].flat().filter(Boolean));
  }).catch((g) => {
    if (wadah.dataset.token !== token) return;
    wadah.replaceChildren(keadaanGalat(g, jalan));
  });
  // Hanya panggilan terakhir per wadah yang mendengar panel simulasi.
  wadah._lepasPanel?.();
  const dengar = () => jalan();
  document.addEventListener('simulasi:ubah', dengar);
  wadah._lepasPanel = () => document.removeEventListener('simulasi:ubah', dengar);
}

/* === Daftar dengan pencarian dan pagination === */
function halamanDaftar({ judul, judulTampil = judul, sub, ambil, kartu, namaJamak, kolomGrid = '' }) {
  setJudul(judul, sub);
  const wadah = el('div', { id: 'hasil' });
  const petunjuk = el('p.field__bantuan', { id: 'petunjuk-cari' });
  let halaman = 1;
  let q = '';
  let waktu = null;
  const input = el('input.input', { type: 'search', id: 'cari', placeholder: 'Cari judul', autocomplete: 'off', 'aria-describedby': 'petunjuk-cari',
    oninput: (e) => {
      clearTimeout(waktu);
      waktu = setTimeout(() => {
        q = e.target.value.trim();
        petunjuk.textContent = q.length === 1 ? 'Ketik minimal 2 karakter.' : '';
        halaman = 1;
        muat();
      }, 300);
    } });
  const qAktif = () => (q.length >= 2 ? q : '');
  function muat() {
    muatKe(wadah, {
      kerangka: () => kerangkaKartu(9),
      ambil: () => ambil({ page: halaman, q: qAktif() }),
      kosong: () => keadaan(qAktif()
        ? { judul: 'Tidak ada hasil', teks: `Tidak ada ${namaJamak} dengan judul "${qAktif()}". Coba kata lain.`, namaIkon: 'cari' }
        : { judul: `Belum ada ${namaJamak}`, teks: `${namaJamak[0].toUpperCase()}${namaJamak.slice(1)} akan tampil di sini setelah dipublikasikan.` }),
      gambarkan: (hasil) => [
        el(`div.grid-kartu${kolomGrid}`, null, hasil.items.map(kartu)),
        pagination(hasil, (n) => { halaman = n; muat(); wadah.scrollIntoView({ block: 'start' }); }),
      ],
    });
  }
  utama.append(
    breadcrumb([{ label: judul }]),
    kepalaHalaman(judulTampil, sub),
    el('div.wadah.daftar', null,
      el('div.field.cari-blok', null, el('label', { for: 'cari' }, `Cari ${namaJamak}`), el('div.cari', null, ikon('cari', 20), input), petunjuk),
      wadah));
  muat();
}

function tidakDitemukan(label = 'Halaman') {
  setJudul('Halaman tidak ditemukan');
  document.querySelector('meta[name="robots"]')?.setAttribute('content', 'noindex');
  utama.replaceChildren(
    breadcrumb([{ label: 'Halaman tidak ditemukan' }]),
    el('section.wadah.tak-ditemukan', null,
      el('p.tak-ditemukan__kode', { 'aria-hidden': 'true' }, '404'),
      el('h1', null, `${label} tidak ditemukan`),
      el('p.teks-sekunder', null, 'Tautan mungkin sudah berubah atau kontennya tidak lagi tersedia.'),
      tautanTombol('Kembali ke Beranda', 'utama', 'index.html')));
}

/* === Halaman === */
const HALAMAN = {
  index() {
    setJudul('Beranda', K.hero.sub);
    document.title = judulSitus;
    const wadahDok = el('div');
    const wadahTesti = el('div');
    const wadahBerita = el('div');
    utama.append(
      el('section.hero', { 'aria-labelledby': 'judul-hero' },
        el('canvas.hero__bola', { 'aria-hidden': 'true' }),
        el('div.wadah.hero__isi', null,
          el('div.hero__teks', null,
            el('h1', { id: 'judul-hero' }, K.hero.judul),
            el('p', null, K.hero.sub),
            el('div.hero__aksi', null,
              tautanTombol('Diskusikan Kebutuhan AI', 'terang', K.kontak.whatsapp_url, { target: '_blank', rel: 'noopener' }, 'pesan'),
              tautanTombol('Lihat Layanan', 'garis-terang', '#layanan'))))),
      el('section.lembar.lembar--putih', { id: 'layanan', 'aria-labelledby': 'judul-layanan' }, el('div.wadah', null,
        el('h2', { id: 'judul-layanan' }, 'Dua cara kami membantu Anda'),
        el('div.layanan-grid', null, K.layanan.map((l, i) => el('article', { class: `layanan-kartu layanan-kartu--${i ? 'gelap' : 'terang'}` },
          el('h3', null, l.nama), el('p', null, l.deskripsi),
          tautanTombol(i ? 'Lihat Solusi AI' : 'Lihat Pelatihan', i ? 'terang' : 'utama', l.href, null, 'panah')))))),
      el('section.lembar.lembar--muda', { 'aria-labelledby': 'judul-dok' }, el('div.wadah', null,
        el('div.judul-baris', null, el('h2', { id: 'judul-dok' }, 'Galeri Dokumentasi'), el('a.tautan-panah', { href: 'dokumentasi.html' }, 'Lihat Galeri Dokumentasi', ikon('panah', 18))),
        wadahDok)),
      el('section.lembar.lembar--langit', { 'aria-labelledby': 'judul-testi' }, el('div.wadah.testi', null,
        el('h2', { id: 'judul-testi' }, 'Cerita dari mereka yang belajar dan bekerja bersama kami'),
        wadahTesti)),
      el('section.lembar.lembar--putih', { 'aria-labelledby': 'judul-berita' }, el('div.wadah', null,
        el('div.judul-baris', null, el('h2', { id: 'judul-berita' }, 'Berita terbaru'), el('a.tautan-panah', { href: 'berita.html' }, 'Semua berita', ikon('panah', 18))),
        wadahBerita)),
      bagianKontak('Punya kebutuhan AI? Mari berdiskusi.', 'Ceritakan kebutuhan belajar atau penerapan AI Anda. Tim kami siap membahas langkah berikutnya.'));

    BolaNeural.pasang(utama.querySelector('.hero__bola'), { area: utama.querySelector('.hero') });

    muatKe(wadahDok, {
      kerangka: () => kerangkaKartu(3),
      ambil: () => api.daftarDokumentasi({ limit: 3 }),
      gambarkan: (h) => el('div.grid-kartu.grid-kartu--tiga', null, h.items.map(kartuDokumentasi)),
      kosong: () => keadaan({ judul: 'Belum ada dokumentasi', teks: 'Dokumentasi kegiatan akan tampil di sini.' }),
    });
    muatKe(wadahTesti, {
      kerangka: () => el('div.kerangka', { style: 'height:16rem', role: 'status', 'aria-label': 'Memuat data' }),
      ambil: () => api.daftarTestimoni(),
      gambarkan: (h) => carousel(h.items.map((t) => el('figure.testi__kutip', null,
        el('blockquote', null, t.teks),
        el('figcaption', null,
          t.foto ? el('img.testi__foto', { src: gambar(t.foto), alt: '' }) : el('span.testi__inisial', { 'aria-hidden': 'true' }, Format.inisial(t.nama)),
          el('span', null, el('strong', null, t.nama), el('br'), t.posisi)))), { label: 'Testimoni klien', intervalMs: 10000 }),
      kosong: () => keadaan({ judul: 'Belum ada testimoni', teks: 'Testimoni akan tampil di sini.' }),
    });
    muatKe(wadahBerita, {
      kerangka: () => kerangkaKartu(3),
      ambil: () => api.daftarBerita({ limit: 3 }),
      gambarkan: (h) => el('div.grid-kartu.grid-kartu--tiga', null, h.items.map(kartuBerita)),
      kosong: () => keadaan({ judul: 'Belum ada berita', teks: 'Berita terbaru akan tampil di sini.' }),
    });
    aktifkanLembar();
  },

  pelatihan() {
    const p = K.pelatihan;
    setJudul('Pelatihan', p.hero_sub);
    utama.append(
      breadcrumb([{ label: 'Pelatihan' }]),
      el('section.sampul.sampul--belajar', null, el('div.wadah', null,
        el('h1', null, p.hero_judul), el('p', null, p.hero_sub),
        el('div.hero__aksi', null,
          tautanTombol('Pelatihan Reguler', 'utama', '#reguler'),
          tautanTombol('Konsultasi Gratis', 'kedua', '#khusus')))),
      el('section.bagian.wadah', { id: 'reguler', 'aria-labelledby': 'j-reguler' },
        el('h2', { id: 'j-reguler' }, 'Pelatihan Reguler'),
        el('div.segera', null, el('strong', null, 'Segera hadir'), el('p.teks-sekunder', null, p.reguler))),
      el('section.bagian.bagian--muda', { id: 'khusus', 'aria-labelledby': 'j-khusus' }, el('div.wadah', null,
        el('h2', { id: 'j-khusus' }, 'Pelatihan Khusus'),
        el('p.bagian__sub', null, p.khusus_sub),
        el('ol.alur', null, p.alur_konsultasi.map((l, i) => el('li', null, el('span.alur__nomor', { 'aria-hidden': 'true' }, String(i + 1)), el('h3', null, l.judul), el('p.teks-sekunder', null, l.deskripsi)))),
        tautanTombol('Tanya Pelatihan Khusus', 'utama', K.kontak.whatsapp_url, { target: '_blank', rel: 'noopener' }, 'pesan'))),
      el('section.bagian.wadah', { id: 'workshop', 'aria-labelledby': 'j-workshop' },
        el('h2', { id: 'j-workshop' }, 'Workshop atau Seminar untuk Bisnis'),
        el('p.bagian__sub', null, p.workshop_sub),
        tautanTombol('Tanya Workshop Bisnis', 'utama', K.kontak.whatsapp_url, { target: '_blank', rel: 'noopener' }, 'pesan')));
  },

  inference() {
    const n = K.inference;
    setJudul('Inference and Solutions', n.hero_sub);
    utama.append(
      breadcrumb([{ label: 'Inference and Solutions' }]),
      el('section.sampul.sampul--enterprise', null, el('div.wadah.sampul__dua', null,
        el('div', null, el('h1', null, n.hero_judul), el('p', null, n.hero_sub)),
        carousel(n.pertanyaan.map((t) => el('p.pertanyaan', null, t)), { label: 'Pertanyaan untuk Anda', intervalMs: 6000 }))),
      el('section.bagian.wadah', { 'aria-labelledby': 'j-jenis' },
        el('h2', { id: 'j-jenis' }, 'Dua area untuk memulai diskusi.'),
        el('div.jenis-grid', null, n.jenis.map((j) => el('article.jenis', null, el('h3', null, j.nama), el('p.teks-sekunder', null, j.deskripsi))))),
      el('section.bagian.bagian--muda', { 'aria-labelledby': 'j-alur' }, el('div.wadah', null,
        el('h2', { id: 'j-alur' }, 'Tahapan penerapan model AI.'),
        el('ol.alur.alur--lima', null, n.alur.map((l, i) => el('li', null, el('span.alur__nomor', { 'aria-hidden': 'true' }, String(i + 1)), el('h3', null, l.judul), el('p.teks-sekunder', null, l.deskripsi)))))),
      el('section.bagian.wadah', { id: 'portofolio', 'aria-labelledby': 'j-porto' },
        el('div.porto__kepala', null,
          el('p.tentang-label', null, 'Portofolio'),
          el('h2', { id: 'j-porto' }, 'Jelajahi Portofolio Kami.'),
          el('p.bagian__sub', null, 'Kenali contoh penerapan teknologi AI melalui karya dan demonstrasi aplikasi.')),
        el('div.grid-kartu.grid-kartu--tiga.porto', null, n.portofolio.map(kartuPortofolio))),
      bagianKontak('Diskusikan kebutuhan AI Anda.', 'Ceritakan model, data, atau tantangan yang sedang Anda hadapi. Tim AI Center UBAYA dapat membantu membahas langkah berikutnya.', { wa: 'Diskusikan Kebutuhan AI' }));
  },

  'jejak-alumni'() {
    setJudul('JejakAlumni (Prototipe)', K.jejak_alumni.disclaimer);
    utama.append(
      breadcrumb([{ label: 'Inference and Solutions', href: 'inference.html' }, { label: 'JejakAlumni' }]),
      JejakAlumni.halaman(),
      bagianKontak('Ingin membahas data lulusan di institusi Anda?', 'Ceritakan kebutuhan evaluasi lulusan Anda. Tim AI Center UBAYA dapat membantu membahas langkah berikutnya.', { wa: 'Diskusikan Kebutuhan AI' }));
  },

  berita() {
    halamanDaftar({ judul: 'Berita', judulTampil: 'Kabar Terbaru dari AI Center UBAYA.', sub: 'Ikuti kegiatan, kolaborasi, dan perkembangan AI Center UBAYA melalui cerita dan informasi terbaru kami.', ambil: api.daftarBerita, kartu: kartuBerita, namaJamak: 'berita', kolomGrid: '.grid-kartu--berita' });
  },

  dokumentasi() {
    halamanDaftar({ judul: 'Galeri Dokumentasi', sub: 'Melihat lebih dekat kegiatan, kolaborasi, dan momen bersama AI Center UBAYA.', ambil: api.daftarDokumentasi, kartu: kartuAlbum, namaJamak: 'album', kolomGrid: '.grid-kartu--album' });
  },

  'berita-detail'() {
    const slug = parameter.get('slug');
    const wadah = el('div');
    utama.append(wadah);
    muatDetail(wadah, () => api.detailBerita(slug), 'Berita', (b) => {
      const isi = Format.paragraf(b.isi);
      setJudul(b.judul, Format.potongan(b.isi, 160));
      return [
        breadcrumb([{ label: 'Berita', href: 'berita.html' }, { label: b.judul }]),
        el('article.artikel', null,
          el('header.artikel__kepala', null,
            el('h1', null, b.judul),
            el('p.artikel__meta', null, el('time.tanggal', null, b.tanggal_publikasi), el('span', null, 'Tim Media Ubaya AI Center'))),
          el('img.artikel__gambar', { src: gambar(b.gambar), alt: '' }),
          el('div.artikel__isi', null, isi.map((p) => el('p', null, p))),
          el('a.tautan-panah.artikel__kembali', { href: 'berita.html' }, 'Semua berita', ikon('panah', 18))),
      ];
    });
  },

  'dokumentasi-detail'() {
    const slug = parameter.get('slug');
    const wadah = el('div');
    utama.append(wadah);
    muatDetail(wadah, () => api.detailDokumentasi(slug), 'Dokumentasi', (d) => {
      setJudul(d.judul, `Dokumentasi kegiatan ${d.judul} — ${d.tanggal_publikasi}`);
      return [
        breadcrumb([{ label: 'Dokumentasi', href: 'dokumentasi.html' }, { label: d.judul }]),
        el('article.wadah.dok', null,
          el('header.artikel__kepala', null, el('h1', null, d.judul), el('p.artikel__meta', null, el('time.tanggal', null, d.tanggal_publikasi))),
          el(`div.galeri.galeri--${d.gambar.length}`, null, d.gambar.map((g, i) => el('img', { src: gambar(g.key), alt: `Foto ${i + 1} dari ${d.gambar.length}: ${d.judul}`, loading: i ? 'lazy' : null }))),
          el('a.tautan-panah.artikel__kembali', { href: 'dokumentasi.html' }, 'Semua dokumentasi', ikon('panah', 18))),
      ];
    });
  },

  tim() {
    const judul = 'Tim';
    const sub = 'Kenali tim yang berkontribusi dalam riset, pengembangan, dan pembelajaran AI di AI Center UBAYA.';
    setJudul(judul, sub);
    const wadah = el('div');
    const petunjuk = el('p.field__bantuan', { id: 'petunjuk-cari', role: 'status' });
    let semua = [];
    let q = '';
    function saring() {
      const hasil = q ? semua.filter((t) => `${t.nama} ${t.posisi} ${t.bidang_riset}`.toLowerCase().includes(q.toLowerCase())) : semua;
      petunjuk.textContent = q ? `${hasil.length} anggota ditemukan.` : `${hasil.length} anggota tim.`;
      wadah.replaceChildren(hasil.length
        ? el('div.grid-tim', null, hasil.map(kartuTim))
        : keadaan({ judul: 'Tidak ada hasil', teks: `Tidak ada anggota tim yang cocok dengan "${q}".`, namaIkon: 'cari' }));
    }
    function muat() {
      muatKe(wadah, {
        kerangka: () => kerangkaKartu(6),
        ambil: () => api.daftarTim(),
        kosong: () => keadaan({ judul: 'Belum ada data tim', teks: 'Susunan tim akan tampil di sini.' }),
        gambarkan: (h) => { semua = h.items; const c = el('div'); queueMicrotask(saring); return c; },
      });
    }
    utama.append(breadcrumb([{ label: judul }]),
      el('header.kepala-halaman.wadah', null, el('h1', null, 'Orang-Orang di Balik Inovasi AI.'), el('p', null, sub)),
      el('div.wadah.daftar.daftar--tim', null,
        el('div.field.cari-blok', null, el('label', { for: 'cari' }, 'Cari anggota tim'),
          el('div.cari', null, ikon('cari', 20), el('input.input', { type: 'search', id: 'cari', placeholder: 'Cari nama, jabatan, atau bidang', 'aria-describedby': 'petunjuk-cari',
            oninput: (e) => { q = e.target.value.trim().replace(/\s+/g, ' '); if (semua.length) saring(); } })), petunjuk),
        wadah),
      el('section.kontak.kontak--tim', { 'aria-labelledby': 'judul-kontak-tim' }, el('div.wadah.kontak__dalam', null,
        el('div', null,
          el('h2', { id: 'judul-kontak-tim' }, 'Mari Diskusikan Kebutuhan Anda.'),
          el('p', null, 'Punya pertanyaan, ide kolaborasi, atau ingin mengetahui bagaimana AI dapat membantu? Hubungi AI Center UBAYA untuk memulai percakapan.')),
        tautanTombol('Hubungi via WhatsApp', 'terang', K.kontak.whatsapp_url, { target: '_blank', rel: 'noopener' }, 'pesan'))));
    muat();
  },

  'tentang-kami'() {
    const t = K.tentang;
    setJudul('Tentang Kami', t.deskripsi);
    utama.append(
      breadcrumb([{ label: 'Tentang Kami' }]),
      el('section.sampul.tentang-sampul', { 'aria-labelledby': 'judul-tentang' }, el('div.wadah', null,
        el('p.tentang-label', null, 'Tentang Kami'),
        el('h1', { id: 'judul-tentang' }, t.hero_judul),
        el('p.tentang-sampul__sub', null, t.deskripsi))),
      el('section.bagian.wadah.tentang-cerita', { 'aria-labelledby': 'judul-cerita' },
        el('div.tentang-pengantar.tentang-ungkap', null,
          el('h2', { id: 'judul-cerita' }, 'Arah yang Menjadi Landasan Langkah Kami.'),
          el('p.teks-sekunder', null, t.visi_pengantar)),
        el('div.visi-misi.tentang-ungkap', null,
          el('div', null, el('h3', null, 'Visi'), el('p.visi-misi__judul', null, t.visi_judul), el('p.teks-sekunder', null, t.visi)),
          el('div', null, el('h3', null, 'Misi'), el('ol.misi', null, t.misi.map((m) => el('li', null, m)))))),
      el('section.bagian.bagian--muda.tentang-ruang', { 'aria-labelledby': 'judul-ruang' }, el('div.wadah', null,
        el('div.tentang-pengantar.tentang-ungkap', null,
          el('h2', { id: 'judul-ruang' }, 'Ruang untuk Belajar, Berkolaborasi, dan Berinovasi.'),
          el('p.teks-sekunder', null, t.fasilitas_pengantar)),
        el('div.fasilitas.tentang-ungkap', null, t.fasilitas.map((f) => el('figure', null,
          f.gambar ? el('img', { src: gambar(f.gambar), alt: f.nama, loading: 'lazy', width: 1000, height: 625 })
            : el('div.fasilitas__kosong', { role: 'img', 'aria-label': 'Foto ruang demo belum tersedia' }, 'Foto Ruangan Coming Soon'),
          el('figcaption', null, f.nama)))))),
      el('section.kontak.kontak--tentang', { 'aria-labelledby': 'judul-kontak' }, el('div.wadah.kontak__dalam', null,
        el('div.tentang-ungkap', null,
          el('h2', { id: 'judul-kontak' }, 'Langkah Berikutnya Dimulai di Sini.'),
          el('p', null, 'Ingin mengembangkan keterampilan AI atau menerapkannya untuk kebutuhan nyata? Jelajahi layanan kami dan temukan langkah yang sesuai untuk Anda.')),
        el('div.tentang-layanan', null,
          el('a.tentang-layanan__tautan', { href: 'pelatihan.html' },
            el('span.tentang-layanan__isi', null, el('strong', null, 'Pelatihan'), el('span', null, 'Kembangkan pengetahuan dan keterampilan AI.')),
            ikon('panah', 20)),
          el('a.tentang-layanan__tautan', { href: 'inference.html' },
            el('span.tentang-layanan__isi', null, el('strong', null, 'Inference and Solutions'), el('span', null, 'Konsultasikan kebutuhan AI dan jelajahi solusi penerapannya.')),
            ikon('panah', 20))))));
    if ('IntersectionObserver' in window && !matchMedia('(prefers-reduced-motion: reduce)').matches) {
      const pengamat = new IntersectionObserver((entri) => entri.forEach((e) => {
        if (e.isIntersecting) { e.target.classList.add('tentang--masuk'); pengamat.unobserve(e.target); }
      }), { rootMargin: '0px 0px -8% 0px' });
      utama.querySelectorAll('.tentang-ungkap').forEach((bagian) => pengamat.observe(bagian));
    }
  },

  404() { tidakDitemukan(); },
};

/** Detail: 404 sisi client untuk slug tidak dikenal; keadaan loading dan galat tersedia. */
function muatDetail(wadah, ambil, label, gambarkan) {
  const jalan = () => {
    wadah.replaceChildren(el('div.wadah.detail-kerangka', { role: 'status', 'aria-label': 'Memuat data' }, el('div.kerangka', { style: 'height:3rem;width:70%' }), el('div.kerangka', { style: 'height:24rem' })));
    ambil().then((d) => wadah.replaceChildren(...gambarkan(d))).catch((g) => {
      if (g.status === 404) tidakDitemukan(label);
      else wadah.replaceChildren(el('div.wadah.daftar', null, keadaanGalat(g, jalan)));
    });
  };
  // Panel simulasi: muat ulang detail, juga setelah layar 404 menggantikan wadah.
  document.addEventListener('simulasi:ubah', () => { utama.replaceChildren(wadah); jalan(); });
  jalan();
}

/* Pergantian section Beranda: lembar berikutnya naik menutupi yang sebelumnya. */
function aktifkanLembar() {
  if (matchMedia('(prefers-reduced-motion: reduce)').matches || !('IntersectionObserver' in window)) return;
  const pengamat = new IntersectionObserver((entri) => entri.forEach((e) => {
    if (e.isIntersecting) { e.target.classList.add('lembar--masuk'); pengamat.unobserve(e.target); }
  }), { rootMargin: '0px 0px -8% 0px' });
  document.querySelectorAll('.lembar').forEach((l) => {
    if (l.getBoundingClientRect().top > innerHeight) { l.classList.add('lembar--siap'); pengamat.observe(l); }
  });
}

/* === Mulai === */
const namaHalaman = document.body.dataset.halaman;
const menuAktif = ['pelatihan', 'inference', 'jejak-alumni'].includes(namaHalaman) ? 'layanan' : namaHalaman.replace('-detail', '');
document.body.prepend(el('a.lewati', { href: '#utama' }, 'Lewati ke konten'), navbar(menuAktif));
document.body.append(footer());
HALAMAN[namaHalaman]();
