package services

import (
	"database/sql"
	"ricin9/fiber-chat/config"
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

func GetUserIds(rows *sql.Rows) (ids []int, err error) {
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}
	return ids, nil
}
