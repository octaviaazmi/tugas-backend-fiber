package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"siakad-mini-fiber/app/repository"
	"siakad-mini-fiber/app/service"
	"siakad-mini-fiber/config"
	"siakad-mini-fiber/database"
	"siakad-mini-fiber/helper"
	"siakad-mini-fiber/middleware"
	"siakad-mini-fiber/route"
	"siakad-mini-fiber/seeder"

	"github.com/gofiber/fiber/v2"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := database.NewPool(ctx)
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool, "migrations"); err != nil {
		logger.Error("migrasi gagal", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// ---- perakitan ----
	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "siakad-mini"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 60))*time.Minute,
	)

	authService := service.NewAuthService(userRepo, studentRepo, jwtManager)
	studentService := service.NewStudentService(pool, studentRepo, userRepo, courseRepo, enrollmentRepo)
	courseService := service.NewCourseService(courseRepo)
	enrollmentService := service.NewEnrollmentService(pool, enrollmentRepo, studentRepo, courseRepo)

	// ---- seeder ----
	if err := seeder.Run(ctx, pool, studentRepo, courseRepo); err != nil {
		logger.Error("seeder gagal", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// ---- aplikasi ----
	app := fiber.New(fiber.Config{
		AppName:   config.GetEnv("APP_NAME", "siakad-mini"),
		BodyLimit: 1 * 1024 * 1024,
	})

	middleware.Register(app, config.GetEnv("ALLOWED_ORIGINS", ""))

	route.Register(app, route.Dependencies{
		Pool:              pool,
		JWT:               jwtManager,
		AuthService:       authService,
		StudentService:    studentService,
		CourseService:     courseService,
		EnrollmentService: enrollmentService,
	})

	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "Endpoint tidak ditemukan")
	})

	port := config.GetEnv("APP_PORT", "8080")
	go func() {
		logger.Info("server berjalan", slog.String("port", port))
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("sinyal berhenti diterima, menutup server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("gagal menutup server dengan rapi", slog.String("error", err.Error()))
	}
	logger.Info("server berhenti dengan rapi")
}
