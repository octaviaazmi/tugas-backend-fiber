package helper

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// CurrentUserID adalah simulasi untuk mengambil ID user yang login.
// Di aplikasi nyata, ini diambil dari token JWT.
// Untuk testing, kita ambil dari Header "X-User-ID".
func CurrentUserID(c *fiber.Ctx) int {
	idStr := c.Get("X-User-ID", "1") // Default ke ID 1 jika header kosong
	id, _ := strconv.Atoi(idStr)
	return id
}
