package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"
	"strconv"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

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

	unsub := make(chan string, 10)
	newsub := make(chan string, 10)

	done := make(chan struct{})
	go handeleSubscription(c, rooms, &writeMu, done, unsub, newsub)
	go handleRoomChanges(c, uid, &writeMu, done, unsub, newsub)
	handleIncoming(c, uid, done)

}

func handleRoomChanges(c *websocket.Conn, uid int, writeMu *sync.Mutex, done chan struct{}, unsub, newsub chan string) {
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
				newsub <- strconv.Itoa(payload.ID)
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

			} else if payload.Type == RoomChangeLeave {
				unsub <- strconv.Itoa(payload.ID)

				// refactor to template later
				html := fmt.Sprintf(`<ul id="room-list" hx-swap-oob="delete:#room-%d"></ul>
				<div class="flex items-start justify-center space-x-2 mt-4" hx-swap-oob="outerHTML:#room-%d-messages">
  				<div>
    				<p class="bg-white p-2 rounded-lg shadow text-center">You have been kicked from this room</p>
  				</div>
				</div>
				<div hx-swap-oob='delete:#message-input:has(form[hx-vals$=":%d}"])'></div>`,
					payload.ID, payload.ID, payload.ID)

				writeMu.Lock()
				if err := c.WriteMessage(1, []byte(html)); err != nil {
					log.Fatalln("write error: ", err)
				}
				writeMu.Unlock()
			}
		}
	}

}

func handeleSubscription(c *websocket.Conn, rooms []string, writeMu *sync.Mutex, done chan struct{}, unsub, newsub chan string) {
	defer c.Close()

	rdb := config.RedisClient
	pubsub := rdb.Subscribe(context.Background(), rooms...)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-done:
			return
		case room := <-unsub:
			pubsub.Unsubscribe(context.Background(), room)
		case room := <-newsub:
			pubsub.Subscribe(context.Background(), room)
		case msg := <-ch:

			var payload services.PublishMessagePayload
			err := json.Unmarshal([]byte(msg.Payload), &payload)
			if err != nil {
				fmt.Println("error unmarshalling subscribed message", err)
				return
			}

			tmpl, err := template.ParseFiles("views/partials/message-with-room-reorder.html",
				"views/partials/message.html",
				"views/partials/message-right.html",
				"views/partials/message-left.html",
				"views/partials/message-middle.html")
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
		var msg services.WsIncomingMessage
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

		err := services.PersistPublishMessage(context.Background(), uid, msg)
		if err != nil {
			log.Println("error persisting message", err)
			continue
		}

	}
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

	var exists bool
	err := db.QueryRowContext(context.Background(), "select 1 from room_users where room_id = ? and user_id = ?", roomID, uid).Scan(&exists)
	if err != nil {
		return false
	}

	return true
}

func messageBelongsToRoom(messageID int, roomID int) bool {
	db := config.Db

	var exists bool
	err := db.QueryRowContext(context.Background(), "select 1 from messages where message_id = ? and room_id = ?", messageID, roomID).Scan(&exists)
	if err != nil {
		return false
	}

	return true
}
