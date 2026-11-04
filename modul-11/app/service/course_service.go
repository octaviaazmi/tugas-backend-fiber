package service

import (
	"context"

	"modul-11/app/model"
	"modul-11/app/repository"
)

type CourseService struct {
	outbox repository.OutboxRepository
	tx     *repository.TxManager
}

func NewCourseService(outbox repository.OutboxRepository, tx *repository.TxManager) *CourseService {
	return &CourseService{outbox: outbox, tx: tx}
}

func (s *CourseService) Enroll(ctx context.Context, userID int, email string, courseCode string) error {
	return s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		return s.outbox.WithTx(tx).Insert(ctx, model.EventEnrollmentCreated, model.EnrollmentEvent{
			UserID:     userID,
			Email:      email,
			CourseID:   1,
			CourseCode: courseCode,
			CourseName: "Pemrograman Backend",
		})
	})
}

// UpdateQuota <-- BARU (Untuk Tugas Mandiri D.2)
func (s *CourseService) UpdateQuota(ctx context.Context, courseID int, courseCode string, courseName string, newQuota int) error {
	return s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		return s.outbox.WithTx(tx).Insert(ctx, model.EventQuotaChanged, model.QuotaChangedEvent{
			CourseID:   courseID,
			CourseCode: courseCode,
			CourseName: courseName,
			NewQuota:   newQuota,
			// UserID sengaja dikosongkan dulu, nanti Relay yang akan mencarikan untuk semua mahasiswa
		})
	})
}
