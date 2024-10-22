package storage

import (
	"database/sql"
	"github.com/AvdeevK/url-cutter.git/internal/models"
)

type Storage interface {
	SaveURL(shortURL, originalURL string) error
	GetOriginalURL(shortURL string) (string, error)
	Ping() error
	SaveBatch([]models.AddNewURLRecord) error
	SaveBatchTransaction(*sql.Tx, string, string) error
	GetStorageName() (string, error)
}
