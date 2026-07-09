package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"telegram-business-bot/config"
	"telegram-business-bot/internal/app"
	"telegram-business-bot/internal/delivery/telegram"
	"telegram-business-bot/internal/repository"
	"telegram-business-bot/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	messageRepo := repository.NewMessageRepository(db)

	// Initialize services
	messageService := service.NewMessageService(messageRepo)

	// Initialize Telegram bot
	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	botAPI.Debug = false
	log.Printf("Authorized on account %s", botAPI.Self.UserName)

	// Initialize handler
	handler := telegram.NewTelegramHandler(botAPI, messageService, cfg.AdminChatID)

	// Start admin dashboard
	monitor := app.NewMonitor(messageService, cfg.AdminDashboardToken)
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		if err := monitor.Run(addr); err != nil && err != http.ErrServerClosed {
			log.Printf("Admin dashboard stopped with error: %v", err)
		}
	}()

	// Get business updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.AllowedUpdates = []string{
		"message",
		"edited_message",
		"business_message",
		"edited_business_message",
		"deleted_business_messages",
	}

	updates := getBusinessUpdates(botAPI, u)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Bot started. Waiting for updates...")

	for {
		select {
		case update, ok := <-updates:
			if !ok {
				log.Println("Updates channel closed, restarting...")
				updates = getBusinessUpdates(botAPI, u)
				continue
			}
			if err := handler.HandleUpdates(update); err != nil {
				log.Printf("Error handling update: %v", err)
			}

		case <-sigChan:
			log.Println("Shutting down...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := monitor.Shutdown(shutdownCtx); err != nil {
				log.Printf("Failed to shutdown admin dashboard: %v", err)
			}
			return
		}
	}
}

// initDatabase initializes database connection pool
func initDatabase(cfg *config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	// Test connection
	err = db.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")

	// Run migrations
	err = runMigrations(db)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations creates necessary tables if they don't exist
func runMigrations(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	migrationSQL := `
	-- Messages table
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		chat_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		message_id BIGINT NOT NULL,
		business_message_id BIGINT,
		text TEXT,
		media_type VARCHAR(50),
		media_file_id VARCHAR(255),
		media_path VARCHAR(500),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP,
		UNIQUE(chat_id, message_id)
	);

	-- Message edits history
	CREATE TABLE IF NOT EXISTS message_edits (
		id SERIAL PRIMARY KEY,
		message_id BIGINT NOT NULL REFERENCES messages(id),
		old_text TEXT,
		new_text TEXT,
		edited_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Message deletions
	CREATE TABLE IF NOT EXISTS message_deletions (
		id SERIAL PRIMARY KEY,
		message_id BIGINT NOT NULL REFERENCES messages(id),
		deleted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Business users
	CREATE TABLE IF NOT EXISTS business_users (
		id SERIAL PRIMARY KEY,
		business_account_id BIGINT NOT NULL,
		chat_id BIGINT NOT NULL,
		username VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(business_account_id, chat_id)
	);

	-- Business subscribers for connected users
	CREATE TABLE IF NOT EXISTS business_subscribers (
		id SERIAL PRIMARY KEY,
		business_connection_id TEXT NOT NULL,
		chat_id BIGINT NOT NULL,
		username VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(business_connection_id, chat_id)
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
	CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id);
	CREATE INDEX IF NOT EXISTS idx_messages_business_message_id ON messages(business_message_id);
	CREATE INDEX IF NOT EXISTS idx_message_edits_message_id ON message_edits(message_id);
	CREATE INDEX IF NOT EXISTS idx_message_deletions_message_id ON message_deletions(message_id);
	CREATE INDEX IF NOT EXISTS idx_business_users_chat_id ON business_users(chat_id);

	-- Ensure username exists on messages
	ALTER TABLE messages ADD COLUMN IF NOT EXISTS username VARCHAR(255);
	`

	_, err := db.Exec(ctx, migrationSQL)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

func getBusinessUpdates(bot *tgbotapi.BotAPI, config tgbotapi.UpdateConfig) <-chan telegram.BusinessUpdate {
	updates := make(chan telegram.BusinessUpdate, bot.Buffer)

	go func() {
		defer close(updates)

		for {
			resp, err := bot.Request(config)
			if err != nil {
				log.Printf("Failed to get updates: %v", err)
				time.Sleep(3 * time.Second)
				continue
			}

			var rawUpdates []telegram.BusinessUpdate
			err = json.Unmarshal(resp.Result, &rawUpdates)
			if err != nil {
				log.Printf("Failed to unmarshal updates: %v", err)
				log.Printf("Raw update payload: %s", string(resp.Result))
				time.Sleep(3 * time.Second)
				continue
			}

			for _, update := range rawUpdates {
				updates <- update
			}

			if len(rawUpdates) > 0 {
				config.Offset = rawUpdates[len(rawUpdates)-1].UpdateID + 1
			}
		}
	}()

	return updates
}
