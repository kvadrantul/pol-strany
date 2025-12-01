// Telegram Web App API
let tg = window.Telegram?.WebApp;
if (tg) {
  tg.ready();
  tg.expand();
}

// API базовый URL
const API_URL = window.location.origin;

// Состояние приложения
let currentUser = null;
let tariffs = {};
let selectedTariff = null;
let mapCanvas = null;
let mapCtx = null;
let animationInProgress = false;

// Координаты центра карты (начальная точка)
const CENTER_X = 0.5; // 50% от ширины
const CENTER_Y = 0.5; // 50% от высоты

// Инициализация
document.addEventListener('DOMContentLoaded', async () => {
  await initApp();
  setupEventListeners();
  // Инициализируем карту после небольшой задержки, чтобы DOM был готов
  setTimeout(() => {
    initMap();
  }, 100);
});

// Инициализация приложения
async function initApp() {
  // Получаем данные пользователя из Telegram
  const telegramUser = tg?.initDataUnsafe?.user;
  
  if (telegramUser) {
    const telegramId = telegramUser.id;
    
    // Проверяем, есть ли пользователь в БД
    try {
      const response = await fetch(`${API_URL}/api/user/${telegramId}`);
      if (response.ok) {
        const data = await response.json();
        currentUser = data.user;
        if (!currentUser || currentUser.role !== 'client') {
          // Создаем или обновляем как клиента
          await saveUserRole('client');
        }
      } else {
        // Создаем клиента
        await saveUserRole('client');
      }
    } catch (error) {
      console.error('Ошибка загрузки пользователя:', error);
      await saveUserRole('client');
    }
  } else {
    // Для тестирования
    currentUser = { telegram_id: 123456789, role: 'client' };
  }

  // Загружаем тарифы
  await loadTariffs();
}

// Сохранение роли пользователя
async function saveUserRole(role) {
  const telegramUser = tg?.initDataUnsafe?.user;
  if (!telegramUser) {
    currentUser = { telegram_id: 123456789, role };
    return;
  }

  try {
    const response = await fetch(`${API_URL}/api/user`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        telegram_id: telegramUser.id,
        role: role,
        name: `${telegramUser.first_name} ${telegramUser.last_name || ''}`.trim(),
        avatar_url: telegramUser.photo_url
      })
    });

    if (response.ok) {
      const data = await response.json();
      currentUser = data.user;
    }
  } catch (error) {
    console.error('Ошибка сохранения роли:', error);
  }
}

// Загрузка тарифов
async function loadTariffs() {
  try {
    const response = await fetch(`${API_URL}/api/tariffs`);
    if (response.ok) {
      tariffs = await response.json();
      populateTariffSelect();
    }
  } catch (error) {
    console.error('Ошибка загрузки тарифов:', error);
  }
}

// Заполнение select тарифов
function populateTariffSelect() {
  const select = document.getElementById('tariff-select');
  select.innerHTML = '<option value="">Выберите тариф...</option>';
  
  Object.entries(tariffs).forEach(([key, tariff]) => {
    // Проверяем оба варианта написания
    if (tariff.isAddon || tariff.IsAddon) return; // Пропускаем дополнения
    
    const name = tariff.name || tariff.Name || key;
    const minPrice = tariff.priceRange?.min || tariff.PriceRange?.Min || 0;
    const maxPrice = tariff.priceRange?.max || tariff.PriceRange?.Max || 0;
    
    const option = document.createElement('option');
    option.value = key;
    option.textContent = `${name} - ${minPrice}-${maxPrice} ₽/м²`;
    select.appendChild(option);
  });

  select.addEventListener('change', (e) => {
    if (e.target.value) {
      selectedTariff = e.target.value;
    }
  });
}

// Настройка обработчиков событий
function setupEventListeners() {
  // Поиск бригады
  document.getElementById('search-btn').addEventListener('click', searchContractors);
  
  // Закрытие результатов
  document.getElementById('close-results-btn').addEventListener('click', () => {
    document.getElementById('search-results').classList.add('hidden');
    document.getElementById('search-form').classList.remove('hidden');
  });
}

// Инициализация карты
function initMap() {
  mapCanvas = document.getElementById('map-canvas');
  if (!mapCanvas) return;
  
  mapCtx = mapCanvas.getContext('2d');
  
  // Устанавливаем размер canvas равным размеру контейнера
  function resizeCanvas() {
    const container = mapCanvas.parentElement;
    mapCanvas.width = container.offsetWidth;
    mapCanvas.height = container.offsetHeight;
  }
  
  resizeCanvas();
  window.addEventListener('resize', resizeCanvas);
  
  // Создаем фоновое изображение карты
  createMapBackground();
}

// Создание фонового изображения карты
function createMapBackground() {
  const mapImage = document.getElementById('map-image');
  if (!mapImage) return;
  
  // Создаем реалистичный паттерн карты города
  const canvas = document.createElement('canvas');
  const ctx = canvas.getContext('2d');
  canvas.width = 400;
  canvas.height = 600;
  
  // Темный фон
  ctx.fillStyle = '#1a1a2e';
  ctx.fillRect(0, 0, canvas.width, canvas.height);
  
  // Сетка улиц
  ctx.strokeStyle = 'rgba(255, 255, 255, 0.15)';
  ctx.lineWidth = 2;
  
  // Горизонтальные улицы
  for (let y = 0; y < canvas.height; y += 80) {
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(canvas.width, y);
    ctx.stroke();
  }
  
  // Вертикальные улицы
  for (let x = 0; x < canvas.width; x += 80) {
    ctx.beginPath();
    ctx.moveTo(x, 0);
    ctx.lineTo(x, canvas.height);
    ctx.stroke();
  }
  
  // Здания (квадраты разной высоты)
  ctx.fillStyle = 'rgba(0, 0, 0, 0.4)';
  for (let x = 10; x < canvas.width - 10; x += 90) {
    for (let y = 10; y < canvas.height - 10; y += 90) {
      const height = Math.random() * 40 + 30;
      ctx.fillRect(x, y, 60, height);
    }
  }
  
  // Конвертируем canvas в data URL для фона
  const dataURL = canvas.toDataURL();
  mapImage.style.backgroundImage = `url(${dataURL})`;
  mapImage.style.backgroundSize = 'cover';
  mapImage.style.backgroundPosition = 'center';
}

// Поиск бригад
async function searchContractors() {
  const tariff = document.getElementById('tariff-select').value;
  const area = document.getElementById('area-input').value;
  
  if (!tariff) {
    alert('Пожалуйста, выберите тариф');
    return;
  }

  if (!area || area <= 0) {
    alert('Пожалуйста, укажите площадь');
    return;
  }

  if (animationInProgress) return;
  
  animationInProgress = true;

  try {
    const telegramUser = tg?.initDataUnsafe?.user;
    const telegramId = telegramUser?.id || currentUser?.telegram_id || 123456789;

    // Создаем заказ
    const orderResponse = await fetch(`${API_URL}/api/orders`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        telegram_id: telegramId,
        category: tariff,
        area: parseFloat(area),
        address: null
      })
    });

    if (!orderResponse.ok) {
      throw new Error('Ошибка создания заказа');
    }

    // Ищем бригадиров
    const searchResponse = await fetch(`${API_URL}/api/contractors/search?category=${tariff}`);
    
    if (searchResponse.ok) {
      const data = await searchResponse.json();
      console.log('Search response:', data);
      const contractors = data.contractors || [];
      console.log('Found contractors:', contractors.length);

      if (contractors.length > 0) {
        // Берем первые 5 бригад для анимации (как в требованиях)
        const resultsForAnimation = contractors.slice(0, 5);
        console.log('Animating for', resultsForAnimation.length, 'contractors');
        
        // Запускаем анимацию на карте с 5 бригадами
        await animateMapSearch(resultsForAnimation);
        
        // Показываем все результаты поиска (может быть больше 5)
        displaySearchResults(contractors, tariff, parseFloat(area));
      } else {
        alert('К сожалению, сейчас нет доступных бригад для этого тарифа');
      }
    } else {
      const errorText = await searchResponse.text();
      console.error('Search error:', searchResponse.status, errorText);
      alert('Ошибка поиска бригад: ' + searchResponse.status);
    }
  } catch (error) {
    console.error('Ошибка поиска:', error);
    alert('Произошла ошибка при поиске бригады');
  } finally {
    animationInProgress = false;
  }
}

// Анимация поиска на карте
async function animateMapSearch(contractors) {
  if (!mapCanvas || !mapCtx) {
    console.error('Canvas not initialized');
    return;
  }
  
  console.log('Starting animation for', contractors.length, 'contractors');
  
  // Убеждаемся что canvas правильно размещен
  const container = mapCanvas.parentElement;
  if (!container) {
    console.error('Container not found');
    return;
  }
  
  // Обновляем размеры canvas если нужно
  mapCanvas.width = container.offsetWidth || window.innerWidth;
  mapCanvas.height = container.offsetHeight || window.innerHeight;
  
  const centerX = mapCanvas.width * CENTER_X;
  const centerY = mapCanvas.height * CENTER_Y;
  
  console.log('Canvas size:', mapCanvas.width, mapCanvas.height);
  console.log('Center:', centerX, centerY);
  
  // Очищаем canvas
  mapCtx.clearRect(0, 0, mapCanvas.width, mapCanvas.height);
  
  // Генерируем случайные позиции для бригад (не слишком близко к краям)
  const padding = 50;
  const positions = contractors.map(() => ({
    x: padding + Math.random() * (mapCanvas.width - padding * 2),
    y: padding + Math.random() * (mapCanvas.height - padding * 2)
  }));
  
  console.log('Positions:', positions);
  
  // Создаем маркеры
  const markersContainer = document.getElementById('map-markers');
  if (markersContainer) {
    markersContainer.innerHTML = '';
    
    positions.forEach((pos, index) => {
      const marker = document.createElement('div');
      marker.className = 'map-marker';
      marker.style.left = `${pos.x}px`;
      marker.style.top = `${pos.y}px`;
      markersContainer.appendChild(marker);
    });
  }
  
  // Анимация линий - рисуем одновременно ко всем точкам
  return new Promise((resolve) => {
    let progress = 0;
    const duration = 2000; // 2 секунды
    const startTime = Date.now();
    
    function animate() {
      const elapsed = Date.now() - startTime;
      progress = Math.min(elapsed / duration, 1);
      
      // Очищаем canvas
      mapCtx.clearRect(0, 0, mapCanvas.width, mapCanvas.height);
      
      // Рисуем линии ко всем точкам одновременно
      positions.forEach((pos) => {
        const currentX = centerX + (pos.x - centerX) * progress;
        const currentY = centerY + (pos.y - centerY) * progress;
        
        // Линия
        mapCtx.beginPath();
        mapCtx.moveTo(centerX, centerY);
        mapCtx.lineTo(currentX, currentY);
        mapCtx.strokeStyle = '#09B3AF';
        mapCtx.lineWidth = 3;
        mapCtx.stroke();
        
        // Анимированная точка на конце линии
        mapCtx.beginPath();
        mapCtx.arc(currentX, currentY, 6, 0, Math.PI * 2);
        mapCtx.fillStyle = '#09B3AF';
        mapCtx.fill();
        mapCtx.strokeStyle = '#FFFFFF';
        mapCtx.lineWidth = 2;
        mapCtx.stroke();
      });
      
      if (progress < 1) {
        requestAnimationFrame(animate);
      } else {
        // Анимация завершена - рисуем финальные линии
        positions.forEach((pos) => {
          mapCtx.beginPath();
          mapCtx.moveTo(centerX, centerY);
          mapCtx.lineTo(pos.x, pos.y);
          mapCtx.strokeStyle = '#09B3AF';
          mapCtx.lineWidth = 3;
          mapCtx.stroke();
        });
        console.log('Animation completed');
        resolve();
      }
    }
    
    animate();
  });
}

// Отображение результатов поиска
function displaySearchResults(contractors, tariffKey, area) {
  console.log('Displaying results:', contractors.length, 'contractors');
  console.log('Tariff key:', tariffKey);
  console.log('Tariffs:', tariffs);
  
  const tariff = tariffs[tariffKey];
  console.log('Selected tariff:', tariff);
  
  const resultsList = document.getElementById('results-list');
  if (!resultsList) {
    console.error('Results list element not found');
    return;
  }
  
  resultsList.innerHTML = '';
  
  if (contractors.length === 0) {
    resultsList.innerHTML = '<p>Бригады не найдены</p>';
    return;
  }
  
  contractors.forEach((contractor, index) => {
    console.log(`Contractor ${index}:`, contractor);
    
    // API возвращает поля с заглавной буквы, но можем получить и с маленькой
    const name = contractor.Name || contractor.name || 'Бригада';
    const rating = contractor.Rating !== undefined ? contractor.Rating : (contractor.rating || 0);
    const experience = contractor.ExperienceYears !== undefined ? contractor.ExperienceYears : (contractor.experience_years || 0);
    const orders = contractor.CompletedOrders !== undefined ? contractor.CompletedOrders : (contractor.completed_orders || 0);
    const telegramId = contractor.TelegramID || contractor.telegram_id;
    
    // Получаем цену из тарифа (API возвращает с маленькой буквы)
    let pricePerM2 = 0;
    if (tariff) {
      if (tariff.priceRange) {
        pricePerM2 = (tariff.priceRange.min + tariff.priceRange.max) / 2;
      } else if (tariff.PriceRange) {
        pricePerM2 = (tariff.PriceRange.Min + tariff.PriceRange.Max) / 2;
      }
    }
    const totalPrice = Math.round(pricePerM2 * area);
    const tariffName = tariff ? (tariff.name || tariff.Name || tariffKey) : tariffKey;
    
    const card = document.createElement('div');
    card.className = 'result-card';
    card.innerHTML = `
      <div class="result-card-header">
        <div>
          <div class="result-card-title">${name}</div>
          <div class="result-card-subtitle">${tariffName}</div>
        </div>
      </div>
      <div class="result-card-meta">
        <div class="result-card-rating">
          ⭐ ${rating.toFixed(1)}
        </div>
        <div>Опыт: ${experience || 0} ${experience === 1 ? 'год' : experience < 5 ? 'года' : 'лет'}</div>
        <div>Заказов: ${orders || 0}</div>
      </div>
      <div class="result-card-price">${totalPrice} ₽</div>
    `;
    
    card.addEventListener('click', () => {
      contactContractor({ ...contractor, telegram_id: telegramId });
    });
    
    resultsList.appendChild(card);
  });
  
  // Показываем результаты
  const searchForm = document.getElementById('search-form');
  const searchResults = document.getElementById('search-results');
  
  if (searchForm) searchForm.classList.add('hidden');
  if (searchResults) searchResults.classList.remove('hidden');
  
  console.log('Results displayed');
}

// Контакт с бригадой
function contactContractor(contractor) {
  const telegramId = contractor.telegram_id || contractor.TelegramID;
  
  if (telegramId) {
    if (tg) {
      tg.openTelegramLink(`https://t.me/${telegramId}`);
    } else {
      window.open(`https://t.me/${telegramId}`, '_blank');
    }
  } else {
    alert('Telegram ID бригады не найден');
  }
}
