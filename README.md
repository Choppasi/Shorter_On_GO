# 🔗 GoURL - URL Shortener

### README сделал ии

Простой и быстрый сокращатель ссылок на Go.

## 🚀 Возможности

- Сокращение длинных URL в короткие ссылки
- Перенаправление с подсчётом кликов
- Удаление ссылок
- Копирование в один клик
- Современный тёмный UI
- REST API

## 🛠 Установка

```bash
git clone https://github.com/YOUR_USERNAME/url-shortener.git
cd url-shortener
go run main.go
```

## 📖 Использование

1. Открой http://localhost:8080
2. Вставь длинную ссылку
3. Получи короткую ссылку

## 🌐 Доступ с других устройств

Сервер автоматически определяет локальный IP. Для использования с другими устройствами в сети:

```bash
# Узнать IP
ipconfig getifaddr en0
```

Или указать свой домен:

```bash
BASE_URL=https://example.com ./server
```

## 📡 API Endpoints

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/` | Главная страница |
| POST | `/api/shorten` | Создать ссылку |
| GET | `/r/{code}` | Перенаправление |
| GET | `/api/links` | Все ссылки |
| DELETE | `/api/delete/{code}` | Удалить ссылку |

## 🧰 Технологии

- Go 1.22+
- net/http
- In-Memory Storage
- HTML/CSS/JS

## 📝 Лицензия

MIT
