package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/auth"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store/db"
)

// aturan menyatakan syarat status sebuah aksi admin; dipakai menjelaskan mengapa UPDATE/DELETE bersyarat tidak mengenai baris.
type aturan int

const (
	aturanDraftBerversi  aturan = iota // ubah, post: harus draft, belum di-trash, version cocok
	aturanPostedBerversi               // retract: harus posted, belum di-trash, version cocok
	aturanDraftAktif                   // hapus: harus draft dan belum di-trash
	aturanDiTrash                      // pulihkan, hapus permanen: harus di-trash
	aturanVersiSaja                    // tim: hanya version yang dicek (tanpa status)
)

// putuskanKlasifikasi memilih galat domain dari keadaan baris yang ada. Status yang salah didahulukan daripada version
// yang salah. Bila baris ternyata memenuhi syarat, berarti berubah di antara dua kueri (balapan): dianggap bentrok
// agar klien memuat ulang.
func putuskanKlasifikasi(a aturan, dihapus bool, status string, versiBaris, versiDiminta int32) error {
	switch a {
	case aturanDraftBerversi:
		if dihapus || status != "draft" {
			return admin.ErrStatusTidakValid
		}
		if versiBaris != versiDiminta {
			return admin.ErrVersiBentrok
		}
	case aturanPostedBerversi:
		if dihapus || status != "posted" {
			return admin.ErrStatusTidakValid
		}
		if versiBaris != versiDiminta {
			return admin.ErrVersiBentrok
		}
	case aturanDraftAktif:
		if dihapus || status != "draft" {
			return admin.ErrStatusTidakValid
		}
	case aturanDiTrash:
		if !dihapus {
			return admin.ErrStatusTidakValid
		}
	case aturanVersiSaja:
		if versiBaris != versiDiminta {
			return admin.ErrVersiBentrok
		}
	}
	return admin.ErrVersiBentrok
}

// fotoPg mengubah key menjadi kolom nullable; string kosong menjadi NULL.
func fotoPg(key string) pgtype.Text { return pgtype.Text{String: key, Valid: key != ""} }

// dalamTx menjalankan f dalam satu transaksi: commit bila f sukses, rollback bila galat (atau panic). Galat dari f
// dikembalikan apa adanya (dibungkus bila commit gagal) sehingga pemanggil tetap dapat mengenali galat PostgreSQL.
func (s *Store) dalamTx(ctx context.Context, f func(q *db.Queries) error) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("memulai transaksi: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // tidak berefek setelah Commit
	if a, ok := auth.AktorDariContext(ctx); ok {
		if err := setAktor(ctx, tx, a); err != nil { // trigger log aktivitas memakai aktor ini
			return err
		}
	}
	if err := f(s.Q.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaksi: %w", err)
	}
	return nil
}
