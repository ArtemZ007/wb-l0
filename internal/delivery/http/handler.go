package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/sirupsen/logrus"
)

// OrderHandler структура для HTTP-обработчиков, связанных с заказами.
type OrderHandler struct {
	Cache *cache.Cache
	DB    *sql.DB
}

// NewOrderHandler создает новый экземпляр OrderHandler с предоставленным кэшем и базой данных.
func NewOrderHandler(c *cache.Cache, db *sql.DB) *OrderHandler {
	return &OrderHandler{
		Cache: c,
		DB:    db,
	}
}

// RegisterRoutes регистрирует маршруты для обработчиков заказов.
func (h *OrderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/order/", h.handleOrder)
}

func (h *OrderHandler) handleOrder(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetOrder(w, r)
	default:
		logrus.Warn("Получен запрос с неподдерживаемым методом")
		http.Error(w, "Неподдерживаемый метод", http.StatusMethodNotAllowed)
	}
}

// GetOrder обрабатывает GET-запросы для получения данных о заказе по ID.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/api/order/")
	order, found := h.Cache.GetOrder(orderID)
	if !found {
		// Если заказ не найден в кэше, пытаемся извлечь его из базы данных
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var err error
		order, err = db.GetOrderByID(ctx, h.DB, orderID)
		if err != nil {
			logrus.WithField("orderID", orderID).WithError(err).Warn("Заказ не найден")
			writeJSONError(w, "Заказ не найден", http.StatusNotFound)
			return
		}

		// Добавляем заказ в кэш после успешного извлечения из базы данных
		h.Cache.SetOrder(orderID, order)
	}

	logrus.WithField("orderID", orderID).Info("Заказ успешно извлечен")
	writeJSONResponse(w, order, http.StatusOK)
}

func writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logrus.WithError(err).Error("Ошибка при кодировании ответа в JSON")
	}
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	writeJSONResponse(w, map[string]string{"error": message}, statusCode)
}
