package main

import (
	"encoding/json"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// Templates кэш шаблонов
var templates = template.Must(template.ParseFiles("templates/index.html"))

// Store глобальное хранилище
var store = NewStore()

// BaseURL - базовый URL для коротких ссылок
var BaseURL = getBaseURL()

func getBaseURL() string {
	// Проверяем переменную окружения
	if envURL := os.Getenv("BASE_URL"); envURL != "" {
		return envURL
	}

	// Получаем локальный IP
	ip := getLocalIP()
	if ip != "" {
		return "http://" + ip + ":8080"
	}

	return "http://localhost:8080"
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Статические файлы
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Маршруты
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/shorten", handleShorten)
	http.HandleFunc("/api/links", handleLinks)
	http.HandleFunc("/api/delete", handleDelete)

	// Перенаправление по коротким ссылкам (должен быть в конце)
	http.HandleFunc("/r/", handleRedirect)

	log.Printf("🚀 Сервер запущен на %s", BaseURL)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// LinkResponse ответ API для ссылки
type LinkResponse struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	ShortCode   string `json:"short_code"`
	Clicks      int    `json:"clicks"`
	CreatedAt   string `json:"created_at"`
}

// handleIndex главная страница
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	links := store.GetAll()
	linkResponses := make([]LinkResponse, len(links))

	for i, link := range links {
		linkResponses[i] = LinkResponse{
			OriginalURL: link.OriginalURL,
			ShortURL:    BaseURL + "/r/" + link.ShortCode,
			ShortCode:   link.ShortCode,
			Clicks:      link.Clicks,
			CreatedAt:   link.CreatedAt.Format("02.01.2006 15:04"),
		}
	}

	data := struct {
		Links []LinkResponse
	}{
		Links: linkResponses,
	}

	templates.ExecuteTemplate(w, "index.html", data)
}

// handleShorten создание короткой ссылки
func handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	originalURL := strings.TrimSpace(r.Form.Get("url"))

	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Форматируем URL
	originalURL = FormatURL(originalURL)

	if !ValidateURL(originalURL) {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли уже такая ссылка
	if existingLink, exists := store.GetByOriginalURL(originalURL); exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LinkResponse{
			OriginalURL: existingLink.OriginalURL,
			ShortURL:    BaseURL + "/r/" + existingLink.ShortCode,
			ShortCode:   existingLink.ShortCode,
			Clicks:      existingLink.Clicks,
			CreatedAt:   existingLink.CreatedAt.Format("02.01.2006 15:04"),
		})
		return
	}

	// Генерируем уникальный код
	shortCode := generateUniqueCode()

	// Добавляем в хранилище
	link := store.Add(originalURL, shortCode)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LinkResponse{
		OriginalURL: link.OriginalURL,
		ShortURL:    BaseURL + "/r/" + link.ShortCode,
		ShortCode:   link.ShortCode,
		Clicks:      link.Clicks,
		CreatedAt:   link.CreatedAt.Format("02.01.2006 15:04"),
	})
}

// handleLinks получение всех ссылок
func handleLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	links := store.GetAll()
	linkResponses := make([]LinkResponse, len(links))

	for i, link := range links {
		linkResponses[i] = LinkResponse{
			OriginalURL: link.OriginalURL,
			ShortURL:    BaseURL + "/r/" + link.ShortCode,
			ShortCode:   link.ShortCode,
			Clicks:      link.Clicks,
			CreatedAt:   link.CreatedAt.Format("02.01.2006 15:04"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(linkResponses)
}

// handleDelete удаление ссылки
func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := strings.TrimPrefix(r.URL.Path, "/api/delete/")
	shortCode = strings.TrimPrefix(shortCode, "/")

	if shortCode == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	success := store.Delete(shortCode)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": success})
}

// handleRedirect перенаправление по короткой ссылке
func handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/r/")
	shortCode = strings.TrimPrefix(shortCode, "/")

	if shortCode == "" {
		http.NotFound(w, r)
		return
	}

	link, exists := store.GetByShortCode(shortCode)
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Увеличиваем счётчик переходов
	store.IncrementClicks(shortCode)

	// Перенаправляем на оригинальный URL
	http.Redirect(w, r, link.OriginalURL, http.StatusMovedPermanently)
}

// generateUniqueCode генерирует уникальный короткий код
func generateUniqueCode() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 6

	for attempts := 0; attempts < 100; attempts++ {
		code := make([]byte, codeLength)
		for i := 0; i < codeLength; i++ {
			code[i] = chars[rand.Intn(len(chars))]
		}

		shortCode := string(code)
		if _, exists := store.GetByShortCode(shortCode); !exists {
			return shortCode
		}
	}

	// Если не удалось найти уникальный код за 100 попыток, увеличиваем длину
	code := make([]byte, 8)
	for i := 0; i < 8; i++ {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}
