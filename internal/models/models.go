package models

type Request struct {
	RequestURL string `json:"url"`
}

type Response struct {
	ResponseAddress string `json:"result"`
}
