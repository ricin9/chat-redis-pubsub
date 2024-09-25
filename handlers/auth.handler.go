package handlers

import (
	"fmt"
	"log"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"
	"ricin9/fiber-chat/utils"
	"ricin9/fiber-chat/views/layouts"
	"ricin9/fiber-chat/views/pages"
	"ricin9/fiber-chat/views/partials"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/mattn/go-sqlite3"
)

type (
	User struct {
		Username string `validate:"required,min=3,max=32"`
		Password string `validate:"required,min=6,max=32"`
	}
)

func SignUpView(c *fiber.Ctx) error {
	templHandler := templ.Handler(layouts.AuthLayout("Login - Chat App", pages.Signup()))
	return adaptor.HTTPHandler(templHandler)(c)
}

func LoginView(c *fiber.Ctx) error {
	log.Println("====sup 1")
	templHandler := templ.Handler(layouts.AuthLayout("Login - Chat App", pages.Login()))
	log.Println("====sup 2")
	return adaptor.HTTPHandler(templHandler)(c)
}

func Signup(c *fiber.Ctx) error {
	validate := config.Validate

	user := &User{
		Username: c.FormValue("username"),
		Password: c.FormValue("password"),
	}

	err := validate.Struct(user)
	if err != nil {
		errors := utils.FormatErrors(err)
		templHandler := templ.Handler(partials.SignupForm(
			partials.LoginFormData{Errors: errors, Username: user.Username, Password: user.Password}))
		return adaptor.HTTPHandler(templHandler)(c)
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

	message := fmt.Sprintf("New user has joined, welcome %s", user.Username)
	err = services.PersistPublishMessage(c.Context(), 0, services.WsIncomingMessage{RoomID: 1, Content: message})
	if err != nil {
		log.Println(err)
		return c.Format("failed to notify users of promotion")
	}

	c.Set("HX-Redirect", "/")
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
		templHandler := templ.Handler(partials.LoginForm(
			partials.LoginFormData{Errors: errors, Username: user.Username, Password: user.Password}))
		return adaptor.HTTPHandler(templHandler)(c)
	}

	db := config.Db
	var uid int
	var hash string
	err = db.QueryRowContext(c.Context(), "SELECT user_id, password FROM users WHERE username = ?", user.Username).Scan(&uid, &hash)
	if err != nil {
		templHandler := templ.Handler(partials.LoginForm(partials.LoginFormData{
			Username: user.Username, Password: user.Password, Message: "Invalid username or password"}))
		return adaptor.HTTPHandler(templHandler)(c)
	}

	same, err := utils.ComparePassword(hash, user.Password)
	if err != nil || !same {
		templHandler := templ.Handler(partials.LoginForm(partials.LoginFormData{
			Username: user.Username, Password: user.Password, Message: "Invalid username or password"}))
		return adaptor.HTTPHandler(templHandler)(c)
	}

	err = utils.CreateSession(c, uid)
	if err != nil {
		return c.Format("User created but couldn't create a session")
	}

	c.Set("HX-Redirect", "/")
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
