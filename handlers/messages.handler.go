package handlers

import (
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"
	"ricin9/fiber-chat/views/partials"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func GetMessages(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)
	roomID, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	cursor := c.QueryInt("cursor", 999999999999999999)

	db := config.Db

	var exists bool
	err = db.QueryRowContext(c.Context(), "select 1 from room_users where room_id = ? and user_id = ?", roomID, uid).Scan(&exists)
	if err != nil {
		return c.Format("you are not a member of this room")
	}

	messages, err := services.GetMessages(c.Context(), uid, roomID, cursor)
	if err != nil {
		return c.Format("Error getting messages")
	}

	if len(messages) == 0 {
		return c.SendString("")
	}

	newCursor := messages[0].ID

	templHandler := templ.Handler(partials.MessagesRange(services.Room{ID: roomID}, messages, newCursor))
	return adaptor.HTTPHandler(templHandler)(c)
}
