# Telegram Business Bot

Telegram-бот для отслеживания и сохранения удаленных сообщений, отредактированных сообщений и исчезающих медиа-файлов в Telegram Business.

## Возможности

✅ Сохранение новых бизнес-сообщений в БД  
✅ Отслеживание отредактированных сообщений (история изменений)  
✅ Обработка удаленных сообщений  
✅ Сохранение медиа-файлов (фото, видео, документы, аудио)  
✅ Clean Architecture  
✅ PostgreSQL для хранения логов  

## Требования

- Go 1.22+
- PostgreSQL 12+
- Git

## Установка

### 1. Клонируйте репозиторий

```bash
git clone <repository-url>
cd telegram-business-bot
```

### 2. Установите зависимости

```bash
go mod download
go mod tidy
```

### 3. Настройте окружение

Создайте файл `.env` на основе `.env.example`:

```bash
cp .env.example .env
```

Отредактируйте `.env`:

```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=telegram_bot
SERVER_PORT=8080
```

### 4. Создайте БД PostgreSQL

```bash
createdb telegram_bot
```

## Запуск

### Локально

```bash
go run cmd/main/main.go
```

### С Docker Compose

```bash
docker-compose up -d
```

## Структура проекта

```
telegram-business-bot/
├── cmd/main/
│   └── main.go              # Точка входа приложения
├── config/
│   └── config.go            # Загрузка конфигурации
├── internal/
│   ├── domain/
│   │   └── models.go        # Модели данных (Message, MessageEdit и т.д.)
│   ├── repository/
│   │   └── message.go       # Repository для работы с БД
│   ├── service/
│   │   └── message.go       # Бизнес-логика
│   └── delivery/telegram/
│       └── handler.go       # Обработчики Telegram API
├── migrations/
│   └── schema.sql           # SQL миграции
├── media/                   # Локальное хранилище медиа
├── go.mod                   # Модули Go
├── go.sum                   # Контрольные суммы модулей
├── .env.example             # Пример переменных окружения
├── .gitignore               # Исключения Git
└── README.md                # Этот файл
```

## API Бизнес-Аккаунтов Telegram

Бот обрабатывает следующие обновления:

### 1. BusinessMessage (новое сообщение)
```go
update.BusinessMessage // Новое сообщение с медиа
```

### 2. EditedBusinessMessage (отредактированное сообщение)
```go
update.EditedBusinessMessage // Отредактированное сообщение
```

### 3. DeletedBusinessMessages (удаленные сообщения)
```go
update.DeletedBusinessMessages // Массив ID удаленных сообщений
```

## Настройка Бота в Telegram Business

1. Создайте бот через [@BotFather](https://t.me/botfather)
2. Активируйте режим Business через: **Настройки → Бизнес → Приложения и услуги → Боты**
3. Подключите вашего бота к бизнес-аккаунту
4. Вставьте токен в `.env`

## Функции

### SaveMessage
Сохраняет новое бизнес-сообщение в БД.

```go
message := &domain.Message{
    ChatID:    123456,
    UserID:    789012,
    MessageID: 1,
    Text:      "Hello",
}
err := messageService.HandleNewBusinessMessage(ctx, message)
```

### HandleEditedBusinessMessage
Сохраняет историю изменений сообщения.

```go
err := messageService.HandleEditedBusinessMessage(ctx, oldMsg, newMsg)
```

### HandleDeletedBusinessMessages
Обрабатывает удаленные сообщения и уведомляет пользователя.

```go
err := messageService.HandleDeletedBusinessMessages(ctx, messageID, userID)
```

## Расширение функционала

### Добавление поддержки S3

Замените локальное хранилище на AWS S3:

```go
// В config/config.go
StorageConfig struct {
    Type            string // "s3"
    S3Bucket        string
    S3Region        string
}
```

### Добавление уведомлений

Реализуйте отправку уведомлений при удалении/редактировании:

```go
err := handler.SendNotification(chatID, "Сообщение было отредактировано")
```

## Логирование

Логи выводятся в консоль. Для логирования в файл добавьте:

```go
logFile, _ := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
log.SetOutput(logFile)
```

## Развертывание

### На VPS (Linux)

```bash
# Установите Go
wget https://golang.org/dl/go1.22.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.22.linux-amd64.tar.gz

# Клонируйте репозиторий
git clone <repository-url> /opt/telegram-bot
cd /opt/telegram-bot

# Запустите как systemd сервис
sudo nano /etc/systemd/system/telegram-bot.service
```

Содержимое файла:
```ini
[Unit]
Description=Telegram Business Bot
After=network.target

[Service]
Type=simple
User=app
WorkingDirectory=/opt/telegram-bot
ExecStart=/usr/local/go/bin/go run cmd/main/main.go
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable telegram-bot
sudo systemctl start telegram-bot
```

## Решение проблем

### Ошибка подключения к БД

```
failed to create database connection pool: error
```

Проверьте:
- PostgreSQL запущен
- Учетные данные в `.env` правильные
- БД существует

### Бот не получает обновления

- Проверьте токен в `.env`
- Убедитесь, что бот добавлен в бизнес-аккаунт
- Проверьте логи: `docker-compose logs -f bot`

## Лицензия

MIT

## Контакты

По вопросам откройте Issue в репозитории.
