const sqlite3 = require('sqlite3').verbose();
const path = require('path');
const fs = require('fs');

class Database {
  constructor(dbPath) {
    // Создаем директорию для БД если её нет
    const dbDir = path.dirname(dbPath);
    if (!fs.existsSync(dbDir)) {
      fs.mkdirSync(dbDir, { recursive: true });
    }

    this.db = new sqlite3.Database(dbPath, (err) => {
      if (err) {
        console.error('Ошибка подключения к БД:', err.message);
      } else {
        console.log('Подключено к SQLite БД');
      }
    });

    this.init();
  }

  init() {
    // Таблица пользователей
    this.db.run(`
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

    // Таблица профилей бригадиров
    this.db.run(`
      CREATE TABLE IF NOT EXISTS contractor_profiles (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        experience_years INTEGER,
        rating REAL DEFAULT 0,
        completed_orders INTEGER DEFAULT 0,
        categories TEXT, -- JSON массив категорий
        is_active BOOLEAN DEFAULT 1,
        current_order_id INTEGER,
        FOREIGN KEY (user_id) REFERENCES users(id)
      )
    `);

    // Таблица заказов
    this.db.run(`
      CREATE TABLE IF NOT EXISTS orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        client_id INTEGER NOT NULL,
        contractor_id INTEGER,
        category TEXT NOT NULL,
        area REAL, -- площадь в м²
        address TEXT,
        status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'accepted', 'in_progress', 'completed', 'cancelled')),
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        accepted_at DATETIME,
        completed_at DATETIME,
        FOREIGN KEY (client_id) REFERENCES users(id),
        FOREIGN KEY (contractor_id) REFERENCES users(id)
      )
    `);

    // Таблица отзывов
    this.db.run(`
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
    return new Promise((resolve, reject) => {
      this.db.get(
        'SELECT * FROM users WHERE telegram_id = ?',
        [telegramId],
        (err, row) => {
          if (err) reject(err);
          else resolve(row);
        }
      );
    });
  }

  async createUser(telegramId, role, name = null, phone = null, avatarUrl = null) {
    return new Promise((resolve, reject) => {
      this.db.run(
        'INSERT INTO users (telegram_id, role, name, phone, avatar_url) VALUES (?, ?, ?, ?, ?)',
        [telegramId, role, name, phone, avatarUrl],
        function(err) {
          if (err) reject(err);
          else resolve(this.lastID);
        }
      );
    });
  }

  async updateUser(telegramId, updates) {
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
    return new Promise((resolve, reject) => {
      this.db.run(
        `UPDATE users SET ${fields.join(', ')} WHERE telegram_id = ?`,
        values,
        (err) => {
          if (err) reject(err);
          else resolve();
        }
      );
    });
  }

  // Профили бригадиров
  async getContractorProfile(userId) {
    return new Promise((resolve, reject) => {
      this.db.get(
        `SELECT cp.*, u.name, u.phone, u.avatar_url, u.telegram_id 
         FROM contractor_profiles cp 
         JOIN users u ON cp.user_id = u.id 
         WHERE u.id = ?`,
        [userId],
        (err, row) => {
          if (err) reject(err);
          else resolve(row);
        }
      );
    });
  }

  async createOrUpdateContractorProfile(userId, profileData) {
    return new Promise((resolve, reject) => {
      this.db.get(
        'SELECT id FROM contractor_profiles WHERE user_id = ?',
        [userId],
        (err, row) => {
          if (err) {
            reject(err);
            return;
          }

          if (row) {
            // Обновляем существующий профиль
            const categories = JSON.stringify(profileData.categories || []);
            this.db.run(
              `UPDATE contractor_profiles 
               SET experience_years = ?, categories = ?, is_active = ?
               WHERE user_id = ?`,
              [
                profileData.experience_years,
                categories,
                profileData.is_active !== false ? 1 : 0,
                userId
              ],
              (err) => {
                if (err) reject(err);
                else resolve();
              }
            );
          } else {
            // Создаем новый профиль
            const categories = JSON.stringify(profileData.categories || []);
            this.db.run(
              `INSERT INTO contractor_profiles (user_id, experience_years, categories, is_active)
               VALUES (?, ?, ?, ?)`,
              [
                userId,
                profileData.experience_years,
                categories,
                profileData.is_active !== false ? 1 : 0
              ],
              function(err) {
                if (err) reject(err);
                else resolve(this.lastID);
              }
            );
          }
        }
      );
    });
  }

  async getAvailableContractors(category) {
    return new Promise((resolve, reject) => {
      this.db.all(
        `SELECT cp.*, u.name, u.phone, u.avatar_url, u.telegram_id, u.id as user_id
         FROM contractor_profiles cp
         JOIN users u ON cp.user_id = u.id
         WHERE cp.is_active = 1 
         AND (cp.current_order_id IS NULL OR cp.current_order_id = 0)
         AND (cp.categories LIKE ? OR cp.categories = '[]')
         ORDER BY cp.rating DESC, cp.completed_orders DESC
         LIMIT 10`,
        [`%${category}%`],
        (err, rows) => {
          if (err) reject(err);
          else {
            // Парсим JSON категории
            const contractors = rows.map(row => ({
              ...row,
              categories: JSON.parse(row.categories || '[]')
            }));
            resolve(contractors);
          }
        }
      );
    });
  }

  // Заказы
  async createOrder(clientId, category, area = null, address = null) {
    return new Promise((resolve, reject) => {
      this.db.run(
        'INSERT INTO orders (client_id, category, area, address) VALUES (?, ?, ?, ?)',
        [clientId, category, area, address],
        function(err) {
          if (err) reject(err);
          else resolve(this.lastID);
        }
      );
    });
  }

  async getOrder(orderId) {
    return new Promise((resolve, reject) => {
      this.db.get(
        `SELECT o.*, 
                uc.name as client_name, uc.telegram_id as client_telegram_id,
                uct.name as contractor_name, uct.telegram_id as contractor_telegram_id
         FROM orders o
         LEFT JOIN users uc ON o.client_id = uc.id
         LEFT JOIN users uct ON o.contractor_id = uct.id
         WHERE o.id = ?`,
        [orderId],
        (err, row) => {
          if (err) reject(err);
          else resolve(row);
        }
      );
    });
  }

  async getPendingOrdersForContractor(contractorId) {
    return new Promise((resolve, reject) => {
      this.db.all(
        `SELECT o.*, u.name as client_name, u.telegram_id as client_telegram_id
         FROM orders o
         JOIN users u ON o.client_id = u.id
         WHERE o.status = 'pending'
         ORDER BY o.created_at DESC`,
        (err, rows) => {
          if (err) reject(err);
          else resolve(rows);
        }
      );
    });
  }

  async getAllPendingOrders() {
    return new Promise((resolve, reject) => {
      this.db.all(
        `SELECT o.*, u.name as client_name, u.telegram_id as client_telegram_id
         FROM orders o
         JOIN users u ON o.client_id = u.id
         WHERE o.status = 'pending'
         ORDER BY o.created_at DESC`,
        (err, rows) => {
          if (err) reject(err);
          else resolve(rows);
        }
      );
    });
  }

  async getContractorOrders(contractorId) {
    return new Promise((resolve, reject) => {
      this.db.all(
        `SELECT o.*, u.name as client_name, u.telegram_id as client_telegram_id
         FROM orders o
         JOIN users u ON o.client_id = u.id
         WHERE o.contractor_id = ?
         ORDER BY o.created_at DESC`,
        [contractorId],
        (err, rows) => {
          if (err) reject(err);
          else resolve(rows);
        }
      );
    });
  }

  async acceptOrder(orderId, contractorId) {
    return new Promise((resolve, reject) => {
      this.db.serialize(() => {
        this.db.run('BEGIN TRANSACTION');
        
        // Обновляем заказ
        this.db.run(
          `UPDATE orders 
           SET contractor_id = ?, status = 'accepted', accepted_at = CURRENT_TIMESTAMP
           WHERE id = ? AND status = 'pending'`,
          [contractorId, orderId],
          function(err) {
            if (err) {
              this.db.run('ROLLBACK');
              reject(err);
              return;
            }

            // Обновляем профиль бригадира
            this.db.run(
              `UPDATE contractor_profiles 
               SET current_order_id = ?
               WHERE user_id = ?`,
              [orderId, contractorId],
              (err) => {
                if (err) {
                  this.db.run('ROLLBACK');
                  reject(err);
                } else {
                  this.db.run('COMMIT');
                  resolve();
                }
              }
            );
          }
        );
      });
    });
  }

  async completeOrder(orderId) {
    return new Promise((resolve, reject) => {
      this.db.serialize(() => {
        this.db.run('BEGIN TRANSACTION');

        // Получаем contractor_id из заказа
        this.db.get(
          'SELECT contractor_id FROM orders WHERE id = ?',
          [orderId],
          (err, order) => {
            if (err) {
              this.db.run('ROLLBACK');
              reject(err);
              return;
            }

            // Обновляем заказ
            this.db.run(
              `UPDATE orders 
               SET status = 'completed', completed_at = CURRENT_TIMESTAMP
               WHERE id = ?`,
              [orderId],
              (err) => {
                if (err) {
                  this.db.run('ROLLBACK');
                  reject(err);
                  return;
                }

                // Освобождаем бригадира
                if (order && order.contractor_id) {
                  this.db.run(
                    `UPDATE contractor_profiles 
                     SET current_order_id = NULL,
                         completed_orders = completed_orders + 1
                     WHERE user_id = ?`,
                    [order.contractor_id],
                    (err) => {
                      if (err) {
                        this.db.run('ROLLBACK');
                        reject(err);
                      } else {
                        this.db.run('COMMIT');
                        resolve();
                      }
                    }
                  );
                } else {
                  this.db.run('COMMIT');
                  resolve();
                }
              }
            );
          }
        );
      });
    });
  }

  async cancelOrder(orderId) {
    return new Promise((resolve, reject) => {
      this.db.run(
        `UPDATE orders SET status = 'cancelled' WHERE id = ?`,
        [orderId],
        (err) => {
          if (err) reject(err);
          else resolve();
        }
      );
    });
  }

  close() {
    return new Promise((resolve, reject) => {
      this.db.close((err) => {
        if (err) reject(err);
        else resolve();
      });
    });
  }
}

module.exports = Database;

