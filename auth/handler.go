package auth

import (
	"fmt"
	"ricin9/fiber-chat/config"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type (
	User struct {
		Username string `validate:"required,min=3,max=32"`
		Password string `validate:"required,min=6,max=32"`
	}
)

func Signup(c *fiber.Ctx) error {
	validate := config.Validate

	user := &User{
		Username: c.FormValue("username"),
		Password: c.FormValue("password"),
	}

	err := validate.Struct(user)
	if err != nil {
		// todo, function to format error
		msg := fmt.Sprintf("Validation error: %s", err.Error())
		return c.Format(msg)
	}

	db := config.Db

	res, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, user.Password)
	if err != nil {
		return c.Format("Error inserting user")
	}

	userId, err := res.LastInsertId()
	if err != nil {
		return c.Format("Error getting last insert id")
	}

	return c.Format("aight, user valid, user id : " + strconv.FormatInt(userId, 10))
}
