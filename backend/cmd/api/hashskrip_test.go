package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMuatHashSkrip(t *testing.T) {
	dir := t.TempDir()
	if got := muatHashSkrip(dir, logUji); got != nil {
		t.Errorf("tanpa berkas: %v", got)
	}
	baik := "sha256-1BvDVWLr3QCC20uHhUDi27eDFVJpk1uMQJGIWKUks2I="
	isi := baik + "\n\n  " + baik + "  \nsha256-pendek\n'unsafe-inline'\nsha256-1BvDVWLr3QCC20uHhUDi27eDFVJpk1uMQJGIWKUks2I=x\n"
	if err := os.WriteFile(filepath.Join(dir, "csp-hash-skrip.txt"), []byte(isi), 0o644); err != nil {
		t.Fatal(err)
	}
	got := muatHashSkrip(dir, logUji)
	if len(got) != 2 || got[0] != baik || got[1] != baik {
		t.Errorf("hash sah saja yang dimuat: %v", got)
	}
}
