package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type App struct {
	db *sql.DB
}

// Тарифы
var TARIFFS = map[string]Tariff{
	"econom": {
		Name:        "ЭКОНОМ",
		Description: "Мокрая, ручная",
		PriceRange: PriceRange{
			Min: 400,
			Max: 450,
		},
		Days:     "28 дней",
		Features: []string{"Классика", "Низкая цена материалов", "Долгий срок высыхания", "Высокий риск трещин"},
	},
	"comfort": {
		Name:        "КОМФОРТ",
		Description: "Полусухая механизированная",
		PriceRange: PriceRange{
			Min: 550,
			Max: 850,
		},
		Days:     "5-7 дней (плитка — 2 дня, ламинат — 14–20 дней)",
		Features: []string{"Оптимальный баланс", "Минимум усадки", "Можно ходить через 12 часов", "Самый популярный выбор"},
	},
	"business": {
		Name:        "БИЗНЕС",
		Description: "С армированием",
		PriceRange: PriceRange{
			Min: 150,
			Max: 300,
		},
		Days:     "Как у базового тарифа",
		Features: []string{"Повышенная прочность", "Надбавка за армирование сеткой или фиброй"},
		IsAddon:  true,
	},
	"premium": {
		Name:        "ПРЕМИУМ",
		Description: "Сухая стяжка Кнауф",
		PriceRange: PriceRange{
			Min: 800,
			Max: 1000,
		},
		Days:     "1-2 дня",
		Features: []string{"Нет мокрых процессов", "Идеальная геометрия", "Теплоизоляция", "Высокая цена материалов"},
	},
	"universal": {
		Name:        "УНИВЕРСАЛ",
		Description: "Плавающая / Утепленная",
		PriceRange: PriceRange{
			Min: 250,
			Max: 600,
		},
		Days:     "Как у базового тарифа",
		Features: []string{"Зависит от вида утеплителя", "Включает слой изоляции"},
		IsAddon:  true,
	},
	"self-leveling": {
		Name:        "САМОВЫРАВНИВАТЕЛЬ",
		Description: "Финишный слой",
		PriceRange: PriceRange{
			Min: 250,
			Max: 500,
		},
		Days:     "1-3 дня",
		Features: []string{"Финишный слой"},
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

func main() {
	// Загружаем .env файл
	godotenv.Load()

	// Подключение к БД
	databaseURL := os.Getenv("DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL не установлен")
	}

	// Формируем DSN для libsql
	// Turso использует формат: libsql://database-url?authToken=token
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
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	app := &App{db: db}

	// Инициализация БД
	if err := app.initDB(); err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}

	// Настройка роутера
	r := mux.NewRouter()

	// CORS middleware
	r.Use(corsMiddleware)

	// API routes
	api := r.PathPrefix("/api").Subrouter()
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
	api.HandleFunc("/orders/{orderId}/reject", app.rejectOrder).Methods("POST")

	// Отдаем статические файлы из корня проекта (на уровень выше backend/)
	// Статика регистрируется ПОСЛЕ API, чтобы не конфликтовать с /api/*

	// Явно отдаем index.html для корня
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.ServeFile(w, r, "../index.html")
		} else {
			http.NotFound(w, r)
		}
	}).Methods("GET")

	// Отдаем остальные статические файлы (JS, CSS и т.д.)
	r.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../app.js")
	}).Methods("GET")
	r.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../styles.css")
	}).Methods("GET")
	r.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../index.html")
	}).Methods("GET")

	// Отдаем изображения и другие статические файлы
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("../images/")))).Methods("GET")

	// Для всех остальных GET запросов (fallback для статики)
	// Используем NotFoundHandler для перехвата только несуществующих роутов
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Только для GET запросов, которые не начинаются с /api, пытаемся отдать файл из корня
		if r.Method == "GET" && !strings.HasPrefix(r.URL.Path, "/api") {
			filePath := "../" + r.URL.Path
			// Проверяем, что файл существует
			if _, err := os.Stat(filePath); err == nil {
				http.ServeFile(w, r, filePath)
				return
			}
		}
		http.NotFound(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
