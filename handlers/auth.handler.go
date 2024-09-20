package handlers

import (
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/utils"

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
		errors := utils.FormatErrors(err)
		return c.Render("partials/signup-form", fiber.Map{"Errors": errors, "User": user})
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

	// add to general chat
	// TODO: add and fix trigger migration, it errors now with sql-migrate
	go func() {
		_, err := db.Exec("INSERT INTO room_users (room_id, user_id) VALUES (?, ?)", 1, uid)
		if err != nil {
			return
		}
	}()
	err = utils.CreateSession(c, int(uid))
	if err != nil {
		return c.Format("User created but couldn't create a session")
	}

	c.Set("HX-Location", "/")
	return c.SendStatus(201)
}

func Login(c *fiber.Ctx) error {
	validate := config.Validate

	user := &User{
		Username: c.FormValue("username"),
		Password: c.FormValue("password"),
	}

	err := validate.Struct(user)
	if err != nil {
		errors := utils.FormatErrors(err)
		return c.Render("partials/signup-form", fiber.Map{"Errors": errors, "User": user})
	}

	db := config.Db
	var uid int
	var hash string
	err = db.QueryRow("SELECT user_id, password FROM users WHERE username = ?", user.Username).Scan(&uid, &hash)
	if err != nil {
		return c.Render("partials/login-form", fiber.Map{"Message": "Invalid username or password", "User": user})
	}

	same, err := utils.ComparePassword(hash, user.Password)
	if err != nil || !same {
		return c.Render("partials/login-form", fiber.Map{"Message": "Invalid username or password", "User": user})
	}

	err = utils.CreateSession(c, uid)
	if err != nil {
		return c.Format("User created but couldn't create a session")
	}

	c.Set("HX-Location", "/")
	return c.SendStatus(200)
}

func Logout(c *fiber.Ctx) error {
	err := utils.DestroySession(c)
	if err != nil {
		return c.Format("Error destroying session")
	}

	c.Set("HX-Redirect", "/login")
	return c.SendStatus(200)
}
