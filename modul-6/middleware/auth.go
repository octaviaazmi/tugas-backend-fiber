package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"modul6/helper"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return helper.Fail(c, fiber.StatusUnauthorized, "token tidak ditemukan")
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")

		claims, err := jwtManager.Verify(tokenString)
		if err != nil {
			return helper.Fail(c, fiber.StatusUnauthorized, "token tidak valid atau kadaluarsa")
		}

		helper.SetCurrentUser(c, helper.AuthUser{
			UserID: claims.UserID,
			Role:   claims.Role,
		})
		return c.Next()
	}
}
