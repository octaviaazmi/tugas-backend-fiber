package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul-11/app/model"
	"modul-11/app/repository"
)

type OutboxRelay struct {
	outbox   repository.OutboxRepository
	tx       *repository.TxManager
	queue    *asynq.Client
	pool     *pgxpool.Pool // Tambahan untuk query enrollments
	interval time.Duration
	batch    int
}

func NewOutboxRelay(
	outbox repository.OutboxRepository,
	tx *repository.TxManager,
	queue *asynq.Client,
	pool *pgxpool.Pool,
	interval time.Duration,
	batch int,
) *OutboxRelay {
	return &OutboxRelay{outbox: outbox, tx: tx, queue: queue, pool: pool, interval: interval, batch: batch}
}

func (r *OutboxRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.RelayOnce(ctx)
		}
	}
}

func (r *OutboxRelay) RelayOnce(ctx context.Context) {
	_ = r.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		repo := r.outbox.WithTx(tx)
		events, err := repo.FetchUnpublished(ctx, r.batch)
		if err != nil {
			return err
		}

		ids := make([]int64, 0, len(events))
		for _, e := range events {
			if err := r.publish(ctx, e); err != nil {
				log.Printf("Gagal enqueue outbox %d: %v", e.ID, err)
				break
			}
			ids = append(ids, e.ID)
		}

		if len(ids) > 0 {
			if err := repo.MarkPublished(ctx, ids); err != nil {
				return err
			}
			log.Printf("[RELAY] Berhasil memindahkan %d event ke Redis\n", len(ids))
		}
		return nil
	})
}

func (r *OutboxRelay) publish(ctx context.Context, e model.OutboxEvent) error {
	// KALAU EVENT UBAH KUOTA, KIRIM KE BANYAK MAHASISWA
	if e.EventType == model.EventQuotaChanged {
		var qEvent model.QuotaChangedEvent
		_ = json.Unmarshal(e.Payload, &qEvent)

		// Ambil semua mahasiswa di kelas tersebut dari database
		rows, err := r.pool.Query(ctx, "SELECT student_id FROM enrollments WHERE course_id = $1", qEvent.CourseID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var studentIDs []int
		for rows.Next() {
			var sID int
			_ = rows.Scan(&sID)
			studentIDs = append(studentIDs, sID)
		}

		// Buat 1 task untuk setiap mahasiswa
		for _, sID := range studentIDs {
			// Buat payload spesifik untuk mahasiswa ini
			specificEvent := qEvent
			specificEvent.UserID = sID
			payloadBytes, _ := json.Marshal(specificEvent)

			job := model.NotificationJob{
				OutboxID:  e.ID,
				EventType: e.EventType,
				Payload:   payloadBytes,
			}
			taskPayload, _ := json.Marshal(job)

			task := asynq.NewTask(model.TaskSendNotification, taskPayload,
				// TaskID dibedakan per mahasiswa agar tidak dianggap duplikat
				asynq.TaskID(fmt.Sprintf("outbox-%d-user-%d", e.ID, sID)),
				asynq.Queue(model.QueueNotifications),
				asynq.MaxRetry(5),
				asynq.Retention(24*time.Hour),
			)

			_, err := r.queue.Enqueue(task)
			if err != nil && !errors.Is(err, asynq.ErrTaskIDConflict) {
				return err
			}
		}
		return nil
	}

	// LOGIKA NORMAL (Untuk Enrollment)
	job := model.NotificationJob{
		OutboxID:  e.ID,
		EventType: e.EventType,
		Payload:   e.Payload,
	}
	payloadBytes, _ := json.Marshal(job)

	task := asynq.NewTask(model.TaskSendNotification, payloadBytes,
		asynq.TaskID(fmt.Sprintf("outbox-%d", e.ID)),
		asynq.Queue(model.QueueNotifications),
		asynq.MaxRetry(5),
		asynq.Retention(24*time.Hour),
	)

	_, err := r.queue.Enqueue(task)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil
	}
	return err
}
