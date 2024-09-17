package config

import (
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3/v2"
	_ "github.com/mattn/go-sqlite3"
)

var (
	SessionStore *session.Store
)

func SetupStores() {
	SetupSessionStore()
	// going to use more stores later like limiter and more
}

func SetupSessionStore() {
	storage := sqlite3.New(sqlite3.Config{
		Database: SqliteSrc,
	})

	SessionStore = session.New(session.Config{
		Storage: storage,
	})
}
