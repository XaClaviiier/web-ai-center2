import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

// vitest selalu berjalan dari root proyek (root config = cwd), jadi pakai cwd
// karena import.meta.url bukan scheme file:// di environment jsdom.
const html = readFileSync(join(process.cwd(), 'index.html'), 'utf8');

describe('index.html (SEO & performa)', () => {
  it('font identitas di-self-host (public/fonts) & dipreload — tanpa CDN eksternal', () => {
    // General Sans (isi) + Cabinet Grotesk (judul) diunduh ke public/fonts
    expect(html).toContain('as="font"');
    expect(html).toContain('href="/fonts/GeneralSans-Regular.woff2"');
    expect(html).toContain('href="/fonts/CabinetGrotesk-Bold.woff2"');
    // API Fontshare dulu membalas Satoshi → General Sans gagal & jatuh ke Helvetica
    expect(html).not.toContain('fontshare.com');
    // Geist Mono di-self-host lewat index.css (public/fonts) — HTML tanpa CDN font
    expect(html).not.toContain('fonts.googleapis.com');
    expect(html).not.toContain('Geist+Mono');
    // Font lama sudah tidak dipakai
    expect(html).not.toContain('Poppins');
    expect(html).not.toContain('Open+Sans');
  });

  it('meta wajib: description, theme-color, viewport aman (tanpa zoom-disable)', () => {
    expect(html).toMatch(/<meta[^>]+name="description"[^>]+content="[^"]+"/);
    expect(html).toContain('name="theme-color"');
    expect(html).not.toMatch(/user-scalable\s*=\s*no|maximum-scale\s*=\s*1/);
  });

  it('preload LCP menunjuk aset yang ada (bukan path usang)', () => {
    expect(html).toContain('rel="preload"');
    expect(html).not.toContain('hero-1.jpg');
  });

  it('hero beranda di-preload kondisional: srcset cocok dengan atribut <img> + guard pathname', () => {
    // Preload dibuat lewat script agar hanya aktif di rute beranda (LCP hero);
    // halaman lain tidak ikut agar tidak berebut bandwidth.
    expect(html).toContain("l.rel = 'preload'");
    expect(html).toContain("l.as = 'image'");
    expect(html).toContain('hero-1-800.webp 800w, /hero-1-1600.webp 1600w');
    // imagesizes HARUS sama dengan atribut img hero (100vw) agar pilihan varian
    // identik — kalau beda, browser mengunduh dua kali (malah merusak skor).
    expect(html).toContain('imagesizes\', \'100vw');
    // Guard: hanya rute beranda / base — bukan preload buta untuk semua halaman.
    expect(html).toContain('beranda');
  });

  it('chunk rute lazy di-modulepreload memakai peta build (window.__PETA_CHUNK)', () => {
    // Peta diisi plugin `peta-chunk-rute` (URL ber-hash → tak bisa hardcode);
    // tanpa peta (mis. dev) skrip keluar tanpa aksi apa pun.
    expect(html).toContain('window.__PETA_CHUNK');
    expect(html).toContain("m.rel = 'modulepreload'");
    // Guard agar preload chunk tidak membocor ke halaman lain.
    expect(html).toContain('p.indexOf(k) === -1');
  });
});
