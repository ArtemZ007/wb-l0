package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ArtemZ007/wb-l0/internal/cache"
	"github.com/ArtemZ007/wb-l0/internal/model"
)

// Handler структура для HTTP-обработчиков с ссылкой на кэш и базу данных.
type Handler struct {
	Cache *cache.Cache
	DB    *sql.DB // Добавляем объект базы данных
}

// NewHandler создает новый экземпляр Handler с кэшем и базой данных.
func NewHandler(c *cache.Cache, db *sql.DB) *Handler {
	return &Handler{
		Cache: c,
		DB:    db,
	}
}

// GetOrder обрабатывает GET-запросы, извлекая заказ по его ID из кэша.
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	order, found := h.Cache.GetOrder(orderID)
	if !found {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Ошибка при формировании ответа", http.StatusInternalServerError)
	}
}

// UpdateOrder обрабатывает POST-запросы для обновления информации о заказе.
func (h *Handler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	order.OrderUID = orderID // Установка ID заказа из URL

	// Исправленный вызов с передачей объекта базы данных
	if err := h.Cache.UpdateOrder(h.DB, orderID, &order); err != nil {
		http.Error(w, "Ошибка при обновлении заказа", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Ошибка при формировании ответа", http.StatusInternalServerError)
	}
}

// RegisterRoutes регистрирует маршруты для обработчиков.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/orders/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/orders/")
		if id == "" && r.Method == http.MethodPost {
			h.UpdateOrder(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.GetOrder(w, r)
		case http.MethodPost:
			h.UpdateOrder(w, r)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	// Обработка статических файлов и корневого пути остается без изменений
	staticFilesPath := "../web/static/"
	fileServer := http.FileServer(http.Dir(staticFilesPath))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			path := filepath.Join(staticFilesPath, r.URL.Path)
			if strings.HasSuffix(r.URL.Path, ".js") || strings.HasSuffix(r.URL.Path, ".css") || strings.HasSuffix(r.URL.Path, ".json") {
				http.ServeFile(w, r, path)
				return
			}
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticFilesPath, "index.html"))
	})
}
