package storage

import (
	"database/sql"
	"errors"
	"github.com/AvdeevK/url-cutter.git/internal/models"
)

type PostgresStorage struct {
	db          *sql.DB
	storageName string
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{
		db:          db,
		storageName: "postgres storage",
	}
}

func (db *PostgresStorage) SaveURL(shortURL, originalURL string) error {
	_, err := db.db.Exec("INSERT INTO urls (short_url, original_url) VALUES ($1, $2)", shortURL, originalURL)
	return err
}

func (db *PostgresStorage) SaveBatchTransaction(tx *sql.Tx, shortURL string, originalURL string) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}

	query := "INSERT INTO urls (short_url, original_url) VALUES ($1, $2)"
	_, err := tx.Exec(query, shortURL, originalURL)
	return err
}

func (db *PostgresStorage) GetOriginalURL(shortURL string) (string, error) {
	var originalURL string
	err := db.db.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", shortURL).Scan(&originalURL)
	if err == sql.ErrNoRows {
		return "", errors.New("URL not found")
	}
	return originalURL, err
}

func (db *PostgresStorage) Ping() error {
	return db.db.Ping()
}

func (db *PostgresStorage) GetStorageName() (string, error) {
	return db.storageName, nil
}

func (db *PostgresStorage) SaveBatch(records []models.AddNewURLRecord) error {
	return errors.New("not implemented")
}
