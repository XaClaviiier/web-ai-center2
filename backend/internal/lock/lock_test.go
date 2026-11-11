package lock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/testutil"
)

var (
	adminA = Pemegang{Sub: "a", Nama: "Admin A"}
	adminB = Pemegang{Sub: "b", Nama: "Admin B"}
)

func TestPasangBaruDanIdempotenUntukPemegangSama(t *testing.T) {
	rdb := testutil.RedisTes(t)
	l := &Lock{C: rdb, TTL: time.Minute}
	ctx := context.Background()

	info, err := l.Pasang(ctx, Berita, 5, adminA)
	if err != nil || info.Sub != "a" || info.Nama != "Admin A" {
		t.Fatalf("pasang: %+v err=%v", info, err)
	}
	if d := time.Until(info.KedaluwarsaPada); d < 55*time.Second || d > time.Minute {
		t.Errorf("kedaluwarsa tidak wajar: %v", d)
	}
	if ttl := rdb.TTL(ctx, "lock:berita:5").Val(); ttl <= 0 || ttl > time.Minute {
		t.Errorf("TTL key = %v (format key harus lock:{jenis}:{id})", ttl)
	}

	rdb.PExpire(ctx, "lock:berita:5", 5*time.Second)
	if _, err := l.Pasang(ctx, Berita, 5, adminA); err != nil {
		t.Fatalf("pemasangan ulang oleh pemegang sama harus idempoten: %v", err)
	}
	if ttl := rdb.PTTL(ctx, "lock:berita:5").Val(); ttl < 50*time.Second {
		t.Errorf("TTL harus diperbarui, dapat %v", ttl)
	}
}

func TestPasangDitolakBilaDipegangOrangLain(t *testing.T) {
	l := &Lock{C: testutil.RedisTes(t), TTL: time.Minute}
	ctx := context.Background()
	if _, err := l.Pasang(ctx, Berita, 5, adminA); err != nil {
		t.Fatal(err)
	}
	_, err := l.Pasang(ctx, Berita, 5, adminB)
	var dp *ErrDipegang
	if !errors.As(err, &dp) || dp.Info.Nama != "Admin A" || dp.Info.Sub != "a" {
		t.Fatalf("harus ErrDipegang oleh A: %v", err)
	}
	// Item atau jenis lain tidak terpengaruh.
	if _, err := l.Pasang(ctx, Berita, 6, adminB); err != nil {
		t.Errorf("id lain: %v", err)
	}
	if _, err := l.Pasang(ctx, Tim, 5, adminB); err != nil {
		t.Errorf("jenis lain: %v", err)
	}
}

func TestPerpanjangHanyaPemegang(t *testing.T) {
	rdb := testutil.RedisTes(t)
	l := &Lock{C: rdb, TTL: time.Minute}
	ctx := context.Background()
	if _, err := l.Perpanjang(ctx, Berita, 5, "a"); !errors.Is(err, ErrHilang) {
		t.Errorf("tanpa kunci harus ErrHilang: %v", err)
	}
	_, _ = l.Pasang(ctx, Berita, 5, adminA)
	if _, err := l.Perpanjang(ctx, Berita, 5, "b"); !errors.Is(err, ErrHilang) {
		t.Errorf("bukan pemegang harus ErrHilang: %v", err)
	}
	rdb.PExpire(ctx, "lock:berita:5", 5*time.Second)
	info, err := l.Perpanjang(ctx, Berita, 5, "a")
	if err != nil || info.Nama != "Admin A" {
		t.Fatalf("perpanjang: %+v %v", info, err)
	}
	if ttl := rdb.PTTL(ctx, "lock:berita:5").Val(); ttl < 50*time.Second {
		t.Errorf("TTL harus diperpanjang, dapat %v", ttl)
	}
}

func TestLepasHanyaPemegangDanIdempoten(t *testing.T) {
	rdb := testutil.RedisTes(t)
	l := &Lock{C: rdb, TTL: time.Minute}
	ctx := context.Background()
	_, _ = l.Pasang(ctx, Berita, 5, adminA)

	if err := l.Lepas(ctx, Berita, 5, "b"); err != nil {
		t.Fatal(err)
	}
	if rdb.Exists(ctx, "lock:berita:5").Val() != 1 {
		t.Fatal("kunci milik orang lain tidak boleh dilepas")
	}
	if err := l.Lepas(ctx, Berita, 5, "a"); err != nil {
		t.Fatal(err)
	}
	if rdb.Exists(ctx, "lock:berita:5").Val() != 0 {
		t.Fatal("pemegang harus bisa melepas")
	}
	if err := l.Lepas(ctx, Berita, 5, "a"); err != nil {
		t.Errorf("lepas ulang harus idempoten: %v", err)
	}
}

func TestBaca(t *testing.T) {
	l := &Lock{C: testutil.RedisTes(t), TTL: time.Minute}
	ctx := context.Background()
	if info, err := l.Baca(ctx, Berita, 5); err != nil || info != nil {
		t.Fatalf("tanpa kunci harus nil: %+v %v", info, err)
	}
	_, _ = l.Pasang(ctx, Berita, 5, adminA)
	info, err := l.Baca(ctx, Berita, 5)
	if err != nil || info == nil || info.Sub != "a" || info.Nama != "Admin A" || time.Until(info.KedaluwarsaPada) < 50*time.Second {
		t.Fatalf("baca: %+v %v", info, err)
	}
}

func TestKunciKedaluwarsaBisaDiambilOrangLain(t *testing.T) {
	l := &Lock{C: testutil.RedisTes(t), TTL: 300 * time.Millisecond}
	ctx := context.Background()
	_, _ = l.Pasang(ctx, Berita, 5, adminA)
	time.Sleep(450 * time.Millisecond)
	if _, err := l.Perpanjang(ctx, Berita, 5, "a"); !errors.Is(err, ErrHilang) {
		t.Errorf("heartbeat setelah kedaluwarsa harus ErrHilang: %v", err)
	}
	if _, err := l.Pasang(ctx, Berita, 5, adminB); err != nil {
		t.Errorf("setelah kedaluwarsa orang lain harus bisa memasang: %v", err)
	}
}

func TestSatuPemenangDariPemasangBersamaan(t *testing.T) {
	l := &Lock{C: testutil.RedisTes(t), TTL: time.Minute}
	ctx := context.Background()
	const n = 20
	hasil := make(chan error, n)
	for i := 0; i < n; i++ {
		p := Pemegang{Sub: "u" + string(rune('a'+i)), Nama: "U"}
		go func() { _, err := l.Pasang(ctx, Berita, 9, p); hasil <- err }()
	}
	menang := 0
	for i := 0; i < n; i++ {
		if err := <-hasil; err == nil {
			menang++
		}
	}
	if menang != 1 {
		t.Errorf("tepat satu pemasang harus menang, dapat %d", menang)
	}
}
