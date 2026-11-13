// Dev: Astro (publik, :4321) + Vite admin (:5174). Astro meneruskan /admin ke Vite agar satu origin.
const procs = [
  Bun.spawn([process.execPath, 'x', 'vite', '--config', 'vite.admin.config.ts'], { stdout: 'inherit', stderr: 'inherit' }),
  Bun.spawn([process.execPath, 'x', 'astro', 'dev'], { stdout: 'inherit', stderr: 'inherit' }),
];
const stop = () => { for (const p of procs) p.kill(); process.exit(0); };
process.on('SIGINT', stop);
process.on('SIGTERM', stop);
await Promise.race(procs.map((p) => p.exited));
stop();
