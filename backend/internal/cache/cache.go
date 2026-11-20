// Package cache menyimpan body JSON respons publik di Redis (cache-aside, prefix "cache:").
//
// Pembatalan memakai penghitung versi per jenis konten (cache:ver:{jenis}) yang ikut menjadi bagian key,
// jadi membatalkan cukup satu INCR; key lama kedaluwarsa sendiri lewat TTL (docs/database.md §7).
package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// Jenis konten yang di-cache; dipakai sebagai bagian key dan nama penghitung versi.
const (
	Berita      = "berita"
	Dokumentasi = "dokumentasi"
	Testimoni   = "testimoni"
	Tim         = "tim"
	Sitemap     = "sitemap" // dibatalkan bersama berita dan dokumentasi
)

// Cache adalah cache body JSON di Redis. Semua metode aman dipanggil bersamaan.
type Cache struct {
	C   redis.Cmdable
	TTL time.Duration // default 20 menit bila <= 0

	sf singleflight.Group
}

func (c *Cache) ttl() time.Duration {
	if c.TTL <= 0 {
		return 20 * time.Minute
	}
	return c.TTL
}

// Kunci adalah pegangan ke satu entri cache pada versi tertentu. Versi dibaca SEBELUM data dimuat dari database,
// jadi data yang dimuat sebelum pembatalan tidak pernah tersimpan di bawah versi baru.
type Kunci struct {
	key string
}

// Kunci membaca versi jenis lalu menyusun key "cache:{jenis}:v{versi}:{sub}".
func (c *Cache) Kunci(ctx context.Context, jenis, sub string) (Kunci, error) {
	v, err := c.C.Get(ctx, "cache:ver:"+jenis).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return Kunci{}, fmt.Errorf("cache: baca versi %s: %w", jenis, err)
	}
	return Kunci{key: "cache:" + jenis + ":v" + strconv.FormatInt(v, 10) + ":" + sub}, nil
}

// Ambil mengembalikan body tersimpan; ok=false bila tidak ada (miss).
func (c *Cache) Ambil(ctx context.Context, k Kunci) (body []byte, ok bool, err error) {
	b, err := c.C.Get(ctx, k.key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("cache: baca %s: %w", k.key, err)
	}
	return b, true, nil
}

// Simpan menulis body dengan TTL.
func (c *Cache) Simpan(ctx context.Context, k Kunci, body []byte) error {
	if err := c.C.Set(ctx, k.key, body, c.ttl()).Err(); err != nil {
		return fmt.Errorf("cache: tulis %s: %w", k.key, err)
	}
	return nil
}

// Batalkan menaikkan versi jenis sehingga semua key lama tak terbaca lagi. Panggil SETELAH perubahan
// di database berhasil. Penghitung versi tidak diberi TTL (kecil, satu per jenis).
func (c *Cache) Batalkan(ctx context.Context, jenis string) error {
	if err := c.C.Incr(ctx, "cache:ver:"+jenis).Err(); err != nil {
		return fmt.Errorf("cache: batalkan %s: %w", jenis, err)
	}
	return nil
}

// Satukan menjalankan muat sekali untuk semua pemanggil bersamaan dengan key yang sama (anti-stampede).
// Hasil dan galat dibagikan; pemanggil tidak boleh mengubah isi slice yang dikembalikan.
func (c *Cache) Satukan(k Kunci, muat func() ([]byte, error)) ([]byte, error) {
	v, err, _ := c.sf.Do(k.key, func() (any, error) { return muat() })
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}
