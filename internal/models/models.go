package models

type Request struct {
	RequestURL string `json:"url"`
}

type Response struct {
	ResponseAddress string `json:"result"`
}

type AddNewURLRecord struct {
	ID          string `json:"correlation_id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

var PairsOfURLs = make(map[string]string)
