/* Komponen bersama. Konten selalu lewat textContent, tidak pernah innerHTML. */
const IKON = {
  menu: 'M4 6h16M4 12h16M4 18h16',
  tutup: 'M6 6l12 12M18 6L6 18',
  bawah: 'M6 9l6 6 6-6',
  kiri: 'M15 6l-6 6 6 6',
  kanan: 'M9 6l6 6-6 6',
  panah: 'M5 12h14M13 6l6 6-6 6',
  cari: 'M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16zM21 21l-4.3-4.3',
  surat: 'M3 6h18v12H3zM3 7l9 7 9-7',
  pesan: 'M21 12a8.5 8.5 0 0 1-12.6 7.4L3 21l1.6-5.4A8.5 8.5 0 1 1 21 12z',
  pin: 'M12 21s7-6.1 7-11.5A7 7 0 0 0 5 9.5C5 14.9 12 21 12 21zM12 12a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5z',
  kamera: 'M4 4h16v16H4zM12 16a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM16.5 7.5v.01',
  tambah: 'M12 5v14M5 12h14',
  kunci: 'M6 11h12v9H6zM8.5 11V7.5a3.5 3.5 0 0 1 7 0V11',
  ulang: 'M4 12a8 8 0 0 1 13.7-5.7L20 8.5M20 4v4.5h-4.5M20 12a8 8 0 0 1-13.7 5.7L4 15.5M4 20v-4.5h4.5',
  peringatan: 'M12 4L2.5 20h19zM12 10v4.5M12 17.5v.01',
  keluar: 'M15 4h4v16h-4M10 8l-4 4 4 4M6 12h9',
  berita: 'M5 4h14v16H5zM8 8h8M8 12h8M8 16h5',
  galeri: 'M4 5h16v14H4zM4 16l5-5 4 4 2-2 5 5M15 9.5v.01',
  kutip: 'M9 7H5v6h4v-1c0 2-1 3-3 4M19 7h-4v6h4v-1c0 2-1 3-3 4',
  orang: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM4 21c1-4 4-6 8-6s7 2 8 6',
  cek: 'M5 12.5l4.5 4.5L19 7.5',
  kosong: 'M4 13l2.5-8h11l2.5 8v6H4zM4 13h5l1.5 2.5h3L15 13h5',
};

/** Buat elemen: el('a.kelas', {href}, anak...). Anak string menjadi text node. */
function el(tagKelas, atribut, ...anak) {
  const [tag, ...kelas] = tagKelas.split('.');
  const node = document.createElement(tag || 'div');
  if (kelas.length) node.className = kelas.join(' ');
  for (const [k, v] of Object.entries(atribut ?? {})) {
    if (v === null || v === undefined || v === false) continue;
    if (k.startsWith('on')) node.addEventListener(k.slice(2), v);
    else if (k === 'nilai') node.value = v;
    else node.setAttribute(k, v === true ? '' : v);
  }
  node.append(...anak.flat().filter((a) => a !== null && a !== undefined && a !== false));
  return node;
}

function ikon(nama, ukuran = 20) {
  const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  svg.setAttribute('viewBox', '0 0 24 24');
  svg.setAttribute('width', ukuran);
  svg.setAttribute('height', ukuran);
  svg.setAttribute('fill', 'none');
  svg.setAttribute('stroke', 'currentColor');
  svg.setAttribute('stroke-width', '1.8');
  svg.setAttribute('stroke-linecap', 'round');
  svg.setAttribute('stroke-linejoin', 'round');
  svg.setAttribute('aria-hidden', 'true');
  const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
  path.setAttribute('d', IKON[nama]);
  svg.append(path);
  return svg;
}

function tombol(label, varian, atribut, namaIkon) {
  return el(`button.tombol.tombol--${varian}`, { type: 'button', ...atribut }, namaIkon && ikon(namaIkon, 18), label);
}

function tautanTombol(label, varian, href, atribut, namaIkon) {
  return el(`a.tombol.tombol--${varian}`, { href, ...atribut }, namaIkon && ikon(namaIkon, 18), label);
}

function badgeStatus(status) {
  const label = { draft: 'Draft', posted: 'Posted', trash: 'Trash' }[status] ?? status;
  return el(`span.badge.badge--${status}`, null, label);
}

/** Keadaan kosong dan galat. */
function keadaan({ judul, teks, namaIkon = 'kosong', galat = false, aksi }) {
  return el(`div.keadaan${galat ? '.keadaan--galat' : ''}`, { role: galat ? 'alert' : null }, ikon(namaIkon, 32), el('h3', null, judul), teks && el('p', null, teks), aksi);
}

/** Pesan galat aman dari respons API; 429 menyebut waktu tunggu. */
function pesanGalat(galat) {
  if (galat?.status === 429) return `Terlalu banyak permintaan. Coba lagi dalam ${galat.retry_after ?? 30} detik.`;
  return galat?.error?.message ?? 'Terjadi kendala. Coba lagi beberapa saat lagi.';
}

function keadaanGalat(galat, cobaLagi) {
  return keadaan({
    judul: galat?.status === 429 ? 'Permintaan sedang dibatasi' : 'Data belum bisa dimuat',
    teks: pesanGalat(galat),
    namaIkon: 'peringatan',
    galat: true,
    aksi: cobaLagi && tombol('Coba lagi', 'kedua', { onclick: cobaLagi }, 'ulang'),
  });
}

function kerangkaKartu(jumlah = 3) {
  return el('div.grid-kartu', { 'aria-busy': 'true', 'aria-label': 'Memuat data', role: 'status' }, Array.from({ length: jumlah }, () => el('div.kerangka.kerangka--kartu')));
}

function kerangkaBaris(jumlah = 5) {
  return el('div', { 'aria-busy': 'true', 'aria-label': 'Memuat data', role: 'status' }, Array.from({ length: jumlah }, () => el('div.kerangka.kerangka--baris')));
}

function pagination({ page, limit, total }, keHalaman) {
  const jumlah = Math.ceil(total / limit);
  if (jumlah <= 1) return null;
  const nomor = Array.from({ length: jumlah }, (_, i) => i + 1);
  return el('nav.halaman-nav', { 'aria-label': 'Pagination' },
    el('button', { type: 'button', disabled: page <= 1, 'aria-label': 'Halaman sebelumnya', onclick: () => keHalaman(page - 1) }, ikon('kiri', 18)),
    nomor.map((n) => el('button', { type: 'button', 'aria-current': n === page ? 'page' : null, 'aria-label': `Halaman ${n}`, onclick: () => keHalaman(n) }, String(n))),
    el('button', { type: 'button', disabled: page >= jumlah, 'aria-label': 'Halaman berikutnya', onclick: () => keHalaman(page + 1) }, ikon('kanan', 18)));
}

/** Carousel: otomatis hanya di perangkat ber-hover (berhenti saat disorot atau difokus);
    layar sentuh dan reduced motion manual lewat panah, karena tanpa tombol jeda (WCAG 2.2.2). */
function carousel(slides, { label, intervalMs }) {
  let posisi = 0;
  const berjalan = matchMedia('(hover: hover)').matches && !matchMedia('(prefers-reduced-motion: reduce)').matches;
  let timer = null;
  const nodeSlide = slides.map((isi, i) => el('div.carousel__slide', { role: 'group', 'aria-roledescription': 'slide', 'aria-label': `${i + 1} dari ${slides.length}` }, isi));
  const teksPosisi = el('span.carousel__posisi', { 'aria-hidden': 'true' });

  function tampil(i) {
    posisi = (i + slides.length) % slides.length;
    nodeSlide.forEach((s, n) => s.classList.toggle('aktif', n === posisi));
    teksPosisi.textContent = `${posisi + 1} / ${slides.length}`;
  }
  function atur() {
    clearInterval(timer);
    if (berjalan && slides.length > 1) timer = setInterval(() => tampil(posisi + 1), intervalMs);
  }
  const akar = el('div.carousel', { role: 'region', 'aria-roledescription': 'carousel', 'aria-label': label,
    onmouseenter: () => clearInterval(timer), onmouseleave: atur, onfocusin: () => clearInterval(timer), onfocusout: atur },
    nodeSlide,
    slides.length > 1 && el('div.carousel__kontrol', null,
      el('button.carousel__tombol', { type: 'button', 'aria-label': 'Sebelumnya', onclick: () => tampil(posisi - 1) }, ikon('kiri', 18)),
      teksPosisi,
      el('button.carousel__tombol', { type: 'button', 'aria-label': 'Berikutnya', onclick: () => tampil(posisi + 1) }, ikon('kanan', 18))));
  tampil(0);
  atur();
  return akar;
}

/** Dialog konfirmasi berbasis <dialog>; hasilnya true bila dikonfirmasi. */
function konfirmasi({ judul, teks, labelYa, varianYa = 'utama', labelTidak = 'Batal' }) {
  return new Promise((selesai) => {
    const dialog = el('dialog.modal', { 'aria-labelledby': 'modal-judul' },
      el('div.modal__isi', null,
        el('h2', { id: 'modal-judul', style: 'font-size:var(--judul-3)' }, judul),
        el('p', null, teks),
        el('div.modal__aksi', null,
          labelTidak && tombol(labelTidak, 'kedua', { onclick: () => tutup(false) }),
          tombol(labelYa, varianYa, { onclick: () => tutup(true) }))));
    function tutup(hasil) { dialog.close(); dialog.remove(); selesai(hasil); }
    dialog.addEventListener('cancel', (e) => { e.preventDefault(); tutup(false); });
    document.body.append(dialog);
    dialog.showModal();
  });
}

function toast(pesan, galat = false) {
  let wadah = document.querySelector('.toast-wadah');
  if (!wadah) {
    wadah = el('div.toast-wadah', { role: 'status', 'aria-live': 'polite' });
    document.body.append(wadah);
  }
  const item = el(`div.toast${galat ? '.toast--galat' : ''}`, null, ikon(galat ? 'peringatan' : 'cek', 18), el('span', null, pesan));
  wadah.append(item);
  setTimeout(() => item.remove(), 5000);
}
