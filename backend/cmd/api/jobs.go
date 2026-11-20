package main

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/config"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/jobs"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/store"
)

// bangunJob merakit job berkala: hapus permanen trash kedaluwarsa (berita, dokumentasi, testimoni; tim tidak punya
// trash) dan pembersihan file yatim. Penghapusan trash memakai layanan admin yang sama dengan hapus manual.
func bangunJob(cfg config.Config, log *slog.Logger, rdb redis.Cmdable, st *store.Store, gudang *media.S3Store,
	berita admin.BeritaLayanan, dokumentasi admin.DokumentasiLayanan, testimoni admin.TestimoniLayanan) *jobs.Job {
	sumber := func(nama string, hapus func(ctx context.Context, id int64) error) jobs.Sumber {
		return jobs.Sumber{
			Nama: nama,
			Kedaluwarsa: func(ctx context.Context, hari, batas int32) ([]int64, error) {
				return st.TrashKedaluwarsa(ctx, nama, hari, batas)
			},
			HapusPermanen: hapus,
		}
	}
	return &jobs.Job{
		Log: log, Redis: rdb, Gudang: gudang, Dipakai: st.KeyDipakai,
		HitungLog: st.HitungLogKedaluwarsa, HapusLog: st.HapusLogKedaluwarsa,
		Sumber: []jobs.Sumber{
			sumber("berita", berita.HapusPermanen),
			sumber("dokumentasi", dokumentasi.HapusPermanen),
			sumber("testimoni", testimoni.HapusPermanen),
		},
		Cfg: jobs.Config{
			Interval:         cfg.JobInterval,
			LockTTL:          cfg.JobLockTTL,
			RetensiTrashHari: int32(cfg.TrashRetentionDays),
			RetensiLogBulan:  int32(cfg.LogRetentionMonths),
			UmurMinYatim:     cfg.OrphanMinAge,
			MaksHapusYatim:   cfg.OrphanMaxDelete,
			DryRun:           cfg.JobsDryRun,
		},
	}
}
