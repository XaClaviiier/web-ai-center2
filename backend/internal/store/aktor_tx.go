package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/auth"
)

// setAktor mengirim identitas aktor ke database untuk transaksi ini saja (set_config lokal-transaksi). Trigger
// catat_aktivitas (migrasi 00003) membacanya untuk menulis log aktivitas di transaksi yang sama dengan perubahan data.
func setAktor(ctx context.Context, tx pgx.Tx, a auth.Aktor) error {
	if _, err := tx.Exec(ctx, "SELECT set_config('app.sub', $1, true), set_config('app.username', $2, true)", a.Sub, a.Username); err != nil {
		return fmt.Errorf("set aktor transaksi: %w", err)
	}
	return nil
}

// adalahTulis benar bila statement sqlc berupa INSERT/UPDATE/DELETE (baris komentar "-- name:" di awal diabaikan).
func adalahTulis(sql string) bool {
	for _, baris := range strings.Split(sql, "\n") {
		b := strings.TrimSpace(baris)
		if b == "" || strings.HasPrefix(b, "--") {
			continue
		}
		b = strings.ToUpper(b)
		return strings.HasPrefix(b, "INSERT") || strings.HasPrefix(b, "UPDATE") || strings.HasPrefix(b, "DELETE")
	}
	return false
}

// dbtxAktor membungkus pool untuk sqlc: statement TULIS yang dijalankan dengan aktor di context dijalankan dalam
// transaksi singkat yang lebih dulu men-set aktor, sehingga log aktivitas tertulis atomik bersama perubahan datanya
// (gagal menulis log menggagalkan perubahan). Tanpa aktor (job purge, tes, impor) dan untuk statement baca,
// perilakunya sama dengan pool biasa.
type dbtxAktor struct{ pool *pgxpool.Pool }

func (d dbtxAktor) mulai(ctx context.Context, a auth.Aktor) (pgx.Tx, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("memulai transaksi: %w", err)
	}
	if err := setAktor(ctx, tx, a); err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}
	return tx, nil
}

func (d dbtxAktor) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	a, ok := auth.AktorDariContext(ctx)
	if !ok || !adalahTulis(sql) {
		return d.pool.Exec(ctx, sql, args...)
	}
	tx, err := d.mulai(ctx, a)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	tag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		_ = tx.Rollback(ctx)
		return tag, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tag, fmt.Errorf("commit transaksi: %w", err)
	}
	return tag, nil
}

// Query meneruskan ke pool. Tidak ada statement tulis :many di sqlc proyek ini; bila kelak ada, ia harus dijalankan
// lewat dalamTx agar tercatat di log.
func (d dbtxAktor) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return d.pool.Query(ctx, sql, args...)
}

func (d dbtxAktor) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	a, ok := auth.AktorDariContext(ctx)
	if !ok || !adalahTulis(sql) {
		return d.pool.QueryRow(ctx, sql, args...)
	}
	tx, err := d.mulai(ctx, a)
	if err != nil {
		return galatRow{err}
	}
	return &rowTx{row: tx.QueryRow(ctx, sql, args...), tx: tx, ctx: ctx}
}

type galatRow struct{ err error }

func (g galatRow) Scan(...any) error { return g.err }

// rowTx menyelesaikan transaksi setelah baris dibaca: commit bila Scan sukses, rollback bila gagal (termasuk
// pgx.ErrNoRows, yang diklasifikasi pemanggil).
type rowTx struct {
	row pgx.Row
	tx  pgx.Tx
	ctx context.Context
}

func (r *rowTx) Scan(dest ...any) error {
	if err := r.row.Scan(dest...); err != nil {
		_ = r.tx.Rollback(r.ctx)
		return err
	}
	if err := r.tx.Commit(r.ctx); err != nil {
		return fmt.Errorf("commit transaksi: %w", err)
	}
	return nil
}
