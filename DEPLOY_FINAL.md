# 🚀 Финальная инструкция по деплою

## ✅ У вас есть все данные:

- ✅ Токен бота: `8510455686:AAEvlK1P3_yt6btesENX_-d2OMDn5H6T1AQ`
- ✅ Turso Database URL: `libsql://pol-strany-hun7eee.aws-eu-west-1.turso.io`
- ✅ Turso Auth Token: `eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9...`
- ✅ Токен Vercel: `AMfApuKRnYM2Kwd4UC1GxRVb`

## 📋 Шаги деплоя:

### 1. Проверка подключения к Turso (опционально)

```bash
cd pol-strany
npm install
npm run test-turso
```

Если видите "✅ Все тесты пройдены" - подключение работает!

### 2. Деплой на Vercel

```bash
# Если еще не установлен Vercel CLI
npm i -g vercel

# Войти в Vercel
vercel login

# Деплой
cd pol-strany
vercel
```

**Следуйте инструкциям:**
- Set up and deploy? **Y**
- Which scope? Выберите ваш аккаунт
- Link to existing project? **N**
- What's your project's name? **pol-strany** (или любое другое)
- In which directory is your code located? **./**

### 3. Добавить переменные окружения

После деплоя откройте Vercel Dashboard:

1. Выберите проект **pol-strany**
2. **Settings** → **Environment Variables**
3. Добавьте переменные (см. `VERCEL_ENV_VARS.md`):

```
TELEGRAM_BOT_TOKEN = 8510455686:AAEvlK1P3_yt6btesENX_-d2OMDn5H6T1AQ
DATABASE_URL = libsql://pol-strany-hun7eee.aws-eu-west-1.turso.io
TURSO_AUTH_TOKEN = eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJhIjoicnciLCJpYXQiOjE3NjQ2MjA5MzgsImlkIjoiYjIzMDljZmMtZGFhMi00MWViLTkyYTAtNzM3NDJiM2UwNzA4IiwicmlkIjoiYWFiNTY0M2EtOWZiOS00MjQ1LWFhYmQtZDNmNzEzNmIwZDVjIn0.fF8XEVaKVVR4bgJ3BEBtnoU6v3AvnryFS5rKzJ-iity9WYvwHyNxHjjGLGhjMUOo9vIDITN8EW0Z7W5wxrGxDA
NODE_ENV = production
```

4. **Environment:** Выберите "Production"
5. Нажмите "Save"

### 4. Добавить APP_URL

После первого деплоя вы получите URL проекта (например: `https://pol-strany.vercel.app`)

Добавьте еще одну переменную:

```
APP_URL = https://pol-strany.vercel.app
```

(Замените на ваш реальный URL)

### 5. Передеплой

После добавления переменных **обязательно** передеплойте:

```bash
vercel --prod
```

Или через Dashboard: **Deployments** → выберите последний → **Redeploy**

### 6. Настроить Telegram Mini App

1. Откройте [@BotFather](https://t.me/BotFather)
2. Отправьте `/newapp`
3. Выберите вашего бота
4. Укажите:
   - **Title:** Пол Страны
   - **Description:** Приложение для поиска бригад по стяжке пола
   - **Web App URL:** `https://pol-strany.vercel.app` (ваш URL из Vercel)
   - **Short name:** pol-strany

### 7. Проверка работы

1. Откройте вашего бота в Telegram
2. Отправьте `/start`
3. Нажмите кнопку "Открыть приложение"
4. Приложение должно открыться! 🎉

## 🔍 Проверка API:

Откройте в браузере:
```
https://pol-strany.vercel.app/api/tariffs
```

Должен вернуться JSON с тарифами.

## 🐛 Если что-то не работает:

1. **Проверьте логи:** Vercel Dashboard → Deployments → View Function Logs
2. **Проверьте переменные:** Убедитесь что все переменные добавлены
3. **Проверьте Turso:** Зайдите на turso.tech и убедитесь что БД активна
4. **Передеплойте:** После изменений переменных всегда нужен передеплой

## ✅ Готово!

Ваше приложение должно работать! 

Если возникнут проблемы - проверьте логи в Vercel Dashboard.

