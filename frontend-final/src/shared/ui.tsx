/* Komponen bersama. Konten selalu dirender sebagai teks oleh React, tidak pernah innerHTML. */
import type { ReactNode } from 'react';
import { ApiError } from './types';

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
  kalender: 'M4 6h16v14H4zM4 10h16M8 3.5V7M16 3.5V7',
} as const;
export type IconName = keyof typeof IKON;

export function Icon({ name, size = 20 }: { name: IconName; size?: number }) {
  return (
    <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" strokeWidth={1.8} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d={IKON[name]} />
    </svg>
  );
}

/** Pesan galat aman dari respons API; 429 menyebut waktu tunggu. */
export function errorMessage(error: unknown): string {
  if (error instanceof ApiError && error.status === 429) return `Terlalu banyak permintaan. Coba lagi dalam ${error.retryAfter ?? 30} detik.`;
  return error instanceof ApiError ? error.error.message : 'Terjadi kendala. Coba lagi beberapa saat lagi.';
}

/** Keadaan kosong dan galat. Bila `error` diisi, judul dan teks mengikuti galat. */
export function State({ title, text, error, action, icon }: { title?: string; text?: string; error?: unknown; action?: ReactNode; icon?: IconName }) {
  const galat = error !== undefined;
  const judul = galat ? (error instanceof ApiError && error.status === 429 ? 'Permintaan sedang dibatasi' : 'Data belum bisa dimuat') : title;
  const teks = galat ? errorMessage(error) : text;
  return (
    <div className={galat ? 'keadaan keadaan--galat' : 'keadaan'} role={galat ? 'alert' : undefined}>
      <Icon name={icon ?? (galat ? 'peringatan' : 'kosong')} size={32} />
      <h3>{judul}</h3>
      {teks && <p>{teks}</p>}
      {action}
    </div>
  );
}

export function Loading({ count, rows = false }: { count?: number; rows?: boolean }) {
  const jumlah = count ?? (rows ? 5 : 3);
  return (
    <div className={rows ? undefined : 'grid-kartu'} aria-busy="true" aria-label="Memuat data" role="status">
      {Array.from({ length: jumlah }, (_, i) => <div key={i} className={rows ? 'kerangka kerangka--baris' : 'kerangka kerangka--kartu'} />)}
    </div>
  );
}

export function Pagination({ page, limit, total, onPage }: { page: number; limit: number; total: number; onPage: (page: number) => void }) {
  const jumlah = Math.ceil(total / limit);
  if (jumlah <= 1) return null;
  return (
    <nav className="halaman-nav" aria-label="Pagination">
      <button type="button" disabled={page <= 1} aria-label="Halaman sebelumnya" onClick={() => onPage(page - 1)}><Icon name="kiri" size={18} /></button>
      {Array.from({ length: jumlah }, (_, i) => i + 1).map((n) => (
        <button key={n} type="button" aria-current={n === page ? 'page' : undefined} aria-label={`Halaman ${n}`} onClick={() => onPage(n)}>{n}</button>
      ))}
      <button type="button" disabled={page >= jumlah} aria-label="Halaman berikutnya" onClick={() => onPage(page + 1)}><Icon name="kanan" size={18} /></button>
    </nav>
  );
}
