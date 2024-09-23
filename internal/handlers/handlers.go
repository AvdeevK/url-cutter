package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/AvdeevK/url-cutter.git/internal/config"
	"github.com/AvdeevK/url-cutter.git/internal/logger"
	"github.com/AvdeevK/url-cutter.git/internal/models"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"strconv"
)

var PairsOfURLs = make(map[string]string)

type AddNewURLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var lastUUID int

func createShortURL(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func addURLToFile(newURL AddNewURLRecord) error {
	file, err := os.OpenFile(config.Configs.FileStoragePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	if err := enc.Encode(&newURL); err != nil {
		return err
	}

	return nil
}

func LoadURLsFromFile() error {
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
		var record AddNewURLRecord
		if err := dec.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		PairsOfURLs["/"+record.ShortURL] = record.OriginalURL
		lastUUID, err = strconv.Atoi(record.UUID)
		if err != nil {
			logger.Log.Info("can't to get las uuid")
		}
	}

	return nil
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

	PairsOfURLs["/"+shortURL] = req.RequestURL

	lastUUID += 1
	record := AddNewURLRecord{
		UUID:        strconv.Itoa(lastUUID),
		ShortURL:    shortURL,
		OriginalURL: req.RequestURL,
	}

	if err := addURLToFile(record); err != nil {
		logger.Log.Info("error appending URL to file", zap.Error(err))
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

	PairsOfURLs["/"+shortURL] = url

	lastUUID += 1
	record := AddNewURLRecord{
		UUID:        strconv.Itoa(lastUUID),
		ShortURL:    shortURL,
		OriginalURL: url,
	}

	if err := addURLToFile(record); err != nil {
		logger.Log.Info("error appending URL to file", zap.Error(err))
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", config.Configs.ResponseAddress, shortURL)))
}

func GetURLHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := r.URL.Path
	if len(shortURL) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, exists := PairsOfURLs[shortURL]

	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
