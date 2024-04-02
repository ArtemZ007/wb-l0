package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/ArtemZ007/wb-l0/internal/cache"
	"github.com/ArtemZ007/wb-l0/internal/model"
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
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	orderIDStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	if orderIDStr == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	// Преобразование orderID из строки в int
	// orderID, err := orderIDStr
	// if err != nil {
	// 	http.Error(w, "Invalid order ID format", http.StatusBadRequest)
	// 	return
	// }

	order, found := h.Cache.GetOrder(orderIDStr)
	if !found {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	response, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Failed to marshal order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(response); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// UpdateOrder обрабатывает запросы на обновление заказа
func (h *Handler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { // Или http.MethodPut, если вы предпочитаете использовать PUT для обновлений
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	orderIDStr := strings.TrimPrefix(r.URL.Path, "/orders/update/")
	if orderIDStr == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	// Преобразование orderID из строки в int
	// orderID, err := strconv.Atoi(orderIDStr)
	// if err != nil {
	// 	http.Error(w, "Invalid order ID format", http.StatusBadRequest)
	// 	return
	// }

	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid order data", http.StatusBadRequest)
		return
	}

	if err := h.Cache.UpdateOrder(orderIDStr, &order); err != nil {
		http.Error(w, "Failed to update order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RegisterRoutes регистрирует маршруты для обработчиков
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/update/") {
			h.UpdateOrder(w, r)
		} else {
			h.GetOrder(w, r)
		}
	})
}
