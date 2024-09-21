package services

import (
	"ricin9/fiber-chat/config"
	"slices"
	"time"
)

type Message struct {
	ID        int
	Content   string
	CreatedAt time.Time
	UserID    int
	Username  string
}

func GetMessages(uid int, roomID int, page int) (messages []Message, err error) {
	db := config.Db

	rows, err := db.Query("SELECT m.message_id, m.content, m.created_at, u.user_id, u.username FROM messages m JOIN users u ON m.user_id = u.user_id WHERE m.room_id = ? ORDER BY m.message_id DESC limit 50 offset ?", roomID, (page-1)*50)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var message Message
		err := rows.Scan(&message.ID, &message.Content, &message.CreatedAt, &message.UserID, &message.Username)
		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}
	slices.Reverse(messages)

	return messages, nil
}
