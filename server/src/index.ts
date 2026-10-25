/**
 * Entry point backend: resolve config → buka DB → rakit app → listen.
 * Jalankan: `bun run dev` (watch) atau `bun run start`.
 */
import { createApp } from './app';
import { DEFAULT_ADMIN_PASSWORD, parseAdminPassword, resolveConfig } from './config';
import { createRustFsImageStorage } from './lib/objectStorage';
import { createDb, resolveDbTarget } from './db/client';
import { createRepositories } from './repositories/drizzle';

const config = resolveConfig(process.env);
const target = resolveDbTarget(config.databaseUrl);

if (parseAdminPassword(process.env.ADMIN_PASSWORD) === DEFAULT_ADMIN_PASSWORD) {
  if (process.env.NODE_ENV === 'production') {
    throw new Error('ADMIN_PASSWORD wajib diisi saat NODE_ENV=production');
  }
  console.warn(
    '[admin] ADMIN_PASSWORD belum di-set di .env — memakai fallback "admin-dev". ' +
      'Ganti di server/.env sebelum dipakai di luar lokal.',
  );
}

const storage = await createRustFsImageStorage({
  endpoint: config.rustfsEndpoint,
  accessKey: config.rustfsAccessKey,
  secretKey: config.rustfsSecretKey,
  bucket: config.rustfsBucket,
});
const handle = await createDb(target);
const app = createApp({
  repos: createRepositories(handle.db),
  corsOrigins: config.corsOrigins,
  adminPassword: config.adminPassword,
  uploadStorage: storage,
});

try {
  app.listen(config.port);
  console.log(
    `API AI Center berjalan di http://localhost:${config.port}/api/health (db: ${target.kind})`,
  );
} catch (error) {
  console.error('Gagal menjalankan server:', error);
  await handle.close();
  process.exit(1);
}

/** Tutup koneksi rapi saat proses dihentikan. */
async function shutdown(signal: string): Promise<void> {
  console.log(`\nMenerima ${signal}, menutup server…`);
  await app.stop();
  await handle.close();
  process.exit(0);
}

process.on('SIGINT', () => void shutdown('SIGINT'));
process.on('SIGTERM', () => void shutdown('SIGTERM'));
