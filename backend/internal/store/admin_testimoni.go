package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store/db"
)

var _ admin.TestimoniRepo = RepoTestimoni{}

// RepoTestimoni adalah akses data testimoni untuk admin (adaptor di atas Store).
type RepoTestimoni struct{ *Store }

// AdminTestimoni mengembalikan repositori testimoni untuk admin.
func (s *Store) AdminTestimoni() RepoTestimoni { return RepoTestimoni{s} }

func testimoniAdmin(r db.Testimoni) admin.Testimoni {
	t := admin.Testimoni{
		ID: r.ID, Nama: r.Nama, Teks: r.Teks, Posisi: r.Posisi, FotoKey: teks(r.Foto), Status: r.Status, Version: r.Version,
		DibuatPada: r.CreatedAt.Time.UTC(), DibuatOleh: r.CreatedBy, DiubahOleh: r.UpdatedBy,
	}
	if r.DeletedAt.Valid {
		d := r.DeletedAt.Time.UTC()
		t.DihapusPada = &d
	}
	return t
}

func (s RepoTestimoni) klasifikasi(ctx context.Context, id int64, a aturan, version int32) error {
	r, err := s.Q.GetTestimoniAdmin(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.ErrTidakDitemukan
	}
	if err != nil {
		return fmt.Errorf("klasifikasi testimoni: %w", err)
	}
	return putuskanKlasifikasi(a, r.DeletedAt.Valid, r.Status, r.Version, version)
}

func (s RepoTestimoni) Daftar(ctx context.Context, filter string) ([]admin.Testimoni, error) {
	baris, err := s.Q.ListTestimoniAdmin(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("daftar testimoni admin: %w", err)
	}
	out := make([]admin.Testimoni, 0, len(baris))
	for _, r := range baris {
		out = append(out, testimoniAdmin(r))
	}
	return out, nil
}

func (s RepoTestimoni) Detail(ctx context.Context, id int64) (admin.Testimoni, error) {
	r, err := s.Q.GetTestimoniAdmin(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Testimoni{}, admin.ErrTidakDitemukan
	}
	if err != nil {
		return admin.Testimoni{}, fmt.Errorf("detail testimoni admin: %w", err)
	}
	return testimoniAdmin(r), nil
}

func (s RepoTestimoni) Buat(ctx context.Context, in admin.InputTestimoni, oleh string) (admin.Testimoni, error) {
	r, err := s.Q.BuatTestimoni(ctx, db.BuatTestimoniParams{Nama: in.Nama, Teks: in.Teks, Posisi: in.Posisi, Foto: fotoPg(in.Foto), Oleh: oleh})
	if adalahUnik(err, "testimoni_foto_key") {
		return admin.Testimoni{}, admin.ErrGambarDipakai
	}
	if err != nil {
		return admin.Testimoni{}, fmt.Errorf("buat testimoni: %w", err)
	}
	return testimoniAdmin(r), nil
}

func (s RepoTestimoni) Ubah(ctx context.Context, id int64, version int32, in admin.InputTestimoni, oleh string) (admin.Testimoni, error) {
	r, err := s.Q.UbahTestimoni(ctx, db.UbahTestimoniParams{
		Nama: in.Nama, Teks: in.Teks, Posisi: in.Posisi, Foto: fotoPg(in.Foto), Oleh: oleh, ID: id, Version: version,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Testimoni{}, s.klasifikasi(ctx, id, aturanDraftBerversi, version)
	}
	if adalahUnik(err, "testimoni_foto_key") {
		return admin.Testimoni{}, admin.ErrGambarDipakai
	}
	if err != nil {
		return admin.Testimoni{}, fmt.Errorf("ubah testimoni: %w", err)
	}
	return testimoniAdmin(r), nil
}

func (s RepoTestimoni) Post(ctx context.Context, id int64, version int32, oleh string) (admin.Testimoni, error) {
	r, err := s.Q.PostTestimoni(ctx, db.PostTestimoniParams{Oleh: oleh, ID: id, Version: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Testimoni{}, s.klasifikasi(ctx, id, aturanDraftBerversi, version)
	}
	if err != nil {
		return admin.Testimoni{}, fmt.Errorf("post testimoni: %w", err)
	}
	return testimoniAdmin(r), nil
}

func (s RepoTestimoni) Retract(ctx context.Context, id int64, version int32, oleh string) (admin.Testimoni, error) {
	r, err := s.Q.RetractTestimoni(ctx, db.RetractTestimoniParams{Oleh: oleh, ID: id, Version: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Testimoni{}, s.klasifikasi(ctx, id, aturanPostedBerversi, version)
	}
	if err != nil {
		return admin.Testimoni{}, fmt.Errorf("retract testimoni: %w", err)
	}
	return testimoniAdmin(r), nil
}

func (s RepoTestimoni) Hapus(ctx context.Context, id int64, oleh string) error {
	_, err := s.Q.HapusTestimoni(ctx, db.HapusTestimoniParams{Oleh: oleh, ID: id})
	if errors.Is(err, pgx.ErrNoRows) {
		return s.klasifikasi(ctx, id, aturanDraftAktif, 0)
	}
	if err != nil {
		return fmt.Errorf("hapus testimoni: %w", err)
	}
	return nil
}

func (s RepoTestimoni) Pulihkan(ctx context.Context, id int64, oleh string) (admin.Testimoni, error) {
	r, err := s.Q.PulihkanTestimoni(ctx, db.PulihkanTestimoniParams{Oleh: oleh, ID: id})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Testimoni{}, s.klasifikasi(ctx, id, aturanDiTrash, 0)
	}
	if err != nil {
		return admin.Testimoni{}, fmt.Errorf("pulihkan testimoni: %w", err)
	}
	return testimoniAdmin(r), nil
}

func (s RepoTestimoni) HapusPermanen(ctx context.Context, id int64) (string, error) {
	foto, err := s.Q.HapusPermanenTestimoni(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", s.klasifikasi(ctx, id, aturanDiTrash, 0)
	}
	if err != nil {
		return "", fmt.Errorf("hapus permanen testimoni: %w", err)
	}
	return teks(foto), nil
}

func (s RepoTestimoni) FotoDipakai(ctx context.Context, key string) (bool, error) {
	ada, err := s.Q.FotoTestimoniDipakai(ctx, fotoPg(key))
	if err != nil {
		return false, fmt.Errorf("periksa pemakaian foto testimoni: %w", err)
	}
	return ada, nil
}
