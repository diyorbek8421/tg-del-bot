# Решение проблем при первом запуске

## 🔴 Ошибка: "missing go.sum entry"

```
go: go.sum is out of sync
```

**Решение:**
```bash
go mod tidy
go mod download
```

---

## 🔴 Ошибка: "unknown module"

```
go: no required module provides package ...
```

**Решение:**
```bash
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
go get github.com/jackc/pgx/v5
go get github.com/joho/godotenv
go mod tidy
```

---

## 🔴 Ошибка: "TELEGRAM_BOT_TOKEN is required"

```
Failed to load config: TELEGRAM_BOT_TOKEN is required
```

**Решение:**
1. Скопируйте `.env.example` → `.env`
2. Отредактируйте `.env` и добавьте реальный токен:
```env
TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
```

**Как получить токен:**
1. Откройте Telegram и найдите [@BotFather](https://t.me/botfather)
2. Напишите `/newbot`
3. Следуйте инструкциям и скопируйте токен

---

## 🔴 Ошибка: "failed to create database connection pool"

```
Failed to initialize database: failed to create database connection pool: ...
```

**Решение 1 (Docker):**
```bash
docker run --name postgres-bot \
  -e POSTGRES_PASSWORD=password \
  -d -p 5432:5432 \
  postgres:15-alpine

# Убедитесь что контейнер запущен
docker ps
```

**Решение 2 (Локально на Windows):**
- Скачайте PostgreSQL с [postgresql.org](https://www.postgresql.org/download/windows/)
- Установите
- Проверьте в Services что PostgreSQL запущен
- В `.env` убедитесь что пароль совпадает

---

## 🔴 Ошибка: "connection refused"

```
failed to ping database: cannot connect to database
```

**Это значит:**
- PostgreSQL не запущен
- Неправильные учетные данные в `.env`
- Неправильный хост/порт

**Проверьте:**

**Docker:**
```bash
docker ps | grep postgres
# должен быть контейнер postgres-bot

# Если нет:
docker run --name postgres-bot \
  -e POSTGRES_PASSWORD=password \
  -d -p 5432:5432 \
  postgres:15-alpine
```

**Локально на Windows:**
```powershell
# Проверьте что PostgreSQL запущен
Get-Service postgres*

# Если нет:
Start-Service postgresql-x64-15

# Проверьте подключение
psql -U postgres
# Должна открыться консоль PostgreSQL
```

**Проверьте `.env`:**
```env
DB_HOST=localhost           # ✅ правильно
DB_PORT=5432               # ✅ стандартный порт
DB_USER=postgres           # ✅ дефолтный пользователь
DB_PASSWORD=password       # ✅ пароль при установке

# Если подключаетесь с Docker контейнера:
DB_HOST=postgres           # ❌ НЕправильно (будет работать только в Docker)
DB_HOST=host.docker.internal  # ✅ правильно для доступа из контейнера
# или используйте docker-compose.yml
```

---

## 🔴 Ошибка: "database does not exist"

```
database "telegram_bot" does not exist
```

**Решение:**

**Docker (автоматически):**
```bash
# Просто пересоздайте контейнер
docker rm postgres-bot
docker run --name postgres-bot \
  -e POSTGRES_INITDB_ARGS="-c max_connections=200" \
  -e POSTGRES_DB=telegram_bot \
  -e POSTGRES_PASSWORD=password \
  -d -p 5432:5432 \
  postgres:15-alpine
```

**Локально на Windows:**
```powershell
# Создайте базу вручную
psql -U postgres -c "CREATE DATABASE telegram_bot;"

# Или кроме того
createdb -U postgres telegram_bot

# Проверьте что создалась
psql -U postgres -c "\l"
```

---

## 🔴 Ошибка: "failed to run migrations"

```
Error handling update: failed to run migrations: ...
```

**Решение:**

Это может быть много причин. Посмотрите полный лог ошибки.

**Частые причины:**

1. **Таблица уже существует** - это нормально (если нет ошибки, таблицы созданы)

2. **Permission denied** - пользователь БД не имеет прав:
```sql
-- Подключитесь как postgres
psql -U postgres

-- Дайте права
ALTER ROLE postgres WITH SUPERUSER;
```

3. **Синтаксис SQL** - проверьте `migrations/schema.sql`

---

## 🔴 Ошибка: "Bot token is invalid"

```
failed to create Telegram bot: Unauthorized
```

**Решение:**

1. В `.env` проверьте что токен полностью скопирован
2. Убедитесь что нет пробелов в начале/конце
3. Токен должен содержать `:` (двоеточие)

Правильный формат: `123456:ABCDefghIJKlmnoPQRstUVwxYZ`

Попробуйте заново получить токен:
1. [@BotFather](https://t.me/botfather)
2. `/mybots`
3. Выберите вашего бота
4. `API Token`
5. Скопируйте и вставьте в `.env`

---

## 🔴 Ошибка: "package not found"

```
cannot find package "..." in any of:
```

**Решение:**
```bash
go mod download
go mod tidy
go build ./...
```

Если не поможет:
```bash
# Удалите go.sum и переулодите
rm go.sum
go mod download
go mod tidy
```

---

## 🔴 Ошибка: "no rows in result set"

```
failed to get existing message: message not found
```

**Это нормально:** когда сообщение еще не было сохранено в БД.

Проверьте логи:
- Бот получает обновления от Telegram?
- Сообщение пришло через Business API?

---

## 🔴 Ошибка: "bot not authorized to perform this action"

```
FORBIDDEN: 403 Forbidden
```

**Решение:**

1. Проверьте что бот добавлен в бизнес-чат
2. В Telegram: **Бизнес → Автоматизация → Боты**
3. Убедитесь что дан доступ к Business API

---

## 🟡 Бот не получает обновления

Если бот запущен, но не получает сообщений:

**Решите:**
1. Убедитесь что бот запущен:
```bash
go run cmd/main/main.go
# должны быть логи "Bot started. Waiting for updates..."
```

2. Проверьте что это Business API сообщение:
- Отправляйте через Business чат, а не обычный
- В логах должно быть "Handling new business message"

3. Если это обычное сообщение:
- В логах должно быть "Handling regular message"
- Бот должен ответить эхо

4. Проверьте логи:
```bash
# Смотрите вывод программы
# Должны быть логи формата:
# "Handling ... message from user ..."
# "Successfully saved ..."
```

---

## 📋 Полный чеклист первого запуска

```
[ ] Go 1.22+ установлен: go version
[ ] PostgreSQL запущен: psql -U postgres -c "\l"
[ ] .env файл создан: cp .env.example .env
[ ] Токен бота вставлен: cat .env | grep TELEGRAM_BOT_TOKEN
[ ] Зависимости загружены: go mod download
[ ] Бот запустился: go run cmd/main/main.go
[ ] Видны логи: "Bot started. Waiting for updates..."
[ ] In Business API отправлено тестовое сообщение
[ ] В логах видно: "Handling new business message"
```

---

## 🔧 Docker команды помощи

```bash
# Посмотреть логи контейнера
docker logs postgres-bot
docker logs telegram-business-bot

# Подключиться к БД в контейнере
docker exec -it postgres-bot psql -U postgres -d telegram_bot

# Просмотреть contents контейнера
docker exec -it telegram-business-bot bash

# Перезагрузить контейнер
docker restart postgres-bot

# Полная переустановка
docker rm postgres-bot telegram-business-bot
docker-compose up -d
```

---

## 🔧 PostgreSQL команды помощи

```bash
# Подключение
psql -U postgres

# Список БД
\l

# Подключиться к БД
\c telegram_bot

# Список таблиц
\dt

# Структура таблицы
\d messages

# Число строк в таблице
SELECT COUNT(*) FROM messages;

# Просмотреть данные
SELECT * FROM messages LIMIT 10;

# Выход
\q
```

---

## 🔍 Отладка Go приложения

```bash
# Запустить с дополнительным логированием
GOTRACEBACK=all go run cmd/main/main.go

# С профилированием памяти
go run -run=MemProfile cmd/main/main.go

# С CPU профилированием
go run -cpufreq=... cmd/main/main.go
```

---

## 💾 Сохраненные данные в БД

Проверьте что данные сохраняются:

```bash
# Подключитесь к БД
psql -U postgres -d telegram_bot

# Посмотреть сообщения
SELECT id, message_id, text, created_at FROM messages ORDER BY created_at DESC LIMIT 5;

# Посмотреть редактирования
SELECT * FROM message_edits;

# Посмотреть удаления
SELECT * FROM message_deletions;
```

---

## 🚀 Если все работает

```
✅ Бот запущен
✅ БД подключена
✅ Таблицы созданы
✅ Сообщения сохраняются
✅ Готово к разработке!
```

Что дальше:
1. Прочитайте [docs/QUICKSTART.md](docs/QUICKSTART.md)
2. Изучите [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)
3. Смотрите [docs/API.md](docs/API.md)
4. Начните добавлять функциональность

---

## 🆘 Если ничего не помогло

1. Скопируйте полный текст ошибки
2. Проверьте все шаги выше
3. Посмотрите полный лог:
```bash
go run cmd/main/main.go 2>&1 | tee bot.log
```
4. Откройте проблему в репозитории с логом

**Удачи!** 🚀
