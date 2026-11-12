/* Halaman publik. Tiap HTML menentukan <body data-halaman>; isi dirender di sini. */
const K = KONTEN;
const utama = document.getElementById('utama');
const gambar = (nama) => (/^(\/|https?:)/.test(nama) ? nama : `/assets/img/${nama}`); // URL dari API (/media/...) dipakai apa adanya
const judulSitus = 'UBAYA AI Center';
const parameter = new URLSearchParams(location.search);
const slugDariUrl = () => { try { return decodeURIComponent(location.pathname.split('/').filter(Boolean).pop() ?? ''); } catch { return ''; } };

function setJudul(judul, deskripsi) {
  document.title = `${judul} | ${judulSitus}`;
  if (deskripsi) document.querySelector('meta[name="description"]')?.setAttribute('content', deskripsi);
}

/* === Kerangka === */
function navbar(aktif) {
  const MENU = [
    { kunci: 'index', label: 'Beranda', href: '/' },
    { kunci: 'layanan', label: 'Layanan', sub: [{ label: 'Pelatihan', href: '/layanan/pelatihan' }, { label: 'Inference and Solutions', href: '/layanan/inference-solutions' }] },
    { kunci: 'contoh-aplikasi', label: 'Contoh Aplikasi', href: '/layanan/inference-solutions#portofolio' },
    { kunci: 'kabar', label: 'Kabar', sub: [{ label: 'Berita', href: '/berita' }, { label: 'Galeri Dokumentasi', href: '/dokumentasi' }] },
    { kunci: 'tim', label: 'Tim', href: '/tim' },
    { kunci: 'tentang-kami', label: 'Tentang Kami', href: '/tentang-kami' },
  ];
  const menu = el('ul.nav__menu', { id: 'menu-utama' }, MENU.map((m) => {
    if (!m.sub) return el('li', null, el('a.nav__tautan', { href: m.href, 'data-nav': m.kunci }, m.label));
    const idSub = `menu-${m.kunci}`;
    const sub = el('ul.nav__sub', { id: idSub, hidden: true }, m.sub.map((s) => el('li', null, el('a', { href: s.href, 'aria-current': location.pathname.replace(/\/$/, '') === s.href ? 'page' : null }, s.label))));
    const btn = el('button.nav__tautan', { type: 'button', 'aria-expanded': 'false', 'aria-controls': idSub, 'data-nav': m.kunci,
      onclick: () => { const buka = sub.hidden; sub.hidden = !buka; btn.setAttribute('aria-expanded', String(buka)); } }, m.label, ikon('bawah', 16));
    const item = el('li.nav__punya-sub', null, btn, sub);
    document.addEventListener('click', (e) => { if (!item.contains(e.target)) { sub.hidden = true; btn.setAttribute('aria-expanded', 'false'); } });
    item.addEventListener('keydown', (e) => { if (e.key === 'Escape' && !sub.hidden) { e.stopPropagation(); sub.hidden = true; btn.setAttribute('aria-expanded', 'false'); btn.focus(); } });
    item.addEventListener('focusout', (e) => { if (!item.contains(e.relatedTarget)) { sub.hidden = true; btn.setAttribute('aria-expanded', 'false'); } });
    return item;
  }));
  const alihkanMenu = () => {
    const buka = menu.classList.toggle('terbuka');
    akar.classList.toggle('nav--menu-terbuka', buka);
    burger.setAttribute('aria-expanded', String(buka));
    burger.setAttribute('aria-label', buka ? 'Tutup menu' : 'Buka menu');
    burger.replaceChildren(ikon(buka ? 'tutup' : 'menu', 24));
  };
  const burger = el('button.nav__burger', { type: 'button', 'aria-label': 'Buka menu', 'aria-expanded': 'false', 'aria-controls': 'menu-utama', onclick: alihkanMenu }, ikon('menu', 24));
  const cta = el('a.nav__cta', { href: K.kontak.whatsapp_url, target: '_blank', rel: 'noopener', 'aria-label': 'Diskusikan Kebutuhan AI melalui WhatsApp (tab baru)' }, 'Diskusikan Kebutuhan AI');
  const akar = el('header.nav', { onkeydown: (e) => { if (e.key === 'Escape' && menu.classList.contains('terbuka')) { tutupMenu(); burger.focus(); } } },
    el('div.wadah.nav__dalam', null,
      el('span.nav__permukaan', { 'aria-hidden': 'true' }),
      el('a.nav__logo', { href: '/', 'aria-label': 'UBAYA AI Center, ke Beranda' },
        el('span.nav__logo-penuh', null, el('img', { src: gambar('AI-Center_Logo.png'), alt: '', width: 440, height: 116 })),
        el('span.nav__logo-simbol', null, el('img', { src: gambar('AI-Center_Logo.png'), alt: '', width: 440, height: 116 }))),
      burger,
      el('nav.nav__navigasi', { 'aria-label': 'Utama' }, menu), cta));
  const desktop = matchMedia('(min-width: 80rem)');
  const minimGerak = matchMedia('(prefers-reduced-motion: reduce)');
  const wadahNav = akar.querySelector('.nav__dalam');
  const permukaan = akar.querySelector('.nav__permukaan');
  const bagianBergerak = [akar.querySelector('.nav__logo'), ...menu.children, cta];
  let animasi = [];
  function hentikanGerak() {
    animasi.forEach((a) => a.cancel());
    animasi = [];
  }
  function ubahKeadaan(jadiRingkas, langsung = false) {
    if (akar.classList.contains('nav--ringkas') === jadiRingkas) return;
    const bergerak = !langsung && !minimGerak.matches && !akar.classList.contains('nav--mobile') && typeof permukaan.animate === 'function';
    const elemen = [...bagianBergerak, permukaan];
    // FLIP membaca posisi yang sedang terlihat agar pembalikan scroll tidak meloncat ke awal.
    const sebelum = bergerak ? elemen.map((n) => n.getBoundingClientRect()) : [];
    const opasitas = bergerak ? getComputedStyle(permukaan).opacity : 0;
    hentikanGerak();
    akar.classList.toggle('nav--ringkas', jadiRingkas);
    if (!bergerak) return;
    const sesudah = elemen.map((n) => n.getBoundingClientRect());
    elemen.forEach((n, i) => {
      const awal = sebelum[i], akhir = sesudah[i];
      if (!akhir.width || !akhir.height) return;
      const permukaanSaja = n === permukaan;
      const pindah = `translate(${awal.left - akhir.left}px, ${awal.top - akhir.top}px)`;
      const dari = { transform: permukaanSaja ? `${pindah} scale(${awal.width / akhir.width}, ${awal.height / akhir.height})` : pindah };
      const ke = { transform: 'none' };
      if (permukaanSaja) { dari.opacity = opasitas; ke.opacity = jadiRingkas ? 1 : 0; }
      animasi.push(n.animate([dari, ke], { duration: 360, easing: 'cubic-bezier(0.77, 0, 0.175, 1)' }));
    });
  }
  akar.classList.toggle('nav--mobile', !desktop.matches);
  let ringkas = scrollY > 96;
  let frameScroll = null;
  let frameUkuran = null;
  function tutupMenu() {
    menu.classList.remove('terbuka');
    akar.classList.remove('nav--menu-terbuka');
    burger.setAttribute('aria-expanded', 'false');
    burger.setAttribute('aria-label', 'Buka menu');
    burger.replaceChildren(ikon('menu', 24));
    akar.querySelectorAll('.nav__sub').forEach((s) => { s.hidden = true; s.previousElementSibling.setAttribute('aria-expanded', 'false'); });
  }
  function aturAktif(dariScroll = false) {
    const halaman = document.body.dataset.halaman;
    let induk = ['berita', 'dokumentasi'].includes(aktif) ? 'kabar' : aktif;
    if (halaman === 'inference' && location.hash === '#portofolio') induk = 'contoh-aplikasi';
    if (dariScroll && ['index', 'inference'].includes(halaman)) {
      const batas = parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--nav-offset')) || 104;
      const sections = [...utama.querySelectorAll('[data-nav-section]')].map((section) => ({ section, posisi: section.getBoundingClientRect() }));
      // Section tanpa pasangan menu mempertahankan pasangan terakhir yang telah dilewati.
      if (sections.some(({ posisi }) => posisi.height)) {
        induk = aktif;
        for (const { section, posisi } of sections) {
          if (posisi.height && posisi.top <= batas + 1) induk = section.dataset.navSection;
        }
        if (scrollY <= 1) induk = aktif;
      }
    }
    akar.querySelectorAll('[data-nav]').forEach((n) => {
      const kunci = n.dataset.nav;
      const dipilih = kunci === induk;
      n.removeAttribute('aria-current');
      n.removeAttribute('data-aktif');
      if (dipilih && n.tagName === 'A') n.setAttribute('aria-current', induk === 'contoh-aplikasi' || induk !== aktif ? 'location' : 'page');
      else if (dipilih) n.setAttribute('data-aktif', 'ya');
    });
  }
  function aturScroll(paksa = false) {
    aturAktif(true);
    if (!paksa && scrollY >= 48 && (akar.contains(document.activeElement) || akar.querySelector('.nav__sub:not([hidden])'))) return;
    if (scrollY > 96) ringkas = true;
    else if (scrollY < 48) ringkas = false;
    ubahKeadaan(ringkas && !akar.classList.contains('nav--mobile'), paksa);
  }
  function aturUkuran() {
    hentikanGerak();
    akar.classList.add('nav--ukur');
    const sebelumnyaRingkas = akar.classList.contains('nav--ringkas');
    akar.classList.remove('nav--ringkas');
    akar.classList.toggle('nav--mobile', !desktop.matches);
    if (desktop.matches) {
      const viewport = document.documentElement.clientWidth || innerWidth;
      const ruang = wadahNav.getBoundingClientRect().width || viewport - 2 * Math.min(64, Math.max(20, viewport * .03));
      const lebarMenu = menu.scrollWidth;
      const lebarCta = cta.getBoundingClientRect().width;
      // Dua sisi sama lebar menjaga menu tepat di tengah, bukan di tengah ruang sisa.
      if (lebarMenu && lebarMenu + 2 * Math.max(152, lebarCta) + 40 > ruang) akar.classList.add('nav--mobile');
      else {
        akar.classList.add('nav--ringkas');
        const lebarRingkas = menu.scrollWidth + 70 + lebarCta + 60;
        akar.style.setProperty('--nav-lebar-ringkas', `${Math.ceil(lebarRingkas)}px`);
      }
    }
    akar.classList.toggle('nav--ringkas', sebelumnyaRingkas && !akar.classList.contains('nav--mobile'));
    if (!akar.classList.contains('nav--mobile')) tutupMenu();
    document.documentElement.style.setProperty('--nav-offset', akar.classList.contains('nav--mobile') ? '92px' : '104px');
    aturScroll(true);
    // Selesaikan pengukuran tanpa memutar ulang crossfade logo saat resize/font siap.
    getComputedStyle(akar.querySelector('.nav__logo-penuh')).opacity;
    akar.classList.remove('nav--ukur');
  }
  akar.addEventListener('focusin', hentikanGerak);
  akar.addEventListener('focusout', () => requestAnimationFrame(() => aturScroll()));
  minimGerak.addEventListener('change', hentikanGerak);
  akar.addEventListener('click', (e) => {
    const a = e.target.closest('a[href]');
    if (!a || e.defaultPrevented || e.button !== 0 || e.ctrlKey || e.metaKey || e.shiftKey || e.altKey) return;
    tutupMenu();
    const tujuan = new URL(a.href, location.href);
    if (tujuan.pathname === location.pathname && tujuan.search === location.search && tujuan.hash === '#portofolio') {
      requestAnimationFrame(() => {
        const target = document.getElementById('portofolio');
        if (target) {
          target.setAttribute('tabindex', '-1');
          if (e.detail > 0) target.dataset.fokusPointer = 'ya';
          else delete target.dataset.fokusPointer;
          target.addEventListener('blur', () => delete target.dataset.fokusPointer, { once: true });
          target.addEventListener('keydown', () => delete target.dataset.fokusPointer, { once: true });
          target.focus({ preventScroll: true });
        }
        aturScroll(true);
      });
    }
  });
  // Jalur event tetap mengenali ikon hamburger yang diganti ketika tombol diklik.
  document.addEventListener('click', (e) => { if (!e.composedPath().includes(akar)) { tutupMenu(); aturScroll(); } });
  window.addEventListener('hashchange', () => { aturAktif(); aturUkuran(); });
  window.addEventListener('scroll', () => {
    if (frameScroll !== null) return;
    frameScroll = requestAnimationFrame(() => { frameScroll = null; aturScroll(); });
  }, { passive: true });
  window.addEventListener('resize', () => {
    if (frameUkuran !== null) return;
    frameUkuran = requestAnimationFrame(() => { frameUkuran = null; aturUkuran(); });
  }, { passive: true });
  aturAktif();
  requestAnimationFrame(aturUkuran);
  document.fonts?.ready.then(aturUkuran);
  if ('ResizeObserver' in window) new ResizeObserver(() => {
    if (frameScroll !== null) return;
    frameScroll = requestAnimationFrame(() => { frameScroll = null; aturAktif(true); });
  }).observe(utama);
  return akar;
}

function footer() {
  const k = K.kontak;
  const tautan = (label, href) => el('li', null, el('a', { href }, label));
  return el('footer.footer', null,
    el('div.wadah.footer__grid', null,
      el('div.footer__profil', null,
        el('img', { src: gambar('AI-Center_Logo.png'), alt: 'UBAYA AI Center', width: 440, height: 116, class: 'footer__logo' }),
        el('p', null, k.deskripsi_footer)),
      el('div', null, el('h2', null, 'Jelajahi'), el('ul', null,
        tautan('Pelatihan', '/layanan/pelatihan'), tautan('Inference and Solutions', '/layanan/inference-solutions'), tautan('Berita', '/berita'),
        tautan('Galeri Dokumentasi', '/dokumentasi'), tautan('Tim', '/tim'), tautan('Tentang Kami', '/tentang-kami'))),
      el('address', null, el('h2', null, 'Kontak'), el('ul', null,
        el('li', null, el('a.footer__kontak', { href: k.whatsapp_url, target: '_blank', rel: 'noopener' }, ikon('pesan', 16), el('span.sr-only', null, 'WhatsApp '), k.whatsapp)),
        el('li', null, el('a.footer__kontak', { href: k.email_url, target: '_blank', rel: 'noopener' }, ikon('surat', 16), k.email)),
        el('li', null, el('a.footer__kontak', { href: k.instagram_url, target: '_blank', rel: 'noopener' }, ikon('kamera', 16), el('span.sr-only', null, 'Instagram '), k.instagram)),
        k.alamat.map((b) => el('li', null, b)))),
      el('div.footer__peta', null,
        el('h2', null, 'Peta lokasi'),
        el('iframe', { src: k.peta_embed, title: 'Peta lokasi UBAYA AI Center', loading: 'lazy', referrerpolicy: 'no-referrer-when-downgrade' }),
        el('a', { href: k.peta_url, target: '_blank', rel: 'noopener' }, 'Buka di Google Maps'))),
    el('div.wadah.footer__bawah', null, el('small', null, `© 2026 UBAYA AI Center`)));
}

function breadcrumb(item) {
  const semua = [{ label: 'Beranda', href: '/' }, ...item];
  return el('nav.remah', { 'aria-label': 'Breadcrumb' }, el('ol.wadah', null, semua.map((c, i) => el('li', null,
    i === semua.length - 1 ? el('span', { 'aria-current': 'page' }, c.label) : el('a', { href: c.href }, c.label)))));
}

function kepalaHalaman(judul, sub) {
  return el('div.kepala-halaman.wadah', null, el('h1', null, judul), sub && el('p', null, sub));
}

/* === Kartu === */
function kartuBerita(b, tingkat = 'h2') {
  const href = `/berita/${encodeURIComponent(b.slug)}`;
  const tanggal = Format.parseTanggal(b.tanggal_publikasi);
  return el('article.kartu.kartu-berita', null,
    el('a.kartu-berita__tautan', { href },
      el('span.kartu-berita__media', null,
        el('img.kartu__gambar', { src: gambar(b.gambar), alt: '', loading: 'lazy' })),
      el('div.kartu__isi', null,
        el('time.tanggal', { datetime: tanggal?.toISOString().slice(0, 10) }, b.tanggal_publikasi),
        el(tingkat, null, b.judul),
        el('p.teks-sekunder', null, Format.potongan(b.isi)),
        el('span.tautan-panah', null, 'Lihat detail', ikon('panah', 18)))));
}

function kartuDokumentasi(d) {
  const href = `/dokumentasi/${encodeURIComponent(d.slug)}`;
  return el('article.kartu', null,
    el('img.kartu__gambar', { src: gambar(d.gambar[0]?.key), alt: '', loading: 'lazy' }),
    el('div.kartu__isi', null,
      el('time.tanggal', null, d.tanggal_publikasi),
      el('h3', null, el('a.tautan-kartu', { href }, d.judul)),
      el('a.tautan-panah', { href, 'aria-label': `Lihat detail: ${d.judul}` }, 'Lihat detail', ikon('panah', 18))));
}

function kartuAlbum(d) {
  const href = `/dokumentasi/${encodeURIComponent(d.slug)}`;
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
    el('img', { src: gambar(t.foto), alt: `Foto ${t.nama}`, loading: 'lazy', width: 96, height: 120 }),
    el('div.kartu-tim__identitas', null, el('h2', null, t.nama), el('p.kartu-tim__jabatan', null, t.posisi)),
    el('div.kartu-tim__riset', null, el('p.kartu-tim__label', null, 'Bidang Penelitian'), el('p', null, t.bidang_riset)));
}

function kartuPortofolio(p) {
  return el('article.kartu.kartu-album', null, el('a.kartu-album__tautan', { href: p.href },
    el('img.kartu__gambar', { src: gambar(p.gambar), alt: '', loading: 'lazy' }),
    el('div.kartu__isi', null,
      el('p.kartu-tim__jabatan', null, p.bidang),
      el('h3', null, p.judul),
      p.subjudul && el('p.porto__subjudul', null, p.subjudul),
      el('p.teks-sekunder', null, p.deskripsi),
      el('div.porto__status', null, p.status.split('·').map((status) => el('span.badge.badge--simulasi', null, status.trim()))),
      el('span.tautan-panah', null, p.cta, ikon('panah', 18)))));
}

/* === Bagian bersama === */
function aksiKontak() {
  const k = K.kontak;
  return el('div.kontak__aksi', null,
    tautanTombol('Hubungi lewat WhatsApp', 'terang', k.whatsapp_url, { target: '_blank', rel: 'noopener' }, 'pesan'),
    tautanTombol('Kirim email', 'garis-terang', k.email_url, null, 'surat'));
}

function bagianKontak(judul, sub) {
  return el('section.kontak', { 'aria-labelledby': 'judul-kontak' }, el('div.wadah.kontak__dalam', null,
    el('div', null, el('h2', { id: 'judul-kontak' }, judul), el('p', null, sub)),
    aksiKontak()));
}

/** Muat data ke wadah, lengkap dengan loading, galat, dan kosong. Dipanggil ulang saat panel berubah. */
function muatKe(wadah, { kerangka, ambil, gambarkan, kosong, selesai }) {
  wadah.dataset.token = String((Number(wadah.dataset.token) || 0) + 1);
  const token = wadah.dataset.token;
  wadah.replaceChildren(kerangka());
  const jalan = () => muatKe(wadah, { kerangka, ambil, gambarkan, kosong, selesai });
  ambil().then((hasil) => {
    if (wadah.dataset.token !== token) return;
    const daftar = hasil.items ?? hasil;
    if (Array.isArray(daftar) && daftar.length === 0) wadah.replaceChildren(kosong(hasil));
    else wadah.replaceChildren(...[gambarkan(hasil)].flat().filter(Boolean));
    selesai?.(hasil);
  }).catch((g) => {
    if (wadah.dataset.token !== token) return;
    wadah.replaceChildren(keadaanGalat(g, jalan));
    selesai?.(null, g);
  });
}

/* === Daftar dengan pencarian dan pagination === */
function halamanDaftar({ judul, judulTampil = judul, sub, ambil, kartu, namaJamak, kolomGrid = '' }) {
  setJudul(judul, sub);
  const wadah = el('div', { id: 'hasil', tabindex: '-1', 'aria-label': 'Hasil pencarian' });
  const status = el('p.field__bantuan', { role: 'status', 'aria-atomic': 'true' });
  const petunjuk = el('p.field__bantuan', { id: 'petunjuk-cari' });
  let halaman = Math.max(1, Number.parseInt(parameter.get('page'), 10) || 1);
  let q = (parameter.get('q') || '').trim();
  let waktu = null;
  const input = el('input.input', { type: 'search', id: 'cari', placeholder: 'Cari judul', autocomplete: 'off', 'aria-describedby': 'petunjuk-cari',
    oninput: (e) => {
      clearTimeout(waktu);
      waktu = setTimeout(() => {
        q = e.target.value.trim();
        petunjuk.textContent = q.length === 1 ? 'Ketik minimal 2 karakter.' : '';
        halaman = 1;
        simpanUrl();
        muat();
      }, 300);
    } });
  const qAktif = () => (q.length >= 2 ? q : '');
  input.value = q;
  function simpanUrl(ganti = false) {
    const url = new URL(location.href);
    q ? url.searchParams.set('q', q) : url.searchParams.delete('q');
    halaman > 1 ? url.searchParams.set('page', halaman) : url.searchParams.delete('page');
    if (url.href !== location.href) history[ganti ? 'replaceState' : 'pushState'](null, '', url);
  }
  function muat(pindahFokus = false) {
    const halamanDiminta = halaman, cariDiminta = qAktif();
    status.textContent = 'Memuat hasil…';
    muatKe(wadah, {
      kerangka: () => kerangkaKartu(9),
      ambil: async () => {
        const hasil = await ambil({ page: halamanDiminta, q: cariDiminta });
        const terakhir = Math.max(1, Math.ceil(hasil.total / hasil.limit));
        if (halamanDiminta > terakhir && halaman === halamanDiminta && qAktif() === cariDiminta) { halaman = terakhir; simpanUrl(true); return ambil({ page: halaman, q: cariDiminta }); }
        return hasil;
      },
      kosong: () => keadaan(qAktif()
        ? { judul: 'Tidak ada hasil', teks: `Tidak ada ${namaJamak} dengan judul "${qAktif()}". Coba kata lain.`, namaIkon: 'cari' }
        : { judul: `Belum ada ${namaJamak}`, teks: `${namaJamak[0].toUpperCase()}${namaJamak.slice(1)} akan tampil di sini setelah dipublikasikan.` }),
      gambarkan: (hasil) => [
        el(`div.grid-kartu${kolomGrid}`, null, hasil.items.map((b) => kartu(b))),
        pagination(hasil, (n) => { halaman = n; simpanUrl(); muat(true); }),
      ],
      selesai: (hasil, galat) => {
        status.textContent = galat ? 'Hasil belum dapat dimuat.' : hasil.total ? `Halaman ${halaman} dari ${Math.ceil(hasil.total / hasil.limit)}. ${hasil.total} ${namaJamak} ditemukan.` : 'Tidak ada hasil.';
        if (pindahFokus && document.activeElement === document.body) { wadah.focus({ preventScroll: true }); wadah.scrollIntoView({ block: 'start' }); }
      },
    });
  }
  utama.append(
    breadcrumb([{ label: judul }]),
    kepalaHalaman(judulTampil, sub),
    el('div.wadah.daftar', null,
      el('div.field.cari-blok', null, el('label', { for: 'cari' }, `Cari ${namaJamak}`), el('div.cari', null, ikon('cari', 20), input), petunjuk),
      status, wadah));
  window.addEventListener('popstate', () => {
    clearTimeout(waktu);
    const params = new URLSearchParams(location.search);
    q = (params.get('q') || '').trim(); input.value = q;
    halaman = Math.max(1, Number.parseInt(params.get('page'), 10) || 1);
    petunjuk.textContent = q.length === 1 ? 'Ketik minimal 2 karakter.' : '';
    muat();
  });
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
      tautanTombol('Kembali ke Beranda', 'utama', '/')));
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
      el('section.hero', { 'aria-labelledby': 'judul-hero', 'data-nav-section': 'index' },
        el('canvas.hero__bola', { 'aria-hidden': 'true' }),
        el('div.wadah.hero__isi', null,
          el('div.hero__teks', null,
            el('h1', { id: 'judul-hero' }, K.hero.judul),
            el('p', null, K.hero.sub),
            el('div.hero__aksi', null,
              tautanTombol('Diskusikan Kebutuhan AI', 'terang', K.kontak.whatsapp_url, { target: '_blank', rel: 'noopener' }, 'pesan'),
              tautanTombol('Lihat Layanan', 'garis-terang', '#layanan'))))),
      el('section.lembar.lembar--putih', { id: 'layanan', 'aria-labelledby': 'judul-layanan', 'data-nav-section': 'layanan' }, el('div.wadah', null,
        el('h2', { id: 'judul-layanan' }, 'Dua cara kami membantu Anda'),
        el('div.layanan-grid', null, K.layanan.map((l, i) => el('article', { class: `layanan-kartu layanan-kartu--${i ? 'gelap' : 'terang'}` },
          el('h3', null, l.nama), el('p', null, l.deskripsi),
          tautanTombol(i ? 'Lihat Solusi AI' : 'Lihat Pelatihan', i ? 'terang' : 'utama', l.href, null, 'panah')))))),
      el('section.lembar.lembar--muda', { 'aria-labelledby': 'judul-dok', 'data-nav-section': 'kabar' }, el('div.wadah', null,
        el('div.judul-baris', null, el('h2', { id: 'judul-dok' }, 'Galeri Dokumentasi'), el('a.tautan-panah', { href: '/dokumentasi' }, 'Lihat Galeri Dokumentasi', ikon('panah', 18))),
        wadahDok)),
      el('section.lembar.lembar--langit', { 'aria-labelledby': 'judul-testi' }, el('div.wadah.testi', null,
        el('h2', { id: 'judul-testi' }, 'Cerita dari mereka yang belajar dan bekerja bersama kami'),
        wadahTesti)),
      el('section.lembar.lembar--putih', { 'aria-labelledby': 'judul-berita', 'data-nav-section': 'kabar' }, el('div.wadah', null,
        el('div.judul-baris', null, el('h2', { id: 'judul-berita' }, 'Berita terbaru'), el('a.tautan-panah', { href: '/berita' }, 'Semua berita', ikon('panah', 18))),
        wadahBerita)),
      bagianKontak('Punya kebutuhan AI? Mari berdiskusi.', 'Ingin mengembangkan kemampuan AI atau mengeksplorasi penerapannya? Ceritakan kebutuhan Anda, dan mari temukan langkah yang sesuai bersama UBAYA AI Center.'));

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
      gambarkan: (h) => el('div.grid-kartu.grid-kartu--tiga', null, h.items.map((b) => kartuBerita(b, 'h3'))),
      kosong: () => keadaan({ judul: 'Belum ada berita', teks: 'Berita terbaru akan tampil di sini.' }),
    });
    aktifkanLembar();
  },

  pelatihan() {
    const p = K.pelatihan;
    setJudul('Pelatihan', p.hero_sub);
    const hasil = el('div', { role: 'status', 'aria-live': 'polite', 'aria-atomic': 'true' },
      el('p.teks-sekunder', null, 'Pilih satu kebutuhan di atas untuk melihat jalur yang dapat dipertimbangkan.'));
    const tautanHasil = el('a.tautan-panah', { hidden: true });
    const pilihan = el('fieldset.pelatihan-pilihan', null,
      el('legend.sr-only', null, 'Pilih kebutuhan pelatihan'),
      p.jalur.map((jalur, i) => el('label.pelatihan-pilihan__item', { for: `jalur-${i}` },
        el('input', { type: 'radio', name: 'jalur-pelatihan', id: `jalur-${i}`, value: String(i),
          onchange: (e) => {
            if (!e.target.checked) return;
            hasil.replaceChildren(el('div.pelatihan-hasil__isi', null,
              el('h3', null, jalur.layanan), el('p', null, jalur.alasan)));
            tautanHasil.setAttribute('href', jalur.href);
            tautanHasil.replaceChildren(jalur.tautan, ikon('panah', 18));
            tautanHasil.hidden = false;
          } }), el('span', null, jalur.kebutuhan))));
    utama.append(
      breadcrumb([{ label: 'Pelatihan' }]),
      el('section.sampul.sampul--belajar', null, el('div.wadah', null,
        el('h1', null, p.hero_judul), el('p', null, p.hero_sub),
        el('div.hero__aksi', null,
          tautanTombol('Konsultasi Gratis', 'utama', K.kontak.whatsapp_url, { target: '_blank', rel: 'noopener' }, 'pesan'),
          tautanTombol('Lihat Pilihan Pelatihan', 'kedua', '#pilihan-pelatihan')))),
      el('section.bagian.wadah.pelatihan-pemilih', { id: 'pilihan-pelatihan', 'aria-labelledby': 'j-pilihan' },
        el('h2', { id: 'j-pilihan' }, 'Pelatihan Mana yang Sesuai dengan Kebutuhan Anda?'),
        el('p.bagian__sub', null, 'Pilih kebutuhan yang paling mendekati tujuan Anda untuk melihat jalur layanan yang dapat dipertimbangkan.'),
        pilihan,
        el('div.pelatihan-hasil', null, hasil, tautanHasil),
        el('p.teks-sekunder', null, 'Jika kebutuhan Anda mencakup keduanya, bentuk kegiatan dapat ditentukan melalui konsultasi.')),
      el('section.bagian.bagian--muda', { id: 'khusus', 'aria-labelledby': 'j-khusus' }, el('div.wadah', null,
        el('p.pelatihan-eyebrow', null, 'Pengembangan Kompetensi'),
        el('h2', { id: 'j-khusus' }, 'Pelatihan Khusus'),
        el('h3.pelatihan-subjudul', null, 'Pembelajaran AI yang Relevan dengan Kebutuhan Anda'),
        el('div.pelatihan-isi', null, p.khusus_isi.map((teks) => el('p', null, teks))),
        el('ul.pelatihan-fokus', null, p.khusus_fokus.map((teks) => el('li', null, teks))),
        tautanTombol('Konsultasi Gratis', 'utama', K.kontak.whatsapp_url, { target: '_blank', rel: 'noopener' }, 'pesan'),
        el('div.pelatihan-workflow', null,
          el('h3', null, 'Empat Langkah Konsultasi Pelatihan'),
          el('ol.alur', { role: 'list' }, p.alur_konsultasi.map((l, i) => el('li', { style: `--alur-urutan:${i}` }, el('span.alur__nomor', null, String(i + 1)), el('h4', null, l.judul), el('p.teks-sekunder', null, l.deskripsi))))))),
      el('section.bagian.wadah', { id: 'workshop', 'aria-labelledby': 'j-workshop' },
        el('p.pelatihan-eyebrow', null, 'Eksplorasi AI untuk Organisasi'),
        el('h2', { id: 'j-workshop' }, 'Workshop & Seminar untuk Bisnis'),
        el('h3.pelatihan-subjudul', null, 'Mengenal Potensi AI dalam Proses Bisnis'),
        el('div.pelatihan-isi', null, p.workshop_isi.map((teks) => el('p', null, teks))),
        el('ul.pelatihan-fokus', null, p.workshop_fokus.map((teks) => el('li', null, teks))),
        tautanTombol('Tanya Workshop Bisnis', 'utama', K.kontak.whatsapp_url, { target: '_blank', rel: 'noopener' }, 'pesan')),
      el('section.bagian.wadah.pelatihan-reguler', { id: 'reguler', 'aria-labelledby': 'j-reguler' },
        el('div.segera', null, el('p.pelatihan-eyebrow', null, 'Coming Soon'),
          el('h2', { id: 'j-reguler' }, 'Pelatihan Reguler'), el('p.teks-sekunder', null, p.reguler))),
      el('section.kontak', { 'aria-labelledby': 'j-kontak-pelatihan' }, el('div.wadah.kontak__dalam', null,
        el('div', null, el('h2', { id: 'j-kontak-pelatihan' }, 'Diskusikan Kebutuhan Pelatihan AI Anda'),
          el('p', null, 'Baik untuk pengembangan kompetensi maupun eksplorasi peluang AI di organisasi, UBAYA AI Center terbuka untuk mendiskusikan bentuk kegiatan yang sesuai dengan kebutuhan Anda.')),
        aksiKontak())));
    const alur = utama.querySelector('.pelatihan-workflow .alur');
    const desktop = matchMedia('(min-width: 64rem)');
    const kurangGerak = matchMedia('(prefers-reduced-motion: reduce)');
    if (desktop.matches && !kurangGerak.matches && 'IntersectionObserver' in window) {
      const pengamat = new IntersectionObserver((entri) => {
        if (entri.some((e) => e.isIntersecting && e.intersectionRatio >= 1)) {
          alur.classList.add('alur--mengalir');
          pengamat.disconnect();
        }
      }, { threshold: 1 });
      pengamat.observe(alur);
      alur.addEventListener('animationend', (e) => {
        if (e.animationName === 'pelatihan-indikator' && e.target === alur.lastElementChild.querySelector('.alur__nomor')) alur.classList.remove('alur--mengalir');
      });
      const hentikan = () => {
        if (!desktop.matches || kurangGerak.matches) {
          pengamat.disconnect();
          alur.classList.remove('alur--mengalir');
        }
      };
      desktop.addEventListener('change', hentikan);
      kurangGerak.addEventListener('change', hentikan);
    }
  },

  inference() {
    const n = K.inference;
    const ilustrasiJenis = ['it-service-management-line.svg', 'data-science-line.svg'];
    const teksHeadline = el('span.inference-headline', null, n.hero_judul);
    const gerakHero = matchMedia('(prefers-reduced-motion: reduce)');
    const warnaPaksa = matchMedia('(forced-colors: active)');
    if (!gerakHero.matches && !warnaPaksa.matches && window.CSS?.supports('background-clip', 'text') && window.CSS.supports('box-decoration-break', 'slice')) {
      const selesai = () => {
        teksHeadline.classList.remove('inference-headline--reveal');
        gerakHero.removeEventListener('change', selesai);
        warnaPaksa.removeEventListener('change', selesai);
      };
      teksHeadline.classList.add('inference-headline--reveal');
      teksHeadline.addEventListener('animationend', selesai, { once: true });
      teksHeadline.addEventListener('animationcancel', selesai, { once: true });
      gerakHero.addEventListener('change', selesai);
      warnaPaksa.addEventListener('change', selesai);
    }
    setJudul('Inference and Solutions', n.hero_sub);
    utama.append(
      breadcrumb([{ label: 'Inference and Solutions' }]),
      el('section.sampul.sampul--enterprise', { 'data-nav-section': 'layanan' }, el('div.wadah.sampul__dua', null,
        el('div', null, el('h1', null, teksHeadline), el('p', null, n.hero_sub),
          el('div.hero__aksi', null, tautanTombol('Lihat Contoh Penerapan', 'terang', '#portofolio', null, 'panah'))),
        carousel(n.pertanyaan.map((t) => el('p.pertanyaan', null, `“${t}”`)), { label: 'Pertanyaan untuk Anda', intervalMs: 6000 }))),
      el('section.bagian.wadah', { 'aria-labelledby': 'j-jenis' },
        el('h2', { id: 'j-jenis' }, 'Jelajahi Peluang Penerapan AI'),
        el('p.bagian__sub', null, 'Solusi inference utama yang kami tawarkan.'),
        el('div.jenis-grid', null, n.jenis.map((j, i) => el('article.jenis', null,
          el('img.jenis__ilustrasi', { src: gambar(ilustrasiJenis[i]), alt: '', width: 180, height: 108, loading: 'lazy' }),
          el('h3', null, j.nama), el('p.teks-sekunder', null, j.deskripsi))))),
      el('section.bagian.bagian--muda', { 'aria-labelledby': 'j-alur' }, el('div.wadah', null,
        el('div.inference-workflow', null,
          el('h2', { id: 'j-alur' }, 'Tahapan penerapan model AI.'),
          el('ol.alur.alur--inference', { role: 'list' }, n.alur.map((l, i) => el('li', { style: `--alur-urutan:${i}` }, el('span.alur__nomor', null, String(i + 1)), el('h3', null, l.judul), el('p.teks-sekunder', null, l.deskripsi))))))),
      el('section.bagian.wadah', { id: 'portofolio', tabindex: '-1', 'aria-labelledby': 'j-porto', 'data-nav-section': 'contoh-aplikasi' },
        el('div.porto__kepala', null,
          el('p.tentang-label', null, 'Portofolio'),
          el('h2', { id: 'j-porto' }, 'Jelajahi Portofolio Kami.'),
          el('p.bagian__sub', null, 'Kenali contoh penerapan teknologi AI melalui demonstrasi aplikasi.')),
        el('div.grid-kartu.grid-kartu--tiga.porto', null, n.portofolio.map(kartuPortofolio))),
      bagianKontak('Diskusikan kebutuhan AI Anda.', 'Punya model AI, data, atau ide penerapan yang ingin dikembangkan? Ceritakan kepada tim UBAYA AI Center, dan mari diskusikan langkah yang paling sesuai.'));
    const alur = utama.querySelector('.alur--inference');
    const lebar = matchMedia('(min-width: 80rem)');
    const kurangGerak = matchMedia('(prefers-reduced-motion: reduce)');
    if (lebar.matches && !kurangGerak.matches && 'IntersectionObserver' in window) {
      const pengamat = new IntersectionObserver((entri) => {
        if (entri.some((e) => e.isIntersecting && e.intersectionRatio >= 1)) {
          alur.classList.add('inference--mengalir');
          pengamat.disconnect();
        }
      }, { threshold: 1 });
      pengamat.observe(alur);
      alur.addEventListener('animationend', (e) => {
        if (e.animationName === 'inference-indikator' && e.target === alur.lastElementChild.querySelector('.alur__nomor')) alur.classList.remove('inference--mengalir');
      });
      const hentikan = () => {
        if (!lebar.matches || kurangGerak.matches) {
          pengamat.disconnect();
          alur.classList.remove('inference--mengalir');
        }
      };
      lebar.addEventListener('change', hentikan);
      kurangGerak.addEventListener('change', hentikan);
    }
  },

  'jejak-alumni'() {
    setJudul('JejakAlumni (Demo Konsep)', K.jejak_alumni.disclaimer);
    utama.append(
      breadcrumb([{ label: 'Inference and Solutions', href: '/layanan/inference-solutions' }, { label: 'JejakAlumni' }]),
      JejakAlumni.halaman(),
      bagianKontak('Ingin membahas data lulusan di institusi Anda?', 'Ceritakan kebutuhan evaluasi lulusan Anda. Tim UBAYA AI Center dapat membantu membahas langkah berikutnya.'));
  },

  'perencanaan-implan-gigi'() {
    const d = K.implan_gigi;
    setJudul(`${d.judul} (Konseptual)`, d.disclaimer);
    utama.append(
      breadcrumb([{ label: 'Inference and Solutions', href: '/layanan/inference-solutions' }, { label: d.judul }]),
      ...ImplanGigi.bagian(),
      bagianKontak('Tertarik Mengembangkan Solusi AI Serupa?', 'Diskusikan peluang kolaborasi riset atau pengembangan solusi AI sesuai kebutuhan institusi Anda.'));
  },

  'document-intelligence'() {
    const d = K.document_intelligence;
    setJudul(`${d.judul} (Demo Konsep)`, d.disclaimer);
    utama.append(
      breadcrumb([{ label: 'Inference and Solutions', href: '/layanan/inference-solutions' }, { label: d.judul }]),
      ...DocNext.bagian(),
      bagianKontak('Punya Proses Bisnis yang Ingin Dibantu AI?', 'Diskusikan kebutuhan pengolahan dokumen, integrasi sistem, atau implementasi AI di lingkungan internal perusahaan bersama UBAYA AI Center.'));
  },

  berita() {
    halamanDaftar({ judul: 'Berita', judulTampil: 'Kabar Terbaru dari UBAYA AI Center.', sub: 'Ikuti kegiatan, kolaborasi, dan perkembangan UBAYA AI Center melalui cerita dan informasi terbaru kami.', ambil: api.daftarBerita, kartu: kartuBerita, namaJamak: 'berita', kolomGrid: '.grid-kartu--berita' });
  },

  dokumentasi() {
    halamanDaftar({ judul: 'Galeri Dokumentasi', sub: 'Melihat lebih dekat kegiatan, kolaborasi, dan momen bersama UBAYA AI Center.', ambil: api.daftarDokumentasi, kartu: kartuAlbum, namaJamak: 'album', kolomGrid: '.grid-kartu--album' });
  },

  'berita-detail'() {
    const slug = slugDariUrl();
    const wadah = el('div');
    utama.replaceChildren(wadah);
    muatDetail(wadah, () => api.detailBerita(slug), 'Berita', (b) => {
      const isi = Format.paragraf(b.isi);
      setJudul(b.judul, Format.potongan(b.isi, 160));
      return [
        breadcrumb([{ label: 'Berita', href: '/berita' }, { label: b.judul }]),
        el('article.artikel', null,
          el('header.artikel__kepala', null,
            el('h1', null, b.judul),
            el('p.artikel__meta', null, el('time.tanggal', null, b.tanggal_publikasi), ' ', el('span', null, 'oleh Tim Media Ubaya AI Center'))),
          el('img.artikel__gambar', { src: gambar(b.gambar), alt: '' }),
          el('div.artikel__isi', null, isi.map((p) => el('p', null, p))),
          el('a.tautan-panah.artikel__kembali', { href: '/berita' }, 'Semua berita', ikon('panah', 18))),
      ];
    });
  },

  'dokumentasi-detail'() {
    const slug = slugDariUrl();
    const wadah = el('div');
    utama.replaceChildren(wadah);
    muatDetail(wadah, () => api.detailDokumentasi(slug), 'Dokumentasi', (d) => {
      setJudul(d.judul, `Dokumentasi kegiatan ${d.judul} — ${d.tanggal_publikasi}`);
      return [
        breadcrumb([{ label: 'Galeri Dokumentasi', href: '/dokumentasi' }, { label: d.judul }]),
        el('article.wadah.dok', null,
          el('header.artikel__kepala', null, el('h1', null, d.judul), el('p.artikel__meta', null, el('time.tanggal', null, d.tanggal_publikasi))),
          el(`div.galeri.galeri--${d.gambar.length}`, null, d.gambar.map((g, i) => el('img', { src: gambar(g.key), alt: `Foto ${i + 1} dari ${d.gambar.length}: ${d.judul}`, loading: i ? 'lazy' : null }))),
          el('a.tautan-panah.artikel__kembali', { href: '/dokumentasi' }, 'Semua Galeri Dokumentasi', ikon('panah', 18))),
      ];
    });
  },

  tim() {
    const judul = 'Tim';
    const sub = 'Kenali tim yang berkontribusi dalam riset, pengembangan, dan pembelajaran AI di UBAYA AI Center.';
    setJudul(judul, sub);
    const wadah = el('div');
    function muat() {
      muatKe(wadah, {
        kerangka: () => kerangkaKartu(6),
        ambil: () => api.daftarTim(),
        kosong: () => keadaan({ judul: 'Belum ada data tim', teks: 'Susunan tim akan tampil di sini.' }),
        gambarkan: (h) => el('div.grid-tim', null, h.items.map(kartuTim)),
      });
    }
    utama.append(breadcrumb([{ label: judul }]),
      el('header.kepala-halaman.wadah', null, el('h1', null, 'Berkenalan dengan Tim Kami.'), el('p', null, sub)),
      el('div.wadah.daftar.daftar--tim', null, wadah),
      el('section.kontak.kontak--tim', { 'aria-labelledby': 'judul-kontak-tim' }, el('div.wadah.kontak__dalam', null,
        el('div', null,
          el('h2', { id: 'judul-kontak-tim' }, 'Mari Diskusikan Kebutuhan Anda.'),
          el('p', null, 'Punya pertanyaan, ide kolaborasi, atau ingin mengetahui bagaimana AI dapat membantu? Hubungi UBAYA AI Center untuk memulai percakapan.')),
        aksiKontak())));
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
        aksiKontak())));
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
const menuAktif = ['pelatihan', 'inference', 'jejak-alumni', 'perencanaan-implan-gigi', 'document-intelligence'].includes(namaHalaman) ? 'layanan' : namaHalaman.replace('-detail', '');
document.body.prepend(el('a.lewati', { href: '#utama' }, 'Lewati ke konten'), navbar(menuAktif));
document.body.append(footer());
HALAMAN[namaHalaman]();

// Target anchor sudah ada setelah render; tunggu font agar posisi awal tidak meleset.
if (namaHalaman === 'inference' && location.hash === '#portofolio') {
  let dibatalkan = false;
  const batalkan = () => { dibatalkan = true; };
  const interaksi = ['wheel', 'touchstart', 'pointerdown', 'keydown'];
  interaksi.forEach((jenis) => window.addEventListener(jenis, batalkan, { once: true, passive: true }));
  Promise.resolve(document.fonts?.ready).then(() => requestAnimationFrame(() => {
    interaksi.forEach((jenis) => window.removeEventListener(jenis, batalkan));
    if (!dibatalkan && location.hash === '#portofolio') document.getElementById('portofolio')?.scrollIntoView({ block: 'start', behavior: 'instant' });
  }));
}
