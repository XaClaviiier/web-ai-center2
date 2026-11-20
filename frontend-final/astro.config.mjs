import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import tailwindcss from '@tailwindcss/vite';

// Halaman publik statis. Admin (React + Vite) dibangun terpisah ke dist/public/admin.
export default defineConfig({
  output: 'static',
  outDir: './dist/public',
  trailingSlash: 'ignore',
  build: { format: 'directory' },
  integrations: [react()],
  vite: {
    plugins: [tailwindcss()],
    // Dev: /admin diteruskan ke server Vite admin (scripts/dev.ts) agar satu origin.
    server: { proxy: { '/admin': { target: 'http://localhost:5174', ws: true } } },
  },
});
