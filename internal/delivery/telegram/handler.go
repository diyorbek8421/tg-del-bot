package telegram

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-business-bot/internal/domain"
	"telegram-business-bot/internal/service"
)

type BusinessMessage struct {
	tgbotapi.Message
	BusinessConnectionID string `json:"business_connection_id,omitempty"`
	BusinessMessageID    int64  `json:"business_message_id,omitempty"`
}

type DeletedBusinessMessages struct {
	BusinessConnectionID string         `json:"business_connection_id,omitempty"`
	Chat                 *tgbotapi.Chat `json:"chat,omitempty"`
	MessageIDs           []int64        `json:"message_ids,omitempty"`
}

type BusinessUpdate struct {
	UpdateID                int                      `json:"update_id"`
	Message                 *tgbotapi.Message        `json:"message,omitempty"`
	EditedMessage           *tgbotapi.Message        `json:"edited_message,omitempty"`
	BusinessMessage         *BusinessMessage         `json:"business_message,omitempty"`
	EditedBusinessMessage   *BusinessMessage         `json:"edited_business_message,omitempty"`
	DeletedBusinessMessages *DeletedBusinessMessages `json:"deleted_business_messages,omitempty"`
}

type TelegramHandler struct {
	bot                *tgbotapi.BotAPI
	messageService     service.MessageService
	adminChatID        int64
	processedUpdateIDs map[int]time.Time
	mu                 sync.Mutex
}

func NewTelegramHandler(bot *tgbotapi.BotAPI, messageService service.MessageService, adminChatID int64) *TelegramHandler {
	return &TelegramHandler{
		bot:                bot,
		messageService:     messageService,
		adminChatID:        adminChatID,
		processedUpdateIDs: make(map[int]time.Time),
	}
}

// HandleUpdates processes incoming Telegram updates
func (h *TelegramHandler) HandleUpdates(update BusinessUpdate) error {
	if update.UpdateID != 0 && h.isDuplicateUpdate(update.UpdateID) {
		log.Printf("Skipping duplicate update: id=%d", update.UpdateID)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Received business update: id=%d message=%t edited_business=%t deleted_business=%t edited=%t normal=%t",
		update.UpdateID,
		update.BusinessMessage != nil,
		update.EditedBusinessMessage != nil,
		update.DeletedBusinessMessages != nil,
		update.EditedMessage != nil,
		update.Message != nil,
	)

	// Handle Business Message (new message)
	if update.BusinessMessage != nil {
		return h.handleBusinessMessage(ctx, update.BusinessMessage)
	}

	// Handle Edited Business Message
	if update.EditedBusinessMessage != nil {
		return h.handleEditedBusinessMessage(ctx, update.EditedBusinessMessage)
	}

	// Handle Deleted Business Messages
	if update.DeletedBusinessMessages != nil {
		return h.handleDeletedBusinessMessages(ctx, update.DeletedBusinessMessages)
	}

	// Handle edited messages
	if update.EditedMessage != nil {
		return h.handleEditedMessage(ctx, update.EditedMessage)
	}

	// Handle regular messages
	if update.Message != nil {
		return h.handleMessage(ctx, update.Message)
	}

	return nil
}

// handleCommand handles bot commands
func (h *TelegramHandler) handleCommand(ctx context.Context, msg *tgbotapi.Message) error {
	if msg == nil || msg.Text == "" {
		return nil
	}

	switch msg.Command() {
	case "start":
		return h.commandStart(msg)
	case "help":
		return h.commandHelp(msg)
	}

	return nil
}

func (h *TelegramHandler) commandStart(msg *tgbotapi.Message) error {
	text := `👋 Добро пожаловать\!
Я — ssancodel bot \- Пока остальные общаются, я внимательно слежу за чатом\. 👀
Если кто\-то удалит сообщение — оно не исчезнет бесследно\. Я сохраню его и покажу удалённый текст\. 🗂️

Для подробной справки используйте /help`

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = "MarkdownV2"
	reply.ReplyToMessageID = msg.MessageID

	_, err := h.bot.Send(reply)
	return err
}

func (h *TelegramHandler) commandHelp(msg *tgbotapi.Message) error {
	text := `📖 *Помощь*

🤖 *Что умеет бот?*
• Отслеживает удалённые сообщения\.
• Отправляет удалённый текст обратно в чат\.
• Работает автоматически после подключения\.

🔧 *Как подключить бота?*

1\. Откройте Telegram\.
2\. Перейдите в ⚙️ Настройки\.
3\. Нажмите ✏️ Изменить профиль\.
4\. Откройте раздел «Автоматизация чатов»\.
5\. Нажмите «Добавить бота»\.
6\. Введите или выберите юзернейм бота \(@ВашБот\)\.
7\. Готово\! Теперь бот будет работать в ваших чатах\.

❓ *Если бот не работает:*
• Проверьте, что он добавлен в «Автоматизацию чатов»\.
• Убедитесь, что указан правильный юзернейм\.
• Попробуйте удалить и добавить бота заново\.

💬 Если возникли вопросы или нашли ошибку — свяжитесь с разработчиком\. `

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = "MarkdownV2"
	reply.ReplyToMessageID = msg.MessageID

	_, err := h.bot.Send(reply)
	return err
}

// handleMessage handles new messages
func (h *TelegramHandler) handleMessage(ctx context.Context, msg *tgbotapi.Message) error {
	if msg == nil || msg.From == nil || msg.Chat == nil {
		return nil
	}

	log.Printf("Handling new message from user %d in chat %d", msg.From.ID, msg.Chat.ID)

	// Handle commands
	if msg.IsCommand() {
		return h.handleCommand(ctx, msg)
	}

	if msg.ReplyToMessage != nil {
		h.handleReplyToMedia(ctx, msg, "")
	}

	message := &domain.Message{
		ChatID:    msg.Chat.ID,
		UserID:    msg.From.ID,
		Username:  msg.From.UserName,
		MessageID: int64(msg.MessageID),
		Text:      msg.Text,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Handle media
	if msg.Photo != nil && len(msg.Photo) > 0 {
		message.MediaType = "photo"
		message.MediaFileID = msg.Photo[len(msg.Photo)-1].FileID
	} else if msg.Video != nil {
		message.MediaType = "video"
		message.MediaFileID = msg.Video.FileID
	} else if msg.Document != nil {
		message.MediaType = "document"
		message.MediaFileID = msg.Document.FileID
	} else if msg.Audio != nil {
		message.MediaType = "audio"
		message.MediaFileID = msg.Audio.FileID
	}

	err := h.messageService.HandleNewBusinessMessage(ctx, message)
	if err != nil {
		log.Printf("Error handling message: %v", err)
		return err
	}

	log.Printf("Successfully saved message %d", msg.MessageID)

	// Send confirmation to user
	textPreview := msg.Text
	if len(textPreview) > 40 {
		textPreview = textPreview[:40] + "…"
	}

	emoji := "📝"
	if msg.Photo != nil && len(msg.Photo) > 0 {
		emoji = "📸"
	} else if msg.Video != nil {
		emoji = "🎬"
	} else if msg.Document != nil {
		emoji = "📄"
	} else if msg.Audio != nil {
		emoji = "🎵"
	}

	replyText := fmt.Sprintf("✅ *Сообщение сохранено*\n\n%s `%s`", emoji, escapeMarkdownV2(textPreview))
	reply := tgbotapi.NewMessage(msg.Chat.ID, replyText)
	reply.ParseMode = "MarkdownV2"
	reply.ReplyToMessageID = msg.MessageID

	_, err = h.bot.Send(reply)
	if err != nil {
		log.Printf("Error sending confirmation: %v", err)
	}

	return nil
}

// handleBusinessMessage handles new business messages
func (h *TelegramHandler) handleBusinessMessage(ctx context.Context, msg *BusinessMessage) error {
	if msg == nil || msg.From == nil || msg.Chat == nil {
		return nil
	}

	log.Printf("Handling new business message from user %d in chat %d, businessConnectionID=%s, businessMessageID=%d",
		msg.From.ID, msg.Chat.ID, msg.BusinessConnectionID, msg.BusinessMessageID)

	if msg.BusinessConnectionID != "" {
		if msg.Chat.ID != h.adminChatID {
			log.Printf("Attempting to save subscriber: businessConnectionID=%s, chatID=%d, username=%s",
				msg.BusinessConnectionID, msg.Chat.ID, msg.From.UserName)
			err := h.messageService.SaveBusinessSubscriber(ctx, &domain.BusinessSubscriber{
				BusinessConnectionID: msg.BusinessConnectionID,
				ChatID:               msg.Chat.ID,
				Username:             msg.From.UserName,
			})
			if err != nil {
				log.Printf("Error saving business subscriber: %v", err)
			} else {
				log.Printf("Successfully saved subscriber for businessConnectionID=%s, chatID=%d", msg.BusinessConnectionID, msg.Chat.ID)
			}
		} else {
			log.Printf("Skipping saving admin chat %d as subscriber", msg.Chat.ID)
		}
	} else {
		log.Printf("WARNING: businessConnectionID is empty for message from user %d in chat %d", msg.From.ID, msg.Chat.ID)
	}

	if msg.ReplyToMessage != nil {
		h.handleReplyToMedia(ctx, &msg.Message, msg.BusinessConnectionID)
	}

	message := &domain.Message{
		ChatID:            msg.Chat.ID,
		UserID:            msg.From.ID,
		Username:          msg.From.UserName,
		MessageID:         int64(msg.MessageID),
		BusinessMessageID: msg.BusinessMessageID,
		Text:              msg.Text,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if msg.Photo != nil && len(msg.Photo) > 0 {
		message.MediaType = "photo"
		message.MediaFileID = msg.Photo[len(msg.Photo)-1].FileID
	} else if msg.Video != nil {
		message.MediaType = "video"
		message.MediaFileID = msg.Video.FileID
	} else if msg.Document != nil {
		message.MediaType = "document"
		message.MediaFileID = msg.Document.FileID
	} else if msg.Audio != nil {
		message.MediaType = "audio"
		message.MediaFileID = msg.Audio.FileID
	}

	err := h.messageService.HandleNewBusinessMessage(ctx, message)
	if err != nil {
		log.Printf("Error handling business message: %v", err)
		return err
	}

	log.Printf("Successfully saved business message %d", msg.MessageID)
	return nil
}

func (h *TelegramHandler) handleReplyToMedia(ctx context.Context, msg *tgbotapi.Message, businessConnectionID string) {
	if msg == nil || msg.ReplyToMessage == nil || !isMediaMessage(msg.ReplyToMessage) {
		return
	}

	caption := fmt.Sprintf("Ответ на медиа от @%s", msg.From.UserName)
	if msg.Text != "" {
		caption = fmt.Sprintf("%s\n%s", caption, msg.Text)
	}

	if businessConnectionID == "" || msg.Chat == nil {
		return
	}

	if msg.Chat.ID == h.adminChatID {
		subs, err := h.messageService.GetBusinessSubscribersByConnectionID(ctx, businessConnectionID)
		if err != nil {
			log.Printf("Error getting business subscribers for admin reply: %v", err)
			return
		}
		for _, sub := range subs {
			if sub.ChatID == h.adminChatID {
				continue
			}
			if err := h.sendReplyMediaToChat(sub.ChatID, msg.ReplyToMessage, caption); err != nil {
				log.Printf("Error sending admin reply media to subscriber %d: %v", sub.ChatID, err)
			}
		}
		return
	}

	log.Printf("Skipping admin notification for subscriber reply media from chat %d", msg.Chat.ID)
}

func isMediaMessage(msg *tgbotapi.Message) bool {
	if msg == nil {
		return false
	}

	return len(msg.Photo) > 0 || msg.Video != nil || msg.Voice != nil || msg.Audio != nil || msg.Document != nil || msg.Animation != nil || msg.VideoNote != nil || msg.Sticker != nil
}

func (h *TelegramHandler) sendReplyMediaToChat(chatID int64, replyMsg *tgbotapi.Message, caption string) error {
	if chatID == 0 || replyMsg == nil {
		return fmt.Errorf("invalid target or reply message")
	}

	send := func(cfg tgbotapi.Chattable, mediaType string) error {
		_, err := h.bot.Send(cfg)
		if err != nil {
			log.Printf("Direct %s send failed for chat %d: %v", mediaType, chatID, err)
		}
		return err
	}

	if len(replyMsg.Photo) > 0 {
		fileID := replyMsg.Photo[len(replyMsg.Photo)-1].FileID
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fileID))
		photo.Caption = caption
		if err := send(photo, "photo"); err == nil {
			return nil
		}
	}

	if replyMsg.Video != nil {
		video := tgbotapi.NewVideo(chatID, tgbotapi.FileID(replyMsg.Video.FileID))
		video.Caption = caption
		if err := send(video, "video"); err == nil {
			return nil
		}
	}

	if replyMsg.Document != nil {
		doc := tgbotapi.NewDocument(chatID, tgbotapi.FileID(replyMsg.Document.FileID))
		doc.Caption = caption
		if err := send(doc, "document"); err == nil {
			return nil
		}
	}

	if replyMsg.Audio != nil {
		audio := tgbotapi.NewAudio(chatID, tgbotapi.FileID(replyMsg.Audio.FileID))
		audio.Caption = caption
		if err := send(audio, "audio"); err == nil {
			return nil
		}
	}

	if replyMsg.Voice != nil {
		voice := tgbotapi.NewVoice(chatID, tgbotapi.FileID(replyMsg.Voice.FileID))
		voice.Caption = caption
		if err := send(voice, "voice"); err == nil {
			return nil
		}
	}

	if replyMsg.Animation != nil {
		animation := tgbotapi.NewAnimation(chatID, tgbotapi.FileID(replyMsg.Animation.FileID))
		animation.Caption = caption
		if err := send(animation, "animation"); err == nil {
			return nil
		}
	}

	if replyMsg.VideoNote != nil {
		videoNote := tgbotapi.NewVideoNote(chatID, int(replyMsg.VideoNote.Length), tgbotapi.FileID(replyMsg.VideoNote.FileID))
		if err := send(videoNote, "video_note"); err == nil {
			return nil
		}
	}

	if replyMsg.Sticker != nil {
		sticker := tgbotapi.NewSticker(chatID, tgbotapi.FileID(replyMsg.Sticker.FileID))
		if err := send(sticker, "sticker"); err == nil {
			return nil
		}
	}

	log.Printf("Direct media send failed for chat %d, attempting download/reupload fallback", chatID)
	if err := h.downloadAndSendReplyMediaToChat(chatID, replyMsg, caption); err == nil {
		return nil
	}

	log.Printf("Download/reupload fallback failed for chat %d, attempting copy/forward fallback", chatID)
	if replyMsg.Chat != nil {
		if err := h.copyMessageToChat(chatID, replyMsg.Chat.ID, replyMsg.MessageID, caption); err == nil {
			return nil
		}
		log.Printf("Copy/forward failed for chat %d, reply from chat %d msg=%d", chatID, replyMsg.Chat.ID, replyMsg.MessageID)
	} else {
		log.Printf("Cannot copy/forward media for chat %d: replyMsg.Chat is nil", chatID)
	}

	return fmt.Errorf("failed to deliver reply media to chat %d", chatID)
}

func (h *TelegramHandler) downloadAndSendReplyMediaToChat(chatID int64, replyMsg *tgbotapi.Message, caption string) error {
	if chatID == 0 || replyMsg == nil {
		return fmt.Errorf("invalid target or reply message")
	}

	download := func(fileID string) ([]byte, error) {
		if fileID == "" {
			return nil, fmt.Errorf("empty file id")
		}

		url, err := h.bot.GetFileDirectURL(fileID)
		if err != nil {
			return nil, err
		}

		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("download failed: status %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	if len(replyMsg.Photo) > 0 {
		fileID := replyMsg.Photo[len(replyMsg.Photo)-1].FileID
		data, err := download(fileID)
		if err != nil {
			return err
		}
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{Name: "photo.jpg", Bytes: data})
		photo.Caption = caption
		_, err = h.bot.Send(photo)
		return err
	}

	if replyMsg.Video != nil {
		data, err := download(replyMsg.Video.FileID)
		if err != nil {
			return err
		}
		video := tgbotapi.NewVideo(chatID, tgbotapi.FileBytes{Name: "video.mp4", Bytes: data})
		video.Caption = caption
		_, err = h.bot.Send(video)
		return err
	}

	if replyMsg.Document != nil {
		fileName := replyMsg.Document.FileName
		if fileName == "" {
			fileName = "document"
		}
		data, err := download(replyMsg.Document.FileID)
		if err != nil {
			return err
		}
		doc := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{Name: fileName, Bytes: data})
		doc.Caption = caption
		_, err = h.bot.Send(doc)
		return err
	}

	if replyMsg.Audio != nil {
		fileName := replyMsg.Audio.FileName
		if fileName == "" {
			fileName = "audio.oga"
		}
		data, err := download(replyMsg.Audio.FileID)
		if err != nil {
			return err
		}
		audio := tgbotapi.NewAudio(chatID, tgbotapi.FileBytes{Name: fileName, Bytes: data})
		audio.Caption = caption
		_, err = h.bot.Send(audio)
		return err
	}

	if replyMsg.Voice != nil {
		data, err := download(replyMsg.Voice.FileID)
		if err != nil {
			return err
		}
		voice := tgbotapi.NewVoice(chatID, tgbotapi.FileBytes{Name: "voice.oga", Bytes: data})
		voice.Caption = caption
		_, err = h.bot.Send(voice)
		return err
	}

	if replyMsg.Animation != nil {
		data, err := download(replyMsg.Animation.FileID)
		if err != nil {
			return err
		}
		animation := tgbotapi.NewAnimation(chatID, tgbotapi.FileBytes{Name: "animation.mp4", Bytes: data})
		animation.Caption = caption
		_, err = h.bot.Send(animation)
		return err
	}

	if replyMsg.VideoNote != nil {
		data, err := download(replyMsg.VideoNote.FileID)
		if err != nil {
			return err
		}
		videoNote := tgbotapi.NewVideoNote(chatID, int(replyMsg.VideoNote.Length), tgbotapi.FileBytes{Name: "videonote.mp4", Bytes: data})
		_, err = h.bot.Send(videoNote)
		return err
	}

	return fmt.Errorf("no supported downloadable media found")
}

func (h *TelegramHandler) copyMessageToChat(chatID int64, fromChatID int64, messageID int, caption string) error {
	copyCfg := tgbotapi.NewCopyMessage(chatID, fromChatID, messageID)
	if caption != "" {
		copyCfg.Caption = caption
	}

	_, err := h.bot.CopyMessage(copyCfg)
	if err == nil {
		return nil
	}
	log.Printf("CopyMessage failed for chat %d: %v", chatID, err)

	forwardCfg := tgbotapi.NewForward(chatID, fromChatID, messageID)
	_, err2 := h.bot.Send(forwardCfg)
	if err2 == nil {
		return nil
	}
	log.Printf("Forward failed for chat %d: %v", chatID, err2)
	return fmt.Errorf("copy failed: %w; forward failed: %v", err, err2)
}

// handleEditedMessage handles edited messages
func messagePreview(msg *domain.Message) string {
	if msg == nil {
		return "❓ \\_не найдено\\_"
	}

	preview := ""
	switch msg.MediaType {
	case "photo":
		preview = "📸 Фото"
		if msg.Text != "" {
			preview += fmt.Sprintf(" \\+ `%s`", escapeMarkdownV2(msg.Text))
		}
	case "video":
		preview = "🎬 Видео"
		if msg.Text != "" {
			preview += fmt.Sprintf(" \\+ `%s`", escapeMarkdownV2(msg.Text))
		}
	case "document":
		preview = "📄 Документ"
		if msg.Text != "" {
			preview += fmt.Sprintf(" \\+ `%s`", escapeMarkdownV2(msg.Text))
		}
	case "audio":
		preview = "🎵 Аудио"
	default:
		if msg.Text == "" {
			preview = "💬 \\_пустое сообщение\\_"
		} else {
			preview = fmt.Sprintf("`%s`", escapeMarkdownV2(msg.Text))
		}
	}

	return preview
}

func escapeMarkdownV2(text string) string {
	specialChars := []string{"\\", "_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

func (h *TelegramHandler) handleEditedMessage(ctx context.Context, msg *tgbotapi.Message) error {
	if msg == nil || msg.From == nil || msg.Chat == nil {
		return nil
	}

	log.Printf("Handling edited message from user %d in chat %d", msg.From.ID, msg.Chat.ID)

	existingMsg, err := h.messageService.GetMessageByID(ctx, int64(msg.MessageID))
	oldPreview := messagePreview(nil)
	if err == nil && existingMsg != nil {
		oldPreview = messagePreview(existingMsg)
	}

	newMessage := &domain.Message{
		ChatID:    msg.Chat.ID,
		UserID:    msg.From.ID,
		MessageID: int64(msg.MessageID),
		Text:      msg.Text,
		UpdatedAt: time.Now(),
	}

	if msg.Photo != nil && len(msg.Photo) > 0 {
		newMessage.MediaType = "photo"
		newMessage.MediaFileID = msg.Photo[len(msg.Photo)-1].FileID
	} else if msg.Video != nil {
		newMessage.MediaType = "video"
		newMessage.MediaFileID = msg.Video.FileID
	} else if msg.Document != nil {
		newMessage.MediaType = "document"
		newMessage.MediaFileID = msg.Document.FileID
	}

	err = h.messageService.HandleEditedBusinessMessage(ctx, &domain.Message{ChatID: msg.Chat.ID, UserID: msg.From.ID, MessageID: int64(msg.MessageID)}, newMessage)
	if err != nil {
		log.Printf("Error handling edited message: %v", err)
		return err
	}

	log.Printf("Successfully saved edited message %d", msg.MessageID)

	// Previews are already formatted for MarkdownV2
	escapedOld := oldPreview
	escapedNew := messagePreview(newMessage)

	notification := fmt.Sprintf("✏️ *Сообщение отредактировано*\n\n📌 *Было:*\n%s\n\n✨ *Стало:*\n%s", escapedOld, escapedNew)
	if msg.Chat.ID == h.adminChatID {
		h.notifyAdmin(notification, "MarkdownV2")
	} else {
		log.Printf("Skipping admin notification for subscriber edited message in chat %d", msg.Chat.ID)
	}
	h.notifyChat(msg.Chat.ID, notification)

	reply := tgbotapi.NewMessage(msg.Chat.ID, "✅ *Редактирование сохранено в архивах*")
	reply.ParseMode = "MarkdownV2"
	reply.ReplyToMessageID = msg.MessageID

	_, err = h.bot.Send(reply)
	if err != nil {
		log.Printf("Error sending confirmation: %v", err)
	}

	return nil
}

// handleEditedBusinessMessage handles edited business messages
func (h *TelegramHandler) handleEditedBusinessMessage(ctx context.Context, msg *BusinessMessage) error {
	if msg == nil || msg.From == nil || msg.Chat == nil {
		return nil
	}

	log.Printf("Handling edited business message from user %d in chat %d", msg.From.ID, msg.Chat.ID)

	if msg.BusinessConnectionID != "" {
		err := h.messageService.SaveBusinessSubscriber(ctx, &domain.BusinessSubscriber{
			BusinessConnectionID: msg.BusinessConnectionID,
			ChatID:               msg.Chat.ID,
			Username:             msg.From.UserName,
		})
		if err != nil {
			log.Printf("Error saving business subscriber: %v", err)
		}
	}

	existingMsg, err := h.messageService.GetMessageByID(ctx, int64(msg.MessageID))
	oldPreview := "(не найдено старое сообщение)"
	if err == nil && existingMsg != nil {
		oldPreview = messagePreview(existingMsg)
	}

	newMessage := &domain.Message{
		ChatID:            msg.Chat.ID,
		UserID:            msg.From.ID,
		MessageID:         int64(msg.MessageID),
		BusinessMessageID: msg.BusinessMessageID,
		Text:              msg.Text,
		UpdatedAt:         time.Now(),
	}

	if msg.Photo != nil && len(msg.Photo) > 0 {
		newMessage.MediaType = "photo"
		newMessage.MediaFileID = msg.Photo[len(msg.Photo)-1].FileID
	} else if msg.Video != nil {
		newMessage.MediaType = "video"
		newMessage.MediaFileID = msg.Video.FileID
	} else if msg.Document != nil {
		newMessage.MediaType = "document"
		newMessage.MediaFileID = msg.Document.FileID
	}

	err = h.messageService.HandleEditedBusinessMessage(ctx, &domain.Message{ChatID: msg.Chat.ID, UserID: msg.From.ID, MessageID: int64(msg.MessageID)}, newMessage)
	if err != nil {
		log.Printf("Error handling edited business message: %v", err)
		return err
	}

	log.Printf("Successfully saved edited business message %d", msg.MessageID)

	formatTime := func(t time.Time) string {
		timeStr := t.Format("02.01.2006 • 15:04:05")
		// Escape dots for MarkdownV2
		return strings.ReplaceAll(timeStr, ".", "\\.")
	}

	notification := fmt.Sprintf("╭─ ✏️ Изменение сообщения\n│\n├ 👤 Пользователь: @%s\n├ 🆔 ID: %d\n├ ⏰ Отправлено: %s\n│\n├ 📄 Было:\n│   %s\n│\n├ ✍️ Стало:\n│   %s\n│\n├ 🕒 Изменено: %s\n╰─ 🤖 @ssancodels_bot",
		escapeMarkdownV2(msg.From.UserName), msg.From.ID, formatTime(time.Unix(int64(msg.EditDate), 0)),
		escapeMarkdownV2(oldPreview), messagePreview(newMessage), formatTime(time.Now()))
	if msg.Chat != nil && msg.Chat.ID == h.adminChatID {
		h.notifyAdmin(notification)
	} else {
		log.Printf("Skipping admin notification for subscriber edit in chat %d", msg.Chat.ID)
	}
	if msg.BusinessConnectionID != "" && msg.Chat != nil && msg.Chat.ID != h.adminChatID {
		h.notifyBusinessSubscribers(ctx, msg.BusinessConnectionID, notification)
	} else if msg.BusinessConnectionID != "" {
		log.Printf("Skipping subscriber notification for admin chat edit %d", msg.Chat.ID)
	}
	return nil
}

// handleDeletedBusinessMessages handles deleted business messages
func (h *TelegramHandler) isDuplicateUpdate(updateID int) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.processedUpdateIDs[updateID]; ok {
		return true
	}

	threshold := time.Now().Add(-1 * time.Minute)
	for id, ts := range h.processedUpdateIDs {
		if ts.Before(threshold) {
			delete(h.processedUpdateIDs, id)
		}
	}

	h.processedUpdateIDs[updateID] = time.Now()
	return false
}

func (h *TelegramHandler) handleDeletedBusinessMessages(ctx context.Context, update *DeletedBusinessMessages) error {
	if update == nil {
		return nil
	}

	chatID := int64(0)
	if update.Chat != nil {
		chatID = update.Chat.ID
	}
	log.Printf("Handling deleted business messages from connection %s chat=%d", update.BusinessConnectionID, chatID)

	connectionID := update.BusinessConnectionID
	if connectionID == "" && chatID != 0 {
		subsByChat, err := h.messageService.GetBusinessSubscribersByChatID(ctx, chatID)
		if err != nil {
			log.Printf("Error looking up subscribers by chatID %d: %v", chatID, err)
		} else if len(subsByChat) > 0 {
			connectionID = subsByChat[0].BusinessConnectionID
			log.Printf("handleDeletedBusinessMessages: fallback connectionID from chatID %d = %s", chatID, connectionID)
		}
	}

	deletedLines := make([]string, 0, len(update.MessageIDs))
	for _, messageID := range update.MessageIDs {
		msg, err := h.messageService.GetMessageByID(ctx, messageID)
		if err != nil || msg == nil {
			log.Printf("Deleted message %d not found in DB, skipping notification details", messageID)
			continue
		}

		preview := messagePreview(msg)
		userLine := fmt.Sprintf("🆔 ID: `%d`", msg.UserID)
		userNameLine := ""
		timeStr := "unknown"
		if msg.Username != "" {
			userNameLine = fmt.Sprintf("👤 Пользователь: @%s", escapeMarkdownV2(msg.Username))
		}
		if !msg.CreatedAt.IsZero() {
			timeStr = strings.ReplaceAll(msg.CreatedAt.Format("02.01.2006 • 15:04:05"), ".", "\\.")
		}

		err = h.messageService.HandleDeletedBusinessMessages(ctx, messageID, 0)
		if err != nil {
			log.Printf("Error handling deleted message %d: %v", messageID, err)
			continue
		}

		log.Printf("Successfully processed deleted message %d", messageID)

		nameLine := userNameLine
		if nameLine == "" {
			nameLine = userLine
		}

		block := fmt.Sprintf("├ %s\n├ ⏰ Отправлено: %s\n│\n├ 📄 Было:\n│   %s",
			nameLine, timeStr, preview)
		deletedLines = append(deletedLines, block)
	}

	formatTime := func(t time.Time) string {
		timeStr := t.Format("02.01.2006 • 15:04:05")
		// Escape dots for MarkdownV2
		return strings.ReplaceAll(timeStr, ".", "\\.")
	}

	if len(deletedLines) == 0 {
		notification := fmt.Sprintf("╭─ 🗑️ Удаление сообщения\n│\n├ 📍 Чат: %d\n├ 🕒 Удалено: %s\n╰─ 🤖 @ssancodels_bot", chatID, formatTime(time.Now()))
		if chatID == h.adminChatID {
			h.notifyAdmin(notification)
		} else {
			log.Printf("Skipping admin notification for subscriber deletion in chat %d", chatID)
		}
		log.Printf("handleDeletedBusinessMessages: no messages found in DB, skipping subscriber notification")
	} else {
		deletedListFormatted := stringJoin(deletedLines, "\n\n")
		notification := fmt.Sprintf("╭─ 🗑️ Удаление сообщения\n│\n%s\n│\n├ 🕒 Удалено: %s\n╰─ 🤖 @ssancodels_bot", deletedListFormatted, formatTime(time.Now()))
		if chatID == h.adminChatID {
			h.notifyAdmin(notification)
		} else {
			log.Printf("Skipping admin notification for subscriber deletion in chat %d", chatID)
		}
		if chatID != 0 && chatID != h.adminChatID {
			h.notifyBusinessSubscribersByChatID(ctx, chatID, notification)
		} else if connectionID != "" && chatID != h.adminChatID {
			h.notifyBusinessSubscribers(ctx, connectionID, notification)
		} else {
			log.Printf("handleDeletedBusinessMessages: skipping subscriber notification for admin chat %d or missing connectionID", chatID)
		}
	}
	return nil
}

type joinLinesFunc func([]string) string

func joinLines(lines []string) string {
	return fmt.Sprintf("%s", stringJoin(lines, "\n"))
}

func joinLinesFormatted(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n\n"
		}
		result += fmt.Sprintf("• %s", line)
	}
	timeStr := time.Now().Format("02.01.2006 15:04:05 (MST), UTC+5")
	timeStr = strings.ReplaceAll(timeStr, ".", "\\.")
	result += fmt.Sprintf("\n\n⏰ `%s`", timeStr)
	return result
}

func joinLinesDeletedFormatted(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	result := ""
	for i, line := range lines {
		// Escape content for MarkdownV2
		safeLine := escapeMarkdownV2(line)
		if i == 0 {
			result += fmt.Sprintf("├ 📄 Удаленное сообщение:\n│   %s", safeLine)
		} else {
			result += fmt.Sprintf("\n│\n├ 📄 Удаленное сообщение:\n│   %s", safeLine)
		}
	}
	return result
}

func stringJoin(elements []string, sep string) string {
	if len(elements) == 0 {
		return ""
	}

	result := elements[0]
	for _, element := range elements[1:] {
		result += sep + element
	}
	return result
}

func (h *TelegramHandler) notifyChat(chatID int64, text string, args ...string) {
	if chatID == 0 {
		return
	}

	msg := tgbotapi.NewMessage(chatID, text)
	useMD := len(args) > 0 && args[0] == "MarkdownV2"
	if useMD {
		msg.ParseMode = "MarkdownV2"
	}
	if _, err := h.bot.Send(msg); err != nil {
		if useMD && (strings.Contains(err.Error(), "parse entities") || strings.Contains(err.Error(), "Italic") || strings.Contains(err.Error(), "can't parse entities")) {
			log.Printf("MarkdownV2 parse error, retrying without ParseMode: %v", err)
			msg2 := tgbotapi.NewMessage(chatID, text)
			if _, err2 := h.bot.Send(msg2); err2 != nil {
				log.Printf("Error sending chat notification to %d without ParseMode: %v", chatID, err2)
			}
			return
		}
		log.Printf("Error sending chat notification to %d: %v", chatID, err)
	}
}

func (h *TelegramHandler) notifyBusinessSubscribers(ctx context.Context, connectionID string, text string) {
	log.Printf("notifyBusinessSubscribers: Getting subscribers for connectionID=%s", connectionID)

	subs, err := h.messageService.GetBusinessSubscribersByConnectionID(ctx, connectionID)
	if err != nil {
		log.Printf("Error getting business subscribers: %v", err)
		return
	}

	log.Printf("notifyBusinessSubscribers: Found %d subscribers for connectionID=%s", len(subs), connectionID)

	sent := make(map[int64]bool)
	for i, sub := range subs {
		if sub.ChatID == 0 {
			continue
		}

		if sub.ChatID == h.adminChatID {
			log.Printf("notifyBusinessSubscribers: skipping admin chatID=%d from subscriber list", sub.ChatID)
			continue
		}

		if sent[sub.ChatID] {
			log.Printf("notifyBusinessSubscribers: skipping duplicate subscriber chatID=%d", sub.ChatID)
			continue
		}

		sent[sub.ChatID] = true
		log.Printf("notifyBusinessSubscribers: Sending notification to subscriber %d/%d: chatID=%d, username=%s",
			i+1, len(subs), sub.ChatID, sub.Username)
		h.notifyChat(sub.ChatID, text, "MarkdownV2")
	}
}

func (h *TelegramHandler) notifyBusinessSubscribersByChatID(ctx context.Context, chatID int64, text string) {
	log.Printf("notifyBusinessSubscribersByChatID: Getting subscribers for chatID=%d", chatID)

	subs, err := h.messageService.GetBusinessSubscribersByChatID(ctx, chatID)
	if err != nil {
		log.Printf("Error getting business subscribers by chatID: %v", err)
		return
	}

	log.Printf("notifyBusinessSubscribersByChatID: Found %d subscribers for chatID=%d", len(subs), chatID)

	sent := make(map[int64]bool)
	for i, sub := range subs {
		if sub.ChatID == 0 {
			continue
		}

		if sub.ChatID == h.adminChatID {
			log.Printf("notifyBusinessSubscribersByChatID: skipping admin chatID=%d from subscriber list", sub.ChatID)
			continue
		}

		if sent[sub.ChatID] {
			log.Printf("notifyBusinessSubscribersByChatID: skipping duplicate subscriber chatID=%d", sub.ChatID)
			continue
		}

		sent[sub.ChatID] = true
		log.Printf("notifyBusinessSubscribersByChatID: Sending notification to subscriber %d/%d: chatID=%d, username=%s",
			i+1, len(subs), sub.ChatID, sub.Username)
		h.notifyChat(sub.ChatID, text, "MarkdownV2")
	}
}

func (h *TelegramHandler) notifyAdmin(text string, args ...string) {
	if h.adminChatID == 0 {
		return
	}

	msg := tgbotapi.NewMessage(h.adminChatID, text)
	useMD := len(args) > 0 && args[0] == "MarkdownV2"
	if useMD {
		msg.ParseMode = "MarkdownV2"
	}
	if _, err := h.bot.Send(msg); err != nil {
		if useMD && (strings.Contains(err.Error(), "parse entities") || strings.Contains(err.Error(), "Italic") || strings.Contains(err.Error(), "can't parse entities")) {
			log.Printf("Admin MarkdownV2 parse error, retrying without ParseMode: %v", err)
			msg2 := tgbotapi.NewMessage(h.adminChatID, text)
			if _, err2 := h.bot.Send(msg2); err2 != nil {
				log.Printf("Error sending admin notification without ParseMode: %v", err2)
			}
			return
		}
		log.Printf("Error sending admin notification: %v", err)
	}
}
