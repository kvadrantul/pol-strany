# Инструкция по деплою на Vercel

## ⚠️ Важно: SQLite не работает на Vercel!

Vercel использует read-only файловую систему, поэтому SQLite не будет работать. Нужно использовать внешнюю базу данных.

## Варианты базы данных для Vercel:

### 1. **Turso** (рекомендуется) - Serverless SQL
- Бесплатный тариф: 500 MB, 1 регион
- Совместим с SQLite (использует libSQL)
- Быстрый и простой в настройке
- Сайт: https://turso.tech

### 2. **PlanetScale** - MySQL serverless
- Бесплатный тариф: 1 БД, 1 GB storage
- Сайт: https://planetscale.com

### 3. **Supabase** - PostgreSQL
- Бесплатный тариф: 500 MB БД
- Сайт: https://supabase.com

### 4. **MongoDB Atlas** - MongoDB
- Бесплатный тариф: 512 MB
- Сайт: https://www.mongodb.com/cloud/atlas

## Шаги деплоя:

### 1. Подготовка базы данных

Выберите один из вариантов выше и создайте БД. Сохраните строку подключения.

### 2. Установка Vercel CLI

```bash
npm i -g vercel
```

### 3. Логин в Vercel

```bash
vercel login
```

### 4. Деплой проекта

```bash
cd pol-strany
vercel
```

Следуйте инструкциям:
- Set up and deploy? **Y**
- Which scope? Выберите ваш аккаунт
- Link to existing project? **N**
- What's your project's name? **pol-strany** (или любое другое)
- In which directory is your code located? **./**

### 5. Настройка переменных окружения

После деплоя перейдите на https://vercel.com/dashboard

1. Выберите ваш проект
2. Settings → Environment Variables
3. Добавьте переменные:

```
TELEGRAM_BOT_TOKEN=your_bot_token_here
DATABASE_URL=your_database_connection_string
APP_URL=https://your-project.vercel.app
NODE_ENV=production
```

**Для Turso:**
```
DATABASE_URL=libsql://your-db-name.turso.io
TURSO_AUTH_TOKEN=your-auth-token
```

**Для PlanetScale:**
```
DATABASE_URL=mysql://user:password@host/database?sslaccept=strict
```

**Для Supabase:**
```
DATABASE_URL=postgresql://user:password@host:5432/database
```

### 6. Обновление кода для работы с внешней БД

Если вы используете **Turso**, нужно обновить `db.js`:

```bash
npm install @libsql/client
```

И обновить код подключения в `db.js`.

### 7. Передеплой

После добавления переменных окружения:

```bash
vercel --prod
```

Или через веб-интерфейс Vercel: Deployments → Redeploy

### 8. Настройка Telegram Mini App

1. Откройте [@BotFather](https://t.me/BotFather)
2. `/myapps` - выберите ваше приложение
3. `/editapp` - выберите приложение
4. Обновите URL на: `https://your-project.vercel.app`

## Структура для Vercel:

```
pol-strany/
├── api/
│   ├── index.js          # Entry point для Vercel
│   └── routes.js         # API роуты
├── public/               # Статические файлы
│   ├── index.html
│   ├── styles.css
│   └── app.js
├── db.js                 # Работа с БД (нужно адаптировать)
├── vercel.json          # Конфигурация Vercel
└── package.json
```

## Что нужно предоставить для деплоя:

1. **Токен Telegram бота** - для переменной `TELEGRAM_BOT_TOKEN`
2. **Строка подключения к БД** - для переменной `DATABASE_URL`
3. **Дополнительные токены** (если нужны, например для Turso)

## Тестирование после деплоя:

1. Откройте `https://your-project.vercel.app` в браузере
2. Проверьте API: `https://your-project.vercel.app/api/tariffs`
3. Откройте бота в Telegram и нажмите кнопку "Открыть приложение"

## Проблемы и решения:

### Ошибка подключения к БД
- Проверьте переменные окружения в Vercel
- Убедитесь, что БД доступна из интернета
- Проверьте firewall настройки БД

### Приложение не открывается
- Проверьте URL в BotFather
- Убедитесь, что приложение задеплоено
- Проверьте логи в Vercel Dashboard

### API возвращает ошибки
- Проверьте логи в Vercel Dashboard (Functions → View Function Logs)
- Убедитесь, что БД инициализирована (таблицы созданы)

