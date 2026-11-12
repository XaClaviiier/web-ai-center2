package testutil

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BersihkanKonten mengosongkan tabel konten di database TES (bukan ai_center) sebelum dan sesudah tes.
// Paket yang memakai database dijalankan berurutan (go test -p 1) agar tidak saling menimpa.
func BersihkanKonten(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	bersih := func() {
		_, err := pool.Exec(context.Background(),
			"TRUNCATE berita, dokumentasi, dokumentasi_gambar, testimoni, tim, log_aktivitas RESTART IDENTITY CASCADE")
		if err != nil {
			t.Fatalf("membersihkan tabel tes: %v", err)
		}
	}
	bersih()
	t.Cleanup(bersih)
}
