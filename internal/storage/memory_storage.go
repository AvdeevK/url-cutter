package storage

import (
	"errors"
	"github.com/AvdeevK/url-cutter.git/internal/models"
)

type MemoryStorage struct {
	urls        map[string]string
	storageName string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls:        models.PairsOfURLs,
		storageName: "memory storage",
	}
}

func (m *MemoryStorage) SaveURL(shortURL, originalURL string) error {
	m.urls[shortURL] = originalURL
	return nil
}

func (m *MemoryStorage) GetOriginalURL(shortURL string) (string, error) {
	originalURL, exists := m.urls[shortURL]
	if !exists {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

func (m *MemoryStorage) Ping() error {
	return nil
}

func (m *MemoryStorage) GetStorageName() (string, error) {
	return m.storageName, nil
}