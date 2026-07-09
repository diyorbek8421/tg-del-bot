# Интеграция Telegram Business API

## Поддерживаемые события

### 1. BusinessMessage

Срабатывает при получении нового сообщения в бизнес-аккаунте.

```go
if update.BusinessMessage != nil {
    msg := update.BusinessMessage
    // msg.MessageID - ID сообщения
    // msg.Text - текст сообщения
    // msg.From - информация об отправителе
    // msg.Photo - массив фото
    // msg.Video - видео
    // msg.Document - документ
}
```

**Структура сообщения:**

```
tgbotapi.Message:
├── MessageID: int
├── From: &User
├── Chat: &Chat
├── Text: string
├── Photo: []PhotoSize
├── Video: *Video
├── Document: *Document
├── Audio: *Audio
├── Sticker: *Sticker
└── Date: int (Unix timestamp)
```

### 2. EditedBusinessMessage

Срабатывает при редактировании сообщения.

```go
if update.EditedBusinessMessage != nil {
    msg := update.EditedBusinessMessage
    // Содержит то же самое, что и BusinessMessage
    // но с обновленным текстом/медиа
}
```

### 3. DeletedBusinessMessages

Срабатывает при удалении сообщений.

```go
if update.DeletedBusinessMessages != nil {
    deleted := update.DeletedBusinessMessages
    // deleted.MessageIDs - массив ID удаленных сообщений
    // deleted.BusinessConnectionID - ID бизнес-подключения
}
```

**Структура:**

```
tgbotapi.DeletedBusinessMessages:
├── BusinessConnectionID: string
└── MessageIDs: []int
```

## Типы медиа

### Фото

```go
if msg.Photo != nil && len(msg.Photo) > 0 {
    photo := msg.Photo[len(msg.Photo)-1] // Последнее фото (наилучшее качество)
    
    fileID := photo.FileID
    fileUniqueID := photo.FileUniqueID
    width := photo.Width
    height := photo.Height
    fileSize := photo.FileSize
}
```

### Видео

```go
if msg.Video != nil {
    video := msg.Video
    
    fileID := video.FileID
    fileUniqueID := video.FileUniqueID
    width := video.Width
    height := video.Height
    duration := video.Duration
    fileSize := video.FileSize
    mimeType := video.MimeType
    thumbnail := video.Thumbnail
}
```

### Документ

```go
if msg.Document != nil {
    doc := msg.Document
    
    fileID := doc.FileID
    fileUniqueID := doc.FileUniqueID
    fileSize := doc.FileSize
    fileName := doc.FileName
    mimeType := doc.MimeType
    thumbnail := doc.Thumbnail
}
```

### Аудио

```go
if msg.Audio != nil {
    audio := msg.Audio
    
    fileID := audio.FileID
    fileUniqueID := audio.FileUniqueID
    duration := audio.Duration
    performer := audio.Performer
    title := audio.Title
    mimeType := audio.MimeType
    fileSize := audio.FileSize
}
```

### Голосовое сообщение

```go
if msg.Voice != nil {
    voice := msg.Voice
    
    fileID := voice.FileID
    duration := voice.Duration
    mimeType := voice.MimeType
    fileSize := voice.FileSize
}
```

## Загрузка файлов

### Получение информации о файле

```go
fileConfig := tgbotapi.FileConfig{FileID: "file_id_here"}
file, err := bot.GetFile(fileConfig)
if err != nil {
    log.Fatal(err)
}

fileURL := file.Link(bot.Token)
fmt.Println("File URL:", fileURL)
```

### Загрузка файла по URL

```go
import "net/http"

resp, err := http.Get(fileURL)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

data, err := io.ReadAll(resp.Body)
if err != nil {
    log.Fatal(err)
}
```

## Отправка сообщений пользователю

### Отправка текстового сообщения

```go
msg := tgbotapi.NewMessage(userID, "Hello, World!")
bot.Send(msg)
```

### Отправка перепосланного сообщения

```go
msg := tgbotapi.NewForward(userID, chatID, messageID)
bot.Send(msg)
```

### Отправка сообщения с кнопками

```go
msg := tgbotapi.NewMessage(userID, "Choose action:")

// Создание кнопок
mainMenu := tgbotapi.NewReplyKeyboard(
    tgbotapi.NewKeyboardButtonRow(
        tgbotapi.NewKeyboardButton("View History"),
        tgbotapi.NewKeyboardButton("Settings"),
    ),
    tgbotapi.NewKeyboardButtonRow(
        tgbotapi.NewKeyboardButton("Help"),
    ),
)

msg.ReplyMarkup = mainMenu
bot.Send(msg)
```

## Уведомления о изменениях

### Уведомление об удалении

```go
notification := fmt.Sprintf(
    "❌ Сообщение было удалено\n\n"+
    "Исходный текст: %s\n"+
    "Удалено в: %s",
    message.Text,
    time.Now().Format("2006-01-02 15:04:05"),
)

msg := tgbotapi.NewMessage(userID, notification)
bot.Send(msg)
```

### Уведомление об редактировании

```go
notification := fmt.Sprintf(
    "✏️ Сообщение было отредактировано\n\n"+
    "Первоначально: %s\n"+
    "Отредактировано на: %s\n"+
    "Время: %s",
    oldText,
    newText,
    time.Now().Format("2006-01-02 15:04:05"),
)

msg := tgbotapi.NewMessage(userID, notification)
bot.Send(msg)
```

## Обработка ошибок Telegram API

### Частые ошибки

```go
if err != nil {
    switch err {
    case tgbotapi.ErrAPIForbidden:
        log.Println("Bot not authorized - check token")
    case tgbotapi.ErrAPIBadRequest:
        log.Println("Bad request - check parameters")
    default:
        if strings.Contains(err.Error(), "chat not found") {
            log.Println("Chat ID is invalid")
        } else if strings.Contains(err.Error(), "message to edit not found") {
            log.Println("Message cannot be edited")
        }
    }
}
```

### Retry Logic

```go
import "time"

func retryGetFile(bot *tgbotapi.BotAPI, fileID string, maxRetries int) (*tgbotapi.File, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
        if err == nil {
            return file, nil
        }
        
        lastErr = err
        time.Sleep(time.Second * time.Duration(1<<uint(i))) // Exponential backoff
    }
    
    return nil, lastErr
}
```

## Лимиты Telegram API

- Максимальный размер файла: 20 MB для скачивания, 50 MB для загрузки
- Лимит запросов: ~30 в секунду
- Timeout на запросы: 30 секунд (установлено в коде)

## Синхронизация состояния

### Восстановление после перезагрузки

```go
func syncWithDatabase(bot *tgbotapi.BotAPI) {
    // Получить все сообщения за последний час
    oneHourAgo := time.Now().Add(-1 * time.Hour)
    
    // SELECT из БД
    query := `
        SELECT id, text FROM messages 
        WHERE created_at > $1
    `
    
    // Проверить их статус
    // Если в Telegram их нет - помечу как удаленные
}
```

## Безопасность

### Валидация токена

```go
if len(cfg.TelegramBotToken) == 0 {
    log.Fatal("Bot token is empty")
}

if !strings.Contains(cfg.TelegramBotToken, ":") {
    log.Fatal("Invalid bot token format")
}
```

### Проверка прав доступа

```go
// Убедитесь, что бот имеет права на лчтение бизнес-сообщений
if !botAPI.Self.IsBot {
    log.Fatal("Token belongs to user, not bot")
}
```

## Логирование Telegram запросов

```go
botAPI.Debug = true  // Включите для отладки
// Будут выводиться все запросы к API
```

## Тестирование

### Mock для тестирования

```go
type MockBotAPI struct {
    messages map[string]*tgbotapi.Message
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
    // Mock implementation
    return tgbotapi.Message{}, nil
}
```
