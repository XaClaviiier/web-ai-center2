/* Format data. Tanggal berupa string DD-MM-YYYY di data maupun tampilan. */
const Format = {
  /** Hari kalender Jakarta untuk pembandingan rentang log (YYYY-MM-DD). */
  hariJakarta(waktu) {
    const bagian = new Intl.DateTimeFormat('en-GB', { timeZone: 'Asia/Jakarta', year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(new Date(waktu));
    const nilai = (tipe) => bagian.find((b) => b.type === tipe).value;
    return `${nilai('year')}-${nilai('month')}-${nilai('day')}`;
  },

  waktuJakarta(waktu) {
    const bagian = new Intl.DateTimeFormat('id-ID', { timeZone: 'Asia/Jakarta', day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit', hourCycle: 'h23' }).formatToParts(new Date(waktu));
    const nilai = (tipe) => bagian.find((b) => b.type === tipe).value;
    return `${nilai('day')} ${nilai('month')} ${nilai('year')}, ${nilai('hour')}:${nilai('minute')} WIB`;
  },

  aktivitas(log) {
    const kata = { tambah: 'Menambahkan', ubah: 'Mengubah', post: 'Mempublikasikan', retract: 'Menarik publikasi', restore: 'Memulihkan', hapus: log.modul === 'tim' ? 'Menghapus permanen' : 'Memindahkan ke trash', hapus_permanen: 'Menghapus permanen' };
    const modul = { berita: 'berita', dokumentasi: 'dokumentasi', testimoni: 'testimoni', tim: 'anggota Tim' };
    return `${kata[log.aksi]} ${modul[log.modul]} "${log.label_item}" (ID ${log.item_id})`;
  },

  /** Ubah "DD-MM-YYYY" menjadi Date; null bila format salah atau tanggal tidak ada. */
  parseTanggal(teks) {
    const m = /^(\d{2})-(\d{2})-(\d{4})$/.exec(String(teks ?? '').trim());
    if (!m) return null;
    const [hari, bulan, tahun] = [Number(m[1]), Number(m[2]), Number(m[3])];
    const d = new Date(Date.UTC(tahun, bulan - 1, hari));
    const sama = d.getUTCFullYear() === tahun && d.getUTCMonth() === bulan - 1 && d.getUTCDate() === hari;
    return sama ? d : null;
  },

  /** Terbaru dulu menurut tanggal publikasi; created_at sebagai pemutus seri. */
  bandingkanTerbaru(a, b) {
    const selisih = Format.parseTanggal(b.tanggal_publikasi) - Format.parseTanggal(a.tanggal_publikasi);
    return selisih || new Date(b.created_at) - new Date(a.created_at);
  },

  /** Paragraf dipisah baris kosong. */
  paragraf(isi) {
    return String(isi ?? '').split(/\n\s*\n/).map((p) => p.trim()).filter(Boolean);
  },

  /** Potongan kartu: paragraf pertama, dipotong di batas kata. */
  potongan(isi, batas = 140) {
    const pertama = Format.paragraf(isi)[0] ?? '';
    if (pertama.length <= batas) return pertama;
    const spasi = pertama.lastIndexOf(' ', batas);
    return pertama.slice(0, spasi > 0 ? spasi : batas) + '…';
  },

  inisial(nama) {
    return String(nama ?? '').split(/\s+/).filter(Boolean).slice(0, 2).map((k) => k[0].toUpperCase()).join('');
  },
};
