package storage

import (
	"fmt"
	"io"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MediaDownloader struct {
	botAPI  *tgbotapi.BotAPI
	storage Storage
}

func NewMediaDownloader(botAPI *tgbotapi.BotAPI, storage Storage) *MediaDownloader {
	return &MediaDownloader{
		botAPI:  botAPI,
		storage: storage,
	}
}

// DownloadPhoto downloads photo from Telegram and saves to storage
func (md *MediaDownloader) DownloadPhoto(fileID string, filename string) (string, error) {
	file, err := md.botAPI.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("failed to get photo file info: %w", err)
	}

	url := file.Link(md.botAPI.Token)

	resp, err := tgbotapi.DoHTTP(tgbotapi.ClientConfig{}, url)
	if err != nil {
		return "", fmt.Errorf("failed to download photo: %w", err)
	}
	defer resp.Body.Close()

	path, err := md.storage.SaveFile(filename, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save photo: %w", err)
	}

	return path, nil
}

// DownloadVideo downloads video from Telegram and saves to storage
func (md *MediaDownloader) DownloadVideo(fileID string, filename string) (string, error) {
	file, err := md.botAPI.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("failed to get video file info: %w", err)
	}

	url := file.Link(md.botAPI.Token)

	resp, err := tgbotapi.DoHTTP(tgbotapi.ClientConfig{}, url)
	if err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}
	defer resp.Body.Close()

	path, err := md.storage.SaveFile(filename, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save video: %w", err)
	}

	return path, nil
}

// DownloadDocument downloads document from Telegram and saves to storage
func (md *MediaDownloader) DownloadDocument(fileID string, filename string) (string, error) {
	file, err := md.botAPI.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("failed to get document file info: %w", err)
	}

	url := file.Link(md.botAPI.Token)

	resp, err := tgbotapi.DoHTTP(tgbotapi.ClientConfig{}, url)
	if err != nil {
		return "", fmt.Errorf("failed to download document: %w", err)
	}
	defer resp.Body.Close()

	path, err := md.storage.SaveFile(filename, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save document: %w", err)
	}

	return path, nil
}

// DownloadAudio downloads audio from Telegram and saves to storage
func (md *MediaDownloader) DownloadAudio(fileID string, filename string) (string, error) {
	file, err := md.botAPI.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("failed to get audio file info: %w", err)
	}

	url := file.Link(md.botAPI.Token)

	resp, err := tgbotapi.DoHTTP(tgbotapi.ClientConfig{}, url)
	if err != nil {
		return "", fmt.Errorf("failed to download audio: %w", err)
	}
	defer resp.Body.Close()

	path, err := md.storage.SaveFile(filename, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save audio: %w", err)
	}

	return path, nil
}
