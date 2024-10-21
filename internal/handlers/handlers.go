package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/AvdeevK/url-cutter.git/internal/config"
	"github.com/AvdeevK/url-cutter.git/internal/logger"
	"github.com/AvdeevK/url-cutter.git/internal/models"
	"github.com/AvdeevK/url-cutter.git/internal/storage"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
)

var (
	DB *sql.DB
)

var store storage.Storage

func InitializeStorage(s storage.Storage) {
	store = s
}

func CreateTable(db *sql.DB) error {
	// SQL-запрос для создания таблицы, если её нет
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_url VARCHAR(255) NOT NULL UNIQUE,
		original_url TEXT NOT NULL
	);`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}

	return nil
}

func createShortURL(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func NotAllowedMethodsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func PostJSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req models.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(req.RequestURL) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := createShortURL(8)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = store.SaveURL(shortURL, req.RequestURL)
	if err != nil {
		log.Printf("Error saving URL: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := models.Response{
		ResponseAddress: fmt.Sprintf("%s/%s", config.Configs.ResponseAddress, shortURL),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		logger.Log.Info(fmt.Sprintf("error encoding response: %s", err))
		return
	}
}

func PostURLHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	url := string(body)

	shortURL, err := createShortURL(8)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = store.SaveURL(shortURL, url)
	if err != nil {
		log.Printf("Error saving URL: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", config.Configs.ResponseAddress, shortURL)))
}

func GetURLHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		logger.Log.Info(fmt.Sprintf("incoming HTTP request isn't get"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := r.URL.Path[1:]
	if len(shortURL) == 0 {
		logger.Log.Info(fmt.Sprintf("requested url is empty"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, err := store.GetOriginalURL(shortURL)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("requested %s url, which isn't found", shortURL))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func PingDBHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := DB.Ping()
	if err != nil {
		logger.Log.Error("error of ping: ", zap.Error(err))
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
