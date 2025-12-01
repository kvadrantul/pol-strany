package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var app *App

type App struct {
	db *sql.DB
}

// Тарифы
var TARIFFS = map[string]Tariff{
	"econom": {
		Name:        "ЭКОНОМ",
		Description: "Мокрая, ручная",
		PriceRange: PriceRange{Min: 400, Max: 450},
		Days:        "28 дней",
		Features:    []string{"Классика", "Низкая цена материалов", "Долгий срок высыхания", "Высокий риск трещин"},
	},
	"comfort": {
		Name:        "КОМФОРТ",
		Description: "Полусухая механизированная",
		PriceRange:  PriceRange{Min: 550, Max: 850},
		Days:        "5-7 дней (плитка — 2 дня, ламинат — 14–20 дней)",
		Features:    []string{"Оптимальный баланс", "Минимум усадки", "Можно ходить через 12 часов", "Самый популярный выбор"},
	},
	"business": {
		Name:        "БИЗНЕС",
		Description: "С армированием",
		PriceRange:  PriceRange{Min: 150, Max: 300},
		Days:        "Как у базового тарифа",
		Features:    []string{"Повышенная прочность", "Надбавка за армирование сеткой или фиброй"},
		IsAddon:     true,
	},
	"premium": {
		Name:        "ПРЕМИУМ",
		Description: "Сухая стяжка Кнауф",
		PriceRange:  PriceRange{Min: 800, Max: 1000},
		Days:        "1-2 дня",
		Features:    []string{"Нет мокрых процессов", "Идеальная геометрия", "Теплоизоляция", "Высокая цена материалов"},
	},
	"universal": {
		Name:        "УНИВЕРСАЛ",
		Description: "Плавающая / Утепленная",
		PriceRange:  PriceRange{Min: 250, Max: 600},
		Days:        "Как у базового тарифа",
		Features:    []string{"Зависит от вида утеплителя", "Включает слой изоляции"},
		IsAddon:     true,
	},
	"self-leveling": {
		Name:        "САМОВЫРАВНИВАТЕЛЬ",
		Description: "Финишный слой",
		PriceRange:  PriceRange{Min: 250, Max: 500},
		Days:        "1-3 дня",
		Features:    []string{"Финишный слой"},
	},
}

type Tariff struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	PriceRange  PriceRange `json:"priceRange"`
	Days        string     `json:"days"`
	Features    []string   `json:"features"`
	IsAddon     bool       `json:"isAddon,omitempty"`
}

type PriceRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type User struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Role       string    `json:"role"`
	Name       *string   `json:"name"`
	Phone      *string   `json:"phone"`
	AvatarURL  *string   `json:"avatar_url"`
	CreatedAt  time.Time `json:"created_at"`
}

type ContractorProfile struct {
	ID              int64   `json:"id"`
	UserID          int64   `json:"user_id"`
	ExperienceYears *int    `json:"experience_years"`
	Rating          float64 `json:"rating"`
	CompletedOrders int     `json:"completed_orders"`
	Categories      string  `json:"categories"`
	IsActive        bool    `json:"is_active"`
	CurrentOrderID  *int64  `json:"current_order_id"`
	Name            *string `json:"name"`
	Phone           *string `json:"phone"`
	AvatarURL       *string `json:"avatar_url"`
	TelegramID      *int64  `json:"telegram_id"`
}

type Order struct {
	ID                   int64      `json:"id"`
	ClientID             int64      `json:"client_id"`
	ContractorID         *int64     `json:"contractor_id"`
	Category             string     `json:"category"`
	Area                 *float64   `json:"area"`
	Address              *string    `json:"address"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	AcceptedAt           *time.Time `json:"accepted_at"`
	CompletedAt          *time.Time `json:"completed_at"`
	ClientName           *string    `json:"client_name"`
	ClientTelegramID     *int64     `json:"client_telegram_id"`
	ContractorName       *string    `json:"contractor_name"`
	ContractorTelegramID *int64     `json:"contractor_telegram_id"`
}

var dbInitialized bool

func initDBIfNeeded() error {
	if dbInitialized {
		return nil
	}

	databaseURL := os.Getenv("DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	if databaseURL == "" {
		// Не падаем, просто не инициализируем БД
		// Это позволит отдавать статические файлы
		return fmt.Errorf("DATABASE_URL не установлен")
	}

	dsn := databaseURL
	if authToken != "" {
		if strings.Contains(dsn, "?") {
			dsn += "&authToken=" + authToken
		} else {
			dsn += "?authToken=" + authToken
		}
	}

	db, err := sql.Open("libsql", dsn)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	app = &App{db: db}

	if err := app.initDB(); err != nil {
		return fmt.Errorf("ошибка инициализации БД: %w", err)
	}

	dbInitialized = true
	return nil
}

func (app *App) initDB() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id INTEGER UNIQUE NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('client', 'contractor')),
			name TEXT,
			phone TEXT,
			avatar_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS contractor_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			experience_years INTEGER,
			rating REAL DEFAULT 0,
			completed_orders INTEGER DEFAULT 0,
			categories TEXT,
			is_active BOOLEAN DEFAULT 1,
			current_order_id INTEGER,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
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
		)`,
		`CREATE TABLE IF NOT EXISTS reviews (
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
		)`,
	}

	for _, query := range queries {
		if _, err := app.db.Exec(query); err != nil {
			return fmt.Errorf("ошибка создания таблицы: %w", err)
		}
	}
	return nil
}

func (app *App) getUserByTelegramID(telegramID int64) (*User, error) {
	row := app.db.QueryRow("SELECT id, telegram_id, role, name, phone, avatar_url, created_at FROM users WHERE telegram_id = ?", telegramID)
	var user User
	var createdAt sql.NullString
	err := row.Scan(&user.ID, &user.TelegramID, &user.Role, &user.Name, &user.Phone, &user.AvatarURL, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if createdAt.Valid && createdAt.String != "" {
		user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}
	return &user, nil
}

func (app *App) createUser(telegramID int64, role string, name, phone, avatarURL *string) (int64, error) {
	result, err := app.db.Exec("INSERT INTO users (telegram_id, role, name, phone, avatar_url) VALUES (?, ?, ?, ?, ?)", telegramID, role, name, phone, avatarURL)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}

func (app *App) updateUser(telegramID int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	setParts := []string{}
	args := []interface{}{}
	if name, ok := updates["name"].(*string); ok {
		setParts = append(setParts, "name = ?")
		args = append(args, name)
	}
	if phone, ok := updates["phone"].(*string); ok {
		setParts = append(setParts, "phone = ?")
		args = append(args, phone)
	}
	if avatarURL, ok := updates["avatar_url"].(*string); ok {
		setParts = append(setParts, "avatar_url = ?")
		args = append(args, avatarURL)
	}
	if role, ok := updates["role"].(string); ok {
		setParts = append(setParts, "role = ?")
		args = append(args, role)
	}
	if len(setParts) == 0 {
		return nil
	}
	args = append(args, telegramID)
	query := fmt.Sprintf("UPDATE users SET %s WHERE telegram_id = ?", strings.Join(setParts, ", "))
	_, err := app.db.Exec(query, args...)
	return err
}

func (app *App) getContractorProfile(userID int64) (*ContractorProfile, error) {
	row := app.db.QueryRow(`SELECT cp.id, cp.user_id, cp.experience_years, cp.rating, cp.completed_orders, cp.categories, cp.is_active, cp.current_order_id, u.name, u.phone, u.avatar_url, u.telegram_id FROM contractor_profiles cp JOIN users u ON cp.user_id = u.id WHERE u.id = ?`, userID)
	var profile ContractorProfile
	err := row.Scan(&profile.ID, &profile.UserID, &profile.ExperienceYears, &profile.Rating, &profile.CompletedOrders, &profile.Categories, &profile.IsActive, &profile.CurrentOrderID, &profile.Name, &profile.Phone, &profile.AvatarURL, &profile.TelegramID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (app *App) createOrUpdateContractorProfile(userID int64, experienceYears *int, categories []string, isActive bool) error {
	categoriesJSON, _ := json.Marshal(categories)
	categoriesStr := string(categoriesJSON)
	row := app.db.QueryRow("SELECT id FROM contractor_profiles WHERE user_id = ?", userID)
	var existingID int64
	err := row.Scan(&existingID)
	if err == nil {
		_, err = app.db.Exec(`UPDATE contractor_profiles SET experience_years = ?, categories = ?, is_active = ? WHERE user_id = ?`, experienceYears, categoriesStr, isActive, userID)
		return err
	} else if err == sql.ErrNoRows {
		_, err = app.db.Exec(`INSERT INTO contractor_profiles (user_id, experience_years, categories, is_active) VALUES (?, ?, ?, ?)`, userID, experienceYears, categoriesStr, isActive)
		return err
	}
	return err
}

func (app *App) getAvailableContractors(category string) ([]ContractorProfile, error) {
	rows, err := app.db.Query(`SELECT cp.id, cp.user_id, cp.experience_years, cp.rating, cp.completed_orders, cp.categories, cp.is_active, cp.current_order_id, u.name, u.phone, u.avatar_url, u.telegram_id FROM contractor_profiles cp JOIN users u ON cp.user_id = u.id WHERE cp.is_active = 1 AND (cp.current_order_id IS NULL OR cp.current_order_id = 0) AND (cp.categories LIKE ? OR cp.categories = '[]') ORDER BY cp.rating DESC, cp.completed_orders DESC LIMIT 10`, "%"+category+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contractors []ContractorProfile
	for rows.Next() {
		var profile ContractorProfile
		err := rows.Scan(&profile.ID, &profile.UserID, &profile.ExperienceYears, &profile.Rating, &profile.CompletedOrders, &profile.Categories, &profile.IsActive, &profile.CurrentOrderID, &profile.Name, &profile.Phone, &profile.AvatarURL, &profile.TelegramID)
		if err != nil {
			return nil, err
		}
		contractors = append(contractors, profile)
	}
	return contractors, nil
}

func (app *App) createOrder(clientID int64, category string, area *float64, address *string) (int64, error) {
	result, err := app.db.Exec("INSERT INTO orders (client_id, category, area, address) VALUES (?, ?, ?, ?)", clientID, category, area, address)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}

func (app *App) getOrder(orderID int64) (*Order, error) {
	row := app.db.QueryRow(`SELECT o.id, o.client_id, o.contractor_id, o.category, o.area, o.address, o.status, o.created_at, o.accepted_at, o.completed_at, uc.name, uc.telegram_id, uct.name, uct.telegram_id FROM orders o LEFT JOIN users uc ON o.client_id = uc.id LEFT JOIN users uct ON o.contractor_id = uct.id WHERE o.id = ?`, orderID)
	var order Order
	var createdAt, acceptedAt, completedAt sql.NullString
	err := row.Scan(&order.ID, &order.ClientID, &order.ContractorID, &order.Category, &order.Area, &order.Address, &order.Status, &createdAt, &acceptedAt, &completedAt, &order.ClientName, &order.ClientTelegramID, &order.ContractorName, &order.ContractorTelegramID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if createdAt.Valid && createdAt.String != "" {
		order.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}
	if acceptedAt.Valid && acceptedAt.String != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", acceptedAt.String)
		order.AcceptedAt = &t
	}
	if completedAt.Valid && completedAt.String != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
		order.CompletedAt = &t
	}
	return &order, nil
}

func (app *App) getContractorOrders(contractorID int64) ([]Order, error) {
	rows, err := app.db.Query(`SELECT o.id, o.client_id, o.contractor_id, o.category, o.area, o.address, o.status, o.created_at, o.accepted_at, o.completed_at, u.name, u.telegram_id FROM orders o JOIN users u ON o.client_id = u.id WHERE o.contractor_id = ? ORDER BY o.created_at DESC`, contractorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []Order
	for rows.Next() {
		var order Order
		var createdAt, acceptedAt, completedAt sql.NullString
		err := rows.Scan(&order.ID, &order.ClientID, &order.ContractorID, &order.Category, &order.Area, &order.Address, &order.Status, &createdAt, &acceptedAt, &completedAt, &order.ClientName, &order.ClientTelegramID)
		if err != nil {
			return nil, err
		}
		if createdAt.Valid && createdAt.String != "" {
			order.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}
		if acceptedAt.Valid && acceptedAt.String != "" {
			t, _ := time.Parse("2006-01-02 15:04:05", acceptedAt.String)
			order.AcceptedAt = &t
		}
		if completedAt.Valid && completedAt.String != "" {
			t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
			order.CompletedAt = &t
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (app *App) getAllPendingOrders() ([]Order, error) {
	rows, err := app.db.Query(`SELECT o.id, o.client_id, o.contractor_id, o.category, o.area, o.address, o.status, o.created_at, o.accepted_at, o.completed_at, u.name, u.telegram_id FROM orders o JOIN users u ON o.client_id = u.id WHERE o.status = 'pending' ORDER BY o.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []Order
	for rows.Next() {
		var order Order
		var createdAt, acceptedAt, completedAt sql.NullString
		err := rows.Scan(&order.ID, &order.ClientID, &order.ContractorID, &order.Category, &order.Area, &order.Address, &order.Status, &createdAt, &acceptedAt, &completedAt, &order.ClientName, &order.ClientTelegramID)
		if err != nil {
			return nil, err
		}
		if createdAt.Valid && createdAt.String != "" {
			order.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (app *App) acceptOrder(orderID, contractorID int64) error {
	queries := []string{
		`UPDATE orders SET contractor_id = ?, status = 'accepted', accepted_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'pending'`,
		`UPDATE contractor_profiles SET current_order_id = ? WHERE user_id = ?`,
	}
	args := [][]interface{}{
		{contractorID, orderID},
		{orderID, contractorID},
	}
	for i, query := range queries {
		if _, err := app.db.Exec(query, args[i]...); err != nil {
			return err
		}
	}
	return nil
}

func (app *App) completeOrder(orderID int64) error {
	row := app.db.QueryRow("SELECT contractor_id FROM orders WHERE id = ?", orderID)
	var contractorID sql.NullInt64
	err := row.Scan(&contractorID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	_, err = app.db.Exec(`UPDATE orders SET status = 'completed', completed_at = CURRENT_TIMESTAMP WHERE id = ?`, orderID)
	if err != nil {
		return err
	}
	if contractorID.Valid {
		_, err = app.db.Exec(`UPDATE contractor_profiles SET current_order_id = NULL, completed_orders = completed_orders + 1 WHERE user_id = ?`, contractorID.Int64)
		return err
	}
	return nil
}

func (app *App) cancelOrder(orderID int64) error {
	_, err := app.db.Exec("UPDATE orders SET status = 'cancelled' WHERE id = ?", orderID)
	return err
}

func (app *App) getTariffs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TARIFFS)
}

func (app *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	telegramID, err := strconv.ParseInt(vars["telegramId"], 10, 64)
	if err != nil {
		http.Error(w, "Неверный telegram ID", http.StatusBadRequest)
		return
	}
	user, err := app.getUserByTelegramID(telegramID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}
	var profile *ContractorProfile
	if user.Role == "contractor" {
		profile, err = app.getContractorProfile(user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"user": user, "profile": profile})
}

func (app *App) createOrUpdateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TelegramID int64   `json:"telegram_id"`
		Role       string  `json:"role"`
		Name       *string `json:"name"`
		Phone      *string `json:"phone"`
		AvatarURL  *string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	user, err := app.getUserByTelegramID(req.TelegramID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		_, err := app.createUser(req.TelegramID, req.Role, req.Name, req.Phone, req.AvatarURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user, err = app.getUserByTelegramID(req.TelegramID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		updates := map[string]interface{}{}
		if req.Role != "" {
			updates["role"] = req.Role
		}
		if req.Name != nil {
			updates["name"] = req.Name
		}
		if req.Phone != nil {
			updates["phone"] = req.Phone
		}
		if req.AvatarURL != nil {
			updates["avatar_url"] = req.AvatarURL
		}
		if err := app.updateUser(req.TelegramID, updates); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user, err = app.getUserByTelegramID(req.TelegramID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"user": user})
}

func (app *App) updateContractorProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TelegramID      int64    `json:"telegram_id"`
		ExperienceYears *int     `json:"experience_years"`
		Categories      []string `json:"categories"`
		IsActive        bool     `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	user, err := app.getUserByTelegramID(req.TelegramID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil || user.Role != "contractor" {
		http.Error(w, "Пользователь не является бригадиром", http.StatusBadRequest)
		return
	}
	if err := app.createOrUpdateContractorProfile(user.ID, req.ExperienceYears, req.Categories, req.IsActive); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	profile, err := app.getContractorProfile(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"profile": profile})
}

func (app *App) searchContractors(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		http.Error(w, "Не указана категория", http.StatusBadRequest)
		return
	}
	contractors, err := app.getAvailableContractors(category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type ContractorResponse struct {
		ID              int64    `json:"id"`
		UserID          int64    `json:"user_id"`
		ExperienceYears *int     `json:"experience_years"`
		Rating          float64  `json:"rating"`
		CompletedOrders int      `json:"completed_orders"`
		Categories      []string `json:"categories"`
		IsActive        bool     `json:"is_active"`
		CurrentOrderID  *int64   `json:"current_order_id"`
		Name            *string  `json:"name"`
		Phone           *string  `json:"phone"`
		AvatarURL       *string  `json:"avatar_url"`
		TelegramID      *int64   `json:"telegram_id"`
	}
	response := make([]ContractorResponse, 0, len(contractors))
	for i := range contractors {
		var categories []string
		if contractors[i].Categories != "" {
			json.Unmarshal([]byte(contractors[i].Categories), &categories)
		}
		response = append(response, ContractorResponse{
			ID:              contractors[i].ID,
			UserID:          contractors[i].UserID,
			ExperienceYears: contractors[i].ExperienceYears,
			Rating:          contractors[i].Rating,
			CompletedOrders: contractors[i].CompletedOrders,
			Categories:      categories,
			IsActive:        contractors[i].IsActive,
			CurrentOrderID:  contractors[i].CurrentOrderID,
			Name:            contractors[i].Name,
			Phone:           contractors[i].Phone,
			AvatarURL:       contractors[i].AvatarURL,
			TelegramID:      contractors[i].TelegramID,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"contractors": response})
}

func (app *App) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TelegramID int64    `json:"telegram_id"`
		Category   string   `json:"category"`
		Area       *float64 `json:"area"`
		Address    *string  `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	user, err := app.getUserByTelegramID(req.TelegramID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil || user.Role != "client" {
		http.Error(w, "Пользователь не является клиентом", http.StatusBadRequest)
		return
	}
	orderID, err := app.createOrder(user.ID, req.Category, req.Area, req.Address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	order, err := app.getOrder(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"order": order})
}

func (app *App) handleGetContractorOrders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	telegramID, err := strconv.ParseInt(vars["telegramId"], 10, 64)
	if err != nil {
		http.Error(w, "Неверный telegram ID", http.StatusBadRequest)
		return
	}
	user, err := app.getUserByTelegramID(telegramID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil || user.Role != "contractor" {
		http.Error(w, "Пользователь не является бригадиром", http.StatusBadRequest)
		return
	}
	orders, err := app.getContractorOrders(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"orders": orders})
}

func (app *App) getPendingOrders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	telegramID, err := strconv.ParseInt(vars["telegramId"], 10, 64)
	if err != nil {
		http.Error(w, "Неверный telegram ID", http.StatusBadRequest)
		return
	}
	user, err := app.getUserByTelegramID(telegramID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil || user.Role != "contractor" {
		http.Error(w, "Пользователь не является бригадиром", http.StatusBadRequest)
		return
	}
	orders, err := app.getAllPendingOrders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"orders": orders})
}

func (app *App) handleAcceptOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		http.Error(w, "Неверный order ID", http.StatusBadRequest)
		return
	}
	var req struct {
		TelegramID int64 `json:"telegram_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	user, err := app.getUserByTelegramID(req.TelegramID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil || user.Role != "contractor" {
		http.Error(w, "Пользователь не является бригадиром", http.StatusBadRequest)
		return
	}
	if err := app.acceptOrder(orderID, user.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	order, err := app.getOrder(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"order": order})
}

func (app *App) handleCompleteOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		http.Error(w, "Неверный order ID", http.StatusBadRequest)
		return
	}
	var req struct {
		TelegramID int64 `json:"telegram_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	user, err := app.getUserByTelegramID(req.TelegramID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil || user.Role != "contractor" {
		http.Error(w, "Пользователь не является бригадиром", http.StatusBadRequest)
		return
	}
	if err := app.completeOrder(orderID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	order, err := app.getOrder(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"order": order})
}

func (app *App) handleRejectOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		http.Error(w, "Неверный order ID", http.StatusBadRequest)
		return
	}
	if err := app.cancelOrder(orderID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// Handler - экспортированная функция для Vercel
// Обрабатывает только API запросы, статика обслуживается Vercel автоматически из корня проекта
func Handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	// Обрабатываем только API запросы
	if strings.HasPrefix(path, "/api") {
		handleAPI(w, r)
		return
	}
	
	// Все остальные запросы должны обслуживаться Vercel как статика
	// Возвращаем 404 только если это не статика (на случай ошибок)
	http.NotFound(w, r)
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	// Инициализируем БД если нужно
	if err := initDBIfNeeded(); err != nil {
		http.Error(w, "База данных не настроена: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// CORS middleware
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Настройка роутера для API
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/tariffs", app.getTariffs).Methods("GET")
	api.HandleFunc("/user/{telegramId}", app.getUser).Methods("GET")
	api.HandleFunc("/user", app.createOrUpdateUser).Methods("POST")
	api.HandleFunc("/contractor/profile", app.updateContractorProfile).Methods("POST")
	api.HandleFunc("/contractors/search", app.searchContractors).Methods("GET")
	api.HandleFunc("/orders", app.handleCreateOrder).Methods("POST")
	api.HandleFunc("/contractor/orders/{telegramId}", app.handleGetContractorOrders).Methods("GET")
	api.HandleFunc("/contractor/pending-orders/{telegramId}", app.getPendingOrders).Methods("GET")
	api.HandleFunc("/orders/{orderId}/accept", app.handleAcceptOrder).Methods("POST")
	api.HandleFunc("/orders/{orderId}/complete", app.handleCompleteOrder).Methods("POST")
	api.HandleFunc("/orders/{orderId}/reject", app.handleRejectOrder).Methods("POST")

	router.ServeHTTP(w, r)
}
