/* Konfigurasi e2e Playwright. Server HARUS sudah berjalan (APP_ENV=development, JOBS_ENABLED=false) dengan Keycloak aktif,
   dan ALLOWED_ORIGINS harus memuat BASE_URL. Login lewat form Keycloak sungguhan. Variabel lingkungan:
     BASE_URL  (default http://localhost:8080)
     KC_USER, KC_PASS            akun Keycloak ber-role direktur/pegawai (wajib untuk spec admin)
     KC_USER2, KC_PASS2          akun ber-role kedua yang berbeda, opsional: tes dua akun dilewati bila kosong
     KC_USER_TANPA_ROLE, KC_PASS_TANPA_ROLE  akun tanpa role CMS, opsional: tes akses_ditolak dilewati bila kosong
   Data uji berlabel "[E2E-PW]" dibuat dan dibersihkan oleh tes itu sendiri (hanya baris berlabel itu).
   Artefak (tangkapan layar, trace) ditulis ke output/playwright/. */
const path = require('node:path');
const { defineConfig, devices } = require('@playwright/test');

module.exports = defineConfig({
  testDir: __dirname,
  testMatch: '*.spec.cjs',
  fullyParallel: false,
  workers: 1, // batas tulis admin 30/menit per akun: jangan paralel
  retries: 0,
  timeout: 90_000,
  expect: { timeout: 10_000 },
  outputDir: path.join(__dirname, '..', '..', 'output', 'playwright', 'hasil'),
  reporter: [['list']],
  use: {
    baseURL: process.env.BASE_URL ?? 'http://localhost:8080',
    headless: true,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    locale: 'id-ID',
    timezoneId: 'Asia/Jakarta',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
});
