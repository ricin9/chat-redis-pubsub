package routes

import (
	"context"
	"fmt"
	"log"
	"ricin9/fiber-chat/auth"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/middleware"
	"ricin9/fiber-chat/views"
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func Setup(app *fiber.App) {

	app.Get("/", middleware.Authenticate, func(c *fiber.Ctx) error {
		fmt.Println(c.Locals("uid"))
		todos := []*views.Todo{}
		index := views.Index(todos)
		handler := adaptor.HTTPHandler(templ.Handler(index))
		return handler(c)
	})

	app.Post("/signup", auth.Signup)
	// Create a /api/v1 endpoint
	v1 := app.Group("/api/v1")

	// websockets

	rdb := config.RedisClient

	v1.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	v1.Get("/ws", websocket.New(func(c *websocket.Conn) {
		var (
			mt  int
			msg []byte
			err error
		)

		chatQuery := c.Query("chats")
		chats := strings.Split(chatQuery, ",")

		pubsub := rdb.Subscribe(context.Background(), chats...)
		ch := pubsub.Channel()
		defer pubsub.Close()

		go func() {
			for msg := range ch {
				if err := c.WriteMessage(1, []byte(msg.Channel+": "+msg.Payload)); err != nil {
					log.Fatalln("write error: ", err)
				}
			}
		}()

		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)

			parts := strings.Split(string(msg), ":")
			log.Println("parts", parts)
			if len(parts) != 2 {
				log.Println("invalid message")
				continue
			}
			chat := parts[0]
			msg := parts[1]

			if err := rdb.Publish(context.Background(), chat, msg).Err(); err != nil {
				log.Println("publish:", err)
				break
			}
			_ = mt
		}

	}))
}
