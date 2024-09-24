package utils

import (
	"context"
	"log"
)

func GetUserId(ctx context.Context) int {
	if uid, ok := ctx.Value("uid").(int); ok {
		return uid
	}
	log.Fatalln("can't get user id from context in template")
	return 0
}

func GetUsername(ctx context.Context) string {
	if uid, ok := ctx.Value("username").(string); ok {
		return uid
	}
	log.Fatalln("can't get user id from context in template")
	return ""
}
