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

## API client

`src/api/api.ts` is the real client for the Go backend (`/api/*`, same origin, HttpOnly session cookie, `X-CSRF-Token` on writes, image upload on Save, soft lock). The old in-memory simulation panel and account picker are in `deprecated/` (not built). The temporary dev login page is `admin-masuk/` (replaced by Keycloak later).

Known limits versus production: detail pages for news and documentation are one generic skeleton (`src/pages/{berita,dokumentasi}/kerangka-seo.astro`) served by Go for every slug, which injects the SEO head and initial data; HTTP 404 status for unknown slugs comes from Go, and exists in `preview` only as a static fallback.
