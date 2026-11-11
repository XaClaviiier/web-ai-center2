// Package config membaca konfigurasi dari environment variable.
// Fungsi murni (input lookup -> output Config) supaya mudah diuji tanpa menyentuh os.Environ.
package config

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTest        = "test"

	// NamaCookieProduksi wajib di produksi (security.md §2).
	NamaCookieProduksi = "__Host-session"

	portDefault          = 8080
	namaCookieDevDefault = "session"

	idleTTLDefaultDetik       = 3600  // 1 jam
	absoluteTTLDefaultDetik   = 28800 // 8 jam
	batasLoginDefault         = 10
	jobIntervalDefaultMenit   = 60
	jobLockTTLDefaultDetik    = 1800
	trashRetensiDefaultHari   = 30 // docs/requirement.md bagian 5
	orphanUmurDefaultJam      = 24 // docs/database.md bagian 5
	orphanMaksHapusDefault    = 200
	staticDirDefault          = "../web/dist"
	siteNameDefault           = "AI Center"
	lockTTLDefaultDetik       = 300  // 5 menit (docs/database.md §4)
	lockHeartbeatDefaultDetik = 60   // 1 menit
	batasAdminBacaDefault     = 180  // docs/security.md §6: per user ID
	batasAdminTulisDefault    = 30   // termasuk heartbeat lock
	batasPublikMissDefault    = 300  // docs/stack.md §6: baca publik per IP, hanya saat cache miss
	cacheTTLDefaultDetik      = 1200 // 20 menit (docs/database.md §7)

	uploadMaxMBDefault = 2  // docs/database.md §3.1
	uploadMaxMBBatas   = 20 // pagar pengaman terhadap salah ketik
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
	// Idle diperpanjang tiap request; batas mutlak tidak pernah dilewati (security.md §2).
	SessionIdleTTL     time.Duration
	SessionAbsoluteTTL time.Duration

	// Batas request ke endpoint login per IP per menit (security.md §6).
	RateLimitLoginPerMinute int
	// Batas request baca publik per IP per menit; hanya dihitung saat cache miss.
	RateLimitPublicMissPerMinute int

	// Batas request admin per user ID per menit (docs/security.md §6).
	RateLimitAdminReadPerMinute  int
	RateLimitAdminWritePerMinute int

	// Proxy tepercaya (RP-03): X-Forwarded-For hanya dipercaya bila RemoteAddr ada di daftar ini. Kosong = diabaikan.
	TrustedProxies []netip.Prefix

	// Job berkala (hapus permanen trash kedaluwarsa dan pembersihan file yatim). Default: aktif hanya di production,
	// supaya menjalankan server di laptop tidak otomatis membersihkan bucket dev.
	JobsEnabled        bool
	JobsDryRun         bool          // hanya mencatat apa yang akan dihapus
	JobInterval        time.Duration // jarak antar siklus
	JobLockTTL         time.Duration // masa hidup kunci Redis lock:purge
	TrashRetentionDays int           // lama item di trash sebelum dihapus permanen
	OrphanMinAge       time.Duration // umur minimum objek sebelum dianggap yatim
	OrphanMaxDelete    int           // batas hapus file yatim per siklus

	// Folder hasil build Astro: Go membaca kerangka detail dan 404.html dari sini (RP-12).
	StaticDir string
	// Nama situs untuk title, og:site_name, dan JSON-LD.
	SiteName string

	// Masa hidup entri cache konten publik di Redis.
	CacheTTL time.Duration

	// Login development (password) hanya aktif bila APP_ENV=development (security.md §1).
	DevLoginAktif bool
	AdminPassword string
	// Akun admin development kedua (opsional) untuk menguji dua admin yang saling mengunci; kosong = nonaktif.
	AdminPassword2 string

	// Soft lock (docs/database.md §4): masa hidup kunci dan interval heartbeat yang disarankan ke klien.
	LockTTL       time.Duration
	LockHeartbeat time.Duration

	RustFSEndpoint  string
	RustFSAccessKey string
	RustFSSecretKey string
	RustFSBucket    string

	// Batas ukuran satu gambar unggahan (docs: 2 MB; bisa disesuaikan lewat UPLOAD_MAX_MB).
	UploadMaxBytes int64
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
		AdminPassword2:  get("ADMIN_PASSWORD_2"),
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
	if cfg.RedisURL == "" {
		errs = append(errs, errors.New("REDIS_URL wajib diisi"))
	} else if u, perr := url.Parse(cfg.RedisURL); perr != nil || (u.Scheme != "redis" && u.Scheme != "rediss") || u.Host == "" {
		errs = append(errs, errors.New("REDIS_URL harus berbentuk redis://[:password@]host:port[/db]"))
	}

	for _, w := range []struct{ nama, nilai string }{
		{"RUSTFS_ENDPOINT", cfg.RustFSEndpoint},
		{"RUSTFS_ACCESS_KEY", cfg.RustFSAccessKey},
		{"RUSTFS_SECRET_KEY", cfg.RustFSSecretKey},
		{"RUSTFS_BUCKET", cfg.RustFSBucket},
	} {
		if w.nilai == "" {
			errs = append(errs, fmt.Errorf("%s wajib diisi", w.nama))
		}
	}
	if cfg.RustFSEndpoint != "" {
		if err := validasiOrigin("RUSTFS_ENDPOINT", strings.TrimRight(cfg.RustFSEndpoint, "/")); err != nil {
			errs = append(errs, err)
		}
	}
	uploadMB, err := parsePositifInt("UPLOAD_MAX_MB", get("UPLOAD_MAX_MB"), uploadMaxMBDefault)
	if err != nil {
		errs = append(errs, err)
	} else if uploadMB > uploadMaxMBBatas {
		errs = append(errs, fmt.Errorf("UPLOAD_MAX_MB %d terlalu besar (maksimum %d)", uploadMB, uploadMaxMBBatas))
	}
	cfg.UploadMaxBytes = int64(uploadMB) << 20

	idle, err := parsePositifInt("SESSION_IDLE_TTL_SECONDS", get("SESSION_IDLE_TTL_SECONDS"), idleTTLDefaultDetik)
	if err != nil {
		errs = append(errs, err)
	}
	mutlak, err := parsePositifInt("SESSION_ABSOLUTE_TTL_SECONDS", get("SESSION_ABSOLUTE_TTL_SECONDS"), absoluteTTLDefaultDetik)
	if err != nil {
		errs = append(errs, err)
	}
	if idle > mutlak {
		errs = append(errs, errors.New("SESSION_IDLE_TTL_SECONDS tidak boleh lebih besar dari SESSION_ABSOLUTE_TTL_SECONDS"))
	}
	cfg.SessionIdleTTL = time.Duration(idle) * time.Second
	cfg.SessionAbsoluteTTL = time.Duration(mutlak) * time.Second

	batasLogin, err := parsePositifInt("RATE_LIMIT_LOGIN_PER_MINUTE", get("RATE_LIMIT_LOGIN_PER_MINUTE"), batasLoginDefault)
	if err != nil {
		errs = append(errs, err)
	}
	cfg.RateLimitLoginPerMinute = batasLogin

	batasMiss, err := parsePositifInt("RATE_LIMIT_PUBLIC_MISS_PER_MINUTE", get("RATE_LIMIT_PUBLIC_MISS_PER_MINUTE"), batasPublikMissDefault)
	if err != nil {
		errs = append(errs, err)
	}
	cfg.RateLimitPublicMissPerMinute = batasMiss

	baca, err := parsePositifInt("RATE_LIMIT_ADMIN_READ_PER_MINUTE", get("RATE_LIMIT_ADMIN_READ_PER_MINUTE"), batasAdminBacaDefault)
	if err != nil {
		errs = append(errs, err)
	}
	cfg.RateLimitAdminReadPerMinute = baca
	tulis, err := parsePositifInt("RATE_LIMIT_ADMIN_WRITE_PER_MINUTE", get("RATE_LIMIT_ADMIN_WRITE_PER_MINUTE"), batasAdminTulisDefault)
	if err != nil {
		errs = append(errs, err)
	}
	cfg.RateLimitAdminWritePerMinute = tulis

	proxies, err := parseProxies(get("TRUSTED_PROXIES"))
	if err != nil {
		errs = append(errs, err)
	}
	cfg.TrustedProxies = proxies

	cacheTTL, err := parsePositifInt("CACHE_TTL_SECONDS", get("CACHE_TTL_SECONDS"), cacheTTLDefaultDetik)
	if err != nil {
		errs = append(errs, err)
	}
	cfg.CacheTTL = time.Duration(cacheTTL) * time.Second

	cfg.StaticDir = get("STATIC_DIR")
	if cfg.StaticDir == "" {
		cfg.StaticDir = staticDirDefault
	}
	cfg.SiteName = get("SITE_NAME")
	if cfg.SiteName == "" {
		cfg.SiteName = siteNameDefault
	}

	cfg.PublicBaseURL = strings.TrimRight(get("PUBLIC_BASE_URL"), "/")
	if cfg.PublicBaseURL == "" {
		if produksi {
			errs = append(errs, errors.New("PUBLIC_BASE_URL wajib diisi di production"))
		} else if cfg.Port > 0 {
			// Development/test: canonical dan OG tetap valid saat dicoba lokal.
			cfg.PublicBaseURL = fmt.Sprintf("http://localhost:%d", cfg.Port)
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

	lockTTL, err := parsePositifInt("LOCK_TTL_SECONDS", get("LOCK_TTL_SECONDS"), lockTTLDefaultDetik)
	if err != nil {
		errs = append(errs, err)
	}
	lockHB, err := parsePositifInt("LOCK_HEARTBEAT_SECONDS", get("LOCK_HEARTBEAT_SECONDS"), lockHeartbeatDefaultDetik)
	if err != nil {
		errs = append(errs, err)
	}
	if lockHB >= lockTTL {
		errs = append(errs, errors.New("LOCK_HEARTBEAT_SECONDS harus lebih kecil dari LOCK_TTL_SECONDS"))
	}
	cfg.LockTTL = time.Duration(lockTTL) * time.Second
	cfg.LockHeartbeat = time.Duration(lockHB) * time.Second

	cfg.JobsEnabled = produksi
	if raw := get("JOBS_ENABLED"); raw != "" {
		b, perr := strconv.ParseBool(raw)
		if perr != nil {
			errs = append(errs, fmt.Errorf("JOBS_ENABLED %q bukan boolean", raw))
		}
		cfg.JobsEnabled = b
	}
	if raw := get("JOBS_DRY_RUN"); raw != "" {
		b, perr := strconv.ParseBool(raw)
		if perr != nil {
			errs = append(errs, fmt.Errorf("JOBS_DRY_RUN %q bukan boolean", raw))
		}
		cfg.JobsDryRun = b
	}
	jobMenit, err := parsePositifInt("JOB_INTERVAL_MINUTES", get("JOB_INTERVAL_MINUTES"), jobIntervalDefaultMenit)
	if err != nil {
		errs = append(errs, err)
	}
	jobTTL, err := parsePositifInt("JOB_LOCK_TTL_SECONDS", get("JOB_LOCK_TTL_SECONDS"), jobLockTTLDefaultDetik)
	if err != nil {
		errs = append(errs, err)
	}
	retensi, err := parsePositifInt("TRASH_RETENTION_DAYS", get("TRASH_RETENTION_DAYS"), trashRetensiDefaultHari)
	if err != nil {
		errs = append(errs, err)
	}
	umurYatim, err := parsePositifInt("ORPHAN_MIN_AGE_HOURS", get("ORPHAN_MIN_AGE_HOURS"), orphanUmurDefaultJam)
	if err != nil {
		errs = append(errs, err)
	}
	maksYatim, err := parsePositifInt("ORPHAN_MAX_DELETE_PER_RUN", get("ORPHAN_MAX_DELETE_PER_RUN"), orphanMaksHapusDefault)
	if err != nil {
		errs = append(errs, err)
	}
	cfg.JobInterval = time.Duration(jobMenit) * time.Minute
	cfg.JobLockTTL = time.Duration(jobTTL) * time.Second
	cfg.TrashRetentionDays = retensi
	cfg.OrphanMinAge = time.Duration(umurYatim) * time.Hour
	cfg.OrphanMaxDelete = maksYatim

	cfg.DevLoginAktif = cfg.AppEnv == EnvDevelopment
	if cfg.AdminPassword2 != "" {
		if !cfg.DevLoginAktif {
			errs = append(errs, errors.New("ADMIN_PASSWORD_2 hanya boleh diisi di development"))
		} else if cfg.AdminPassword2 == cfg.AdminPassword {
			errs = append(errs, errors.New("ADMIN_PASSWORD_2 harus berbeda dari ADMIN_PASSWORD"))
		}
	}
	if cfg.DevLoginAktif && cfg.AdminPassword == "" {
		errs = append(errs, errors.New("ADMIN_PASSWORD wajib diisi untuk login development"))
	}

	if len(errs) > 0 {
		return Config{}, fmt.Errorf("konfigurasi tidak valid: %w", errors.Join(errs...))
	}
	return cfg, nil
}

func parsePositifInt(nama, raw string, def int) (int, error) {
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return def, fmt.Errorf("%s %q harus bilangan bulat minimal 1", nama, raw)
	}
	return n, nil
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

// parseProxies membaca daftar CIDR atau IP tunggal dipisah koma (mis. "10.0.0.0/8, 192.168.1.5").
// Kosong berarti tidak ada proxy tepercaya.
func parseProxies(raw string) ([]netip.Prefix, error) {
	var out []netip.Prefix
	var errs []error
	for _, bagian := range strings.Split(raw, ",") {
		bagian = strings.TrimSpace(bagian)
		if bagian == "" {
			continue
		}
		if p, err := netip.ParsePrefix(bagian); err == nil {
			out = append(out, p.Masked())
			continue
		}
		a, err := netip.ParseAddr(bagian)
		if err != nil {
			errs = append(errs, fmt.Errorf("TRUSTED_PROXIES: %q bukan CIDR atau alamat IP yang valid", bagian))
			continue
		}
		a = a.Unmap()
		out = append(out, netip.PrefixFrom(a, a.BitLen()))
	}
	return out, errors.Join(errs...)
}
