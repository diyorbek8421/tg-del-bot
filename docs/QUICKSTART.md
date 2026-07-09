# Быстрый старт

## 5 минут до первого запуска

### Шаг 1: Создайте бота в Telegram

1. Откройте [@BotFather](https://t.me/botfather)
2. Отправьте `/start`
3. Отправьте `/newbot`
4. Следуйте инструкциям:
   - Придумайте имя: `My Business Bot`
   - Придумайте username: `my_business_bot` (с нижним подчеркиванием, без пробелов)
5. Скопируйте токен (выглядит как `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`)

### Шаг 2: Клонируйте проект

```bash
cd C:\Users\user\Desktop\Front\del
git clone <repository-url> telegram-business-bot
cd telegram-business-bot
```

### Шаг 3: Установите PostgreSQL

**Вариант 1: Локально на Windows**

- Скачайте с [postgresql.org](https://www.postgresql.org/download/windows/)
- Установите с дефолтными параметрами
- Пароль для `postgres`: запомните!

**Вариант 2: Docker (рекомендуется)**

```bash
docker run --name postgres-bot -e POSTGRES_PASSWORD=password -d -p 5432:5432 postgres:15-alpine
```

### Шаг 4: Настройте переменные окружения

```bash
cp .env.example .env
```

Отредактируйте `.env`:

```env
TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=telegram_bot
```

### Шаг 5: Запустите приложение

```bash
# Загрузите зависимости
go mod download
go mod tidy

# Запустите бота
go run cmd/main/main.go
```

Вы должны увидеть:

```
Authorized on account my_business_bot
Successfully connected to database
Migrations completed successfully
Bot started. Waiting for updates...
```

## Тестирование бота

### Локально (без Business API)

1. Напишите боту в личный чат
2. Бот должен ответить эхо

### С Business API

1. Создайте бизнес-аккаунт в Telegram
2. Перейдите: **Бизнес → Автоматизация → Боты**
3. Добавьте вашего бота
4. Отправьте сообщение через бизнес-чат
5. Проверьте логи:

```
Handling new business message from user 123 in chat 456
Successfully saved business message 1
```

## Структура БД

Бот автоматически создает таблицы при запуске:

```
messages
├── id (Primary Key)
├── chat_id
├── user_id
├── message_id
├── text
├── media_type (photo, video, document, audio)
├── media_file_id
├── media_path
└── created_at, updated_at, deleted_at

message_edits
├── id
├── message_id (Foreign Key)
├── old_text
├── new_text
└── edited_at

message_deletions
├── id
├── message_id (Foreign Key)
└── deleted_at

business_users
├── id
├── business_account_id
├── chat_id
└── username
```

## Полезные команды

### Makefile команды

```bash
# Запуск бота
make run

# Docker
make run-docker
make stop
make logs

# Проверка кода
make fmt
make lint

# Тесты
make test

# Очистка
make clean
```

### Прямые команды

```bash
# Загрузить зависимости
go mod download

# Форматировать код
go fmt ./...

# Запустить тесты
go test -v ./...

# Собрать бинарник
go build -o bot.exe cmd/main/main.go
```

## Отладка

### Включите debug режим

В `internal/delivery/telegram/handler.go`:

```go
botAPI.Debug = true
log.SetFlags(log.LstdFlags | log.Llongfile) // Показывает файл и строку
```

### Посмотрите логи БД

```bash
# PostgreSQL
psql -U postgres -d telegram_bot

# Проверьте данные
SELECT * FROM messages;
SELECT * FROM message_edits WHERE message_id = 1;
SELECT * FROM message_deletions WHERE message_id = 1;
```

### Проверьте соединение с Telegram

```go
me, err := bot.GetMe()
if err != nil {
    log.Fatal(err)
}
fmt.Println("Bot username:", me.UserName)
fmt.Println("Bot ID:", me.ID)
```

## Частые ошибки

### ❌ "TELEGRAM_BOT_TOKEN is required"

**Решение:** Проверьте `.env` файл, токен должен быть вроде `123456:ABC-DEF1234...`

### ❌ "failed to create database connection pool"

**Решение:** 
```bash
# Проверьте PostgreSQL запущен
psql -U postgres -c "SELECT version();"

# Создайте БД
createdb telegram_bot
```

### ❌ "connection refused"

**Решение:** 
- Docker: `docker ps` (должен быть postgres контейнер)
- Локально: PostgreSQL должен быть запущен (Services в Windows)

### ❌ "Bot token is invalid"

**Решение:** Скопируйте токен точно без пробелов

### ❌ "no rows in result set"

**Решение:** Сообщение не было сохранено. Проверьте:
- Бот подключен к Business API
- Сообщение отправлено через Business чат
- Логи содержат "Successfully saved business message"

## Следующие шаги

1. **Добавьте загрузку медиа:**
   - В [handler.go](internal/delivery/telegram/handler.go) вызовите downloader
   - Медиа будет сохраняться в папку `media/`

2. **Реализуйте отправку уведомлений:**
   - При удалении сообщения
   - При редактировании сообщения
   - При сохранении медиа

3. **Добавьте веб-интерфейс:**
   - REST API для просмотра истории
   - Dashboard для отображения логов
   - Экспорт в CSV/PDF

4. **Развертывание:**
   - Docker Compose для production
   - GitHub Actions для CI/CD
   - Nginx для reverse proxy

## Полезные ссылки

- [go-telegram-bot-api документация](https://pkg.go.dev/github.com/go-telegram-bot-api/telegram-bot-api/v5)
- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Business API](https://core.telegram.org/bots/business)
- [PostgreSQL документация](https://www.postgresql.org/docs/)
- [Go документация](https://golang.org/doc/)

## Поддержка

Если что-то не работает:

1. Проверьте логи (`make logs` для Docker)
2. Убедитесь, что все переменные в `.env` установлены
3. Перезагрузитесь: `docker-compose restart bot`
4. Читайте документацию в папке `docs/`

Happy coding! 🚀
