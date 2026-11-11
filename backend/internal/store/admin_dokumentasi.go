package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store/db"
)

var _ admin.DokumentasiRepo = RepoDokumentasi{}

// RepoDokumentasi adalah akses data dokumentasi untuk admin (adaptor di atas Store).
type RepoDokumentasi struct{ *Store }

// AdminDokumentasi mengembalikan repositori dokumentasi untuk admin.
func (s *Store) AdminDokumentasi() RepoDokumentasi { return RepoDokumentasi{s} }

const constraintKeyGambarDok = "dokumentasi_gambar_key_key"

func dokumentasiAdmin(r db.Dokumentasi, gambar []db.ListGambarDokumentasiRow) admin.Dokumentasi {
	d := admin.Dokumentasi{
		ID: r.ID, Slug: r.Slug, Judul: r.Judul, TanggalPublikasi: tanggal(r.TanggalPublikasi), Status: r.Status, Version: r.Version,
		DibuatPada: r.CreatedAt.Time.UTC(), DibuatOleh: r.CreatedBy, DiubahOleh: r.UpdatedBy,
		Gambar: make([]admin.GambarDokumentasi, 0, len(gambar)),
	}
	if r.DeletedAt.Valid {
		t := r.DeletedAt.Time.UTC()
		d.DihapusPada = &t
	}
	for _, g := range gambar {
		d.Gambar = append(d.Gambar, admin.GambarDokumentasi{Key: g.Key, Alt: teks(g.Alt)})
	}
	return d
}

func (s RepoDokumentasi) klasifikasi(ctx context.Context, id int64, a aturan, version int32) error {
	r, err := s.Q.GetDokumentasiAdmin(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.ErrTidakDitemukan
	}
	if err != nil {
		return fmt.Errorf("klasifikasi dokumentasi: %w", err)
	}
	return putuskanKlasifikasi(a, r.DeletedAt.Valid, r.Status, r.Version, version)
}

func (s RepoDokumentasi) Daftar(ctx context.Context, p admin.ParamDaftar) ([]admin.RingkasDokumentasi, int64, error) {
	baris, err := s.Q.ListDokumentasiAdmin(ctx, db.ListDokumentasiAdminParams{Filter: p.Filter, Pola: p.Pola, Batas: p.Batas, Geser: p.Geser})
	if err != nil {
		return nil, 0, fmt.Errorf("daftar dokumentasi admin: %w", err)
	}
	total, err := s.Q.HitungDokumentasiAdmin(ctx, db.HitungDokumentasiAdminParams{Filter: p.Filter, Pola: p.Pola})
	if err != nil {
		return nil, 0, fmt.Errorf("hitung dokumentasi admin: %w", err)
	}
	out := make([]admin.RingkasDokumentasi, 0, len(baris))
	for _, r := range baris {
		x := admin.RingkasDokumentasi{
			ID: r.ID, Slug: r.Slug, Judul: r.Judul, TanggalPublikasi: tanggal(r.TanggalPublikasi), GambarUtamaKey: teks(r.GambarUtama),
			JumlahGambar: int(r.JumlahGambar), Status: r.Status, Version: r.Version,
			DibuatPada: r.CreatedAt.Time.UTC(), DibuatOleh: r.CreatedBy, DiubahOleh: r.UpdatedBy,
		}
		if r.DeletedAt.Valid {
			t := r.DeletedAt.Time.UTC()
			x.DihapusPada = &t
		}
		out = append(out, x)
	}
	return out, total, nil
}

func (s RepoDokumentasi) Detail(ctx context.Context, id int64) (admin.Dokumentasi, error) {
	r, err := s.Q.GetDokumentasiAdmin(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Dokumentasi{}, admin.ErrTidakDitemukan
	}
	if err != nil {
		return admin.Dokumentasi{}, fmt.Errorf("detail dokumentasi admin: %w", err)
	}
	g, err := s.Q.ListGambarDokumentasi(ctx, id)
	if err != nil {
		return admin.Dokumentasi{}, fmt.Errorf("gambar dokumentasi admin: %w", err)
	}
	return dokumentasiAdmin(r, g), nil
}

func (s RepoDokumentasi) Buat(ctx context.Context, slug string, in admin.InputDokumentasi, oleh string) (admin.Dokumentasi, bool, error) {
	tgl, err := tanggalPg(in.TanggalPublikasi)
	if err != nil {
		return admin.Dokumentasi{}, false, err
	}
	var id int64
	dibuat := false
	err = s.dalamTx(ctx, func(q *db.Queries) error {
		r, err := q.BuatDokumentasi(ctx, db.BuatDokumentasiParams{Slug: slug, Judul: in.Judul, TanggalPublikasi: tgl, Oleh: oleh})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // slug sudah dipakai; tidak ada yang tersimpan
		}
		if err != nil {
			return err
		}
		id, dibuat = r.ID, true
		for i, g := range in.Gambar {
			if err := q.TambahGambarDokumentasi(ctx, db.TambahGambarDokumentasiParams{
				DokumentasiID: r.ID, Key: g.Key, Urutan: int16(i + 1), Alt: altPg(g.Alt),
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if adalahUnik(err, constraintKeyGambarDok) {
		return admin.Dokumentasi{}, false, admin.ErrGambarDipakai
	}
	if err != nil {
		return admin.Dokumentasi{}, false, fmt.Errorf("buat dokumentasi: %w", err)
	}
	if !dibuat {
		return admin.Dokumentasi{}, false, nil
	}
	d, err := s.Detail(ctx, id)
	return d, true, err
}

func (s RepoDokumentasi) Ubah(ctx context.Context, id int64, version int32, in admin.InputDokumentasi, oleh string) (admin.Dokumentasi, error) {
	tgl, err := tanggalPg(in.TanggalPublikasi)
	if err != nil {
		return admin.Dokumentasi{}, err
	}
	keys := make([]string, 0, len(in.Gambar))
	for _, g := range in.Gambar {
		keys = append(keys, g.Key)
	}
	err = s.dalamTx(ctx, func(q *db.Queries) error {
		if _, err := q.UbahDokumentasi(ctx, db.UbahDokumentasiParams{
			Judul: in.Judul, TanggalPublikasi: tgl, Oleh: oleh, ID: id, Version: version,
		}); err != nil {
			return err // pgx.ErrNoRows diklasifikasi di luar transaksi
		}
		// 1. buang gambar yang tidak ada dalam daftar akhir
		if err := q.HapusGambarDokumentasiTakDipertahankan(ctx, db.HapusGambarDokumentasiTakDipertahankanParams{DokumentasiID: id, Dipertahankan: keys}); err != nil {
			return err
		}
		// 2. perbarui urutan dan alt gambar yang sudah ada; sisipkan yang baru (urutan = posisi dalam daftar, 1..n).
		//    UNIQUE (dokumentasi_id, urutan) deferred, jadi pertukaran urutan sah selama hasil akhirnya unik.
		for i, g := range in.Gambar {
			urutan := int16(i + 1)
			n, err := q.PerbaruiGambarDokumentasi(ctx, db.PerbaruiGambarDokumentasiParams{Urutan: urutan, Alt: altPg(g.Alt), DokumentasiID: id, Key: g.Key})
			if err != nil {
				return err
			}
			if n == 0 {
				if err := q.TambahGambarDokumentasi(ctx, db.TambahGambarDokumentasiParams{DokumentasiID: id, Key: g.Key, Urutan: urutan, Alt: altPg(g.Alt)}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Dokumentasi{}, s.klasifikasi(ctx, id, aturanDraftBerversi, version)
	}
	if adalahUnik(err, constraintKeyGambarDok) {
		return admin.Dokumentasi{}, admin.ErrGambarDipakai
	}
	if err != nil {
		return admin.Dokumentasi{}, fmt.Errorf("ubah dokumentasi: %w", err)
	}
	return s.Detail(ctx, id)
}

func (s RepoDokumentasi) Post(ctx context.Context, id int64, version int32, oleh string) (admin.Dokumentasi, error) {
	_, err := s.Q.PostDokumentasi(ctx, db.PostDokumentasiParams{Oleh: oleh, ID: id, Version: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Dokumentasi{}, s.klasifikasi(ctx, id, aturanDraftBerversi, version)
	}
	if err != nil {
		return admin.Dokumentasi{}, fmt.Errorf("post dokumentasi: %w", err)
	}
	return s.Detail(ctx, id)
}

func (s RepoDokumentasi) Retract(ctx context.Context, id int64, version int32, oleh string) (admin.Dokumentasi, error) {
	_, err := s.Q.RetractDokumentasi(ctx, db.RetractDokumentasiParams{Oleh: oleh, ID: id, Version: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Dokumentasi{}, s.klasifikasi(ctx, id, aturanPostedBerversi, version)
	}
	if err != nil {
		return admin.Dokumentasi{}, fmt.Errorf("retract dokumentasi: %w", err)
	}
	return s.Detail(ctx, id)
}

func (s RepoDokumentasi) Hapus(ctx context.Context, id int64, oleh string) error {
	_, err := s.Q.HapusDokumentasi(ctx, db.HapusDokumentasiParams{Oleh: oleh, ID: id})
	if errors.Is(err, pgx.ErrNoRows) {
		return s.klasifikasi(ctx, id, aturanDraftAktif, 0)
	}
	if err != nil {
		return fmt.Errorf("hapus dokumentasi: %w", err)
	}
	return nil
}

func (s RepoDokumentasi) Pulihkan(ctx context.Context, id int64, oleh string) (admin.Dokumentasi, error) {
	_, err := s.Q.PulihkanDokumentasi(ctx, db.PulihkanDokumentasiParams{Oleh: oleh, ID: id})
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.Dokumentasi{}, s.klasifikasi(ctx, id, aturanDiTrash, 0)
	}
	if err != nil {
		return admin.Dokumentasi{}, fmt.Errorf("pulihkan dokumentasi: %w", err)
	}
	return s.Detail(ctx, id)
}

func (s RepoDokumentasi) HapusPermanen(ctx context.Context, id int64) ([]string, error) {
	var keys []string
	err := s.dalamTx(ctx, func(q *db.Queries) error {
		// daftar key dibaca SEBELUM penghapusan (baris gambar ikut hilang lewat ON DELETE CASCADE)
		gambar, err := q.ListGambarDokumentasi(ctx, id)
		if err != nil {
			return err
		}
		if _, err := q.HapusPermanenDokumentasi(ctx, id); err != nil {
			return err // pgx.ErrNoRows diklasifikasi di luar transaksi
		}
		keys = make([]string, 0, len(gambar))
		for _, g := range gambar {
			keys = append(keys, g.Key)
		}
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, s.klasifikasi(ctx, id, aturanDiTrash, 0)
	}
	if err != nil {
		return nil, fmt.Errorf("hapus permanen dokumentasi: %w", err)
	}
	return keys, nil
}

func (s RepoDokumentasi) GambarDipakai(ctx context.Context, key string) (bool, error) {
	ada, err := s.Q.GambarDokumentasiDipakai(ctx, key)
	if err != nil {
		return false, fmt.Errorf("periksa pemakaian gambar dokumentasi: %w", err)
	}
	return ada, nil
}
