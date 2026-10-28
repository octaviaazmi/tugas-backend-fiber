package helper

import (
	"github.com/gofiber/fiber/v2"

	"modul-10/app/model"
)

const LocalsAuthUser = "authUser"

// CurrentUser mengambil data user yang sedang login dari token JWT
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}
