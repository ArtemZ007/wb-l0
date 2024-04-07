package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ArtemZ007/wb-l0/internal/cache"
)

// Handler структура для HTTP-обработчиков с ссылкой на кэш.
type Handler struct {
	Cache *cache.Cache
}

// NewHandler создает новый экземпляр Handler.
func NewHandler(c *cache.Cache) *Handler {
	return &Handler{Cache: c}
}

// GetOrder обрабатывает запросы на получение заказа по ID.
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Реализация обработчика
}

// UpdateOrder обрабатывает запросы на обновление заказа.
func (h *Handler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	// Реализация обработчика
}

// RegisterRoutes регистрирует маршруты для обработчиков.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Регистрация обработчиков для конкретных методов и путей
	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetOrder(w, r)
		case http.MethodPost:
			h.UpdateOrder(w, r)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	// Обработка статических файлов
	staticFilesPath := "../web/static/"
	fileServer := http.FileServer(http.Dir(staticFilesPath))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Добавление обработчика для корневого пути для отдачи index.html
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// Обработка запросов к статическим файлам
			path := filepath.Join(staticFilesPath, r.URL.Path)
			if strings.HasSuffix(r.URL.Path, ".js") || strings.HasSuffix(r.URL.Path, ".css") || strings.HasSuffix(r.URL.Path, ".json") {
				http.ServeFile(w, r, path)
				return
			}
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticFilesPath, "index.html"))
	})
}

// NewRouter создает и возвращает новый экземпляр роутера с настроенными маршрутами.
func NewRouter(c *cache.Cache) *http.ServeMux {
	handler := NewHandler(c)
	router := http.NewServeMux()
	handler.RegisterRoutes(router)
	return router
}
