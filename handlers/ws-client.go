package handlers

import (
	"context"
	"fmt"
	"log"
	"ricin9/fiber-chat/config"
	"ricin9/fiber-chat/services"
	"ricin9/fiber-chat/views/partials"
	"strconv"
	"sync"

	"github.com/gofiber/contrib/websocket"
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

func (c *WsClient) Write(p []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return len(p), c.WsConn.WriteMessage(1, p)
}

func (c *WsClient) JoinRoom(data PSJoinRoom) error {
	err := c.pubsub.Subscribe(context.Background(), strconv.Itoa(data.RoomID))
	if err != nil {
		log.Println(err)
	}

	component := partials.JoinRoomOOB(services.Room{ID: data.RoomID, Name: data.Name})
	if err := component.Render(context.Background(), c); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (c *WsClient) KickedFromRoom(data PSKickedFromRoom) error {
	err := c.pubsub.Unsubscribe(context.Background(), strconv.Itoa(data.RoomID))
	if err != nil {
		log.Println(err)
	}

	component := partials.KickedNotificationOOB(data.RoomID)
	if err := component.Render(context.Background(), c); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (c *WsClient) LeaveRoom(data PSLeaveRoom) error {
	c.pubsub.Unsubscribe(context.Background(), strconv.Itoa(data.RoomID))
	return nil
}

func (c *WsClient) SendMessage(payload PSMessageBroadcast) error {

	component := partials.NewMessageOOB(payload.RoomID, services.Message{
		ID:        payload.ID,
		UserID:    payload.UserID,
		Username:  payload.Username,
		Content:   payload.Content,
		CreatedAt: payload.CreatedAt,
	})

	if err := component.Render(context.WithValue(context.Background(), "uid", c.Uid), c); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
