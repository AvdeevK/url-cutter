package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/AvdeevK/url-cutter.git/internal/logger"
	"github.com/AvdeevK/url-cutter.git/internal/models"
	"go.uber.org/zap"
	"io"
	"net/http"

	"github.com/AvdeevK/url-cutter.git/internal/config"
	"github.com/go-chi/chi/v5"
)

var pairsOfURLs = make(map[string]string)

func createShortURL(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func notAllowedMethodsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func postJSONHandler(w http.ResponseWriter, r *http.Request) {
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

	pairsOfURLs["/"+shortURL] = req.RequestURL

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

func postURLHandler(w http.ResponseWriter, r *http.Request) {

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

	pairsOfURLs["/"+shortURL] = url

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", config.Configs.ResponseAddress, shortURL)))
}

func getURLHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := r.URL.Path
	if len(shortURL) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, exists := pairsOfURLs[shortURL]

	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func run(r chi.Router) error {
	if err := logger.Initialize("Info"); err != nil {
		return err
	}
	logger.Log.Info("Running server", zap.String("address", config.Configs.RequestAddress))

	r.MethodNotAllowed(logger.RequestLogger(logger.ResponseLogger(notAllowedMethodsHandler)))
	r.Post("/", logger.RequestLogger(logger.ResponseLogger(postURLHandler)))
	r.Get("/{link}", logger.RequestLogger(logger.ResponseLogger(getURLHandler)))
	r.Post("/api/shorten", logger.RequestLogger(logger.ResponseLogger(postJSONHandler)))
	return http.ListenAndServe(config.Configs.RequestAddress, r)
}

func main() {
	r := chi.NewRouter()

	config.ParseFlags()
	if err := run(r); err != nil {
		panic(err)
	}
}
