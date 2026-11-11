package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/konten"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store/db"
)

// Pastikan Store memenuhi kontrak baca publik.
var _ konten.BacaPublik = (*Store)(nil)

func tanggal(d pgtype.Date) string { return d.Time.Format("2006-01-02") }

func teks(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func (s *Store) ListBerita(ctx context.Context, p konten.ParamDaftar) ([]konten.BeritaRingkas, int64, error) {
	var (
		baris []db.ListBeritaPublikRow
		total int64
		err   error
	)
	if p.Pola == "" {
		if baris, err = s.Q.ListBeritaPublik(ctx, db.ListBeritaPublikParams{Batas: p.Batas, Geser: p.Geser}); err != nil {
			return nil, 0, fmt.Errorf("daftar berita: %w", err)
		}
		if total, err = s.Q.HitungBeritaPublik(ctx); err != nil {
			return nil, 0, fmt.Errorf("hitung berita: %w", err)
		}
	} else {
		cari, err := s.Q.CariBeritaPublik(ctx, db.CariBeritaPublikParams{Pola: p.Pola, Batas: p.Batas, Geser: p.Geser})
		if err != nil {
			return nil, 0, fmt.Errorf("cari berita: %w", err)
		}
		for _, r := range cari {
			baris = append(baris, db.ListBeritaPublikRow(r))
		}
		if total, err = s.Q.HitungCariBeritaPublik(ctx, p.Pola); err != nil {
			return nil, 0, fmt.Errorf("hitung cari berita: %w", err)
		}
	}
	out := make([]konten.BeritaRingkas, 0, len(baris))
	for _, r := range baris {
		out = append(out, konten.BeritaRingkas{
			ID: r.ID, Slug: r.Slug, Judul: r.Judul, Ringkasan: teks(r.Ringkasan),
			GambarKey: r.Gambar, TanggalPublikasi: tanggal(r.TanggalPublikasi),
		})
	}
	return out, total, nil
}

func (s *Store) BeritaBySlug(ctx context.Context, slug string) (konten.BeritaDetail, error) {
	r, err := s.Q.GetBeritaPublik(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return konten.BeritaDetail{}, konten.ErrTidakDitemukan
	}
	if err != nil {
		return konten.BeritaDetail{}, fmt.Errorf("detail berita: %w", err)
	}
	return konten.BeritaDetail{
		BeritaRingkas: konten.BeritaRingkas{
			ID: r.ID, Slug: r.Slug, Judul: r.Judul, Ringkasan: teks(r.Ringkasan),
			GambarKey: r.Gambar, TanggalPublikasi: tanggal(r.TanggalPublikasi),
		},
		Isi: r.Isi,
	}, nil
}

func (s *Store) ListDokumentasi(ctx context.Context, p konten.ParamDaftar) ([]konten.DokumentasiRingkas, int64, error) {
	var (
		baris []db.ListDokumentasiPublikRow
		total int64
		err   error
	)
	if p.Pola == "" {
		if baris, err = s.Q.ListDokumentasiPublik(ctx, db.ListDokumentasiPublikParams{Batas: p.Batas, Geser: p.Geser}); err != nil {
			return nil, 0, fmt.Errorf("daftar dokumentasi: %w", err)
		}
		if total, err = s.Q.HitungDokumentasiPublik(ctx); err != nil {
			return nil, 0, fmt.Errorf("hitung dokumentasi: %w", err)
		}
	} else {
		cari, err := s.Q.CariDokumentasiPublik(ctx, db.CariDokumentasiPublikParams{Pola: p.Pola, Batas: p.Batas, Geser: p.Geser})
		if err != nil {
			return nil, 0, fmt.Errorf("cari dokumentasi: %w", err)
		}
		for _, r := range cari {
			baris = append(baris, db.ListDokumentasiPublikRow(r))
		}
		if total, err = s.Q.HitungCariDokumentasiPublik(ctx, p.Pola); err != nil {
			return nil, 0, fmt.Errorf("hitung cari dokumentasi: %w", err)
		}
	}
	out := make([]konten.DokumentasiRingkas, 0, len(baris))
	for _, r := range baris {
		out = append(out, konten.DokumentasiRingkas{
			ID: r.ID, Slug: r.Slug, Judul: r.Judul,
			TanggalPublikasi: tanggal(r.TanggalPublikasi), GambarUtamaKey: teks(r.GambarUtama),
		})
	}
	return out, total, nil
}

func (s *Store) DokumentasiBySlug(ctx context.Context, slug string) (konten.DokumentasiDetail, error) {
	r, err := s.Q.GetDokumentasiPublik(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return konten.DokumentasiDetail{}, konten.ErrTidakDitemukan
	}
	if err != nil {
		return konten.DokumentasiDetail{}, fmt.Errorf("detail dokumentasi: %w", err)
	}
	gambar, err := s.Q.ListGambarDokumentasi(ctx, r.ID)
	if err != nil {
		return konten.DokumentasiDetail{}, fmt.Errorf("gambar dokumentasi: %w", err)
	}
	out := konten.DokumentasiDetail{
		ID: r.ID, Slug: r.Slug, Judul: r.Judul, TanggalPublikasi: tanggal(r.TanggalPublikasi),
		Gambar: make([]konten.GambarDokumentasi, 0, len(gambar)),
	}
	for _, g := range gambar {
		out.Gambar = append(out.Gambar, konten.GambarDokumentasi{Key: g.Key, Alt: teks(g.Alt)})
	}
	return out, nil
}

func (s *Store) ListTestimoni(ctx context.Context) ([]konten.Testimoni, error) {
	baris, err := s.Q.ListTestimoniPublik(ctx)
	if err != nil {
		return nil, fmt.Errorf("daftar testimoni: %w", err)
	}
	out := make([]konten.Testimoni, 0, len(baris))
	for _, r := range baris {
		out = append(out, konten.Testimoni{ID: r.ID, Nama: r.Nama, Teks: r.Teks, Posisi: r.Posisi, FotoKey: teks(r.Foto)})
	}
	return out, nil
}

func (s *Store) ListTim(ctx context.Context) ([]konten.AnggotaTim, error) {
	baris, err := s.Q.ListTim(ctx)
	if err != nil {
		return nil, fmt.Errorf("daftar tim: %w", err)
	}
	out := make([]konten.AnggotaTim, 0, len(baris))
	for _, r := range baris {
		out = append(out, konten.AnggotaTim{ID: r.ID, Nama: r.Nama, Posisi: r.Posisi, BidangRiset: r.BidangRiset, FotoKey: r.Foto})
	}
	return out, nil
}
