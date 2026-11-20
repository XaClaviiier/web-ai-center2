package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

// Tes integrasi: Redis Docker lokal, DB 15 (dikosongkan oleh testutil.RedisTes).
func TestRedisStoreBulatPenuh(t *testing.T) {
	c := testutil.RedisTes(t)
	st := RedisStore{C: c}
	ctx := context.Background()
	id := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ" // 43 karakter sah
	if !IDSah(id) {
		t.Fatal("ID uji tidak sah")
	}

	want := Session{Identitas: dev, CSRFToken: "csrf-token", DibuatPada: time.Date(2026, 11, 9, 8, 0, 0, 0, time.UTC)}
	if err := st.Simpan(ctx, id, want, time.Hour); err != nil {
		t.Fatal(err)
	}
	got, err := st.Ambil(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Identitas != want.Identitas || got.CSRFToken != want.CSRFToken || !got.DibuatPada.Equal(want.DibuatPada) {
		t.Fatalf("data tidak kembali utuh: %+v", got)
	}

	ttl := c.TTL(ctx, "session:"+id).Val()
	if ttl <= 0 || ttl > time.Hour {
		t.Errorf("TTL di Redis salah: %v", ttl)
	}
	if err := st.Perpanjang(ctx, id, 10*time.Minute); err != nil {
		t.Fatal(err)
	}
	if ttl := c.TTL(ctx, "session:"+id).Val(); ttl <= 9*time.Minute || ttl > 10*time.Minute {
		t.Errorf("TTL setelah Perpanjang salah: %v", ttl)
	}

	if err := st.Hapus(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Ambil(ctx, id); !errors.Is(err, ErrTidakAda) {
		t.Errorf("setelah Hapus harus ErrTidakAda, dapat %v", err)
	}
	if err := st.Perpanjang(ctx, id, time.Minute); !errors.Is(err, ErrTidakAda) {
		t.Errorf("Perpanjang key hilang harus ErrTidakAda, dapat %v", err)
	}
}

func TestRedisStoreIsiRusakDianggapTidakAda(t *testing.T) {
	c := testutil.RedisTes(t)
	ctx := context.Background()
	id := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ"
	if err := c.Set(ctx, "session:"+id, "bukan json", time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := (RedisStore{C: c}).Ambil(ctx, id); !errors.Is(err, ErrTidakAda) {
		t.Fatalf("isi rusak harus ErrTidakAda, dapat %v", err)
	}
}

func TestManagerDenganRedisNyata(t *testing.T) {
	c := testutil.RedisTes(t)
	ctx := context.Background()
	m := NewManager(RedisStore{C: c}, time.Hour, 8*time.Hour)

	id, s, err := m.Buat(ctx, dev)
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.Validasi(ctx, id)
	if err != nil || got.CSRFToken != s.CSRFToken {
		t.Fatalf("Validasi: %+v %v", got, err)
	}
	if err := m.Hapus(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Validasi(ctx, id); !errors.Is(err, ErrTidakAda) {
		t.Fatalf("setelah logout harus ErrTidakAda, dapat %v", err)
	}
}
