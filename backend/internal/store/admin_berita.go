package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store/db"
)

var _ admin.BeritaRepo = RepoBerita{}

// RepoBerita adalah akses data berita untuk admin (adaptor di atas Store; nama metode sama dengan jenis konten lain).
type RepoBerita struct{ *Store }

// AdminBerita mengembalikan repositori berita untuk admin.
func (s *Store) AdminBerita() RepoBerita { return RepoBerita{s} }

func tanggalPg(s string) (pgtype.Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("tanggal %q: %w", s, err)
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func beritaAdmin(r db.Beritum) admin.Berita {
	b := admin.Berita{
		ID: r.ID, Slug: r.Slug, Judul: r.Judul, Ringkasan: teks(r.Ringkasan), Isi: r.Isi,
		TanggalPublikasi: tanggal(r.TanggalPublikasi), GambarKey: r.Gambar, Status: r.Status, Version: r.Version,
		DibuatPada: r.CreatedAt.Time.UTC(), DibuatOleh: r.CreatedBy, DiubahOleh: r.UpdatedBy,
	}
	if r.DeletedAt.Valid {
		t := r.DeletedAt.Time.UTC()
		b.DihapusPada = &t
	}
	return b
}

// adalahUnik mengenali pelanggaran UNIQUE pada constraint bernama tertentu.
func adalahUnik(err error, constraint string) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505" && pg.ConstraintName == constraint
}

func (s RepoBerita) Daftar(ctx context.Context, p admin.ParamDaftar) ([]admin.Berita, int64, error) {
	baris, err := s.Q.ListBeritaAdmin(ctx, db.ListBeritaAdminParams{Filter: p.Filter, Pola: p.Pola, Batas: p.Batas, Geser: p.Geser})
	if err != nil {
		return nil, 0, fmt.Errorf("daftar berita admin: %w", err)
	}
	total, err := s.Q.HitungBeritaAdmin(ctx, db.HitungBeritaAdminParams{Filter: p.Filter, Pola: p.Pola})
	if err != nil {
		return nil, 0, fmt.Errorf("hitung berita admin: %w", err)
	}
	out := make([]admin.Berita, 0, len(baris))
	for _, r := range baris {
		out = append(out, beritaAdmin(r))
	}
	return out, total, nil
}

func (s RepoBerita) Detail(ctx context.Context, id int64) (admin.Berita, error) {
	r, err := s.Q.GetBeritaAdmin(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Berita{}, admin.ErrTidakDitemukan
	}
	if err != nil {
		return admin.Berita{}, fmt.Errorf("detail berita admin: %w", err)
	}
	return beritaAdmin(r), nil
}

func (s RepoBerita) Buat(ctx context.Context, slug string, in admin.InputBerita, oleh string) (admin.Berita, bool, error) {
	tgl, err := tanggalPg(in.TanggalPublikasi)
	if err != nil {
		return admin.Berita{}, false, err
	}
	r, err := s.Q.BuatBerita(ctx, db.BuatBeritaParams{Slug: slug, Judul: in.Judul, Isi: in.Isi, TanggalPublikasi: tgl, Gambar: in.Gambar, Oleh: oleh})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Berita{}, false, nil // slug sudah dipakai
	}
	if adalahUnik(err, "berita_gambar_key") {
		return admin.Berita{}, false, admin.ErrGambarDipakai
	}
	if err != nil {
		return admin.Berita{}, false, fmt.Errorf("buat berita: %w", err)
	}
	return beritaAdmin(r), true, nil
}

// klasifikasi menjelaskan mengapa UPDATE/DELETE bersyarat tidak mengenai baris mana pun (aturan bersama di admin_umum.go).
func (s RepoBerita) klasifikasi(ctx context.Context, id int64, a aturan, version int32) error {
	r, err := s.Q.GetBeritaAdmin(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.ErrTidakDitemukan
	}
	if err != nil {
		return fmt.Errorf("klasifikasi berita: %w", err)
	}
	return putuskanKlasifikasi(a, r.DeletedAt.Valid, r.Status, r.Version, version)
}

func (s RepoBerita) Ubah(ctx context.Context, id int64, version int32, in admin.InputBerita, oleh string) (admin.Berita, error) {
	tgl, err := tanggalPg(in.TanggalPublikasi)
	if err != nil {
		return admin.Berita{}, err
	}
	r, err := s.Q.UbahBerita(ctx, db.UbahBeritaParams{
		Judul: in.Judul, Isi: in.Isi, TanggalPublikasi: tgl, Gambar: in.Gambar, Oleh: oleh, ID: id, Version: version,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Berita{}, s.klasifikasi(ctx, id, aturanDraftBerversi, version)
	}
	if adalahUnik(err, "berita_gambar_key") {
		return admin.Berita{}, admin.ErrGambarDipakai
	}
	if err != nil {
		return admin.Berita{}, fmt.Errorf("ubah berita: %w", err)
	}
	return beritaAdmin(r), nil
}

func (s RepoBerita) Post(ctx context.Context, id int64, version int32, oleh string) (admin.Berita, error) {
	r, err := s.Q.PostBerita(ctx, db.PostBeritaParams{Oleh: oleh, ID: id, Version: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Berita{}, s.klasifikasi(ctx, id, aturanDraftBerversi, version)
	}
	if err != nil {
		return admin.Berita{}, fmt.Errorf("post berita: %w", err)
	}
	return beritaAdmin(r), nil
}

func (s RepoBerita) Retract(ctx context.Context, id int64, version int32, oleh string) (admin.Berita, error) {
	r, err := s.Q.RetractBerita(ctx, db.RetractBeritaParams{Oleh: oleh, ID: id, Version: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Berita{}, s.klasifikasi(ctx, id, aturanPostedBerversi, version)
	}
	if err != nil {
		return admin.Berita{}, fmt.Errorf("retract berita: %w", err)
	}
	return beritaAdmin(r), nil
}

func (s RepoBerita) Hapus(ctx context.Context, id int64, oleh string) error {
	_, err := s.Q.HapusBerita(ctx, db.HapusBeritaParams{Oleh: oleh, ID: id})
	if errors.Is(err, pgx.ErrNoRows) {
		return s.klasifikasi(ctx, id, aturanDraftAktif, 0)
	}
	if err != nil {
		return fmt.Errorf("hapus berita: %w", err)
	}
	return nil
}

func (s RepoBerita) Pulihkan(ctx context.Context, id int64, oleh string) (admin.Berita, error) {
	r, err := s.Q.PulihkanBerita(ctx, db.PulihkanBeritaParams{Oleh: oleh, ID: id})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Berita{}, s.klasifikasi(ctx, id, aturanDiTrash, 0)
	}
	if err != nil {
		return admin.Berita{}, fmt.Errorf("pulihkan berita: %w", err)
	}
	return beritaAdmin(r), nil
}

func (s RepoBerita) HapusPermanen(ctx context.Context, id int64) (string, error) {
	key, err := s.Q.HapusPermanenBerita(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", s.klasifikasi(ctx, id, aturanDiTrash, 0)
	}
	if err != nil {
		return "", fmt.Errorf("hapus permanen berita: %w", err)
	}
	return key, nil
}

func (s RepoBerita) GambarDipakai(ctx context.Context, key string) (bool, error) {
	ada, err := s.Q.GambarBeritaDipakai(ctx, key)
	if err != nil {
		return false, fmt.Errorf("periksa pemakaian gambar: %w", err)
	}
	return ada, nil
}
