package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AvdeevK/url-cutter.git/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestPostURLHandler(t *testing.T) {
	// описываем ожидаемое тело ответа при успешном запросе

	requredPathOfResponseBody := config.Configs.ResponseAddress

	testCases := []struct {
		testName     string
		method       string
		expectedCode int
		requestBody  string
	}{
		{
			testName:     "Тест с обычным URL в запросе",
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
			requestBody:  "https://dzen.ru/",
		},
		{
			testName:     "Тест с пустым телом запроса",
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			requestBody:  "",
		},
		{
			testName:     "Тест с длинным  URL",
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
			requestBody: "https:/oAJIzeMUgAXdCOxIwlsqKqFrIiDtQDGoxyIw" +
				"FvtsuiuBTHkjXQtpkoANYiFbnYIoJUJOWOlxvUIY.ru/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.requestBody))
			w := httptest.NewRecorder()

			// вызовем хендлер как обычную функцию, без запуска самого сервера
			postURLHandler(w, r)

			// проверим корректность полученного тела ответа, если мы его ожидаем

			if tc.requestBody != "" {
				assert.Contains(t, w.Body.String(), requredPathOfResponseBody)
			}
			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestGetURLHandler(t *testing.T) {
	// описываем ожидаемое тело ответа при успешном запросе

	pairsOfURLs["/qMBUnCeI"] = "http://yandex.ru"
	pairsOfURLs["/hbflpNSd"] = "http://wLlvfmtuXUcjYopEUIpsmFORoKlQyINZQwucmqLKzLzJM" +
		"oAdIDWcMfAiJhDZZZlQbZWsolaiYEFUtQGZTBfvQGMZzbVaCWdOFLSZ.com"

	testCases := []struct {
		testName       string
		method         string
		expectedCode   int
		path           string
		headerLocation string
	}{
		{
			testName:       "Тест с существующим коротким URL в запросе",
			method:         http.MethodGet,
			expectedCode:   http.StatusTemporaryRedirect,
			path:           "/qMBUnCeI",
			headerLocation: pairsOfURLs["/qMBUnCeI"],
		},
		{
			testName:       "Тест с пустым телом запроса",
			method:         http.MethodGet,
			expectedCode:   http.StatusBadRequest,
			path:           "/",
			headerLocation: "",
		},
		{
			testName:       "Тест с коротким URL и длинным исходным URL",
			method:         http.MethodGet,
			expectedCode:   http.StatusTemporaryRedirect,
			path:           "/hbflpNSd",
			headerLocation: pairsOfURLs["/hbflpNSd"],
		},
		{
			testName:       "Тест с несуществующим коротким  URL",
			method:         http.MethodGet,
			expectedCode:   http.StatusBadRequest,
			path:           "/GNBlDGPP",
			headerLocation: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			// вызовем хендлер как обычную функцию, без запуска самого сервера
			getURLHandler(w, r)

			// проверим корректность полученного тела ответа, если мы его ожидаем

			if !assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не соответствует 307") {
				assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не соответствует 400")
			}

			assert.Equal(t, tc.headerLocation, w.Header().Get("Location"))
		})
	}
}
