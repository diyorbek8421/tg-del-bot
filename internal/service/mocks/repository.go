package mocks

import (
	"context"

	"telegram-business-bot/internal/domain"
)

// MockMessageRepository - mock for testing
type MockMessageRepository struct {
	messages map[int64]*domain.Message
}

func NewMockMessageRepository() *MockMessageRepository {
	return &MockMessageRepository{
		messages: make(map[int64]*domain.Message),
	}
}

func (m *MockMessageRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	m.messages[msg.MessageID] = msg
	return nil
}

func (m *MockMessageRepository) GetMessageByID(ctx context.Context, messageID int64) (*domain.Message, error) {
	if msg, ok := m.messages[messageID]; ok {
		return msg, nil
	}
	return nil, nil
}

func (m *MockMessageRepository) UpdateMessage(ctx context.Context, msg *domain.Message) error {
	m.messages[msg.MessageID] = msg
	return nil
}

func (m *MockMessageRepository) SaveMessageEdit(ctx context.Context, edit *domain.MessageEdit) error {
	return nil
}

func (m *MockMessageRepository) SaveMessageDeletion(ctx context.Context, deletion *domain.MessageDeletion) error {
	return nil
}

func (m *MockMessageRepository) SaveBusinessSubscriber(ctx context.Context, sub *domain.BusinessSubscriber) error {
	return nil
}

func (m *MockMessageRepository) GetMessagesByChatID(ctx context.Context, chatID int64, limit int) ([]*domain.Message, error) {
	return nil, nil
}

func (m *MockMessageRepository) GetBusinessSubscribersByConnectionID(ctx context.Context, connectionID string) ([]*domain.BusinessSubscriber, error) {
	return nil, nil
}

func (m *MockMessageRepository) GetAllBusinessSubscribers(ctx context.Context) ([]*domain.BusinessSubscriber, error) {
	return nil, nil
}

func (m *MockMessageRepository) GetBusinessSubscribersByChatID(ctx context.Context, chatID int64) ([]*domain.BusinessSubscriber, error) {
	return nil, nil
}

func (m *MockMessageRepository) MarkMessageAsDeleted(ctx context.Context, messageID int64) error {
	return nil
}
