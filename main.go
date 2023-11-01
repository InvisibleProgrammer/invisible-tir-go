package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {

	log.Println("App starting: http://localhost:3000")

	app := fiber.New()

	// routing.RegisterRoutes(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// app.Listen(":3000")

	go func() {
		err := app.Listen(":3000") // Change the port as needed
		if err != nil {
			log.Fatalf("Error starting the server: %v", err)
		}
	}()

	// Log "App started" message after the server has started
	defer func() {
		log.Println("App started")
	}()

	// You can add any other initialization or setup code here

	// Block the main goroutine (e.g., with a select{}) to keep the application running
	select {}

}
