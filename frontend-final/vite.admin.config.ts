import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('./admin', import.meta.url));

export default defineConfig({
  root,
  base: '/admin/',
  publicDir: fileURLToPath(new URL('./public', import.meta.url)),
  plugins: [react(), tailwindcss()],
  server: { port: 5174, strictPort: true },
  // Aset publik (font, gambar) dilayani dari root situs oleh build Astro, bukan dari /admin/.
  experimental: { renderBuiltUrl: (file, { type }) => (type === 'public' ? `/${file}` : { relative: false }) },
  build: { outDir: fileURLToPath(new URL('./dist/public/admin', import.meta.url)), emptyOutDir: true, copyPublicDir: false },
});
