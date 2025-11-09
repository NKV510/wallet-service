# Wallet Service

## Функциональность

- **Создание кошельков** - автоматическое создание при первой операции
- **Операции с балансом** - пополнение (deposit) и списание (withdraw) средств
- **Проверка баланса** - получение текущего состояния счета
- **Валидация операций** - проверка корректности входных данных
- **Конкурентная безопасность** - гарантия целостности данных при параллельных операциях

### Быстрый запуск

```bash
# Клонирование репозитория
git clone https://github.com/NKV510/wallet-service.git
cd wallet-service

# Запуск сервиса
make up
```

Сервис будет доступен по адресу: `http://localhost:8080`

### Ручная сборка

```bash
# Сборка контейнеров
make build

# Запуск сервиса
make up

# Остановка сервиса
make down
```

## API Endpoints

### Операция с кошельком

**POST** `/api/v1/wallet`

```json
{
  "walletId": "uuid-кошелька",
  "operationType": "DEPOSIT|WITHDRAW",
  "amount": 1000
}
```

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/api/v1/wallet \
  -H "Content-Type: application/json" \
  -d '{
    "walletId": "123e4567-e89b-12d3-a456-426614174000",
    "operationType": "DEPOSIT", 
    "amount": 1000
  }'
```

**Ответ:**
```json
{
  "status": "success"
}
```

### Получение баланса

**GET** `/api/v1/wallets/{walletId}`

**Пример запроса:**
```bash
curl http://localhost:8080/api/v1/wallets/123e4567-e89b-12d3-a456-426614174000
```

**Ответ:**
```json
{
  "walletId": "123e4567-e89b-12d3-a456-426614174000",
  "balance": 1000
}
```

## База данных

Сервис использует PostgreSQL с автоматическим применением миграций при запуске.

**Структура таблицы:**
```sql
CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Тестирование

### Интеграционные тесты

```bash
# Запуск интеграционных тестов
make test-integration

# Или вручную
docker compose -f docker-compose.test.yml up -d
go test ./internal/... -v -tags=integration
```

## Структура проекта

```
wallet-service/
├── cmd/
│   └── main.go                 
├── internal/
│   ├── handlers/               
│   ├── repository/            
│   ├── models/                 
│   └── config/      
├── migrations/               
├── docker-compose.yml          
├── docker-compose.test.yml    
└── Makefile                   
```

## Команды Makefile

```bash
make help      # Справка по командам
make build     # Сборка контейнеров
make up        # Запуск сервиса
make down      # Остановка сервиса
make restart   # Перезапуск сервиса
make logs      # Просмотр логов
make ps        # Статус контейнеров
```

## Конфигурация

Настройки по умолчанию:

- **Порт приложения**: 8080
- **PostgreSQL порт**: 5432
- **База данных**: wallet
- **Пользователь**: postgres
- **Пароль**: password

Для изменения настроек отредактируйте `docker-compose.yml`
