package helper

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// AuthUser adalah identitas hasil verifikasi JWT, dipasang ke request oleh RequireAuth.
type AuthUser struct {
	UserID int
	Role   string
}

const localsKeyUser = "current_user"

func SetCurrentUser(c *fiber.Ctx, user AuthUser) {
	c.Locals(localsKeyUser, user)
}

func CurrentUser(c *fiber.Ctx) (AuthUser, bool) {
	value := c.Locals(localsKeyUser)
	user, ok := value.(AuthUser)
	return user, ok
}

func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Context(), 5*time.Second)
}
