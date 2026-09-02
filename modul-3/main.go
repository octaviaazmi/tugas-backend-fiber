package main

import (
	"log"

	"modul-3/app/repository"
	"modul-3/config"
	"modul-3/database"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Muat konfigurasi dari file .env
	config.LoadEnv()

	// 2. Inisialisasi koneksi database PostgreSQL pakai pgxpool
	dbPool, err := database.ConnectPostgres()
	if err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}
	defer dbPool.Close()

	// 3. Inisialisasi Repository dan Handler
	studentRepo := repository.NewStudentRepository(dbPool)
	studentHandler := NewStudentHandler(studentRepo)

	// 4. Inisialisasi framework Fiber
	app := fiber.New()

	// 5. Daftarkan Routing API (Sesuai standar RESTful)
	api := app.Group("/api/v1")

	// Endpoint Health Check untuk menguji server dan koneksi database
	api.Get("/health", func(c *fiber.Ctx) error {
		if err := dbPool.Ping(c.UserContext()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":  "error",
				"message": "database tidak dapat dihubungi",
			})
		}
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "server dan database berjalan",
		})
	})

	students := api.Group("/students")
	students.Get("", studentHandler.ListStudents)
	students.Get("/:id", studentHandler.GetStudent)
	students.Post("", studentHandler.CreateStudent)
	students.Put("/:id", studentHandler.ReplaceStudent)
	students.Delete("/:id", studentHandler.DeleteStudent)

	// 6. Jalankan Server di port 3000 (atau sesuai .env)
	port := config.GetEnv("PORT", "3000")
	log.Printf("Server berjalan di port %s...", port)
	log.Fatal(app.Listen(":" + port))
}
