/**
 * Perakitan aplikasi: CORS (@elysia/cors, whitelist origin), penanganan error
 * eksplisit, lalu daftarkan semua route. DB disuntikkan lewat `Repositories`.
 */
import { Elysia } from "elysia";
import { cors } from "@elysia/cors";
import { registerRoutes, type AnyElysia } from "./routes";
import type { Repositories } from "./repositories/types";
import type { ImageStorage } from "./lib/upload";

export interface AppDeps {
  repos: Repositories;
  /** Origin frontend yang diizinkan CORS. */
  corsOrigins: string[];
  /** Password admin (env ADMIN_PASSWORD) — dipakai login & verifikasi token tulis. */
  adminPassword: string;
  /** Storage gambar; default filesystem hanya untuk unit test lokal. */
  uploadStorage?: ImageStorage;
}

export function createApp({
  repos,
  corsOrigins,
  adminPassword,
  uploadStorage,
}: AppDeps): AnyElysia {
  const allowedOrigins = new Set(corsOrigins);

  const app = new Elysia()
    .use(
      cors({
        origin: corsOrigins,
        methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
        allowedHeaders: ["content-type", "authorization"],
        maxAge: 86400,
        preflight: true,
      }),
    )
    .onError(({ code, error, set }) => {
      if (code === "NOT_FOUND") {
        set.status = 404;
        return { error: "Not Found" };
      }
      // Body bukan JSON valid (ParseError Elysia) → 400 eksplisit, bukan 500:
      // body rusak adalah kesalahan klien, bukan kegagalan server.
      if (code === "PARSE") {
        set.status = 400;
        return { error: "Body bukan JSON valid" };
      }
      // Log eksplisit — jangan pernah menelan error tanpa jejak.
      console.error("[api] error tidak tertangani:", error);
      set.status = 500;
      return { error: "Internal Server Error" };
    });

  return registerRoutes(app, repos, adminPassword, { storage: uploadStorage });
}
