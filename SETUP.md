# Инструкция по настройке

## 1. Установка зависимостей

```bash
npm install
```

## 2. Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```env
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
PORT=3000
DB_PATH=./data/database.db
APP_URL=https://your-domain.com
NODE_ENV=development
```

### Как получить токен бота:

1. Откройте Telegram и найдите [@BotFather](https://t.me/BotFather)
2. Отправьте команду `/newbot`
3. Следуйте инструкциям для создания бота
4. Скопируйте полученный токен в `.env`

## 3. Инициализация базы данных

```bash
npm run init-db
```

Это создаст SQLite базу данных со всеми необходимыми таблицами.

## 4. Настройка Telegram Mini App

1. Откройте [@BotFather](https://t.me/BotFather)
2. Отправьте команду `/newapp`
3. Выберите вашего бота
4. Укажите название приложения: "Пол Страны"
5. Укажите описание
6. Укажите URL вашего приложения (например, `https://your-domain.com`)
7. Загрузите иконку (опционально)

## 5. Запуск приложения

### Локальная разработка:

```bash
# Запустить только сервер (для тестирования в браузере)
npm start

# Или с автоперезагрузкой
npm run dev
```

### Запуск с ботом:

В одном терминале:
```bash
npm start
```

В другом терминале:
```bash
npm run start:bot
```

## 6. Тестирование

### Без Telegram (локально):

1. Откройте `http://localhost:3000` в браузере
2. Приложение будет работать, но без интеграции с Telegram API

### В Telegram:

1. Найдите вашего бота в Telegram
2. Отправьте команду `/start`
3. Нажмите кнопку "Открыть приложение"
4. Mini App откроется внутри Telegram

## 7. Деплой

### Vercel:

1. Установите Vercel CLI: `npm i -g vercel`
2. Войдите: `vercel login`
3. Деплой: `vercel`
4. Обновите `APP_URL` в `.env` на URL от Vercel
5. Обновите URL Mini App в BotFather

### Другие платформы:

- **Cloud.ru**: Используйте стандартный Node.js деплой
- **Яндекс.Облако**: Используйте Cloud Functions или VM
- **Heroku**: Используйте стандартный Node.js buildpack

## Структура базы данных

- **users** - Пользователи (клиенты и бригадиры)
- **contractor_profiles** - Профили бригадиров
- **orders** - Заказы
- **reviews** - Отзывы (для будущего использования)

## API Endpoints

- `GET /api/tariffs` - Получить список тарифов
- `GET /api/user/:telegramId` - Получить пользователя
- `POST /api/user` - Создать/обновить пользователя
- `POST /api/contractor/profile` - Обновить профиль бригадира
- `GET /api/contractors/search?category=...` - Поиск бригадиров
- `POST /api/orders` - Создать заказ
- `GET /api/contractor/orders/:telegramId` - Заказы бригадира
- `GET /api/contractor/pending-orders/:telegramId` - Входящие заявки
- `POST /api/orders/:orderId/accept` - Принять заказ
- `POST /api/orders/:orderId/complete` - Завершить заказ
- `POST /api/orders/:orderId/reject` - Отклонить заказ

