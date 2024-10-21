package storage

type Storage interface {
	SaveURL(shortURL, originalURL string) error
	GetOriginalURL(shortURL string) (string, error)
	Ping() error
	GetStorageName() (string, error)
}
