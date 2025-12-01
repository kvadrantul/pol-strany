# Исправление проблемы с отдачей статики

## Проблема:
404: NOT_FOUND - статические файлы не отдаются

## Причина:
Конфигурация `vercel.json` с `routes` перехватывает все запросы, включая статические файлы.

## Решение:

### 1. Изменить `vercel.json`:

Использовать `rewrites` вместо `routes` - это позволит статике отдаваться автоматически из корня, а API обрабатываться через Go handler.

### 2. Проверить структуру:

Статические файлы должны быть в корне:
- `index.html`
- `app.js`
- `styles.css`
- `images/logo.svg`

### 3. Убедиться, что Handler не перехватывает статику:

Handler в `api/index.go` должен обрабатывать только `/api/*` запросы, а не все запросы.

## Текущая конфигурация:

```json
{
  "version": 2,
  "outputDirectory": ".",
  "builds": [
    {
      "src": "api/index.go",
      "use": "@vercel/go"
    }
  ],
  "rewrites": [
    {
      "source": "/api/(.*)",
      "destination": "/api/index.go"
    }
  ]
}
```

## Как работает:

1. Статические файлы отдаются автоматически из корня (`outputDirectory: "."`)
2. API запросы (`/api/*`) перенаправляются через `rewrites` на Go handler
3. Все остальные запросы обрабатываются как статика

## После изменения:

1. Закоммитьте изменения:
   ```bash
   git add vercel.json
   git commit -m "Исправить конфигурацию для отдачи статики"
   git push
   ```

2. Дождитесь автоматического деплоя

3. Проверьте:
   - `https://your-url.vercel.app/` - главная страница
   - `https://your-url.vercel.app/app.js` - JavaScript
   - `https://your-url.vercel.app/styles.css` - CSS
   - `https://your-url.vercel.app/api/tariffs` - API

