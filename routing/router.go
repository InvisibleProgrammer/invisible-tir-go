package routing

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(app *fiber.App) {

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

}
