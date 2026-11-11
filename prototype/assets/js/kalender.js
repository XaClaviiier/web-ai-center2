/* Komponen kalender: kolom teks DD-MM-YYYY plus popup bulan. Tanggal setelah hari ini
   tidak bisa dipilih. "Hari ini" memakai tanggal lokal perangkat, bukan UTC. */
IKON.kalender = 'M4 6h16v14H4zM4 10h16M8 3.5V7M16 3.5V7';

const Kalender = (() => {
  const BULAN = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];
  const HARI = ['Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu', 'Minggu'];

  const hariIni = () => { const n = new Date(); return new Date(n.getFullYear(), n.getMonth(), n.getDate()); };
  const dua = (n) => String(n).padStart(2, '0');
  const keTeks = (d) => `${dua(d.getDate())}-${dua(d.getMonth() + 1)}-${d.getFullYear()}`;
  const hariIniTeks = () => keTeks(hariIni());
  /** Tanggal lokal dari "DD-MM-YYYY"; null bila format salah atau tanggal tidak ada. */
  function dariTeks(teks) {
    const u = Format.parseTanggal(teks);
    return u ? new Date(u.getUTCFullYear(), u.getUTCMonth(), u.getUTCDate()) : null;
  }
  const diMasaDepan = (d) => d > hariIni();
  const sama = (a, b) => a && b && a.getTime() === b.getTime();
  const indeksHari = (d) => (d.getDay() + 6) % 7; // Senin = 0
  const labelPenuh = (d) => `${HARI[indeksHari(d)]}, ${d.getDate()} ${BULAN[d.getMonth()]} ${d.getFullYear()}`;

  /** Pasang kalender pada input teks. Mengembalikan elemen pembungkus. */
  function pasang(input, { label }) {
    let fokus = hariIni();
    let bulanTampil = new Date(fokus.getFullYear(), fokus.getMonth(), 1);

    const buka = el('button.tombol.tombol--kedua.kalender__buka', { type: 'button', 'aria-label': `Buka kalender ${label}`, 'aria-expanded': 'false', onclick: () => (popup.hidden ? bukaPopup() : tutup(true)) }, ikon('kalender', 20));
    const judul = el('p.kalender__judul', { 'aria-live': 'polite' });
    const sebelum = el('button.carousel__tombol', { type: 'button', 'aria-label': 'Bulan sebelumnya', onclick: () => geserBulan(-1) }, ikon('kiri', 18));
    const berikut = el('button.carousel__tombol', { type: 'button', 'aria-label': 'Bulan berikutnya', onclick: () => geserBulan(1) }, ikon('kanan', 18));
    const isiTabel = el('tbody');
    const popup = el('div.kalender', { role: 'dialog', 'aria-label': `Pilih ${label}`, hidden: true, onkeydown: tombolKeyboard },
      el('div.kalender__kepala', null, sebelum, judul, berikut),
      el('table.kalender__tabel', { 'aria-label': 'Tanggal' },
        el('thead', null, el('tr', null, HARI.map((h) => el('th', { scope: 'col', abbr: h, title: h }, h.slice(0, 3))))),
        isiTabel));

    function gambar() {
      const terpilih = dariTeks(input.value);
      const sekarang = hariIni();
      judul.textContent = `${BULAN[bulanTampil.getMonth()]} ${bulanTampil.getFullYear()}`;
      const jumlahHari = new Date(bulanTampil.getFullYear(), bulanTampil.getMonth() + 1, 0).getDate();
      const sel = Array.from({ length: indeksHari(bulanTampil) }, () => el('td'));
      for (let tgl = 1; tgl <= jumlahHari; tgl++) {
        const d = new Date(bulanTampil.getFullYear(), bulanTampil.getMonth(), tgl);
        const masaDepan = diMasaDepan(d);
        const kelas = ['kalender__hari', sama(d, sekarang) && 'kalender__hari--ini', sama(d, terpilih) && 'kalender__hari--pilih'].filter(Boolean).join(' ');
        sel.push(el('td', null, el('button', { type: 'button', class: kelas, 'data-tanggal': keTeks(d), 'aria-label': labelPenuh(d),
          'aria-current': sama(d, sekarang) ? 'date' : null, 'aria-pressed': sama(d, terpilih) ? 'true' : 'false',
          tabindex: sama(d, fokus) ? '0' : '-1', disabled: masaDepan, onclick: () => pilih(d) }, String(tgl))));
      }
      while (sel.length % 7) sel.push(el('td'));
      const baris = [];
      for (let i = 0; i < sel.length; i += 7) baris.push(el('tr', null, sel.slice(i, i + 7)));
      isiTabel.replaceChildren(...baris);
      // Bulan yang seluruhnya di masa depan tidak bisa dibuka.
      berikut.disabled = new Date(bulanTampil.getFullYear(), bulanTampil.getMonth() + 1, 1) > sekarang;
    }

    function fokuskanHari() {
      popup.querySelector(`[data-tanggal="${keTeks(fokus)}"]`)?.focus();
    }

    function geserBulan(arah) {
      const baru = new Date(bulanTampil.getFullYear(), bulanTampil.getMonth() + arah, 1);
      if (baru > hariIni()) return;
      bulanTampil = baru;
      const akhir = new Date(baru.getFullYear(), baru.getMonth() + 1, 0).getDate();
      fokus = new Date(baru.getFullYear(), baru.getMonth(), Math.min(fokus.getDate(), akhir));
      if (diMasaDepan(fokus)) fokus = hariIni();
      gambar();
      if (berikut.disabled) sebelum.focus();
    }

    function tombolKeyboard(e) {
      if (e.key === 'Escape') { e.preventDefault(); tutup(true); return; }
      const langkah = { ArrowLeft: -1, ArrowRight: 1, ArrowUp: -7, ArrowDown: 7 }[e.key];
      if (!langkah || !e.target.matches('.kalender__hari')) return;
      e.preventDefault();
      const baru = new Date(fokus.getFullYear(), fokus.getMonth(), fokus.getDate() + langkah);
      if (diMasaDepan(baru)) return;
      fokus = baru;
      bulanTampil = new Date(baru.getFullYear(), baru.getMonth(), 1);
      gambar();
      fokuskanHari();
    }

    function pilih(d) {
      input.value = keTeks(d);
      input.dispatchEvent(new Event('change', { bubbles: true }));
      tutup(true);
    }

    function diLuar(e) {
      if (!pembungkus.isConnected) { lepasPendengar(); return; }
      if (!pembungkus.contains(e.target)) tutup(false);
    }
    function lepasPendengar() {
      document.removeEventListener('mousedown', diLuar);
      document.removeEventListener('focusin', diLuar);
    }
    function bukaPopup() {
      const nilai = dariTeks(input.value);
      fokus = nilai && !diMasaDepan(nilai) ? nilai : hariIni();
      bulanTampil = new Date(fokus.getFullYear(), fokus.getMonth(), 1);
      gambar();
      popup.hidden = false;
      buka.setAttribute('aria-expanded', 'true');
      document.addEventListener('mousedown', diLuar);
      document.addEventListener('focusin', diLuar);
      fokuskanHari();
    }
    function tutup(kembalikanFokus) {
      popup.hidden = true;
      buka.setAttribute('aria-expanded', 'false');
      lepasPendengar();
      if (kembalikanFokus) buka.focus();
    }

    const pembungkus = el('div.kalender-field', null, el('div.kalender-field__baris', null, input, !input.disabled && buka), popup);
    return pembungkus;
  }

  return { pasang, dariTeks, diMasaDepan, hariIniTeks };
})();
