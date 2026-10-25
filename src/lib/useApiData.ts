/**
 * Hook React untuk memuat data dari backend dengan fallback data dummy.
 *
 * Perilaku:
 * - Render pertama langsung menampilkan `fallback` (tanpa skeleton/kedip), lalu
 *   data backend menggantikan bila berhasil dimuat.
 * - Same-origin production memakai `/api`; override `VITE_API_BASE_URL` kosong
 *   atau mode test → selalu `fallback`, tidak ada request sama sekali.
 * - Gagal memuat → `ambilJson`/`ambilDaftar` sudah mengembalikan `fallback`
 *   (fallback hanya dicatat via console.warn saat DEV), jadi halaman tidak pernah kosong.
 * - `path` berubah (navigasi antar detail) → state ikut di-reset ke `fallback`
 *   baru saat render juga, supaya tidak ada konten lama yang tersisa.
 *
 * Catatan: `fallback` harus identitas stabil (konstanta modul / hasil lookup
 * dari array modul) dan `muat` harus fungsi modul (`ambilJson`/`ambilDaftar`)
 * — keduanya sengaja tidak dimasukkan ke dependensi efek.
 */
import { useEffect, useRef, useState } from 'react';
import { BASE_URL_API, ambilDaftar, ambilJson, harusFetch } from './api.ts';

type Loader<T> = (path: string, fallback: T) => Promise<T>;

function useApiData<T>(path: string, fallback: T, muat: Loader<T>): T {
  const [state, setState] = useState<{ path: string; data: T }>({ path, data: fallback });

  // "Derived state": path berubah → reset sinkron sebelum paint (anti flash
  // konten lama saat pindah halaman detail dengan komponen yang sama).
  if (state.path !== path) {
    setState({ path, data: fallback });
  }

  // Ref nilai-terbaru: efek utama hanya depend ke `path`, tapi tetap membaca
  // fallback paling baru tanpa memicu fetch ulang.
  const fallbackRef = useRef(fallback);
  useEffect(() => {
    fallbackRef.current = fallback;
  });

  useEffect(() => {
    if (!harusFetch(BASE_URL_API, import.meta.env.MODE)) return;
    let aktif = true;
    void muat(path, fallbackRef.current)
      .then((data) => {
        if (aktif) setState({ path, data });
      })
      .catch((error: unknown) => {
        // Loader seharusnya sudah menangkap error internal; ini jaring pengaman.
        console.error(`[api] gagal memuat ${path} — memakai data fallback:`, error);
      });
    return () => {
      aktif = false;
    };
    // `muat` fungsi modul (stabil); fallback lewat ref — lihat komentar atas.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [path]);

  return state.data;
}

/** Daftar dari endpoint `{ items: [...] }` — fallback saat integrasi mati/gagal. */
export function useApiDaftar<T>(path: string, fallback: T[]): T[] {
  return useApiData<T[]>(path, fallback, ambilDaftar);
}

/** Objek detail dari endpoint tunggal — fallback saat integrasi mati/gagal. */
export function useApiObjek<T>(path: string, fallback: T | undefined): T | undefined {
  return useApiData<T | undefined>(path, fallback, ambilJson);
}
