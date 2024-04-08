package api

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/ArtemZ007/wb-l0/internal/cache"
	"github.com/ArtemZ007/wb-l0/internal/model"
	"github.com/sirupsen/logrus"
)

// OrderHandler структура для HTTP-обработчиков, связанных с заказами.
type OrderHandler struct {
	Cache *cache.Cache
}

// NewOrderHandler создает новый экземпляр OrderHandler с предоставленным кэшем.
func NewOrderHandler(c *cache.Cache) *OrderHandler {
	return &OrderHandler{Cache: c}
}

// handleOrder обрабатывает HTTP-запросы к заказам.
func (h *OrderHandler) handleOrder(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetOrder(w, r)
	case http.MethodPost:
		h.UpdateOrder(w, r)
	default:
		logrus.Warn("Получен запрос с неподдерживаемым методом")
		http.Error(w, "Неподдерживаемый метод", http.StatusMethodNotAllowed)
	}
}

// GetOrder обрабатывает GET-запросы, извлекая заказ по его ID из кэша и отображая его в виде HTML страницы с JSON.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	order, found := h.Cache.GetOrder(orderID)
	if !found {
		logrus.WithField("orderID", orderID).Warn("Заказ не найден в кэше")
		http.NotFound(w, r)
		return
	}

	orderJSON, err := json.MarshalIndent(order, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Ошибка при маршалинге заказа в JSON")
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("order").Parse(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>Заказ</title>
        </head>
        <body>
            <pre>{{.}}</pre>
        </body>
        </html>
    `))

	if err := tmpl.Execute(w, string(orderJSON)); err != nil {
		logrus.WithError(err).Error("Ошибка при рендеринге страницы заказа")
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}

// UpdateOrder обрабатывает POST-запросы для обновления информации о заказе.
func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		logrus.WithError(err).Error("Ошибка при декодировании тела запроса")
		http.Error(w, "Неверное тело запроса", http.StatusBadRequest)
		return
	}

	if err := h.Cache.UpdateOrder(order.OrderUID, &order); err != nil {
		logrus.WithError(err).WithField("orderID", order.OrderUID).Error("Ошибка при обновлении заказа в кэше")
		http.Error(w, "Ошибка при сохранении заказа", http.StatusInternalServerError)
		return
	}

	logrus.WithField("orderID", order.OrderUID).Info("Заказ успешно обновлен в кэше")
	writeJSONResponse(w, order, http.StatusOK)
}

// RegisterRoutes регистрирует маршруты для обработчиков заказов.
func (h *OrderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/orders/", h.handleOrder)
}

// writeJSONResponse отправляет ответ в формате JSON с указанным статус-кодом.
func writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logrus.WithError(err).Error("Ошибка при кодировании ответа в JSON")
	}
}

// writeJSONError отправляет сообщение об ошибке в формате JSON.
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	writeJSONResponse(w, map[string]string{"error": message}, statusCode)
}
