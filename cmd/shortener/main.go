package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

var pairsOrURLs = make(map[string]string)

func createShortURL(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
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

	pairsOrURLs[shortURL] = url

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://localhost:8080/%s", shortURL)))
}

func getURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := r.URL.Path[1:]
	if len(shortURL) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, exists := pairsOrURLs[shortURL]

	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(originalURL))
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == "/" {
		postURLHandler(w, r)
	} else if r.Method == http.MethodGet && r.URL.Path != "/" {
		getURLHandler(w, r)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func run() error {
	return http.ListenAndServe(`:8080`, http.HandlerFunc(mainHandler))
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
