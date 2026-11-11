// Package store membungkus koneksi PostgreSQL (pgx) dan query hasil sqlc.
package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/store/db"
)

// Store memegang pool koneksi dan query bertipe (sqlc).
type Store struct {
	Pool *pgxpool.Pool
	Q    *db.Queries
}

// Open membuka pool dan memastikan database terjangkau.
func Open(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("membuka pool database: %w", err)
	}
	s := &Store{Pool: pool, Q: db.New(pool)}
	if err := s.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

// Ping menjalankan kueri ringan untuk memeriksa koneksi.
func (s *Store) Ping(ctx context.Context) error {
	if _, err := s.Q.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	return nil
}

// Close menutup pool.
func (s *Store) Close() { s.Pool.Close() }
