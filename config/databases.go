package config

import (
	"database/sql"
	"log"
	"os"

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
	dbpath := os.Getenv("DATABASE_PATH")
	if len(dbpath) != 0 {
		SqliteSrc = dbpath
	} else {
		SqliteSrc = "test.db"
	}

	db, err := sql.Open("sqlite3", SqliteSrc)
	if err != nil {
		log.Fatalf("SQLite3 database init has faield, src : %s", SqliteSrc)
	}

	Db = db
}

func SetupRedis() {
	var opts *redis.Options

	rediscConnStr := os.Getenv("REDIS_CONN_STRING")
	if len(rediscConnStr) != 0 {
		parsed, err := redis.ParseURL(rediscConnStr)
		if err != nil {
			log.Fatalln("REDIS_CONN_STRING Redis connection string is invalid")
		}

		opts = parsed
	} else {
		opts = &redis.Options{
			Addr: ":6379",
		}
	}

	RedisClient = redis.NewClient(opts)
}
