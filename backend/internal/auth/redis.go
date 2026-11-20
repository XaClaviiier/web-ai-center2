package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const prefixSession = "session:"

// RedisStore menyimpan session di Redis pada key session:{id}.
type RedisStore struct {
	C redis.Cmdable
}

type sessionJSON struct {
	Sub    string    `json:"sub"`
	Nama   string    `json:"nama"`
	Role   string    `json:"role"`
	Uname  string    `json:"username,omitempty"`
	CSRF   string    `json:"csrf"`
	Dibuat time.Time `json:"dibuat"`
}

func (r RedisStore) Simpan(ctx context.Context, id string, s Session, ttl time.Duration) error {
	b, err := json.Marshal(sessionJSON{Sub: s.Sub, Nama: s.Nama, Role: s.Role, Uname: s.Username, CSRF: s.CSRFToken, Dibuat: s.DibuatPada})
	if err != nil {
		return err
	}
	if err := r.C.Set(ctx, prefixSession+id, b, ttl).Err(); err != nil {
		return fmt.Errorf("redis SET session: %w", err)
	}
	return nil
}

func (r RedisStore) Ambil(ctx context.Context, id string) (Session, error) {
	raw, err := r.C.Get(ctx, prefixSession+id).Bytes()
	if errors.Is(err, redis.Nil) {
		return Session{}, ErrTidakAda
	}
	if err != nil {
		return Session{}, fmt.Errorf("redis GET session: %w", err)
	}
	var j sessionJSON
	if err := json.Unmarshal(raw, &j); err != nil {
		// Isi rusak diperlakukan sebagai tidak ada supaya pengguna login ulang, bukan terkunci.
		return Session{}, ErrTidakAda
	}
	return Session{Identitas: Identitas{Sub: j.Sub, Nama: j.Nama, Role: j.Role, Username: j.Uname}, CSRFToken: j.CSRF, DibuatPada: j.Dibuat}, nil
}

func (r RedisStore) Perpanjang(ctx context.Context, id string, ttl time.Duration) error {
	ok, err := r.C.Expire(ctx, prefixSession+id, ttl).Result()
	if err != nil {
		return fmt.Errorf("redis EXPIRE session: %w", err)
	}
	if !ok {
		return ErrTidakAda
	}
	return nil
}

func (r RedisStore) Hapus(ctx context.Context, id string) error {
	if err := r.C.Del(ctx, prefixSession+id).Err(); err != nil {
		return fmt.Errorf("redis DEL session: %w", err)
	}
	return nil
}
