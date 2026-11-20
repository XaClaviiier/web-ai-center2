package store

import (
	"context"
	"fmt"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/store/db"
)

// TrashKedaluwarsa mengembalikan id item di trash yang sudah lebih dari hari hari (jam database), terlama dulu,
// paling banyak batas. sumber: "berita", "dokumentasi", atau "testimoni" (tim tidak punya trash).
func (s *Store) TrashKedaluwarsa(ctx context.Context, sumber string, hari, batas int32) ([]int64, error) {
	var (
		ids []int64
		err error
	)
	switch sumber {
	case "berita":
		ids, err = s.Q.TrashBeritaKedaluwarsa(ctx, db.TrashBeritaKedaluwarsaParams{Hari: hari, Batas: batas})
	case "dokumentasi":
		ids, err = s.Q.TrashDokumentasiKedaluwarsa(ctx, db.TrashDokumentasiKedaluwarsaParams{Hari: hari, Batas: batas})
	case "testimoni":
		ids, err = s.Q.TrashTestimoniKedaluwarsa(ctx, db.TrashTestimoniKedaluwarsaParams{Hari: hari, Batas: batas})
	default:
		return nil, fmt.Errorf("trash kedaluwarsa: sumber %q tidak dikenal", sumber)
	}
	if err != nil {
		return nil, fmt.Errorf("trash %s kedaluwarsa: %w", sumber, err)
	}
	return ids, nil
}

// KeyDipakai melaporkan apakah key gambar dipakai baris mana pun pada berita, dokumentasi, testimoni, atau tim
// (termasuk draft dan trash).
func (s *Store) KeyDipakai(ctx context.Context, key string) (bool, error) {
	ada, err := s.Q.KeyGambarDipakai(ctx, key)
	if err != nil {
		return false, fmt.Errorf("periksa pemakaian key: %w", err)
	}
	return ada, nil
}
