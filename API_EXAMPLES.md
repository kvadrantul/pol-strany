# Примеры использования API

## Тестирование через curl или Postman

### 1. Получить список тарифов

```bash
curl http://localhost:3000/api/tariffs
```

### 2. Создать пользователя (клиент)

```bash
curl -X POST http://localhost:3000/api/user \
  -H "Content-Type: application/json" \
  -d '{
    "telegram_id": 123456789,
    "role": "client",
    "name": "Иван Иванов",
    "phone": "+79001234567"
  }'
```

### 3. Создать пользователя (бригадир)

```bash
curl -X POST http://localhost:3000/api/user \
  -H "Content-Type: application/json" \
  -d '{
    "telegram_id": 987654321,
    "role": "contractor",
    "name": "Петр Петров",
    "phone": "+79009876543"
  }'
```

### 4. Обновить профиль бригадира

```bash
curl -X POST http://localhost:3000/api/contractor/profile \
  -H "Content-Type: application/json" \
  -d '{
    "telegram_id": 987654321,
    "experience_years": 5,
    "categories": ["comfort", "premium", "business"],
    "is_active": true
  }'
```

### 5. Создать заказ

```bash
curl -X POST http://localhost:3000/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "telegram_id": 123456789,
    "category": "comfort",
    "area": 25.5,
    "address": "Москва, ул. Ленина, д. 1"
  }'
```

### 6. Поиск бригадиров

```bash
curl "http://localhost:3000/api/contractors/search?category=comfort"
```

### 7. Получить заказы бригадира

```bash
curl http://localhost:3000/api/contractor/orders/987654321
```

### 8. Получить входящие заявки

```bash
curl http://localhost:3000/api/contractor/pending-orders/987654321
```

### 9. Принять заказ

```bash
curl -X POST http://localhost:3000/api/orders/1/accept \
  -H "Content-Type: application/json" \
  -d '{
    "telegram_id": 987654321
  }'
```

### 10. Завершить заказ

```bash
curl -X POST http://localhost:3000/api/orders/1/complete \
  -H "Content-Type: application/json" \
  -d '{
    "telegram_id": 987654321
  }'
```

### 11. Отклонить заказ

```bash
curl -X POST http://localhost:3000/api/orders/1/reject \
  -H "Content-Type: application/json"
```

## Тестирование полного цикла

### Шаг 1: Создать клиента и бригадира

```bash
# Клиент
curl -X POST http://localhost:3000/api/user \
  -H "Content-Type: application/json" \
  -d '{"telegram_id": 111, "role": "client", "name": "Клиент"}'

# Бригадир
curl -X POST http://localhost:3000/api/user \
  -H "Content-Type: application/json" \
  -d '{"telegram_id": 222, "role": "contractor", "name": "Бригадир"}'
```

### Шаг 2: Настроить профиль бригадира

```bash
curl -X POST http://localhost:3000/api/contractor/profile \
  -H "Content-Type: application/json" \
  -d '{
    "telegram_id": 222,
    "experience_years": 3,
    "categories": ["comfort", "econom"],
    "is_active": true
  }'
```

### Шаг 3: Создать заказ от клиента

```bash
curl -X POST http://localhost:3000/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "telegram_id": 111,
    "category": "comfort",
    "area": 30,
    "address": "Тестовый адрес"
  }'
```

### Шаг 4: Бригадир видит заявку

```bash
curl http://localhost:3000/api/contractor/pending-orders/222
```

### Шаг 5: Бригадир принимает заказ

```bash
curl -X POST http://localhost:3000/api/orders/1/accept \
  -H "Content-Type: application/json" \
  -d '{"telegram_id": 222}'
```

### Шаг 6: Бригадир завершает заказ

```bash
curl -X POST http://localhost:3000/api/orders/1/complete \
  -H "Content-Type: application/json" \
  -d '{"telegram_id": 222}'
```

## Проверка результатов

После выполнения команд можно проверить данные в БД:

```bash
sqlite3 data/database.db "SELECT * FROM users;"
sqlite3 data/database.db "SELECT * FROM orders;"
sqlite3 data/database.db "SELECT * FROM contractor_profiles;"
```

