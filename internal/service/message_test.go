package service

import (
	"context"
	"testing"
	"time"

	"telegram-business-bot/internal/domain"
	"telegram-business-bot/internal/service/mocks"
)

func TestHandleNewBusinessMessage(t *testing.T) {
	mockRepo := mocks.NewMockMessageRepository()
	service := NewMessageService(mockRepo)

	msg := &domain.Message{
		ChatID:    123456,
		UserID:    789,
		MessageID: 1,
		Text:      "Test message",
	}

	err := service.HandleNewBusinessMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify message was saved
	savedMsg, _ := mockRepo.GetMessageByID(context.Background(), msg.MessageID)
	if savedMsg == nil {
		t.Fatalf("Expected message to be saved")
	}

	if savedMsg.Text != "Test message" {
		t.Fatalf("Expected text 'Test message', got '%s'", savedMsg.Text)
	}
}

func TestHandleEditedBusinessMessage(t *testing.T) {
	mockRepo := mocks.NewMockMessageRepository()
	service := NewMessageService(mockRepo)

	originalMsg := &domain.Message{
		ID:        1,
		ChatID:    123456,
		UserID:    789,
		MessageID: 1,
		Text:      "Original text",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save original
	mockRepo.SaveMessage(context.Background(), originalMsg)

	// Edit message
	editedMsg := &domain.Message{
		ID:        1,
		ChatID:    123456,
		UserID:    789,
		MessageID: 1,
		Text:      "Edited text",
	}

	err := service.HandleEditedBusinessMessage(context.Background(), originalMsg, editedMsg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHandleDeletedBusinessMessages(t *testing.T) {
	mockRepo := mocks.NewMockMessageRepository()
	service := NewMessageService(mockRepo)

	msg := &domain.Message{
		ID:        1,
		MessageID: 1,
		ChatID:    123456,
		UserID:    789,
		Text:      "Message to delete",
		CreatedAt: time.Now(),
	}

	// Save message first
	mockRepo.SaveMessage(context.Background(), msg)

	// Delete message
	err := service.HandleDeletedBusinessMessages(context.Background(), msg.MessageID, msg.UserID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
