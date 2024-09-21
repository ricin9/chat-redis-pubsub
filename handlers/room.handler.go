package handlers

import (
	"fmt"
	"log"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"
	"ricin9/fiber-chat/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type CreateRoomInput struct {
	Name  string `validate:"required,min=3,max=32"`
	Users string `validate:"required"`
}

func GetRoom(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)
	roomID, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	db := config.Db

	var roomName string
	err = db.QueryRow("select rooms.name from room_users join rooms using (room_id) where room_id = ? and user_id = ?", roomID, uid).Scan(&roomName)
	if err != nil {
		return c.Format("you are not a member of this room")
	}

	messages, err := services.GetMessages(uid, roomID, 1)
	if err != nil {
		return c.Format("Error getting messages")
	}

	if c.Get("HX-Request") != "" {
		return c.Render("partials/room-content", fiber.Map{"Messages": messages, "RoomID": roomID, "RoomName": roomName})
	}

	rooms, err := services.GetRoomsFor(uid)
	if err != nil {
		log.Println("Error getting rooms: ", err)
	}

	return c.Render("layouts/main", fiber.Map{"Messages": messages, "Rooms": rooms, "RoomID": roomID, "RoomName": roomName})
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
		return c.Render("partials/create-room-form", fiber.Map{"Errors": errors, "Room": room})
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

	uid := c.Locals("uid")

	args := []any{roomID, uid}
	args = append(args, usersAny...)

	fmt.Println("creating room,", room)
	_, err = db.Exec(fmt.Sprintf(`INSERT INTO room_users (room_id, user_id)
	 select ? as room_id, user_id from users where user_id = ? OR username IN (%s)`, sqlInStatement), args...)

	if err != nil {
		log.Println("Error adding users to room: ", err)
		return c.Format("Error adding users to room")
	}

	fmt.Println("room created,", room)
	c.Set("HX-Redirect", "/")
	return c.SendStatus(201)
}
