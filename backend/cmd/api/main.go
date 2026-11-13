// Command api menjalankan server HTTP backend AI Center.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/auth"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/cache"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/config"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/httpapi"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/lock"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/ratelimit"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store"
)

func main() {
	if err := jalankan(); err != nil {
		fmt.Fprintln(os.Stderr, "gagal:", err)
		os.Exit(1)
	}
}

func jalankan() error {
	// .env hanya untuk development; variabel yang sudah ada di environment tidak ditimpa.
	if err := config.LoadDotEnv(".env"); err != nil {
		return fmt.Errorf("membaca .env: %w", err)
	}
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if cfg.AppEnv == config.EnvProduction {
		log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	slog.SetDefault(log)
	peringatanKonfigurasi(cfg, log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bukaCtx, batal := context.WithTimeout(ctx, 10*time.Second)
	st, err := store.Open(bukaCtx, cfg.DatabaseURL)
	batal()
	if err != nil {
		return err
	}
	defer st.Close()

	rdbOpt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("REDIS_URL tidak valid: %w", err)
	}
	rdb := redis.NewClient(rdbOpt)
	defer rdb.Close()
	pingCtx, batalPing := context.WithTimeout(ctx, 5*time.Second)
	err = rdb.Ping(pingCtx).Err()
	batalPing()
	if err != nil {
		return fmt.Errorf("Redis tidak terjangkau: %w", err)
	}

	gambar := media.NewS3(media.S3Config{
		Endpoint:  cfg.RustFSEndpoint,
		AccessKey: cfg.RustFSAccessKey,
		SecretKey: cfg.RustFSSecretKey,
		Bucket:    cfg.RustFSBucket,
	})
	// Bucket dibuat otomatis hanya di luar production (kemudahan dev); di production harus sudah ada.
	bktCtx, batalBkt := context.WithTimeout(ctx, 10*time.Second)
	err = gambar.PastikanBucket(bktCtx, cfg.AppEnv != config.EnvProduction)
	batalBkt()
	if err != nil {
		return fmt.Errorf("storage gambar: %w", err)
	}

	seoDeps, err := muatSEO(cfg, log)
	if err != nil {
		return err
	}
	seoDeps.Sitemap = st

	layananBerita := admin.BeritaLayanan{Repo: st.AdminBerita(), Media: gambar, Log: log}
	layananDokumentasi := admin.DokumentasiLayanan{Repo: st.AdminDokumentasi(), Media: gambar, Log: log}
	layananTestimoni := admin.TestimoniLayanan{Repo: st.AdminTestimoni(), Media: gambar, Log: log}
	layananTim := admin.TimLayanan{Repo: st.AdminTim(), Media: gambar, Log: log}

	handler := httpapi.NewRouter(httpapi.Deps{
		SEO:                seoDeps,
		HSTS:               cfg.AppEnv == config.EnvProduction,
		Static:             &httpapi.DepsStatic{Dir: cfg.StaticDir, DevLoginAktif: cfg.DevLoginAktif, Halaman404: seoDeps.Halaman404, NamaSitus: cfg.SiteName},
		Log:                log,
		DB:                 st,
		Konten:             st,
		Media:              gambar,
		UploadMaxBytes:     cfg.UploadMaxBytes,
		Cache:              &cache.Cache{C: rdb, TTL: cfg.CacheTTL},
		PublikLimiter:      ratelimit.Redis{C: rdb, Jendela: time.Minute},
		BatasPublik:        cfg.RateLimitPublicMissPerMinute,
		BatasAdminBaca:     cfg.RateLimitAdminReadPerMinute,
		BatasAdminTulis:    cfg.RateLimitAdminWritePerMinute,
		ProxyTepercaya:     cfg.TrustedProxies,
		Lock:               &lock.Lock{C: rdb, TTL: cfg.LockTTL},
		LockHeartbeatDetik: int(cfg.LockHeartbeat.Seconds()),
		Admin:              httpapi.DepsAdmin{Berita: layananBerita, Dokumentasi: layananDokumentasi, Testimoni: layananTestimoni, Tim: layananTim, Log: admin.LogLayanan{Repo: st.AdminLog()}},
		Auth: httpapi.DepsAuth{
			Sesi:           auth.NewManager(auth.RedisStore{C: rdb}, cfg.SessionIdleTTL, cfg.SessionAbsoluteTTL),
			Limiter:        ratelimit.Redis{C: rdb, Jendela: time.Minute},
			DevLoginAktif:  cfg.DevLoginAktif,
			AdminPassword:  cfg.AdminPassword,
			AdminPassword2: cfg.AdminPassword2,
			BatasLogin:     cfg.RateLimitLoginPerMinute,
			CookieName:     cfg.SessionCookieName,
			CookieSecure:   cfg.SessionCookieSecure,
			CookieMaxAge:   cfg.SessionAbsoluteTTL,
			AllowedOrigins: cfg.AllowedOrigins,
		},
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Job berkala berhenti bersama server (ctx dibatalkan oleh sinyal) dan ditunggu sebelum keluar.
	var selesaiJob sync.WaitGroup
	if cfg.JobsEnabled {
		job := bangunJob(cfg, log, rdb, st, gambar, layananBerita, layananDokumentasi, layananTestimoni)
		selesaiJob.Add(1)
		go func() {
			defer selesaiJob.Done()
			job.Jalankan(ctx)
		}()
	} else {
		log.Info("job berkala nonaktif (JOBS_ENABLED=false)")
	}

	galat := make(chan error, 1)
	go func() {
		log.Info("server berjalan", "addr", srv.Addr, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			galat <- err
		}
	}()

	select {
	case err := <-galat:
		return err
	case <-ctx.Done():
		log.Info("menerima sinyal berhenti, menutup server")
	}

	tutupCtx, selesai := context.WithTimeout(context.Background(), 10*time.Second)
	defer selesai()
	err = srv.Shutdown(tutupCtx)
	stop() // pastikan job menerima pembatalan juga bila keluar lewat jalur galat
	selesaiJob.Wait()
	return err
}

// peringatanKonfigurasi mencatat WARN untuk kombinasi konfigurasi yang berbahaya tetapi sah secara teknis.
// Tidak menggagalkan start: keputusan nilai ada pada operator.
func peringatanKonfigurasi(cfg config.Config, log *slog.Logger) {
	if cfg.AppEnv == config.EnvProduction && len(cfg.TrustedProxies) == 0 {
		log.Warn("TRUSTED_PROXIES kosong di production: bila server berada di belakang proxy, semua klien terlihat berasal dari IP proxy dan batas laju per-IP dipakai bersama; isi dengan CIDR proxy tepercaya")
	}
	if cfg.AppEnv != config.EnvProduction && !hostLokal(cfg.PublicBaseURL) {
		log.Warn("APP_ENV bukan production padahal PUBLIC_BASE_URL bukan host lokal: login development dan pengaturan longgar mungkin aktif di server publik", "app_env", cfg.AppEnv, "public_base_url", cfg.PublicBaseURL)
	}
}

// hostLokal melaporkan apakah URL menunjuk ke mesin lokal (localhost, 127.0.0.1, ::1) atau kosong.
func hostLokal(rawURL string) bool {
	if rawURL == "" {
		return true
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	switch u.Hostname() {
	case "localhost", "127.0.0.1", "::1", "":
		return true
	}
	return false
}
