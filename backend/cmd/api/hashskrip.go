package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var polaHashSkrip = regexp.MustCompile(`^sha256-[A-Za-z0-9+/]{43}=$`)

// muatHashSkrip membaca STATIC_DIR/csp-hash-skrip.txt (satu hash CSP per baris, ditulis oleh scripts/bangun-web-final.cjs)
// untuk script-src. Berkas tidak ada: hanya 'self' (halaman Astro dengan island tidak akan terhidrasi, jadi diberi peringatan).
// Baris yang bukan hash sah dilewati.
func muatHashSkrip(dir string, log *slog.Logger) []string {
	isi, err := os.ReadFile(filepath.Join(dir, "csp-hash-skrip.txt"))
	if err != nil {
		log.Warn("csp-hash-skrip.txt tidak ditemukan; script-src hanya 'self' (jalankan scripts/bangun-web-final.cjs)", "dir", dir)
		return nil
	}
	var out []string
	for _, baris := range strings.Split(string(isi), "\n") {
		baris = strings.TrimSpace(baris)
		if polaHashSkrip.MatchString(baris) {
			out = append(out, baris)
		}
	}
	return out
}
