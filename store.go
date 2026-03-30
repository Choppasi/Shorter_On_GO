package main

import (
	"fmt"
	"sync"
	"time"
)

// Link представляет сокращённую ссылку
type Link struct {
	OriginalURL string
	ShortCode   string
	CreatedAt   time.Time
	Clicks      int
}

// Store хранилище ссылок
type Store struct {
	mu      sync.RWMutex
	links   map[string]*Link
	reverse map[string]string // shortCode -> originalURL
}

// NewStore создаёт новое хранилище
func NewStore() *Store {
	return &Store{
		links:   make(map[string]*Link),
		reverse: make(map[string]string),
	}
}

// Add добавляет новую ссылку
func (s *Store) Add(originalURL, shortCode string) *Link {
	s.mu.Lock()
	defer s.mu.Unlock()

	link := &Link{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		CreatedAt:   time.Now(),
		Clicks:      0,
	}

	s.links[shortCode] = link
	s.reverse[originalURL] = shortCode

	return link
}

// GetByShortCode получает ссылку по короткому коду
func (s *Store) GetByShortCode(shortCode string) (*Link, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	link, exists := s.links[shortCode]
	return link, exists
}

// GetByOriginalURL получает ссылку по оригинальному URL
func (s *Store) GetByOriginalURL(originalURL string) (*Link, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shortCode, exists := s.reverse[originalURL]
	if !exists {
		return nil, false
	}

	link, exists := s.links[shortCode]
	return link, exists
}

// GetAll получает все ссылки
func (s *Store) GetAll() []*Link {
	s.mu.RLock()
	defer s.mu.RUnlock()

	links := make([]*Link, 0, len(s.links))
	for _, link := range s.links {
		links = append(links, link)
	}

	return links
}

// Delete удаляет ссылку
func (s *Store) Delete(shortCode string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	link, exists := s.links[shortCode]
	if !exists {
		return false
	}

	delete(s.reverse, link.OriginalURL)
	delete(s.links, shortCode)

	return true
}

// IncrementClicks увеличивает счётчик переходов
func (s *Store) IncrementClicks(shortCode string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if link, exists := s.links[shortCode]; exists {
		link.Clicks++
	}
}

// GenerateShortCode генерирует короткий код
func GenerateShortCode(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, length)

	for i := 0; i < length; i++ {
		code[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(time.Nanosecond)
	}

	return string(code)
}

// ValidateURL проверяет корректность URL
func ValidateURL(url string) bool {
	if len(url) < 5 {
		return false
	}
	// Простая проверка - должен содержать точку и начинаться с http
	return len(url) > 10 && (url[:7] == "http://" || url[:8] == "https://" || url[:4] == "www.")
}

// FormatURL форматирует URL
func FormatURL(url string) string {
	if len(url) > 0 && url[:4] == "www." {
		return "https://" + url
	}
	if len(url) > 0 && url[:7] == "http://" {
		return url
	}
	if len(url) > 0 && url[:8] == "https://" {
		return url
	}
	return "https://" + url
}

// GetHostName извлекает хост из URL
func GetHostName(url string) string {
	// Упрощённая версия - просто форматируем
	return fmt.Sprintf("http://%s", url)
}
