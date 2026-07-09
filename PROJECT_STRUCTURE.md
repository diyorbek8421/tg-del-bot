# Структура проекта

## Карта файлов

### 📁 Корневая директория

```
telegram-business-bot/
├── cmd/main/main.go              # 🚀 Точка входа приложения
├── config/config.go              # ⚙️ Загрузка конфигурации из .env
├── internal/
│   ├── domain/models.go          # 📊 Модели данных
│   ├── repository/
│   │   └── message.go            # 💾 Интерфейсы и реализация работы с БД
│   ├── service/
│   │   ├── message.go            # 🧠 Бизнес-логика
│   │   ├── message_test.go       # ✅ Юнит-тесты
│   │   └── mocks/
│   │       └── repository.go     # 🎭 Mock-объекты для тестирования
│   ├── storage/
│   │   ├── local.go              # 📁 Локальное хранилище файлов
│   │   └── downloader.go         # 🔽 Загрузка медиа из Telegram
│   └── delivery/telegram/
│       └── handler.go            # 🤖 Обработчики Telegram событий
├── docs/
│   ├── QUICKSTART.md             # ⚡ Быстрый старт за 5 минут
│   ├── API.md                    # 📚 АПФ документация
│   └── TELEGRAM_INTEGRATION.md   # 🔗 Интеграция с Telegram
├── examples/
│   ├── main.go                   # 💡 Примеры использования
│   └── media_download.go         # 📸 Примеры загрузки медиа
├── migrations/
│   └── schema.sql                # 🗄️ SQL миграции для БД
├── go.mod                        # 📦 Зависимости проекта
├── Dockerfile                    # 🐳 Docker образ
├── docker-compose.yml            # 🐪 Docker Compose конфигурация
├── Makefile                      # 🔨 Полезные команды
├── .env.example                  # 🔐 Пример переменных окружения
└── README.md                     # 📖 Основная документация
```

## Описание файлов

### `cmd/main/main.go` 🚀
**Точка входа приложения**

Что делает:
- Загружает конфигурацию
- Инициализирует БД PostgreSQL
- Запускает миграции
- Создает Telegram бота
- Получает обновления в бесконечном цикле

Как запустить:
```bash
go run cmd/main/main.go
```

### `config/config.go` ⚙️
**Управление конфигурацией**

Структуры:
- `Config` - основная конфигурация
- `DatabaseConfig` - параметры БД
- `ServerConfig` - параметры сервера
- `StorageConfig` - параметры хранилища

Функция:
```go
cfg, err := config.Load() // Загружает из .env
```

### `internal/domain/models.go` 📊
**Модели данных**

Структуры:
- `Message` - сообщение
- `MessageEdit` - редактирование
- `MessageDeletion` - удаление
- `BusinessUser` - пользователь бизнеса

### `internal/repository/message.go` 💾
**Слой доступа к данным**

Интерфейс `MessageRepository`:
- `SaveMessage()` - сохраняет сообщение
- `GetMessageByID()` - получает сообщение
- `UpdateMessage()` - обновляет сообщение
- `SaveMessageEdit()` - сохраняет редактирование
- `SaveMessageDeletion()` - сохраняет удаление
- `GetMessagesByChatID()` - получает историю
- `MarkMessageAsDeleted()` - помечает удаленным

Реализация использует `pgxpool` для работы с PostgreSQL.

### `internal/service/message.go` 🧠
**Бизнес-логика**

Интерфейс `MessageService`:
- `HandleNewBusinessMessage()` - обработка нового сообщения
- `HandleEditedBusinessMessage()` - отслеживание редактирования
- `HandleDeletedBusinessMessages()` - обработка удаления
- `GetMessageHistory()` - получение истории

### `internal/service/message_test.go` ✅
**Юнит-тесты сервиса**

Тестирует:
- Сохранение нового сообщения
- Обработку редактирования
- Обработку удаления

Запуск:
```bash
go test -v ./internal/service
```

### `internal/service/mocks/repository.go` 🎭
**Mock-объект для тестирования**

Используется в тестах вместо настоящей БД:
```go
mockRepo := mocks.NewMockMessageRepository()
service := NewMessageService(mockRepo)
```

### `internal/storage/local.go` 📁
**Локальное хранилище файлов**

Методы:
- `SaveFile()` - сохраняет файл на диск
- `GetFile()` - читает файл
- `DeleteFile()` - удаляет файл

### `internal/storage/downloader.go` 🔽
**Загрузка медиа из Telegram**

Методы:
- `DownloadPhoto()` - скачивает фото
- `DownloadVideo()` - скачивает видео
- `DownloadDocument()` - скачивает документ
- `DownloadAudio()` - скачивает аудио

### `internal/delivery/telegram/handler.go` 🤖
**Обработчики Telegram событий**

Методы:
- `HandleUpdates()` - маршрутизирует обновления
- `handleBusinessMessage()` - обработка новых сообщений
- `handleEditedBusinessMessage()` - обработка редактирования
- `handleDeletedBusinessMessages()` - обработка удаления
- `SendNotification()` - отправка уведомления пользователю

### `migrations/schema.sql` 🗄️
**SQL миграции**

Создает таблицы:
- `messages` - основные сообщения
- `message_edits` - история редактирования
- `message_deletions` - история удаления
- `business_users` - пользователи бизнеса

Автоматически выполняется при запуске `main.go`.

### `Dockerfile` 🐳
**Docker образ**

Создает контейнер с Go приложением:
```bash
docker build -t telegram-bot .
docker run telegram-bot
```

### `docker-compose.yml` 🐪
**Docker Compose**

Запускает два контейнера:
- `postgres` - БД PostgreSQL
- `bot` - приложение

```bash
docker-compose up -d
```

### `Makefile` 🔨
**Полезные команды**

```bash
make setup        # Инициализация проекта
make deps         # Загрузка зависимостей
make run          # Локальный запуск
make run-docker   # Запуск с Docker
make fmt          # Форматирование кода
make test         # Запуск тестов
make clean        # Очистка
```

### `.env.example` 🔐
**Пример конфигурации**

Переменные:
- `TELEGRAM_BOT_TOKEN` - токен бота
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - параметры БД
- `SERVER_PORT` - порт сервера
- `STORAGE_TYPE`, `LOCAL_STORAGE_PATH` - хранилище

Используется для создания `.env`:
```bash
cp .env.example .env
```

## Поток данных

```
Telegram API
    ↓
main.go (получает обновления)
    ↓
delivery/telegram/handler.go (парсит события)
    ↓
service/message.go (бизнес-логика)
    ↓
repository/message.go (записывает в БД)
    ↓
PostgreSQL (хранит данные)
```

## Архитектура (Clean Architecture)

```
┌─────────────────────────────────────┐
│ Delivery Layer (handler.go)         │
│ - Получает Telegram обновления      │
│ - Парсит события                    │
└──────────────┬──────────────────────┘
               ↓
┌─────────────────────────────────────┐
│ Service Layer (service/message.go)  │
│ - Бизнес-логика                     │
│ - Обработка событий                 │
└──────────────┬──────────────────────┘
               ↓
┌─────────────────────────────────────┐
│ Repository Layer (repository/)      │
│ - Интерфейсы для работы с данными   │
└──────────────┬──────────────────────┘
               ↓
┌─────────────────────────────────────┐
│ Domain Layer (domain/)              │
│ - Модели данных                     │
│ - Интерфейсы                        │
└─────────────────────────────────────┘
```

## Зависимости

```
go.mod
├── github.com/go-telegram-bot-api/telegram-bot-api/v5
│   └── Работа с Telegram API
├── github.com/jackc/pgx/v5
│   └── Драйвер PostgreSQL
└── github.com/joho/godotenv
    └── Загрузка .env файла
```

## Расширение проекта

### Добавить новый хендлер

1. Создайте функцию в `internal/delivery/telegram/handler.go`:
```go
func (h *TelegramHandler) handleMyEvent(ctx context.Context, event *tgbotapi.SomeEvent) error {
    // Логика обработки
}
```

2. Вызовите из `HandleUpdates()`:
```go
if update.MyEvent != nil {
    return h.handleMyEvent(ctx, update.MyEvent)
}
```

### Добавить новый сервис

1. Создайте интерфейс в `internal/service/`:
```go
type MyService interface {
    DoSomething(ctx context.Context) error
}
```

2. Реализуйте в struct:
```go
type myService struct {
    repo repository.MyRepository
}
```

### Добавить новый repository

1. Создайте интерфейс и реализацию в `internal/repository/`
2. Используйте в `main.go` при инициализации

## Сборка и развертывание

### Локально

```bash
go run cmd/main/main.go
```

### Бинарник

```bash
go build -o bot cmd/main/main.go
./bot
```

### Docker

```bash
docker build -t telegram-bot .
docker run -e TELEGRAM_BOT_TOKEN=... telegram-bot
```

### Docker Compose

```bash
docker-compose up -d
```

## Отладка

Включите логирование в `main.go`:
```go
log.SetFlags(log.LstdFlags | log.Llongfile)
log.SetOutput(os.Stdout)
```

Включите debug для Telegram API:
```go
botAPI.Debug = true
```

## Производительность

- Асинхронная обработка обновлений
- Connection pool для БД (pgxpool)
- Индексы на часто используемых полях
- Контексты с timeout для предотвращения зависаний

## Безопасность

- Переменные окружения для чувствительных данных
- SQL инъекции защищены параметризованными запросами
- Валидация входных данных
- Обработка ошибок Telegram API

---

Это базовая архитектура. Вы можете расширять проект добавляя новые фанилы, сервисы и обработчики! 🚀
