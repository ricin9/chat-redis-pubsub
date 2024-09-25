package handlers

import (
	"context"
	"encoding/json"
	"log"
	"ricin9/fiber-chat/services"

	"github.com/gofiber/contrib/websocket"
)

type PSBase struct {
	Type int `json:"type"`
}

type PSLeaveRoom struct {
	PSBase
	RoomID int `json:"room_id"`
}

type PSKickedFromRoom struct {
	PSBase
	RoomID int `json:"room_id"`
}

type PSJoinRoom struct {
	PSBase
	RoomID int    `json:"room_id"`
	Name   string `json:"name"`
}

type PSMessageBroadcast struct {
	PSBase
	services.PublishMessagePayload
}

const (
	CMessageBroadcast = iota + 1
	CJoinRoom
	CKickedFromRoom
	CLeaveRoom
)

func Websocket(c *websocket.Conn) {
	uid := c.Locals("uid").(int)
	done := make(chan struct{})

	client := &WsClient{WsConn: c, Uid: uid}
	client.Init()
	defer client.pubsub.Close()

	go handleSubscriptions(client, done)

	handleIncomingMessage(c, uid, done)
}

func handleSubscriptions(client *WsClient, done <-chan struct{}) {
	psCh := client.pubsub.Channel()

	for {
		select {
		case <-done:
			return
		case msg := <-psCh:
			var psmsg PSBase
			err := json.Unmarshal([]byte(msg.Payload), &psmsg)
			if err != nil {
				log.Println(err)
				continue
			}

			switch psmsg.Type {
			case CMessageBroadcast:
				var payload PSMessageBroadcast
				err := json.Unmarshal([]byte(msg.Payload), &payload)
				if err != nil {
					log.Println("error unmarshalling message", err)
					continue
				}
				client.SendMessage(payload)
			case CJoinRoom:
				var payload PSJoinRoom
				err := json.Unmarshal([]byte(msg.Payload), &payload)
				if err != nil {
					log.Println("error unmarshalling room join", err)
					continue
				}
				client.JoinRoom(payload)
			case CKickedFromRoom:
				var payload PSKickedFromRoom
				err := json.Unmarshal([]byte(msg.Payload), &payload)
				if err != nil {
					log.Println("error unmarshalling room join", err)
					continue
				}
				client.KickedFromRoom(payload)
			case CLeaveRoom:
				var payload PSLeaveRoom
				err := json.Unmarshal([]byte(msg.Payload), &payload)
				if err != nil {
					log.Println("error unmarshalling room join", err)
					continue
				}
				client.LeaveRoom(payload)

			}
		}
	}
}

func handleIncomingMessage(c *websocket.Conn, uid int, done chan struct{}) {
	defer c.Close()
	defer close(done)

	for {
		var msg services.WsIncomingMessage
		if err := c.ReadJSON(&msg); err != nil {
			// special ws goodbye message
			break
		}

		if !services.RoomHasUser(msg.RoomID, uid) {
			log.Println("user does not belong to room")
			continue
		}

		if msg.ReplyTo != 0 && !services.RoomHasMessage(msg.RoomID, msg.ReplyTo) {
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
