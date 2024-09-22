package handlers

import (
	"fmt"
	"log"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"

	"github.com/gofiber/fiber/v2"
)

func AddRoomMember(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)
	roomID, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	db := config.Db

	var exists bool
	err = db.QueryRow("select 1 from room_users where room_id = ? and user_id = ? and admin = 1 and 1 = 0",
		roomID, uid).Scan(&exists)

	if err != nil || !exists {
		return c.Format("you are not an admin of this room")
	}

	// TODO

	// add user to room
	// publish event to room
	// publish event to user

	return c.Format("cool beans")
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
	err = db.QueryRow("select username from room_users join users using (user_id) where room_id = ? and user_id = ? and admin = 1",
		roomID, uid).Scan(&username)

	if err != nil {
		return c.Format("you are not an admin of this room")
	}

	var memberUsername string
	err = db.QueryRow("select username from room_users join users using (user_id) where room_id = ? and user_id = ?", roomID, memberId).Scan(&memberUsername)
	if err != nil {
		return c.Format("user does not belong to this room")
	}

	_, err = db.Exec("delete from room_users where room_id = ? and user_id = ?", roomID, memberId)
	if err != nil {
		// maybe use member li template with message error
		return c.Format("there was an error deleting the user")
	}

	message := fmt.Sprintf("%s has kicked %s", username, memberUsername)
	err = services.PersistPublishMessage(0, services.WsIncomingMessage{RoomID: roomID, Content: message})
	if err != nil {
		log.Println(err)
		return c.Format("failed to notify users of kicking")
	}

	return c.SendStatus(204)
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
	err = db.QueryRow("select username from room_users join users using (user_id) where room_id = ? and user_id = ? and admin = 1",
		roomID, uid).Scan(&username)

	if err != nil {
		return c.Format("you are not an admin of this room")
	}

	var memberUsername string
	err = db.QueryRow("select username from room_users join users using (user_id) where room_id = ? and user_id = ?", roomID, memberId).Scan(&memberUsername)
	if err != nil {
		return c.Format("user does not belong to this room")
	}

	_, err = db.Exec("update room_users set admin = 1 where room_id = ? and user_id = ?", roomID, memberId)
	if err != nil {
		// maybe use member li template with message error
		return c.Format("there was an error promoting the user")
	}

	message := fmt.Sprintf("%s has promoted %s to admin", username, memberUsername)
	err = services.PersistPublishMessage(0, services.WsIncomingMessage{RoomID: roomID, Content: message})
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
	err = db.QueryRow("select username from room_users join users using (user_id) where room_id = ? and user_id = ? and admin = 1",
		roomID, uid).Scan(&username)

	if err != nil {
		return c.Format("you are not an admin of this room")
	}

	var memberUsername string
	err = db.QueryRow("select username from room_users join users using (user_id) where room_id = ? and user_id = ?", roomID, memberId).Scan(&memberUsername)
	if err != nil {
		return c.Format("user does not belong to this room")
	}

	_, err = db.Exec("update room_users set admin = 0 where room_id = ? and user_id = ?", roomID, memberId)
	if err != nil {
		// maybe use member li template with message error
		return c.Format("there was an error demoting the user")
	}
	message := fmt.Sprintf("%s has demoted %s", username, memberUsername)
	err = services.PersistPublishMessage(0, services.WsIncomingMessage{RoomID: roomID, Content: message})
	if err != nil {
		log.Println(err)
		return c.Format("failed to notify users of demotion")
	}

	return c.Render("partials/room-info-member-li", fiber.Map{"Username": memberUsername,
		"Admin": false, "RoomID": roomID, "ID": memberId})
}
