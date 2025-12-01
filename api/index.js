// Vercel serverless function entry point
const express = require('express');
const path = require('path');
const bodyParser = require('body-parser');
const cors = require('cors');

// Импортируем роуты
const apiRoutes = require('./routes');

const app = express();

// Middleware
app.use(cors());
app.use(bodyParser.json());

// API routes
app.use('/api', apiRoutes);

// Serve static files from public directory
app.use(express.static(path.join(__dirname, '../public')));

// Главная страница
app.get('/', (req, res) => {
  res.sendFile(path.join(__dirname, '../public', 'index.html'));
});

// Export для Vercel
module.exports = app;

