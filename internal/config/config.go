package config

import (
	"flag"
	"os"
)

var Configs struct {
	RequestAddress  string
	ResponseAddress string
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
}
