// Preview hasil build: file statis dist/public; /admin/* fallback ke SPA admin; selain itu 404.html (status 404).
import { join, normalize } from 'node:path';

const root = join(import.meta.dir, '..', 'dist', 'public');
const port = Number(process.env.PORT ?? 4321);

async function file(path: string): Promise<Bun.BunFile | null> {
  const f = Bun.file(join(root, path));
  return (await f.exists()) ? f : null;
}

Bun.serve({
  port,
  async fetch(req) {
    const path = normalize(decodeURIComponent(new URL(req.url).pathname)).replace(/^([/\\])+/, '');
    if (path.startsWith('..')) return new Response('Bad request', { status: 400 });
    const found = (await file(path)) ?? (await file(join(path, 'index.html')));
    if (found) return new Response(found);
    if (/^admin([/\\]|$)/.test(path)) return new Response(Bun.file(join(root, 'admin', 'index.html')));
    return new Response(Bun.file(join(root, '404.html')), { status: 404, headers: { 'content-type': 'text/html; charset=utf-8' } });
  },
});
console.log(`Preview: http://localhost:${port}`);
