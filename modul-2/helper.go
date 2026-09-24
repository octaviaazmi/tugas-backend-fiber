package main

import (
	"strconv"
	"strings"

	""api-students/app/model"

	"github.com/gofiber/fiber/v2"
)

// ==========================================
// FUNGSI BANTUAN UNTUK MENGIRIM RESPONS (AMPLOP)
// ==========================================

func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func okList(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true, Message: message, Data: data, Meta: meta,
	})
}

func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location) // Memberi tahu klien di mana data baru berada
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent) // 204: Berhasil dihapus, tanpa body
}

func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.WebResponse{Success: false, Message: message})
}

func failValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(model.WebResponse{
		Success: false, Message: "validasi gagal", Errors: errs,
	})
}

// ==========================================
// FUNGSI BANTUAN UNTUK MEMBACA QUERY STRING
// ==========================================

// Daftar putih (whitelist) field yang boleh dipakai untuk mengurutkan (sort).
var allowedSort = map[string]bool{
	"id": true, "name": true, "grade": true,
}

// parseListQuery membaca parameter dari URL dan memberi nilai bawaan (default) yang aman.
func parseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   strings.ToLower(c.Query("sort", "id")),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}

	// Batas atas limit untuk mencegah klien mengambil terlalu banyak data sekaligus
	if q.Limit > 50 {
		q.Limit = 50
	}

	// Cek apakah field sort ada di dalam daftar putih, kalau tidak, kembalikan ke "id"
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}

	if q.Order != "desc" {
		q.Order = "asc"
	}

	// Membaca filter is_active jika dikirim oleh klien
	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}

	return q
}
