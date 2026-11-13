/* Format data. Tanggal berupa string DD-MM-YYYY di data maupun tampilan. */
import type { Log } from './types';

function bagianJakarta(waktu: string | number | Date, opsi: Intl.DateTimeFormatOptions, lokal: string): (tipe: Intl.DateTimeFormatPartTypes) => string {
  const bagian = new Intl.DateTimeFormat(lokal, { timeZone: 'Asia/Jakarta', ...opsi }).formatToParts(new Date(waktu));
  return (tipe) => bagian.find((b) => b.type === tipe)?.value ?? '';
}

const KATA: Record<string, string> = { tambah: 'Menambahkan', ubah: 'Mengubah', post: 'Mempublikasikan', retract: 'Menarik publikasi', restore: 'Memulihkan', hapus_permanen: 'Menghapus permanen' };
const MODUL: Record<Log['modul'], string> = { berita: 'berita', dokumentasi: 'dokumentasi', testimoni: 'testimoni', tim: 'anggota Tim' };

export const Format = {
  /** Hari kalender Jakarta untuk pembandingan rentang log (YYYY-MM-DD). */
  hariJakarta(waktu: string | number | Date): string {
    const nilai = bagianJakarta(waktu, { year: 'numeric', month: '2-digit', day: '2-digit' }, 'en-GB');
    return `${nilai('year')}-${nilai('month')}-${nilai('day')}`;
  },

  waktuJakarta(waktu: string | number | Date): string {
    const nilai = bagianJakarta(waktu, { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit', hourCycle: 'h23' }, 'id-ID');
    return `${nilai('day')} ${nilai('month')} ${nilai('year')}, ${nilai('hour')}:${nilai('minute')} WIB`;
  },

  aktivitas(log: Log): string {
    const kata = log.aksi === 'hapus' ? (log.modul === 'tim' ? 'Menghapus permanen' : 'Memindahkan ke trash') : KATA[log.aksi];
    return `${kata} ${MODUL[log.modul]} "${log.labelItem}" (ID ${log.itemId})`;
  },

  /** Ubah "DD-MM-YYYY" menjadi Date; null bila format salah atau tanggal tidak ada. */
  parseTanggal(teks: unknown): Date | null {
    const m = /^(\d{2})-(\d{2})-(\d{4})$/.exec(String(teks ?? '').trim());
    if (!m) return null;
    const [hari, bulan, tahun] = [Number(m[1]), Number(m[2]), Number(m[3])];
    const d = new Date(Date.UTC(tahun, bulan - 1, hari));
    const sama = d.getUTCFullYear() === tahun && d.getUTCMonth() === bulan - 1 && d.getUTCDate() === hari;
    return sama ? d : null;
  },

  /** Terbaru dulu menurut tanggal publikasi; createdAt sebagai pemutus seri. */
  bandingkanTerbaru(a: { tanggalPublikasi?: string; createdAt: string }, b: { tanggalPublikasi?: string; createdAt: string }): number {
    const selisih = (Format.parseTanggal(b.tanggalPublikasi)?.getTime() ?? 0) - (Format.parseTanggal(a.tanggalPublikasi)?.getTime() ?? 0);
    return selisih || Date.parse(b.createdAt) - Date.parse(a.createdAt);
  },

  /** Paragraf dipisah baris kosong. */
  paragraf(isi: unknown): string[] {
    return String(isi ?? '').split(/\n\s*\n/).map((p) => p.trim()).filter(Boolean);
  },

  /** Potongan kartu: paragraf pertama, dipotong di batas kata. */
  potongan(isi: unknown, batas = 140): string {
    const pertama = Format.paragraf(isi)[0] ?? '';
    if (pertama.length <= batas) return pertama;
    const spasi = pertama.lastIndexOf(' ', batas);
    return pertama.slice(0, spasi > 0 ? spasi : batas) + '…';
  },

  inisial(nama: unknown): string {
    return String(nama ?? '').split(/\s+/).filter(Boolean).slice(0, 2).map((k) => k[0]!.toUpperCase()).join('');
  },
};
