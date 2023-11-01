package main

import (
	"invisible-tir-go/infrastructure"
	"invisible-tir-go/routing"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {

	db, err := infrastructure.InitDb()
	if err != nil {
		log.Panicf("Error in initializing db: %v", err)
	}

	log.Println("App starting: http://localhost:3000")

	app := fiber.New()

	routing.RegisterRoutes(app, db)

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
