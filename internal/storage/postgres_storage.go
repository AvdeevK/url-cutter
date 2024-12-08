package storage

import (
	"database/sql"
	"errors"
	"github.com/AvdeevK/url-cutter.git/internal/models"
	"github.com/lib/pq"
	"log"
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

func (db *PostgresStorage) SaveURL(shortURL, originalURL, userID string) (string, error) {
	query := `
        INSERT INTO urls (user_id, short_url, original_url)
        VALUES ($1, $2, $3)
        ON CONFLICT (original_url) DO NOTHING
        RETURNING short_url;
    `

	var existingShortURL string
	err := db.db.QueryRow(query, userID, shortURL, originalURL).Scan(&existingShortURL)

	if err != nil {
		if err == sql.ErrNoRows {
			return db.GetShortURLByOriginal(originalURL)
		}
		return "", err
	}
	return existingShortURL, nil
}

func (db *PostgresStorage) SaveBatchTransaction(
	tx *sql.Tx, shortURL string, originalURL string, userID string) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}

	query := "INSERT INTO urls (user_id, short_url, original_url) VALUES ($1, $2, $3)"
	_, err := tx.Exec(query, userID, shortURL, originalURL)
	return err
}

func (db *PostgresStorage) GetOriginalURL(shortURL string) models.OriginalURLSelectionResult {
	var (
		originalURL string
		isDeleted   bool
	)
	err := db.db.QueryRow("SELECT original_url, is_deleted FROM urls WHERE short_url = $1", shortURL).Scan(&originalURL, &isDeleted)
	if err == sql.ErrNoRows {
		return models.OriginalURLSelectionResult{
			OriginalURL: "",
			IsDeleted:   false,
			Error:       errors.New("not found"),
		}
	}
	return models.OriginalURLSelectionResult{
		OriginalURL: originalURL,
		IsDeleted:   isDeleted,
		Error:       nil,
	}
}

func (db *PostgresStorage) GetShortURLByOriginal(originalURL string) (string, error) {
	query := `
        SELECT short_url FROM urls WHERE original_url = $1;
    `
	var shortURL string
	err := db.db.QueryRow(query, originalURL).Scan(&shortURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return shortURL, errors.New("conflict")
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

func (db *PostgresStorage) GetAllUserURLs(userID string) ([]models.BasePairsOfURLsResponse, error) {
	query := `SELECT short_url, original_url FROM urls WHERE user_id = $1`
	rows, err := db.db.Query(query, userID)
	if err != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("Error closing rows: %v", cerr)
		}
	}()

	var records []models.BasePairsOfURLsResponse
	for rows.Next() {
		var record models.BasePairsOfURLsResponse
		if err := rows.Scan(&record.ShortURL, &record.OriginalURL); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func (db *PostgresStorage) MarkURLsAsDeleted(userID string, urlIDs []string) error {
	query := `
		UPDATE urls
		SET is_deleted = TRUE
		WHERE user_id = $1 AND short_url = ANY($2);
	`
	_, err := db.db.Exec(query, userID, pq.Array(urlIDs))
	return err
}
