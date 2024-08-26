package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

type stringMap = map[string]string

var urls stringMap
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func main() {
	urls = stringMap{}
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", webhook)
	return http.ListenAndServe(`:8080`, mux)
}

func webhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		requestBody := string(body)
		_, err = url.ParseRequestURI(requestBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Ошибка при чтении URL: %v", err)))
		} else {
			result := addUrl(requestBody)
			baseUrl := getBaseUrl(r)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(baseUrl + "/" + result))
		}
	case http.MethodGet:
		id := extractIdFromUrl(r)
		result, err := readUrl(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Ошибка при поиске URL: %v", err)))
		} else {
			w.Header().Add("Location", result)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	default:
		http.Error(w, "Неправильный метод запроса", http.StatusMethodNotAllowed)
	}
}

func addUrl(requestBody string) string {
	var shortUrl string
	for {
		shortUrl = generateUrl(8)
		checkUrl := urls[requestBody]
		if checkUrl != "" {
			continue
		}
		break
	}
	urls[shortUrl] = requestBody
	return shortUrl
}

func readUrl(requestBody string) (string, error) {
	if len(urls) == 0 {
		return "", errors.New("Пока нет закладок, добавьте корректный URL")

	}

	shortUrl := urls[requestBody]
	if shortUrl == "" {
		return "", errors.New("Такого URL не найдено")
	}

	return shortUrl, nil

}

func generateUrl(n int) string {
	arr := make([]rune, n)
	for i := range arr {
		arr[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(arr)
}

func extractIdFromUrl(r *http.Request) string {
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/") // Удаляем начальный слеш, если он есть
	return id
}

func getBaseUrl(r *http.Request) string {
	return "http://" + r.Host
}
