package service

import (
	"github.com/gofiber/fiber/v2"

	"modul-11/app/repository"
	"modul-11/helper"
)

type NotificationService struct {
	notifications repository.NotificationRepository
}

func NewNotificationService(notifications repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifications: notifications}
}

// GET /api/v1/notifications/mine
func (s *NotificationService) Mine(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID := helper.CurrentUserID(c)

	list, err := s.notifications.FindByUser(ctx, userID, 20)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "gagal mengambil notifikasi"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "notifikasi berhasil diambil",
		"data":    list,
	})
}
