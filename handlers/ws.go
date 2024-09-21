package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"ricin9/fiber-chat/config"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type PublishMessagePayload struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	RoomID    int       `json:"room_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type IncomingMessage struct {
	Content string `json:"content"`
	RoomID  int    `json:"room_id"`
	ReplyTo int    `json:"reply_to"`
}

type RoomChangePayload struct {
	ID   int    `json:"room_id"`
	Name string `json:"name"`
	Type int    `json:"type"` // 1: join, 2: leave
}

var (
	RoomChangeJoin  = 1
	RoomChangeLeave = 2
)

func Websocket(c *websocket.Conn) {
	uid := c.Locals("uid").(int)
	rooms := userRooms(uid)

	writeMu := sync.Mutex{}

	done := make(chan struct{})
	go handeleSubscription(c, rooms, &writeMu, done)
	go handleRoomChanges(c, uid, &writeMu, done)
	handleIncoming(c, uid, done)

}

func handleRoomChanges(c *websocket.Conn, uid int, writeMu *sync.Mutex, done chan struct{}) {
	defer c.Close()

	rdb := config.RedisClient
	pubsub := rdb.Subscribe(context.Background(), "user:"+strconv.Itoa(int(uid)))
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-done:
			return
		case msg := <-ch:
			var payload RoomChangePayload
			err := json.Unmarshal([]byte(msg.Payload), &payload)
			if err != nil {
				log.Println("error unmarshalling room change message", err)
				continue
			}

			if payload.Type == RoomChangeJoin {

				tmpl, err := template.ParseFiles("views/partials/room.html")
				if err != nil {
					fmt.Println("could not find template file", err)
					break
				}

				var output bytes.Buffer
				err = tmpl.Execute(&output, payload)
				if err != nil {
					fmt.Println("error executing template", err)
					break
				}

				result := output.String()

				result = `<ul id="room-list" hx-swap-oob="afterbegin">` + result + `</ul>`

				writeMu.Lock()
				if err := c.WriteMessage(1, []byte(result)); err != nil {
					log.Fatalln("write error: ", err)
				}
				writeMu.Unlock()

			}
		}
	}

}

func handeleSubscription(c *websocket.Conn, rooms []string, writeMu *sync.Mutex, done chan struct{}) {
	defer c.Close()

	rdb := config.RedisClient
	pubsub := rdb.Subscribe(context.Background(), rooms...)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-done:
			return
		case msg := <-ch:

			var payload PublishMessagePayload
			err := json.Unmarshal([]byte(msg.Payload), &payload)
			if err != nil {
				fmt.Println("error unmarshalling subscribed message", err)
				return
			}

			tmpl, err := template.ParseFiles("views/partials/message-with-room-reorder.html",
				"views/partials/message.html",
				"views/partials/message-right.html",
				"views/partials/message-left.html")
			if err != nil {
				fmt.Println("could not find template file", err)
				break
			}

			var output bytes.Buffer
			err = tmpl.Execute(&output, fiber.Map{
				"ID":        payload.ID,
				"UserID":    payload.UserID,
				"RoomID":    payload.RoomID,
				"Username":  payload.Username,
				"Content":   payload.Content,
				"CreatedAt": payload.CreatedAt,
				"uid":       c.Locals("uid"),
			})

			fmt.Println("message user id", payload.UserID, "current user id", c.Locals("uid"))
			if err != nil {
				fmt.Println("error executing template", err)
				break
			}

			result := output.String()

			writeMu.Lock()
			if err := c.WriteMessage(1, []byte(result)); err != nil {
				log.Fatalln("write error: ", err)
			}
			writeMu.Unlock()
		}

	}
}

func handleIncoming(c *websocket.Conn, uid int, done chan struct{}) {
	defer c.Close()
	defer close(done)

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

		if msg.ReplyTo != 0 && !messageBelongsToRoom(msg.ReplyTo, msg.RoomID) {
			log.Println("message does not belong to room")
			continue
		}

		messageId, err := persistMessage(uid, msg)
		if err != nil {
			log.Println("error persisting message", err)
			continue
		}

		username := getUsername(uid)
		outgoing := PublishMessagePayload{
			ID:        messageId,
			UserID:    uid,
			RoomID:    msg.RoomID,
			Username:  username,
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

func getUsername(uid int) string {
	db := config.Db

	var username string
	err := db.QueryRow("SELECT username FROM users WHERE user_id = ?", uid).Scan(&username)
	if err != nil {
		log.Println("error getting username", err)
		return "unknown"
	}

	return username
}
func userRooms(uid int) []string {
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

func userBelongsTo(uid int, roomID int) bool {
	db := config.Db

	err := db.QueryRow("select 1 from room_users where room_id = ? and user_id = ?", roomID, uid).Err()
	if err != nil {
		return false
	}

	return true
}

func messageBelongsToRoom(messageID int, roomID int) bool {
	db := config.Db

	err := db.QueryRow("select 1 from messages where message_id = ? and room_id = ?", messageID, roomID).Err()
	if err != nil {
		return false
	}

	return true
}
func persistMessage(uid int, msg IncomingMessage) (messageId int, err error) {
	db := config.Db

	var replyto sql.NullInt64
	if msg.ReplyTo != 0 {
		replyto = sql.NullInt64{Int64: int64(msg.ReplyTo), Valid: true}
	}

	result, err := db.Exec("INSERT INTO messages (content, room_id, user_id, reply_to) VALUES (?, ?, ?, ?)", msg.Content, msg.RoomID, uid, replyto)
	if err != nil {
		return 0, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(messageID), nil
}
