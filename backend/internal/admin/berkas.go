package admin

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
)

const waktuOperasiFile = 15 * time.Second

// pengelolaFile menerapkan aturan file gambar yang sama untuk semua jenis konten (database.md §5):
// validasi key dari klien, hapus file lama SETELAH commit, dan hapus file baru yang gagal disimpan (dengan pengaman).
type pengelolaFile struct {
	Media  media.Store
	Log    *slog.Logger
	Sumber string // media.SumberBerita, SumberTim, SumberTestimoni, ...
	// Dipakai melaporkan apakah key dipakai baris mana pun pada jenis konten ini.
	Dipakai func(ctx context.Context, key string) (bool, error)
}

// periksaKey memastikan key dari klien berbentuk buatan kita, bersumber sama dengan jenis konten, dan objeknya ada di storage.
func (p pengelolaFile) periksaKey(ctx context.Context, key string) error {
	sumber, ok := media.SumberDariKey(key)
	if !ok || sumber != p.Sumber {
		return salah("Field gambar bukan key gambar " + p.Sumber + " yang sah")
	}
	ada, err := media.Ada(ctx, p.Media, key)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStorage, err)
	}
	if !ada {
		return salah("Gambar belum diunggah atau sudah dihapus; unggah ulang")
	}
	return nil
}

// ada memeriksa bahwa objek ada (untuk Post); kegagalan storage dibungkus ErrStorage.
func (p pengelolaFile) ada(ctx context.Context, key string) (bool, error) {
	ada, err := media.Ada(ctx, p.Media, key)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrStorage, err)
	}
	return ada, nil
}

// bersihkanBaru menghapus file baru yang gagal disimpan (database.md §5: gagal sebelum commit = file baru dihapus kembali).
// Pengaman: hanya key sah bersumber sama, bukan key lama, ada di storage, dan TIDAK dipakai baris mana pun.
func (p pengelolaFile) bersihkanBaru(key, lama string) {
	if key == "" || key == lama {
		return
	}
	if sumber, ok := media.SumberDariKey(key); !ok || sumber != p.Sumber {
		return
	}
	ctx, batal := context.WithTimeout(context.Background(), waktuOperasiFile)
	defer batal()
	dipakai, err := p.Dipakai(ctx, key)
	if err != nil || dipakai {
		return
	}
	if err := p.Media.Hapus(ctx, key); err != nil {
		p.Log.Warn("gagal menghapus gambar baru yang tidak jadi disimpan", "key", key, "err", err)
	}
}

// hapusLama menghapus file yang tidak lagi dipakai SETELAH commit; gagal hanya dicatat (job file yatim akan membersihkannya).
func (p pengelolaFile) hapusLama(key string) {
	if key == "" {
		return
	}
	ctx, batal := context.WithTimeout(context.Background(), waktuOperasiFile)
	defer batal()
	if err := p.Media.Hapus(ctx, key); err != nil {
		p.Log.Warn("gagal menghapus gambar lama setelah commit", "key", key, "err", err)
	}
}
