package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

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
	result, err := app.db.Exec(
		"INSERT INTO users (telegram_id, role, name, phone, avatar_url) VALUES (?, ?, ?, ?, ?)",
		telegramID, role, name, phone, avatarURL,
	)
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
	row := app.db.QueryRow(
		`SELECT cp.id, cp.user_id, cp.experience_years, cp.rating, cp.completed_orders, 
			cp.categories, cp.is_active, cp.current_order_id,
			u.name, u.phone, u.avatar_url, u.telegram_id
		 FROM contractor_profiles cp
		 JOIN users u ON cp.user_id = u.id
		 WHERE u.id = ?`,
		userID,
	)

	var profile ContractorProfile
	err := row.Scan(
		&profile.ID, &profile.UserID, &profile.ExperienceYears, &profile.Rating,
		&profile.CompletedOrders, &profile.Categories, &profile.IsActive, &profile.CurrentOrderID,
		&profile.Name, &profile.Phone, &profile.AvatarURL, &profile.TelegramID,
	)
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

	// Проверяем существование
	row := app.db.QueryRow("SELECT id FROM contractor_profiles WHERE user_id = ?", userID)
	var existingID int64
	err := row.Scan(&existingID)

	if err == nil {
		// Обновляем
		_, err = app.db.Exec(
			`UPDATE contractor_profiles 
			 SET experience_years = ?, categories = ?, is_active = ?
			 WHERE user_id = ?`,
			experienceYears, categoriesStr, isActive, userID,
		)
		return err
	} else if err == sql.ErrNoRows {
		// Создаем
		_, err = app.db.Exec(
			`INSERT INTO contractor_profiles (user_id, experience_years, categories, is_active)
			 VALUES (?, ?, ?, ?)`,
			userID, experienceYears, categoriesStr, isActive,
		)
		return err
	}
	return err
}

func (app *App) getAvailableContractors(category string) ([]ContractorProfile, error) {
	rows, err := app.db.Query(
		`SELECT cp.id, cp.user_id, cp.experience_years, cp.rating, cp.completed_orders,
			cp.categories, cp.is_active, cp.current_order_id,
			u.name, u.phone, u.avatar_url, u.telegram_id
		 FROM contractor_profiles cp
		 JOIN users u ON cp.user_id = u.id
		 WHERE cp.is_active = 1 
		 AND (cp.current_order_id IS NULL OR cp.current_order_id = 0)
		 AND (cp.categories LIKE ? OR cp.categories = '[]')
		 ORDER BY cp.rating DESC, cp.completed_orders DESC
		 LIMIT 10`,
		"%"+category+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contractors []ContractorProfile
	for rows.Next() {
		var profile ContractorProfile
		err := rows.Scan(
			&profile.ID, &profile.UserID, &profile.ExperienceYears, &profile.Rating,
			&profile.CompletedOrders, &profile.Categories, &profile.IsActive, &profile.CurrentOrderID,
			&profile.Name, &profile.Phone, &profile.AvatarURL, &profile.TelegramID,
		)
		if err != nil {
			return nil, err
		}
		contractors = append(contractors, profile)
	}

	return contractors, nil
}

func (app *App) createOrder(clientID int64, category string, area *float64, address *string) (int64, error) {
	result, err := app.db.Exec(
		"INSERT INTO orders (client_id, category, area, address) VALUES (?, ?, ?, ?)",
		clientID, category, area, address,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}

func (app *App) getOrder(orderID int64) (*Order, error) {
	row := app.db.QueryRow(
		`SELECT o.id, o.client_id, o.contractor_id, o.category, o.area, o.address,
			o.status, o.created_at, o.accepted_at, o.completed_at,
			uc.name, uc.telegram_id,
			uct.name, uct.telegram_id
		 FROM orders o
		 LEFT JOIN users uc ON o.client_id = uc.id
		 LEFT JOIN users uct ON o.contractor_id = uct.id
		 WHERE o.id = ?`,
		orderID,
	)

	var order Order
	var createdAt, acceptedAt, completedAt sql.NullString
	err := row.Scan(
		&order.ID, &order.ClientID, &order.ContractorID, &order.Category,
		&order.Area, &order.Address, &order.Status,
		&createdAt, &acceptedAt, &completedAt,
		&order.ClientName, &order.ClientTelegramID,
		&order.ContractorName, &order.ContractorTelegramID,
	)
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
	rows, err := app.db.Query(
		`SELECT o.id, o.client_id, o.contractor_id, o.category, o.area, o.address,
			o.status, o.created_at, o.accepted_at, o.completed_at,
			u.name, u.telegram_id
		 FROM orders o
		 JOIN users u ON o.client_id = u.id
		 WHERE o.contractor_id = ?
		 ORDER BY o.created_at DESC`,
		contractorID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		var createdAt, acceptedAt, completedAt sql.NullString
		err := rows.Scan(
			&order.ID, &order.ClientID, &order.ContractorID, &order.Category,
			&order.Area, &order.Address, &order.Status,
			&createdAt, &acceptedAt, &completedAt,
			&order.ClientName, &order.ClientTelegramID,
		)
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
	rows, err := app.db.Query(
		`SELECT o.id, o.client_id, o.contractor_id, o.category, o.area, o.address,
			o.status, o.created_at, o.accepted_at, o.completed_at,
			u.name, u.telegram_id
		 FROM orders o
		 JOIN users u ON o.client_id = u.id
		 WHERE o.status = 'pending'
		 ORDER BY o.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		var createdAt, acceptedAt, completedAt sql.NullString
		err := rows.Scan(
			&order.ID, &order.ClientID, &order.ContractorID, &order.Category,
			&order.Area, &order.Address, &order.Status,
			&createdAt, &acceptedAt, &completedAt,
			&order.ClientName, &order.ClientTelegramID,
		)
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
	// Используем batch для атомарности
	queries := []string{
		`UPDATE orders 
		 SET contractor_id = ?, status = 'accepted', accepted_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND status = 'pending'`,
		`UPDATE contractor_profiles 
		 SET current_order_id = ?
		 WHERE user_id = ?`,
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
	// Получаем contractor_id
	row := app.db.QueryRow("SELECT contractor_id FROM orders WHERE id = ?", orderID)
	var contractorID sql.NullInt64
	err := row.Scan(&contractorID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	var contractorIDPtr *int64
	if contractorID.Valid {
		id := contractorID.Int64
		contractorIDPtr = &id
	}

	// Обновляем заказ
	_, err = app.db.Exec(
		`UPDATE orders 
		 SET status = 'completed', completed_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		orderID,
	)
	if err != nil {
		return err
	}

	// Освобождаем бригадира
	if contractorIDPtr != nil {
		_, err = app.db.Exec(
			`UPDATE contractor_profiles 
			 SET current_order_id = NULL,
			     completed_orders = completed_orders + 1
			 WHERE user_id = ?`,
			*contractorIDPtr,
		)
		return err
	}

	return nil
}

func (app *App) cancelOrder(orderID int64) error {
	_, err := app.db.Exec("UPDATE orders SET status = 'cancelled' WHERE id = ?", orderID)
	return err
}
