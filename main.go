package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"modul6/app/repository"
	"modul6/app/service"
	"modul6/config"
	"modul6/helper"
	"modul6/middleware"
	"modul6/route"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()
	pool, err := config.NewPool(ctx)
	if err != nil {
		logger.Error("gagal konek database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	jwtManager := helper.NewJWTManager(
		config.GetEnv("JWT_SECRET", "ubah-secret-ini"),
		config.GetEnvInt("JWT_TTL_MINUTES", 60),
	)

	userRepository := repository.NewUserRepository(pool)
	roleRepository := repository.NewRoleRepository(pool)
	studentRepository := repository.NewStudentRepository(pool)

	// Permission dimuat SEKALI saat start. Kalau gagal, aplikasi menolak menyala
	// (fail closed) daripada berjalan dengan PermissionSet kosong.
	rawPermissions, err := roleRepository.LoadPermissions(ctx)
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	authService := service.NewAuthService(userRepository, jwtManager, permissions)
	userService := service.NewUserService(userRepository, permissions)
	studentService := service.NewStudentService(studentRepository, permissions)

	app := fiber.New()
	app.Use(middleware.RequestLogger(logger))

	route.Register(app, route.Dependencies{
		Pool:           pool,
		JWT:            jwtManager,
		Permissions:    permissions,
		AuthService:    authService,
		UserService:    userService,
		StudentService: studentService,
	})

	port := config.GetEnv("APP_PORT", "3000")
	logger.Info("server berjalan", slog.String("port", port))
	if err := app.Listen(":" + port); err != nil {
		logger.Error("server berhenti", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
