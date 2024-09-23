package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"
	"ricin9/fiber-chat/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AddRoomMember(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)
	roomID, err := c.ParamsInt("id")

	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	memberUsername := strings.TrimSpace(c.FormValue("username"))

	err = config.Validate.Var(memberUsername, "required,min=3,max=32")
	if err != nil {
		errmsg := "username" + utils.FormatErrors(err)[""]
		return c.Render("partials/room-add-member-form", fiber.Map{"Error": errmsg, "RoomID": roomID, "Username": memberUsername})
	}

	db := config.Db

	var username string
	err = db.QueryRowContext(c.Context(), "select username from room_users join users using (user_id) where room_id = ? and user_id = ? and admin = 1",
		roomID, uid).Scan(&username)

	if err != nil {
		return c.Format("you are not an admin of this room")
	}

	var memberId int
	err = db.QueryRowContext(c.Context(), "select user_id from users where username = ?", memberUsername).Scan(&memberId)
	if err != nil {
		msg := fmt.Sprintf("user %s does not exist", memberUsername)
		return c.Render("partials/room-add-member-form", fiber.Map{"Error": msg, "RoomID": roomID, "Username": memberUsername})
	}

	var exists bool
	err = db.QueryRowContext(c.Context(), "select 1 from room_users join users using (user_id) where room_id = ? and user_id = ?",
		roomID, memberId).Scan(&exists)

	if err == nil {
		msg := fmt.Sprintf("user %s is already a member of this room", memberUsername)
		return c.Render("partials/room-add-member-form", fiber.Map{"Error": msg, "RoomID": roomID, "Username": memberUsername})
	}

	_, err = db.Exec("insert into room_users (room_id, user_id) values (?, ?)", roomID, memberId)

	if err != nil {
		log.Println(err)
		return c.Format("Unknown Error adding user to room")
	}

	room, err := services.GetRoomById(c.Context(), roomID)
	if err != nil {
		log.Println(err)
		c.Format("added member, but failed to notify users of new member")
	}

	// notify user of room change
	rdb := config.RedisClient
	roomJoinPayload := PSJoinRoom{PSBase: PSBase{Type: CJoinRoom},
		RoomID: int(roomID), Name: room.Name}

	roomChangeJson, err := json.Marshal(roomJoinPayload)
	if err != nil {
		log.Println(err)
		return c.Format("error notifying users of new room")
	}

	rdb.Publish(c.Context(), "user:"+strconv.Itoa(memberId), roomChangeJson)

	message := fmt.Sprintf("%s has added %s to the room", username, memberUsername)
	err = services.PersistPublishMessage(c.Context(), 0, services.WsIncomingMessage{RoomID: roomID, Content: message})
	if err != nil {
		log.Println(err)
		return c.Format("failed to notify users of new member addition")
	}

	return c.Render("partials/room-add-member-sucess", fiber.Map{
		"Form":   fiber.Map{"Sucess": "member added successfully", "RoomID": roomID},
		"Member": fiber.Map{"Username": memberUsername, "Admin": false, "RoomID": roomID, "ID": memberId}})
}

func KickMember(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)
	roomID, err := c.ParamsInt("roomId")
	memberId, err := c.ParamsInt("userId")

	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	db := config.Db

	var username string
	err = db.QueryRowContext(c.Context(), "select username from room_users join users using (user_id) where room_id = ? and user_id = ? and admin = 1",
		roomID, uid).Scan(&username)

	if err != nil {
		return c.Format("you are not an admin of this room")
	}

	var memberUsername string
	err = db.QueryRowContext(c.Context(), "select username from room_users join users using (user_id) where room_id = ? and user_id = ?", roomID, memberId).Scan(&memberUsername)
	if err != nil {
		return c.Format("user does not belong to this room")
	}

	_, err = db.Exec("delete from room_users where room_id = ? and user_id = ?", roomID, memberId)
	if err != nil {
		// maybe use member li template with message error
		return c.Format("there was an error deleting the user")
	}

	// notify user of room change
	rdb := config.RedisClient
	roomKickedPayload := PSKickedFromRoom{PSBase: PSBase{Type: CKickedFromRoom},
		RoomID: int(roomID)}

	roomChangeJson, err := json.Marshal(roomKickedPayload)
	if err != nil {
		log.Println(err)
		return c.Format("error notifying users of new room")
	}
	rdb.Publish(c.Context(), "user:"+strconv.Itoa(memberId), roomChangeJson)

	message := fmt.Sprintf("%s has kicked %s", username, memberUsername)
	err = services.PersistPublishMessage(c.Context(), 0, services.WsIncomingMessage{RoomID: roomID, Content: message})
	if err != nil {
		log.Println(err)
		return c.Format("failed to notify users of kicking")
	}

	return c.SendStatus(200)
}

func PromoteMember(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)
	roomID, err := c.ParamsInt("roomId")
	memberId, err := c.ParamsInt("userId")

	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	db := config.Db

	var username string
	err = db.QueryRowContext(c.Context(), "select username from room_users join users using (user_id) where room_id = ? and user_id = ? and admin = 1",
		roomID, uid).Scan(&username)

	if err != nil {
		return c.Format("you are not an admin of this room")
	}

	var memberUsername string
	err = db.QueryRowContext(c.Context(), "select username from room_users join users using (user_id) where room_id = ? and user_id = ?", roomID, memberId).Scan(&memberUsername)
	if err != nil {
		return c.Format("user does not belong to this room")
	}

	_, err = db.Exec("update room_users set admin = 1 where room_id = ? and user_id = ?", roomID, memberId)
	if err != nil {
		// maybe use member li template with message error
		return c.Format("there was an error promoting the user")
	}

	message := fmt.Sprintf("%s has promoted %s to admin", username, memberUsername)
	err = services.PersistPublishMessage(c.Context(), 0, services.WsIncomingMessage{RoomID: roomID, Content: message})
	if err != nil {
		log.Println(err)
		return c.Format("failed to notify users of promotion")
	}

	return c.Render("partials/room-info-member-li", fiber.Map{"Username": memberUsername,
		"Admin": true, "RoomID": roomID, "ID": memberId})
}

func DemoteMember(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)
	roomID, err := c.ParamsInt("roomId")
	memberId, err := c.ParamsInt("userId")

	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	db := config.Db

	var username string
	err = db.QueryRowContext(c.Context(), "select username from room_users join users using (user_id) where room_id = ? and user_id = ? and admin = 1",
		roomID, uid).Scan(&username)

	if err != nil {
		return c.Format("you are not an admin of this room")
	}

	var memberUsername string
	err = db.QueryRowContext(c.Context(), "select username from room_users join users using (user_id) where room_id = ? and user_id = ?", roomID, memberId).Scan(&memberUsername)
	if err != nil {
		return c.Format("user does not belong to this room")
	}

	_, err = db.Exec("update room_users set admin = 0 where room_id = ? and user_id = ?", roomID, memberId)
	if err != nil {
		// maybe use member li template with message error
		return c.Format("there was an error demoting the user")
	}
	message := fmt.Sprintf("%s has demoted %s", username, memberUsername)
	err = services.PersistPublishMessage(c.Context(), 0, services.WsIncomingMessage{RoomID: roomID, Content: message})
	if err != nil {
		log.Println(err)
		return c.Format("failed to notify users of demotion")
	}

	return c.Render("partials/room-info-member-li", fiber.Map{"Username": memberUsername,
		"Admin": false, "RoomID": roomID, "ID": memberId})
}
