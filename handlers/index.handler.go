package handlers

import (
	"log"
	"ricin9/fiber-chat/services"

	"github.com/gofiber/fiber/v2"
)

func IndexPage(c *fiber.Ctx) error {
	{
		uid := c.Locals("uid").(int)

		rooms, err := services.GetRoomsFor(uid)
		if err != nil {
			log.Println("error getting rooms for user", err)
		}

		return c.Render("pages/index", fiber.Map{"Rooms": rooms}, "layouts/base")
	}
}
