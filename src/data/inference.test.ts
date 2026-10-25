import { describe, expect, it } from 'vitest';
import {
  alurInference,
  apaItuInference,
  contohInference,
  contohInferenceIntro,
  kebutuhanInference,
  kontenInferenceDummy,
  waDiskusiInference,
} from './inference.ts';
import { portofolioDummy } from './portofolio.ts';

describe('data inference (disusun dari pemahaman layanan — situs lama tidak punya section ini)', () => {
  it('panel "Apa itu" lengkap: judul + menjelaskan inference, API, dan pendampingan tim', () => {
    expect(apaItuInference.judul).toBe('Apa itu Inference Solution?');
    expect(apaItuInference.deskripsi).toContain('Inference');
    expect(apaItuInference.deskripsi).toContain('API');
    expect(apaItuInference.deskripsi).toContain('Ubaya AI Center');
  });

  it('alur kerja: tepat 5 langkah berurutan 01–05, tanpa judul/deskripsi kosong', () => {
    expect(alurInference.map((l) => l.nomor)).toEqual(['01', '02', '03', '04', '05']);
    for (const l of alurInference) {
      expect(l.judul).toBeTruthy();
      expect(l.deskripsi).toBeTruthy();
    }
  });

  it('tanda kebutuhan: minimal 4 poin dan tidak ada duplikat', () => {
    expect(kebutuhanInference.length).toBeGreaterThanOrEqual(4);
    expect(new Set(kebutuhanInference).size).toBe(kebutuhanInference.length);
    for (const k of kebutuhanInference) expect(k).toBeTruthy();
  });

  it('contoh penerapan sinkron dengan portofolioDummy (judul tidak diduplikasi manual)', () => {
    expect(contohInference).toHaveLength(portofolioDummy.length);
    expect(contohInference.map((c) => c.judul)).toEqual(portofolioDummy.map((k) => k.judul));
    expect(contohInferenceIntro).toBeTruthy();
  });

  it('CTA WhatsApp diskusi inference ter-encode dan memuat konteks layanan', () => {
    const wa = waDiskusiInference();
    expect(wa).toMatch(/^https:\/\/wa\.me\/62\d+\?text=/);
    expect(decodeURIComponent(wa)).toContain('inference');
  });
});

describe('kontenInferenceDummy (konten statis halaman)', () => {
  it('menggabungkan konstanta panel/alur/kebutuhan/contoh jadi satu objek', () => {
    expect(kontenInferenceDummy).toEqual({
      judulApaItu: apaItuInference.judul,
      deskripsiApaItu: apaItuInference.deskripsi,
      kebutuhan: kebutuhanInference,
      alur: alurInference,
      contohIntro: contohInferenceIntro,
      contoh: contohInference,
    });
  });

  it('semua field terisi — tidak ada daftar kosong yang membuat section kosong', () => {
    expect(kontenInferenceDummy.judulApaItu.length).toBeGreaterThan(0);
    expect(kontenInferenceDummy.deskripsiApaItu.length).toBeGreaterThan(0);
    expect(kontenInferenceDummy.contohIntro.length).toBeGreaterThan(0);
    expect(kontenInferenceDummy.kebutuhan.length).toBeGreaterThan(0);
    expect(kontenInferenceDummy.alur.length).toBeGreaterThan(0);
    expect(kontenInferenceDummy.contoh.length).toBeGreaterThan(0);
  });

  it('daftar alur & contoh tidak di-copy ulang (referensi konstanta modul)', () => {
    expect(kontenInferenceDummy.alur).toBe(alurInference);
    expect(kontenInferenceDummy.kebutuhan).toBe(kebutuhanInference);
    expect(kontenInferenceDummy.contoh).toBe(contohInference);
  });
});
