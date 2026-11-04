package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	"modul-11/app/model"
	"modul-11/app/repository"
	"modul-11/helper"
)

type NotificationWorker struct {
	notifications repository.NotificationRepository
	tx            *repository.TxManager
	mailer        helper.Mailer
}

func NewNotificationWorker(
	notifications repository.NotificationRepository,
	tx *repository.TxManager,
	mailer helper.Mailer,
) *NotificationWorker {
	return &NotificationWorker{notifications: notifications, tx: tx, mailer: mailer}
}

func (w *NotificationWorker) HandleSend(ctx context.Context, t *asynq.Task) error {
	var job model.NotificationJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return fmt.Errorf("%w: payload task rusak", asynq.SkipRetry)
	}

	var userID int
	var payload any

	// Cek tipe event untuk membaca payload yang sesuai
	if job.EventType == model.EventQuotaChanged {
		var e model.QuotaChangedEvent
		if err := json.Unmarshal(job.Payload, &e); err != nil {
			return fmt.Errorf("%w: payload quota rusak", asynq.SkipRetry)
		}
		userID = e.UserID
		payload = e
	} else {
		var e model.EnrollmentEvent
		if err := json.Unmarshal(job.Payload, &e); err != nil {
			return fmt.Errorf("%w: payload enrollment rusak", asynq.SkipRetry)
		}
		userID = e.UserID
		payload = e
	}

	msg, err := BuildMessage(job.EventType, payload)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	alreadySent := false
	err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
			OutboxID: job.OutboxID,
			UserID:   userID,
			Channel:  "email",
			Subject:  msg.Subject,
			Body:     msg.Body,
		})
		if err != nil {
			return err
		}

		if !claimed {
			alreadySent = true
			return nil // Sudah pernah dikirim
		}

		return w.mailer.Send(ctx, msg)
	})

	if errors.Is(err, helper.ErrPermanentDelivery) {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if err != nil {
		return err
	}

	if alreadySent {
		log.Printf("[WORKER] Event %d dilewati (sudah pernah dikirim)\n", job.OutboxID)
	} else {
		log.Printf("[WORKER] Email event %d berhasil dikirim ke %s\n", job.OutboxID, msg.To)
	}

	return nil
}
