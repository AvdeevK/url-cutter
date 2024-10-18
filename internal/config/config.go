package config

import (
	"flag"
	"os"
)

var Configs struct {
	RequestAddress  string
	ResponseAddress string
	DatabaseAddress string
	FileStoragePath string
}

func ParseFlags() {

	flag.StringVar(&Configs.RequestAddress, "a", "localhost:8080", "server listening port")
	flag.StringVar(&Configs.ResponseAddress, "b", "http://localhost:8080", "url availiable at port")

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		Configs.RequestAddress = envRunAddr
	}

	if envReqAddr := os.Getenv("BASE_URL"); envReqAddr != "" {
		Configs.ResponseAddress = envReqAddr
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		Configs.FileStoragePath = envFileStoragePath
	} else {
		flag.StringVar(&Configs.FileStoragePath, "f", "backup_url.json", "file storage path")
		flag.Parse()
	}

	if envDatabaseAddress := os.Getenv("DATABASE_DSN"); envDatabaseAddress != "" {
		Configs.DatabaseAddress = envDatabaseAddress
	} else {
		flag.StringVar(&Configs.DatabaseAddress, "d", "localhost:5432", "database availiable at port")
		flag.Parse()
	}
}
