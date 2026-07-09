# Управление зависимостями

## Инициализация проекта

### Первый запуск

```bash
cd telegram-business-bot

# Загрузить все зависимости
go mod download

# Привести в порядок go.mod и go.sum
go mod tidy
```

## Файлы зависимостей

### `go.mod`
Содержит список всех прямых зависимостей проекта.

**Что означает:**
```
require (
    github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
    // Формат: module_path version
)
```

**Версионирование:**
- `v5.5.1` - точная версия
- `^v5.5.1` - совместимо с 5.5.1 (up to 6.0.0)
- `~v5.5.1` - совместимо с 5.5.x (up to 5.6.0)

### `go.sum`
Контрольные суммы всех зависимостей для безопасности.

**Не редактируйте вручную!** Go управляет автоматически.

## Добавление новой зависимости

### Способ 1: Автоматический

```bash
go get github.com/user/package

# Конкретная версия
go get github.com/user/package@v1.2.3

# Последняя версия
go get github.com/user/package@latest
```

### Способ 2: Ручной (редко)

1. Отредактируйте `go.mod` вручную
2. Выполните `go mod tidy`

## Обновление зависимостей

### Обновить все
```bash
go get -u ./...
```

### Обновить конкретную
```bash
go get -u github.com/go-telegram-bot-api/telegram-bot-api/v5
```

### Проверить обновления
```bash
go list -u -m all
```

## Удаление зависимости

```bash
# Если не используется, просто удалите import
# Затем выполните:
go mod tidy
```

## Проверка целостности

```bash
# Проверить, что все OK
go mod verify

# Загрузить и проверить
go mod tidy
```

## Локальные зависимости (для разработки)

Если вы разрабатываете локальный модуль:

```go
// В go.mod
replace github.com/mymodule => ../mymodule
```

## Вендоринг (если нужно)

```bash
# Скопировать все зависимости в vendor/
go mod vendor

# Использовать вендор (добавить флаг)
go build -mod=vendor
```

## Структура загруженных модулей

Все модули скачиваются в `$GOPATH/pkg/mod/`

На Windows обычно: `C:\Users\<user>\go\pkg\mod\`

## Очистка кеша

```bash
# Если что-то пошло не так
go clean -modcache

# Переулодить зависимости
go mod download
```

## Рекомендации

1. **Всегда фиксируйте версии** в go.mod
2. **Регулярно обновляйте** зависимости
3. **Проверяйте совместимость** перед обновлением
4. **Не удаляйте go.sum** - это критично для безопасности
5. **Добавляйте go.mod и go.sum в git**, но не vendor/

## Основные зависимости проекта

| Пакет | Версия | Назначение |
|-------|--------|-----------|
| go-telegram-bot-api | v5.5.1 | Telegram Bot API |
| pgx | v5.5.5 | PostgreSQL драйвер |
| godotenv | v1.5.1 | Загрузка .env файла |

## Проблемы и решения

### "missing go.sum entry"

```bash
# Решение:
go mod download
go mod verify
go mod tidy
```

### "package version not available"

```bash
# Убедитесь, что версия существует:
go list -m github.com/package@latest

# Попробуйте:
go get github.com/package@latest
```

### "cannot find package"

```bash
# Убедитесь, что module существует
go mod graph | grep package

# Переулодите:
rm go.sum
go mod download
```

## Go Version

Проект требует **Go 1.22+**

Проверка версии:
```bash
go version
```

Установина новой версии:
- Windows: https://golang.org/dl/
- WSL: `wget ...` и распаковка

## Производительность

Если сборка медленная:

```bash
# Параллельная сборка
go build -p 8

# Кэширование
go build -cache
```

## CI/CD

В GitHub Actions:

```yaml
- name: Download modules
  run: go mod download

- name: Verify
  run: go mod verify

- name: Build
  run: go build -v ./...
```

## Безопасность

Проверка уязвимостей (требует Go 1.18+):

```bash
go list -json -m all | nancy sleuth

# Или используйте
govulncheck ./...
```

Установка:
```bash
go install github.com/sonatype-nexus-community/nancy@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

---

Дополнительная информация: [Go Modules](https://golang.org/ref/mod)
