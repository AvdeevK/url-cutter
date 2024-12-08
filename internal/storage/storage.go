package storage

import (
	"database/sql"
	"github.com/AvdeevK/url-cutter.git/internal/models"
)

type Storage interface {
	SaveURL(shortURL, originalURL, userID string) (string, error)
	GetOriginalURL(shortURL string) models.OriginalURLSelectionResult
	Ping() error
	SaveBatch([]models.AddNewURLRecord) error
	SaveBatchTransaction(*sql.Tx, string, string, string) error
	GetStorageName() (string, error)
	GetAllUserURLs(string) ([]models.BasePairsOfURLsResponse, error)
	MarkURLsAsDeleted(string, []string) error
}
