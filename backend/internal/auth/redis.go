package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	prefixSession = "session:"
	// prefixKunciSegar: kunci penyegaran token per session (SET NX), terpisah dari key session.
	prefixKunciSegar = "session-segar:"
)

// RedisStore menyimpan session di Redis pada key session:{id}. Refresh token dienkripsi dengan Sandi; tanpa Sandi,
// session yang membawa refresh token ditolak (tidak pernah disimpan sebagai teks biasa).
type RedisStore struct {
	C     redis.Cmdable
	Sandi *Sandi
}

type sessionJSON struct {
	Sub        string    `json:"sub"`
	Nama       string    `json:"nama"`
	Role       string    `json:"role"`
	Uname      string    `json:"username,omitempty"`
	CSRF       string    `json:"csrf"`
	Dibuat     time.Time `json:"dibuat"`
	Refresh    []byte    `json:"rt,omitempty"` // nonce || ciphertext AES-256-GCM
	AksesHabis time.Time `json:"akses_habis,omitzero"`
}

// errTanpaSandi: refresh token tidak boleh disimpan tanpa enkripsi.
var errTanpaSandi = errors.New("refresh token tidak dapat disimpan tanpa kunci enkripsi session")

func (r RedisStore) kodekan(id string, s Session) ([]byte, error) {
	j := sessionJSON{Sub: s.Sub, Nama: s.Nama, Role: s.Role, Uname: s.Username, CSRF: s.CSRFToken, Dibuat: s.DibuatPada, AksesHabis: s.AksesHabis}
	if s.Refresh != "" {
		if r.Sandi == nil {
			return nil, errTanpaSandi
		}
		segel, err := r.Sandi.Segel([]byte(s.Refresh), []byte(prefixSession+id))
		if err != nil {
			return nil, err
		}
		j.Refresh = segel
	}
	return json.Marshal(j)
}

func (r RedisStore) Simpan(ctx context.Context, id string, s Session, ttl time.Duration) error {
	b, err := r.kodekan(id, s)
	if err != nil {
		return err
	}
	if err := r.C.Set(ctx, prefixSession+id, b, ttl).Err(); err != nil {
		return fmt.Errorf("redis SET session: %w", err)
	}
	return nil
}

// GantiBilaAda memakai SET XX: atomik, tidak pernah membuat ulang key yang sudah dihapus logout.
func (r RedisStore) GantiBilaAda(ctx context.Context, id string, s Session, ttl time.Duration) error {
	b, err := r.kodekan(id, s)
	if err != nil {
		return err
	}
	ok, err := r.C.SetXX(ctx, prefixSession+id, b, ttl).Result()
	if err != nil {
		return fmt.Errorf("redis SET XX session: %w", err)
	}
	if !ok {
		return ErrTidakAda
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
	s := Session{
		Identitas:  Identitas{Sub: j.Sub, Nama: j.Nama, Role: j.Role, Username: j.Uname},
		Token:      Token{AksesHabis: j.AksesHabis},
		CSRFToken:  j.CSRF,
		DibuatPada: j.Dibuat,
	}
	if len(j.Refresh) > 0 {
		if r.Sandi == nil {
			return Session{}, ErrTidakAda
		}
		// Kunci berganti atau data rusak: sama dengan isi rusak, login ulang.
		teks, err := r.Sandi.Buka(j.Refresh, []byte(prefixSession+id))
		if err != nil {
			return Session{}, ErrTidakAda
		}
		s.Refresh = string(teks)
	}
	return s, nil
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

func (r RedisStore) KunciSegar(ctx context.Context, id, pemilik string, ttl time.Duration) (bool, error) {
	ok, err := r.C.SetNX(ctx, prefixKunciSegar+id, pemilik, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis SET NX kunci segar: %w", err)
	}
	return ok, nil
}

var skripLepasKunciSegar = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then return redis.call('DEL', KEYS[1]) end
return 0`)

func (r RedisStore) LepasKunciSegar(ctx context.Context, id, pemilik string) error {
	if err := skripLepasKunciSegar.Run(ctx, r.C, []string{prefixKunciSegar + id}, pemilik).Err(); err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("redis lepas kunci segar: %w", err)
	}
	return nil
}
