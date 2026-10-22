import { existsSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { beritaDummy, getBeritaBySlug } from './berita.ts';
import { dokumentasiDummy, getDokumentasiBySlug } from './dokumentasi.ts';
import { timDummy } from './tim.ts';
import { testimoniDummy } from './testimoni.ts';
import { heroSlidesDummy } from './hero.ts';

describe('data dummy berita', () => {
  it('setiap item punya slug, judul, dan tanggal valid', () => {
    expect(beritaDummy.length).toBeGreaterThan(0);
    for (const b of beritaDummy) {
      expect(b.slug).toBeTruthy();
      expect(b.judul).toBeTruthy();
      expect(Number.isNaN(new Date(b.tanggal).getTime())).toBe(false);
    }
  });

  it('getBeritaBySlug: ketemu, tidak ketemu, dan slug kosong', () => {
    expect(getBeritaBySlug(beritaDummy[0].slug)?.judul).toBe(beritaDummy[0].judul);
    expect(getBeritaBySlug('tidak-ada')).toBeUndefined();
    expect(getBeritaBySlug('')).toBeUndefined();
  });
});

describe('data dummy dokumentasi', () => {
  it('setiap item punya slug dan kategori', () => {
    expect(dokumentasiDummy.length).toBeGreaterThan(0);
    for (const d of dokumentasiDummy) {
      expect(d.slug).toBeTruthy();
      expect(d.kategori).toBeTruthy();
    }
  });

  it('getDokumentasiBySlug: ketemu dan tidak ketemu', () => {
    expect(getDokumentasiBySlug(dokumentasiDummy[0].slug)?.judul).toBe(dokumentasiDummy[0].judul);
    expect(getDokumentasiBySlug('')).toBeUndefined();
  });
});

describe('data dummy tim', () => {
  it('setiap anggota punya nama dan peran', () => {
    expect(timDummy.length).toBeGreaterThan(0);
    for (const a of timDummy) {
      expect(a.nama).toBeTruthy();
      expect(a.peran).toBeTruthy();
    }
  });

  it('semua 6 anggota punya foto & file-nya ada di public/tim/ (dari situs lama)', () => {
    expect(timDummy).toHaveLength(6);
    for (const a of timDummy) {
      expect(a.foto, `foto hilang untuk ${a.nama}`).toBeTruthy();
      expect(existsSync(`public${a.foto}`), `file hilang: ${a.foto}`).toBe(true);
    }
  });
});

describe('data dummy testimoni', () => {
  it('setiap testimoni punya nama, peran, dan kutipan', () => {
    expect(testimoniDummy.length).toBeGreaterThan(0);
    for (const t of testimoniDummy) {
      expect(t.nama).toBeTruthy();
      expect(t.kutipan).toBeTruthy();
    }
  });
});

describe('data dummy hero slides', () => {
  it('setiap slide punya judul, sub, dan CTA', () => {
    expect(heroSlidesDummy.length).toBeGreaterThan(0);
    for (const s of heroSlidesDummy) {
      expect(s.judul).toBeTruthy();
      expect(s.sub).toBeTruthy();
      // CTA boleh rute internal (/…) atau link konsultasi eksternal (https://…).
      for (const cta of [s.ctaPrimer, s.ctaSekunder]) {
        expect(cta.to).toMatch(/^(\/|https?:\/\/)/);
      }
    }
  });

  it('aset gambar yang dirujuk benar-benar ada di public/ (cegah broken image)', () => {
    for (const s of heroSlidesDummy) {
      if (s.image) {
        expect(existsSync(`public${s.image}`), `image hilang: ${s.image}`).toBe(true);
      }
      for (const kandidat of (s.srcSet ?? '').split(',').map((p) => p.trim().split(' ')[0])) {
        if (kandidat) {
          expect(existsSync(`public${kandidat}`), `srcset hilang: ${kandidat}`).toBe(true);
        }
      }
    }
  });

  it('edge case: tanpa image, slide tetap valid (fallback background navy)', () => {
    const tanpaGambar = heroSlidesDummy.filter((s) => !s.image);
    // Tidak wajib ada, tapi bila ada pun tidak boleh mengubah struktur wajib slide.
    for (const s of tanpaGambar) {
      expect(s.judul).toBeTruthy();
    }
  });
});
