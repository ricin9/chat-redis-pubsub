package services

import (
	"context"
	"database/sql"
	"log"
	"ricin9/fiber-chat/config"
	"strconv"
)

type Room struct {
	ID          int
	Name        string
	LastMessage sql.NullTime
}

type Member struct {
	ID       int
	Username string
	Admin    bool
}

func GetRoomsFor(uid int) (rooms []Room, err error) {
	db := config.Db

	rows, err := db.Query(`
	SELECT r.room_id, r.name, (select created_at from messages where room_id = r.room_id order by message_id desc limit 1) as last_message
	FROM rooms r
	JOIN room_users using (room_id)
	WHERE user_id = ?
	ORDER BY ifnull(last_message, r.created_at) DESC`,
		uid)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var room Room
		err := rows.Scan(&room.ID, &room.Name, &room.LastMessage)
		if err != nil {
			return nil, err
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}

func GetRoomMembers(roomID int) (members []Member, err error) {
	db := config.Db

	rows, err := db.Query(`SELECT user_id, username, admin from room_users
	 JOIN users using (user_id) WHERE room_id = ?`, roomID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var member Member
		err := rows.Scan(&member.ID, &member.Username, &member.Admin)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	return members, nil
}

func GetRoomById(ctx context.Context, roomId int) (room Room, err error) {
	db := config.Db

	err = db.QueryRowContext(ctx, "SELECT room_id, name FROM rooms WHERE room_id = ?", roomId).Scan(&room.ID, &room.Name)
	if err != nil {
		return Room{}, err
	}

	return room, nil
}

func RoomIdsForUser(uid int) []string {
	db := config.Db

	rows, err := db.Query("SELECT room_id FROM room_users WHERE user_id = ?", uid)
	if err != nil {
		log.Println("[/ Rooms] err: ", err)
		return nil
	}
	defer rows.Close()

	var rooms []string
	for rows.Next() {
		var roomID int
		err := rows.Scan(&roomID)
		if err != nil {
			log.Println("error scanning chat rooms")
			return nil
		}

		rooms = append(rooms, strconv.Itoa(roomID))
	}

	return rooms
}

func RoomHasUser(roomID int, uid int) bool {
	db := config.Db

	var exists bool
	err := db.QueryRowContext(context.Background(), "select 1 from room_users where room_id = ? and user_id = ?", roomID, uid).Scan(&exists)
	if err != nil {
		return false
	}

	return true
}

func RoomHasMessage(roomID int, messageID int) bool {
	db := config.Db

	var exists bool
	err := db.QueryRowContext(context.Background(), "select 1 from messages where message_id = ? and room_id = ?", messageID, roomID).Scan(&exists)
	if err != nil {
		return false
	}

	return true
}
