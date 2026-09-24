package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul-5/app/service"
	"modul-5/helper"
	"modul-5/middleware"
	"modul-5/route"
)

// NewApp merakit seluruh komponen aplikasi Fiber.
func NewApp(
	logger *slog.Logger,
	pool *pgxpool.Pool,
	jwtManager *helper.JWTManager,
	studentService *service.StudentService,
	achievementService *service.AchievementService,
	authService *service.AuthService,
) *fiber.App {
	app := fiber.New(fiber.Config{
		// ErrorHandler kustom agar format pesan error seragam (JSON)
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return helper.Fail(c, code, err.Error())
		},
	})

	allowedOrigins := GetEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")

	// Pasang middleware global
	middleware.Register(app, logger, allowedOrigins)

	// Daftarkan seluruh rute API melalui struct Dependencies
	route.Register(app, route.Dependencies{
		Pool:               pool,
		JWT:                jwtManager,
		StudentService:     studentService,
		AchievementService: achievementService,
		AuthService:        authService,
	})

	return app
}
