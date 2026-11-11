package jobs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/media"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

// pakaiKonteks: semua objek yang baru diunggah dianggap berumur 48 jam lewat jam palsu.
func jamMaju() time.Time { return time.Now().Add(48 * time.Hour) }

type lingkungan struct {
	j            *Job
	gudang       *media.Memori
	dipakai      map[string]bool
	rdb          *redis.Client
	galatDipakai error
}

func siapkan(t *testing.T) *lingkungan {
	t.Helper()
	e := &lingkungan{gudang: media.NewMemori(), dipakai: map[string]bool{}, rdb: testutil.RedisTes(t)}
	e.j = &Job{
		Log: slog.New(slog.NewTextHandler(io.Discard, nil)), Redis: e.rdb, Gudang: e.gudang,
		Dipakai: func(_ context.Context, key string) (bool, error) {
			if e.galatDipakai != nil {
				return false, e.galatDipakai
			}
			return e.dipakai[key], nil
		},
		Cfg: Config{
			Interval: time.Hour, LockTTL: 30 * time.Minute, RetensiTrashHari: 30,
			UmurMinYatim: 24 * time.Hour, MaksHapusYatim: 200,
		},
		Sekarang: jamMaju,
	}
	return e
}

func (e *lingkungan) taruh(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if err := e.gudang.Put(context.Background(), k, []byte("x"), "image/png"); err != nil {
			t.Fatal(err)
		}
	}
}

func (e *lingkungan) ada(key string) bool {
	_, err := e.gudang.Head(context.Background(), key)
	return err == nil
}

func TestYatimDihapusYangDipakaiDanDiLuarPrefixTidak(t *testing.T) {
	e := siapkan(t)
	e.taruh(t, "berita/yatim.png", "berita/dipakai.png", "dokumentasi/yatim.png", "testimoni/yatim.png", "tim/yatim.png", "lain/bukan-konten.png", "logo.png")
	e.dipakai["berita/dipakai.png"] = true

	n, err := e.j.bersihkanYatim(context.Background())
	if err != nil || n != 4 {
		t.Fatalf("hapus: n=%d err=%v", n, err)
	}
	for _, k := range []string{"berita/yatim.png", "dokumentasi/yatim.png", "testimoni/yatim.png", "tim/yatim.png"} {
		if e.ada(k) {
			t.Errorf("%s harus terhapus", k)
		}
	}
	for _, k := range []string{"berita/dipakai.png", "lain/bukan-konten.png", "logo.png"} {
		if !e.ada(k) {
			t.Errorf("%s tidak boleh terhapus", k)
		}
	}
}

func TestObjekMudaTidakDisentuh(t *testing.T) {
	e := siapkan(t)
	e.j.Sekarang = time.Now // objek baru diunggah = berumur ~0
	e.taruh(t, "berita/baru.png")
	n, err := e.j.bersihkanYatim(context.Background())
	if err != nil || n != 0 || !e.ada("berita/baru.png") {
		t.Errorf("objek muda harus aman: n=%d err=%v", n, err)
	}
}

func TestPemutusMembatalkanPenghapusanMassal(t *testing.T) {
	e := siapkan(t)
	for i := 0; i < 30; i++ {
		e.taruh(t, fmt.Sprintf("berita/%02d.png", i)) // semua yatim: tampak seperti database salah arah
	}
	n, err := e.j.bersihkanYatim(context.Background())
	if !errors.Is(err, ErrPemutus) || n != 0 || e.gudang.Jumlah() != 30 {
		t.Fatalf("pemutus harus aktif tanpa menghapus: n=%d err=%v sisa=%d", n, err, e.gudang.Jumlah())
	}

	// 5 yatim dari 30 (di bawah minimum 20 dan di bawah 50%) tetap dibersihkan.
	for i := 0; i < 25; i++ {
		e.dipakai[fmt.Sprintf("berita/%02d.png", i)] = true
	}
	n, err = e.j.bersihkanYatim(context.Background())
	if err != nil || n != 5 || e.gudang.Jumlah() != 25 {
		t.Errorf("pembersihan wajar: n=%d err=%v sisa=%d", n, err, e.gudang.Jumlah())
	}
}

func TestPemutusTidakAktifBilaKandidatSedikitMeskiSemuaYatim(t *testing.T) {
	e := siapkan(t)
	e.taruh(t, "berita/a.png", "berita/b.png") // 100% yatim tetapi hanya 2 (di bawah minimum)
	if n, err := e.j.bersihkanYatim(context.Background()); err != nil || n != 2 {
		t.Errorf("n=%d err=%v", n, err)
	}
}

func TestBatasHapusPerSiklus(t *testing.T) {
	e := siapkan(t)
	e.j.Cfg.MaksHapusYatim = 3
	for i := 0; i < 15; i++ {
		e.taruh(t, fmt.Sprintf("berita/y%02d.png", i))
	}
	for i := 0; i < 40; i++ { // banyak yang dipakai agar pemutus tidak aktif
		k := fmt.Sprintf("tim/p%02d.png", i)
		e.taruh(t, k)
		e.dipakai[k] = true
	}
	n, err := e.j.bersihkanYatim(context.Background())
	if err != nil || n != 3 || e.gudang.Jumlah() != 52 {
		t.Errorf("harus berhenti di 3: n=%d err=%v sisa=%d", n, err, e.gudang.Jumlah())
	}
}

func TestGalatDatabaseTidakMenghapusApaPun(t *testing.T) {
	e := siapkan(t)
	e.taruh(t, "berita/a.png", "berita/b.png")
	e.galatDipakai = errors.New("db mati")
	n, err := e.j.bersihkanYatim(context.Background())
	if err == nil || n != 0 || e.gudang.Jumlah() != 2 {
		t.Errorf("galat DB harus menghentikan tanpa menghapus: n=%d err=%v", n, err)
	}
}

func TestDryRunTidakMenghapusApaPun(t *testing.T) {
	e := siapkan(t)
	e.j.Cfg.DryRun = true
	e.taruh(t, "berita/a.png", "berita/b.png")
	var hapusDipanggil atomic.Int32
	e.j.Sumber = []Sumber{{
		Nama:        "berita",
		Kedaluwarsa: func(context.Context, int32, int32) ([]int64, error) { return []int64{1, 2}, nil },
		HapusPermanen: func(context.Context, int64) error {
			hapusDipanggil.Add(1)
			return nil
		},
	}}
	e.j.Siklus(context.Background())
	if e.gudang.Jumlah() != 2 || hapusDipanggil.Load() != 0 {
		t.Errorf("dry-run tidak boleh menghapus: objek=%d hapusTrash=%d", e.gudang.Jumlah(), hapusDipanggil.Load())
	}
}

func TestHapusTrashParameterGalatDanBatas(t *testing.T) {
	e := siapkan(t)
	var (
		mu                    sync.Mutex
		dihapus               []string
		hariDapat, batasDapat int32
	)
	catat := func(s string) func(ctx context.Context, id int64) error {
		return func(_ context.Context, id int64) error {
			mu.Lock()
			defer mu.Unlock()
			switch id {
			case 2:
				return admin.ErrTidakDitemukan // sudah dihapus pihak lain: wajar
			case 3:
				return admin.ErrStatusTidakValid // sudah dipulihkan: wajar
			case 4:
				return errors.New("storage mati") // tidak boleh menghentikan yang lain
			}
			dihapus = append(dihapus, fmt.Sprintf("%s:%d", s, id))
			return nil
		}
	}
	mk := func(nama string) Sumber {
		return Sumber{
			Nama: nama,
			Kedaluwarsa: func(_ context.Context, hari, batas int32) ([]int64, error) {
				hariDapat, batasDapat = hari, batas
				return []int64{1, 2, 3, 4, 5}, nil
			},
			HapusPermanen: catat(nama),
		}
	}
	e.j.Sumber = []Sumber{mk("berita"), mk("dokumentasi"),
		{Nama: "testimoni", Kedaluwarsa: func(context.Context, int32, int32) ([]int64, error) { return nil, errors.New("db") }, HapusPermanen: catat("testimoni")}}

	n := e.j.hapusTrash(context.Background())
	if n != 4 || strings.Join(dihapus, ",") != "berita:1,berita:5,dokumentasi:1,dokumentasi:5" {
		t.Errorf("n=%d dihapus=%v", n, dihapus)
	}
	if hariDapat != 30 || batasDapat != 100 {
		t.Errorf("parameter: hari=%d batas=%d", hariDapat, batasDapat)
	}
}

func TestSiklusMelepasKunciDanHanyaMilikSendiri(t *testing.T) {
	e := siapkan(t)
	e.j.Siklus(context.Background())
	if e.rdb.Exists(context.Background(), KunciJob).Val() != 0 {
		t.Error("kunci harus dilepas setelah siklus")
	}

	// TTL habis di tengah siklus lalu instance lain mengambil kunci: tidak boleh dihapus oleh siklus pertama.
	e.j.Sumber = []Sumber{{
		Nama:        "berita",
		Kedaluwarsa: func(context.Context, int32, int32) ([]int64, error) { return []int64{1}, nil },
		HapusPermanen: func(ctx context.Context, _ int64) error {
			return e.rdb.Set(ctx, KunciJob, "instance-lain", time.Minute).Err()
		},
	}}
	e.j.Siklus(context.Background())
	if v, _ := e.rdb.Get(context.Background(), KunciJob).Result(); v != "instance-lain" {
		t.Errorf("kunci milik instance lain tidak boleh dilepas, nilai %q", v)
	}
}

func TestSiklusDilewatiBilaKunciDipegangAtauRedisMati(t *testing.T) {
	e := siapkan(t)
	var jalan atomic.Int32
	e.j.Sumber = []Sumber{{
		Nama:        "berita",
		Kedaluwarsa: func(context.Context, int32, int32) ([]int64, error) { jalan.Add(1); return nil, nil },
	}}
	e.rdb.Set(context.Background(), KunciJob, "instance-lain", time.Minute)
	e.j.Siklus(context.Background())
	if jalan.Load() != 0 {
		t.Error("kunci dipegang instance lain: siklus harus dilewati")
	}

	mati := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 200 * time.Millisecond, MaxRetries: -1})
	t.Cleanup(func() { _ = mati.Close() })
	e.j.Redis = mati
	e.j.Siklus(context.Background())
	if jalan.Load() != 0 {
		t.Error("Redis mati: siklus harus dilewati")
	}
}

func TestDuaInstanceBersamaanHanyaSatuBekerja(t *testing.T) {
	e := siapkan(t)
	var jalan, serentak, maks atomic.Int32
	e.j.Sumber = []Sumber{{
		Nama: "berita",
		Kedaluwarsa: func(context.Context, int32, int32) ([]int64, error) {
			jalan.Add(1)
			n := serentak.Add(1)
			for {
				m := maks.Load()
				if n <= m || maks.CompareAndSwap(m, n) {
					break
				}
			}
			time.Sleep(200 * time.Millisecond)
			serentak.Add(-1)
			return nil, nil
		},
	}}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); e.j.Siklus(context.Background()) }()
	}
	wg.Wait()
	if jalan.Load() != 1 || maks.Load() != 1 {
		t.Errorf("hanya satu siklus yang boleh bekerja: dijalankan %d kali, serentak maks %d", jalan.Load(), maks.Load())
	}
}

func TestJalankanBerulangDanBerhentiBersamaKonteks(t *testing.T) {
	e := siapkan(t)
	e.j.Cfg.TundaAwal = 20 * time.Millisecond
	e.j.Cfg.Interval = 30 * time.Millisecond
	var jalan atomic.Int32
	e.j.Sumber = []Sumber{{
		Nama:        "berita",
		Kedaluwarsa: func(context.Context, int32, int32) ([]int64, error) { jalan.Add(1); return nil, nil },
	}}
	ctx, batal := context.WithCancel(context.Background())
	selesai := make(chan struct{})
	go func() { e.j.Jalankan(ctx); close(selesai) }()
	time.Sleep(400 * time.Millisecond)
	batal()
	select {
	case <-selesai:
	case <-time.After(3 * time.Second):
		t.Fatal("Jalankan harus berhenti saat konteks dibatalkan")
	}
	if jalan.Load() < 3 {
		t.Errorf("siklus harus berulang, dijalankan %d kali", jalan.Load())
	}
}
