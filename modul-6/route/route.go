package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul6/app/service"
	"modul6/helper"
	"modul6/middleware"
)

type Dependencies struct {
	Pool           *pgxpool.Pool
	JWT            *helper.JWTManager
	Permissions    *helper.PermissionSet
	AuthService    *service.AuthService
	UserService    *service.UserService
	StudentService *service.StudentService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return helper.Success(c, fiber.StatusOK, "server hidup", nil)
	})

	// --- autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", deps.AuthService.Login)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	perms := deps.Permissions

	// --- users: wajib login, hak diperiksa per endpoint ---
	users := api.Group("/users", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))

	users.Get("/", middleware.RequirePermission(perms, "user:list"), deps.UserService.List)
	users.Get("/:id", deps.UserService.Get)    // ownership diperiksa di service
	users.Put("/:id", deps.UserService.Update) // ownership diperiksa di service
	users.Delete("/:id", middleware.RequirePermission(perms, "user:delete"), deps.UserService.Delete)
	users.Patch("/:id/role", middleware.RequirePermission(perms, "role:assign"), deps.UserService.AssignRole)

	// --- students: wajib login, hak diperiksa per endpoint ---
	students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))

	students.Get("/", middleware.RequirePermission(perms, "student:list"), deps.StudentService.List)
	students.Post("/", middleware.RequirePermission(perms, "student:create"), deps.StudentService.Create)
	students.Get("/:id", deps.StudentService.Get)    // ownership diperiksa di service
	students.Put("/:id", deps.StudentService.Update) // ownership diperiksa di service
	students.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), deps.StudentService.Delete)
}
