package handlers

import (
	"log"
	"ricin9/fiber-chat/services"
	"ricin9/fiber-chat/views/layouts"
	"ricin9/fiber-chat/views/pages"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func IndexPage(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)

	if c.Get("HX-Request") != "" {
		body := pages.Index()
		templHandler := templ.Handler(body)
		return adaptor.HTTPHandler(templHandler)(c)
	}

	rooms, err := services.GetRoomsFor(uid)
	if err != nil {
		log.Println("Error getting rooms: ", err)
	}

	templHandler := templ.Handler(layouts.Main("Chat App", rooms, pages.Index()))
	return adaptor.HTTPHandler(templHandler)(c)
}
