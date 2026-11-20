// Package ratelimit membatasi jumlah request per jendela waktu dengan penghitung Redis (security.md §6).
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Keputusan adalah hasil pemeriksaan satu request.
type Keputusan struct {
	Diizinkan bool
	// TungguDetik adalah sisa jendela (untuk header Retry-After) bila ditolak; minimal 1.
	TungguDetik int
}

// Limiter menghitung satu percobaan untuk (bucket, id) dan memutuskan boleh atau tidak.
type Limiter interface {
	Hit(ctx context.Context, bucket, id string, batas int) (Keputusan, error)
}

// Redis menghitung per jendela tetap: INCR pada key ratelimit:{bucket}:{id}, EXPIRE hanya saat key baru.
type Redis struct {
	C       redis.Cmdable
	Jendela time.Duration // default 1 menit
}

func (r Redis) Hit(ctx context.Context, bucket, id string, batas int) (Keputusan, error) {
	jendela := r.Jendela
	if jendela <= 0 {
		jendela = time.Minute
	}
	key := fmt.Sprintf("ratelimit:%s:%s", bucket, id)

	var (
		incr *redis.IntCmd
		pttl *redis.DurationCmd
	)
	_, err := r.C.Pipelined(ctx, func(p redis.Pipeliner) error {
		incr = p.Incr(ctx, key)
		p.ExpireNX(ctx, key, jendela) // memastikan key selalu punya TTL, tanpa memperpanjang jendela berjalan
		pttl = p.PTTL(ctx, key)
		return nil
	})
	if err != nil {
		return Keputusan{}, fmt.Errorf("redis rate limit: %w", err)
	}
	if incr.Val() <= int64(batas) {
		return Keputusan{Diizinkan: true}, nil
	}
	tunggu := int((pttl.Val() + time.Second - 1) / time.Second)
	if tunggu < 1 {
		tunggu = 1
	}
	return Keputusan{Diizinkan: false, TungguDetik: tunggu}, nil
}
