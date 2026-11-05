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

	"github.com/XaClaviiier/web-ai-center2/backend/internal/config"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/httpapi"
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

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           httpapi.NewRouter(httpapi.Deps{Log: log, DB: st}),
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
