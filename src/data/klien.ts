// Data statis frontend yang disiapkan untuk section Our Client.
// Section masih nonaktif; data ini bukan fallback API atau konten CMS.
export interface Klien {
  /** Id baris DB — hanya ada pada data backend. */
  id?: number;
  nama: string;
  bidang: string;
  /** Urutan tampil — hanya ada pada data backend. */
  urutan?: number;
}

export const klienDummy: Klien[] = [
  { nama: 'FTB', bidang: 'Mitra Fakultas (placeholder)' },
  { nama: 'CAW', bidang: 'Mitra (placeholder)' },
  { nama: 'Ubaya', bidang: 'Universitas Surabaya' },
];
