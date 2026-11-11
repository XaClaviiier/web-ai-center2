// Package lock menyimpan soft lock edit di Redis (docs/database.md §4): key lock:{jenis}:{id} berisi pemegang
// (JSON {sub,nama}), dibuat hanya bila belum ada, dengan TTL. Hanya pemegang yang dapat memperpanjang atau
// melepas; semua pemeriksaan "hanya pemegang" dijalankan atomik di Redis lewat skrip Lua.
package lock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Jenis konten yang bisa dikunci.
const (
	Berita      = "berita"
	Dokumentasi = "dokumentasi"
	Testimoni   = "testimoni"
	Tim         = "tim"
)

// Pemegang adalah identitas pemegang kunci (sub dari session; nama untuk ditampilkan ke pengguna lain).
type Pemegang struct {
	Sub  string `json:"sub"`
	Nama string `json:"nama"`
}

// Info adalah keadaan sebuah kunci.
type Info struct {
	Pemegang
	KedaluwarsaPada time.Time
}

// ErrDipegang berarti kunci sedang dipegang orang lain.
type ErrDipegang struct{ Info Info }

func (e *ErrDipegang) Error() string { return "konten sedang dikunci oleh " + e.Info.Nama }

// ErrHilang berarti kunci sudah kedaluwarsa atau bukan milik pemanggil.
var ErrHilang = errors.New("kunci tidak ada atau bukan milik pemanggil")

// Lock adalah soft lock berbasis Redis.
type Lock struct {
	C   redis.Cmdable
	TTL time.Duration // default 5 menit bila <= 0
}

func (l *Lock) ttl() time.Duration {
	if l.TTL <= 0 {
		return 5 * time.Minute
	}
	return l.TTL
}

func key(jenis string, id int64) string { return fmt.Sprintf("lock:%s:%d", jenis, id) }

// Pasang membuat kunci (SET NX). Pemegang yang sama memasang lagi: idempoten, TTL diperbarui.
// Dipegang orang lain: *ErrDipegang.
var skrinPasang = redis.NewScript(`
local v = redis.call('GET', KEYS[1])
if not v then
  redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[3])
  return {1, ARGV[1], tonumber(ARGV[3])}
end
local cur = cjson.decode(v)
if cur.sub == ARGV[2] then
  redis.call('PEXPIRE', KEYS[1], ARGV[3])
  return {1, v, tonumber(ARGV[3])}
end
return {0, v, redis.call('PTTL', KEYS[1])}
`)

func (l *Lock) Pasang(ctx context.Context, jenis string, id int64, p Pemegang) (Info, error) {
	val, err := json.Marshal(p)
	if err != nil {
		return Info{}, err
	}
	hasil, err := skrinPasang.Run(ctx, l.C, []string{key(jenis, id)}, string(val), p.Sub, l.ttl().Milliseconds()).Slice()
	if err != nil {
		return Info{}, fmt.Errorf("lock: pasang: %w", err)
	}
	info, err := uraiHasil(hasil)
	if err != nil {
		return Info{}, err
	}
	if hasil[0].(int64) == 0 {
		return Info{}, &ErrDipegang{Info: info}
	}
	return info, nil
}

// Perpanjang (heartbeat) memperpanjang TTL bila kunci milik sub; selain itu ErrHilang.
var skrinPerpanjang = redis.NewScript(`
local v = redis.call('GET', KEYS[1])
if not v then return {0} end
local cur = cjson.decode(v)
if cur.sub ~= ARGV[1] then return {0} end
redis.call('PEXPIRE', KEYS[1], ARGV[2])
return {1, v, tonumber(ARGV[2])}
`)

func (l *Lock) Perpanjang(ctx context.Context, jenis string, id int64, sub string) (Info, error) {
	hasil, err := skrinPerpanjang.Run(ctx, l.C, []string{key(jenis, id)}, sub, l.ttl().Milliseconds()).Slice()
	if err != nil {
		return Info{}, fmt.Errorf("lock: perpanjang: %w", err)
	}
	if hasil[0].(int64) == 0 {
		return Info{}, ErrHilang
	}
	return uraiHasil(hasil)
}

// Lepas menghapus kunci hanya bila milik sub. Tidak ada atau milik orang lain: tidak melakukan apa-apa (idempoten).
var skrinLepas = redis.NewScript(`
local v = redis.call('GET', KEYS[1])
if not v then return 0 end
local cur = cjson.decode(v)
if cur.sub ~= ARGV[1] then return 0 end
return redis.call('DEL', KEYS[1])
`)

func (l *Lock) Lepas(ctx context.Context, jenis string, id int64, sub string) error {
	if err := skrinLepas.Run(ctx, l.C, []string{key(jenis, id)}, sub).Err(); err != nil {
		return fmt.Errorf("lock: lepas: %w", err)
	}
	return nil
}

// Baca mengembalikan keadaan kunci, atau nil bila tidak ada.
func (l *Lock) Baca(ctx context.Context, jenis string, id int64) (*Info, error) {
	var (
		get  *redis.StringCmd
		pttl *redis.DurationCmd
	)
	_, err := l.C.Pipelined(ctx, func(p redis.Pipeliner) error {
		get = p.Get(ctx, key(jenis, id))
		pttl = p.PTTL(ctx, key(jenis, id))
		return nil
	})
	if errors.Is(get.Err(), redis.Nil) {
		return nil, nil // tidak ada kunci
	}
	if err != nil {
		return nil, fmt.Errorf("lock: baca: %w", err)
	}
	var p Pemegang
	if err := json.Unmarshal([]byte(get.Val()), &p); err != nil {
		return nil, fmt.Errorf("lock: isi kunci rusak: %w", err)
	}
	sisa := pttl.Val()
	if sisa < 0 {
		sisa = 0
	}
	return &Info{Pemegang: p, KedaluwarsaPada: time.Now().Add(sisa)}, nil
}

// uraiHasil membaca balasan skrip: [status, json pemegang, pttl ms].
func uraiHasil(h []any) (Info, error) {
	if len(h) < 3 {
		return Info{}, errors.New("lock: balasan skrip tidak lengkap")
	}
	s, _ := h[1].(string)
	ms, _ := h[2].(int64)
	var p Pemegang
	if err := json.Unmarshal([]byte(s), &p); err != nil {
		return Info{}, fmt.Errorf("lock: isi kunci rusak: %w", err)
	}
	if ms < 0 {
		ms = 0
	}
	return Info{Pemegang: p, KedaluwarsaPada: time.Now().Add(time.Duration(ms) * time.Millisecond)}, nil
}
