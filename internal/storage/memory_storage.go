package storage

import (
	"errors"
	"github.com/AvdeevK/url-cutter.git/internal/models"
)

type MemoryStorage struct {
	urls        map[string]models.OriginalURLSelectionResult
	storageName string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls:        models.PairsOfURLs,
		storageName: "memory storage",
	}
}

func (m *MemoryStorage) SaveURL(shortURL, originalURL string, userID string) (string, error) {
	m.urls[shortURL] = models.OriginalURLSelectionResult{
		OriginalURL: originalURL,
		IsDeleted:   false,
		UserID:      userID,
	}
	return "", nil
}

func (m *MemoryStorage) GetOriginalURL(shortURL string) models.OriginalURLSelectionResult {
	attributes, exists := m.urls[shortURL]
	if !exists {
		return models.OriginalURLSelectionResult{
			OriginalURL: "",
			IsDeleted:   false,
			Error:       errors.New("URL not found"),
			UserID:      "",
		}
	}
	return models.OriginalURLSelectionResult{
		OriginalURL: attributes.OriginalURL,
		IsDeleted:   attributes.IsDeleted,
		Error:       attributes.Error,
		UserID:      attributes.UserID,
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
		m.urls[record.ShortURL] = models.OriginalURLSelectionResult{
			OriginalURL: record.OriginalURL,
			IsDeleted:   record.DeletedFlag,
			UserID:      record.UserID,
		}
	}
	return nil
}

func (m *MemoryStorage) GetAllUserURLs(userID string) ([]models.BasePairsOfURLsResponse, error) {
	result := make([]models.BasePairsOfURLsResponse, 0)
	for key, val := range m.urls {
		if val.UserID == userID && !val.IsDeleted {
			result = append(result, models.BasePairsOfURLsResponse{
				OriginalURL: val.OriginalURL,
				ShortURL:    key,
			})
		}
	}
	return result, nil
}

func (m *MemoryStorage) MarkURLsAsDeleted(userID string, urlIDs []string) error {
	for _, id := range urlIDs {
		if url, exists := m.urls[id]; exists {
			if url.UserID == userID {
				url.IsDeleted = true
				m.urls[id] = url
			}
		}
	}
	return nil
}
