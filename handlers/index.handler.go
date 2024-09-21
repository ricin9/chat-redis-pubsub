package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func IndexPage(c *fiber.Ctx) error {
	return c.Redirect("/rooms/1")
}
