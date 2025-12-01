// Адаптер для работы с Turso (libSQL)
const { createClient } = require('@libsql/client');

class TursoDatabase {
  constructor() {
    const databaseUrl = process.env.DATABASE_URL;
    const authToken = process.env.TURSO_AUTH_TOKEN;

    if (!databaseUrl) {
      throw new Error('DATABASE_URL не установлен');
    }

    this.db = createClient({
      url: databaseUrl,
      authToken: authToken
    });

    // Инициализация будет выполнена при первом запросе
    this.initialized = false;
  }

  async ensureInitialized() {
    if (!this.initialized) {
      await this.init();
      this.initialized = true;
    }
  }

  async init() {
    // Создаем таблицы если их нет
    await this.db.execute(`
      CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        telegram_id INTEGER UNIQUE NOT NULL,
        role TEXT NOT NULL CHECK(role IN ('client', 'contractor')),
        name TEXT,
        phone TEXT,
        avatar_url TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
      )
    `);

    await this.db.execute(`
      CREATE TABLE IF NOT EXISTS contractor_profiles (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        experience_years INTEGER,
        rating REAL DEFAULT 0,
        completed_orders INTEGER DEFAULT 0,
        categories TEXT,
        is_active BOOLEAN DEFAULT 1,
        current_order_id INTEGER,
        FOREIGN KEY (user_id) REFERENCES users(id)
      )
    `);

    await this.db.execute(`
      CREATE TABLE IF NOT EXISTS orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        client_id INTEGER NOT NULL,
        contractor_id INTEGER,
        category TEXT NOT NULL,
        area REAL,
        address TEXT,
        status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'accepted', 'in_progress', 'completed', 'cancelled')),
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        accepted_at DATETIME,
        completed_at DATETIME,
        FOREIGN KEY (client_id) REFERENCES users(id),
        FOREIGN KEY (contractor_id) REFERENCES users(id)
      )
    `);

    await this.db.execute(`
      CREATE TABLE IF NOT EXISTS reviews (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_id INTEGER NOT NULL,
        contractor_id INTEGER NOT NULL,
        client_id INTEGER NOT NULL,
        rating INTEGER CHECK(rating >= 1 AND rating <= 5),
        comment TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (order_id) REFERENCES orders(id),
        FOREIGN KEY (contractor_id) REFERENCES users(id),
        FOREIGN KEY (client_id) REFERENCES users(id)
      )
    `);
  }

  // Пользователи
  async getUserByTelegramId(telegramId) {
    await this.ensureInitialized();
    const result = await this.db.execute({
      sql: 'SELECT * FROM users WHERE telegram_id = ?',
      args: [telegramId]
    });
    // Конвертируем Row в объект
    if (result.rows.length > 0) {
      const row = result.rows[0];
      return {
        id: row.id,
        telegram_id: row.telegram_id,
        role: row.role,
        name: row.name,
        phone: row.phone,
        avatar_url: row.avatar_url,
        created_at: row.created_at
      };
    }
    return null;
  }

  async createUser(telegramId, role, name = null, phone = null, avatarUrl = null) {
    await this.ensureInitialized();
    const result = await this.db.execute({
      sql: 'INSERT INTO users (telegram_id, role, name, phone, avatar_url) VALUES (?, ?, ?, ?, ?)',
      args: [telegramId, role, name, phone, avatarUrl]
    });
    return Number(result.lastInsertRowid);
  }

  async updateUser(telegramId, updates) {
    await this.ensureInitialized();
    const fields = [];
    const values = [];
    
    if (updates.name !== undefined) {
      fields.push('name = ?');
      values.push(updates.name);
    }
    if (updates.phone !== undefined) {
      fields.push('phone = ?');
      values.push(updates.phone);
    }
    if (updates.avatar_url !== undefined) {
      fields.push('avatar_url = ?');
      values.push(updates.avatar_url);
    }
    if (updates.role !== undefined) {
      fields.push('role = ?');
      values.push(updates.role);
    }

    if (fields.length === 0) return;

    values.push(telegramId);
    await this.db.execute({
      sql: `UPDATE users SET ${fields.join(', ')} WHERE telegram_id = ?`,
      args: values
    });
  }

  // Профили бригадиров
  async getContractorProfile(userId) {
    await this.ensureInitialized();
    const result = await this.db.execute({
      sql: `SELECT cp.*, u.name, u.phone, u.avatar_url, u.telegram_id 
            FROM contractor_profiles cp 
            JOIN users u ON cp.user_id = u.id 
            WHERE u.id = ?`,
      args: [userId]
    });
    if (result.rows.length > 0) {
      const row = result.rows[0];
      return {
        id: row.id,
        user_id: row.user_id,
        experience_years: row.experience_years,
        rating: row.rating,
        completed_orders: row.completed_orders,
        categories: row.categories,
        is_active: row.is_active,
        current_order_id: row.current_order_id,
        name: row.name,
        phone: row.phone,
        avatar_url: row.avatar_url,
        telegram_id: row.telegram_id
      };
    }
    return null;
  }

  async createOrUpdateContractorProfile(userId, profileData) {
    await this.ensureInitialized();
    const existing = await this.db.execute({
      sql: 'SELECT id FROM contractor_profiles WHERE user_id = ?',
      args: [userId]
    });

    const categories = JSON.stringify(profileData.categories || []);

    if (existing.rows.length > 0) {
      await this.db.execute({
        sql: `UPDATE contractor_profiles 
              SET experience_years = ?, categories = ?, is_active = ?
              WHERE user_id = ?`,
        args: [
          profileData.experience_years,
          categories,
          profileData.is_active !== false ? 1 : 0,
          userId
        ]
      });
    } else {
      await this.db.execute({
        sql: `INSERT INTO contractor_profiles (user_id, experience_years, categories, is_active)
              VALUES (?, ?, ?, ?)`,
        args: [
          userId,
          profileData.experience_years,
          categories,
          profileData.is_active !== false ? 1 : 0
        ]
      });
    }
  }

  async getAvailableContractors(category) {
    await this.ensureInitialized();
    const result = await this.db.execute({
      sql: `SELECT cp.*, u.name, u.phone, u.avatar_url, u.telegram_id, u.id as user_id
            FROM contractor_profiles cp
            JOIN users u ON cp.user_id = u.id
            WHERE cp.is_active = 1 
            AND (cp.current_order_id IS NULL OR cp.current_order_id = 0)
            AND (cp.categories LIKE ? OR cp.categories = '[]')
            ORDER BY cp.rating DESC, cp.completed_orders DESC
            LIMIT 10`,
      args: [`%${category}%`]
    });

    return result.rows.map(row => ({
      id: row.id,
      user_id: row.user_id,
      experience_years: row.experience_years,
      rating: row.rating,
      completed_orders: row.completed_orders,
      categories: JSON.parse(row.categories || '[]'),
      is_active: row.is_active,
      current_order_id: row.current_order_id,
      name: row.name,
      phone: row.phone,
      avatar_url: row.avatar_url,
      telegram_id: row.telegram_id
    }));
  }

  // Заказы
  async createOrder(clientId, category, area = null, address = null) {
    await this.ensureInitialized();
    const result = await this.db.execute({
      sql: 'INSERT INTO orders (client_id, category, area, address) VALUES (?, ?, ?, ?)',
      args: [clientId, category, area, address]
    });
    return Number(result.lastInsertRowid);
  }

  async getOrder(orderId) {
    await this.ensureInitialized();
    const result = await this.db.execute({
      sql: `SELECT o.*, 
                   uc.name as client_name, uc.telegram_id as client_telegram_id,
                   uct.name as contractor_name, uct.telegram_id as contractor_telegram_id
            FROM orders o
            LEFT JOIN users uc ON o.client_id = uc.id
            LEFT JOIN users uct ON o.contractor_id = uct.id
            WHERE o.id = ?`,
      args: [orderId]
    });
    if (result.rows.length > 0) {
      return this.rowToObject(result.rows[0]);
    }
    return null;
  }

  rowToObject(row) {
    const obj = {};
    for (const key in row) {
      obj[key] = row[key];
    }
    return obj;
  }

  async getContractorOrders(contractorId) {
    await this.ensureInitialized();
    const result = await this.db.execute({
      sql: `SELECT o.*, u.name as client_name, u.telegram_id as client_telegram_id
            FROM orders o
            JOIN users u ON o.client_id = u.id
            WHERE o.contractor_id = ?
            ORDER BY o.created_at DESC`,
      args: [contractorId]
    });
    return result.rows.map(row => this.rowToObject(row));
  }

  async getAllPendingOrders() {
    await this.ensureInitialized();
    const result = await this.db.execute({
      sql: `SELECT o.*, u.name as client_name, u.telegram_id as client_telegram_id
            FROM orders o
            JOIN users u ON o.client_id = u.id
            WHERE o.status = 'pending'
            ORDER BY o.created_at DESC`
    });
    return result.rows.map(row => this.rowToObject(row));
  }

  async acceptOrder(orderId, contractorId) {
    await this.ensureInitialized();
    // Turso поддерживает транзакции через batch
    await this.db.batch([
      {
        sql: `UPDATE orders 
              SET contractor_id = ?, status = 'accepted', accepted_at = CURRENT_TIMESTAMP
              WHERE id = ? AND status = 'pending'`,
        args: [contractorId, orderId]
      },
      {
        sql: `UPDATE contractor_profiles 
              SET current_order_id = ?
              WHERE user_id = ?`,
        args: [orderId, contractorId]
      }
    ]);
  }

  async completeOrder(orderId) {
    await this.ensureInitialized();
    // Получаем contractor_id из заказа
    const orderResult = await this.db.execute({
      sql: 'SELECT contractor_id FROM orders WHERE id = ?',
      args: [orderId]
    });

    const order = orderResult.rows[0];
    const contractorId = order ? order.contractor_id : null;

    if (contractorId) {
      // Используем batch для атомарности
      await this.db.batch([
        {
          sql: `UPDATE orders 
                SET status = 'completed', completed_at = CURRENT_TIMESTAMP
                WHERE id = ?`,
          args: [orderId]
        },
        {
          sql: `UPDATE contractor_profiles 
                SET current_order_id = NULL,
                    completed_orders = completed_orders + 1
                WHERE user_id = ?`,
          args: [contractorId]
        }
      ]);
    } else {
      await this.db.execute({
        sql: `UPDATE orders 
              SET status = 'completed', completed_at = CURRENT_TIMESTAMP
              WHERE id = ?`,
        args: [orderId]
      });
    }
  }

  async cancelOrder(orderId) {
    await this.ensureInitialized();
    await this.db.execute({
      sql: `UPDATE orders SET status = 'cancelled' WHERE id = ?`,
      args: [orderId]
    });
  }

  async close() {
    // Turso клиент не требует явного закрытия
    return Promise.resolve();
  }
}

module.exports = TursoDatabase;

