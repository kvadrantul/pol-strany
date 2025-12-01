require('dotenv').config();
const express = require('express');
const path = require('path');
const bodyParser = require('body-parser');
const cors = require('cors');
const Database = require('./db');

const app = express();
const PORT = process.env.PORT || 3000;

// Для локальной разработки используем SQLite
const db = new Database(process.env.DB_PATH || './data/database.db');

// Middleware
app.use(cors());
app.use(bodyParser.json());
app.use(express.static('public'));

// Тарифы
const TARIFFS = {
  'econom': {
    name: 'ЭКОНОМ',
    description: 'Мокрая, ручная',
    priceRange: { min: 400, max: 450 },
    days: 28,
    features: ['Классика', 'Низкая цена материалов', 'Долгий срок высыхания', 'Высокий риск трещин']
  },
  'comfort': {
    name: 'КОМФОРТ',
    description: 'Полусухая механизированная',
    priceRange: { min: 550, max: 850 },
    days: '5-7 дней (плитка — 2 дня, ламинат — 14–20 дней)',
    features: ['Оптимальный баланс', 'Минимум усадки', 'Можно ходить через 12 часов', 'Самый популярный выбор']
  },
  'business': {
    name: 'БИЗНЕС',
    description: 'С армированием',
    priceRange: { min: 150, max: 300 },
    days: 'Как у базового тарифа',
    features: ['Повышенная прочность', 'Надбавка за армирование сеткой или фиброй'],
    isAddon: true
  },
  'premium': {
    name: 'ПРЕМИУМ',
    description: 'Сухая стяжка Кнауф',
    priceRange: { min: 800, max: 1000 },
    days: '1-2 дня',
    features: ['Нет мокрых процессов', 'Идеальная геометрия', 'Теплоизоляция', 'Высокая цена материалов']
  },
  'universal': {
    name: 'УНИВЕРСАЛ',
    description: 'Плавающая / Утепленная',
    priceRange: { min: 250, max: 600 },
    days: 'Как у базового тарифа',
    features: ['Зависит от вида утеплителя', 'Включает слой изоляции'],
    isAddon: true
  },
  'self-leveling': {
    name: 'САМОВЫРАВНИВАТЕЛЬ',
    description: 'Финишный слой',
    priceRange: { min: 250, max: 500 },
    days: '1-3 дня',
    features: ['Финишный слой']
  }
};

// API Routes

// Получить тарифы
app.get('/api/tariffs', (req, res) => {
  res.json(TARIFFS);
});

// Получить информацию о пользователе
app.get('/api/user/:telegramId', async (req, res) => {
  try {
    const user = await db.getUserByTelegramId(parseInt(req.params.telegramId));
    if (!user) {
      return res.status(404).json({ error: 'Пользователь не найден' });
    }

    let profile = null;
    if (user.role === 'contractor') {
      profile = await db.getContractorProfile(user.id);
    }

    res.json({ user, profile });
  } catch (error) {
    console.error('Ошибка получения пользователя:', error);
    res.status(500).json({ error: 'Внутренняя ошибка сервера' });
  }
});

// Создать или обновить пользователя
app.post('/api/user', async (req, res) => {
  try {
    const { telegram_id, role, name, phone, avatar_url } = req.body;
    
    let user = await db.getUserByTelegramId(telegram_id);
    if (!user) {
      const userId = await db.createUser(telegram_id, role, name, phone, avatar_url);
      user = await db.getUserByTelegramId(telegram_id);
    } else {
      await db.updateUser(telegram_id, { role, name, phone, avatar_url });
      user = await db.getUserByTelegramId(telegram_id);
    }

    res.json({ user });
  } catch (error) {
    console.error('Ошибка создания пользователя:', error);
    res.status(500).json({ error: 'Внутренняя ошибка сервера' });
  }
});

// Обновить профиль бригадира
app.post('/api/contractor/profile', async (req, res) => {
  try {
    const { telegram_id, experience_years, categories, is_active } = req.body;
    
    const user = await db.getUserByTelegramId(telegram_id);
    if (!user || user.role !== 'contractor') {
      return res.status(400).json({ error: 'Пользователь не является бригадиром' });
    }

    await db.createOrUpdateContractorProfile(user.id, {
      experience_years,
      categories,
      is_active
    });

    const profile = await db.getContractorProfile(user.id);
    res.json({ profile });
  } catch (error) {
    console.error('Ошибка обновления профиля:', error);
    res.status(500).json({ error: 'Внутренняя ошибка сервера' });
  }
});

// Поиск доступных бригадиров
app.get('/api/contractors/search', async (req, res) => {
  try {
    const { category } = req.query;
    if (!category) {
      return res.status(400).json({ error: 'Не указана категория' });
    }

    const contractors = await db.getAvailableContractors(category);
    res.json({ contractors });
  } catch (error) {
    console.error('Ошибка поиска бригадиров:', error);
    res.status(500).json({ error: 'Внутренняя ошибка сервера' });
  }
});

// Создать заказ
app.post('/api/orders', async (req, res) => {
  try {
    const { telegram_id, category, area, address } = req.body;
    
    const user = await db.getUserByTelegramId(telegram_id);
    if (!user || user.role !== 'client') {
      return res.status(400).json({ error: 'Пользователь не является клиентом' });
    }

    const orderId = await db.createOrder(user.id, category, area, address);
    const order = await db.getOrder(orderId);
    
    res.json({ order });
  } catch (error) {
    console.error('Ошибка создания заказа:', error);
    res.status(500).json({ error: 'Внутренняя ошибка сервера' });
  }
});

// Получить заказы бригадира
app.get('/api/contractor/orders/:telegramId', async (req, res) => {
  try {
    const user = await db.getUserByTelegramId(parseInt(req.params.telegramId));
    if (!user || user.role !== 'contractor') {
      return res.status(400).json({ error: 'Пользователь не является бригадиром' });
    }

    const orders = await db.getContractorOrders(user.id);
    res.json({ orders });
  } catch (error) {
    console.error('Ошибка получения заказов:', error);
    res.status(500).json({ error: 'Внутренняя ошибка сервера' });
  }
});

// Получить входящие заявки для бригадира
app.get('/api/contractor/pending-orders/:telegramId', async (req, res) => {
  try {
    const user = await db.getUserByTelegramId(parseInt(req.params.telegramId));
    if (!user || user.role !== 'contractor') {
      return res.status(400).json({ error: 'Пользователь не является бригадиром' });
    }

    // Получаем все pending заказы
    const orders = await db.getAllPendingOrders();
    res.json({ orders });
  } catch (error) {
    console.error('Ошибка получения заказов:', error);
    res.status(500).json({ error: 'Внутренняя ошибка сервера' });
  }
});

// Принять заказ
app.post('/api/orders/:orderId/accept', async (req, res) => {
  try {
    const { telegram_id } = req.body;
    const orderId = parseInt(req.params.orderId);
    
    const user = await db.getUserByTelegramId(telegram_id);
    if (!user || user.role !== 'contractor') {
      return res.status(400).json({ error: 'Пользователь не является бригадиром' });
    }

    await db.acceptOrder(orderId, user.id);
    const order = await db.getOrder(orderId);
    
    res.json({ order });
  } catch (error) {
    console.error('Ошибка принятия заказа:', error);
    res.status(500).json({ error: 'Внутренняя ошибка сервера' });
  }
});

// Завершить заказ
app.post('/api/orders/:orderId/complete', async (req, res) => {
  try {
    const { telegram_id } = req.body;
    const orderId = parseInt(req.params.orderId);
    
    const user = await db.getUserByTelegramId(telegram_id);
    if (!user || user.role !== 'contractor') {
      return res.status(400).json({ error: 'Пользователь не является бригадиром' });
    }

    await db.completeOrder(orderId);
    const order = await db.getOrder(orderId);
    
    res.json({ order });
  } catch (error) {
    console.error('Ошибка завершения заказа:', error);
    res.status(500).json({ error: 'Внутренняя ошибка сервера' });
  }
});

// Отклонить заказ
app.post('/api/orders/:orderId/reject', async (req, res) => {
  try {
    const orderId = parseInt(req.params.orderId);
    await db.cancelOrder(orderId);
    res.json({ success: true });
  } catch (error) {
    console.error('Ошибка отклонения заказа:', error);
    res.status(500).json({ error: 'Внутренняя ошибка сервера' });
  }
});

// Главная страница
app.get('/', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

app.listen(PORT, () => {
  console.log(`Сервер запущен на порту ${PORT}`);
  console.log(`Откройте http://localhost:${PORT} для тестирования`);
});

// Graceful shutdown
process.on('SIGINT', async () => {
  console.log('\nЗавершение работы...');
  await db.close();
  process.exit(0);
});

