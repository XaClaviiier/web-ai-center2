package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

// PanjangKunciSandi adalah panjang kunci AES-256 (byte mentah, SESSION_ENCRYPTION_KEY_FILE).
const PanjangKunciSandi = 32

// Sandi mengenkripsi refresh token yang disimpan di session Redis dengan AES-256-GCM. Setiap penyegelan memakai
// nonce acak baru; data tambahan (AAD) mengikat ciphertext ke key session sehingga tidak bisa dipindah ke session lain.
type Sandi struct {
	aead cipher.AEAD
}

// BuatSandi membuat Sandi dari kunci 32 byte. Kunci tidak pernah dicatat atau dikembalikan dalam galat.
func BuatSandi(kunci []byte) (*Sandi, error) {
	if len(kunci) != PanjangKunciSandi {
		return nil, fmt.Errorf("kunci enkripsi session harus tepat %d byte, dapat %d byte", PanjangKunciSandi, len(kunci))
	}
	blok, err := aes.NewCipher(kunci)
	if err != nil {
		return nil, fmt.Errorf("menyiapkan AES: %w", err)
	}
	aead, err := cipher.NewGCM(blok)
	if err != nil {
		return nil, fmt.Errorf("menyiapkan GCM: %w", err)
	}
	return &Sandi{aead: aead}, nil
}

// Segel mengembalikan nonce || ciphertext.
func (s *Sandi) Segel(teks, aad []byte) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize(), s.aead.NonceSize()+len(teks)+s.aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("membuat nonce: %w", err)
	}
	return s.aead.Seal(nonce, nonce, teks, aad), nil
}

// errSegelRusak berarti ciphertext tidak dapat dibuka (rusak, kunci berganti, atau AAD lain).
var errSegelRusak = errors.New("data terenkripsi tidak dapat dibuka")

// Buka membalik Segel.
func (s *Sandi) Buka(segel, aad []byte) ([]byte, error) {
	n := s.aead.NonceSize()
	if len(segel) < n+s.aead.Overhead() {
		return nil, errSegelRusak
	}
	teks, err := s.aead.Open(nil, segel[:n], segel[n:], aad)
	if err != nil {
		return nil, errSegelRusak
	}
	return teks, nil
}
