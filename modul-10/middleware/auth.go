package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"modul-10/helper"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if header == "" {
			return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak ditemukan")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return helper.Fail(c, fiber.StatusUnauthorized, "Format token tidak valid")
		}

		authUser, err := jwtManager.Parse(parts[1])
		if err != nil {
			return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid atau kadaluarsa")
		}

		// Simpan data user ke context Fiber
		c.Locals(helper.LocalsAuthUser, authUser)
		return c.Next()
	}
}

func RequireJSON(c *fiber.Ctx) error {
	if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut || c.Method() == fiber.MethodPatch {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return helper.Fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
		}
	}
	return c.Next()
}
