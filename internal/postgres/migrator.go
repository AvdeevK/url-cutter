package postgres

import (
	"database/sql"
	"github.com/pressly/goose/v3"
	"log"
)

func RunMigrations(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(db, migrationsDir); err != nil {
		return err
	}
	log.Println("Migrations applied successfully")
	return nil
}
