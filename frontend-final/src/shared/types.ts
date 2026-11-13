export type Entity = 'berita' | 'dokumentasi' | 'testimoni' | 'tim';
export type Status = 'draft' | 'posted';
export interface Image { id?: number; key: string; urutan: number; dokumentasiId?: number }
export interface Item {
  id: number; version: number; createdAt: string; createdBy?: string; updatedBy?: string;
  status?: Status; deletedAt?: string | null; slug?: string; judul?: string; isi?: string;
  tanggalPublikasi?: string; gambar?: string | Image[]; nama?: string; posisi?: string;
  teks?: string; foto?: string | null; bidangRiset?: string; urutan?: number; dikunciOleh?: string | null;
}
export interface News extends Item { judul: string; isi: string; tanggalPublikasi: string; gambar: string; slug: string }
export interface Album extends Item { judul: string; tanggalPublikasi: string; gambar: Image[]; slug: string }
export interface Member extends Item { nama: string; posisi: string; bidangRiset: string; foto: string; urutan: number }
export interface Testimonial extends Item { nama: string; posisi: string; teks: string }
export interface Account { id: string; sub: string; username: string; nama: string; role: string | null }
export interface Page<T> { items: T[]; page: number; limit: number; total: number }
export interface Log {
  id: number; terjadiPada: string; penggunaSub: string; username: string;
  jenis: 'tambah' | 'ubah' | 'hapus'; aksi: string; modul: Entity; itemId: number; labelItem: string;
}
export type Payload = Partial<Item> & { gambarDaftar?: Image[] };
export class ApiError extends Error {
  readonly error: { code: string; message: string };
  constructor(public status: number, code: string, message: string, public retryAfter?: number) {
    super(message); this.name = 'ApiError'; this.error = { code, message };
  }
}
