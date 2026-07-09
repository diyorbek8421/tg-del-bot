package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-business-bot/internal/delivery/telegram"
	"telegram-business-bot/internal/service"
	"telegram-business-bot/internal/service/mocks"
)

func main() {
	// Load Telegram bot token from environment
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatalf("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	// Initialize Telegram bot
	botAPI, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	botAPI.Debug = false
	log.Printf("✅ Authorized on account @%s", botAPI.Self.UserName)

	// Initialize mock repository and service
	mockRepo := mocks.NewMockMessageRepository()
	messageService := service.NewMessageService(mockRepo)

	// Initialize handler
	handler := telegram.NewTelegramHandler(botAPI, messageService, 0)

	// Get updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botAPI.GetUpdatesChan(u)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("🚀 Bot started (DEMO MODE - In-Memory Storage)")
	log.Println("⚠️  Note: This demo uses in-memory storage, data will be lost on restart")
	log.Println("📝 For persistent storage, run with PostgreSQL database")
	log.Println("Waiting for updates...")

	for {
		select {
		case update := <-updates:
			if err := handler.HandleUpdates(telegram.BusinessUpdate{
				UpdateID:      update.UpdateID,
				Message:       update.Message,
				EditedMessage: update.EditedMessage,
			}); err != nil {
				log.Printf("Error handling update: %v", err)
			}

		case <-sigChan:
			log.Println("Shutting down...")
			return
		}
	}
}
