package helper

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestContext memberi timeout 5 detik untuk setiap request ke database
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// ParamID mengambil angka :id dari URL (misal: /courses/1/reviews)
func ParamID(c *fiber.Ctx) (int64, bool) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}
