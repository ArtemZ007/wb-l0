package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ArtemZ007/wb-l0/internal/cache"
	"github.com/ArtemZ007/wb-l0/internal/model"
)

// OrderHandler структура для HTTP-обработчиков, связанных с заказами.
type OrderHandler struct {
	Cache *cache.Cache
}

// NewOrderHandler создает новый экземпляр OrderHandler.
func NewOrderHandler(c *cache.Cache) *OrderHandler {
	return &OrderHandler{Cache: c}
}

// handleOrder обрабатывает запросы к заказам.
func (h *OrderHandler) handleOrder(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetOrder(w, r)
	case http.MethodPost:
		h.UpdateOrder(w, r)
	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}

// GetOrder обрабатывает GET-запросы на получение заказа по ID.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/order/")
	order, found := h.Cache.Get(orderID)
	if !found {
		writeJSONError(w, "Order not found", http.StatusNotFound)
		return
	}

	writeJSONResponse(w, order, http.StatusOK)
}

// UpdateOrder обрабатывает POST-запросы на обновление заказа.
func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.Cache.Set(order.OrderUID, &order)
	writeJSONResponse(w, order, http.StatusOK)
}

// RegisterRoutes регистрирует маршруты для обработчиков.
func (h *OrderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/order/", h.handleOrder) // Используется один обработчик для всех методов
}

// writeJSONResponse упрощает отправку JSON ответов.
func writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeJSONError упрощает отправку JSON ошибок.
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	writeJSONResponse(w, map[string]string{"error": message}, statusCode)
}
