package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ArtemZ007/cache"
	"github.com/ArtemZ007/model"
)

// Handler структура для HTTP-обработчиков с ссылкой на кэш
type Handler struct {
	Cache *cache.Cache
}

// NewHandler создает новый экземпляр Handler
func NewHandler(c *cache.Cache) *Handler {
	return &Handler{Cache: c}
}

// GetOrder обрабатывает запросы на получение заказа по ID
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Извлечение ID заказа из URL
	orderID := strings.TrimPrefix(r.URL.Path, "/orders/")

	// Поиск заказа в кэше
	order, found := h.Cache.GetOrder(orderID)
	if !found {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Возврат заказа в ответе
	response, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Failed to marshal order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// UpdateOrder обрабатывает запросы на обновление заказа
func (h *Handler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	// Извлечение ID заказа из URL
	orderID := strings.TrimPrefix(r.URL.Path, "/orders/")

	// Декодирование полученного заказа из запроса
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid order data", http.StatusBadRequest)
		return
	}

	// Обновление заказа в кэше и базе данных
	if err := h.Cache.UpdateOrder(orderID, &order); err != nil {
		http.Error(w, "Failed to update order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RegisterRoutes регистрирует маршруты для обработчиков
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/orders/", h.GetOrder)
	mux.HandleFunc("/orders/update/", h.UpdateOrder)
}
