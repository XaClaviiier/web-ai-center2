import { useState, type ReactNode } from 'react';
import type { Entity, Image, Item } from '../../src/shared/types';
import type { IconName } from '../../src/shared/ui';
import { Format } from '../../src/shared/format';
import { image } from '../../src/shared/content';

// Batas judul konten (spasi dihitung, nilai di-trim). Salinan di src/api/api.ts.
const MAKS_JUDUL = 70;

export type TextName = 'judul' | 'tanggalPublikasi' | 'isi' | 'nama' | 'teks' | 'posisi' | 'bidangRiset';

export type FieldDef = { label: string; wajib: boolean; bantuan?: string } & (
  | { tipe: 'teks'; nama: TextName; maks?: number }
  | { tipe: 'tanggal' | 'area'; nama: TextName }
  | { tipe: 'angka'; nama: 'urutan' }
  | { tipe: 'gambar'; nama: 'gambar' | 'foto' }
  | { tipe: 'galeri'; nama: 'gambarDaftar' }
);

interface Column {
  label: string;
  utama?: boolean;
  isi: (b: Item) => ReactNode;
}

export interface EntityDef {
  label: string;
  nama: string;
  ikon: IconName;
  status: boolean;
  field: FieldDef[];
  kolom: Column[];
}

export function daftarGambar(b: Item | null): Image[] {
  const gambar = b?.gambar;
  return Array.isArray(gambar) ? gambar : [];
}

function Thumb({ gambar, potret = false }: { gambar: string | null | undefined; potret?: boolean }) {
  return gambar ? <img className={potret ? 'thumb thumb--potret' : 'thumb'} src={image(gambar)} alt="" /> : null;
}

/** Thumbnail bulat; tanpa foto atau gambar gagal dimuat tampil inisial. */
function Avatar({ nama, foto }: { nama: string; foto: string | null | undefined }) {
  const [gagal, setGagal] = useState(false);
  if (!foto || gagal) {
    return <span className="avatar avatar--inisial" role="img" aria-label={`Tanpa foto: ${nama}`}>{Format.inisial(nama)}</span>;
  }
  return <img className="avatar" src={image(foto)} alt={`Foto ${nama}`} loading="lazy" width={44} height={44} onError={() => setGagal(true)} />;
}

export const ENTITAS: Record<Entity, EntityDef> = {
  berita: {
    label: 'Berita', nama: 'berita', ikon: 'berita', status: true,
    field: [
      { nama: 'gambar', label: 'Gambar', tipe: 'gambar', wajib: true },
      { nama: 'judul', label: 'Judul', tipe: 'teks', wajib: true, maks: MAKS_JUDUL },
      { nama: 'tanggalPublikasi', label: 'Tanggal publikasi', tipe: 'tanggal', wajib: true },
      { nama: 'isi', label: 'Konten berita', tipe: 'area', wajib: true, bantuan: 'Pisahkan paragraf dengan satu baris kosong.' },
    ],
    kolom: [
      { label: 'Gambar', isi: (b) => <Thumb gambar={typeof b.gambar === 'string' ? b.gambar : null} /> },
      { label: 'Judul', utama: true, isi: (b) => b.judul },
      { label: 'Tanggal publikasi', isi: (b) => <span className="tanggal">{b.tanggalPublikasi}</span> },
    ],
  },
  dokumentasi: {
    label: 'Dokumentasi', nama: 'dokumentasi', ikon: 'galeri', status: true,
    field: [
      { nama: 'gambarDaftar', label: 'Gambar', tipe: 'galeri', wajib: true },
      { nama: 'judul', label: 'Judul kegiatan', tipe: 'teks', wajib: true, maks: MAKS_JUDUL },
      { nama: 'tanggalPublikasi', label: 'Tanggal publikasi', tipe: 'tanggal', wajib: true },
    ],
    kolom: [
      { label: 'Headline', isi: (d) => <Thumb gambar={daftarGambar(d)[0]?.key} /> },
      { label: 'Judul kegiatan', utama: true, isi: (d) => d.judul },
      { label: 'Tanggal publikasi', isi: (d) => <span className="tanggal">{d.tanggalPublikasi}</span> },
      { label: 'Gambar', isi: (d) => `${daftarGambar(d).length} gambar` },
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
      { label: 'Foto', isi: (t) => <Avatar nama={t.nama ?? ''} foto={t.foto} /> },
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
      { nama: 'bidangRiset', label: 'Bidang riset', tipe: 'teks', wajib: true },
      { nama: 'urutan', label: 'Urutan', tipe: 'angka', wajib: true, bantuan: 'Angka kecil tampil lebih dulu.' },
    ],
    kolom: [
      { label: 'Foto', isi: (t) => <Thumb gambar={t.foto} potret /> },
      { label: 'Nama', utama: true, isi: (t) => t.nama },
      { label: 'Posisi / jabatan', isi: (t) => t.posisi },
      { label: 'Bidang riset', isi: (t) => t.bidangRiset },
      { label: 'Urutan', isi: (t) => <span className="tanggal">{String(t.urutan)}</span> },
    ],
  },
};

export const isEntity = (nilai: string | undefined): nilai is Entity => nilai !== undefined && Object.hasOwn(ENTITAS, nilai);
