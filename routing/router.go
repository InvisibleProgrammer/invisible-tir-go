package routing

import (
	"invisible-tir-go/cmd/user"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB) {

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Post("/users", func(c *fiber.Ctx) error {
		return user.RegisterUser(c, db)
	})

	app.Post("/users/login", func(c *fiber.Ctx) error {
		return user.LoginUser(c, db)
	})

	app.Get("/users/me", func(c *fiber.Ctx) error {
		return user.GetProfile(c, db)
	})

	app.Put("/users/:id", func(c *fiber.Ctx) error {
		return user.UpdateProfile(c, db)
	})

	app.Put("/users/:id/password", func(c *fiber.Ctx) error {
		return user.UpdatePassword(c, db)
	})

	app.Put("/users/:id/role", func(c *fiber.Ctx) error {
		return user.AddRole(c, db)
	})

	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		return user.DeleteProfile(c, db)
	})

	app.Get("/thematics", func(c *fiber.Ctx) error {
		return thematics.ListThematics(c, db)
	})

	NotFoundHandler(app)
}
