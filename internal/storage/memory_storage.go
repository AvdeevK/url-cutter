package storage

import (
	"database/sql"
	"errors"
	"github.com/AvdeevK/url-cutter.git/internal/models"
)

type MemoryStorage struct {
	urls        map[string][]string
	storageName string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls:        models.PairsOfURLs,
		storageName: "memory storage",
	}
}

func (m *MemoryStorage) SaveURL(shortURL, originalURL string, userID string) (string, error) {
	attributes := &[]string{originalURL, userID}
	m.urls[shortURL] = *attributes
	return "", nil
}

func (m *MemoryStorage) GetOriginalURL(shortURL string) (string, bool, error) {
	attributes, exists := m.urls[shortURL]
	if !exists {
		return "", false, errors.New("URL not found")
	}
	originalURL := attributes[0]
	return originalURL, false, nil
}

func (m *MemoryStorage) Ping() error {
	return nil
}

func (m *MemoryStorage) GetStorageName() (string, error) {
	return m.storageName, nil
}

func (m *MemoryStorage) SaveBatch(records []models.AddNewURLRecord) error {
	for _, record := range records {
		attributes := &[]string{record.ShortURL, record.UserID}
		m.urls[record.ShortURL] = *attributes
	}
	return nil
}

func (m *MemoryStorage) SaveBatchTransaction(tx *sql.Tx, shortURL, originalURL, userID string) error {
	return errors.New("not implemented")
}

func (m *MemoryStorage) GetAllUserURLs(userID string) ([]models.BasePairsOfURLsResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *MemoryStorage) MarkURLsAsDeleted(userID string, urlIDs []string) error {
	return errors.New("not implemented")
}
