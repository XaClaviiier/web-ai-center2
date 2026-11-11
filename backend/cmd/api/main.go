// Command api menjalankan server HTTP backend AI Center.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/auth"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/cache"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/config"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/httpapi"
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

	handler := httpapi.NewRouter(httpapi.Deps{
		Log:             log,
		DB:              st,
		Konten:          st,
		Media:           gambar,
		UploadMaxBytes:  cfg.UploadMaxBytes,
		Cache:           &cache.Cache{C: rdb, TTL: cfg.CacheTTL},
		PublikLimiter:   ratelimit.Redis{C: rdb, Jendela: time.Minute},
		BatasPublik:     cfg.RateLimitPublicMissPerMinute,
		BatasAdminBaca:  cfg.RateLimitAdminReadPerMinute,
		BatasAdminTulis: cfg.RateLimitAdminWritePerMinute,
		ProxyTepercaya:  cfg.TrustedProxies,
		Admin: httpapi.DepsAdmin{
			Berita:      admin.BeritaLayanan{Repo: st.AdminBerita(), Media: gambar, Log: log},
			Dokumentasi: admin.DokumentasiLayanan{Repo: st.AdminDokumentasi(), Media: gambar, Log: log},
			Testimoni:   admin.TestimoniLayanan{Repo: st.AdminTestimoni(), Media: gambar, Log: log},
			Tim:         admin.TimLayanan{Repo: st.AdminTim(), Media: gambar, Log: log},
		},
		Auth: httpapi.DepsAuth{
			Sesi:           auth.NewManager(auth.RedisStore{C: rdb}, cfg.SessionIdleTTL, cfg.SessionAbsoluteTTL),
			Limiter:        ratelimit.Redis{C: rdb, Jendela: time.Minute},
			DevLoginAktif:  cfg.DevLoginAktif,
			AdminPassword:  cfg.AdminPassword,
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
	return srv.Shutdown(tutupCtx)
}
