# Habit Tracker Microservices

Микросервисная архитектура для трекера привычек на Go с gRPC и PostgreSQL 17.

## Структура проекта

- **api-gateway** - HTTP REST API Gateway
- **services/user-service** - Управление пользователями и авторизация
- **services/habits-service** - Управление хорошими привычками
- **services/bad-habits-service** - Отслеживание плохих привычек

## Технологии

- **Go 1.21+**
- **PostgreSQL 17** (с UUIDv7)
- **Redis 7** (для сессий и кэша)
- **Kafka 3.6** (для событий между сервисами)
- **gRPC** (для межсервисного взаимодействия)
- **Protocol Buffers v3**

## Быстрый старт

### Вариант 1: Запуск через Docker Compose (Рекомендуется)

```bash
# Запустить всю систему (инфраструктура + сервисы)
make up

# Или вручную
cd deployments
docker-compose -f docker-compose.dev.yml up -d

# Посмотреть логи
make logs

# Остановить все сервисы
make down
```

Это запустит:
- PostgreSQL 18 на порту 5432 (с автоматическим созданием БД)
- Redis на порту 6379
- Kafka на порту 9092
- Kafka UI на http://localhost:8090
- pgAdmin на http://localhost:5050
- User Service на порту 50053
- API Gateway на http://localhost:8080

### Вариант 2: Локальный запуск для разработки

#### 1. Установите зависимости

```bash
# Установите golang-migrate для миграций
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Или через brew (macOS)
brew install golang-migrate
```

#### 2. Запустите только инфраструктуру

```bash
cd deployments
docker-compose -f docker-compose.dev.yml up -d postgres redis kafka
```

#### 3. Примените миграции

```bash
make migrate-up

# Или вручную для конкретного сервиса
./scripts/migrate.sh up user-service
```

#### 4. Запустите сервисы локально

```bash
# User Service
cd services/user-service
go run cmd/main.go

# API Gateway
cd api-gateway
go run cmd/main.go
```

## Makefile команды

```bash
make help          # Показать все доступные команды
make build         # Собрать Docker образы
make up            # Запустить все сервисы
make down          # Остановить все сервисы
make logs          # Показать логи всех сервисов
make logs-user     # Показать логи user-service
make logs-gateway  # Показать логи api-gateway
make restart       # Перезапустить все сервисы
make clean         # Удалить все контейнеры, volumes и образы
make migrate-up    # Применить миграции
make migrate-down  # Откатить миграции
make proto         # Сгенерировать proto файлы
make ps            # Показать запущенные контейнеры
```

## Управление миграциями

### Применить все миграции
```bash
make migrate-up
# или
./scripts/migrate.sh up user-service
```

### Откатить все миграции
```bash
make migrate-down
# или
./scripts/migrate.sh down user-service
```

### Миграции для конкретного сервиса
```bash
./scripts/migrate.sh up user-service      # Применить
./scripts/migrate.sh down user-service    # Откатить
./scripts/migrate.sh force user-service 1 # Установить версию 1
```

## Подключение к БД

### pgAdmin
- URL: http://localhost:5050
- Email: admin@habit-tracker.com
- Password: admin

### Прямое подключение через psql
```bash
psql -h localhost -p 5432 -U postgres -d user_service
```

## Примеры API запросов

### Регистрация
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "password": "securepassword",
    "first_name": "John",
    "timezone": "Europe/Moscow"
  }'
```

### Вход
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email_or_username": "user@example.com",
    "password": "securepassword"
  }'
```

### Получение профиля
```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Выход
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Health Check
```bash
curl http://localhost:8080/health
```

## Конфигурация

Каждый сервис использует `go.uber.org/config` для управления конфигурацией.

Файлы конфигурации: `services/*/config/base.yaml`

Переопределение через переменные окружения:
```bash
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export JWT_SECRET=your-secret-key
export LOG_LEVEL=debug
export REDIS_ADDR=localhost:6379
```

## Разработка

### Генерация proto файлов
```bash
./scripts/generate-proto.sh
```

### Создание новой миграции
```bash
migrate create -ext sql -dir services/user-service/migrations -seq add_new_field
```

## Порты

- API Gateway: 8080
- User Service gRPC: 50053
- Habits Service gRPC: 50051
- Bad Habits Service gRPC: 50052
- PostgreSQL: 5432
- Redis: 6379
- Kafka: 9092
- Kafka UI: 8090
- pgAdmin: 5050
- Prometheus Metrics: 9090

## Архитектура

```
┌─────────────┐
│ API Gateway │ :8080
└──────┬──────┘
       │
       ├──────────┐──────────┐
       │          │          │
   ┌───▼───┐  ┌───▼───┐  ┌───▼────┐
   │ User  │  │Habits │  │  Bad   │
   │Service│  │Service│  │ Habits │
   └───┬───┘  └───┬───┘  └───┬────┘
       │          │          │
   ┌───▼──────────▼──────────▼───┐
   │      PostgreSQL 17 + Redis  │
   └─────────────────────────────┘
```

## Статус реализации

### User Service ✅
- [x] Proto файлы (13 RPC методов)
- [x] Миграции (users, sessions таблицы с UUIDv7)
- [x] Domain entities (User, Session)
- [x] Repository слой (PostgreSQL)
- [x] Redis session storage
- [x] Service слой (UserService, AuthService)
- [x] JWT токены (access + refresh)
- [x] gRPC handlers
- [x] Конфигурация (go.uber.org/config)
- [x] Docker образ
- [x] Docker Compose интеграция

### API Gateway ✅
- [x] HTTP handlers (register, login, logout, profile)
- [x] Middleware (auth, logging, rate limiting)
- [x] Router с маршрутами
- [x] gRPC клиенты
- [x] Конфигурация
- [x] Docker образ
- [x] Docker Compose интеграция

### TODO

#### Habits Service
- [ ] Proto файлы
- [ ] Миграции
- [ ] Repository слой
- [ ] Service слой
- [ ] gRPC handlers

#### Bad Habits Service
- [ ] Proto файлы
- [ ] Миграции
- [ ] Repository слой
- [ ] Service слой
- [ ] gRPC handlers

#### Общее
- [ ] Настроить Prometheus метрики
- [ ] Настроить структурированное логирование (zap/zerolog)
- [ ] Добавить Kubernetes манифесты
- [ ] Настроить CI/CD
- [ ] Добавить unit тесты
- [ ] Добавить integration тесты

## Лицензия

MIT
