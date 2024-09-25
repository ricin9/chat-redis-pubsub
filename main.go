package main

import (
	"embed"
	"net/http"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/handlers"
	"ricin9/fiber-chat/utils"

	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	//go:embed static
	embedStaticDir embed.FS

	//go:embed migrations/*
	embedMigrationsDir embed.FS
)

func main() {
	config.Setup()
	utils.Migrate(embedMigrationsDir)

	// Create fiber app
	app := fiber.New(fiber.Config{
		PassLocalsToViews: true,
	})

	// Global Middleware
	app.Use(recover.New())
	app.Use(logger.New())

	// Routes
	handlers.Setup(app)

	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(embedStaticDir),
		PathPrefix: "/static",
	}))

	// Listen on port $PORT or 3000
	log.Fatal(app.Listen(config.Port))
}
