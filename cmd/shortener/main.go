package main

import (
	"database/sql"
	"fmt"
	"github.com/AvdeevK/url-cutter.git/internal/handlers"
	"github.com/AvdeevK/url-cutter.git/internal/logger"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strings"

	"github.com/AvdeevK/url-cutter.git/internal/config"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func gzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		h.ServeHTTP(ow, r)
	}
}

func run(r chi.Router) error {
	if err := logger.Initialize("Info"); err != nil {
		return err
	}
	logger.Log.Info("Running server", zap.String("address", config.Configs.RequestAddress))
	logger.Log.Info("Connection to DB with", zap.String("address", config.Configs.DatabaseAddress))

	if err := handlers.LoadURLsFromFile(); err != nil {
		logger.Log.Error("error loading URLs from file", zap.Error(err))
	}

	r.MethodNotAllowed(logger.RequestLogger(logger.ResponseLogger(gzipMiddleware(handlers.NotAllowedMethodsHandler))))
	r.Post("/", logger.RequestLogger(logger.ResponseLogger(gzipMiddleware(handlers.PostURLHandler))))
	r.Get("/ping", logger.RequestLogger(logger.ResponseLogger(gzipMiddleware(handlers.PingDBHandler))))
	r.Get("/{link}", logger.RequestLogger(logger.ResponseLogger(gzipMiddleware(handlers.GetURLHandler))))
	r.Post("/api/shorten", logger.RequestLogger(logger.ResponseLogger(gzipMiddleware(handlers.PostJSONHandler))))
	return http.ListenAndServe(config.Configs.RequestAddress, r)
}

func main() {
	r := chi.NewRouter()
	var err error

	config.ParseFlags()

	handlers.DB, err = sql.Open("pgx", fmt.Sprintf("host=%s", config.Configs.DatabaseAddress))
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer handlers.DB.Close()

	if err := run(r); err != nil {
		panic(err)
	}
}
