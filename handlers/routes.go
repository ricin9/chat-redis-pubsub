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
		return c.Render("pages/login", fiber.Map{"Guest": true})
	})
	app.Post("/login", Login)

	app.Get("/signup", func(c *fiber.Ctx) error {
		return c.Render("pages/signup", fiber.Map{"Guest": true}, "layouts/base")
	})
	app.Post("/signup", Signup)

	app.Get("/logout", Logout)

	app.Get("/create-room", middleware.Authenticate, func(c *fiber.Ctx) error {
		return c.Render("pages/create-room", nil, "layouts/base")
	})

	app.Post("/create-room", middleware.Authenticate, CreateRoom)

	app.Get("/rooms/:id/messages", middleware.Authenticate, GetMessages)

	// websockets
	app.Use("/ws", middleware.Authenticate, func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(Websocket))

	app.Static("/", "./static")

}
