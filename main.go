package main

import (
	"context"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/handlers"

	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var ctx = context.Background()

func main() {
	config.Setup()

	// run migrations, dont know where to put this yet
	// migrations := &migrate.FileMigrationSource{
	// 	Dir: "migrations",
	// }

	// Create fiber app
	app := fiber.New(fiber.Config{
		PassLocalsToViews: true,
	})

	// Global Middleware
	app.Use(recover.New())
	app.Use(logger.New())

	// Routes
	handlers.Setup(app)

	// Listen on port 3000
	log.Fatal(app.Listen(config.Port)) // go run app.go -port=:3000
}
