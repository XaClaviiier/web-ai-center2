package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store/db"
)

var _ admin.TimRepo = RepoTim{}

// RepoTim adalah akses data tim untuk admin (adaptor di atas Store).
type RepoTim struct{ *Store }

// AdminTim mengembalikan repositori tim untuk admin.
func (s *Store) AdminTim() RepoTim { return RepoTim{s} }

func timAdmin(r db.Tim) admin.Tim {
	return admin.Tim{
		ID: r.ID, Nama: r.Nama, Posisi: r.Posisi, BidangRiset: r.BidangRiset, FotoKey: r.Foto, Urutan: r.Urutan,
		Version: r.Version, DibuatPada: r.CreatedAt.Time.UTC(), DibuatOleh: r.CreatedBy, DiubahOleh: r.UpdatedBy,
	}
}

func (s RepoTim) Daftar(ctx context.Context) ([]admin.Tim, error) {
	baris, err := s.Q.ListTimAdmin(ctx)
	if err != nil {
		return nil, fmt.Errorf("daftar tim admin: %w", err)
	}
	out := make([]admin.Tim, 0, len(baris))
	for _, r := range baris {
		out = append(out, timAdmin(r))
	}
	return out, nil
}

func (s RepoTim) Detail(ctx context.Context, id int64) (admin.Tim, error) {
	r, err := s.Q.GetTimAdmin(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Tim{}, admin.ErrTidakDitemukan
	}
	if err != nil {
		return admin.Tim{}, fmt.Errorf("detail tim admin: %w", err)
	}
	return timAdmin(r), nil
}

func (s RepoTim) Buat(ctx context.Context, in admin.InputTim, oleh string) (admin.Tim, error) {
	r, err := s.Q.BuatTim(ctx, db.BuatTimParams{Nama: in.Nama, Posisi: in.Posisi, BidangRiset: in.BidangRiset, Foto: in.Foto, Oleh: oleh})
	if adalahUnik(err, "tim_foto_key") {
		return admin.Tim{}, admin.ErrGambarDipakai
	}
	if err != nil {
		return admin.Tim{}, fmt.Errorf("buat tim: %w", err)
	}
	return timAdmin(r), nil
}

func (s RepoTim) Ubah(ctx context.Context, id int64, version int32, in admin.InputTim, oleh string) (admin.Tim, error) {
	r, err := s.Q.UbahTim(ctx, db.UbahTimParams{
		Nama: in.Nama, Posisi: in.Posisi, BidangRiset: in.BidangRiset, Foto: in.Foto, Oleh: oleh, ID: id, Version: version,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		cur, gerr := s.Q.GetTimAdmin(ctx, id)
		if errors.Is(gerr, pgx.ErrNoRows) {
			return admin.Tim{}, admin.ErrTidakDitemukan
		}
		if gerr != nil {
			return admin.Tim{}, fmt.Errorf("klasifikasi tim: %w", gerr)
		}
		return admin.Tim{}, putuskanKlasifikasi(aturanVersiSaja, false, "", cur.Version, version)
	}
	if adalahUnik(err, "tim_foto_key") {
		return admin.Tim{}, admin.ErrGambarDipakai
	}
	if err != nil {
		return admin.Tim{}, fmt.Errorf("ubah tim: %w", err)
	}
	return timAdmin(r), nil
}

func (s RepoTim) Hapus(ctx context.Context, id int64) (string, error) {
	key, err := s.Q.HapusTim(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", admin.ErrTidakDitemukan
	}
	if err != nil {
		return "", fmt.Errorf("hapus tim: %w", err)
	}
	return key, nil
}

func (s RepoTim) FotoDipakai(ctx context.Context, key string) (bool, error) {
	ada, err := s.Q.FotoTimDipakai(ctx, key)
	if err != nil {
		return false, fmt.Errorf("periksa pemakaian foto tim: %w", err)
	}
	return ada, nil
}
