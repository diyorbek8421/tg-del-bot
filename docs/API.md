# API Документация

## Структуры данных

### Message

```go
type Message struct {
    ID                int64          // ID в БД
    ChatID            int64          // ID чата
    UserID            int64          // ID пользователя
    MessageID         int64          // ID сообщения в Telegram
    BusinessMessageID int64          // Business Message ID
    Text              string         // Текст сообщения
    MediaType         string         // Тип медиа (photo, video, document, audio)
    MediaFileID       string         // File ID из Telegram
    MediaPath         string         // Локальный путь к сохраненному файлу
    CreatedAt         time.Time      // Время создания
    UpdatedAt         time.Time      // Время последнего обновления
    DeletedAt         *time.Time     // Время удаления
}
```

### MessageEdit

```go
type MessageEdit struct {
    ID        int64
    MessageID int64      // ID сообщения в БД
    OldText   string     // Старый текст
    NewText   string     // Новый текст
    EditedAt  time.Time  // Время редактирования
}
```

### MessageDeletion

```go
type MessageDeletion struct {
    ID        int64
    MessageID int64      // ID сообщения в БД
    DeletedAt time.Time  // Время удаления
}
```

## Service методы

### HandleNewBusinessMessage

Сохраняет новое бизнес-сообщение.

```go
message := &domain.Message{
    ChatID:    123456,
    UserID:    789,
    MessageID: 1,
    Text:      "Hello world",
}

err := messageService.HandleNewBusinessMessage(ctx, message)
if err != nil {
    log.Fatal(err)
}
```

### HandleEditedBusinessMessage

Отслеживает изменения сообщения.

```go
oldMsg := &domain.Message{
    MessageID: 1,
}

newMsg := &domain.Message{
    MessageID: 1,
    Text:      "Updated text",
}

err := messageService.HandleEditedBusinessMessage(ctx, oldMsg, newMsg)
if err != nil {
    log.Fatal(err)
}
```

### HandleDeletedBusinessMessages

Обрабатывает удаленные сообщения.

```go
err := messageService.HandleDeletedBusinessMessages(ctx, messageID, userID)
if err != nil {
    log.Fatal(err)
}
```

### GetMessageHistory

Получает историю сообщений для чата.

```go
messages, err := messageService.GetMessageHistory(ctx, chatID)
if err != nil {
    log.Fatal(err)
}

for _, msg := range messages {
    fmt.Printf("ID: %d, Text: %s\n", msg.ID, msg.Text)
}
```

## Repository методы

### SaveMessage

```go
err := messageRepo.SaveMessage(ctx, message)
```

### GetMessageByID

```go
msg, err := messageRepo.GetMessageByID(ctx, messageID)
if err != nil {
    log.Fatal(err)
}
```

### UpdateMessage

```go
err := messageRepo.UpdateMessage(ctx, message)
```

### SaveMessageEdit

```go
edit := &domain.MessageEdit{
    MessageID: 1,
    OldText:   "Old",
    NewText:   "New",
}
err := messageRepo.SaveMessageEdit(ctx, edit)
```

### SaveMessageDeletion

```go
deletion := &domain.MessageDeletion{
    MessageID: 1,
}
err := messageRepo.SaveMessageDeletion(ctx, deletion)
```

### GetMessagesByChatID

```go
messages, err := messageRepo.GetMessagesByChatID(ctx, chatID, 100)
```

### MarkMessageAsDeleted

```go
err := messageRepo.MarkMessageAsDeleted(ctx, messageID)
```

## Storage интерфейс

### Сохранение файла

```go
storage, _ := storage.NewLocalStorage("./media")
path, err := storage.SaveFile("photos/file_123.jpg", data)
```

### Получение файла

```go
file, err := storage.GetFile("photos/file_123.jpg")
defer file.Close()
```

### Удаление файла

```go
err := storage.DeleteFile("photos/file_123.jpg")
```

## MediaDownloader

### Загрузка фото

```go
downloader := storage.NewMediaDownloader(botAPI, storage)
path, err := downloader.DownloadPhoto(fileID, "photos/photo_123.jpg")
```

### Загрузка видео

```go
path, err := downloader.DownloadVideo(fileID, "videos/video_123.mp4")
```

### Загрузка документа

```go
path, err := downloader.DownloadDocument(fileID, "documents/doc_123.pdf")
```

### Загрузка аудио

```go
path, err := downloader.DownloadAudio(fileID, "audio/audio_123.mp3")
```

## Примеры использования

### Пример 1: Обработка нового сообщения с медиа

```go
package main

import (
    "context"
    "log"
    "time"
    
    "telegram-business-bot/internal/domain"
    "telegram-business-bot/internal/storage"
)

func handleNewMessage(msg *domain.Message) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Download media if present
    if msg.MediaType == "photo" {
        downloader := storage.NewMediaDownloader(botAPI, localStorage)
        path, err := downloader.DownloadPhoto(msg.MediaFileID, "photos/"+msg.MediaFileID+".jpg")
        if err != nil {
            log.Printf("Error downloading photo: %v", err)
        } else {
            msg.MediaPath = path
        }
    }

    // Save to database
    err := messageService.HandleNewBusinessMessage(ctx, msg)
    if err != nil {
        return err
    }

    return nil
}
```

### Пример 2: Получение истории отредактирования

```go
package main

import (
    "context"
    "log"
)

func getEditHistory(messageID int64) {
    query := `
        SELECT old_text, new_text, edited_at
        FROM message_edits
        WHERE message_id = $1
        ORDER BY edited_at DESC
    `
    
    rows, err := db.Query(context.Background(), query, messageID)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var oldText, newText string
        var editedAt time.Time
        
        err = rows.Scan(&oldText, &newText, &editedAt)
        if err != nil {
            log.Fatal(err)
        }

        log.Printf("Edited at %s: '%s' -> '%s'", editedAt, oldText, newText)
    }
}
```

### Пример 3: Получение удаленных сообщений

```go
package main

import (
    "context"
    "log"
)

func getDeletedMessages(chatID int64) {
    query := `
        SELECT m.text, md.deleted_at
        FROM message_deletions md
        JOIN messages m ON m.id = md.message_id
        WHERE m.chat_id = $1
        ORDER BY md.deleted_at DESC
    `
    
    rows, err := db.Query(context.Background(), query, chatID)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var text string
        var deletedAt time.Time
        
        err = rows.Scan(&text, &deletedAt)
        if err != nil {
            log.Fatal(err)
        }

        log.Printf("Deleted at %s: %s", deletedAt, text)
    }
}
```

## Обработка ошибок

Все функции возвращают `error` интерфейс. Всегда проверяйте ошибки:

```go
if err != nil {
    log.Printf("Error: %v", err)
    // Handle error appropriately
}
```

Рекомендуемые ошибки:

```go
go
fmt.Errorf("failed to save message: %w", err)
```

Используйте `%w` для оборачивания ошибок для лучшей трассировки.
