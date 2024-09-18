package utils

import (
	"ricin9/fiber-chat/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateSession(c *fiber.Ctx, uid int64) error {
	sid := uuid.NewString()
	ip := c.Context().RemoteIP().String()
	userAgent := string(c.Request().Header.UserAgent())
	expires := time.Now().Add(30 * 24 * time.Hour)

	db := config.Db
	_, err := db.Exec("INSERT INTO sessions (session_id, user_id, ip, user_agent, expires_at) VALUES (?, ?, ?, ?, ?)",
		sid, uid, ip, userAgent, expires)
	if err != nil {
		return err
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "session"
	cookie.Value = sid
	cookie.Expires = expires
	cookie.Secure = *config.Prod
	cookie.HTTPOnly = true
	cookie.SameSite = "Strict"

	c.Cookie(cookie)

	return nil
}

func DestroySession(c *fiber.Ctx) error {
	sid := c.Cookies("session")

	db := config.Db
	_, err := db.Exec("DELETE FROM sessions WHERE session_id = ?", sid)
	if err != nil {
		return err
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "session"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-time.Hour)
	cookie.Secure = *config.Prod
	cookie.HTTPOnly = true
	cookie.SameSite = "Strict"

	c.Cookie(cookie)

	return nil
}
