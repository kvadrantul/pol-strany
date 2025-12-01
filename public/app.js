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
let currentRole = null;
let tariffs = {};
let selectedTariff = null;

// Инициализация
document.addEventListener('DOMContentLoaded', async () => {
  await initApp();
  setupEventListeners();
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
        currentRole = data.user?.role;
        
        if (currentRole) {
          showRoleScreen(currentRole);
        } else {
          showRoleSelection();
        }
      } else {
        // Пользователь не найден, показываем выбор роли
        showRoleSelection();
      }
    } catch (error) {
      console.error('Ошибка загрузки пользователя:', error);
      showRoleSelection();
    }
  } else {
    // Для тестирования без Telegram
    showRoleSelection();
  }

  // Загружаем тарифы
  await loadTariffs();
}

// Загрузка тарифов
async function loadTariffs() {
  try {
    const response = await fetch(`${API_URL}/api/tariffs`);
    if (response.ok) {
      tariffs = await response.json();
      renderTariffs();
      populateTariffSelect();
    }
  } catch (error) {
    console.error('Ошибка загрузки тарифов:', error);
  }
}

// Отображение тарифов
function renderTariffs() {
  const container = document.getElementById('tariffs-list');
  container.innerHTML = '';

  Object.entries(tariffs).forEach(([key, tariff]) => {
    const card = document.createElement('div');
    card.className = 'tariff-card';
    card.dataset.tariff = key;
    card.innerHTML = `
      <div class="tariff-name">${tariff.name}</div>
      <div class="tariff-desc">${tariff.description}</div>
      <div class="tariff-price">${tariff.priceRange.min} - ${tariff.priceRange.max} ₽/м²</div>
    `;
    card.addEventListener('click', () => selectTariff(key));
    container.appendChild(card);
  });
}

// Выбор тарифа
function selectTariff(key) {
  selectedTariff = key;
  document.querySelectorAll('.tariff-card').forEach(card => {
    card.classList.remove('selected');
  });
  document.querySelector(`[data-tariff="${key}"]`).classList.add('selected');
  document.getElementById('tariff-select').value = key;
}

// Заполнение select тарифов
function populateTariffSelect() {
  const select = document.getElementById('tariff-select');
  select.innerHTML = '<option value="">Выберите тариф...</option>';
  
  Object.entries(tariffs).forEach(([key, tariff]) => {
    const option = document.createElement('option');
    option.value = key;
    option.textContent = `${tariff.name} - ${tariff.priceRange.min}-${tariff.priceRange.max} ₽/м²`;
    select.appendChild(option);
  });

  select.addEventListener('change', (e) => {
    if (e.target.value) {
      selectTariff(e.target.value);
    }
  });
}

// Показ экрана выбора роли
function showRoleSelection() {
  hideAllScreens();
  document.getElementById('role-selection').classList.add('active');
}

// Показ экрана роли
function showRoleScreen(role) {
  hideAllScreens();
  currentRole = role;
  
  if (role === 'client') {
    document.getElementById('client-screen').classList.add('active');
    loadClientData();
  } else if (role === 'contractor') {
    document.getElementById('contractor-screen').classList.add('active');
    loadContractorData();
  }
}

// Скрыть все экраны
function hideAllScreens() {
  document.querySelectorAll('.screen').forEach(screen => {
    screen.classList.remove('active');
  });
}

// Настройка обработчиков событий
function setupEventListeners() {
  // Выбор роли
  document.querySelectorAll('.role-btn').forEach(btn => {
    btn.addEventListener('click', async (e) => {
      const role = e.currentTarget.dataset.role;
      await saveUserRole(role);
      showRoleScreen(role);
    });
  });

  // Поиск бригады
  document.getElementById('search-btn').addEventListener('click', searchContractor);

  // Просмотр профиля
  document.getElementById('view-profile-btn').addEventListener('click', showContractorProfile);
  document.getElementById('close-profile-modal').addEventListener('click', () => {
    document.getElementById('contractor-profile-modal').classList.add('hidden');
  });

  // Контакт с бригадиром
  document.getElementById('contact-btn').addEventListener('click', contactContractor);

  // Бригадир: редактирование профиля
  document.getElementById('edit-profile-btn').addEventListener('click', () => {
    document.getElementById('contractor-profile-section').classList.add('hidden');
    document.getElementById('profile-form').classList.remove('hidden');
    loadProfileForm();
  });

  document.getElementById('save-profile-btn').addEventListener('click', saveContractorProfile);
  document.getElementById('cancel-profile-btn').addEventListener('click', () => {
    document.getElementById('contractor-profile-section').classList.remove('hidden');
    document.getElementById('profile-form').classList.add('hidden');
  });

  // Меню
  document.getElementById('client-menu-btn').addEventListener('click', () => {
    document.getElementById('menu-modal').classList.remove('hidden');
  });

  document.getElementById('contractor-menu-btn').addEventListener('click', () => {
    document.getElementById('menu-modal').classList.remove('hidden');
  });

  document.getElementById('close-menu-modal').addEventListener('click', () => {
    document.getElementById('menu-modal').classList.add('hidden');
  });

  document.getElementById('change-role-btn').addEventListener('click', () => {
    document.getElementById('menu-modal').classList.add('hidden');
    showRoleSelection();
  });
}

// Сохранение роли пользователя
async function saveUserRole(role) {
  const telegramUser = tg?.initDataUnsafe?.user;
  if (!telegramUser) {
    // Для тестирования
    currentUser = { telegram_id: 123456789, role };
    currentRole = role;
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

// Поиск бригадира
async function searchContractor() {
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

  // Показываем анимацию поиска
  document.getElementById('search-animation').classList.remove('hidden');
  document.getElementById('found-contractor').classList.add('hidden');

  // Создаем заказ
  const address = document.getElementById('address-input').value;
  
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
        address: address || null
      })
    });

    if (!orderResponse.ok) {
      throw new Error('Ошибка создания заказа');
    }

    // Ищем бригадиров
    await new Promise(resolve => setTimeout(resolve, 2000)); // Имитация поиска

    const searchResponse = await fetch(`${API_URL}/api/contractors/search?category=${tariff}`);
    
    if (searchResponse.ok) {
      const data = await searchResponse.json();
      const contractors = data.contractors;

      if (contractors && contractors.length > 0) {
        // Берем первого доступного бригадира
        const contractor = contractors[0];
        displayFoundContractor(contractor);
      } else {
        alert('К сожалению, сейчас нет доступных бригадиров для этого тарифа');
        document.getElementById('search-animation').classList.add('hidden');
      }
    }
  } catch (error) {
    console.error('Ошибка поиска:', error);
    alert('Произошла ошибка при поиске бригады');
    document.getElementById('search-animation').classList.add('hidden');
  }
}

// Отображение найденного бригадира
function displayFoundContractor(contractor) {
  document.getElementById('search-animation').classList.add('hidden');
  document.getElementById('found-contractor').classList.remove('hidden');

  document.getElementById('contractor-name').textContent = contractor.name || 'Бригадир';
  document.getElementById('contractor-rating').textContent = `⭐ ${contractor.rating || 0.0}`;
  document.getElementById('contractor-experience').textContent = `Опыт: ${contractor.experience_years || 0} лет`;
  
  if (contractor.avatar_url) {
    document.getElementById('contractor-avatar').src = contractor.avatar_url;
  } else {
    document.getElementById('contractor-avatar').src = 'https://via.placeholder.com/80';
  }

  // Сохраняем данные бригадира для просмотра профиля
  document.getElementById('found-contractor').dataset.contractorId = contractor.user_id;
  document.getElementById('found-contractor').dataset.contractorData = JSON.stringify(contractor);
}

// Просмотр профиля бригадира
function showContractorProfile() {
  const contractorData = JSON.parse(
    document.getElementById('found-contractor').dataset.contractorData || '{}'
  );

  const details = document.getElementById('profile-details');
  details.innerHTML = `
    <div class="profile-detail">
      <strong>Имя:</strong> ${contractorData.name || 'Не указано'}
    </div>
    <div class="profile-detail">
      <strong>Телефон:</strong> ${contractorData.phone || 'Не указан'}
    </div>
    <div class="profile-detail">
      <strong>Опыт работы:</strong> ${contractorData.experience_years || 0} лет
    </div>
    <div class="profile-detail">
      <strong>Рейтинг:</strong> ⭐ ${contractorData.rating || 0.0}
    </div>
    <div class="profile-detail">
      <strong>Завершено заказов:</strong> ${contractorData.completed_orders || 0}
    </div>
    <div class="profile-detail">
      <strong>Категории:</strong> ${(contractorData.categories || []).join(', ') || 'Не указаны'}
    </div>
  `;

  document.getElementById('contractor-profile-modal').classList.remove('hidden');
}

// Контакт с бригадиром через Telegram
function contactContractor() {
  const contractorData = JSON.parse(
    document.getElementById('found-contractor').dataset.contractorData || '{}'
  );

  if (contractorData.telegram_id) {
    // Открываем чат в Telegram
    if (tg) {
      tg.openTelegramLink(`https://t.me/${contractorData.telegram_id}`);
    } else {
      window.open(`https://t.me/${contractorData.telegram_id}`, '_blank');
    }
  } else {
    alert('Telegram ID бригадира не найден');
  }
}

// Загрузка данных клиента
async function loadClientData() {
  // Данные уже загружены при инициализации
}

// Загрузка данных бригадира
async function loadContractorData() {
  const telegramUser = tg?.initDataUnsafe?.user;
  const telegramId = telegramUser?.id || currentUser?.telegram_id || 123456789;

  try {
    const response = await fetch(`${API_URL}/api/user/${telegramId}`);
    if (response.ok) {
      const data = await response.json();
      if (data.profile) {
        displayContractorProfile(data.profile);
      }
    }

    // Загружаем заказы
    await loadContractorOrders(telegramId);
  } catch (error) {
    console.error('Ошибка загрузки данных бригадира:', error);
  }
}

// Отображение профиля бригадира
function displayContractorProfile(profile) {
  const statusBadge = document.getElementById('profile-status');
  if (profile.experience_years && profile.categories) {
    statusBadge.textContent = 'Профиль заполнен';
    statusBadge.className = 'status-badge completed';
  }
}

// Загрузка формы профиля
function loadProfileForm() {
  const telegramUser = tg?.initDataUnsafe?.user;
  const name = telegramUser ? `${telegramUser.first_name} ${telegramUser.last_name || ''}`.trim() : '';
  
  document.getElementById('profile-name').value = name;
  
  // Загружаем существующий профиль
  loadContractorData().then(() => {
    // Заполняем форму если есть данные
  });

  // Создаем чекбоксы для категорий
  const checkboxesContainer = document.getElementById('categories-checkboxes');
  checkboxesContainer.innerHTML = '';

  Object.entries(tariffs).forEach(([key, tariff]) => {
    const label = document.createElement('label');
    label.className = 'checkbox-item';
    label.innerHTML = `
      <input type="checkbox" value="${key}" id="cat-${key}">
      <span>${tariff.name}</span>
    `;
    checkboxesContainer.appendChild(label);
  });
}

// Сохранение профиля бригадира
async function saveContractorProfile() {
  const name = document.getElementById('profile-name').value;
  const phone = document.getElementById('profile-phone').value;
  const experience = parseInt(document.getElementById('profile-experience').value) || 0;
  
  const checkboxes = document.querySelectorAll('#categories-checkboxes input[type="checkbox"]:checked');
  const categories = Array.from(checkboxes).map(cb => cb.value);

  if (!name) {
    alert('Пожалуйста, укажите имя');
    return;
  }

  if (categories.length === 0) {
    alert('Пожалуйста, выберите хотя бы одну категорию');
    return;
  }

  try {
    const telegramUser = tg?.initDataUnsafe?.user;
    const telegramId = telegramUser?.id || currentUser?.telegram_id || 123456789;

    // Обновляем пользователя
    await fetch(`${API_URL}/api/user`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        telegram_id: telegramId,
        role: 'contractor',
        name: name,
        phone: phone
      })
    });

    // Сохраняем профиль
    const response = await fetch(`${API_URL}/api/contractor/profile`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        telegram_id: telegramId,
        experience_years: experience,
        categories: categories,
        is_active: true
      })
    });

    if (response.ok) {
      alert('Профиль успешно сохранен!');
      document.getElementById('contractor-profile-section').classList.remove('hidden');
      document.getElementById('profile-form').classList.add('hidden');
      await loadContractorData();
    } else {
      throw new Error('Ошибка сохранения профиля');
    }
  } catch (error) {
    console.error('Ошибка сохранения профиля:', error);
    alert('Произошла ошибка при сохранении профиля');
  }
}

// Загрузка заказов бригадира
async function loadContractorOrders(telegramId) {
  try {
    const response = await fetch(`${API_URL}/api/contractor/orders/${telegramId}`);
    if (response.ok) {
      const data = await response.json();
      renderOrders(data.orders);
    }

    // Загружаем входящие заявки
    await loadPendingOrders(telegramId);
  } catch (error) {
    console.error('Ошибка загрузки заказов:', error);
  }
}

// Загрузка входящих заявок
async function loadPendingOrders(telegramId) {
  try {
    // Упрощенная версия - получаем все pending заказы
    const response = await fetch(`${API_URL}/api/contractor/pending-orders/${telegramId}`);
    if (response.ok) {
      const data = await response.json();
      renderPendingOrders(data.orders || []);
    }
  } catch (error) {
    console.error('Ошибка загрузки заявок:', error);
  }
}

// Отображение заказов
function renderOrders(orders) {
  const activeList = document.getElementById('active-orders-list');
  const completedList = document.getElementById('completed-orders-list');

  activeList.innerHTML = '';
  completedList.innerHTML = '';

  orders.forEach(order => {
    const card = createOrderCard(order);
    
    if (order.status === 'accepted' || order.status === 'in_progress') {
      activeList.appendChild(card);
    } else if (order.status === 'completed') {
      completedList.appendChild(card);
    }
  });

  if (activeList.innerHTML === '') {
    activeList.innerHTML = '<p style="color: #666; text-align: center;">Нет активных заказов</p>';
  }

  if (completedList.innerHTML === '') {
    completedList.innerHTML = '<p style="color: #666; text-align: center;">Нет завершенных заказов</p>';
  }
}

// Отображение входящих заявок
function renderPendingOrders(orders) {
  const list = document.getElementById('pending-orders-list');
  list.innerHTML = '';

  if (orders.length === 0) {
    list.innerHTML = '<p style="color: #666; text-align: center;">Нет входящих заявок</p>';
    return;
  }

  orders.forEach(order => {
    const card = createOrderCard(order, true);
    list.appendChild(card);
  });
}

// Создание карточки заказа
function createOrderCard(order, isPending = false) {
  const card = document.createElement('div');
  card.className = 'order-card';
  
  const tariff = tariffs[order.category] || { name: order.category };
  const statusLabels = {
    pending: 'Ожидает',
    accepted: 'Принят',
    in_progress: 'В работе',
    completed: 'Завершен',
    cancelled: 'Отменен'
  };

  card.innerHTML = `
    <div class="order-header">
      <div class="order-title">${tariff.name}</div>
      <div class="order-status ${order.status}">${statusLabels[order.status] || order.status}</div>
    </div>
    <div class="order-details">
      ${order.area ? `<div>Площадь: ${order.area} м²</div>` : ''}
      ${order.address ? `<div>Адрес: ${order.address}</div>` : ''}
      ${order.client_name ? `<div>Клиент: ${order.client_name}</div>` : ''}
      <div>Создан: ${new Date(order.created_at).toLocaleString('ru-RU')}</div>
    </div>
    ${isPending ? `
      <div class="order-actions">
        <button class="btn-accept" onclick="acceptOrder(${order.id})">✓ Принять</button>
        <button class="btn-reject" onclick="rejectOrder(${order.id})">✗ Отклонить</button>
      </div>
    ` : order.status === 'accepted' || order.status === 'in_progress' ? `
      <div class="order-actions">
        <button class="btn-complete" onclick="completeOrder(${order.id})">✓ Завершить</button>
      </div>
    ` : ''}
  `;

  return card;
}

// Принять заказ
async function acceptOrder(orderId) {
  try {
    const telegramUser = tg?.initDataUnsafe?.user;
    const telegramId = telegramUser?.id || currentUser?.telegram_id || 123456789;

    const response = await fetch(`${API_URL}/api/orders/${orderId}/accept`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ telegram_id: telegramId })
    });

    if (response.ok) {
      alert('Заказ принят!');
      await loadContractorData();
    } else {
      throw new Error('Ошибка принятия заказа');
    }
  } catch (error) {
    console.error('Ошибка принятия заказа:', error);
    alert('Произошла ошибка при принятии заказа');
  }
}

// Отклонить заказ
async function rejectOrder(orderId) {
  if (!confirm('Вы уверены, что хотите отклонить этот заказ?')) {
    return;
  }

  try {
    const response = await fetch(`${API_URL}/api/orders/${orderId}/reject`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    });

    if (response.ok) {
      alert('Заказ отклонен');
      await loadContractorData();
    } else {
      throw new Error('Ошибка отклонения заказа');
    }
  } catch (error) {
    console.error('Ошибка отклонения заказа:', error);
    alert('Произошла ошибка при отклонении заказа');
  }
}

// Завершить заказ
async function completeOrder(orderId) {
  if (!confirm('Завершить этот заказ?')) {
    return;
  }

  try {
    const telegramUser = tg?.initDataUnsafe?.user;
    const telegramId = telegramUser?.id || currentUser?.telegram_id || 123456789;

    const response = await fetch(`${API_URL}/api/orders/${orderId}/complete`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ telegram_id: telegramId })
    });

    if (response.ok) {
      alert('Заказ завершен!');
      await loadContractorData();
    } else {
      throw new Error('Ошибка завершения заказа');
    }
  } catch (error) {
    console.error('Ошибка завершения заказа:', error);
    alert('Произошла ошибка при завершении заказа');
  }
}

// Делаем функции доступными глобально для onclick
window.acceptOrder = acceptOrder;
window.rejectOrder = rejectOrder;
window.completeOrder = completeOrder;

