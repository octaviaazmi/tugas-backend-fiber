package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Daftar metode yang wajib bawa body
var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// requireJSON menolak request berisi body yang Content-Type-nya bukan JSON (Status 415)
func requireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
		}
	}
	return c.Next()
}

func main() {
	app := fiber.New(fiber.Config{
		AppName: "Praktikum Backend Lanjut - Tugas 2 Students",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			pesan := "terjadi kesalahan pada server"
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
				pesan = e.Message
			}
			return fail(c, status, pesan)
		},
	})

	// Middleware global
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${locals:requestid} ${method} ${path} ${status} ${latency}\n",
	}))
	app.Use(cors.New())

	// Endpoint dasar
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("REST API Students - Modul 2 Berjalan!")
	})

	// Grup API versi 1
	api := app.Group("/api/v1")

	// Grup endpoint /students dipasangi middleware requireJSON
	s := api.Group("/students", requireJSON)

	s.Get("/", listStudents)        // GET semua data
	s.Get("/:id", getStudent)       // GET satu data
	s.Post("/", createStudent)      // POST tambah data
	s.Put("/:id", replaceStudent)   // PUT ganti semua data
	s.Patch("/:id", patchStudent)   // PATCH ubah sebagian data
	s.Delete("/:id", deleteStudent) // DELETE hapus data

	// Endpoint (URL) yang tidak dikenal akan dialihkan ke 404
	app.Use(func(c *fiber.Ctx) error {
		return fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	fmt.Println("Server berjalan di http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
