package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

// Tes integrasi: Redis Docker lokal, DB 15.
func TestBatasDanRetryAfter(t *testing.T) {
	c := testutil.RedisTes(t)
	l := Redis{C: c, Jendela: time.Minute}
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		k, err := l.Hit(ctx, "login", "1.2.3.4", 3)
		if err != nil || !k.Diizinkan {
			t.Fatalf("percobaan %d harus diizinkan: %+v %v", i, k, err)
		}
	}
	k, err := l.Hit(ctx, "login", "1.2.3.4", 3)
	if err != nil {
		t.Fatal(err)
	}
	if k.Diizinkan || k.TungguDetik < 1 || k.TungguDetik > 60 {
		t.Fatalf("percobaan ke-4 harus ditolak dengan Retry-After 1..60: %+v", k)
	}
	ttl := c.TTL(ctx, "ratelimit:login:1.2.3.4").Val()
	if ttl <= 0 || ttl > time.Minute {
		t.Errorf("key harus punya TTL <= 1 menit, dapat %v", ttl)
	}
}

func TestIdDanBucketTerpisah(t *testing.T) {
	c := testutil.RedisTes(t)
	l := Redis{C: c, Jendela: time.Minute}
	ctx := context.Background()

	if k, _ := l.Hit(ctx, "login", "a", 1); !k.Diizinkan {
		t.Fatal("a pertama harus diizinkan")
	}
	if k, _ := l.Hit(ctx, "login", "a", 1); k.Diizinkan {
		t.Fatal("a kedua harus ditolak")
	}
	if k, _ := l.Hit(ctx, "login", "b", 1); !k.Diizinkan {
		t.Error("IP lain tidak boleh terpengaruh")
	}
	if k, _ := l.Hit(ctx, "tulis", "a", 1); !k.Diizinkan {
		t.Error("bucket lain tidak boleh terpengaruh")
	}
}

func TestJendelaBerakhirMenghitungUlang(t *testing.T) {
	c := testutil.RedisTes(t)
	l := Redis{C: c, Jendela: time.Second}
	ctx := context.Background()

	if k, _ := l.Hit(ctx, "login", "x", 1); !k.Diizinkan {
		t.Fatal("pertama harus diizinkan")
	}
	if k, _ := l.Hit(ctx, "login", "x", 1); k.Diizinkan {
		t.Fatal("kedua dalam jendela harus ditolak")
	}
	time.Sleep(1200 * time.Millisecond)
	if k, _ := l.Hit(ctx, "login", "x", 1); !k.Diizinkan {
		t.Fatal("setelah jendela berakhir harus diizinkan lagi")
	}
}

func TestJendelaTidakDiperpanjangOlehPercobaanBerikutnya(t *testing.T) {
	c := testutil.RedisTes(t)
	l := Redis{C: c, Jendela: 2 * time.Second}
	ctx := context.Background()

	_, _ = l.Hit(ctx, "login", "y", 5)
	time.Sleep(800 * time.Millisecond)
	_, _ = l.Hit(ctx, "login", "y", 5)
	if ttl := c.PTTL(ctx, "ratelimit:login:y").Val(); ttl > 1500*time.Millisecond {
		t.Errorf("TTL tidak boleh di-reset oleh percobaan berikutnya, sisa %v", ttl)
	}
}
