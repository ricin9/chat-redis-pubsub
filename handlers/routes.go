package handlers

import (
	"ricin9/fiber-chat/middleware"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {

	app.Get("/", middleware.Authenticate, IndexPage)

	app.Get("/rooms/:id", middleware.Authenticate, GetRoom)

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("pages/login", fiber.Map{"Guest": true}, "layouts/base")
	})
	app.Post("/login", Login)

	app.Get("/signup", func(c *fiber.Ctx) error {
		return c.Render("pages/signup", fiber.Map{"Guest": true}, "layouts/base")
	})
	app.Post("/signup", Signup)

	app.Get("/logout", Logout)
	// Create a /api/v1 endpoint
	v1 := app.Group("/api/v1")

	app.Get("/create-room", middleware.Authenticate, func(c *fiber.Ctx) error {
		return c.Render("pages/create-room", nil, "layouts/base")
	})

	app.Post("/create-room", middleware.Authenticate, CreateRoom)

	// websockets
	v1.Use("/ws", middleware.Authenticate, func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	v1.Get("/ws", websocket.New(Websocket))

	app.Static("/", "./static")

}
