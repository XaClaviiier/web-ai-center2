package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

// Memori adalah Store di memori untuk tes handler (tanpa RustFS). Err (bila diisi) dikembalikan semua operasi.
type Memori struct {
	mu   sync.Mutex
	data map[string]memObjek
	Err  error
}

type memObjek struct {
	data   []byte
	ctype  string
	diubah time.Time
}

// NewMemori membuat Store kosong.
func NewMemori() *Memori { return &Memori{data: map[string]memObjek{}} }

func (m *Memori) Put(_ context.Context, key string, data []byte, contentType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return m.Err
	}
	m.data[key] = memObjek{data: append([]byte(nil), data...), ctype: contentType, diubah: time.Now()}
	return nil
}

func (m *Memori) ambil(key string) (memObjek, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return memObjek{}, m.Err
	}
	o, ok := m.data[key]
	if !ok {
		return memObjek{}, ErrTidakAda
	}
	return o, nil
}

func metaDari(o memObjek) Meta {
	h := sha256.Sum256(o.data)
	return Meta{ContentType: o.ctype, Ukuran: int64(len(o.data)), ETag: fmt.Sprintf("%q", fmt.Sprintf("%x", h[:8]))}
}

func (m *Memori) Get(_ context.Context, key string) (Objek, error) {
	o, err := m.ambil(key)
	if err != nil {
		return Objek{}, err
	}
	return Objek{Meta: metaDari(o), Body: io.NopCloser(bytes.NewReader(o.data))}, nil
}

func (m *Memori) Head(_ context.Context, key string) (Meta, error) {
	o, err := m.ambil(key)
	if err != nil {
		return Meta{}, err
	}
	return metaDari(o), nil
}

func (m *Memori) Hapus(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return m.Err
	}
	delete(m.data, key)
	return nil
}

func (m *Memori) Daftar(_ context.Context, prefix string) ([]Info, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return nil, m.Err
	}
	var out []Info
	for k, o := range m.data {
		if strings.HasPrefix(k, prefix) {
			out = append(out, Info{Key: k, Ukuran: int64(len(o.data)), Diubah: o.diubah})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

// Jumlah mengembalikan banyak objek tersimpan (untuk asersi di tes).
func (m *Memori) Jumlah() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.data)
}
