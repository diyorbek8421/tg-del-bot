package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"telegram-business-bot/config"
	"telegram-business-bot/internal/domain"
	"telegram-business-bot/internal/repository"
	"telegram-business-bot/internal/service"
)

// Пример 1: Сохранение нового сообщения
func exampleSaveMessage(db *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messageRepo := repository.NewMessageRepository(db)
	messageService := service.NewMessageService(messageRepo)

	message := &domain.Message{
		ChatID:      123456,
		UserID:      789,
		MessageID:   1,
		Text:        "Hello from Business Account",
		MediaType:   "text",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := messageService.HandleNewBusinessMessage(ctx, message)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	log.Printf("Successfully saved message: %+v", message)
}

// Пример 2: Отслеживание редактирования
func exampleTrackMessageEdit(db *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messageRepo := repository.NewMessageRepository(db)
	messageService := service.NewMessageService(messageRepo)

	// Original message
	oldMessage := &domain.Message{
		MessageID: 1,
		Text:      "Original text",
	}

	// Edited message
	newMessage := &domain.Message{
		MessageID: 1,
		Text:      "Updated text with more info",
		UpdatedAt: time.Now(),
	}

	err := messageService.HandleEditedBusinessMessage(ctx, oldMessage, newMessage)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	log.Println("Message edit tracked successfully")
}

// Пример 3: Обработка удаления
func exampleHandleMessageDeletion(db *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messageRepo := repository.NewMessageRepository(db)
	messageService := service.NewMessageService(messageRepo)

	messageID := int64(1)
	userID := int64(789)

	err := messageService.HandleDeletedBusinessMessages(ctx, messageID, userID)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	log.Println("Message deletion processed")
}

// Пример 4: Получение истории сообщений
func exampleGetMessageHistory(db *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messageRepo := repository.NewMessageRepository(db)
	messageService := service.NewMessageService(messageRepo)

	chatID := int64(123456)

	messages, err := messageService.GetMessageHistory(ctx, chatID)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	log.Printf("Found %d messages:", len(messages))
	for _, msg := range messages {
		log.Printf("  ID: %d, Text: %s, CreatedAt: %s", msg.ID, msg.Text, msg.CreatedAt)
	}
}

// Пример 5: Получение редактирования из БД
func exampleGetMessageEdits(db *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT old_text, new_text, edited_at
		FROM message_edits
		WHERE message_id = $1
		ORDER BY edited_at ASC
	`

	rows, err := db.Query(ctx, query, 1)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer rows.Close()

	log.Println("Message edit history:")
	for rows.Next() {
		var oldText, newText string
		var editedAt time.Time

		err := rows.Scan(&oldText, &newText, &editedAt)
		if err != nil {
			log.Printf("Error scanning: %v", err)
			continue
		}

		log.Printf("  %s: '%s' → '%s'", editedAt.Format("2006-01-02 15:04:05"), oldText, newText)
	}
}

// Пример 6: Получение удаленных сообщений
func exampleGetDeletedMessages(db *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT m.text, md.deleted_at
		FROM message_deletions md
		JOIN messages m ON m.id = md.message_id
		WHERE m.chat_id = $1
		ORDER BY md.deleted_at DESC
	`

	rows, err := db.Query(ctx, query, 123456)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer rows.Close()

	log.Println("Deleted messages:")
	for rows.Next() {
		var text string
		var deletedAt time.Time

		err := rows.Scan(&text, &deletedAt)
		if err != nil {
			log.Printf("Error scanning: %v", err)
			continue
		}

		log.Printf("  %s (удалено в %s)", text, deletedAt.Format("2006-01-02 15:04:05"))
	}
}

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	connStr := "postgres://" + cfg.Database.User + ":" + cfg.Database.Password + "@" + 
		cfg.Database.Host + ":" + string(rune(cfg.Database.Port)) + "/" + cfg.Database.Name + "?sslmode=disable"
	
	db, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}
	defer db.Close()

	// Run examples
	log.Println("Example 1: Save Message")
	exampleSaveMessage(db)

	log.Println("\nExample 2: Track Message Edit")
	exampleTrackMessageEdit(db)

	log.Println("\nExample 3: Handle Message Deletion")
	exampleHandleMessageDeletion(db)

	log.Println("\nExample 4: Get Message History")
	exampleGetMessageHistory(db)

	log.Println("\nExample 5: Get Message Edits")
	exampleGetMessageEdits(db)

	log.Println("\nExample 6: Get Deleted Messages")
	exampleGetDeletedMessages(db)
}
