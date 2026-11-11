import { existsSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { portofolioDummy } from './portofolio.ts';

// Konten migrasi dari index.html situs lama (#portofolio) — 6 karya.
describe('data portofolio (migrasi situs lama)', () => {
  it('memuat tepat 6 karya dengan slug unik', () => {
    expect(portofolioDummy).toHaveLength(6);
    const slugs = portofolioDummy.map((k) => k.slug);
    expect(new Set(slugs).size).toBe(slugs.length);
  });

  it('setiap karya punya kategori, teknologi, judul, deskripsi, dan alt', () => {
    for (const karya of portofolioDummy) {
      expect(karya.kategori).toBeTruthy();
      expect(karya.teknologi).toBeTruthy();
      expect(karya.judul).toBeTruthy();
      expect(karya.deskripsi.length).toBeGreaterThan(10);
      expect(karya.alt).toBeTruthy();
      expect(karya.lebar).toBeGreaterThan(0);
      expect(karya.tinggi).toBeGreaterThan(0);
    }
  });

  it('semua gambar dirujuk ada di public/ (cegah broken image)', () => {
    for (const karya of portofolioDummy) {
      expect(karya.gambar).toMatch(/^\/portofolio\/[a-z0-9-]+\.(jpg|svg)$/);
      const path = join(process.cwd(), 'public', karya.gambar);
      expect(existsSync(path), `hilang: ${karya.gambar}`).toBe(true);
    }
  });

  it('edge case: karya kesehatan tetap punya teknologi yang jelas', () => {
    const kesehatan = portofolioDummy.filter((k) => k.kategori === 'Kesehatan');
    expect(kesehatan.length).toBeGreaterThanOrEqual(2);
    for (const karya of kesehatan) {
      expect(karya.teknologi).not.toBe(karya.judul);
    }
  });
});
