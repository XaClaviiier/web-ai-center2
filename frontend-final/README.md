# Frontend final — UBAYA AI Center

Implementation of `prototype/` with the stack agreed in `docs/stack.md`: Astro + TypeScript (static public pages), React + Vite (admin at `/admin`), Tailwind CSS v4 (no preflight, prototype CSS kept), shadcn/ui primitives, GSAP/motion, Embla, TanStack Query, Bun.

## Commands

Bun is vendored locally at `.tooling/package/bin/bun.exe` (not on PATH).

```sh
./.tooling/package/bin/bun.exe install
./.tooling/package/bin/bun.exe run dev      # Astro :4321, admin via /admin (Vite :5174 proxied)
./.tooling/package/bin/bun.exe run check    # astro check + admin tsc
./.tooling/package/bin/bun.exe run build    # dist/public (public) + dist/public/admin
./.tooling/package/bin/bun.exe run preview  # serve dist/public, /admin/* SPA fallback, 404.html
```

## Simulation (temporary)

`src/simulation/` replaces the backend until integration: in-memory API (`api.ts`, 300 ms delay, reset on reload or logout), the Simulation panel (loading/empty/error/429/401/409/heartbeat/lock states), and simulated account selection instead of Keycloak. No cookie, session, CSRF, or persistence. Remove `src/simulation/` and swap `api` for real `/api/*` calls when the Go backend is integrated.

Known limits versus production: detail pages are pre-generated only for seed slugs (unknown slugs and SEO injection are handled by Go later); HTTP 404 status exists only in `preview`.
