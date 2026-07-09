package main

import (
	"context"
	"log"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-business-bot/internal/storage"
)

// Пример загрузки и сохранения медиа-файлов

func exampleDownloadMedia(botAPI *tgbotapi.BotAPI, localStorage *storage.LocalStorage) {
	ctx := context.Background()

	// Создаем downloader
	downloader := storage.NewMediaDownloader(botAPI, localStorage)

	// Пример загрузки фото
	photoFileID := "your_photo_file_id_here"
	photoPath, err := downloader.DownloadPhoto(photoFileID, filepath.Join("photos", "photo_123.jpg"))
	if err != nil {
		log.Printf("Error downloading photo: %v", err)
	} else {
		log.Printf("Photo saved to: %s", photoPath)
	}

	// Пример загрузки видео
	videoFileID := "your_video_file_id_here"
	videoPath, err := downloader.DownloadVideo(videoFileID, filepath.Join("videos", "video_456.mp4"))
	if err != nil {
		log.Printf("Error downloading video: %v", err)
	} else {
		log.Printf("Video saved to: %s", videoPath)
	}

	// Пример загрузки документа
	docFileID := "your_document_file_id_here"
	docPath, err := downloader.DownloadDocument(docFileID, filepath.Join("documents", "doc_789.pdf"))
	if err != nil {
		log.Printf("Error downloading document: %v", err)
	} else {
		log.Printf("Document saved to: %s", docPath)
	}

	// Пример загрузки аудио
	audioFileID := "your_audio_file_id_here"
	audioPath, err := downloader.DownloadAudio(audioFileID, filepath.Join("audio", "audio_012.mp3"))
	if err != nil {
		log.Printf("Error downloading audio: %v", err)
	} else {
		log.Printf("Audio saved to: %s", audioPath)
	}
}

func exampleReadMedia(localStorage *storage.LocalStorage) {
	ctx := context.Background()

	// Получение файла
	filename := filepath.Join("photos", "photo_123.jpg")
	file, err := localStorage.GetFile(filename)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return
	}
	defer file.Close()

	log.Printf("Successfully opened file: %s", filename)
}

func exampleDeleteMedia(localStorage *storage.LocalStorage) {
	ctx := context.Background()

	filename := filepath.Join("photos", "old_photo.jpg")
	err := localStorage.DeleteFile(filename)
	if err != nil {
		log.Printf("Error deleting file: %v", err)
		return
	}

	log.Printf("Successfully deleted file: %s", filename)
}

func main() {
	// Initialize storage
	localStorage, err := storage.NewLocalStorage("./media")
	if err != nil {
		log.Fatalf("Failed to initialize local storage: %v", err)
	}

	// Initialize Telegram bot (you need to provide a token)
	botAPI, err := tgbotapi.NewBotAPI("YOUR_BOT_TOKEN_HERE")
	if err != nil {
		log.Fatalf("Failed to create bot API: %v", err)
	}

	// Run examples
	log.Println("Example: Download Media")
	exampleDownloadMedia(botAPI, localStorage)

	log.Println("\nExample: Read Media")
	exampleReadMedia(localStorage)

	log.Println("\nExample: Delete Media")
	exampleDeleteMedia(localStorage)
}
