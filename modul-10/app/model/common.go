package model

// WebResponse adalah amplop standar untuk semua balasan API
type WebResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Meta dipakai untuk respons yang butuh info pagination
type Meta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
}

// CourseListQuery adalah parameter untuk list mata kuliah
type CourseListQuery struct {
	Semester  int
	Search    string
	Available bool
}
