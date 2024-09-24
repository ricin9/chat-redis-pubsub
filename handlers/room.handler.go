package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"
	"ricin9/fiber-chat/utils"
	"ricin9/fiber-chat/views/layouts"
	"ricin9/fiber-chat/views/pages"
	"ricin9/fiber-chat/views/partials"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
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
	err = db.QueryRowContext(c.Context(), "select rooms.name from room_users join rooms using (room_id) where room_id = ? and user_id = ?", roomID, uid).Scan(&roomName)
	if err != nil {
		return c.Format("you are not a member of this room")
	}

	messages, err := services.GetMessages(c.Context(), uid, roomID, 1)
	if err != nil {
		log.Println("[GET / Rooms/:id] getMessages err: ", err)
		return c.Format("Error getting messages")
	}

	if c.Get("HX-Request") != "" {
		body := pages.RoomContent(services.Room{ID: roomID, Name: roomName}, messages)
		templHandler := templ.Handler(body)
		return adaptor.HTTPHandler(templHandler)(c)
	}

	rooms, err := services.GetRoomsFor(uid)
	if err != nil {
		log.Println("Error getting rooms: ", err)
	}

	roomContent := pages.RoomContent(services.Room{ID: roomID, Name: roomName}, messages)
	body := layouts.Main("Chat App", rooms, roomContent)
	templHandler := templ.Handler(body)
	return adaptor.HTTPHandler(templHandler)(c)
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
		templHandler := templ.Handler(partials.CreateRoomForm(partials.CreateRoomFormData{
			Name:   room.Name,
			Users:  room.Users,
			Errors: errors,
		}))
		return adaptor.HTTPHandler(templHandler)(c)
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

	args := []any{roomID}
	args = append(args, usersAny...)

	_, err = db.Exec("insert into room_users (room_id, user_id, admin) values (?, ?, 1)", roomID, uid)
	if err != nil {
		log.Println(err)
		return c.Format("Error adding users to room")
	}

	newMembers, err := db.Query(fmt.Sprintf(`INSERT INTO room_users (room_id, user_id)
	select ? as room_id, user_id from users where username IN (%s) returning user_id`, sqlInStatement), args...)
	if err != nil {
		log.Println("Error adding users to room: ", err)
		return c.Format("Error adding users to room")
	}

	userIds, err := utils.GetUserIds(newMembers)
	if err != nil {
		log.Println(err)
		return c.Format("error adding users")
	}

	// notify users
	rdb := config.RedisClient

	roomJoinPayload := PSJoinRoom{PSBase: PSBase{Type: CJoinRoom},
		RoomID: int(roomID), Name: room.Name}
	roomChangeJson, err := json.Marshal(roomJoinPayload)
	if err != nil {
		log.Println(err)
		return c.Format("error notifying users of new room")
	}
	// optimize later
	username := services.GetUsername(c.Context(), uid.(int))
	var systemMessages []string

	systemMessages = append(systemMessages, fmt.Sprintf("%s has created room %s", username, room.Name))
	for _, id := range userIds {
		if id != uid {
			systemMessages = append(systemMessages, fmt.Sprintf("%s has added %s to the room", username, services.GetUsername(c.Context(), id)))
		}
	}

	for _, id := range userIds {
		rdb.Publish(context.Background(), "user:"+strconv.Itoa(id), roomChangeJson)
	}

	for _, message := range systemMessages {
		err := services.PersistPublishMessage(c.Context(), 0, services.WsIncomingMessage{RoomID: int(roomID), Content: message})
		if err != nil {
			log.Println(err)
			return c.Format("unknown error has occured")
		}
	}
	// optimize above later, multi sql statements are fast in sqlite anyway, https://www3.sqlite.org/np1queryprob.html
	// but desperately needs refactoring lol

	// this is disconnecting the ws connection, htmx reconnects flawlessly but i want to optimize it
	c.Set("HX-Redirect", fmt.Sprintf("/rooms/%d", roomID))
	return c.SendStatus(201)
}

func GetRoomInfo(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)
	roomID, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	db := config.Db

	var roomName string
	var admin bool
	err = db.QueryRowContext(c.Context(), "select rooms.name, admin from room_users join rooms using (room_id) where room_id = ? and user_id = ?", roomID, uid).Scan(&roomName, &admin)
	if err != nil {
		return c.Format("you are not a member of this room")
	}

	members, err := services.GetRoomMembers(roomID)
	if err != nil {
		return c.Format("Error getting members")
	}

	templHandler := templ.Handler(partials.RoomInfoModal(
		services.Room{ID: roomID, Name: roomName}, members, admin))
	return adaptor.HTTPHandler(templHandler)(c)
}

func LeaveRoom(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)
	roomID, err := c.ParamsInt("id")

	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	db := config.Db

	_, err = db.Exec("delete from room_users where room_id = ? and user_id = ?", roomID, uid)
	if err != nil {
		return c.Format("there was an error leaving the room")
	}

	// notify user of room change
	rdb := config.RedisClient
	roomChangePayload := PSLeaveRoom{PSBase{Type: CLeaveRoom}, roomID}
	roomChangeJson, err := json.Marshal(roomChangePayload)
	if err != nil {
		log.Println(err)
		return c.Format("error notifying users of new room")
	}
	rdb.Publish(c.Context(), "user:"+strconv.Itoa(uid), roomChangeJson)

	message := fmt.Sprintf("%s has left the room", services.GetUsername(c.Context(), uid))
	err = services.PersistPublishMessage(c.Context(), 0, services.WsIncomingMessage{RoomID: roomID, Content: message})
	if err != nil {
		log.Println(err)
		return c.Format("failed to notify users of room leave")
	}

	templHandler := templ.Handler(partials.LeaveRoomOOB(roomID))
	return adaptor.HTTPHandler(templHandler)(c)
}
