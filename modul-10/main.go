package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"modul-10/app/model"
	"modul-10/app/repository"
	"modul-10/app/service"
	"modul-10/config"
	"modul-10/database"
	"modul-10/helper"
	"modul-10/middleware"
	"modul-10/route"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Load .env & Setup Logger
	config.LoadEnv()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	ctx := context.Background()

	// 2. Koneksi ke PostgreSQL
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("Gagal konek Postgres: %v", err)
	}
	defer pool.Close()
	log.Println("✅ Berhasil konek ke PostgreSQL")

	// 3. Koneksi ke MongoDB
	mongoClient, err := database.NewMongoClient(ctx)
	if err != nil {
		logger.Error("Gagal konek MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer mongoClient.Disconnect(context.Background())
	log.Println("✅ Berhasil konek ke MongoDB")

	mongoDB := mongoClient.Database(config.GetEnv("MONGO_DATABASE", "praktikum_backend"))

	// 4. Inisialisasi Repository
	courseRepo := repository.NewCourseRepository(pool)
	reviewRepo := repository.NewReviewRepository(mongoDB)
	portfolioRepo := repository.NewPortfolioRepository(mongoDB)

	// 5. BUAT INDEX MONGODB (Wajib dijalankan saat start)
	if err := reviewRepo.EnsureIndexes(ctx); err != nil {
		logger.Error("Gagal membuat index Review", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := portfolioRepo.EnsureIndexes(ctx); err != nil {
		logger.Error("Gagal membuat index Portfolio", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Println("✅ Index MongoDB siap")

	// 6. Inisialisasi Service
	courseService := service.NewCourseService(courseRepo)
	reviewService := service.NewReviewService(reviewRepo, courseRepo)
	portfolioService := service.NewPortfolioService(pool, portfolioRepo)

	// 7. Inisialisasi JWT Manager
	jwtManager := helper.NewJWTManager(
		config.GetEnv("JWT_SECRET", "rahasia_modul10_super_panjang_sekali_32_karakter"),
		config.GetEnv("JWT_ISSUER", "modul-10"),
		60*time.Minute,
	)

	// 8. Setup Aplikasi Fiber
	app := fiber.New(fiber.Config{
		AppName:   "Modul 10 - MongoDB",
		BodyLimit: 1 * 1024 * 1024,
	})
	middleware.Register(app, config.GetEnv("ALLOWED_ORIGINS", ""))

	// =====================================================================
	// ROUTE SEMENTARA BUAT TEST POSTMAN (Hapus/comment jika sudah production)
	// =====================================================================
	app.Post("/dev/token", func(c *fiber.Ctx) error {
		// Generate token bohongan dengan Role mahasiswa (ID: 1)
		token, _ := jwtManager.GenerateAccess(model.User{
			ID:    1,
			Email: "mhs1@siakad.test",
			Role:  "mahasiswa",
		})
		return c.JSON(fiber.Map{"access_token": token})
	})
	// =====================================================================

	// 9. Daftarkan Route Utama
	route.Register(app, route.Dependencies{
		Pool:             pool,
		JWT:              jwtManager,
		CourseService:    courseService,
		ReviewService:    reviewService,
		PortfolioService: portfolioService,
	})

	// 10. Jalankan Server di Background
	port := config.GetEnv("APP_PORT", "8080")
	go func() {
		logger.Info("🚀 Server berjalan", slog.String("port", port))
		if err := app.Listen(":" + port); err != nil {
			logger.Error("Server berhenti", slog.String("error", err.Error()))
		}
	}()

	// 11. Graceful Shutdown (Tunggu Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Sinyal berhenti diterima, menutup server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("Gagal menutup server", slog.String("error", err.Error()))
	}
	logger.Info("Server berhenti dengan rapi")
}
