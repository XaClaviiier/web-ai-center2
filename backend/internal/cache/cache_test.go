package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

func TestSimpanAmbilDanTTL(t *testing.T) {
	rdb := testutil.RedisTes(t)
	c := &Cache{C: rdb, TTL: 90 * time.Second}
	ctx := context.Background()
	k, err := c.Kunci(ctx, Berita, "list:p1:l9")
	if err != nil {
		t.Fatal(err)
	}
	if _, ada, err := c.Ambil(ctx, k); err != nil || ada {
		t.Fatalf("awalnya harus miss: ada=%v err=%v", ada, err)
	}
	if err := c.Simpan(ctx, k, []byte(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
	b, ada, err := c.Ambil(ctx, k)
	if err != nil || !ada || string(b) != `{"a":1}` {
		t.Fatalf("ambil: %q ada=%v err=%v", b, ada, err)
	}
	ttl := rdb.TTL(ctx, k.key).Val()
	if ttl <= 0 || ttl > 90*time.Second {
		t.Errorf("TTL = %v", ttl)
	}
	if k.key != "cache:berita:v0:list:p1:l9" {
		t.Errorf("format key = %q", k.key)
	}
}

func TestBatalkanMenaikkanVersiHanyaUntukJenisItu(t *testing.T) {
	rdb := testutil.RedisTes(t)
	c := &Cache{C: rdb}
	ctx := context.Background()
	kb, _ := c.Kunci(ctx, Berita, "x")
	kt, _ := c.Kunci(ctx, Tim, "x")
	_ = c.Simpan(ctx, kb, []byte("b"))
	_ = c.Simpan(ctx, kt, []byte("t"))

	if err := c.Batalkan(ctx, Berita); err != nil {
		t.Fatal(err)
	}
	kb2, _ := c.Kunci(ctx, Berita, "x")
	if _, ada, _ := c.Ambil(ctx, kb2); ada {
		t.Error("setelah dibatalkan key baru harus miss")
	}
	if kb2.key != "cache:berita:v1:x" {
		t.Errorf("key setelah batal = %q", kb2.key)
	}
	if b, ada, _ := c.Ambil(ctx, kt); !ada || string(b) != "t" {
		t.Error("jenis lain tidak boleh terpengaruh")
	}
}

func TestSatukanMenjalankanSekaliUntukKeySama(t *testing.T) {
	c := &Cache{}
	k := Kunci{key: "k"}
	var n atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b, err := c.Satukan(k, func() ([]byte, error) {
				n.Add(1)
				time.Sleep(100 * time.Millisecond)
				return []byte("ok"), nil
			})
			if err != nil || string(b) != "ok" {
				t.Errorf("hasil %q err %v", b, err)
			}
		}()
	}
	wg.Wait()
	if n.Load() != 1 {
		t.Errorf("pemuat berjalan %d kali", n.Load())
	}
	sentinel := errors.New("gagal")
	if _, err := c.Satukan(k, func() ([]byte, error) { return nil, sentinel }); !errors.Is(err, sentinel) {
		t.Errorf("galat harus diteruskan: %v", err)
	}
}
