package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"telegram-business-bot/internal/domain"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, msg *domain.Message) error
	GetMessageByID(ctx context.Context, messageID int64) (*domain.Message, error)
	UpdateMessage(ctx context.Context, msg *domain.Message) error
	SaveMessageEdit(ctx context.Context, edit *domain.MessageEdit) error
	SaveMessageDeletion(ctx context.Context, deletion *domain.MessageDeletion) error
	GetMessagesByChatID(ctx context.Context, chatID int64, limit int) ([]*domain.Message, error)
	MarkMessageAsDeleted(ctx context.Context, messageID int64) error
	SaveBusinessSubscriber(ctx context.Context, sub *domain.BusinessSubscriber) error
	GetAllBusinessSubscribers(ctx context.Context) ([]*domain.BusinessSubscriber, error)
	GetBusinessSubscribersByConnectionID(ctx context.Context, connectionID string) ([]*domain.BusinessSubscriber, error)
	GetBusinessSubscribersByChatID(ctx context.Context, chatID int64) ([]*domain.BusinessSubscriber, error)
}

type messageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	query := `
		INSERT INTO messages (chat_id, user_id, username, message_id, business_message_id, text, media_type, media_file_id, media_path, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (chat_id, message_id) DO UPDATE SET
			text = $6,
			media_type = $7,
			media_file_id = $8,
			media_path = $9,
			updated_at = $11,
			username = COALESCE(EXCLUDED.username, messages.username),
			business_message_id = COALESCE(EXCLUDED.business_message_id, messages.business_message_id)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		msg.ChatID,
		msg.UserID,
		msg.Username,
		msg.MessageID,
		msg.BusinessMessageID,
		msg.Text,
		msg.MediaType,
		msg.MediaFileID,
		msg.MediaPath,
		msg.CreatedAt,
		msg.UpdatedAt,
	).Scan(&msg.ID)

	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

func (r *messageRepository) GetMessageByID(ctx context.Context, messageID int64) (*domain.Message, error) {
	query := `
		SELECT id, chat_id, user_id, username, message_id, business_message_id, text, media_type, media_file_id, media_path, created_at, updated_at, deleted_at
		FROM messages
		WHERE message_id = $1 OR business_message_id = $1
	`

	msg := &domain.Message{}
	var username *string
	err := r.db.QueryRow(ctx, query, messageID).Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.UserID,
		&username,
		&msg.MessageID,
		&msg.BusinessMessageID,
		&msg.Text,
		&msg.MediaType,
		&msg.MediaFileID,
		&msg.MediaPath,
		&msg.CreatedAt,
		&msg.UpdatedAt,
		&msg.DeletedAt,
	)
	if username != nil {
		msg.Username = *username
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return msg, nil
}

func (r *messageRepository) UpdateMessage(ctx context.Context, msg *domain.Message) error {
	query := `
		UPDATE messages
		SET text = $1, media_type = $2, media_file_id = $3, media_path = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(ctx, query,
		msg.Text,
		msg.MediaType,
		msg.MediaFileID,
		msg.MediaPath,
		time.Now(),
		msg.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

func (r *messageRepository) SaveMessageEdit(ctx context.Context, edit *domain.MessageEdit) error {
	query := `
		INSERT INTO message_edits (message_id, old_text, new_text, edited_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query,
		edit.MessageID,
		edit.OldText,
		edit.NewText,
		edit.EditedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save message edit: %w", err)
	}

	return nil
}

func (r *messageRepository) SaveMessageDeletion(ctx context.Context, deletion *domain.MessageDeletion) error {
	query := `
		INSERT INTO message_deletions (message_id, deleted_at)
		VALUES ($1, $2)
	`

	_, err := r.db.Exec(ctx, query,
		deletion.MessageID,
		deletion.DeletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save message deletion: %w", err)
	}

	return nil
}

func (r *messageRepository) GetMessagesByChatID(ctx context.Context, chatID int64, limit int) ([]*domain.Message, error) {
	query := `
		SELECT id, chat_id, user_id, username, message_id, business_message_id, text, media_type, media_file_id, media_path, created_at, updated_at, deleted_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, chatID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		var username *string
		err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.UserID,
			&username,
			&msg.MessageID,
			&msg.BusinessMessageID,
			&msg.Text,
			&msg.MediaType,
			&msg.MediaFileID,
			&msg.MediaPath,
			&msg.CreatedAt,
			&msg.UpdatedAt,
			&msg.DeletedAt,
		)
		if username != nil {
			msg.Username = *username
		}
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *messageRepository) MarkMessageAsDeleted(ctx context.Context, messageID int64) error {
	query := `
		UPDATE messages
		SET deleted_at = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, time.Now(), messageID)
	if err != nil {
		return fmt.Errorf("failed to mark message as deleted: %w", err)
	}

	return nil
}

func (r *messageRepository) SaveBusinessSubscriber(ctx context.Context, sub *domain.BusinessSubscriber) error {
	query := `
		INSERT INTO business_subscribers (business_connection_id, chat_id, username, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (business_connection_id, chat_id) DO UPDATE SET username = EXCLUDED.username
	`

	result, err := r.db.Exec(ctx, query,
		sub.BusinessConnectionID,
		sub.ChatID,
		sub.Username,
		sub.CreatedAt,
	)

	if err != nil {
		log.Printf("Error executing SaveBusinessSubscriber query: %v", err)
		return err
	}

	rows := result.RowsAffected()
	log.Printf("SaveBusinessSubscriber: connID=%s chatID=%d username=%s rows_affected=%d",
		sub.BusinessConnectionID, sub.ChatID, sub.Username, rows)

	return nil
}

func (r *messageRepository) GetBusinessSubscribersByConnectionID(ctx context.Context, connectionID string) ([]*domain.BusinessSubscriber, error) {
	query := `
		SELECT id, business_connection_id, chat_id, username, created_at
		FROM business_subscribers
		WHERE business_connection_id = $1
	`

	log.Printf("GetBusinessSubscribersByConnectionID: Querying subscribers for connectionID=%s", connectionID)

	rows, err := r.db.Query(ctx, query, connectionID)
	if err != nil {
		log.Printf("GetBusinessSubscribersByConnectionID: Query error: %v", err)
		return nil, fmt.Errorf("failed to get business subscribers: %w", err)
	}
	defer rows.Close()

	var subs []*domain.BusinessSubscriber
	for rows.Next() {
		sub := &domain.BusinessSubscriber{}
		var username *string
		err := rows.Scan(&sub.ID, &sub.BusinessConnectionID, &sub.ChatID, &username, &sub.CreatedAt)
		if err != nil {
			log.Printf("GetBusinessSubscribersByConnectionID: Scan error: %v", err)
			return nil, fmt.Errorf("failed to scan business subscriber: %w", err)
		}
		if username != nil {
			sub.Username = *username
		}
		subs = append(subs, sub)
		log.Printf("GetBusinessSubscribersByConnectionID: Scanned subscriber: ID=%d, chatID=%d, username=%s",
			sub.ID, sub.ChatID, sub.Username)
	}

	log.Printf("GetBusinessSubscribersByConnectionID: Found %d total subscribers for connectionID=%s", len(subs), connectionID)

	return subs, nil
}

func (r *messageRepository) GetBusinessSubscribersByChatID(ctx context.Context, chatID int64) ([]*domain.BusinessSubscriber, error) {
	query := `
		SELECT id, business_connection_id, chat_id, username, created_at
		FROM business_subscribers
		WHERE chat_id = $1
	`

	log.Printf("GetBusinessSubscribersByChatID: Querying subscriptions for chatID=%d", chatID)

	rows, err := r.db.Query(ctx, query, chatID)
	if err != nil {
		log.Printf("GetBusinessSubscribersByChatID: Query error: %v", err)
		return nil, fmt.Errorf("failed to get business subscribers by chat id: %w", err)
	}
	defer rows.Close()

	var subs []*domain.BusinessSubscriber
	for rows.Next() {
		sub := &domain.BusinessSubscriber{}
		var username *string
		err := rows.Scan(&sub.ID, &sub.BusinessConnectionID, &sub.ChatID, &username, &sub.CreatedAt)
		if err != nil {
			log.Printf("GetBusinessSubscribersByChatID: Scan error: %v", err)
			return nil, fmt.Errorf("failed to scan business subscriber: %w", err)
		}
		if username != nil {
			sub.Username = *username
		}
		subs = append(subs, sub)
		log.Printf("GetBusinessSubscribersByChatID: Scanned subscriber: ID=%d, chatID=%d, username=%s", sub.ID, sub.ChatID, sub.Username)
	}

	log.Printf("GetBusinessSubscribersByChatID: Found %d total subscriptions for chatID=%d", len(subs), chatID)

	return subs, nil
}

func (r *messageRepository) GetAllBusinessSubscribers(ctx context.Context) ([]*domain.BusinessSubscriber, error) {
	query := `
		SELECT id, business_connection_id, chat_id, username, created_at
		FROM business_subscribers
		ORDER BY created_at DESC
	`

	log.Printf("GetAllBusinessSubscribers: Querying all business subscribers")

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		log.Printf("GetAllBusinessSubscribers: Query error: %v", err)
		return nil, fmt.Errorf("failed to get all business subscribers: %w", err)
	}
	defer rows.Close()

	var subs []*domain.BusinessSubscriber
	for rows.Next() {
		sub := &domain.BusinessSubscriber{}
		var username *string
		err := rows.Scan(&sub.ID, &sub.BusinessConnectionID, &sub.ChatID, &username, &sub.CreatedAt)
		if err != nil {
			log.Printf("GetAllBusinessSubscribers: Scan error: %v", err)
			return nil, fmt.Errorf("failed to scan business subscriber: %w", err)
		}
		if username != nil {
			sub.Username = *username
		}
		subs = append(subs, sub)
		log.Printf("GetAllBusinessSubscribers: Scanned subscriber: ID=%d, chatID=%d, username=%s", sub.ID, sub.ChatID, sub.Username)
	}

	log.Printf("GetAllBusinessSubscribers: Found %d total subscribers", len(subs))

	return subs, nil
}
