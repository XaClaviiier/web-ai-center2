package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// MasaTransaksiLogin adalah batas waktu antara GET /api/auth/login dan callback dari Keycloak.
const MasaTransaksiLogin = 5 * time.Minute

const prefixTransaksiLogin = "oidc:login:"

// ErrTransaksiTidakAda berarti state login tidak dikenal, kedaluwarsa, atau sudah dipakai.
var ErrTransaksiTidakAda = errors.New("transaksi login tidak ada, kedaluwarsa, atau sudah dipakai")

// TransaksiLogin adalah data satu percobaan login yang menunggu callback. HashIkatan adalah Hash(nilai cookie
// pengikat) sehingga callback hanya diterima dari browser yang memulai login.
type TransaksiLogin struct {
	Verifier   string `json:"verifier"` // PKCE code_verifier
	Nonce      string `json:"nonce"`
	HashIkatan string `json:"ikatan"`
}

// RedisTransaksi menyimpan transaksi login di Redis pada key oidc:login:{sha256(state)}; state mentah tidak disimpan.
type RedisTransaksi struct {
	C redis.Cmdable
}

// Simpan menyimpan transaksi dengan masa hidup ttl.
func (r RedisTransaksi) Simpan(ctx context.Context, state string, t TransaksiLogin, ttl time.Duration) error {
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	if err := r.C.Set(ctx, prefixTransaksiLogin+Hash(state), b, ttl).Err(); err != nil {
		return fmt.Errorf("redis SET transaksi login: %w", err)
	}
	return nil
}

// Ambil mengambil sekaligus menghapus transaksi (GETDEL, atomik): satu state hanya bisa dipakai sekali.
func (r RedisTransaksi) Ambil(ctx context.Context, state string) (TransaksiLogin, error) {
	raw, err := r.C.GetDel(ctx, prefixTransaksiLogin+Hash(state)).Bytes()
	if errors.Is(err, redis.Nil) {
		return TransaksiLogin{}, ErrTransaksiTidakAda
	}
	if err != nil {
		return TransaksiLogin{}, fmt.Errorf("redis GETDEL transaksi login: %w", err)
	}
	var t TransaksiLogin
	if err := json.Unmarshal(raw, &t); err != nil || t.Verifier == "" || t.Nonce == "" || t.HashIkatan == "" {
		return TransaksiLogin{}, ErrTransaksiTidakAda
	}
	return t, nil
}

// Hash mengembalikan SHA-256 heksadesimal (key transaksi dan pengikat cookie; nilai mentah tidak disimpan).
func Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
