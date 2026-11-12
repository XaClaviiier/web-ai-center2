package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store/db"
)

var _ admin.LogRepo = RepoLog{}

// RepoLog adalah akses baca log aktivitas (adaptor di atas Store). Log ditulis trigger database, bukan oleh Go.
type RepoLog struct{ *Store }

// AdminLog mengembalikan repositori log aktivitas.
func (s *Store) AdminLog() RepoLog { return RepoLog{s} }

func waktuPg(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func (s RepoLog) Daftar(ctx context.Context, p admin.ParamRepoLog) ([]admin.EntriLog, error) {
	baris, err := s.Q.ListLogAktivitas(ctx, db.ListLogAktivitasParams{
		Username: p.Username, Jenis: p.Jenis, Dari: waktuPg(p.Dari), Sampai: waktuPg(p.Sampai), Batas: p.Batas, Geser: p.Geser,
	})
	if err != nil {
		return nil, fmt.Errorf("daftar log aktivitas: %w", err)
	}
	out := make([]admin.EntriLog, 0, len(baris))
	for _, r := range baris {
		out = append(out, admin.EntriLog{
			ID: r.ID, TerjadiPada: r.TerjadiPada.Time.UTC(), PenggunaSub: r.PenggunaSub, Username: r.Username,
			Jenis: r.Jenis, Aksi: r.Aksi, Modul: r.Modul, ItemID: r.ItemID, LabelItem: r.LabelItem,
		})
	}
	return out, nil
}

func (s RepoLog) Hitung(ctx context.Context, p admin.ParamRepoLog) (int64, error) {
	n, err := s.Q.HitungLogAktivitas(ctx, db.HitungLogAktivitasParams{
		Username: p.Username, Jenis: p.Jenis, Dari: waktuPg(p.Dari), Sampai: waktuPg(p.Sampai),
	})
	if err != nil {
		return 0, fmt.Errorf("hitung log aktivitas: %w", err)
	}
	return n, nil
}

func (s RepoLog) Usernames(ctx context.Context) ([]string, error) {
	nama, err := s.Q.ListUsernameLog(ctx)
	if err != nil {
		return nil, fmt.Errorf("username log aktivitas: %w", err)
	}
	if nama == nil {
		nama = []string{}
	}
	return nama, nil
}
