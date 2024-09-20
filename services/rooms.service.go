package services

import (
	"ricin9/fiber-chat/config"
)

type Room struct {
	ID   int
	Name string
}

func GetRoomsFor(uid int64) (rooms []Room, err error) {
	db := config.Db

	rows, err := db.Query("SELECT r.room_id, r.name FROM rooms r JOIN room_users ru ON r.room_id = ru.room_id WHERE ru.user_id = ?", uid)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var room Room
		err := rows.Scan(&room.ID, &room.Name)
		if err != nil {
			return nil, err
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}
