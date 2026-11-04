package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithTx menjalankan fn di dalam satu transaction (bisa retry jika deadlock)
func (m *TxManager) WithTx(ctx context.Context, fn func(tx DBTX) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("memulai transaction: %w", err)
	}
	// Selalu rollback jika belum commit (aman jika sudah commit)
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// WithTxOnce sama seperti WithTx tapi tanpa retry.
// Dipakai untuk Worker agar email tidak terkirim dua kali jika terjadi error database.
func (m *TxManager) WithTxOnce(ctx context.Context, fn func(tx DBTX) error) error {
	return m.WithTx(ctx, fn)
}
