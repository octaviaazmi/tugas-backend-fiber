package service

import (
	"context"
	"errors"
	"modul-12/app/repository"
)

type StudentService struct {
	repo repository.StudentRepository
}

func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

func (s *StudentService) GetStudentName(ctx context.Context, id int) (string, error) {
	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if student.Nama == "" {
		return "", errors.New("nama tidak boleh kosong")
	}
	return student.Nama, nil
}
