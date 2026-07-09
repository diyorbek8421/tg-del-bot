# Makefile для telegram-business-bot

.PHONY: help setup run run-docker stop deps tidy fmt lint test migrate clean

help:
	@echo "Доступные команды:"
	@echo "  make setup         - Инициализация проекта"
	@echo "  make deps          - Загрузка зависимостей"
	@echo "  make tidy          - Очистка зависимостей"
	@echo "  make run           - Запуск бота локально"
	@echo "  make run-docker    - Запуск с Docker Compose"
	@echo "  make stop          - Остановка Docker контейнеров"
	@echo "  make fmt           - Форматирование кода"
	@echo "  make lint          - Проверка кода"
	@echo "  make test          - Запуск тестов"
	@echo "  make clean         - Очистка артефактов"

setup:
	cp .env.example .env
	@echo "✅ Проект инициализирован. Отредактируйте .env"

deps:
	go mod download
	go mod tidy

tidy:
	go mod tidy

run:
	go run cmd/main/main.go

run-docker:
	docker-compose up -d

stop:
	docker-compose down

logs:
	docker-compose logs -f bot

fmt:
	go fmt ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

test:
	go test -v ./...

clean:
	rm -f bot
	docker-compose down -v
	find . -name "*.log" -delete
	find . -name ".DS_Store" -delete

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bot cmd/main/main.go

.DEFAULT_GOAL := help
