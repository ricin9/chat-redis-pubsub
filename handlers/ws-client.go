package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"
	"strconv"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type WsClient struct {
	WsConn  *websocket.Conn
	Uid     int
	pubsub  *redis.PubSub
	writeMu sync.Mutex
}

func (c *WsClient) Init() {
	if c.WsConn == nil || c.Uid == 0 {
		log.Fatalln("ws connection or pubsub is nil")
		return
	}
	c.writeMu = sync.Mutex{}

	rooms := services.RoomIdsForUser(c.Uid)
	userNotifCh := fmt.Sprintf("user:%d", c.Uid)
	chans := append(rooms, userNotifCh)
	c.pubsub = config.RedisClient.Subscribe(context.Background(), chans...)
}

func (c *WsClient) Write(p []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.WsConn.WriteMessage(1, p)
}

func (c *WsClient) JoinRoom(data PSJoinRoom) error {
	err := c.pubsub.Subscribe(context.Background(), strconv.Itoa(data.RoomID))
	if err != nil {
		log.Println(err)
	}

	tmpl, err := template.ParseFiles("views/partials/room.html")
	if err != nil {
		fmt.Println("could not find template file", err)
		return err
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, fiber.Map{"ID": data.RoomID, "Name": data.Name})
	if err != nil {
		fmt.Println("error executing template", err)
		return err
	}

	result := output.String()

	result = `<ul id="room-list" hx-swap-oob="afterbegin">` + result + `</ul>`

	if err := c.Write([]byte(result)); err != nil {
		log.Fatalln("write error: ", err)
		return err
	}
	return nil
}

func (c *WsClient) KickedFromRoom(data PSKickedFromRoom) error {
	err := c.pubsub.Unsubscribe(context.Background(), strconv.Itoa(data.RoomID))
	if err != nil {
		log.Println(err)
	}

	tmpl, err := template.ParseFiles("views/partials/kicked-notif.html",
		"views/partials/message-middle.html")

	if err != nil {
		fmt.Println("could not find template file", err)
		return err
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, fiber.Map{
		"Content": "You have been kicked from this room",
		"RoomID":  data.RoomID,
	})

	if err != nil {
		fmt.Println("error executing template", err)
		return err
	}

	result := output.Bytes()

	if err := c.Write(result); err != nil {
		log.Fatalln("write error: ", err)
		return err
	}
	return nil
}

func (c *WsClient) LeaveRoom(data PSLeaveRoom) error {
	c.pubsub.Unsubscribe(context.Background(), strconv.Itoa(data.RoomID))
	return nil
}

func (c *WsClient) SendMessage(payload PSMessageBroadcast) error {
	tmpl, err := template.ParseFiles("views/partials/message-with-room-reorder.html",
		"views/partials/message.html",
		"views/partials/message-right.html",
		"views/partials/message-left.html",
		"views/partials/message-middle.html")
	if err != nil {
		fmt.Println("could not find template file", err)
		return err
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, fiber.Map{
		"ID":        payload.ID,
		"UserID":    payload.UserID,
		"RoomID":    payload.RoomID,
		"Username":  payload.Username,
		"Content":   payload.Content,
		"CreatedAt": payload.CreatedAt,
		"uid":       c.Uid,
	})

	if err != nil {
		fmt.Println("error executing template", err)
		return err
	}

	result := output.String()

	if err := c.Write([]byte(result)); err != nil {
		log.Fatalln("write error: ", err)
		return err
	}
	return nil
}
