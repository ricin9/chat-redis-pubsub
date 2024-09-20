package handlers

import (
	"fmt"
	"log"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"
	"ricin9/fiber-chat/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CreateRoomInput struct {
	Name  string `validate:"required,min=3,max=32"`
	Users string `validate:"required"`
}

func GetRoom(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int64)
	roomID, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	db := config.Db

	err = db.QueryRow("select 1 from room_users where room_id = ? and user_id = ?", roomID, uid).Err()
	if err != nil {
		return c.Format("you are not a member of this room")
	}

	rows, err := db.Query("SELECT m.message_id, m.content, m.created_at, u.user_id, u.username FROM messages m JOIN users u ON m.user_id = u.user_id WHERE m.room_id = ? ORDER BY m.message_id ASC", roomID)
	if err != nil {
		return c.Format("error fetching messages")
	}

	type Message struct {
		ID        int
		Content   string
		CreatedAt time.Time
		UserID    int
		Username  string
	}

	var messages []Message
	for rows.Next() {
		var message Message
		err := rows.Scan(&message.ID, &message.Content, &message.CreatedAt, &message.UserID, &message.Username)
		if err != nil {
			return c.Format("error scanning messages")
		}

		messages = append(messages, message)
	}

	if c.Get("HX-Request") != "" {
		return c.Render("partials/messages", fiber.Map{"Messages": messages, "RoomID": roomID})
	}

	rooms, err := services.GetRoomsFor(uid)
	if err != nil {
		log.Println("Error getting rooms: ", err)
	}

	return c.Render("pages/index", fiber.Map{"Messages": messages, "Rooms": rooms, "RoomID": roomID}, "layouts/base")
}

func CreateRoom(c *fiber.Ctx) error {
	validate := config.Validate

	room := &CreateRoomInput{
		Name:  c.FormValue("name"),
		Users: c.FormValue("users"),
	}

	err := validate.Struct(room)
	if err != nil {
		errors := utils.FormatErrors(err)
		return c.Render("partials/signup-form", fiber.Map{"Errors": errors, "Room": room})
	}

	users := strings.Split(room.Users, ",")

	usersAny := make([]interface{}, len(users)) // hack to spread users in sql query parameter, users... doesn't work

	for i, v := range users {
		usersAny[i] = strings.TrimSpace(v)
	}

	db := config.Db

	res, err := db.Exec("INSERT INTO rooms (name) VALUES (?)", room.Name)
	if err != nil {
		return c.Format("Error creating room")
	}

	roomID, err := res.LastInsertId()
	if err != nil {
		return c.Format("Error getting last insert id")
	}

	sqlInStatement := strings.Repeat("?,", len(users))
	sqlInStatement, _ = strings.CutSuffix(sqlInStatement, ",") // remove trailing comma

	uid := c.Locals("uid").(int64)

	args := []any{roomID, uid}
	args = append(args, usersAny...)

	_, err = db.Exec(fmt.Sprintf(`INSERT INTO room_users (room_id, user_id)
	 select ? as room_id, user_id from users where user_id = ? OR username IN (%s)`, sqlInStatement), args...)

	if err != nil {
		log.Println("Error adding users to room: ", err)
		return c.Format("Error adding users to room")
	}

	c.Set("HX-Redirect", "/")
	return c.SendStatus(201)
}
