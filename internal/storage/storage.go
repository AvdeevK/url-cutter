package storage

import (
	"github.com/AvdeevK/url-cutter.git/internal/models"
)

type Storage interface {
	SaveURL(shortURL, originalURL, userID string) (string, error)
	GetOriginalURL(shortURL string) models.OriginalURLSelectionResult
	Ping() error
	SaveBatch([]models.AddNewURLRecord) error
	GetStorageName() (string, error)
	GetAllUserURLs(string) ([]models.BasePairsOfURLsResponse, error)
	MarkURLsAsDeleted(string, []string) error
}
