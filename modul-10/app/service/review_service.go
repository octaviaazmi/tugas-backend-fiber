package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"modul-10/app/model"
	"modul-10/app/repository"
	"modul-10/helper"
)

type ReviewService struct {
	reviews repository.ReviewRepository
	courses repository.CourseRepository
}

func NewReviewService(reviews repository.ReviewRepository, courses repository.CourseRepository) *ReviewService {
	return &ReviewService{reviews: reviews, courses: courses}
}

// POST /courses/:id/reviews
func (s *ReviewService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, _ := helper.CurrentUser(c)
	courseID, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id kelas harus berupa angka positif")
	}

	var req model.CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.Comment = strings.TrimSpace(req.Comment)
	req.Tags = NormalizeTags(req.Tags)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	// Cek kelas ada di Postgres (pakai int64 langsung, jangan di-cast ke int)
	if _, err := s.courses.FindByID(ctx, courseID); err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "kelas tidak ditemukan")
	}

	// Cek apakah mahasiswa terdaftar di kelas ini (pakai int64)
	enrolled, err := s.courses.IsEnrolled(ctx, courseID, current.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memeriksa pendaftaran")
	}
	if !enrolled {
		return helper.Fail(c, fiber.StatusForbidden, "hanya peserta kelas yang boleh menulis ulasan")
	}

	// MongoDB model pakai int, jadi di sini baru kita cast ke int
	created, err := s.reviews.Create(ctx, model.Review{
		CourseID: int(courseID),
		UserID:   int(current.UserID),
		Username: current.Email,
		Rating:   req.Rating,
		Comment:  req.Comment,
		Tags:     req.Tags,
	})
	if err != nil {
		return translateReviewError(c, err)
	}
	return helper.Created(c, "ulasan berhasil disimpan", created, "/api/v1/reviews/"+created.ID)
}

// GET /courses/:id/reviews
func (s *ReviewService) ListByCourse(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	courseID, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id kelas tidak valid")
	}

	q := model.ReviewQuery{
		CourseID: int(courseID), // Cast ke int untuk MongoDB
		Tag:      strings.ToLower(strings.TrimSpace(c.Query("tag"))),
		Before:   c.Query("before"),
		Limit:    c.QueryInt("limit", 10),
	}
	if q.Limit < 1 || q.Limit > 50 {
		q.Limit = 10
	}

	want := q.Limit
	q.Limit = want + 1 // Ambil 1 lebih banyak untuk cek has_more

	reviews, err := s.reviews.FindByCourse(ctx, q)
	if err != nil {
		return translateReviewError(c, err)
	}

	hasMore := len(reviews) > want
	if hasMore {
		reviews = reviews[:want]
	}

	next := ""
	if hasMore {
		next = reviews[len(reviews)-1].ID
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "daftar ulasan berhasil diambil",
		"data":    reviews,
		"meta":    fiber.Map{"limit": want, "has_more": hasMore, "next_before": next},
	})
}

// GET /courses/:id/reviews/summary
func (s *ReviewService) Summary(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	courseID, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id kelas tidak valid")
	}

	// Cek ke Postgres (int64)
	if _, err := s.courses.FindByID(ctx, courseID); err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "kelas tidak ditemukan")
	}

	// Panggil Mongo (int)
	summary, err := s.reviews.Summary(ctx, int(courseID))
	if err != nil {
		return translateReviewError(c, err)
	}
	return helper.Success(c, fiber.StatusOK, "ringkasan ulasan berhasil diambil", summary)
}

// GET /reviews/:id
func (s *ReviewService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	review, err := s.reviews.FindByID(ctx, c.Params("id"))
	if err != nil {
		return translateReviewError(c, err)
	}
	return helper.Success(c, fiber.StatusOK, "ulasan ditemukan", review)
}

// PATCH /reviews/:id
func (s *ReviewService) Update(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, _ := helper.CurrentUser(c)

	var req model.UpdateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus JSON valid")
	}
	if req.Tags != nil {
		normalized := NormalizeTags(*req.Tags)
		req.Tags = &normalized
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	// current.UserID bertipe int64, sedangkan UpdateOwn minta int
	updated, err := s.reviews.UpdateOwn(ctx, c.Params("id"), int(current.UserID), req)
	if err != nil {
		return translateReviewError(c, err)
	}

	return helper.Success(c, fiber.StatusOK, "ulasan berhasil diperbarui", updated)
}

// DELETE /reviews/:id
func (s *ReviewService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, _ := helper.CurrentUser(c)

	// TODO: Cek role admin dari JWT (anyOwner)
	anyOwner := false

	if err := s.reviews.Delete(ctx, c.Params("id"), int(current.UserID), anyOwner); err != nil {
		return translateReviewError(c, err)
	}
	return helper.NoContent(c)
}

// POST /reviews/:id/replies
func (s *ReviewService) Reply(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, _ := helper.CurrentUser(c)

	var req model.ReplyRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus JSON valid")
	}
	req.Body = strings.TrimSpace(req.Body)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	// current.UserID bertipe int64, Mongo minta int
	updated, err := s.reviews.AddReply(ctx, c.Params("id"), model.Reply{
		UserID: int(current.UserID), Username: current.Email, Body: req.Body,
	})
	if err != nil {
		return translateReviewError(c, err)
	}

	return helper.Success(c, fiber.StatusCreated, "balasan berhasil ditambahkan", updated)
}

// POST /reviews/:id/helpful
func (s *ReviewService) MarkHelpful(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, _ := helper.CurrentUser(c)
	// current.UserID bertipe int64, Mongo minta int
	updated, err := s.reviews.MarkHelpful(ctx, c.Params("id"), int(current.UserID))
	if err != nil {
		return translateReviewError(c, err)
	}

	return helper.Success(c, fiber.StatusOK, "terima kasih atas penilaiannya", updated)
}

func translateReviewError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.Fail(c, fiber.StatusBadRequest, "id ulasan tidak valid")
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "data tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, "Anda sudah menulis ulasan untuk kelas ini")
	case errors.Is(err, repository.ErrOwnReview):
		return helper.Fail(c, fiber.StatusConflict, "tidak dapat menilai ulasan sendiri")
	case errors.Is(err, repository.ErrAlreadyVoted):
		return helper.Fail(c, fiber.StatusConflict, "Anda sudah menilai ulasan ini")
	case errors.Is(err, repository.ErrReplyLimit):
		return helper.Fail(c, fiber.StatusConflict, "jumlah balasan sudah mencapai batas")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada server")
	}
}
