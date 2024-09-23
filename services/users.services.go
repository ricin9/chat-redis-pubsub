package services

import (
	"context"
	"log"
	"ricin9/fiber-chat/config"
)

func GetUsername(ctx context.Context, uid int) string {
	db := config.Db

	var username string
	err := db.QueryRowContext(ctx, "SELECT username FROM users WHERE user_id = ?", uid).Scan(&username)
	if err != nil {
		log.Println("error getting username", err)
		return "unknown"
	}

	return username
}
