package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	"modul-12/config"
	_ "modul-12/docs"
)

// @title           API KRS Praktikum Backend Lanjut
// @version         1.0
// @description     Ini adalah dokumentasi API untuk Modul 12.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	config.LoadEnv()

	app := fiber.New()

	// Cek apakah SWAGGER_ENABLED bernilai "true"
	if config.GetEnv("SWAGGER_ENABLED", "false") == "true" {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Endpoint testing dasar
	app.Get("/api/v1/health", healthCheck)

	app.Listen(":" + config.GetEnv("APP_PORT", "8080"))
}

// HealthCheck godoc
// @Summary      Cek status server
// @Description  Mengembalikan status server apakah berjalan atau tidak
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /health [get]
func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "message": "Server berjalan"})
}
