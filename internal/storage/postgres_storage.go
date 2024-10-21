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

func (p *PostgresStorage) SaveURL(shortURL, originalURL string) error {
	_, err := p.db.Exec("INSERT INTO urls (short_url, original_url) VALUES ($1, $2)", shortURL, originalURL)
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

func (p *PostgresStorage) GetOriginalURL(shortURL string) (string, error) {
	var originalURL string
	err := p.db.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", shortURL).Scan(&originalURL)
	if err == sql.ErrNoRows {
		return "", errors.New("URL not found")
	}
	return originalURL, err
}

func (p *PostgresStorage) Ping() error {
	return p.db.Ping()
}

func (p *PostgresStorage) GetStorageName() (string, error) {
	return p.storageName, nil
}

func (p *PostgresStorage) SaveBatch(records []models.AddNewURLRecord) error {
	return errors.New("not implemented")
}
