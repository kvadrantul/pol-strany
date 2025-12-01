# Переменные окружения для Vercel

## 📋 Добавьте эти переменные в Vercel Dashboard:

1. Откройте ваш проект в Vercel Dashboard
2. Перейдите в **Settings** → **Environment Variables**
3. Добавьте следующие переменные:

### Обязательные переменные:

```
TELEGRAM_BOT_TOKEN = 8510455686:AAEvlK1P3_yt6btesENX_-d2OMDn5H6T1AQ
```

```
DATABASE_URL = libsql://pol-strany-hun7eee.aws-eu-west-1.turso.io
```

```
TURSO_AUTH_TOKEN = eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJhIjoicnciLCJpYXQiOjE3NjQ2MjA5MzgsImlkIjoiYjIzMDljZmMtZGFhMi00MWViLTkyYTAtNzM3NDJiM2UwNzA4IiwicmlkIjoiYWFiNTY0M2EtOWZiOS00MjQ1LWFhYmQtZDNmNzEzNmIwZDVjIn0.fF8XEVaKVVR4bgJ3BEBtnoU6v3AvnryFS5rKzJ-iity9WYvwHyNxHjjGLGhjMUOo9vIDITN8EW0Z7W5wxrGxDA
```

```
NODE_ENV = production
```

### После деплоя добавьте:

```
APP_URL = https://your-project-name.vercel.app
```

(Замените `your-project-name` на реальное имя вашего проекта)

## ⚠️ Важно:

1. **Environment:** Выберите "Production" для всех переменных
2. После добавления переменных **обязательно передеплойте** проект
3. Переменные применяются только после нового деплоя

## 🚀 После добавления переменных:

```bash
vercel --prod
```

Или через Dashboard: **Deployments** → **Redeploy**

