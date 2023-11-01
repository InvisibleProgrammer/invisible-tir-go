package routing

import "github.com/gofiber/fiber/v2"

func NotFoundHandler(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    403,
			"type":    "NOT_FOUND",
			"message": "Service not found",
		})
	})

}
