package domain

import "time"

// Message represents a business message
type Message struct {
	ID                int64
	ChatID            int64
	UserID            int64
	Username          string
	MessageID         int64
	BusinessMessageID int64
	Text              string
	MediaType         string // "photo", "video", "document", "audio", etc.
	MediaFileID       string
	MediaPath         string // Local path to saved media
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// MessageEdit represents a message edit history
type MessageEdit struct {
	ID                int64
	MessageID         int64
	OldText           string
	NewText           string
	EditedAt          time.Time
}

// MessageDeletion represents a deleted message
type MessageDeletion struct {
	ID        int64
	MessageID int64
	DeletedAt time.Time
}

// BusinessUser represents a business account user
type BusinessSubscriber struct {
	ID                   int64
	BusinessConnectionID string
	ChatID               int64
	Username             string
	CreatedAt            time.Time
}
