package service

import (
	"context"
	"fmt"
	"time"

	"telegram-business-bot/internal/domain"
	"telegram-business-bot/internal/repository"
)

type MessageService interface {
	HandleNewBusinessMessage(ctx context.Context, msg *domain.Message) error
	HandleEditedBusinessMessage(ctx context.Context, oldMsg *domain.Message, newMsg *domain.Message) error
	HandleDeletedBusinessMessages(ctx context.Context, messageID int64, userID int64) error
	SaveBusinessSubscriber(ctx context.Context, sub *domain.BusinessSubscriber) error
	GetAllBusinessSubscribers(ctx context.Context) ([]*domain.BusinessSubscriber, error)
	GetBusinessSubscribersByConnectionID(ctx context.Context, connectionID string) ([]*domain.BusinessSubscriber, error)
	GetBusinessSubscribersByChatID(ctx context.Context, chatID int64) ([]*domain.BusinessSubscriber, error)
	GetMessageByID(ctx context.Context, messageID int64) (*domain.Message, error)
	GetMessageHistory(ctx context.Context, chatID int64) ([]*domain.Message, error)
}

type messageService struct {
	repo repository.MessageRepository
}

func NewMessageService(repo repository.MessageRepository) MessageService {
	return &messageService{repo: repo}
}

// HandleNewBusinessMessage saves a new business message
func (s *messageService) HandleNewBusinessMessage(ctx context.Context, msg *domain.Message) error {
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.UpdatedAt.IsZero() {
		msg.UpdatedAt = time.Now()
	}

	err := s.repo.SaveMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to handle new business message: %w", err)
	}

	return nil
}

// HandleEditedBusinessMessage processes edited message and saves edit history
func (s *messageService) HandleEditedBusinessMessage(ctx context.Context, oldMsg *domain.Message, newMsg *domain.Message) error {
	// Get existing message from DB to compare
	existingMsg, err := s.repo.GetMessageByID(ctx, oldMsg.MessageID)
	if err != nil {
		return fmt.Errorf("failed to get existing message: %w", err)
	}

	// Save edit history
	edit := &domain.MessageEdit{
		MessageID: existingMsg.ID,
		OldText:   existingMsg.Text,
		NewText:   newMsg.Text,
		EditedAt:  time.Now(),
	}

	err = s.repo.SaveMessageEdit(ctx, edit)
	if err != nil {
		return fmt.Errorf("failed to save edit history: %w", err)
	}

	// Update message in DB
	newMsg.ID = existingMsg.ID
	newMsg.UpdatedAt = time.Now()
	err = s.repo.UpdateMessage(ctx, newMsg)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

// HandleDeletedBusinessMessages processes deleted message
func (s *messageService) HandleDeletedBusinessMessages(ctx context.Context, messageID int64, userID int64) error {
	// Get message from DB
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		// Message not found in DB, skip
		return nil
	}

	// Save deletion record
	deletion := &domain.MessageDeletion{
		MessageID: msg.ID,
		DeletedAt: time.Now(),
	}

	err = s.repo.SaveMessageDeletion(ctx, deletion)
	if err != nil {
		return fmt.Errorf("failed to save message deletion: %w", err)
	}

	// Mark message as deleted
	err = s.repo.MarkMessageAsDeleted(ctx, msg.ID)
	if err != nil {
		return fmt.Errorf("failed to mark message as deleted: %w", err)
	}

	return nil
}

func (s *messageService) SaveBusinessSubscriber(ctx context.Context, sub *domain.BusinessSubscriber) error {
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = time.Now()
	}

	err := s.repo.SaveBusinessSubscriber(ctx, sub)
	if err != nil {
		return fmt.Errorf("failed to save business subscriber: %w", err)
	}

	return nil
}

func (s *messageService) GetAllBusinessSubscribers(ctx context.Context) ([]*domain.BusinessSubscriber, error) {
	subs, err := s.repo.GetAllBusinessSubscribers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all business subscribers: %w", err)
	}

	return subs, nil
}

func (s *messageService) GetBusinessSubscribersByConnectionID(ctx context.Context, connectionID string) ([]*domain.BusinessSubscriber, error) {
	subs, err := s.repo.GetBusinessSubscribersByConnectionID(ctx, connectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business subscribers: %w", err)
	}

	return subs, nil
}

func (s *messageService) GetBusinessSubscribersByChatID(ctx context.Context, chatID int64) ([]*domain.BusinessSubscriber, error) {
	subs, err := s.repo.GetBusinessSubscribersByChatID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business subscribers by chat id: %w", err)
	}

	return subs, nil
}

// GetMessageHistory returns message history for a chat
func (s *messageService) GetMessageByID(ctx context.Context, messageID int64) (*domain.Message, error) {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *messageService) GetMessageHistory(ctx context.Context, chatID int64) ([]*domain.Message, error) {
	messages, err := s.repo.GetMessagesByChatID(ctx, chatID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	return messages, nil
}
