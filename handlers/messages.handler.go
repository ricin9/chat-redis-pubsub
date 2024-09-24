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
	page := c.QueryInt("page", 1)

	db := config.Db

	var exists bool
	err = db.QueryRowContext(c.Context(), "select 1 from room_users where room_id = ? and user_id = ?", roomID, uid).Scan(&exists)
	if err != nil {
		return c.Format("you are not a member of this room")
	}

	messages, err := services.GetMessages(c.Context(), uid, roomID, page)
	if err != nil {
		return c.Format("Error getting messages")
	}

	if len(messages) == 0 {
		return c.SendString("")
	}

	nextpage := page + 1
	templHandler := templ.Handler(partials.MessagesRange(services.Room{ID: roomID}, messages, nextpage))
	return adaptor.HTTPHandler(templHandler)(c)
}
