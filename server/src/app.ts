/**
 * Perakitan aplikasi: CORS (whitelist origin), penanganan error eksplisit,
 * lalu daftarkan semua route. DB disuntikkan lewat `Repositories`.
 */
import { Elysia } from 'elysia';
import { registerRoutes, type AnyElysia } from './routes';
import type { Repositories } from './repositories/types';
import type { ImageStorage } from './lib/upload';

export interface AppDeps {
  repos: Repositories;
  /** Origin frontend yang diizinkan CORS. */
  corsOrigins: string[];
  /** Password admin (env ADMIN_PASSWORD) — dipakai login & verifikasi token tulis. */
  adminPassword: string;
  /** Storage gambar; default filesystem hanya untuk unit test lokal. */
  uploadStorage?: ImageStorage;
}

export function createApp({ repos, corsOrigins, adminPassword, uploadStorage }: AppDeps): AnyElysia {
  const allowedOrigins = new Set(corsOrigins);

  const app = new Elysia()
    // CORS sederhana: echo origin hanya bila ada di whitelist.
    .onRequest(({ request, set }) => {
      const origin = request.headers.get('Origin');
      if (origin !== null && allowedOrigins.has(origin)) {
        set.headers['Access-Control-Allow-Origin'] = origin;
        set.headers['Vary'] = 'Origin';
      }
    })
    .onError(({ code, error, set }) => {
      if (code === 'NOT_FOUND') {
        set.status = 404;
        return { error: 'Not Found' };
      }
      // Body bukan JSON valid (ParseError Elysia) → 400 eksplisit, bukan 500:
      // body rusak adalah kesalahan klien, bukan kegagalan server.
      if (code === 'PARSE') {
        set.status = 400;
        return { error: 'Body bukan JSON valid' };
      }
      // Log eksplisit — jangan pernah menelan error tanpa jejak.
      console.error('[api] error tidak tertangani:', error);
      set.status = 500;
      return { error: 'Internal Server Error' };
    })
    // Preflight CORS.
    .options('/*', ({ request, set }) => {
      const origin = request.headers.get('Origin');
      set.status = 204;
      set.headers['Access-Control-Allow-Methods'] = 'GET,POST,PUT,DELETE,OPTIONS';
      set.headers['Access-Control-Allow-Headers'] = 'content-type,authorization';
      set.headers['Access-Control-Max-Age'] = '86400';
      if (origin !== null && allowedOrigins.has(origin)) {
        set.headers['Access-Control-Allow-Origin'] = origin;
        set.headers['Vary'] = 'Origin';
      }
      return null;
    });

  return registerRoutes(app, repos, adminPassword, { storage: uploadStorage });
}
