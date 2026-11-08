/* Dashboard admin: satu file dengan rute hash. Halaman hanya memanggil api.*; login lewat
   redirect (Keycloak), aplikasi tidak punya form password. */
const aplikasi = document.getElementById('aplikasi');
const URL_LOGIN = '../simulasi/keycloak.html';
const urlGambar = (key) => (String(key).startsWith('blob:') ? key : `../assets/img/${key}`);

// Batas judul konten (spasi dihitung, nilai di-trim). Salinan di simulasi/api-simulasi.js.
const MAKS_JUDUL = 70;

const ENTITAS = {
  berita: {
    label: 'Berita', nama: 'berita', ikon: 'berita', status: true,
    field: [
      { nama: 'gambar', label: 'Gambar', tipe: 'gambar', wajib: true },
      { nama: 'judul', label: 'Judul', tipe: 'teks', wajib: true, maks: MAKS_JUDUL },
      { nama: 'tanggal_publikasi', label: 'Tanggal publikasi', tipe: 'tanggal', wajib: true },
      { nama: 'isi', label: 'Konten berita', tipe: 'area', wajib: true, bantuan: 'Pisahkan paragraf dengan satu baris kosong.' },
    ],
    kolom: [
      { label: 'Gambar', isi: (b) => thumb(b.gambar) },
      { label: 'Judul', utama: true, isi: (b) => b.judul },
      { label: 'Tanggal publikasi', isi: (b) => el('span.tanggal', null, b.tanggal_publikasi) },
    ],
  },
  dokumentasi: {
    label: 'Dokumentasi', nama: 'dokumentasi', ikon: 'galeri', status: true,
    field: [
      { nama: 'gambar_daftar', label: 'Gambar', tipe: 'galeri', wajib: true },
      { nama: 'judul', label: 'Judul kegiatan', tipe: 'teks', wajib: true, maks: MAKS_JUDUL },
      { nama: 'tanggal_publikasi', label: 'Tanggal publikasi', tipe: 'tanggal', wajib: true },
    ],
    kolom: [
      { label: 'Headline', isi: (d) => thumb(d.gambar[0]?.key) },
      { label: 'Judul kegiatan', utama: true, isi: (d) => d.judul },
      { label: 'Tanggal publikasi', isi: (d) => el('span.tanggal', null, d.tanggal_publikasi) },
      { label: 'Gambar', isi: (d) => `${d.gambar.length} gambar` },
    ],
  },
  testimoni: {
    label: 'Testimoni', nama: 'testimoni', ikon: 'kutip', status: true,
    field: [
      { nama: 'foto', label: 'Foto', tipe: 'gambar', wajib: false, bantuan: 'Opsional. Tanpa foto, inisial nama yang ditampilkan.' },
      { nama: 'nama', label: 'Nama pemberi testimoni', tipe: 'teks', wajib: true },
      { nama: 'teks', label: 'Teks testimoni', tipe: 'area', wajib: true },
      { nama: 'posisi', label: 'Posisi', tipe: 'teks', wajib: true, bantuan: 'Jabatan atau posisi lainnya.' },
    ],
    kolom: [
      { label: 'Foto', isi: (t) => avatar(t.nama, t.foto) },
      { label: 'Nama', utama: true, isi: (t) => t.nama },
      { label: 'Posisi', isi: (t) => t.posisi },
      { label: 'Teks', isi: (t) => Format.potongan(t.teks, 80) },
    ],
  },
  tim: {
    label: 'Tim', nama: 'anggota tim', ikon: 'orang', status: false,
    field: [
      { nama: 'foto', label: 'Foto', tipe: 'gambar', wajib: true },
      { nama: 'posisi', label: 'Posisi / jabatan', tipe: 'teks', wajib: true },
      { nama: 'nama', label: 'Nama', tipe: 'teks', wajib: true },
      { nama: 'bidang_riset', label: 'Bidang riset', tipe: 'teks', wajib: true },
      { nama: 'urutan', label: 'Urutan', tipe: 'angka', wajib: true, bantuan: 'Angka kecil tampil lebih dulu.' },
    ],
    kolom: [
      { label: 'Foto', isi: (t) => thumb(t.foto, true) },
      { label: 'Nama', utama: true, isi: (t) => t.nama },
      { label: 'Posisi / jabatan', isi: (t) => t.posisi },
      { label: 'Bidang riset', isi: (t) => t.bidang_riset },
      { label: 'Urutan', isi: (t) => el('span.tanggal', null, String(t.urutan)) },
    ],
  },
};

let akun = null;
let berhentikanForm = null; // membersihkan heartbeat dan kunci saat keluar dari form
let formKotor = null; // fungsi: apakah form yang terbuka punya perubahan belum tersimpan
const PESAN_TANPA_PERUBAHAN = 'Tidak ada perubahan yang disimpan.';

function thumb(key, potret) {
  return key ? el(`img.thumb${potret ? '.thumb--potret' : ''}`, { src: urlGambar(key), alt: '' }) : null;
}

/** Thumbnail bulat; tanpa foto atau gambar gagal dimuat tampil inisial. */
function avatar(nama, key) {
  const pengganti = () => el('span.avatar.avatar--inisial', { role: 'img', 'aria-label': `Tanpa foto: ${nama}` }, Format.inisial(nama));
  if (!key) return pengganti();
  const gambar = el('img.avatar', { src: urlGambar(key), alt: `Foto ${nama}`, loading: 'lazy', width: 44, height: 44 });
  gambar.addEventListener('error', () => gambar.replaceWith(pengganti()));
  return gambar;
}

/* === Layar tanpa kerangka === */
function layar({ judul, teks, aksi, galat }) {
  document.title = `${judul} | Admin UBAYA AI Center`;
  aplikasi.replaceChildren(el('main.layar', null, el('div.layar__kartu', null,
    el('img.layar__logo', { src: '../assets/img/AI-Center_Logo.png', alt: 'UBAYA AI Center', height: 44 }),
    el('h1', { class: galat ? 'layar__judul layar__judul--galat' : 'layar__judul' }, judul),
    el('p.teks-sekunder', null, teks),
    aksi)));
}
const layarMasuk = () => layar({ judul: 'Masuk ke admin', teks: 'Masuk memakai akun UBAYA AI Center Anda untuk mengelola konten website.', aksi: tautanTombol('Masuk', 'utama', URL_LOGIN) });
const layar401 = () => { hentikanForm(); layar({ judul: 'Sesi Anda sudah berakhir', teks: 'Silakan masuk kembali untuk melanjutkan.', aksi: tautanTombol('Masuk kembali', 'utama', URL_LOGIN) }); };
const layar403 = () => layar({ judul: 'Anda tidak punya akses', teks: 'Akun ini belum memiliki role untuk mengelola konten. Hubungi pengelola akun Anda.', galat: true, aksi: tombol('Keluar', 'kedua', { onclick: keluar }, 'keluar') });

async function keluar() {
  const teks = `Anda perlu masuk kembali untuk mengelola konten.${formKotor?.() ? ' Perubahan yang belum disimpan akan hilang.' : ''}`;
  const ya = await konfirmasi({ judul: 'Keluar dari admin?', teks, labelYa: 'Keluar' });
  if (ya) { hentikanForm(); location.href = 'index.html'; }
}

/** Galat aksi: 401 ke layar sesi berakhir, lainnya toast. */
function tanganiGalat(g) {
  if (g?.status === 401) layar401();
  else toast(pesanGalat(g), true);
}

/* === Kerangka === */
function kerangka(aktif, isi) {
  const samping = el('nav.samping', { 'aria-label': 'Menu admin' }, Object.entries(ENTITAS).map(([kunci, e]) =>
    el('a', { href: `#/${kunci}`, 'aria-current': kunci === aktif ? 'page' : null }, ikon(e.ikon, 20), e.label)));
  aplikasi.replaceChildren(
    // Hash dipakai router, jadi skip link memindahkan fokus tanpa mengubah hash.
    el('a.lewati', { href: '#isi', onclick: (e) => { e.preventDefault(); document.getElementById('isi').focus(); } }, 'Lewati ke konten'),
    el('header.atas', null,
      el('a.atas__logo', { href: '#/berita', 'aria-label': 'Admin UBAYA AI Center' }, el('img', { src: '../assets/img/AI-Center_Logo.png', alt: '', height: 36 })),
      el('div.atas__akun', null,
        el('span.atas__nama', null, el('strong', null, akun.nama), el('span.teks-sekunder', null, akun.role)),
        tombol('Keluar', 'kedua', { class: 'tombol tombol--kedua tombol--kecil', onclick: keluar }, 'keluar'))),
    el('div.tata', null, samping, el('main.isi', { id: 'isi', tabindex: '-1' }, isi)));
}

/* === Daftar === */
function halamanDaftar(sumber, trash) {
  const e = ENTITAS[sumber];
  const wadah = el('div', { id: 'daftar' });
  document.title = `${e.label}${trash ? ' (Trash)' : ''} | Admin UBAYA AI Center`;
  kerangka(sumber, [
    el('div.kepala', null,
      el('h1', { tabindex: '-1' }, trash ? `Trash ${e.label}` : e.label),
      !trash && tautanTombol(`Tambah ${e.nama}`, 'utama', `#/${sumber}/baru`, null, 'tambah')),
    e.status && el('nav.tab', { 'aria-label': 'Tampilan daftar' },
      el('a', { href: `#/${sumber}`, 'aria-current': trash ? null : 'page' }, 'Semua'),
      el('a', { href: `#/${sumber}/trash`, 'aria-current': trash ? 'page' : null }, 'Trash')),
    trash && el('p.teks-sekunder', null, `Item di sini terhapus permanen setelah 30 hari sejak dipindahkan ke trash.`),
    wadah,
  ]);

  const muat = () => {
    wadah.replaceChildren(kerangkaBaris());
    api.adminDaftar(sumber, { trash }).then(({ items }) => {
      wadah.replaceChildren(items.length ? tabel(sumber, items, trash, muat) : keadaan(trash
        ? { judul: 'Trash kosong', teks: 'Tidak ada item di trash.' }
        : { judul: `Belum ada ${e.nama}`, teks: 'Data yang ditambahkan akan tampil di sini.', aksi: tautanTombol(`Tambah ${e.nama}`, 'utama', `#/${sumber}/baru`, null, 'tambah') }));
    }).catch((g) => {
      if (g?.status === 401) layar401();
      else wadah.replaceChildren(keadaanGalat(g, muat));
    });
  };
  muat();
}

function tabel(sumber, items, trash, muat) {
  const e = ENTITAS[sumber];
  const kolomStatus = e.status && !trash;
  const judulTabel = `${trash ? 'Trash' : 'Daftar'} ${e.nama}`;
  // Wadah bisa difokus agar tabel lebar bisa digulir dengan keyboard.
  return el('div.tabel-wadah', { role: 'region', tabindex: '0', 'aria-label': judulTabel },
    el('table.tabel', null,
      el('caption.sr-only', null, judulTabel),
      el('thead', null, el('tr', null,
        e.kolom.map((k) => el('th', { scope: 'col' }, k.label)),
        kolomStatus && el('th', { scope: 'col' }, 'Status'),
        el('th.tabel__aksi', { scope: 'col' }, 'Aksi'))),
      el('tbody', null, items.map((b) => el('tr', null,
        e.kolom.map((k) => (k.utama
          ? el('td.tabel__utama', null, k.isi(b), b.dikunci_oleh && el('span.tabel__kunci', null, ikon('kunci', 16), `Sedang diedit oleh ${b.dikunci_oleh}`))
          : el('td', null, k.isi(b)))),
        kolomStatus && el('td', null, badgeStatus(b.status)),
        el('td.tabel__aksi', null, el('div.tabel__tombol', null, aksiBaris(sumber, b, trash, muat))))))));
}

/** Tombol sesuai status. Item dikunci orang lain: hanya bisa dibaca. */
function aksiBaris(sumber, b, trash, muat) {
  const e = ENTITAS[sumber];
  const jalankan = (janji, pesan, tombolIni) => {
    tombolIni.disabled = true;
    return janji.then(() => { toast(pesan); muat(); }).catch((g) => { tombolIni.disabled = false; tanganiGalat(g); });
  };
  const namaBaris = b.judul ?? b.nama;
  const kecil = (label, varian, fungsi, atribut) => {
    const t = tombol(label, varian, { class: `tombol tombol--${varian} tombol--kecil`, 'aria-label': `${label}: ${namaBaris}`, ...atribut, onclick: () => fungsi(t) });
    return t;
  };
  const hapus = async (t) => {
    const ya = await konfirmasi(sumber === 'tim'
      ? { judul: 'Hapus anggota tim?', teks: `${b.nama} akan dihapus permanen beserta fotonya.`, labelYa: 'Hapus Permanen', varianYa: 'bahaya' }
      : { judul: `Hapus ${e.nama} ini?`, teks: 'Item dipindahkan ke trash dan terhapus permanen setelah 30 hari.', labelYa: 'Pindahkan ke trash', varianYa: 'bahaya' });
    if (ya) jalankan(api.adminHapus(sumber, b.id), sumber === 'tim' ? 'Anggota tim dihapus.' : 'Dipindahkan ke trash.', t);
  };

  if (trash) {
    return [
      kecil('Restore', 'kedua', (t) => jalankan(api.adminPulihkan(sumber, b.id), 'Dipulihkan sebagai draft.', t), { disabled: Boolean(b.dikunci_oleh) }),
      kecil('Hapus Permanen', 'bahaya', async (t) => {
        const ya = await konfirmasi({ judul: 'Hapus permanen?', teks: 'Data dan semua gambarnya hilang dan tidak bisa dikembalikan.', labelYa: 'Hapus Permanen', varianYa: 'bahaya' });
        if (ya) jalankan(api.adminHapusPermanen(sumber, b.id), 'Dihapus permanen.', t);
      }, { disabled: Boolean(b.dikunci_oleh) }),
    ];
  }
  if (e.status && b.status === 'posted') {
    return [kecil('Retract', 'kedua', (t) => jalankan(api.adminRetract(sumber, b.id), 'Dikembalikan ke draft.', t))];
  }
  const terkunci = Boolean(b.dikunci_oleh);
  const hasil = [
    tautanTombol(terkunci ? 'Lihat' : 'Edit', 'kedua', `#/${sumber}/${b.id}/edit`, { class: 'tombol tombol--kedua tombol--kecil', 'aria-label': `${terkunci ? 'Lihat' : 'Edit'}: ${namaBaris}` }),
    kecil('Hapus', 'bahaya', hapus, { disabled: terkunci }),
    e.status && kecil('Post', 'utama', (t) => jalankan(api.adminPost(sumber, b.id), 'Konten tampil di publik.', t), { disabled: terkunci }),
  ];
  return hasil;
}

/* === Validasi gambar === */
async function validasiBerkas(berkas) {
  if (berkas.size > 2 * 1024 * 1024) return 'Ukuran file melebihi 2 MB.';
  let b;
  try {
    b = new Uint8Array(await berkas.slice(0, 12).arrayBuffer());
  } catch {
    return 'File tidak bisa dibaca. Coba pilih ulang.';
  }
  const teks = String.fromCharCode(...b);
  const jpg = b[0] === 0xff && b[1] === 0xd8 && b[2] === 0xff;
  const png = b[0] === 0x89 && teks.slice(1, 4) === 'PNG';
  const webp = teks.slice(0, 4) === 'RIFF' && teks.slice(8, 12) === 'WEBP';
  return jpg || png || webp ? null : 'Tipe file harus JPG, PNG, atau WEBP.';
}
const ACCEPT = '.jpg,.jpeg,.png,.webp';
// Input file bawaan browser memakai bahasa browser, jadi tombolnya dibuat sendiri.
const labelBerkas = (id, teks) => el('label.tombol.tombol--kedua.tombol--kecil.pilih-berkas__tombol', { for: id }, ikon('tambah', 16), teks);

/* === Field === */
function buatField(f, nilaiAwal, bacaSaja) {
  const id = `f-${f.nama}`;
  const galat = el('p.field__galat', { id: `${id}-galat`, hidden: true });
  const tampilGalat = (pesan) => { galat.textContent = pesan ?? ''; galat.hidden = !pesan; };
  const judul = [f.label, f.wajib && el('span.field__wajib', { 'aria-hidden': 'true' }, ' *')];
  const bantuan = f.bantuan && el('p.field__bantuan', { id: `${id}-bantuan` }, f.bantuan);
  // Penghitung hanya untuk field ber-batas; pengumuman pembaca layar terpisah agar tidak berbunyi tiap ketukan.
  const hitung = f.maks && el('p.field__hitung', { id: `${id}-hitung` });
  const umumkan = f.maks && el('p.sr-only', { role: 'status' });
  const pesanPotong = f.maks && el('p.field__bantuan', { id: `${id}-potong`, role: 'status', hidden: true });
  const describedby = [bantuan && `${id}-bantuan`, hitung && `${id}-hitung`, pesanPotong && `${id}-potong`, `${id}-galat`].filter(Boolean).join(' ');

  if (f.tipe === 'gambar') return fieldGambar(f, nilaiAwal, bacaSaja, { id, judul, bantuan, galat, tampilGalat, describedby });
  if (f.tipe === 'galeri') return fieldGaleri(f, nilaiAwal, bacaSaja, { id, judul, galat, tampilGalat, describedby });

  // Form tambah: tanggal otomatis hari ini. Form edit: nilai tersimpan.
  const nilaiMula = nilaiAwal ?? (f.tipe === 'tanggal' ? Kalender.hariIniTeks() : '');
  const atribut = { id, class: 'input', disabled: bacaSaja, required: f.wajib, 'aria-describedby': describedby, nilai: nilaiMula };
  if (f.tipe === 'tanggal') Object.assign(atribut, { type: 'text', inputmode: 'numeric', placeholder: 'DD-MM-YYYY', maxlength: 10 });
  if (f.tipe === 'angka') Object.assign(atribut, { type: 'number', min: 1, step: 1 });
  if (f.tipe === 'teks') atribut.type = 'text';
  // Judul lama yang sudah melebihi batas tidak dipotong otomatis: maxlength dipasang hanya bila nilai awal muat.
  if (f.maks && nilaiMula.length <= f.maks) atribut.maxlength = f.maks;
  const input = f.tipe === 'area' ? el('textarea', atribut) : el('input', atribut);
  const bantuanTanggal = f.tipe === 'tanggal' && el('p.field__bantuan', null, 'Pilih dari kalender atau ketik DD-MM-YYYY. Tidak boleh setelah hari ini.');

  const kontrol = {
    node: el('div.field', null, el('label', { for: id }, judul), f.tipe === 'tanggal' ? Kalender.pasang(input, { label: f.label }) : input, hitung, umumkan, pesanPotong, bantuan, bantuanTanggal, galat),
    baca: () => (f.tipe === 'angka' ? Number(input.value) : input.value.trim()),
    validasi() {
      const v = input.value.trim();
      let pesan = null;
      if (f.wajib && !v) pesan = `${f.label} wajib diisi.`;
      // Judul lama yang melebihi batas tetap bisa disimpan selama tidak diubah.
      else if (f.maks && v.length > f.maks && v !== String(nilaiMula).trim()) pesan = `${f.label} maksimal ${f.maks} karakter (sekarang ${v.length}).`;
      else if (f.tipe === 'tanggal' && !Kalender.dariTeks(v)) pesan = 'Tanggal harus berformat DD-MM-YYYY dan benar-benar ada.';
      else if (f.tipe === 'tanggal' && Kalender.diMasaDepan(Kalender.dariTeks(v))) pesan = `${f.label} tidak boleh setelah hari ini.`;
      else if (f.tipe === 'angka' && !(Number.isInteger(Number(v)) && Number(v) >= 1)) pesan = 'Urutan harus berupa angka bulat mulai dari 1.';
      tampilGalat(pesan);
      if (pesan) input.setAttribute('aria-invalid', 'true'); else input.removeAttribute('aria-invalid');
      return pesan;
    },
    fokus: () => input.focus(),
  };
  // Pesan tanggal langsung tampil setelah mengetik atau memilih, bukan menunggu Simpan.
  if (f.tipe === 'tanggal') input.addEventListener('change', () => kontrol.validasi());
  if (f.maks) {
    const perbarui = () => {
      pesanPotong.hidden = true;
      const n = input.value.trim().length;
      hitung.textContent = `${n}/${f.maks}`;
      hitung.classList.toggle('field__hitung--lebih', n > f.maks);
      umumkan.textContent = n > f.maks ? `Melebihi batas ${f.maks} karakter.` : n >= f.maks - 10 ? `${n} dari ${f.maks} karakter.` : '';
      if (n > f.maks) kontrol.validasi();
      else if (!galat.hidden && input.getAttribute('aria-invalid') === 'true') kontrol.validasi();
    };
    input.addEventListener('input', perbarui);
    // shortcut: teks yang diseret (drop) terpotong tanpa pesan, tambahkan handler drop bila ternyata dipakai admin.
    if (atribut.maxlength) input.addEventListener('paste', (e) => {
      const sisa = f.maks - (input.value.length - (input.selectionEnd - input.selectionStart));
      const hilang = e.clipboardData.getData('text').trimEnd().length - sisa;
      if (hilang <= 0) return;
      // Ditunda agar tampil setelah event input dari paste, yang menyembunyikan pesan lama.
      setTimeout(() => {
        pesanPotong.textContent = `${hilang} karakter terakhir dipotong karena ${f.label} maksimal ${f.maks} karakter.`;
        pesanPotong.hidden = false;
      });
    });
    perbarui();
  }
  return kontrol;
}

function fieldGambar(f, nilaiAwal, bacaSaja, { id, judul, bantuan, galat, tampilGalat, describedby }) {
  let key = nilaiAwal ?? null;
  const pratinjau = el('div.unggah__pratinjau');
  const input = el('input.pilih-berkas', { id, type: 'file', accept: ACCEPT, disabled: bacaSaja, 'aria-describedby': describedby,
    onchange: async (e) => {
      const berkas = e.target.files[0];
      if (!berkas) return;
      const pesan = await validasiBerkas(berkas);
      tampilGalat(pesan);
      if (pesan) { input.value = ''; return; }
      key = URL.createObjectURL(berkas);
      gambar();
    } });
  function gambar() {
    pratinjau.replaceChildren(...[
      key ? el('img', { src: urlGambar(key), alt: 'Pratinjau gambar' }) : el('span.teks-sekunder', null, 'Belum ada gambar'),
      !f.wajib && key && !bacaSaja && tombol('Hapus foto', 'bahaya', { class: 'tombol tombol--bahaya tombol--kecil', onclick: () => { key = null; input.value = ''; gambar(); } }),
    ].filter(Boolean));
  }
  gambar();
  return {
    node: el('div.field', null, el('label', { for: id }, judul), el('div.unggah', null, pratinjau, input, !bacaSaja && labelBerkas(id, 'Pilih gambar')), bantuan, el('p.field__bantuan', null, 'JPG, PNG, atau WEBP. Maksimal 2 MB.'), galat),
    baca: () => key,
    validasi() { const pesan = f.wajib && !key ? `${f.label} wajib diisi.` : null; tampilGalat(pesan); return pesan; },
    fokus: () => input.focus(),
  };
}

function fieldGaleri(f, nilaiAwal, bacaSaja, { id, judul, galat, tampilGalat, describedby }) {
  let daftar = (nilaiAwal ?? []).map((g) => g.key);
  const wadah = el('ol.galeri-admin');
  const labelTambah = labelBerkas(id, 'Tambah gambar');
  const tambah = el('input.pilih-berkas', { id, type: 'file', accept: ACCEPT, multiple: true, disabled: bacaSaja, 'aria-describedby': describedby,
    onchange: async (e) => {
      const berkas = [...e.target.files];
      tambah.value = '';
      for (const b of berkas) {
        if (daftar.length >= 6) { tampilGalat('Maksimal 6 gambar.'); break; }
        const pesan = await validasiBerkas(b);
        tampilGalat(pesan);
        if (pesan) break;
        daftar.push(URL.createObjectURL(b));
      }
      gambar();
    } });
  function gantiDi(i) {
    const pemilih = el('input', { type: 'file', accept: ACCEPT, onchange: async (e) => {
      const b = e.target.files[0];
      if (!b) return;
      const pesan = await validasiBerkas(b);
      tampilGalat(pesan);
      if (!pesan) { daftar[i] = URL.createObjectURL(b); gambar(); }
    } });
    pemilih.click();
  }
  function gambar() {
    wadah.replaceChildren(...daftar.map((key, i) => el('li', null,
      el('img', { src: urlGambar(key), alt: `Gambar ${i + 1}` }),
      el('span', null, `Gambar ${i + 1}`, i === 0 && el('strong', null, ' · Headline')),
      !bacaSaja && el('span.galeri-admin__aksi', null,
        tombol('Ganti', 'kedua', { class: 'tombol tombol--kedua tombol--kecil', 'aria-label': `Ganti gambar ${i + 1}`, onclick: () => gantiDi(i) }),
        tombol('Hapus', 'bahaya', { class: 'tombol tombol--bahaya tombol--kecil', 'aria-label': `Hapus gambar ${i + 1}`, onclick: () => { daftar.splice(i, 1); tampilGalat(null); gambar(); } })))));
    tambah.disabled = bacaSaja || daftar.length >= 6;
    labelTambah.hidden = tambah.disabled;
  }
  gambar();
  return {
    node: el('div.field', null, el('label', { for: id }, judul), wadah, tambah, labelTambah, el('p.field__bantuan', null, '1 sampai 6 gambar. Gambar pertama menjadi headline. JPG, PNG, atau WEBP, maksimal 2 MB per file.'), galat),
    baca: () => daftar.map((key) => ({ key })),
    validasi() {
      const pesan = daftar.length < 1 ? 'Minimal 1 gambar.' : null;
      tampilGalat(pesan);
      return pesan;
    },
    fokus: () => tambah.focus(),
  };
}

/* === Form === */
function hentikanForm() {
  berhentikanForm?.();
  berhentikanForm = null;
  formKotor = null;
}

async function halamanForm(sumber, id) {
  const e = ENTITAS[sumber];
  const baru = id === null;
  const judulHalaman = baru ? `Tambah ${e.nama}` : `Edit ${e.nama}`;
  document.title = `${judulHalaman} | Admin UBAYA AI Center`;
  const kembali = `#/${sumber}`;
  kerangka(sumber, kerangkaBaris(6));
  const isi = aplikasi.querySelector('#isi');

  let data = null;
  let kunci = { milik: 'saya' };
  try {
    if (!baru) {
      data = await api.adminDetail(sumber, id);
      if (e.status && data.deleted_at) throw { status: 404, error: { message: 'Data ada di trash. Pulihkan dulu sebelum mengedit.' } };
      if (e.status && data.status !== 'draft') {
        isi.replaceChildren(keadaan({ judul: 'Konten ini sudah tayang', teks: 'Retract dulu untuk mengembalikannya ke draft, lalu edit.', aksi: tautanTombol('Kembali ke daftar', 'kedua', kembali) }));
        return;
      }
      kunci = await api.kunciAmbil(sumber, id);
    }
  } catch (g) {
    if (g?.status === 401) return layar401();
    isi.replaceChildren(keadaanGalat(g, () => halamanForm(sumber, id)));
    return;
  }

  const bacaSaja = kunci.milik === 'lain';
  const nilaiAwal = (f) => (f.tipe === 'galeri' ? data?.gambar : data?.[f.nama]);
  const field = e.field.map((f) => ({ ...f, ...buatField(f, nilaiAwal(f), bacaSaja) }));
  const simpan = tombol('Simpan', 'utama', { type: 'submit' });
  const statusKunci = el('p.field__bantuan', { role: 'status' });
  const wadahBanner = el('div');

  const form = el('form.form', { novalidate: true, onsubmit: kirim },
    wadahBanner,
    field.map((f) => f.node),
    !bacaSaja && el('div.form__aksi', null, simpan, tautanTombol('Batal', 'kedua', kembali)),
    bacaSaja && el('div.form__aksi', null, tautanTombol('Kembali ke daftar', 'kedua', kembali)));

  isi.replaceChildren(...[
    el('div.kepala', null, el('h1', { tabindex: '-1' }, bacaSaja ? `Lihat ${e.nama}` : judulHalaman)),
    !baru && !bacaSaja && statusKunci,
    form,
  ].filter(Boolean));
  isi.querySelector('h1').focus();

  if (bacaSaja) {
    wadahBanner.append(el('div.banner.banner--kunci', { role: 'status' },
      el('p.banner__teks', null, ikon('kunci', 20), `Sedang diedit oleh ${kunci.nama}. Anda hanya bisa membaca sampai kunci dilepas.`)));
    return;
  }

  // Dibandingkan dengan nilai awal, sehingga isian yang dikembalikan seperti semula tidak dianggap berubah.
  const nilaiAwalForm = JSON.stringify(field.map((f) => f.baca()));
  formKotor = () => JSON.stringify(field.map((f) => f.baca())) !== nilaiAwalForm;

  /* Heartbeat: kunci diperpanjang tiap 1 menit selama form terbuka. */
  let timer = null;
  const catatDetak = () => { statusKunci.textContent = `Data dikunci untuk Anda. Diperpanjang otomatis, terakhir ${new Date().toLocaleTimeString('id-ID')}.`; };
  async function detak() {
    try {
      await api.kunciHeartbeat();
      catatDetak();
    } catch (g) {
      if (g?.status === 401) return layar401();
      if (g?.status !== 410) return; // galat sementara: coba lagi pada detak berikutnya
      clearInterval(timer);
      simpan.disabled = true;
      statusKunci.textContent = '';
      wadahBanner.replaceChildren(el('div.banner.banner--galat', { role: 'alert' },
        el('p.banner__teks', null, ikon('peringatan', 20), 'Sesi edit berakhir. Isian Anda masih ada di form, tetapi tidak bisa disimpan. Salin yang perlu, lalu muat ulang.'),
        tombol('Muat ulang', 'kedua', { class: 'tombol tombol--kedua tombol--kecil', onclick: () => render() }, 'ulang')));
    }
  }
  if (!baru) {
    catatDetak();
    timer = setInterval(detak, 60000);
    document.addEventListener('simulasi:heartbeat', detak);
    berhentikanForm = () => {
      clearInterval(timer);
      document.removeEventListener('simulasi:heartbeat', detak);
      api.kunciLepas();
    };
  }

  async function kirim(ev) {
    ev.preventDefault();
    if (simpan.disabled) return;
    const salah = field.filter((f) => f.validasi());
    if (salah.length) { salah[0].fokus(); return; }
    if (!baru && !formKotor()) {
      hentikanForm();
      toast(PESAN_TANPA_PERUBAHAN);
      location.hash = kembali;
      return;
    }
    // Nonaktif setelah klik pertama, mencegah klik ganda.
    simpan.disabled = true;
    simpan.textContent = 'Menyimpan…';
    const nilai = Object.fromEntries(field.map((f) => [f.nama, f.baca()]));
    try {
      if (baru) await api.adminTambah(sumber, nilai);
      else await api.adminUbah(sumber, id, nilai, data.version);
      hentikanForm();
      toast(!baru ? 'Perubahan tersimpan.' : e.status ? 'Tersimpan sebagai draft.' : 'Anggota tim ditambahkan.');
      location.hash = kembali;
    } catch (g) {
      simpan.textContent = 'Simpan';
      simpan.disabled = false;
      if (g?.status === 409) {
        const muat = await konfirmasi({ judul: 'Data sudah berubah', teks: 'Pengguna lain menyimpan perubahan lebih dulu. Muat data terbaru untuk melanjutkan; isian Anda yang belum tersimpan akan hilang.', labelYa: 'Muat ulang data terbaru', labelTidak: 'Tutup' });
        if (muat) render();
      } else {
        tanganiGalat(g);
      }
    }
  }
}

/* === Router === */
function render() {
  hentikanForm();
  if (!akun) return layarMasuk();
  if (!akun.role) return layar403();
  const [sumber, a, b] = location.hash.replace(/^#\/?/, '').split('/');
  if (!ENTITAS[sumber]) { location.hash = '#/berita'; return undefined; }
  if (a === 'baru') return halamanForm(sumber, null);
  if (a === 'trash' && ENTITAS[sumber].status) return halamanDaftar(sumber, true);
  if (/^\d+$/.test(a ?? '') && b === 'edit') return halamanForm(sumber, Number(a));
  if (a) {
    kerangka(sumber, keadaan({ judul: 'Halaman tidak ditemukan', teks: 'Tautan admin ini tidak dikenal.', aksi: tautanTombol('Kembali ke daftar', 'kedua', `#/${sumber}`) }));
    return undefined;
  }
  return halamanDaftar(sumber, false);
}

/* Hash sudah berubah saat hashchange terpanggil, jadi membatalkan pindah rute berarti mengembalikan hash.
   Form tidak dirender ulang, sehingga isian dan kunci tetap utuh.
   shortcut: pengembalian hash menambah satu entri riwayat browser, cukup untuk prototype; di produksi pakai blocker router. */
let hashAktif = location.hash;
let menungguKonfirmasi = false;
async function pindahRute() {
  if (location.hash === hashAktif) return;
  if (menungguKonfirmasi) { location.hash = hashAktif; return; }
  if (formKotor?.()) {
    const tujuan = location.hash;
    location.hash = hashAktif;
    menungguKonfirmasi = true;
    const ya = await konfirmasi({ judul: 'Tinggalkan halaman ini?', teks: 'Perubahan yang belum disimpan akan hilang.', labelYa: 'Tinggalkan', labelTidak: 'Tetap di sini', varianYa: 'bahaya' });
    menungguKonfirmasi = false;
    if (!ya) return;
    formKotor = null;
    location.hash = tujuan;
    return;
  }
  if (formKotor) toast(PESAN_TANPA_PERUBAHAN); // form terbuka tanpa perubahan; setelah simpan formKotor sudah null
  hashAktif = location.hash;
  render();
}
window.addEventListener('hashchange', pindahRute);
window.addEventListener('beforeunload', (e) => { if (formKotor?.()) e.preventDefault(); });
document.addEventListener('simulasi:ubah', render);
document.addEventListener('simulasi:catat', (e) => toast(e.detail));
akun = api.masuk(new URLSearchParams(location.search).get('akun'));
render();
