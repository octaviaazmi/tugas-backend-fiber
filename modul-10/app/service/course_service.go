package service

import (
	"modul-10/app/repository"
)

type CourseService struct {
	courses repository.CourseRepository
}

func NewCourseService(courses repository.CourseRepository) *CourseService {
	return &CourseService{courses: courses}
}
