package httpQS

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
)

// DataService интерфейс, определяющий методы для работы с данными.
// Этот интерфейс должен быть реализован сервисом, который занимается получением данных.
type DataService interface {
	GetData() ([]model.Order, error)
	GetOrder(orderUID string) (*model.Order, error)
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
		h.handleOrder(w, r)
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

	// Генерация HTML-страницы
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<html><head><title>Orders</title></head><body>"))
	w.Write([]byte("<h1>Orders</h1>"))
	w.Write([]byte("<table border='1'><tr><th>OrderUID</th><th>TrackNumber</th><th>Entry</th></tr>"))

	for _, order := range data {
		w.Write([]byte("<tr>"))
		w.Write([]byte("<td><a href='/order?uid=" + order.OrderUID + "'>" + order.OrderUID + "</a></td>"))
		w.Write([]byte("<td>" + order.TrackNumber + "</td>"))
		w.Write([]byte("<td>" + order.Entry + "</td>"))
		w.Write([]byte("</tr>"))
	}

	w.Write([]byte("</table></body></html>"))
}

// handleOrder метод для обработки запросов к конкретному заказу.
func (h *Handler) handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSONError(w, "Неподдерживаемый метод", http.StatusMethodNotAllowed)
		return
	}

	orderUID := r.URL.Query().Get("uid")
	if orderUID == "" {
		h.writeJSONError(w, "UID заказа не указан", http.StatusBadRequest)
		return
	}

	order, err := h.dataService.GetOrder(orderUID)
	if err != nil {
		h.logger.Error("Ошибка при получении заказа: ", err)
		h.writeJSONError(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if order == nil {
		h.writeJSONError(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	// Генерация HTML-страницы для конкретного заказа
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<html><head><title>Order Details</title></head><body>"))
	w.Write([]byte("<h1>Order Details</h1>"))
	w.Write([]byte("<p><strong>OrderUID:</strong> " + order.OrderUID + "</p>"))
	w.Write([]byte("<p><strong>TrackNumber:</strong> " + order.TrackNumber + "</p>"))
	w.Write([]byte("<p><strong>Entry:</strong> " + order.Entry + "</p>"))
	w.Write([]byte("<h2>Delivery</h2>"))
	w.Write([]byte("<p><strong>Name:</strong> " + order.Delivery.Name + "</p>"))
	w.Write([]byte("<p><strong>Phone:</strong> " + order.Delivery.Phone + "</p>"))
	w.Write([]byte("<p><strong>Zip:</strong> " + order.Delivery.Zip + "</p>"))
	w.Write([]byte("<p><strong>City:</strong> " + order.Delivery.City + "</p>"))
	w.Write([]byte("<p><strong>Address:</strong> " + order.Delivery.Address + "</p>"))
	w.Write([]byte("<p><strong>Region:</strong> " + order.Delivery.Region + "</p>"))
    w.Write([]byte("<p><strong>Email:</strong> " + *order.Delivery.Email + "</p>"))
    w.Write([]byte("<h2>Payment</h2>"))
    w.Write([]byte("<p><strong>Transaction:</strong> " + order.Payment.Transaction + "</p>"))
    w.Write([]byte("<p><strong>RequestID:</strong> " + order.Payment.RequestID + "</p>"))
    w.Write([]byte("<p><strong>Currency:</strong> " + order.Payment.Currency + "</p>"))
    w.Write([]byte("<p><strong>Provider:</strong> " + order.Payment.Provider + "</p>"))
    w.Write([]byte("<p><strong>Amount:</strong> " + fmt.Sprintf("%d", order.Payment.Amount) + "</p>"))
    w.Write([]byte("<p><strong>PaymentDT:</strong> " + fmt.Sprintf("%d", order.Payment.PaymentDT) + "</p>"))
    w.Write([]byte("<p><strong>Bank:</strong> " + *order.Payment.Bank + "</p>"))
    w.Write([]byte("<p><strong>DeliveryCost:</strong> " + fmt.Sprintf("%d", order.Payment.DeliveryCost) + "</p>"))
    w.Write([]byte("<p><strong>GoodsTotal:</strong> " + fmt.Sprintf("%d", order.Payment.GoodsTotal) + "</p>"))
    w.Write([]byte("<p><strong>CustomFee:</strong> " + fmt.Sprintf("%d", order.Payment.CustomFee) + "</p>"))
    w.Write([]byte("<h2>Items</h2>"))
    w.Write([]byte("<ul>"))
    for _, item := range order.Items {
        w.Write([]byte("<li>"))
        w.Write([]byte("<p><strong>ChrtID:</strong> " + fmt.Sprintf("%d", item.ChrtID) + "</p>"))
        w.Write([]byte("<p><strong>TrackNumber:</strong> " + *item.TrackNumber + "</p>"))
        w.Write([]byte("<p><strong>Price:</strong> " + fmt.Sprintf("%d", item.Price) + "</p>"))
        w.Write([]byte("<p><strong>Name:</strong> " + *item.Name + "</p>"))
        w.Write([]byte("<p><strong>Sale:</strong> " + fmt.Sprintf("%d", item.Sale) + "</p>"))
        w.Write([]byte("<p><strong>Size:</strong> " + *item.Size + "</p>"))
        w.Write([]byte("<p><strong>TotalPrice:</strong> " + fmt.Sprintf("%d", item.TotalPrice) + "</p>"))
        w.Write([]byte("<p><strong>Brand:</strong> " + *item.Brand + "</p>"))
        w.Write([]byte("<p><strong>Status:</strong> " + fmt.Sprintf("%d", item.Status) + "</p>"))
        w.Write([]byte("</li>"))
    }
    w.Write([]byte("</ul>"))
    w.Write([]byte("</body></html>"))

// writeJSONResponse метод для отправки ответа в формате JSON.
func (h *Handler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Ошибка при кодировании JSON ответа: ", err)
	}
}

// writeJSONError метод для отправки сообщения об ошибке в формате JSON.
func (h *Handler) writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	h.writeJSONResponse(w, map[string]string{"error": message}, statusCode)
}
