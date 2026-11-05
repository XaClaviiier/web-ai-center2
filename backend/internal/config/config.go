// Package config membaca konfigurasi dari environment variable.
// Fungsi murni (input lookup -> output Config) supaya mudah diuji tanpa menyentuh os.Environ.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTest        = "test"

	// NamaCookieProduksi wajib di produksi (security.md §2).
	NamaCookieProduksi = "__Host-session"

	portDefault          = 8080
	namaCookieDevDefault = "session"
)

// Config adalah konfigurasi runtime backend.
type Config struct {
	AppEnv         string
	Port           int
	PublicBaseURL  string
	AllowedOrigins []string // hanya untuk pemeriksaan header Origin (CSRF), bukan CORS

	DatabaseURL string
	RedisURL    string

	SessionCookieName   string
	SessionCookieSecure bool

	// Login development (password) hanya aktif bila APP_ENV=development (security.md §1).
	DevLoginAktif bool
	AdminPassword string

	RustFSEndpoint  string
	RustFSAccessKey string
	RustFSSecretKey string
	RustFSBucket    string
}

// Lookup meniru os.LookupEnv; nilai kosong dianggap tidak di-set.
type Lookup func(key string) (string, bool)

// Load menyusun Config dari lookup dan mengumpulkan SEMUA galat validasi sekaligus.
func Load(lookup Lookup) (Config, error) {
	get := func(key string) string {
		v, ok := lookup(key)
		if !ok {
			return ""
		}
		return strings.TrimSpace(v)
	}

	var errs []error
	cfg := Config{
		AppEnv:          get("APP_ENV"),
		DatabaseURL:     get("DATABASE_URL"),
		RedisURL:        get("REDIS_URL"),
		AdminPassword:   get("ADMIN_PASSWORD"),
		RustFSEndpoint:  get("RUSTFS_ENDPOINT"),
		RustFSAccessKey: get("RUSTFS_ACCESS_KEY"),
		RustFSSecretKey: get("RUSTFS_SECRET_KEY"),
		RustFSBucket:    get("RUSTFS_BUCKET"),
	}

	switch cfg.AppEnv {
	case EnvDevelopment, EnvProduction, EnvTest:
	case "":
		errs = append(errs, errors.New("APP_ENV wajib diisi (development, production, atau test)"))
	default:
		errs = append(errs, fmt.Errorf("APP_ENV %q tidak dikenal (development, production, atau test)", cfg.AppEnv))
	}
	produksi := cfg.AppEnv == EnvProduction

	port, err := parsePort(get("PORT"))
	if err != nil {
		errs = append(errs, err)
	}
	cfg.Port = port

	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL wajib diisi"))
	}

	cfg.PublicBaseURL = strings.TrimRight(get("PUBLIC_BASE_URL"), "/")
	if cfg.PublicBaseURL == "" {
		if produksi {
			errs = append(errs, errors.New("PUBLIC_BASE_URL wajib diisi di production"))
		}
	} else if err := validasiOrigin("PUBLIC_BASE_URL", cfg.PublicBaseURL); err != nil {
		errs = append(errs, err)
	}

	origins, err := parseOrigins(get("ALLOWED_ORIGINS"))
	if err != nil {
		errs = append(errs, err)
	}
	if produksi && len(origins) == 0 {
		errs = append(errs, errors.New("ALLOWED_ORIGINS wajib diisi di production"))
	}
	cfg.AllowedOrigins = origins

	// Cookie session: di produksi wajib __Host-session + Secure dan server menolak start bila
	// diset lain; override hanya berlaku di development/test (security.md §2).
	nama, adaNama := lookup("SESSION_COOKIE_NAME")
	nama = strings.TrimSpace(nama)
	secureRaw, adaSecure := lookup("SESSION_COOKIE_SECURE")
	secureRaw = strings.TrimSpace(secureRaw)
	if produksi {
		cfg.SessionCookieName = NamaCookieProduksi
		cfg.SessionCookieSecure = true
		if adaNama && nama != "" && nama != NamaCookieProduksi {
			errs = append(errs, fmt.Errorf("SESSION_COOKIE_NAME di production harus %q", NamaCookieProduksi))
		}
		if adaSecure && secureRaw != "" {
			if b, perr := strconv.ParseBool(secureRaw); perr != nil || !b {
				errs = append(errs, errors.New("SESSION_COOKIE_SECURE di production harus true"))
			}
		}
	} else {
		cfg.SessionCookieName = namaCookieDevDefault
		if nama != "" {
			cfg.SessionCookieName = nama
		}
		if secureRaw != "" {
			b, perr := strconv.ParseBool(secureRaw)
			if perr != nil {
				errs = append(errs, fmt.Errorf("SESSION_COOKIE_SECURE %q bukan boolean", secureRaw))
			}
			cfg.SessionCookieSecure = b
		}
	}

	cfg.DevLoginAktif = cfg.AppEnv == EnvDevelopment
	if cfg.DevLoginAktif && cfg.AdminPassword == "" {
		errs = append(errs, errors.New("ADMIN_PASSWORD wajib diisi untuk login development"))
	}

	if len(errs) > 0 {
		return Config{}, fmt.Errorf("konfigurasi tidak valid: %w", errors.Join(errs...))
	}
	return cfg, nil
}

func parsePort(raw string) (int, error) {
	if raw == "" {
		return portDefault, nil
	}
	p, err := strconv.Atoi(raw)
	if err != nil || p < 1 || p > 65535 {
		return portDefault, fmt.Errorf("PORT %q tidak valid (1-65535)", raw)
	}
	return p, nil
}

func parseOrigins(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	var out []string
	var errs []error
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimRight(strings.TrimSpace(o), "/")
		if o == "" {
			continue
		}
		if err := validasiOrigin("ALLOWED_ORIGINS", o); err != nil {
			errs = append(errs, err)
			continue
		}
		out = append(out, o)
	}
	return out, errors.Join(errs...)
}

// validasiOrigin memastikan nilai berbentuk http(s)://host[:port] tanpa path.
func validasiOrigin(nama, nilai string) error {
	u, err := url.Parse(nilai)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || (u.Path != "" && u.Path != "/") {
		return fmt.Errorf("%s: %q bukan origin http(s)://host[:port]", nama, nilai)
	}
	return nil
}
