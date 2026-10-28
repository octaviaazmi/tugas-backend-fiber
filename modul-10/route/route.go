package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul-10/app/service"
	"modul-10/helper"
	"modul-10/middleware"
)

type Dependencies struct {
	Pool             *pgxpool.Pool
	JWT              *helper.JWTManager
	CourseService    *service.CourseService
	ReviewService    *service.ReviewService
	PortfolioService *service.PortfolioService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// --- HEALTH CHECK ---
	api.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := deps.Pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "PostgreSQL mati")
		}
		return helper.Success(c, fiber.StatusOK, "Server dan PostgreSQL berjalan", nil)
	})

	// --- ROUTE PROTECTED (Butuh Login) ---
	auth := api.Group("", middleware.RequireAuth(deps.JWT))

	// 1. Route Review (Modul 10)
	reviews := auth.Group("/reviews", middleware.RequireJSON)
	reviews.Get("/:id", deps.ReviewService.Get)
	reviews.Patch("/:id", deps.ReviewService.Update)
	reviews.Delete("/:id", deps.ReviewService.Delete)
	reviews.Post("/:id/replies", deps.ReviewService.Reply)
	reviews.Post("/:id/helpful", deps.ReviewService.MarkHelpful)

	coursesReviews := auth.Group("/courses")
	coursesReviews.Get("/:id/reviews", deps.ReviewService.ListByCourse)
	coursesReviews.Get("/:id/reviews/summary", deps.ReviewService.Summary)
	coursesReviews.Post("/:id/reviews", deps.ReviewService.Create)

	// 2. Route Portofolio (Tugas Mandiri)
	portfolios := auth.Group("/students", middleware.RequireJSON)
	portfolios.Get("/:id/portfolio", deps.PortfolioService.Get)
	portfolios.Put("/:id/portfolio", deps.PortfolioService.Upsert)
	portfolios.Post("/:id/portfolio/projects", deps.PortfolioService.AddProject)
	portfolios.Delete("/:id/portfolio/projects/:projectId", deps.PortfolioService.DeleteProject)

	stats := auth.Group("/portfolios")
	stats.Get("/stats/technologies", deps.PortfolioService.Stats)
}
