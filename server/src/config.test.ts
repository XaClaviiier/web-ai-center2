import { describe, expect, test } from 'bun:test';
import {
  DEFAULT_ADMIN_PASSWORD,
  DEFAULT_CORS_ORIGINS,
  DEFAULT_PORT,
  parseAdminPassword,
  parseCorsOrigins,
  parseDatabaseUrl,
  parsePort,
  resolveConfig,
} from './config';

describe('parsePort', () => {
  test('mengembalikan angka valid apa adanya', () => {
    expect(parsePort('3001')).toBe(3001);
    expect(parsePort('1')).toBe(1);
    expect(parsePort('65535')).toBe(65535);
  });

  test('edge: undefined, kosong, bukan angka, di luar rentang jatuh ke default', () => {
    expect(parsePort(undefined)).toBe(DEFAULT_PORT);
    expect(parsePort('')).toBe(DEFAULT_PORT);
    expect(parsePort('   ')).toBe(DEFAULT_PORT);
    expect(parsePort('abc')).toBe(DEFAULT_PORT);
    expect(parsePort('80.5')).toBe(DEFAULT_PORT);
    expect(parsePort('0')).toBe(DEFAULT_PORT);
    expect(parsePort('-1')).toBe(DEFAULT_PORT);
    expect(parsePort('65536')).toBe(DEFAULT_PORT);
  });
});

describe('parseCorsOrigins', () => {
  test('memecah daftar dan membuang entri kosong', () => {
    expect(parseCorsOrigins('http://a.test, http://b.test ,,http://c.test')).toEqual([
      'http://a.test',
      'http://b.test',
      'http://c.test',
    ]);
  });

  test('edge: undefined / semua kosong jatuh ke default', () => {
    expect(parseCorsOrigins(undefined)).toEqual(DEFAULT_CORS_ORIGINS);
    expect(parseCorsOrigins('')).toEqual(DEFAULT_CORS_ORIGINS);
    expect(parseCorsOrigins(' , ,')).toEqual(DEFAULT_CORS_ORIGINS);
  });

  test('default meng-cover dev, vite preview, dan XAMPP tanpa port', () => {
    // Lighthouse/preview jalan di :4173 dan deploy XAMPP di http://localhost/coding/…;
    // tanpa ini fetch API gagal CORS → error console → skor Best Practices turun.
    expect(DEFAULT_CORS_ORIGINS).toEqual([
      'http://localhost:5173',
      'http://localhost:4173',
      'http://localhost',
    ]);
  });
});

describe('parseDatabaseUrl', () => {
  test('string terisi dipertahankan (di-trim)', () => {
    expect(parseDatabaseUrl(' postgres://user:pass@localhost:5432/db ')).toBe(
      'postgres://user:pass@localhost:5432/db',
    );
  });

  test('edge: undefined / kosong → null (tanda pakai PGlite)', () => {
    expect(parseDatabaseUrl(undefined)).toBeNull();
    expect(parseDatabaseUrl('')).toBeNull();
    expect(parseDatabaseUrl('   ')).toBeNull();
  });
});

describe('parseAdminPassword', () => {
  test('nilai terisi dipertahankan (di-trim)', () => {
    expect(parseAdminPassword('  rahasia-lokal  ')).toBe('rahasia-lokal');
  });

  test('edge: undefined / kosong / whitespace jatuh ke fallback', () => {
    expect(parseAdminPassword(undefined)).toBe(DEFAULT_ADMIN_PASSWORD);
    expect(parseAdminPassword('')).toBe(DEFAULT_ADMIN_PASSWORD);
    expect(parseAdminPassword('   ')).toBe(DEFAULT_ADMIN_PASSWORD);
    expect(parseAdminPassword(undefined, 'khusus-test')).toBe('khusus-test');
  });
});

describe('resolveConfig', () => {
  test('env kosong → konfigurasi default siap dev', () => {
    expect(resolveConfig({})).toEqual({
      port: DEFAULT_PORT,
      databaseUrl: null,
      corsOrigins: DEFAULT_CORS_ORIGINS,
      adminPassword: DEFAULT_ADMIN_PASSWORD,
      rustfsEndpoint: 'http://localhost:9000',
      rustfsAccessKey: null,
      rustfsSecretKey: null,
      rustfsBucket: 'ai-center-images',
    });
  });

  test('env lengkap dipetakan sesuai kunci', () => {
    expect(
      resolveConfig({
        PORT: '4000',
        DATABASE_URL: 'postgres://u:p@localhost:5432/ai',
        CORS_ORIGIN: 'https://ai-center.ubaya.ac.id',
        ADMIN_PASSWORD: 'katasandi-lokal',
        RUSTFS_ENDPOINT: 'http://rustfs:9000',
        RUSTFS_ACCESS_KEY: 'ai-key',
        RUSTFS_SECRET_KEY: 'ai-secret',
        RUSTFS_BUCKET: 'gambar-test',
      }),
    ).toEqual({
      port: 4000,
      databaseUrl: 'postgres://u:p@localhost:5432/ai',
      corsOrigins: ['https://ai-center.ubaya.ac.id'],
      adminPassword: 'katasandi-lokal',
      rustfsEndpoint: 'http://rustfs:9000',
      rustfsAccessKey: 'ai-key',
      rustfsSecretKey: 'ai-secret',
      rustfsBucket: 'gambar-test',
    });
  });
});
