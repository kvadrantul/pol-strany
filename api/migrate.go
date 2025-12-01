package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type ContractorData struct {
	Name       string
	Phone      string
	Experience int
	Rating     float64
	Orders     int
	Category   string
	TelegramID int64
}

var contractorsData = []ContractorData{
	// Простой класс (econom) - 5 бригад
	{"Иван Петров", "+7 (999) 111-22-33", 3, 4.2, 12, "econom", 1001},
	{"Сергей Сидоров", "+7 (999) 111-22-34", 2, 4.0, 8, "econom", 1002},
	{"Дмитрий Козлов", "+7 (999) 111-22-35", 4, 4.3, 15, "econom", 1003},
	{"Алексей Новиков", "+7 (999) 111-22-36", 1, 3.9, 5, "econom", 1004},
	{"Михаил Волков", "+7 (999) 111-22-37", 3, 4.1, 10, "econom", 1005},

	// Комфорт класс - 5 бригад
	{"Андрей Соколов", "+7 (999) 222-33-44", 5, 4.6, 28, "comfort", 2001},
	{"Роман Лебедев", "+7 (999) 222-33-45", 6, 4.7, 32, "comfort", 2002},
	{"Николай Орлов", "+7 (999) 222-33-46", 4, 4.5, 25, "comfort", 2003},
	{"Павел Морозов", "+7 (999) 222-33-47", 7, 4.8, 40, "comfort", 2004},
	{"Владимир Смирнов", "+7 (999) 222-33-48", 5, 4.6, 30, "comfort", 2005},

	// Бизнес класс - 4 бригады
	{"Александр Федоров", "+7 (999) 333-44-55", 8, 4.9, 55, "business", 3001},
	{"Евгений Медведев", "+7 (999) 333-44-56", 9, 5.0, 62, "business", 3002},
	{"Игорь Попов", "+7 (999) 333-44-57", 7, 4.8, 48, "business", 3003},
	{"Валерий Степанов", "+7 (999) 333-44-58", 10, 5.0, 70, "business", 3004},

	// Премиум класс - 3 бригады
	{"Виктор Николаев", "+7 (999) 444-55-66", 12, 5.0, 85, "premium", 4001},
	{"Геннадий Павлов", "+7 (999) 444-55-67", 15, 5.0, 95, "premium", 4002},
	{"Юрий Макаров", "+7 (999) 444-55-68", 11, 4.9, 78, "premium", 4003},

	// Универсал - 2 бригады
	{"Олег Захаров", "+7 (999) 555-66-77", 6, 4.7, 35, "universal", 5001},
	{"Константин Белов", "+7 (999) 555-66-78", 8, 4.8, 42, "universal", 5002},

	// Самовыравниватель - 1 бригада
	{"Станислав Романов", "+7 (999) 666-77-88", 5, 4.6, 28, "self-leveling", 6001},
}

func (app *App) handleMigrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var added, updated, errors int
	results := []string{}

	for _, contractor := range contractorsData {
		// Проверяем, существует ли пользователь
		var userID int64
		err := app.db.QueryRow("SELECT id FROM users WHERE telegram_id = ?", contractor.TelegramID).Scan(&userID)

		if err == sql.ErrNoRows {
			// Создаем пользователя
			name := contractor.Name
			phone := contractor.Phone
			result, err := app.db.Exec(
				"INSERT INTO users (telegram_id, role, name, phone) VALUES (?, ?, ?, ?)",
				contractor.TelegramID, "contractor", name, phone,
			)
			if err != nil {
				errors++
				results = append(results, fmt.Sprintf("❌ Ошибка создания пользователя %s: %v", contractor.Name, err))
				continue
			}

			userID, err = result.LastInsertId()
			if err != nil {
				errors++
				results = append(results, fmt.Sprintf("❌ Ошибка получения ID пользователя %s: %v", contractor.Name, err))
				continue
			}
		} else if err != nil {
			errors++
			results = append(results, fmt.Sprintf("❌ Ошибка проверки пользователя %s: %v", contractor.Name, err))
			continue
		}

		// Создаем категории в формате JSON
		categories := []string{contractor.Category}
		categoriesJSON, _ := json.Marshal(categories)

		// Проверяем, существует ли профиль
		var profileID int64
		err = app.db.QueryRow("SELECT id FROM contractor_profiles WHERE user_id = ?", userID).Scan(&profileID)

		if err == sql.ErrNoRows {
			// Создаем профиль
			_, err = app.db.Exec(
				`INSERT INTO contractor_profiles (user_id, experience_years, rating, completed_orders, categories, is_active)
				 VALUES (?, ?, ?, ?, ?, ?)`,
				userID, contractor.Experience, contractor.Rating, contractor.Orders, string(categoriesJSON), true,
			)
			if err != nil {
				errors++
				results = append(results, fmt.Sprintf("❌ Ошибка создания профиля для %s: %v", contractor.Name, err))
				continue
			}
			added++
			results = append(results, fmt.Sprintf("✅ Добавлен: %s (%s)", contractor.Name, contractor.Category))
		} else if err != nil {
			errors++
			results = append(results, fmt.Sprintf("❌ Ошибка проверки профиля для %s: %v", contractor.Name, err))
		} else {
			// Обновляем профиль
			_, err = app.db.Exec(
				`UPDATE contractor_profiles 
				 SET experience_years = ?, rating = ?, completed_orders = ?, categories = ?, is_active = ?
				 WHERE user_id = ?`,
				contractor.Experience, contractor.Rating, contractor.Orders, string(categoriesJSON), true, userID,
			)
			if err != nil {
				errors++
				results = append(results, fmt.Sprintf("❌ Ошибка обновления профиля для %s: %v", contractor.Name, err))
				continue
			}
			updated++
			results = append(results, fmt.Sprintf("🔄 Обновлен: %s (%s)", contractor.Name, contractor.Category))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": errors == 0,
		"added":   added,
		"updated": updated,
		"errors":  errors,
		"total":   len(contractorsData),
		"results": results,
	})
}

