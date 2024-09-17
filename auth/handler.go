package auth

import (
	"fmt"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mattn/go-sqlite3"
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

	hash, err := utils.HashPassword(user.Password)
	if err != nil {
		return c.Format("Error hashing password")
	}

	db := config.Db

	res, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, hash)
	if err != nil {
		if e, ok := err.(sqlite3.Error); ok && e.Code == sqlite3.ErrConstraint {
			return c.Format("Username already exists")
		}
		return c.Format("Error inserting user")
	}

	uid, err := res.LastInsertId()
	if err != nil {
		return c.Format("Error getting last insert id")
	}

	err = utils.CreateSession(c, uid)
	if err != nil {
		return c.Format("User created but couldn't create a session")
	}

	fmt.Println("the created cookie: ", string(c.Response().Header.PeekCookie("session")))
	return c.Format("aight, user valid, user id : " + strconv.FormatInt(uid, 10))
}
