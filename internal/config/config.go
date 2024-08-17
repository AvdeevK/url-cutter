package config

import (
	"flag"
)

var Configs struct {
	RequestAddress  string
	ResponseAddress string
}

func ParseFlags() {
	flag.StringVar(&Configs.RequestAddress, "a", "localhost:8080", "server host")
	flag.StringVar(&Configs.ResponseAddress, "b", "http://localhost:8080", "short url awailiable at host")
	flag.Parse()
}
