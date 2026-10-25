import { describe, expect, it } from 'vitest';
import { urlGambar } from './gambar.ts';
import { aset } from './basis.ts';
import { BASE_URL_API } from './api.ts';

describe('urlGambar', () => {
  it('kosong/undefined → undefined (pemanggil tampilkan placeholder)', () => {
    expect(urlGambar(undefined)).toBeUndefined();
    expect(urlGambar('')).toBeUndefined();
    expect(urlGambar('   ')).toBeUndefined();
  });

  it('URL absolut & data-URI dipakai utuh (kompatibel data lama)', () => {
    expect(urlGambar('https://contoh.test/a.png')).toBe('https://contoh.test/a.png');
    expect(urlGambar('data:image/png;base64,AAA')).toBe('data:image/png;base64,AAA');
  });

  it('path public frontend lewat aset() (basis XAMPP)', () => {
    expect(urlGambar('/tim/foto.jpg')).toBe(aset('/tim/foto.jpg'));
  });

  it('/uploads/ ditempel ke origin backend', () => {
    const hasil = urlGambar('/uploads/123-ab.png');
    expect(hasil).toBe(`${new URL(BASE_URL_API, window.location.origin).origin}/uploads/123-ab.png`);
  });
});
