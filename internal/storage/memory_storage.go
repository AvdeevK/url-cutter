package storage

import (
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

func (m *MemoryStorage) GetOriginalURL(shortURL string) models.OriginalURLSelectionResult {
	attributes, exists := m.urls[shortURL]
	if !exists {
		return models.OriginalURLSelectionResult{
			OriginalURL: "",
			IsDeleted:   false,
			Error:       errors.New("URL not found"),
		}
	}
	originalURL := attributes[0]
	return models.OriginalURLSelectionResult{
		OriginalURL: originalURL,
		IsDeleted:   false,
		Error:       nil,
	}
}

func (m *MemoryStorage) Ping() error {
	return nil
}

func (m *MemoryStorage) GetStorageName() (string, error) {
	return m.storageName, nil
}

func (m *MemoryStorage) SaveBatch(records []models.AddNewURLRecord) error {
	for _, record := range records {
		m.urls[record.ShortURL] = []string{record.ShortURL, record.UserID}
	}
	return nil
}

func (m *MemoryStorage) GetAllUserURLs(userID string) ([]models.BasePairsOfURLsResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *MemoryStorage) MarkURLsAsDeleted(userID string, urlIDs []string) error {
	return errors.New("not implemented")
}
