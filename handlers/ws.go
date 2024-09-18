package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ricin9/fiber-chat/config"
	"strconv"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/redis/go-redis/v9"
)

type PublishMessagePayload struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type IncomingMessage struct {
	Content string `json:"content"`
	RoomID  int    `json:"room_id"`
}

func Websocket(c *websocket.Conn) {
	uid := c.Locals("uid").(int64)
	rooms := userRooms(uid)
	_ = rooms

	rdb := config.RedisClient
	pubsub := rdb.Subscribe(context.Background(), "1")
	defer pubsub.Close()

	ch := pubsub.Channel()

	go handeleSubscription(c, ch)
	handleIncoming(c, uid)

}

func handleIncoming(c *websocket.Conn, uid int64) {
	for {
		var msg IncomingMessage
		if err := c.ReadJSON(&msg); err != nil {
			log.Println("invalid incoming message format", err)
			break
		}

		if !userBelongsTo(uid, msg.RoomID) {
			log.Println("user does not belong to room")
			continue
		}

		messageId, err := persistMessage(uid, msg)
		if err != nil {
			log.Println("error persisting message", err)
			continue
		}

		outgoing := PublishMessagePayload{
			ID:        messageId,
			UserID:    uid,
			Username:  "username", // todo: set and get username in locals
			Content:   msg.Content,
			CreatedAt: time.Now(),
		}

		payload, err := json.Marshal(outgoing)
		if err != nil {
			fmt.Println("error marshalling message", err)
			continue
		}

		rdb := config.RedisClient
		rdb.Publish(context.Background(), strconv.Itoa(msg.RoomID), payload)
	}
}

func handeleSubscription(c *websocket.Conn, ch <-chan *redis.Message) {
	for msg := range ch {
		markup := `<div id="messages" hx-swap-oob="beforeend"><p>hello</p></div>`
		_ = msg
		if err := c.WriteMessage(1, []byte(markup)); err != nil {
			log.Fatalln("write error: ", err)
		}
	}
}

func userRooms(uid int64) []string {
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

func userBelongsTo(uid int64, roomID int) bool {
	db := config.Db

	err := db.QueryRow("select 1 from room_users where room_id = ? and user_id = ?", roomID, uid).Err()
	if err != nil {
		return false
	}

	return true
}

func persistMessage(uid int64, msg IncomingMessage) (messageId int64, err error) {
	db := config.Db

	result, err := db.Exec("INSERT INTO messages (content, room_id, user_id) VALUES (?, ?, ?)", msg.Content, msg.RoomID, uid)
	if err != nil {
		return 0, err
	}

	mid, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return mid, nil
}
