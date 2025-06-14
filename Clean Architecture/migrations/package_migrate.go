package migrate

import (
	"database/sql"
	"log"

	"github.com/pressly/goose/v3"
)

func RunMigrations(db *sql.DB) {
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("Erro ao rodar migrations: %v", err)
	}
}
