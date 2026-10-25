import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig, type Plugin } from 'vite';

// Basis deploy produksi: root domain.
const BASE = '/';

// Peta modul rute lazy → path rute yang memicunya (dipakai plugin di bawah).
// Tiap rute WAJIB terdaftar di sini supaya chunk-nya di-modulepreload oleh
// script inline index.html (tanpa peta, konten menunggu rantai dynamic import
// → jeda FCP→LCP saat audit). Abaikan prefiks yang tumpang tindih: '/layanan/'
// juga dipicu halaman detail pelatihan — itu wajar (LayananDetail ikut kecil).
const RUTE_LAZY: Record<string, string[]> = {
  'src/pages/Beranda.tsx': ['/beranda'],
  'src/pages/PelatihanDetail.tsx': ['/layanan/pelatihan/'],
  'src/pages/LayananDetail.tsx': ['/layanan/'],
  'src/pages/TentangKami.tsx': ['/tentang-kami'],
  'src/pages/Tim.tsx': ['/tim'],
  'src/pages/Berita.tsx': ['/berita'],
  'src/pages/BeritaDetail.tsx': ['/berita/'],
  'src/pages/Dokumentasi.tsx': ['/dokumentasi'],
  'src/pages/DokumentasiDetail.tsx': ['/dokumentasi/'],
  'src/pages/Admin.tsx': ['/admin'],
};

/**
 * Plugin: isi `window.__PETA_CHUNK` (peta rute → URL chunk, sudah termasuk hash
 * build) di index.html dengan MENGGANTI token `window.__PETA_CHUNK = {};`
 * (bukan menyisipkan tag baru — tag sisipan ber-`injectTo:'head'` berakhir SETELAH
 * IIFE pemakai di head, sehingga modul preload tidak pernah jalan dan konten rute
 * menunggu rantai dynamic-import (selisih FCP→LCP ±2 detik di throttling mobile)).
 * Hanya jalan saat build; dev memakai nilai default `{}` (aman — loop kosong).
 */
function petaChunkRute(base: string): Plugin {
  return {
    name: 'peta-chunk-rute',
    apply: 'build',
    transformIndexHtml(html, ctx) {
      const bundle = ctx.bundle;
      if (!bundle) return undefined;
      const peta: Record<string, string[]> = {};
      for (const [modul, rutes] of Object.entries(RUTE_LAZY)) {
        const entry = Object.values(bundle).find(
          (c) =>
            c.type === 'chunk' &&
            c.facadeModuleId !== null &&
            c.facadeModuleId.replace(/\\/g, '/').endsWith(modul),
        );
        if (!entry || entry.type !== 'chunk') continue;
        // Kumpulkan chunk entry + semua import yang dicapainya (rekursif).
        const urls = new Set<string>();
        const antre = [entry];
        const kunjung = new Set<string>();
        while (antre.length > 0) {
          const c = antre.pop();
          if (!c || kunjung.has(c.fileName)) continue;
          kunjung.add(c.fileName);
          urls.add(base + c.fileName);
          for (const nama of [...c.dynamicImports, ...c.imports]) {
            const next = bundle[nama];
            // Chunk entry (React runtime) sudah dipreload Vite sendiri; dilewati
            // supaya rekursi tidak ikut menarik seluruh rute lazy lewat dynamic
            // import milik entry (boros bandwidth — merusak skor yang dituju).
            if (next && next.type === 'chunk' && !next.isEntry) antre.push(next);
          }
        }
        for (const r of rutes) peta[r] = [...urls];
      }
      const token = 'window.__PETA_CHUNK = {};';
      if (!html.includes(token)) {
        throw new Error(
          'peta-chunk-rute: token `window.__PETA_CHUNK = {};` tidak ditemukan di index.html — ' +
            'sesuaikan index.html dengan plugin di vite.config.ts.',
        );
      }
      return html.replace(token, `window.__PETA_CHUNK = ${JSON.stringify(peta)};`);
    },
  };
}

/**
 * Plugin: INLINE css build ke dalam <style> di index.html (build saja).
 * Tag <link rel="stylesheet"> dari Vite adalah render-blocking request (audit
 * Lighthouse: ±191 ms di throttling mobile) — dengan di-inline, HTML sudah
 * membawa CSS-nya: satu permintaan lebih sedikit, byte transfer hampir sama
 * (server sudah gzip HTML & CSS), dan audit render-blocking hilang.
 * Dev tidak tersentuh (apply: 'build'); bila tag css tidak ditemukan, di-skip
 * dengan aman (mis. jumlah file css berubah nanti).
 */
function inlineCssBuild(base: string): Plugin {
  return {
    name: 'inline-css-build',
    apply: 'build',
    transformIndexHtml: {
      order: 'post',
      handler(html, ctx) {
        if (!ctx.bundle) return undefined;
        let hasil = html;
        const pola = /<link rel="stylesheet"[^>]*href="([^"]+\.css)"[^>]*>/g;
        let cocok: RegExpExecArray | null;
        let ada = false;
        while ((cocok = pola.exec(html)) !== null) {
          const fileName = cocok[1].replace(base, '').replace(/^\//, '');
          const asset = ctx.bundle[fileName];
          if (!asset || asset.type !== 'asset') continue;
          const css = (typeof asset.source === 'string' ? asset.source : Buffer.from(asset.source).toString('utf8'))
            // Penjaga: kandungan "</style" di CSS akan menutup tag lebih awal.
            .replace(/<\/style/gi, '<\\/style');
          hasil = hasil.replace(cocok[0], `<style data-vite-inline>${css}</style>`);
          ada = true;
        }
        return ada ? hasil : undefined;
      },
    },
  };
}

// https://vite.dev/config/
export default defineConfig(() => ({
  base: BASE,
  plugins: [react(), tailwindcss(), petaChunkRute(BASE), inlineCssBuild(BASE)],
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    globals: true,
    // Batasi ke frontend saja — test backend memakai bun:test (di-server/),
    // jangan dijalankan vitest.
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
    exclude: ['node_modules/**', 'server/**', 'dist/**'],
  },
}));
