package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul-11/app/repository"
	"modul-11/app/service"
	"modul-11/config"
	"modul-11/database"
)

func main() {
	config.LoadEnv()
	ctx := context.Background()

	// Flag untuk memilih role: api, worker, atau all
	role := flag.String("role", "all", "api | worker | all")
	flag.Parse()

	// Konek Postgres
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("Gagal konek Postgres: %v", err)
	}
	defer pool.Close()

	var stopWorker func()
	// Jalankan Worker (jika role = worker atau all)
	if *role == "worker" || *role == "all" {
		stopWorker = startWorker(ctx, pool)
	}

	// Jalankan API (jika role = api atau all)
	if *role == "api" || *role == "all" {
		go runAPI(pool)
	}

	// Tunggu sinyal Ctrl+C
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Mematikan aplikasi...")

	if stopWorker != nil {
		stopWorker()
	}
	time.Sleep(1 * time.Second)
	log.Println("Aplikasi berhenti dengan rapi.")
}

func runAPI(pool *pgxpool.Pool) {
	app := fiber.New()

	// Init Repositories
	outboxRepo := repository.NewOutboxRepository(pool)
	notifRepo := repository.NewNotificationRepository(pool)
	txManager := repository.NewTxManager(pool)

	// Init Services
	courseService := service.NewCourseService(outboxRepo, txManager)
	notifService := service.NewNotificationService(notifRepo)

	api := app.Group("/api/v1")

	// 1. Endpoint simulasi mendaftar kelas
	api.Post("/enroll", func(c *fiber.Ctx) error {
		var req struct {
			UserID     int    `json:"user_id"`
			Email      string `json:"email"`
			CourseCode string `json:"course_code"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Format body salah"})
		}

		err := courseService.Enroll(c.UserContext(), req.UserID, req.Email, req.CourseCode)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"success": true, "message": "Berhasil daftar kelas. Cek email nanti!"})
	})

	// 2. Endpoint Tugas Mandiri: Ubah Kuota Kelas
	api.Patch("/courses/:id/quota", func(c *fiber.Ctx) error {
		courseID, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "ID kelas harus angka"})
		}

		var req struct {
			Quota int `json:"quota"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Format body salah"})
		}

		// Simulasi ubah kuota (Code dan Name di-hardcode untuk testing)
		err = courseService.UpdateQuota(c.UserContext(), courseID, "IF101", "Pemrograman Backend", req.Quota)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"success": true, "message": "Kuota berhasil diubah. Email sedang dikirim ke semua peserta."})
	})

	// 3. Endpoint melihat notifikasi
	api.Get("/notifications/mine", notifService.Mine)

	port := config.GetEnv("APP_PORT", "8080")
	log.Printf("🚀 API jalan di port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server gagal jalan: %v", err)
	}
}
