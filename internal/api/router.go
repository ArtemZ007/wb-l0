package api

import (
	"net/http"

	"github.com/ArtemZ007/wb-l0/internal/cache"
)

// NewRouter создает и возвращает новый экземпляр роутера с настроенными маршрутами.
func NewRouter(c *cache.Cache) *http.ServeMux {
	router := http.NewServeMux()
	// Здесь добавьте маршруты, например:
	// router.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
	//     // Обработка запроса
	// })
	return router
}
