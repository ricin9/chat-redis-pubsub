package config

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var (
	Db          *sql.DB
	SqliteSrc   string
	RedisClient *redis.Client
)

func SetupDatabasesConfig() {
	SetupSqlite()
	SetupRedis()
}

func SetupSqlite() {
	if *Prod {
		SqliteSrc = "prod.db"
	} else {
		SqliteSrc = "test.db"
	}

	db, err := sql.Open("sqlite3", SqliteSrc)
	if err != nil {
		log.Fatalf("SQLite3 database init has faield, src : %s", SqliteSrc)
	}

	Db = db
	data, err := db.Query("select 1")
	if err != nil {
		fmt.Println("sqlite3 query .tables error", err)
	}

	fmt.Println("data: ", data)

}

func SetupRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: ":6379",
	})
}
