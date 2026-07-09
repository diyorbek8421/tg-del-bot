package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Storage interface {
	SaveFile(filename string, data io.Reader) (string, error)
	GetFile(filename string) (io.ReadCloser, error)
	DeleteFile(filename string) error
}

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Create base directory if it doesn't exist
	err := os.MkdirAll(basePath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// SaveFile saves file to local storage
func (s *LocalStorage) SaveFile(filename string, data io.Reader) (string, error) {
	filePath := filepath.Join(s.basePath, filename)

	// Create directory if needed
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data to file
	_, err = io.Copy(file, data)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

// GetFile retrieves file from local storage
func (s *LocalStorage) GetFile(filename string) (io.ReadCloser, error) {
	filePath := filepath.Join(s.basePath, filename)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// DeleteFile deletes file from local storage
func (s *LocalStorage) DeleteFile(filename string) error {
	filePath := filepath.Join(s.basePath, filename)

	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
