/**
 * Guard sinkronisasi data: SEED_* (backend) wajib identik dengan data dummy
 * frontend (src/data/) — aturan proyek "seed = data frontend" supaya fallback
 * tanpa backend menampilkan copy yang sama dengan live.
 * Pernah drift: fase 2 mengubah layanan.ts (copy LPPM) tanpa memperbarui seed →
 * live menyajikan copy lama. Test ini mencegah kejadian berulang.
 */
import { describe, expect, test } from 'bun:test';
import { beritaDummy } from '../../../src/data/berita';
import { dokumentasiDummy } from '../../../src/data/dokumentasi';
import { heroSlidesDummy } from '../../../src/data/hero';
import { layananDummy } from '../../../src/data/layanan';
import { kursusDummy } from '../../../src/data/pelatihan';
import { profilDummy } from '../../../src/data/profil';
import { kontenInferenceDummy } from '../../../src/data/inference';
import { testimoniDummy } from '../../../src/data/testimoni';
import { timDummy } from '../../../src/data/tim';
import {
  SEED_BERITA,
  SEED_DOKUMENTASI,
  SEED_HERO,
  SEED_INFERENCE,
  SEED_KURSUS,
  SEED_LAYANAN,
  SEED_PROFIL,
  SEED_TESTIMONI,
  SEED_TIM,
} from './seed';

type Baris = Record<string, unknown>;

/** Kembalikan daftar selisih field yang ada di KEDUA sisi (copy identik = []). */
function cariSelisih(nama: string, fe: Baris[], sd: Baris[], kunci: string): string[] {
  const hasil: string[] = [];
  const mapSeed = new Map(sd.map((s) => [String(s[kunci]), s]));
  for (const item of fe) {
    const pasangan = mapSeed.get(String(item[kunci]));
    if (!pasangan) {
      hasil.push(`${nama}: kunci "${String(item[kunci])}" ada di FE, tidak ada di seed`);
      continue;
    }
    for (const k of Object.keys(item)) {
      if (!(k in pasangan)) continue;
      if (JSON.stringify(item[k]) !== JSON.stringify(pasangan[k])) {
        hasil.push(`${nama}: field "${String(item[kunci])}.${k}" FE ≠ seed`);
      }
    }
  }
  const kunciFE = new Set(fe.map((f) => String(f[kunci])));
  for (const s of sd) {
    if (!kunciFE.has(String(s[kunci]))) {
      hasil.push(`${nama}: kunci "${String(s[kunci])}" ada di seed, tidak ada di FE`);
    }
  }
  return hasil;
}

describe('sinkronisasi seed ↔ data frontend', () => {
  test('daftar entitas identik (judul/copy per kunci)', () => {
    const semua = [
      ...cariSelisih('berita', beritaDummy as unknown as Baris[], SEED_BERITA as unknown as Baris[], 'slug'),
      ...cariSelisih(
        'dokumentasi',
        dokumentasiDummy as unknown as Baris[],
        SEED_DOKUMENTASI as unknown as Baris[],
        'slug',
      ),
      ...cariSelisih('tim', timDummy as unknown as Baris[], SEED_TIM as unknown as Baris[], 'nama'),
      ...cariSelisih('layanan', layananDummy as unknown as Baris[], SEED_LAYANAN as unknown as Baris[], 'slug'),
      ...cariSelisih('hero', heroSlidesDummy as unknown as Baris[], SEED_HERO as unknown as Baris[], 'eyebrow'),
      ...cariSelisih('testimoni', testimoniDummy as unknown as Baris[], SEED_TESTIMONI as unknown as Baris[], 'nama'),
      ...cariSelisih('kursus', kursusDummy as unknown as Baris[], SEED_KURSUS as unknown as Baris[], 'kode'),
    ];
    expect(semua).toEqual([]);
  });

  test('profil: visi & misi identik (bentuk DB: visi satu baris, misi dipisah \\n)', () => {
    const visiFe = `${profilDummy.visi.judul} — ${profilDummy.visi.deskripsi}`;
    const misiFe = profilDummy.misi.join('\n');
    expect(String(SEED_PROFIL.visi)).toBe(visiFe);
    expect(String(SEED_PROFIL.misi)).toBe(misiFe);
  });

  test('inference: seluruh field identik dengan konten dummy frontend', () => {
    expect({
      judulApaItu: SEED_INFERENCE.judulApaItu,
      deskripsiApaItu: SEED_INFERENCE.deskripsiApaItu,
      kebutuhan: SEED_INFERENCE.kebutuhan,
      alur: SEED_INFERENCE.alur,
      contohIntro: SEED_INFERENCE.contohIntro,
      contoh: SEED_INFERENCE.contoh,
    }).toEqual(kontenInferenceDummy);
  });
});
