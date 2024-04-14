package httpQS

import (
	"encoding/json"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"net/http"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
)

// DataService интерфейс, определяющий методы для работы с данными.
// Этот интерфейс должен быть реализован сервисом, который занимается получением данных.
type DataService interface {
	GetData() ([]model.Order, error)
}

// Handler структура обработчика HTTP-запросов.
type Handler struct {
	dataService DataService    // Сервис для работы с данными
	logger      *logger.Logger // Логгер для регистрации событий
}

// NewHandler функция для создания нового экземпляра Handler.
// Принимает в качестве аргументов реализацию интерфейса DataService и экземпляр логгера.
func NewHandler(dataService cache.Cache, logger *logger.Logger) *Handler {
	return &Handler{
		dataService: dataService,
		logger:      logger,
	}
}

// ServeHTTP метод для обработки HTTP-запросов.
// Реализует интерфейс http.Handler, что позволяет использовать экземпляры Handler как обработчики в HTTP-сервере.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.handleIndex(w, r)
	default:
		h.handleNotFound(w, r)
	}
}

// handleIndex метод для обработки запросов к корневому маршруту.
func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSONError(w, "Неподдерживаемый метод", http.StatusMethodNotAllowed)
		return
	}

	data, err := h.dataService.GetData()
	if err != nil {
		h.logger.Error("Ошибка при получении данных: ", err)
		h.writeJSONError(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, data, http.StatusOK)
}

// writeJSONResponse метод для отправки ответа в формате JSON.
func (h *Handler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Не удалось закодировать ответ в JSON: ", err)
	}
}

// writeJSONError метод для отправки сообщения об ошибке в формате JSON.
func (h *Handler) writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		h.logger.Error("Ошибка при кодировании ошибки в JSON: ", err)
	}
}

// handleNotFound метод для обработки несуществующих маршрутов.
func (h *Handler) handleNotFound(w http.ResponseWriter, _ *http.Request) {
	h.writeJSONError(w, "Не найдено", http.StatusNotFound)
}
