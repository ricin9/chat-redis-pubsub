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

	// get users ids from multi select field
	users := strings.Split(c.Context().PostArgs().String(), "&")
	var userIds []any
	for _, user := range users {
		parts := strings.Split(user, "=")
		if len(parts) != 2 || parts[0] != "users" {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		id, err := strconv.Atoi(parts[1])
		if err != nil {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}
		userIds = append(userIds, id)
	}

	db := config.Db
	var username string
	err = db.QueryRowContext(c.Context(), "select username from room_users join users using (user_id) where room_id = ? and user_id = ? and admin = 1",
		roomID, uid).Scan(&username)

	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.ErrForbidden.Code)
	}

	sqlInsertValues := strings.Repeat(fmt.Sprintf("(%d,?),", roomID), len(userIds))
	sqlInsertValues, _ = strings.CutSuffix(sqlInsertValues, ",") // remove trailing comma

	newMembers, err := db.Query(fmt.Sprintf(`INSERT INTO room_users (room_id, user_id)
	values %s on conflict(room_users.room_id, room_users.user_id) do nothing returning user_id`,
		sqlInsertValues), userIds...)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.ErrInternalServerError.Code)
	}

	insertedMembers, err := utils.GetMembersByIds(newMembers)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.ErrInternalServerError.Code)
	}

	// notify users
	rdb := config.RedisClient

	var roomName string
	err = db.QueryRowContext(c.Context(), "select name from rooms where room_id = ?", roomID).Scan(&roomName)
	if err != nil {
		log.Println(err)
		return c.SendStatus(fiber.ErrInternalServerError.Code)
	}
	roomJoinPayload := PSJoinRoom{PSBase: PSBase{Type: CJoinRoom},
		RoomID: int(roomID), Name: roomName}
	roomChangeJson, err := json.Marshal(roomJoinPayload)
	if err != nil {
		log.Println(err)
		return c.Format("error notifying users of new room")
	}
	// optimize later
	var systemMessages []string
	for _, member := range insertedMembers {
		systemMessages = append(systemMessages,
			fmt.Sprintf("%s has added %s to the room", username, member.Username))
		rdb.Publish(c.Context(), "user:"+strconv.Itoa(member.ID), roomChangeJson)
	}

	for _, message := range systemMessages {
		err := services.PersistPublishMessage(c.Context(), 0, services.WsIncomingMessage{RoomID: int(roomID), Content: message})
		if err != nil {
			log.Println(err)
			return c.Format("unknown error has occured")
		}
	}

	templHandler := templ.Handler(partials.AddMemberSuccessOOB(services.Room{ID: roomID},
		insertedMembers))
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

func FindNonMembers(c *fiber.Ctx) error {
	uid := c.Locals("uid").(int)
	roomID, err := c.ParamsInt("roomId")
	query := c.Query("q")

	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if len(query) < 3 {
		return c.JSON([]any{})
	}

	db := config.Db

	_ = uid
	var exists bool
	err = db.QueryRowContext(c.Context(), "select 1 from room_users join users using (user_id) where room_id = ? and user_id = ? and admin = 1",
		roomID, uid).Scan(&exists)

	if err != nil {
		return c.Format("you are not an admin of this room")
	}

	query = query + "%"

	rows, err := db.QueryContext(c.Context(), // left outer join
		`select user_id, username from users
		 left join (select user_id, room_id from room_users where room_id = ?)
		 using (user_id) where room_id is null and username like ?
		 limit 10`,
		roomID, query)

	type User struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	}

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username); err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		users = append(users, user)

	}

	return c.JSON(users)
}
