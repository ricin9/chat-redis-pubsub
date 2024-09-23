package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"ricin9/fiber-chat/config"
	"slices"
	"strconv"
	"time"
)

type WsIncomingMessage struct {
	Content string `json:"content"`
	RoomID  int    `json:"room_id"`
	ReplyTo int    `json:"reply_to"`
}

type PublishMessagePayload struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	RoomID    int       `json:"room_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        int
	Content   string
	CreatedAt time.Time
	UserID    int
	Username  string
}

func GetMessages(ctx context.Context, uid int, roomID int, page int) (messages []Message, err error) {
	db := config.Db

	rows, err := db.QueryContext(ctx, "SELECT m.message_id, m.content, m.created_at, ifnull(u.user_id, 0), ifnull(u.username, '') FROM messages m LEFT JOIN users u ON m.user_id = u.user_id WHERE m.room_id = ? ORDER BY m.message_id DESC limit 50 offset ?", roomID, (page-1)*50)

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

func PersistPublishMessage(ctx context.Context, uid int, msg WsIncomingMessage) (err error) {
	db := config.Db

	var replyto sql.NullInt64
	if msg.ReplyTo != 0 {
		replyto = sql.NullInt64{Int64: int64(msg.ReplyTo), Valid: true}
	}

	var muserId sql.NullInt32
	if uid != 0 {
		muserId = sql.NullInt32{Int32: int32(uid), Valid: true}
	}

	result, err := db.Exec("INSERT INTO messages (content, room_id, user_id, reply_to) VALUES (?, ?, ?, ?)", msg.Content, msg.RoomID, muserId, replyto)
	if err != nil {
		return err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	username := ""
	if uid != 0 {
		username = GetUsername(ctx, uid)
	}

	// PSMessageBroadcast, can't import it directly cause dependency circle and i dont want IoC refactor now
	type yeah struct {
		Type int `json:"type"`
		PublishMessagePayload
	}
	outgoing := yeah{
		Type: 1,
		PublishMessagePayload: PublishMessagePayload{
			ID:        int(messageID),
			UserID:    uid,
			RoomID:    msg.RoomID,
			Username:  username,
			Content:   msg.Content,
			CreatedAt: time.Now(),
		}}

	payload, err := json.Marshal(outgoing)
	if err != nil {
		fmt.Println("error marshalling message", err)
		return err
	}

	rdb := config.RedisClient
	rdb.Publish(ctx, strconv.Itoa(msg.RoomID), payload)
	return nil
}
