package service

import (
	"fmt"

	"modul-11/app/model"
)

// BuildMessage menyusun isi pesan email berdasarkan jenis event
func BuildMessage(eventType string, payload any) (model.Message, error) {
	course := ""
	email := ""

	// Ambil data sesuai tipe event
	switch e := payload.(type) {
	case model.EnrollmentEvent:
		course = e.CourseCode + " " + e.CourseName
		email = e.Email
		if email == "" {
			email = fmt.Sprintf("mhs%d@student.unair.ac.id", e.UserID)
		}
	case model.QuotaChangedEvent:
		course = e.CourseCode + " " + e.CourseName
		email = e.Email
		if email == "" {
			email = fmt.Sprintf("mhs%d@student.unair.ac.id", e.UserID)
		}
	default:
		return model.Message{}, fmt.Errorf("tipe payload tidak dikenal")
	}

	var subject, body string
	switch eventType {
	case model.EventEnrollmentCreated:
		subject = "KRS: berhasil terdaftar di " + course
		body = fmt.Sprintf("Halo mahasiswa, Anda tercatat sebagai peserta kelas %s.", course)
	case model.EventEnrollmentDropped:
		subject = "KRS: keluar dari " + course
		body = fmt.Sprintf("Halo mahasiswa, Anda sudah tidak terdaftar di kelas %s.", course)
	case model.EventQuotaChanged: // <-- BARU
		q := payload.(model.QuotaChangedEvent)
		subject = "KRS: kuota " + q.CourseCode + " berubah"
		body = fmt.Sprintf("Halo mahasiswa, kuota kelas %s sekarang menjadi %d.", course, q.NewQuota)
	default:
		return model.Message{}, fmt.Errorf("jenis event tidak dikenal: %s", eventType)
	}

	return model.Message{To: email, Subject: subject, Body: body}, nil
}
