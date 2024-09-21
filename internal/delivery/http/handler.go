package httpQS

import (
	"encoding/json"
	"net/http"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
	contentTypeHTML   = "text/html"
	serverErrorMsg    = "Внутренняя ошибка сервера"
)

// Handler представляет HTTP обработчик
type Handler struct {
	dataService DataService   // Сервис для работы с данными
	logger      logger.Logger // Логгер для регистрации событий
}

// NewHandler создает новый экземпляр HTTP обработчика
func NewHandler(dataService DataService, logger logger.Logger) *Handler {
	return &Handler{
		dataService: dataService,
		logger:      logger,
	}
}

// handleOrder обрабатывает запросы на получение заказа
func (h *Handler) handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSONError(w, "Неподдерживаемый метод", http.StatusMethodNotAllowed)
		return
	}

	orderUID := r.URL.Query().Get("uid")
	if orderUID == "" {
		h.writeJSONError(w, "Отсутствует параметр uid", http.StatusBadRequest)
		return
	}

	order, err := h.dataService.GetOrder(orderUID)
	if err != nil {
		h.logger.Error("Ошибка при получении заказа: ", err)
		h.writeJSONError(w, serverErrorMsg, http.StatusInternalServerError)
		return
	}

	if order == nil {
		h.writeJSONError(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		h.logger.Error("Ошибка при кодировании ответа: ", err)
		h.writeJSONError(w, serverErrorMsg, http.StatusInternalServerError)
	}
}

// writeJSONError записывает ошибку в формате JSON в ответ
func (h *Handler) writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		h.logger.Error("Ошибка при кодировании ошибки: ", err)
	}
}

// DataService интерфейс, определяющий методы для работы с данными.
type DataService interface {
	GetData() ([]model.Order, bool)
	GetOrder(orderUID string) (*model.Order, error)
}

// Service структура, реализующая интерфейс DataService.
type Service struct {
	cache  map[string]*model.Order // Кэш для хранения заказов
	logger logger.Logger           // Логгер для регистрации событий
}

// NewService функция для создания нового экземпляра Service.
func NewService(logger logger.Logger) *Service {
	return &Service{
		cache:  make(map[string]*model.Order),
		logger: logger,
	}
}

// GetData метод для получения всех заказов.
func (s *Service) GetData() ([]model.Order, bool) {
	var orders []model.Order
	for _, order := range s.cache {
		orders = append(orders, *order)
	}
	if len(orders) == 0 {
		return nil, false
	}
	return orders, true
}

// GetOrder метод для получения заказа по его уникальному идентификатору.
func (s *Service) GetOrder(orderUID string) (*model.Order, error) {
	order, exists := s.cache[orderUID]
	if !exists {
		return nil, nil
	}
	return order, nil
}

// ServeHTTP метод для обработки HTTP-запросов.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.handleIndex(w, r)
	default:
		h.handleOrder(w, r)
	}
}

// handleIndex метод для обработки запросов к корневому маршруту.
func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSONError(w, "Неподдерживаемый метод", http.StatusMethodNotAllowed)
		return
	}

	data, ok := h.dataService.GetData()
	if !ok {
		h.logger.Error("Ошибка при получении данных")
		h.writeJSONError(w, serverErrorMsg, http.StatusInternalServerError)
		return
	}

	// Генерация HTML-страницы
	w.Header().Set(contentTypeHeader, contentTypeHTML)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<html><head><title>Orders</title></head><body>"))
	w.Write([]byte("<h1>Заказы</h1>"))
	w.Write([]byte("<table border='1'><tr><th>OrderUID</th><th>TrackNumber</th><th>Entry</th></tr>"))

	for _, order := range data {
		w.Write([]byte("<tr>"))
		w.Write([]byte("<td><a href='/order?uid=" + order.OrderUID + "'>" + order.OrderUID + "</a></td>"))
		if order.TrackNumber != nil && *order.TrackNumber != "" {
			w.Write([]byte("<td>" + *order.TrackNumber + "</td>"))
		} else {
			w.Write([]byte("<td></td>"))
		}
		if order.Entry != nil && *order.Entry != "" {
			w.Write([]byte("<td>" + *order.Entry + "</td>"))
		} else {
			w.Write([]byte("<td></td>"))
		}
		w.Write([]byte("</tr>"))
	}

	w.Write([]byte("</table></body></html>"))
}
