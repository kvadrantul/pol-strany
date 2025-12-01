package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":    user,
		"profile": profile,
	})
}

func (app *App) createOrUpdateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TelegramID int64   `json:"telegram_id"`
		Role       string  `json:"role"`
		Name       *string  `json:"name"`
		Phone      *string  `json:"phone"`
		AvatarURL  *string  `json:"avatar_url"`
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
		// Создаем нового пользователя
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
		// Обновляем существующего
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
		TelegramID     int64    `json:"telegram_id"`
		ExperienceYears *int    `json:"experience_years"`
		Categories     []string `json:"categories"`
		IsActive       bool     `json:"is_active"`
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

	// Парсим JSON категории в массив для каждого подрядчика
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

func (app *App) rejectOrder(w http.ResponseWriter, r *http.Request) {
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

