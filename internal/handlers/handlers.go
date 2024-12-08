package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AvdeevK/url-cutter.git/internal/auth"
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

func generateShortURL(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func ValidateAndSetAuthCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	userID, exists, err := auth.GetAuthCookie(r)
	if !exists || err != nil {
		logger.Log.Warn(fmt.Sprintf("error getting auth cookie or: %v", err))
		logger.Log.Info("start process of creating cookie")
		newUserID, err := auth.GenerateUserID()
		if err != nil {
			logger.Log.Info("error of generating user id")
			return "", errors.New("unable to generate user id")
		}
		userID = newUserID
	}
	return userID, nil
}

func NotAllowedMethodsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func PostURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Получение userID и проверка на ошибку
	logger.Log.Info("start processing of creating cookie")
	userID, err := ValidateAndSetAuthCookie(w, r)
	logger.Log.Info("finish processing of creating cookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if userID == "" {
		logger.Log.Error("got empty user id in cookie, skip processing")
		http.Error(w, "empty user id", http.StatusUnauthorized)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	url := string(body)

	shortURL, err := generateShortURL(8)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	existingShortURL, err := store.SaveURL(shortURL, url, userID)
	if err != nil {
		if err.Error() == "conflict" {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fmt.Sprintf("%s/%s", config.Configs.ResponseAddress, existingShortURL)))
			return
		}
		log.Printf("Error saving URL: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = auth.SetAuthCookie(w, userID)
	if err != nil {
		http.Error(w, "unable to set cookie", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", config.Configs.ResponseAddress, shortURL)))
}

func PostJSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Получение userID и проверка на ошибку
	logger.Log.Info("get auth cookie from request")
	userID, err := ValidateAndSetAuthCookie(w, r)
	logger.Log.Info("finish getting auth cookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if userID == "" {
		logger.Log.Error("got empty user id in cookie, skip processing")
		http.Error(w, "empty user id", http.StatusUnauthorized)
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

	shortURL, err := generateShortURL(8)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	existingShortURL, err := store.SaveURL(shortURL, req.RequestURL, userID)
	if err != nil {
		if err.Error() == "conflict" {
			resp := models.Response{
				ResponseAddress: fmt.Sprintf("%s/%s", config.Configs.ResponseAddress, existingShortURL),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)

			enc := json.NewEncoder(w)
			if err := enc.Encode(resp); err != nil {
				logger.Log.Info(fmt.Sprintf("error encoding response: %s", err))
				return
			}
			return
		}
		log.Printf("Error saving URL: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := models.Response{
		ResponseAddress: fmt.Sprintf("%s/%s", config.Configs.ResponseAddress, shortURL),
	}

	err = auth.SetAuthCookie(w, userID)
	if err != nil {
		http.Error(w, "unable to set cookie", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		logger.Log.Info(fmt.Sprintf("error encoding response: %s", err))
		return
	}
}

func GetURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		logger.Log.Info("incoming HTTP request isn't get")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Получение userID и проверка на ошибку
	logger.Log.Info("get auth cookie from request")
	userID, err := ValidateAndSetAuthCookie(w, r)
	logger.Log.Info("finish getting auth cookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if userID == "" {
		logger.Log.Error("got empty user id in cookie, skip processing")
		http.Error(w, "empty user id", http.StatusUnauthorized)
	}

	shortURL := r.URL.Path[1:]
	if len(shortURL) == 0 {
		logger.Log.Info("requested url is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, isDeleted, err := store.GetOriginalURL(shortURL)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("requested %s url, which isn't found", shortURL))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isDeleted {
		w.WriteHeader(http.StatusGone)
		return
	}

	err = auth.SetAuthCookie(w, userID)
	if err != nil {
		http.Error(w, "unable to set cookie", http.StatusInternalServerError)
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

func PostBatchURLHandler(w http.ResponseWriter, r *http.Request) {
	var records []models.AddNewURLRecord
	var responses []models.BatchResponse

	//Получение userID и проверка на ошибку
	logger.Log.Info("get auth cookie from request")
	userID, err := ValidateAndSetAuthCookie(w, r)
	logger.Log.Info("finish getting auth cookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if userID == "" {
		logger.Log.Error("got empty user id in cookie, skip processing")
		http.Error(w, "empty user id", http.StatusUnauthorized)
	}

	if err := json.NewDecoder(r.Body).Decode(&records); err != nil {
		logger.Log.Error("Error decoding request body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(records) == 0 {
		logger.Log.Warn("Received empty batch")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	storageType, _ := store.GetStorageName()

	if storageType == "postgres storage" {
		tx, err := DB.Begin()
		if err != nil {
			logger.Log.Error("Error starting transaction: ", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		for _, record := range records {
			if record.OriginalURL == "" {
				logger.Log.Warn("Original URL is empty for correlation ID")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			shortURL, err := generateShortURL(8)
			if err != nil {
				logger.Log.Error("Error creating short URL: ", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if err := store.SaveBatchTransaction(tx, shortURL, record.OriginalURL, userID); err != nil {
				logger.Log.Error("Error saving URL in transaction: ", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			responses = append(responses, models.BatchResponse{
				CorrelationID: record.ID,
				ShortURL:      fmt.Sprintf("%s/%s", config.Configs.ResponseAddress, shortURL),
			})
		}

		if err := tx.Commit(); err != nil {
			logger.Log.Error("Error committing transaction: ", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		for _, record := range records {
			if record.OriginalURL == "" {
				logger.Log.Warn("Original URL is empty for correlation ID")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			shortURL, err := generateShortURL(8)
			if err != nil {
				logger.Log.Error("Error creating short URL: ", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if _, err := store.SaveURL(shortURL, record.OriginalURL, userID); err != nil {
				logger.Log.Error("Error saving URL: ", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			responses = append(responses, models.BatchResponse{
				CorrelationID: record.ID,
				ShortURL:      fmt.Sprintf("%s/%s", config.Configs.ResponseAddress, shortURL),
			})
		}
	}

	err = auth.SetAuthCookie(w, userID)
	if err != nil {
		http.Error(w, "unable to set cookie", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(responses); err != nil {
		logger.Log.Error("Error encoding response: ", zap.Error(err))
	}
}

func GetAllUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		logger.Log.Info("incoming HTTP request isn't get")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Получение userID и проверка на ошибку
	logger.Log.Info("get auth cookie from request")
	userID, err := ValidateAndSetAuthCookie(w, r)
	logger.Log.Info("finish getting auth cookie")
	if err != nil {
		logger.Log.Info(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if userID == "" {
		logger.Log.Error("got empty user id in cookie, skip processing")
		http.Error(w, "empty user id", http.StatusUnauthorized)
	}

	records, err := store.GetAllUserURLs(userID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for i := range records {
		records[i].ShortURL = fmt.Sprintf("%s/%s", config.Configs.ResponseAddress, records[i].ShortURL)
		logger.Log.Info("finished processing of creating URL: ", zap.Any("record", records[i]))
	}

	// Если записей нет, возвращаем 204 No Content, но в автотестах ошибка, непонятно, почему тут 401.
	if len(records) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Отправляем записи в формате JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(records)
}

func DeleteUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//Получение userID и проверка на ошибку
	logger.Log.Info("get auth cookie from request")
	userID, err := ValidateAndSetAuthCookie(w, r)
	logger.Log.Info("finish getting auth cookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if userID == "" {
		logger.Log.Error("got empty user id in cookie, skip processing")
		http.Error(w, "empty user id", http.StatusUnauthorized)
	}

	var urlIDs []string
	if err := json.NewDecoder(r.Body).Decode(&urlIDs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(urlIDs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	go func(w http.ResponseWriter) {
		if err := store.MarkURLsAsDeleted(userID, urlIDs); err != nil {
			logger.Log.Error("Failed to mark URLs as deleted", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}(w)

	w.WriteHeader(http.StatusAccepted)
}
