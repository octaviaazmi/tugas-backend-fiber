package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"modul-5/app/repository"
	"modul-5/app/service"
	"modul-5/config"
	"modul-5/database"
	"modul-5/helper"
)

func main() {
	// 1. Inisialisasi Environment & Logger
	config.LoadEnv()
	logger := config.NewLogger()
	logger.Info("memulai aplikasi server...")

	// 2. Inisialisasi Database Connection Pool
	ctx := context.Background()
	dbPool, err := database.NewPool(ctx)
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	// 3. Inisialisasi JWT Manager
	jwtSecret := config.GetEnv("JWT_SECRET", "default-secret-key-yang-panjang")
	issuer := config.GetEnv("JWT_ISSUER", "praktikum-backend")
	accessTTL, _ := time.ParseDuration("15m")
	refreshTTL, _ := time.ParseDuration("168h") // 7 hari
	jwtManager := helper.NewJWTManager(jwtSecret, issuer, accessTTL)

	// 4. Inisialisasi Repository
	studentRepo := repository.NewStudentRepository(dbPool)
	achievementRepo := repository.NewAchievementRepository(dbPool)
	userRepo := repository.NewUserRepository(dbPool)
	tokenRepo := repository.NewTokenRepository(dbPool)

	// 5. Inisialisasi Service
	studentService := service.NewStudentService(studentRepo)
	achievementService := service.NewAchievementService(achievementRepo)
	authService := service.NewAuthService(userRepo, tokenRepo, jwtManager, refreshTTL)

	// 6. Inisialisasi Aplikasi Fiber (Merakit Semua Dependensi)
	app := config.NewApp(
		logger,
		dbPool,
		jwtManager,
		studentService,
		achievementService,
		authService,
	)

	// 7. Jalankan Server di Background (Graceful Shutdown)
	port := config.GetEnv("APP_PORT", "3000")
	go func() {
		logger.Info("server mendengarkan pada port", slog.String("port", port))
		if err := app.Listen(":" + port); err != nil {
			logger.Error("gagal menjalankan server", slog.String("error", err.Error()))
		}
	}()

	// Menunggu sinyal interupsi (Ctrl+C) untuk mematikan server dengan aman
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("mematikan server secara perlahan...")
	_ = app.Shutdown()
	logger.Info("server berhasil dimatikan.")
}
