package middleware

import (
	"database/sql"
	"errors"
	"log"
	"ricin9/fiber-chat/config"

	"github.com/gofiber/fiber/v2"
)

func redirect(c *fiber.Ctx) error {
	if c.Get("HX-Request") != "" {
		c.Set("HX-Redirect", "/login")
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	return c.Redirect("/login")
}
func Authenticate(c *fiber.Ctx) error {
	sid := c.Cookies("session")
	if sid == "" {
		return redirect(c)
	}

	db := config.Db

	var uid int64
	err := db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ? AND expires_at > datetime('now')", sid).Scan(&uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return redirect(c)
		}
		log.Println("[AUTH MIDDLEWARE] Error querying session: ", err)
		return c.Status(500).Format("Internal Server Error")
	}

	c.Locals("uid", uid)
	return c.Next()
}
