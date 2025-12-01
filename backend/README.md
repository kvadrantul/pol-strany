# Go Backend для Пол Страны

Бэкенд на Go для Telegram Mini App "Пол Страны".

## Структура

- `main.go` - точка входа, роутинг, инициализация
- `db.go` - работа с базой данных (Turso/libSQL)
- `handlers.go` - обработчики HTTP запросов

## Локальная разработка

```bash
# Установить зависимости
go mod download

# Запустить сервер
go run main.go db.go handlers.go
```

Или:

```bash
go build -o server .
./server
```

## Переменные окружения

Создайте `.env` файл:

```
DATABASE_URL=libsql://your-database.turso.io
TURSO_AUTH_TOKEN=your-token
PORT=3000
```

## API Endpoints

- `GET /api/tariffs` - получить тарифы
- `GET /api/user/:telegramId` - получить пользователя
- `POST /api/user` - создать/обновить пользователя
- `POST /api/contractor/profile` - обновить профиль бригадира
- `GET /api/contractors/search?category=...` - поиск бригадиров
- `POST /api/orders` - создать заказ
- `GET /api/contractor/orders/:telegramId` - заказы бригадира
- `GET /api/contractor/pending-orders/:telegramId` - входящие заявки
- `POST /api/orders/:orderId/accept` - принять заказ
- `POST /api/orders/:orderId/complete` - завершить заказ
- `POST /api/orders/:orderId/reject` - отклонить заказ

## Деплой на Vercel

Проект уже настроен для деплоя на Vercel. Просто выполните:

```bash
vercel --prod
```

Vercel автоматически определит Go проект и соберет его.

