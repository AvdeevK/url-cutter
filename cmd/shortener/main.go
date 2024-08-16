package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

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
	w.Write([]byte(fmt.Sprintf("http://localhost:8080/%s", shortURL)))
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
	r.MethodNotAllowed(notAllowedMethodsHandler)
	r.Route("/", func(r chi.Router) {
		r.Post("/", postURLHandler)
		r.Route("/{link}", func(r chi.Router) {
			r.Get("/", getURLHandler)
		})
	})
	return http.ListenAndServe(`:8080`, r)
}

func main() {
	r := chi.NewRouter()
	if err := run(r); err != nil {
		panic(err)
	}
}
