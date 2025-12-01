// Скрипт для проверки подключения к Turso
require('dotenv').config();
const { createClient } = require('@libsql/client');

async function testConnection() {
  const databaseUrl = process.env.DATABASE_URL || 'libsql://pol-strany-hun7eee.aws-eu-west-1.turso.io';
  const authToken = process.env.TURSO_AUTH_TOKEN || 'eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJhIjoicnciLCJpYXQiOjE3NjQ2MjA5MzgsImlkIjoiYjIzMDljZmMtZGFhMi00MWViLTkyYTAtNzM3NDJiM2UwNzA4IiwicmlkIjoiYWFiNTY0M2EtOWZiOS00MjQ1LWFhYmQtZDNmNzEzNmIwZDVjIn0.fF8XEVaKVVR4bgJ3BEBtnoU6v3AvnryFS5rKzJ-iity9WYvwHyNxHjjGLGhjMUOo9vIDITN8EW0Z7W5wxrGxDA';

  console.log('🔌 Подключение к Turso...');
  console.log('URL:', databaseUrl);

  try {
    const client = createClient({
      url: databaseUrl,
      authToken: authToken
    });

    // Тест подключения
    console.log('📊 Проверка подключения...');
    const result = await client.execute('SELECT 1 as test');
    console.log('✅ Подключение успешно!');
    console.log('Результат:', result.rows);

    // Проверка существующих таблиц
    console.log('\n📋 Проверка таблиц...');
    const tables = await client.execute(`
      SELECT name FROM sqlite_master 
      WHERE type='table' AND name NOT LIKE 'sqlite_%'
    `);
    
    if (tables.rows.length > 0) {
      console.log('Найдены таблицы:');
      tables.rows.forEach(row => {
        console.log(`  - ${row.name}`);
      });
    } else {
      console.log('⚠️  Таблицы еще не созданы (это нормально для новой БД)');
    }

    // Создание тестовой таблицы
    console.log('\n🔨 Создание тестовой таблицы...');
    await client.execute(`
      CREATE TABLE IF NOT EXISTS test_table (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        message TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
      )
    `);
    console.log('✅ Тестовая таблица создана');

    // Вставка тестовых данных
    console.log('\n💾 Вставка тестовых данных...');
    await client.execute({
      sql: 'INSERT INTO test_table (message) VALUES (?)',
      args: ['Тест подключения к Turso']
    });
    console.log('✅ Данные вставлены');

    // Чтение данных
    console.log('\n📖 Чтение данных...');
    const readResult = await client.execute('SELECT * FROM test_table');
    console.log('Данные из БД:');
    readResult.rows.forEach(row => {
      console.log(`  ID: ${row.id}, Message: ${row.message}, Created: ${row.created_at}`);
    });

    console.log('\n🎉 Все тесты пройдены успешно!');
    console.log('✅ Turso Database работает корректно');

  } catch (error) {
    console.error('❌ Ошибка подключения:', error.message);
    console.error('Детали:', error);
    process.exit(1);
  }
}

testConnection();

