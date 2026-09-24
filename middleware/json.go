package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"modul6/helper"
)

func RequireJSON(c *fiber.Ctx) error {
	if c.Method() == fiber.MethodGet || c.Method() == fiber.MethodDelete {
		return c.Next()
	}
	if !strings.Contains(c.Get("Content-Type"), "application/json") {
		return helper.Fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
	}
	return c.Next()
}
