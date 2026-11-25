/* Konfigurasi tes e2e (framework e2e, github.com/e2e-inc/e2e). Jalankan di WSL; lihat README.md.
   Server CMS dan Keycloak dev HARUS sudah berjalan. Variabel dibaca dari .env di folder ini (lihat .env.example),
   atau dari berkas yang ditunjuk E2E_ENV_FILE (mis. di luar repo). */
import { existsSync } from 'node:fs';
import type { E2EConfig } from 'e2e';
import { web } from '@e2e-dev/web';
import { anthropic } from '@ai-sdk/anthropic';
import { claudeCode } from '@e2e-dev/acp';

const berkasEnv = process.env.E2E_ENV_FILE || '.env';
if (existsSync(berkasEnv)) process.loadEnvFile(berkasEnv);

export default {
  targets: [{ engine: web(), app: { url: process.env.BASE_URL ?? 'http://localhost:8080' } }],
  // Langkah agen memakai Claude: ANTHROPIC_API_KEY bila diisi, selain itu Claude Code (login `claude`, mis. paket Pro).
  // Tes tanpa langkah agen tidak memanggil model.
  agents: {
    default: process.env.ANTHROPIC_API_KEY
      ? { model: anthropic(process.env.E2E_MODEL || 'claude-opus-5-5') }
      : { executor: claudeCode({ model: process.env.E2E_CC_MODEL || 'sonnet' }) },
  },
  credentials: {
    direktur: { username: process.env.KC_USER ?? '', password: process.env.KC_PASS ?? '' },
    pegawai: { username: process.env.KC_USER2 ?? '', password: process.env.KC_PASS2 ?? '' },
  },
  workers: 1, // batas tulis admin 30/menit per akun: jangan paralel
} satisfies E2EConfig;
