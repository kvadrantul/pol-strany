#!/bin/bash

echo "🔍 Проверка деплоя pol-strany..."
echo ""

# Варианты URL для проверки
URLS=(
  "https://pol-strany.vercel.app"
  "https://pol-strany-git-main.vercel.app"
  "https://pol-strany-*.vercel.app"
)

PROJECT_ID="prj_o2q23xamHRnwqnDgQeHkARN6XRai"

echo "📋 Project ID: $PROJECT_ID"
echo ""

echo "🌐 Проверка доступности:"
echo ""

for url in "https://pol-strany.vercel.app"; do
  echo "Проверяю: $url"
  status=$(curl -s -o /dev/null -w "%{http_code}" "$url/" 2>/dev/null)
  if [ "$status" = "200" ]; then
    echo "  ✅ OK (200) - Приложение доступно!"
    echo ""
    echo "📄 Проверка главной страницы:"
    curl -s "$url/" | head -20
    echo ""
    echo ""
    echo "🔌 Проверка API:"
    curl -s "$url/api/tariffs" | head -10
    echo ""
    echo ""
    echo "📦 Проверка статики:"
    curl -s -o /dev/null -w "  app.js: %{http_code}\n" "$url/app.js"
    curl -s -o /dev/null -w "  styles.css: %{http_code}\n" "$url/styles.css"
    curl -s -o /dev/null -w "  index.html: %{http_code}\n" "$url/index.html"
    break
  elif [ "$status" = "404" ]; then
    echo "  ⚠️ 404 - Страница не найдена (возможно деплой еще идет или URL неправильный)"
  else
    echo "  ❌ HTTP $status"
  fi
  echo ""
done

echo ""
echo "📊 Для детальной информации:"
echo "  - Vercel Dashboard: https://vercel.com/dashboard"
echo "  - Project: https://vercel.com/team_igUeUyoVL5L5eyqwTPs4kqVy/pol-strany"
echo ""

