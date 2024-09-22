package handlers

import (
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"

	"github.com/gofiber/fiber/v2"
)

func GetMessages(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)
	roomID, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	page := c.QueryInt("page", 1)

	db := config.Db

	err = db.QueryRow("select 1 from room_users where room_id = ? and user_id = ?", roomID, uid).Err()
	if err != nil {
		return c.Format("you are not a member of this room")
	}

	messages, err := services.GetMessages(uid, roomID, page)
	if err != nil {
		return c.Format("Error getting messages")
	}

	if len(messages) == 0 {
		return c.SendString("")
	}

	return c.Render("partials/message-range-pagination", fiber.Map{"Messages": messages, "RoomID": roomID, "NextPage": page + 1})
}
