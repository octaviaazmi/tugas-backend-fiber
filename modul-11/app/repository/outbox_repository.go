package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"modul-11/app/model"
)

type OutboxRepository interface {
	WithTx(tx DBTX) OutboxRepository
	Insert(ctx context.Context, eventType string, event any) error
	FetchUnpublished(ctx context.Context, limit int) ([]model.OutboxEvent, error)
	MarkPublished(ctx context.Context, ids []int64) error
}

type outboxPostgresRepository struct {
	db DBTX
}

func NewOutboxRepository(db DBTX) OutboxRepository {
	return &outboxPostgresRepository{db: db}
}

func (r *outboxPostgresRepository) WithTx(tx DBTX) OutboxRepository {
	return &outboxPostgresRepository{db: tx}
}

func (r *outboxPostgresRepository) Insert(ctx context.Context, eventType string, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("menyusun payload outbox: %w", err)
	}

	_, err = r.db.Exec(ctx,
		`INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`,
		eventType, payload,
	)
	if err != nil {
		return fmt.Errorf("mencatat outbox: %w", err)
	}
	return nil
}

func (r *outboxPostgresRepository) FetchUnpublished(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, event_type, payload
		 FROM outbox
		 WHERE published_at IS NULL
		 ORDER BY id
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("mengambil outbox: %w", err)
	}
	defer rows.Close()

	events := []model.OutboxEvent{}
	for rows.Next() {
		var e model.OutboxEvent
		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload); err != nil {
			return nil, fmt.Errorf("membaca row outbox: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *outboxPostgresRepository) MarkPublished(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.Exec(ctx,
		`UPDATE outbox SET published_at = NOW() WHERE id = ANY($1)`, ids,
	)
	if err != nil {
		return fmt.Errorf("menandai outbox: %w", err)
	}
	return nil
}
