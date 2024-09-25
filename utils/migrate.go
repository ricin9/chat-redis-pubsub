package utils

import (
	"embed"
	"log"
	"ricin9/fiber-chat/config"

	migrate "github.com/rubenv/sql-migrate"
)

func Migrate(migrations embed.FS) {

	// run migrations
	migrationSrc := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "migrations",
	}

	n, err := migrate.Exec(config.Db, "sqlite3", migrationSrc, migrate.Up)
	if err != nil {
		log.Fatalln("[MIGRATIONS] Error running migrations: ", err)
	}
	log.Println("[MIGRATIONS] Applied ", n, " migrations")
}
