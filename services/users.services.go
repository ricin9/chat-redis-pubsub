package services

import (
	"log"
	"ricin9/fiber-chat/config"
)

func GetUsername(uid int) string {
	db := config.Db

	var username string
	err := db.QueryRow("SELECT username FROM users WHERE user_id = ?", uid).Scan(&username)
	if err != nil {
		log.Println("error getting username", err)
		return "unknown"
	}

	return username
}
