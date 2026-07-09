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

-- Indexes
CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id);
CREATE INDEX IF NOT EXISTS idx_messages_business_message_id ON messages(business_message_id);
CREATE INDEX IF NOT EXISTS idx_message_edits_message_id ON message_edits(message_id);
CREATE INDEX IF NOT EXISTS idx_message_deletions_message_id ON message_deletions(message_id);
CREATE INDEX IF NOT EXISTS idx_business_users_chat_id ON business_users(chat_id);

-- Add username column if missing
ALTER TABLE messages ADD COLUMN IF NOT EXISTS username VARCHAR(255);
