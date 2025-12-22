# Database Migrations Scripts

Скрипты для управления миграциями базы данных всех микросервисов.

## 📋 Доступные сервисы

- `user-service` → `user_service` DB
- `habits-service` → `habits_service` DB
- `bad-habits-service` → `bad_habits_service` DB
- `notification-service` → `notification_service` DB

---

## 🐧 Linux/Mac (Bash)

### Применить все миграции
```bash
./scripts/migrate.sh up
```

### Откатить все миграции
```bash
./scripts/migrate.sh down
```

### Миграции для конкретного сервиса
```bash
# Применить
./scripts/migrate.sh notification up
./scripts/migrate.sh user up

# Откатить
./scripts/migrate.sh notification down
./scripts/migrate.sh user down

# Форсировать версию (для восстановления после ошибок)
./scripts/migrate.sh notification force 1
```

---

## 💻 Windows (PowerShell)

### Применить миграции для одного сервиса
```powershell
.\scripts\migrate.ps1 -Action up -Service notification-service
.\scripts\migrate.ps1 -Action up -Service user-service
```

### Откатить миграции
```powershell
.\scripts\migrate.ps1 -Action down -Service notification-service
```

### Форсировать версию
```powershell
.\scripts\migrate.ps1 -Action force -Service notification-service -Version 1
```

### Проверить текущую версию
```powershell
.\scripts\migrate.ps1 -Action version -Service notification-service
```

---

## 🔧 Переменные окружения

Вы можете настроить подключение к БД через переменные окружения:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_SSL_MODE=disable
```

---

## 📝 Создание новой миграции

### 1. Создайте файлы миграции
```bash
# Формат: NNNN_description.up.sql и NNNN_description.down.sql
# Пример:
touch services/notification-service/migrations/002_add_user_index.up.sql
touch services/notification-service/migrations/002_add_user_index.down.sql
```

### 2. Напишите миграцию

**002_add_user_index.up.sql:**
```sql
CREATE INDEX idx_notifications_user_id_created_at ON notifications(user_id, created_at DESC);
```

**002_add_user_index.down.sql:**
```sql
DROP INDEX IF EXISTS idx_notifications_user_id_created_at;
```

### 3. Примените миграцию
```bash
./scripts/migrate.sh notification up
```

---

## 🐛 Troubleshooting

### "dirty database version" ошибка
Это происходит если миграция была прервана:

```bash
# Проверьте какая версия
./scripts/migrate.sh notification version

# Форсируйте текущую версию
./scripts/migrate.sh notification force 1
```

### "no change" при применении миграций
Все миграции уже применены. Это нормально!

### Проверить таблицы в БД
```bash
docker exec -it habit-tracker-postgres psql -U postgres -d notification_service -c "\dt"
```

---

## 📚 Дополнительная информация

- Используется [golang-migrate](https://github.com/golang-migrate/migrate)
- Миграции должны быть идемпотентными (можно применять несколько раз)
- Всегда используйте `IF EXISTS` / `IF NOT EXISTS`
- Нумерация: 001, 002, 003... (три цифры)

---

## 🎯 Примеры использования

### Полный цикл разработки
```bash
# 1. Создайте новую миграцию
touch services/notification-service/migrations/002_add_feature.up.sql
touch services/notification-service/migrations/002_add_feature.down.sql

# 2. Напишите SQL

# 3. Примените
./scripts/migrate.sh notification up

# 4. Если что-то пошло не так - откатите
./scripts/migrate.sh notification down

# 5. Исправьте и примените снова
./scripts/migrate.sh notification up
```

### Production deployment
```bash
# Всегда проверяйте миграции сначала в dev
./scripts/migrate.sh up

# Затем в production (с правильными env переменными)
DB_HOST=production-db.example.com \
DB_PASSWORD=secure-password \
./scripts/migrate.sh up
```
