package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"
	"ricin9/fiber-chat/utils"
	"ricin9/fiber-chat/views/partials"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
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
		templHandler := templ.Handler(partials.AddMemberForm(
			services.Room{ID: roomID}, partials.AddMemberFormData{Error: errmsg, Username: memberUsername}))
		return adaptor.HTTPHandler(templHandler)(c)
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
		templHandler := templ.Handler(partials.AddMemberForm(
			services.Room{ID: roomID}, partials.AddMemberFormData{Error: msg, Username: memberUsername}))
		return adaptor.HTTPHandler(templHandler)(c)
	}

	var exists bool
	err = db.QueryRowContext(c.Context(), "select 1 from room_users join users using (user_id) where room_id = ? and user_id = ?",
		roomID, memberId).Scan(&exists)

	if err == nil {
		msg := fmt.Sprintf("user %s is already a member of this room", memberUsername)
		templHandler := templ.Handler(partials.AddMemberForm(
			services.Room{ID: roomID}, partials.AddMemberFormData{Error: msg, Username: memberUsername}))
		return adaptor.HTTPHandler(templHandler)(c)
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

	templHandler := templ.Handler(partials.AddMemberSuccessOOB(services.Room{ID: roomID},
		services.Member{ID: memberId, Username: memberUsername, Admin: false}))
	return adaptor.HTTPHandler(templHandler)(c)
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

	templHandler := templ.Handler(partials.Member(services.Room{ID: roomID},
		services.Member{ID: memberId, Username: memberUsername, Admin: true}, true))
	return adaptor.HTTPHandler(templHandler)(c)
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

	templHandler := templ.Handler(partials.Member(services.Room{ID: roomID},
		services.Member{ID: memberId, Username: memberUsername, Admin: false}, true))
	return adaptor.HTTPHandler(templHandler)(c)
}
