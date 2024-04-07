package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ArtemZ007/wb-l0/internal/cache"
	"github.com/ArtemZ007/wb-l0/internal/model"
)

// Handler структура для HTTP-обработчиков с ссылкой на кэш.
type Handler struct {
	Cache *cache.Cache
}

// NewHandler создает новый экземпляр Handler.
func NewHandler(c *cache.Cache) *Handler {
	return &Handler{Cache: c}
}

// GetOrder обрабатывает GET-запросы, извлекая заказ по его ID из кэша.
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/orders/")
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
	orderID := strings.TrimPrefix(r.URL.Path, "/orders/")
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	order.OrderUID = orderID // Установка ID заказа из URL

	if err := h.Cache.UpdateOrder(orderID, &order); err != nil {
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
	// Обновление пути для соответствия ожиданиям клиентского кода
	mux.HandleFunc("/api/orders/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/orders/")
		if id == "" && r.Method != http.MethodPost { // Для POST запросов ID может быть не указан в URL
			http.Error(w, "ID заказа не указан", http.StatusBadRequest)
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
