require('dotenv').config();
const Database = require('../db');
const path = require('path');

const dbPath = process.env.DB_PATH || path.join(__dirname, '..', 'data', 'database.db');
const db = new Database(dbPath);

console.log('База данных инициализирована!');
console.log('Таблицы созданы:');
console.log('- users');
console.log('- contractor_profiles');
console.log('- orders');
console.log('- reviews');

// Закрываем соединение
setTimeout(async () => {
  await db.close();
  process.exit(0);
}, 1000);

