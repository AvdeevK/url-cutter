package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/AvdeevK/url-cutter.git/internal/config"
	"github.com/AvdeevK/url-cutter.git/internal/logger"
	"github.com/AvdeevK/url-cutter.git/internal/models"
	"io"
	"os"
	"strconv"
)

type FileStorage struct {
	filePath    string
	urls        map[string][]string
	storageName string
}

var lastUUID int

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath:    filePath,
		urls:        models.PairsOfURLs,
		storageName: "file storage",
	}
	err := fs.LoadURLsFromFile()
	return fs, err
}

func (f *FileStorage) SaveURL(shortURL, originalURL, userID string) (string, error) {
	lastUUID += 1

	record := models.AddNewURLRecord{
		ID:          strconv.Itoa(lastUUID),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}

	return "", f.saveToFile(record)
}

func (f *FileStorage) GetOriginalURL(shortURL string) models.OriginalURLSelectionResult {
	attributes, exists := f.urls[shortURL]
	if !exists {
		return models.OriginalURLSelectionResult{"", false, errors.New("URL not found")}
	}
	originalURL := attributes[0]
	return models.OriginalURLSelectionResult{originalURL, false, nil}
}

func (f *FileStorage) Ping() error {
	return nil
}

func (f *FileStorage) LoadURLsFromFile() error {
	file, err := os.Open(config.Configs.FileStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	for {
		var record models.AddNewURLRecord
		if err := dec.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		attributes := &[]string{record.OriginalURL, record.UserID}
		f.urls[record.ShortURL] = *attributes
		lastUUID, err = strconv.Atoi(record.ID)
		if err != nil {
			logger.Log.Info("can't to get last uuid")
		}
	}

	return nil
}

func (f *FileStorage) saveToFile(newURL models.AddNewURLRecord) error {
	file, err := os.OpenFile(config.Configs.FileStoragePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	if err := enc.Encode(&newURL); err != nil {
		return err
	}

	attributes := &[]string{newURL.OriginalURL, newURL.UserID}
	f.urls[newURL.ShortURL] = *attributes
	return nil
}

func (f *FileStorage) GetStorageName() (string, error) {
	return f.storageName, nil
}

func (f *FileStorage) SaveBatch(records []models.AddNewURLRecord) error {
	for _, record := range records {
		if err := f.saveToFile(record); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStorage) SaveBatchTransaction(tx *sql.Tx, shortURL, originalURL, userID string) error {
	return errors.New("not implemented")
}

func (f *FileStorage) GetAllUserURLs(userID string) ([]models.BasePairsOfURLsResponse, error) {
	return nil, errors.New("not implemented")
}

func (f *FileStorage) MarkURLsAsDeleted(userID string, urlIDs []string) error {
	return errors.New("not implemented")
}
