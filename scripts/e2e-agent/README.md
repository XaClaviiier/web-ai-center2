# Tes e2e dengan agen (framework e2e)

Tes end-to-end memakai framework [e2e](https://github.com/e2e-inc/e2e) (`e2e@0.24.0`, mesin Playwright). Suite ini melengkapi `scripts/e2e/` (Playwright), bukan menggantikannya: `scripts/e2e/` menguji kontrak API, keamanan, dan CRUD secara rinci; folder ini menguji alur pengguna, termasuk langkah agen Claude yang memeriksa UI seperti pengguna.

| File | Isi | Butuh model? |
|---|---|---|
| `tests/auth.setup.e2e.ts` | Login lewat form Keycloak sungguhan untuk role `direktur` dan `pegawai`, lalu menyimpan sesi | Tidak |
| `tests/admin-sesi.e2e.ts` | Sesi tersimpan membuka `/admin/` langsung dengan role yang benar | Tidak |
| `tests/navigasi.e2e.ts` | Semua tautan menu utama di beranda tidak 404 | Tidak |
| `tests/alur-berita.e2e.ts` | Direktur menerbitkan berita lewat UI, pengunjung melihatnya, lalu berita ditarik ke trash; menu di layar ponsel | Ya (Claude) |

## Prasyarat

- Windows: jalankan di **WSL** (Ubuntu), sesuai syarat framework. WSL harus memakai `networkingMode=mirrored` (di `%USERPROFILE%\.wslconfig`) agar `localhost` Windows (CMS `:8080`, Keycloak `:8180`) terjangkau dari WSL.
- Node.js 24.8 atau lebih baru di dalam WSL.
- Sistem sudah berjalan: `docker compose up -d` (akar repo), Keycloak dev (`keycloak/README.md`), lalu `cd backend && go run ./cmd/api`.
- Akun Keycloak dev ber-role `direktur` dan `pegawai`.
- Untuk tes agen: kunci API Anthropic (`ANTHROPIC_API_KEY`). Model default `claude-opus-5-5`, dapat diganti lewat `E2E_MODEL`.

## Menjalankan

```bash
cd scripts/e2e-agent
cp .env.example .env        # isi KC_PASS, KC_PASS2, dan ANTHROPIC_API_KEY
npm install
export E2E_TELEMETRY_DISABLED=1   # matikan telemetri anonim framework
npx playwright install-deps chromium   # sekali, sebagai root, bila library sistem belum ada
npx e2e run                         # semua tes
npx e2e run tests/navigasi.e2e.ts   # satu file
npx e2e run --headed                # lihat browser
```

Tes tanpa langkah agen tidak memanggil model, jadi dapat dijalankan tanpa kunci API.

Untuk menyimpan rahasia dan cache di luar repo, set `E2E_ENV_FILE=<path ke .env>` dan `E2E_CACHE_DIR=<folder cache>` sebelum `npx e2e run`. Hasil run (`.e2e/`) harus berada di dalam folder proyek (aturan framework). Pindahkan setelah run bila perlu.

## Catatan

- Data uji berlabel `[E2E-AGENT]`. Tes berita memindahkannya ke trash di akhir; trash dihapus permanen oleh job berkala atau manual.
- `workers: 1`, karena batas tulis admin 30 per menit per akun.
- `.env`, `node_modules/`, dan `.e2e/` (hasil run, trace, cache replay) tidak di-commit. Cache replay (`.e2e/cache/`) membuat langkah agen dapat diputar ulang tanpa memanggil model. Commit hanya bila tim sepakat, dan tinjau isinya di PR.
- Framework masih pra-1.0, dan API dapat berubah antar rilis minor. Versi dipin di `package.json`.
