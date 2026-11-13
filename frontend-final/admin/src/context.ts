import { createContext, useContext, useEffect, type RefObject } from 'react';
import { ApiError, type Account } from '../../src/shared/types';

export const PESAN_TANPA_PERUBAHAN = 'Tidak ada perubahan yang disimpan.';
export const URL_LOGIN = '/admin/masuk';

export interface ConfirmOptions {
  judul: string;
  teks: string;
  labelYa: string;
  varianYa?: 'utama' | 'bahaya';
  labelTidak?: string;
}

/** Form yang sedang terbuka; `dirty()` membandingkan isian dengan nilai awal. */
export interface FormGuard {
  dirty(): boolean;
}

export interface Admin {
  akun: Account | null;
  toast(pesan: string, galat?: boolean): void;
  konfirmasi(opsi: ConfirmOptions): Promise<boolean>;
  /** Layar "Sesi Anda sudah berakhir" (401). */
  sesiBerakhir(): void;
  /** Galat aksi: 401 ke layar sesi berakhir, lainnya toast. */
  tanganiGalat(galat: unknown): void;
  /** Render ulang rute aktif dari awal (setara `render()` prototype). */
  render(): void;
  keluar(): Promise<void>;
  formAktif: RefObject<FormGuard | null>;
}

export const AdminContext = createContext<Admin | null>(null);

export function useAdmin(): Admin {
  const admin = useContext(AdminContext);
  if (!admin) throw new Error('AdminContext belum tersedia.');
  return admin;
}

export function useTitle(judul: string): void {
  useEffect(() => {
    document.title = judul;
  }, [judul]);
}

export const is401 = (galat: unknown): boolean => galat instanceof ApiError && galat.status === 401;
