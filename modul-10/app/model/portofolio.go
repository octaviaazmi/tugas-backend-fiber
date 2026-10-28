package model

import "time"

type Project struct {
	ProjectID    string   `bson:"project_id" json:"project_id"`
	Title        string   `bson:"title" json:"title"`
	Technologies []string `bson:"technologies" json:"technologies"`
	Link         string   `bson:"link" json:"link"`
	Year         int      `bson:"year" json:"year"`
}

type Certificate struct {
	Name   string    `bson:"name" json:"name"`
	Issuer string    `bson:"issuer" json:"issuer"`
	Date   time.Time `bson:"date" json:"date"`
}

type Portfolio struct {
	ID           string        `bson:"_id,omitempty" json:"id"`
	StudentID    int           `bson:"student_id" json:"student_id"`
	Headline     string        `bson:"headline" json:"headline"`
	Bio          string        `bson:"bio" json:"bio"`
	Projects     []Project     `bson:"projects" json:"projects"`
	Certificates []Certificate `bson:"certificates" json:"certificates"`
	CreatedAt    time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at" json:"updated_at"`
}

// Request structs untuk validasi input
type UpsertPortfolioRequest struct {
	Headline     string        `json:"headline" validate:"required,min=5,max=100"`
	Bio          string        `json:"bio" validate:"required,min=10,max=1000"`
	Certificates []Certificate `json:"certificates"`
}

type AddProjectRequest struct {
	Title        string   `json:"title" validate:"required,min=3,max=100"`
	Technologies []string `json:"technologies" validate:"required,min=1,dive,min=1,max=30"`
	Link         string   `json:"link" validate:"omitempty,url"`
	Year         int      `json:"year" validate:"required,min=2000,max=2100"`
}
