package routes

import (
	"fmt"
	"ricin9/fiber-chat/auth"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/handlers"
	"ricin9/fiber-chat/middleware"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {

	app.Get("/", middleware.Authenticate, func(c *fiber.Ctx) error {
		uid := c.Locals("uid").(int64)

		db := config.Db

		rows, err := db.Query("SELECT r.room_id, r.name FROM rooms r JOIN room_users ru ON r.room_id = ru.room_id WHERE ru.user_id = ?", uid)
		if err != nil {
			fmt.Println("[/ Rooms] err: ", err)
			return c.Format("error fetching chat rooms")
		}
		defer rows.Close()

		type Room struct {
			ID   int
			Name string
		}
		var rooms []Room
		for rows.Next() {
			var room Room
			err := rows.Scan(&room.ID, &room.Name)
			if err != nil {
				return c.Format("error scanning chat rooms")
			}

			rooms = append(rooms, room)
		}

		return c.Render("pages/index", fiber.Map{"Rooms": rooms}, "layouts/base")
	})

	app.Get("/rooms/:id", middleware.Authenticate, func(c *fiber.Ctx) error {
		uid := c.Locals("uid").(int64)
		roomID, err := c.ParamsInt("id")
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		db := config.Db

		err = db.QueryRow("select 1 from room_users where room_id = ? and user_id = ?", roomID, uid).Err()
		if err != nil {
			return c.Format("you are not a member of this room")
		}

		rows, err := db.Query("SELECT m.message_id, m.content, m.created_at, u.user_id, u.username FROM messages m JOIN users u ON m.user_id = u.user_id WHERE m.room_id = ? ORDER BY m.message_id ASC", roomID)
		if err != nil {
			return c.Format("error fetching messages")
		}

		type Message struct {
			ID        int
			Content   string
			CreatedAt time.Time
			UserID    int
			Username  string
		}

		var messages []Message
		for rows.Next() {
			var message Message
			err := rows.Scan(&message.ID, &message.Content, &message.CreatedAt, &message.UserID, &message.Username)
			if err != nil {
				return c.Format("error scanning messages")
			}

			messages = append(messages, message)
		}

		if c.Get("HX-Request") != "" {
			return c.Render("partials/messages", fiber.Map{"Messages": messages, "RoomID": roomID})
		}

		// todo, fill Rooms data
		return c.Render("pages/index", fiber.Map{"Messages": messages}, "layouts/base")
	})

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("pages/login", fiber.Map{"Guest": true}, "layouts/base")
	})
	app.Post("/login", auth.Login)

	app.Get("/signup", func(c *fiber.Ctx) error {
		return c.Render("pages/signup", fiber.Map{"Guest": true}, "layouts/base")
	})
	app.Post("/signup", auth.Signup)

	app.Get("/logout", auth.Logout)
	// Create a /api/v1 endpoint
	v1 := app.Group("/api/v1")

	// websockets

	v1.Use("/ws", middleware.Authenticate, func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	v1.Get("/ws", websocket.New(handlers.Websocket))

	app.Static("/", "./static")

}
