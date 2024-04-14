// Package http предоставляет утилиты сервера HTTP для обработки запросов и ответов.
// Это включает в себя маршрутизацию и обслуживание HTTP-запросов, обработку ошибок и кодирование ответов.
//
// Автор: ArtemZ007
package http

import (
	"encoding/json"
	"net/http"

	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
)

// Handler структура обработчика HTTP-запросов. Отвечает за обработку входящих запросов и взаимодействие с кэшем.
type Handler struct {
	cacheService cache.Cache    // Используем интерфейс Cache для улучшения интеграции с сервисом кэширования.
	logger       *logger.Logger // Используем конкретную реализацию Logger для унификации логирования.
}

// NewHandler создает новый экземпляр Handler. Принимает сервис кэширования и экземпляр логгера.
func NewHandler(cacheService cache.Cache, logger *logger.Logger) *Handler {
	return &Handler{
		cacheService: cacheService,
		logger:       logger,
	}
}

// ServeHTTP метод для обработки HTTP-запросов. Определяет маршруты и вызывает соответствующие обработчики.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/order":
		h.handleGetOrder(w, r)
	default:
		h.handleNotFound(w, r)
	}
}

// handleGetOrder обрабатывает запросы на получение заказа по его ID. Возвращает данные заказа в формате JSON.
func (h *Handler) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSONError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		h.writeJSONError(w, "Не указан ID заказа", http.StatusBadRequest)
		return
	}

	order, found := h.cacheService.GetOrder(orderID)
	if !found {
		h.logger.Error("Заказ не найден", map[string]interface{}{"orderID": orderID})
		h.writeJSONError(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	h.logger.Info("Заказ успешно найден и отправлен", map[string]interface{}{"orderID": orderID})
	h.writeJSONResponse(w, order, http.StatusOK)
}

// writeJSONResponse отправляет ответ в формате JSON. Устанавливает необходимые заголовки и сериализует данные.
func (h *Handler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Ошибка при сериализации данных в JSON", map[string]interface{}{"error": err})
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}

// writeJSONError отправляет сообщение об ошибке в формате JSON. Устанавливает необходимые заголовки.
func (h *Handler) writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		h.logger.Error("Ошибка при отправке сообщения об ошибке", map[string]interface{}{"error": err})
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}

// handleNotFound обрабатывает несуществующие маршруты. Возвращает сообщение об ошибке.
func (h *Handler) handleNotFound(w http.ResponseWriter, _ *http.Request) {
	h.writeJSONError(w, "Страница не найдена", http.StatusNotFound)
}
