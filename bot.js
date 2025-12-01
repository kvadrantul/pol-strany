require('dotenv').config();
const TelegramBot = require('node-telegram-bot-api');

const token = process.env.TELEGRAM_BOT_TOKEN || '8510455686:AAEvlK1P3_yt6btesENX_-d2OMDn5H6T1AQ';
if (!token) {
  console.error('TELEGRAM_BOT_TOKEN не установлен');
  process.exit(1);
}

const bot = new TelegramBot(token, { polling: true });
const APP_URL = process.env.APP_URL || 'https://pol-strany.vercel.app';

// Команда /start
bot.onText(/\/start/, (msg) => {
  const chatId = msg.chat.id;
  const firstName = msg.from.first_name;

  const options = {
    reply_markup: {
      inline_keyboard: [
        [
          {
            text: '🚀 Открыть приложение',
            web_app: { url: `${APP_URL}` }
          }
        ]
      ]
    }
  };

  bot.sendMessage(
    chatId,
    `Привет, ${firstName}! 👋\n\n` +
    `Добро пожаловать в "Пол Страны" - приложение для поиска бригад по стяжке пола.\n\n` +
    `Нажмите кнопку ниже, чтобы открыть приложение:`,
    options
  );
});

// Обработка callback от кнопок
bot.on('callback_query', (query) => {
  const chatId = query.message.chat.id;
  const data = query.data;

  if (data === 'open_app') {
    const options = {
      reply_markup: {
        inline_keyboard: [
          [
            {
              text: '🚀 Открыть приложение',
              web_app: { url: `${APP_URL}` }
            }
          ]
        ]
      }
    };
    bot.sendMessage(chatId, 'Откройте приложение:', options);
  }

  bot.answerCallbackQuery(query.id);
});

// Обработка сообщений
bot.on('message', (msg) => {
  const chatId = msg.chat.id;
  
  // Игнорируем команды и сообщения с web_app
  if (msg.text && msg.text.startsWith('/')) {
    return;
  }
  if (msg.web_app_data) {
    return;
  }

  // Если это обычное сообщение, предлагаем открыть приложение
  if (msg.text) {
    const options = {
      reply_markup: {
        inline_keyboard: [
          [
            {
              text: '🚀 Открыть приложение',
              web_app: { url: `${APP_URL}` }
            }
          ]
        ]
      }
    };
    bot.sendMessage(
      chatId,
      'Используйте приложение для работы с заказами. Нажмите кнопку ниже:',
      options
    );
  }
});

console.log('Telegram бот запущен!');

module.exports = bot;

