package main

import (
	"context"
	"log"
	"log/slog"

	"siakad-mini-fiber/config"
	"siakad-mini-fiber/database"
	"siakad-mini-fiber/middleware"

	"github.com/gofiber/fiber/v2"
)

func main() {
	config.LoadEnv()

	logger := slog.New(slog.NewJSONHandler(log.Writer(), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx := context.Background()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatalf("migrasi: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:   config.GetEnv("APP_NAME", "siakad-mini"),
		BodyLimit: 1 * 1024 * 1024,
	})

	middleware.Register(app, config.GetEnv("ALLOWED_ORIGINS", ""))

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong"})
	})

	port := config.GetEnv("APP_PORT", "8080")
	log.Printf("Server jalan di port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server gagal jalan: %v", err)
	}
}
