package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func lookupDari(m map[string]string) Lookup {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func dasarDev() map[string]string {
	return map[string]string{
		"APP_ENV":           "development",
		"DATABASE_URL":      "postgres://u:p@localhost:5432/x",
		"REDIS_URL":         "redis://:p@localhost:6379/0",
		"RUSTFS_ENDPOINT":   "http://localhost:9000",
		"RUSTFS_ACCESS_KEY": "ak",
		"RUSTFS_SECRET_KEY": "sk",
		"RUSTFS_BUCKET":     "ai-center-images",
		"ADMIN_PASSWORD":    "rahasia",
	}
}

func dasarProd() map[string]string {
	return map[string]string{
		"APP_ENV":           "production",
		"DATABASE_URL":      "postgres://u:p@db:5432/x",
		"REDIS_URL":         "redis://:p@redis:6379/0",
		"RUSTFS_ENDPOINT":   "http://rustfs:9000",
		"RUSTFS_ACCESS_KEY": "ak",
		"RUSTFS_SECRET_KEY": "sk",
		"RUSTFS_BUCKET":     "ai-center-images",
		"PUBLIC_BASE_URL":   "https://ai-center.example.com/",
		"ALLOWED_ORIGINS":   "https://ai-center.example.com",
	}
}

func TestLoadDevelopmentDefault(t *testing.T) {
	cfg, err := Load(lookupDari(dasarDev()))
	if err != nil {
		t.Fatalf("galat tak terduga: %v", err)
	}
	if cfg.Port != 8080 || cfg.SessionCookieName != "session" || cfg.SessionCookieSecure || !cfg.DevLoginAktif {
		t.Fatalf("default dev salah: %+v", cfg)
	}
	if cfg.SessionIdleTTL != time.Hour || cfg.SessionAbsoluteTTL != 8*time.Hour || cfg.RateLimitLoginPerMinute != 10 {
		t.Fatalf("default TTL/batas login salah: %v %v %d", cfg.SessionIdleTTL, cfg.SessionAbsoluteTTL, cfg.RateLimitLoginPerMinute)
	}
	if cfg.RateLimitPublicMissPerMinute != 300 || cfg.CacheTTL != 20*time.Minute {
		t.Fatalf("default limiter publik/TTL cache salah: %d %v", cfg.RateLimitPublicMissPerMinute, cfg.CacheTTL)
	}
}

func TestLoadBatasUnggahDefaultDanEnv(t *testing.T) {
	cfg, err := Load(lookupDari(dasarDev()))
	if err != nil || cfg.UploadMaxBytes != 2<<20 {
		t.Fatalf("default harus 2 MB: %d %v", cfg.UploadMaxBytes, err)
	}
	m := dasarDev()
	m["UPLOAD_MAX_MB"] = "5"
	cfg, err = Load(lookupDari(m))
	if err != nil || cfg.UploadMaxBytes != 5<<20 {
		t.Fatalf("UPLOAD_MAX_MB=5 harus 5 MB: %d %v", cfg.UploadMaxBytes, err)
	}
}

func TestLoadRustFSDanBatasUnggahGalat(t *testing.T) {
	cases := []struct {
		nama, kunci, nilai, wajib string
	}{
		{"endpoint kosong", "RUSTFS_ENDPOINT", "", "RUSTFS_ENDPOINT wajib"},
		{"endpoint bukan URL", "RUSTFS_ENDPOINT", "localhost:9000", "RUSTFS_ENDPOINT"},
		{"access key kosong", "RUSTFS_ACCESS_KEY", "", "RUSTFS_ACCESS_KEY wajib"},
		{"secret key kosong", "RUSTFS_SECRET_KEY", "", "RUSTFS_SECRET_KEY wajib"},
		{"bucket kosong", "RUSTFS_BUCKET", "", "RUSTFS_BUCKET wajib"},
		{"upload nol", "UPLOAD_MAX_MB", "0", "UPLOAD_MAX_MB"},
		{"upload bukan angka", "UPLOAD_MAX_MB", "dua", "UPLOAD_MAX_MB"},
		{"upload terlalu besar", "UPLOAD_MAX_MB", "500", "terlalu besar"},
	}
	for _, c := range cases {
		t.Run(c.nama, func(t *testing.T) {
			m := dasarDev()
			m[c.kunci] = c.nilai
			_, err := Load(lookupDari(m))
			if err == nil || !strings.Contains(err.Error(), c.wajib) {
				t.Fatalf("diharapkan galat memuat %q, dapat: %v", c.wajib, err)
			}
		})
	}
}

func TestLoadTTLDanBatasLoginDariEnv(t *testing.T) {
	m := dasarDev()
	m["SESSION_IDLE_TTL_SECONDS"] = "600"
	m["SESSION_ABSOLUTE_TTL_SECONDS"] = "7200"
	m["RATE_LIMIT_LOGIN_PER_MINUTE"] = "3"
	m["RATE_LIMIT_PUBLIC_MISS_PER_MINUTE"] = "50"
	m["CACHE_TTL_SECONDS"] = "30"
	cfg, err := Load(lookupDari(m))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SessionIdleTTL != 10*time.Minute || cfg.SessionAbsoluteTTL != 2*time.Hour || cfg.RateLimitLoginPerMinute != 3 {
		t.Fatalf("nilai env tidak terbaca: %+v", cfg)
	}
	if cfg.RateLimitPublicMissPerMinute != 50 || cfg.CacheTTL != 30*time.Second {
		t.Fatalf("env limiter publik/TTL cache tidak terbaca: %d %v", cfg.RateLimitPublicMissPerMinute, cfg.CacheTTL)
	}
}

func TestLoadProduksiMemaksaCookieAman(t *testing.T) {
	cfg, err := Load(lookupDari(dasarProd()))
	if err != nil {
		t.Fatalf("galat tak terduga: %v", err)
	}
	if cfg.SessionCookieName != NamaCookieProduksi || !cfg.SessionCookieSecure || cfg.DevLoginAktif {
		t.Fatalf("produksi harus __Host-session + Secure tanpa dev login: %+v", cfg)
	}
	if cfg.PublicBaseURL != "https://ai-center.example.com" {
		t.Fatalf("slash akhir harus dibuang: %q", cfg.PublicBaseURL)
	}
}

func TestLoadGalat(t *testing.T) {
	cases := []struct {
		nama  string
		ubah  func(m map[string]string)
		dasar func() map[string]string
		wajib string
	}{
		{"APP_ENV kosong", func(m map[string]string) { delete(m, "APP_ENV") }, dasarDev, "APP_ENV wajib"},
		{"APP_ENV asing", func(m map[string]string) { m["APP_ENV"] = "staging" }, dasarDev, "tidak dikenal"},
		{"DATABASE_URL kosong", func(m map[string]string) { delete(m, "DATABASE_URL") }, dasarDev, "DATABASE_URL wajib"},
		{"REDIS_URL kosong", func(m map[string]string) { delete(m, "REDIS_URL") }, dasarDev, "REDIS_URL wajib"},
		{"REDIS_URL skema salah", func(m map[string]string) { m["REDIS_URL"] = "http://x:6379" }, dasarDev, "REDIS_URL harus"},
		{"idle TTL nol", func(m map[string]string) { m["SESSION_IDLE_TTL_SECONDS"] = "0" }, dasarDev, "SESSION_IDLE_TTL_SECONDS"},
		{"mutlak TTL bukan angka", func(m map[string]string) { m["SESSION_ABSOLUTE_TTL_SECONDS"] = "x" }, dasarDev, "SESSION_ABSOLUTE_TTL_SECONDS"},
		{"idle melebihi mutlak", func(m map[string]string) {
			m["SESSION_IDLE_TTL_SECONDS"] = "9000"
			m["SESSION_ABSOLUTE_TTL_SECONDS"] = "100"
		}, dasarDev, "tidak boleh lebih besar"},
		{"batas publik nol", func(m map[string]string) { m["RATE_LIMIT_PUBLIC_MISS_PER_MINUTE"] = "0" }, dasarDev, "RATE_LIMIT_PUBLIC_MISS_PER_MINUTE"},
		{"ttl cache bukan angka", func(m map[string]string) { m["CACHE_TTL_SECONDS"] = "dua" }, dasarDev, "CACHE_TTL_SECONDS"},
		{"batas login negatif", func(m map[string]string) { m["RATE_LIMIT_LOGIN_PER_MINUTE"] = "-1" }, dasarDev, "RATE_LIMIT_LOGIN_PER_MINUTE"},
		{"PORT bukan angka", func(m map[string]string) { m["PORT"] = "abc" }, dasarDev, "PORT"},
		{"PORT di luar rentang", func(m map[string]string) { m["PORT"] = "70000" }, dasarDev, "PORT"},
		{"dev tanpa ADMIN_PASSWORD", func(m map[string]string) { delete(m, "ADMIN_PASSWORD") }, dasarDev, "ADMIN_PASSWORD wajib"},
		{"origin dengan path", func(m map[string]string) { m["ALLOWED_ORIGINS"] = "https://a.example.com/x" }, dasarDev, "ALLOWED_ORIGINS"},
		{"produksi nama cookie salah", func(m map[string]string) { m["SESSION_COOKIE_NAME"] = "session" }, dasarProd, "__Host-session"},
		{"produksi Secure=false", func(m map[string]string) { m["SESSION_COOKIE_SECURE"] = "false" }, dasarProd, "harus true"},
		{"produksi tanpa PUBLIC_BASE_URL", func(m map[string]string) { delete(m, "PUBLIC_BASE_URL") }, dasarProd, "PUBLIC_BASE_URL wajib"},
		{"produksi tanpa ALLOWED_ORIGINS", func(m map[string]string) { delete(m, "ALLOWED_ORIGINS") }, dasarProd, "ALLOWED_ORIGINS wajib"},
	}
	for _, c := range cases {
		t.Run(c.nama, func(t *testing.T) {
			m := c.dasar()
			c.ubah(m)
			_, err := Load(lookupDari(m))
			if err == nil || !strings.Contains(err.Error(), c.wajib) {
				t.Fatalf("diharapkan galat memuat %q, dapat: %v", c.wajib, err)
			}
		})
	}
}

func TestLoadMengumpulkanSemuaGalat(t *testing.T) {
	_, err := Load(lookupDari(map[string]string{"PORT": "0"}))
	if err == nil {
		t.Fatal("harus galat")
	}
	for _, s := range []string{"APP_ENV", "DATABASE_URL", "PORT"} {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("galat harus menyebut %s: %v", s, err)
		}
	}
}

func TestLoadDotEnvTidakMenimpas(t *testing.T) {
	p := filepath.Join(t.TempDir(), ".env")
	isi := "# komentar\nCFG_TEST_BARU=dari-berkas\nCFG_TEST_ADA=dari-berkas\n\nCFG_TEST_SAMA=a=b\n"
	if err := os.WriteFile(p, []byte(isi), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CFG_TEST_ADA", "dari-env")
	t.Cleanup(func() { os.Unsetenv("CFG_TEST_BARU"); os.Unsetenv("CFG_TEST_SAMA") })

	if err := LoadDotEnv(p); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("CFG_TEST_BARU") != "dari-berkas" {
		t.Error("variabel baru harus terisi")
	}
	if os.Getenv("CFG_TEST_ADA") != "dari-env" {
		t.Error("variabel yang sudah ada tidak boleh ditimpa")
	}
	if os.Getenv("CFG_TEST_SAMA") != "a=b" {
		t.Error("nilai yang memuat '=' harus utuh")
	}
	if err := LoadDotEnv(filepath.Join(t.TempDir(), "tidak-ada")); err != nil {
		t.Errorf("berkas tidak ada bukan galat: %v", err)
	}
}

func TestLoadLimiterAdminDanProxyTepercaya(t *testing.T) {
	cfg, err := Load(lookupDari(dasarDev()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RateLimitAdminReadPerMinute != 180 || cfg.RateLimitAdminWritePerMinute != 30 || len(cfg.TrustedProxies) != 0 {
		t.Fatalf("default salah: %d %d %v", cfg.RateLimitAdminReadPerMinute, cfg.RateLimitAdminWritePerMinute, cfg.TrustedProxies)
	}
	m := dasarDev()
	m["RATE_LIMIT_ADMIN_READ_PER_MINUTE"] = "7"
	m["RATE_LIMIT_ADMIN_WRITE_PER_MINUTE"] = "4"
	m["TRUSTED_PROXIES"] = " 10.0.0.0/8, 192.168.1.5 ,::1 "
	cfg, err = Load(lookupDari(m))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RateLimitAdminReadPerMinute != 7 || cfg.RateLimitAdminWritePerMinute != 4 {
		t.Errorf("batas admin: %d %d", cfg.RateLimitAdminReadPerMinute, cfg.RateLimitAdminWritePerMinute)
	}
	var got []string
	for _, p := range cfg.TrustedProxies {
		got = append(got, p.String())
	}
	if len(got) != 3 || got[0] != "10.0.0.0/8" || got[1] != "192.168.1.5/32" || got[2] != "::1/128" {
		t.Errorf("proxy tepercaya: %v", got)
	}
	m["TRUSTED_PROXIES"] = "10.0.0.0/8, bukan-ip, 300.1.1.1"
	if _, err := Load(lookupDari(m)); err == nil || !strings.Contains(err.Error(), "TRUSTED_PROXIES") {
		t.Errorf("proxy tidak valid harus ditolak: %v", err)
	}
	m["TRUSTED_PROXIES"] = ""
	m["RATE_LIMIT_ADMIN_WRITE_PER_MINUTE"] = "0"
	if _, err := Load(lookupDari(m)); err == nil || !strings.Contains(err.Error(), "RATE_LIMIT_ADMIN_WRITE_PER_MINUTE") {
		t.Errorf("batas tulis nol harus ditolak: %v", err)
	}
}

func TestLoadLockDanAkunDevKedua(t *testing.T) {
	cfg, err := Load(lookupDari(dasarDev()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LockTTL != 5*time.Minute || cfg.LockHeartbeat != time.Minute || cfg.AdminPassword2 != "" {
		t.Fatalf("default salah: %v %v %q", cfg.LockTTL, cfg.LockHeartbeat, cfg.AdminPassword2)
	}
	m := dasarDev()
	m["LOCK_TTL_SECONDS"] = "120"
	m["LOCK_HEARTBEAT_SECONDS"] = "30"
	m["ADMIN_PASSWORD_2"] = "beda-dengan-pertama"
	cfg, err = Load(lookupDari(m))
	if err != nil || cfg.LockTTL != 2*time.Minute || cfg.LockHeartbeat != 30*time.Second || cfg.AdminPassword2 != "beda-dengan-pertama" {
		t.Fatalf("nilai env: %+v %v", cfg, err)
	}

	galat := func(ubah func(map[string]string), mau string) {
		t.Helper()
		m := dasarDev()
		ubah(m)
		if _, err := Load(lookupDari(m)); err == nil || !strings.Contains(err.Error(), mau) {
			t.Errorf("mau galat memuat %q, dapat %v", mau, err)
		}
	}
	galat(func(m map[string]string) { m["LOCK_HEARTBEAT_SECONDS"] = "300" }, "LOCK_HEARTBEAT_SECONDS harus lebih kecil")
	galat(func(m map[string]string) { m["LOCK_TTL_SECONDS"] = "0" }, "LOCK_TTL_SECONDS")
	galat(func(m map[string]string) { m["ADMIN_PASSWORD_2"] = m["ADMIN_PASSWORD"] }, "ADMIN_PASSWORD_2 harus berbeda")
}

func TestLoadSEODefaultDanEnv(t *testing.T) {
	cfg, err := Load(lookupDari(dasarDev()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.StaticDir != "../web" || cfg.SiteName != "AI Center" {
		t.Errorf("default salah: %q %q", cfg.StaticDir, cfg.SiteName)
	}
	if cfg.PublicBaseURL != "http://localhost:8080" {
		t.Errorf("development tanpa PUBLIC_BASE_URL harus memakai localhost:{PORT}, dapat %q", cfg.PublicBaseURL)
	}
	m := dasarDev()
	m["STATIC_DIR"] = "/srv/web"
	m["SITE_NAME"] = "Pusat AI"
	m["PORT"] = "9090"
	cfg, err = Load(lookupDari(m))
	if err != nil || cfg.StaticDir != "/srv/web" || cfg.SiteName != "Pusat AI" || cfg.PublicBaseURL != "http://localhost:9090" {
		t.Errorf("env tidak terbaca: %+v %v", cfg, err)
	}
	m["PUBLIC_BASE_URL"] = "https://situs.example.com/"
	if cfg, _ = Load(lookupDari(m)); cfg.PublicBaseURL != "https://situs.example.com" {
		t.Errorf("PUBLIC_BASE_URL eksplisit harus menang: %q", cfg.PublicBaseURL)
	}
}

func TestLoadJobDefaultDanEnv(t *testing.T) {
	dev, err := Load(lookupDari(dasarDev()))
	if err != nil {
		t.Fatal(err)
	}
	if dev.JobsEnabled || dev.JobsDryRun {
		t.Error("job harus nonaktif secara default di development (melindungi bucket dev)")
	}
	if dev.JobInterval != time.Hour || dev.JobLockTTL != 30*time.Minute || dev.TrashRetentionDays != 30 ||
		dev.OrphanMinAge != 24*time.Hour || dev.OrphanMaxDelete != 200 {
		t.Errorf("default salah: %+v", dev)
	}
	prod, err := Load(lookupDari(dasarProd()))
	if err != nil || !prod.JobsEnabled || prod.JobsDryRun {
		t.Errorf("production: job aktif secara default, dry-run mati: %+v %v", prod.JobsEnabled, err)
	}

	m := dasarDev()
	m["JOBS_ENABLED"] = "true"
	m["JOBS_DRY_RUN"] = "true"
	m["JOB_INTERVAL_MINUTES"] = "5"
	m["JOB_LOCK_TTL_SECONDS"] = "120"
	m["TRASH_RETENTION_DAYS"] = "7"
	m["ORPHAN_MIN_AGE_HOURS"] = "48"
	m["ORPHAN_MAX_DELETE_PER_RUN"] = "10"
	cfg, err := Load(lookupDari(m))
	if err != nil || !cfg.JobsEnabled || !cfg.JobsDryRun || cfg.JobInterval != 5*time.Minute || cfg.JobLockTTL != 2*time.Minute ||
		cfg.TrashRetentionDays != 7 || cfg.OrphanMinAge != 48*time.Hour || cfg.OrphanMaxDelete != 10 {
		t.Errorf("env tidak terbaca: %+v %v", cfg, err)
	}

	mati := dasarProd()
	mati["JOBS_ENABLED"] = "false"
	if cfg, err := Load(lookupDari(mati)); err != nil || cfg.JobsEnabled {
		t.Errorf("production boleh mematikan job: %v %v", cfg.JobsEnabled, err)
	}

	for _, kv := range [][2]string{{"JOBS_ENABLED", "mungkin"}, {"JOBS_DRY_RUN", "x"}, {"JOB_INTERVAL_MINUTES", "0"},
		{"TRASH_RETENTION_DAYS", "-1"}, {"ORPHAN_MIN_AGE_HOURS", "abc"}, {"ORPHAN_MAX_DELETE_PER_RUN", "0"}, {"JOB_LOCK_TTL_SECONDS", "0"}} {
		m := dasarDev()
		m[kv[0]] = kv[1]
		if _, err := Load(lookupDari(m)); err == nil || !strings.Contains(err.Error(), kv[0]) {
			t.Errorf("%s=%s harus ditolak: %v", kv[0], kv[1], err)
		}
	}
}
