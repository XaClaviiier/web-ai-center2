import { aset } from './basis.ts';
import { BASE_URL_API } from './api.ts';

/**
 * Selesaikan path gambar menjadi URL siap render.
 *
 * - URL absolut (`https://…`) & data-URI → dipakai utuh (data lama berbentuk URL).
 * - `/uploads/` (hasil `POST /api/admin/upload`) → ditempel ke origin backend,
 *   karena file disajikan server (`GET /uploads/:nama`), bukan folder public
 *   frontend. Dengan `/api` same-origin, origin memakai host browser.
 * - Selain itu → path public frontend via `aset()` (basis Vite).
 */
export function urlGambar(src?: string): string | undefined {
  if (src === undefined || src.trim() === '') return undefined;
  const bersih = src.trim();
  if (/^(https?:)?\/\//i.test(bersih) || bersih.startsWith('data:')) return bersih;
  if (bersih.startsWith('/uploads/')) {
    if (BASE_URL_API === '') return undefined;
    return `${new URL(BASE_URL_API, window.location.origin).origin}${bersih}`;
  }
  return aset(bersih);
}
