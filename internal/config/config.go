package config

import (
	"flag"
)

var Configs struct {
	RequestAddress  string
	ResponseAddress string
}

func ParseFlags() {

	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&Configs.RequestAddress, "a", ":8080", "server listening port")
	flag.StringVar(&Configs.ResponseAddress, "b", ":8080", "url availiable at port")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}
