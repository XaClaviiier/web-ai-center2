// Package jobs menjalankan pekerjaan berkala di proses API: hapus permanen item trash yang kedaluwarsa dan
// pembersihan file gambar yatim di storage (docs/database.md bagian 5, docs/requirement.md bagian 5).
//
// Satu instance saja yang bekerja pada satu waktu: setiap siklus mengambil kunci Redis "lock:purge".
// Pekerjaan sebagian aman dihentikan kapan saja karena tiap item bersifat atomik.
package jobs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
)

const (
	// KunciJob adalah key Redis yang menjaga agar hanya satu instance menjalankan siklus.
	KunciJob = "lock:purge"

	tundaAwalDefault   = 2 * time.Minute
	maksTrashPerSumber = 100 // item per sumber per siklus; sisanya siklus berikutnya

	// Pemutus: pembersihan yatim dibatalkan total bila kandidat melebihi persen ini dari objek berumur di atas
	// batas DAN lebih dari minimum ini. Mencegah penghapusan massal bila aplikasi salah diarahkan ke database kosong.
	pemutusPersen  = 50
	pemutusMinimum = 20
)

// ErrPemutus berarti pembersihan yatim dibatalkan karena jumlah kandidat tidak masuk akal.
var ErrPemutus = errors.New("jumlah file yatim tidak wajar; pembersihan dibatalkan")

// prefix objek yang boleh disentuh job; objek di luar prefix ini tidak pernah didaftar atau dihapus.
var prefixKonten = []string{"berita/", "dokumentasi/", "testimoni/", "tim/"}

// Gudang adalah storage gambar yang bisa didaftar dan dihapus.
type Gudang interface {
	media.Pendaftar
	Hapus(ctx context.Context, key string) error
}

// Sumber adalah satu jenis konten yang punya trash.
type Sumber struct {
	Nama string
	// Kedaluwarsa mengembalikan id item di trash yang lebih tua dari hari, paling banyak batas.
	Kedaluwarsa func(ctx context.Context, hari, batas int32) ([]int64, error)
	// HapusPermanen menghapus baris dan file gambarnya (jalur yang sama dengan hapus manual).
	HapusPermanen func(ctx context.Context, id int64) error
}

// Config adalah pengaturan job.
type Config struct {
	Interval         time.Duration
	TundaAwal        time.Duration // siklus pertama setelah start; default 2 menit
	LockTTL          time.Duration
	BatasSiklus      time.Duration // batas waktu satu siklus; harus lebih pendek dari LockTTL
	RetensiTrashHari int32
	UmurMinYatim     time.Duration
	MaksHapusYatim   int
	DryRun           bool
}

// Job menjalankan siklus berkala.
type Job struct {
	Log     *slog.Logger
	Redis   redis.Cmdable
	Gudang  Gudang
	Sumber  []Sumber
	Dipakai func(ctx context.Context, key string) (bool, error) // key dipakai baris mana pun?
	Cfg     Config

	// Sekarang bisa diganti di tes; default time.Now.
	Sekarang func() time.Time
}

func (j *Job) sekarang() time.Time {
	if j.Sekarang != nil {
		return j.Sekarang()
	}
	return time.Now()
}

// Jalankan memblokir sampai ctx dibatalkan: siklus pertama setelah TundaAwal, lalu tiap Interval.
func (j *Job) Jalankan(ctx context.Context) {
	tunda := j.Cfg.TundaAwal
	if tunda <= 0 {
		tunda = tundaAwalDefault
	}
	j.Log.Info("job berkala aktif", "interval", j.Cfg.Interval.String(), "dry_run", j.Cfg.DryRun, "retensi_trash_hari", j.Cfg.RetensiTrashHari)
	timer := time.NewTimer(tunda)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			j.Siklus(ctx)
			timer.Reset(j.Cfg.Interval)
		}
	}
}

// Siklus menjalankan satu putaran bila kunci Redis berhasil diambil; selain itu dilewati (dicatat di log).
func (j *Job) Siklus(ctx context.Context) {
	token, ok, err := j.ambilKunci(ctx)
	if err != nil {
		j.Log.Error("job dilewati: Redis bermasalah", "err", err)
		return
	}
	if !ok {
		j.Log.Info("job dilewati: instance lain sedang bekerja")
		return
	}
	defer j.lepasKunci(token)

	batas := j.Cfg.BatasSiklus
	if batas <= 0 {
		batas = j.Cfg.LockTTL * 5 / 6
	}
	ctx, batal := context.WithTimeout(ctx, batas)
	defer batal()

	mulai := j.sekarang()
	dihapusTrash := j.hapusTrash(ctx)
	dihapusYatim, err := j.bersihkanYatim(ctx)
	switch {
	case errors.Is(err, ErrPemutus):
		j.Log.Error("pembersihan file yatim dibatalkan oleh pemutus", "err", err)
	case err != nil:
		j.Log.Error("pembersihan file yatim gagal", "err", err)
	}
	j.Log.Info("siklus job selesai", "trash_dihapus", dihapusTrash, "yatim_dihapus", dihapusYatim,
		"dry_run", j.Cfg.DryRun, "durasi", j.sekarang().Sub(mulai).String())
}

func (j *Job) ambilKunci(ctx context.Context) (token string, ok bool, err error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", false, err
	}
	token = hex.EncodeToString(b)
	ok, err = j.Redis.SetNX(ctx, KunciJob, token, j.Cfg.LockTTL).Result()
	if err != nil {
		return "", false, fmt.Errorf("ambil kunci %s: %w", KunciJob, err)
	}
	return token, ok, nil
}

var skrinLepasKunci = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then return redis.call('DEL', KEYS[1]) end
return 0`)

// lepasKunci menghapus kunci hanya bila masih milik siklus ini (aman walau TTL sempat habis).
func (j *Job) lepasKunci(token string) {
	ctx, batal := context.WithTimeout(context.Background(), 5*time.Second)
	defer batal()
	if err := skrinLepasKunci.Run(ctx, j.Redis, []string{KunciJob}, token).Err(); err != nil {
		j.Log.Error("gagal melepas kunci job", "err", err)
	}
}

// hapusTrash menghapus permanen item trash yang lebih tua dari retensi. Galat satu item tidak menghentikan yang lain.
func (j *Job) hapusTrash(ctx context.Context) int {
	total := 0
	for _, s := range j.Sumber {
		ids, err := s.Kedaluwarsa(ctx, j.Cfg.RetensiTrashHari, maksTrashPerSumber)
		if err != nil {
			j.Log.Error("gagal membaca trash kedaluwarsa", "sumber", s.Nama, "err", err)
			continue
		}
		for _, id := range ids {
			if ctx.Err() != nil {
				return total
			}
			if j.Cfg.DryRun {
				j.Log.Info("dry-run: trash akan dihapus permanen", "sumber", s.Nama, "id", id)
				continue
			}
			err := s.HapusPermanen(ctx, id)
			switch {
			case err == nil:
				total++
				j.Log.Info("trash kedaluwarsa dihapus permanen", "sumber", s.Nama, "id", id)
			case errors.Is(err, admin.ErrTidakDitemukan), errors.Is(err, admin.ErrStatusTidakValid):
				// dihapus atau dipulihkan pihak lain di antara pembacaan dan penghapusan: wajar
				j.Log.Info("trash dilewati: sudah berubah", "sumber", s.Nama, "id", id)
			default:
				j.Log.Error("gagal menghapus trash kedaluwarsa", "sumber", s.Nama, "id", id, "err", err)
			}
		}
	}
	return total
}

// bersihkanYatim menghapus objek storage yang tidak dipakai baris mana pun dan lebih tua dari UmurMinYatim.
// Hanya prefix konten yang disentuh. Dibatalkan total (ErrPemutus) bila kandidat tidak wajar banyak.
func (j *Job) bersihkanYatim(ctx context.Context) (int, error) {
	var tua []media.Info
	batasUmur := j.sekarang().Add(-j.Cfg.UmurMinYatim)
	for _, p := range prefixKonten {
		infos, err := j.Gudang.Daftar(ctx, p)
		if err != nil {
			return 0, fmt.Errorf("daftar objek %q: %w", p, err)
		}
		for _, o := range infos {
			if strings.HasPrefix(o.Key, p) && o.Diubah.Before(batasUmur) {
				tua = append(tua, o)
			}
		}
	}

	var kandidat []media.Info
	for _, o := range tua {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		dipakai, err := j.Dipakai(ctx, o.Key)
		if err != nil {
			return 0, fmt.Errorf("periksa pemakaian %q: %w", o.Key, err) // jangan hapus apa pun bila DB bermasalah
		}
		if !dipakai {
			kandidat = append(kandidat, o)
		}
	}
	if len(kandidat) > pemutusMinimum && len(kandidat)*100 > len(tua)*pemutusPersen {
		return 0, fmt.Errorf("%w: %d dari %d objek", ErrPemutus, len(kandidat), len(tua))
	}

	dihapus := 0
	for _, o := range kandidat {
		if ctx.Err() != nil {
			return dihapus, ctx.Err()
		}
		if j.Cfg.MaksHapusYatim > 0 && dihapus >= j.Cfg.MaksHapusYatim {
			j.Log.Warn("batas hapus yatim per siklus tercapai; sisanya siklus berikutnya", "batas", j.Cfg.MaksHapusYatim, "kandidat", len(kandidat))
			break
		}
		// Periksa ulang tepat sebelum menghapus untuk mempersempit celah balapan dengan simpan yang sedang berjalan.
		dipakai, err := j.Dipakai(ctx, o.Key)
		if err != nil {
			return dihapus, fmt.Errorf("periksa ulang %q: %w", o.Key, err)
		}
		if dipakai {
			continue
		}
		if j.Cfg.DryRun {
			j.Log.Info("dry-run: file yatim akan dihapus", "key", o.Key, "ukuran", o.Ukuran, "diubah", o.Diubah)
			continue
		}
		if err := j.Gudang.Hapus(ctx, o.Key); err != nil {
			j.Log.Error("gagal menghapus file yatim", "key", o.Key, "err", err)
			continue
		}
		dihapus++
		j.Log.Info("file yatim dihapus", "key", o.Key, "ukuran", o.Ukuran)
	}
	return dihapus, nil
}
