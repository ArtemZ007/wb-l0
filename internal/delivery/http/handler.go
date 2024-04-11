package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/internal/repository/db"
	"github.com/sirupsen/logrus"
)

// Server структура для HTTP-сервера с конфигурацией и зависимостями.
type Server struct {
	port   string
	server *http.Server
	logger *logrus.Logger
}

// MyHandler структура для обработчика HTTP-запросов с зависимостями.
type MyHandler struct {
	CacheService *cache.Cache
	DBService    *db.DBService
	Logger       *logrus.Logger
}

// NewServer создает новый HTTP-сервер с заданной конфигурацией.
func NewServer(port string, handler http.Handler, logger *logrus.Logger) *Server {
	return &Server{
		port: port,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: handler,
		},
		logger: logger,
	}
}

// Start запускает HTTP-сервер.
func (s *Server) Start(ctx context.Context) {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatalf("Не удалось начать прослушивание на порту %s: %v", s.port, err)
		}
	}()
	s.logger.Infof("Сервер запущен на порту %s", s.port)

	<-ctx.Done()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctxShutdown); err != nil {
		s.logger.Fatalf("Не удалось корректно завершить работу сервера: %+v", err)
	}
	s.logger.Info("Сервер корректно завершил работу")
}

// NewHandler создает новый экземпляр MyHandler с кэшем, базой данных и логгером.
func NewHandler(cacheService *cache.Cache, dbService *db.DBService, logger *logrus.Logger) *MyHandler {
	return &MyHandler{
		CacheService: cacheService,
		DBService:    dbService,
		Logger:       logger,
	}
}

// RegisterRoutes регистрирует маршруты для обработчика.
func (h *MyHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/order/", h.handleOrder)
	// Дополнительные маршруты могут быть зарегистрированы здесь
}

// handleOrder обрабатывает запросы к /api/order/.
func (h *MyHandler) handleOrder(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetOrder(w, r)
	default:
		h.Logger.Warn("Получен запрос с неподдерживаемым методом")
		writeJSONError(w, "Неподдерживаемый метод", http.StatusMethodNotAllowed, h.Logger)
	}
}

// GetOrder обрабатывает GET-запросы для получения данных о заказе по ID.
func (h *MyHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/api/order/")
	order, found := h.CacheService.GetOrder(orderID)
	if !found {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var err error
		order, err = h.DBService.GetOrderByID(ctx, orderID)
		if err != nil {
			h.Logger.WithField("orderID", orderID).WithError(err).Warn("Заказ не найден")
			writeJSONError(w, "Заказ не найден", http.StatusNotFound, h.Logger)
			return
		}

		h.CacheService.AddOrUpdateOrder(order)
	}

	h.Logger.WithField("orderID", orderID).Info("Заказ успешно извлечен")
	writeJSONResponse(w, order, http.StatusOK, h.Logger)
}

func writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int, logger *logrus.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.WithError(err).Error("Ошибка при кодировании ответа в JSON")
	}
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int, logger *logrus.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		logger.WithError(err).Error("Ошибка при кодировании ошибки в JSON")
	}
}
