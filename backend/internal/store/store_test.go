package store

import (
	"context"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

// Tes integrasi: memakai database tes (ai_center_test) milik Docker lokal.
func TestOpenDanPing(t *testing.T) {
	url := testutil.DatabaseTes(t)
	ctx, batal := context.WithTimeout(context.Background(), 5*time.Second)
	defer batal()

	s, err := Open(ctx, url)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()
	if err := s.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestOpenGagalBilaURLSalah(t *testing.T) {
	ctx, batal := context.WithTimeout(context.Background(), 3*time.Second)
	defer batal()
	if _, err := Open(ctx, "postgres://u:p@127.0.0.1:1/tidak-ada?sslmode=disable&connect_timeout=1"); err == nil {
		t.Fatal("harus galat bila database tidak terjangkau")
	}
}
