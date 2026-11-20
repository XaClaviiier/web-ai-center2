package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/store"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

// Tes ujung ke ujung berita: router asli + Store asli + database tes (ai_center_test).
func TestE2EBeritaPublik(t *testing.T) {
	url := testutil.DatabaseTes(t)
	ctx, batal := context.WithTimeout(context.Background(), 20*time.Second)
	defer batal()
	st, err := store.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	// t.Cleanup berjalan terbalik: Close didaftarkan lebih dulu agar pool baru ditutup setelah tabel dibersihkan.
	t.Cleanup(st.Close)
	testutil.BersihkanKonten(t, st.Pool)

	_, err = st.Pool.Exec(ctx, `INSERT INTO berita (slug, judul, isi, tanggal_publikasi, gambar, status, created_by, updated_by, deleted_at) VALUES
		('pertama', 'Berita Pertama', E'Paragraf awal.\n\nLanjutan', '2026-06-01', 'berita/1/a.jpg', 'posted', 'u', 'u', NULL),
		('kedua',   'Berita Kedua',   'Isi kedua',                   '2026-07-01', 'berita/2/b.jpg', 'posted', 'u', 'u', NULL),
		('draft',   'Masih Draft',    'Rahasia',                     '2026-08-01', 'berita/3/c.jpg', 'draft',  'u', 'u', NULL),
		('buang',   'Sudah Dibuang',  'Dibuang',                     '2026-09-01', 'berita/4/d.jpg', 'posted', 'u', 'u', now())`)
	if err != nil {
		t.Fatal(err)
	}

	h := NewRouter(Deps{Log: slog.New(slog.NewTextHandler(io.Discard, nil)), DB: st, Konten: st})

	rec := panggil(h, http.MethodGet, "/api/berita?limit=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("daftar status %d: %s", rec.Code, rec.Body)
	}
	var daftar struct {
		Items []struct {
			Slug             string `json:"slug"`
			Ringkasan        string `json:"ringkasan"`
			Gambar           string `json:"gambar"`
			TanggalPublikasi string `json:"tanggalPublikasi"`
		} `json:"items"`
		Total int64 `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &daftar); err != nil {
		t.Fatal(err)
	}
	if daftar.Total != 2 || len(daftar.Items) != 2 || daftar.Items[0].Slug != "kedua" || daftar.Items[1].Slug != "pertama" {
		t.Fatalf("daftar salah: %s", rec.Body)
	}
	if daftar.Items[1].Ringkasan != "Paragraf awal." || daftar.Items[1].Gambar != "/media/berita/1/a.jpg" || daftar.Items[1].TanggalPublikasi != "01-06-2026" {
		t.Errorf("item salah: %+v", daftar.Items[1])
	}

	if rec := panggil(h, http.MethodGet, "/api/berita/pertama"); rec.Code != http.StatusOK {
		t.Errorf("detail pertama status %d", rec.Code)
	}
	for _, slug := range []string{"draft", "buang", "tidak-ada"} {
		if rec := panggil(h, http.MethodGet, "/api/berita/"+slug); rec.Code != http.StatusNotFound {
			t.Errorf("%s status %d, diharapkan 404", slug, rec.Code)
		}
	}
	if rec := panggil(h, http.MethodGet, "/api/berita?q=kedua"); rec.Code != http.StatusOK {
		t.Errorf("pencarian status %d", rec.Code)
	}
}
