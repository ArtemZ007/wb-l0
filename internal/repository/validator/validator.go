package validator

import (
	"encoding/json"
	"log"
	"regexp"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
)

// ValidationError представляет ошибку валидации для поля.
type ValidationError struct {
	Field   string `json:"field"`   // Поле, в котором произошла ошибка
	Message string `json:"message"` // Сообщение об ошибке
}

// Service предоставляет методы для валидации моделей.
type Service struct {
	logger *log.Logger // Логгер для записи логов
}

// NewService создает новый экземпляр Service.
func NewService(logger *log.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

// ValidateOrder валидирует заказ, используя кастомные правила валидации.
func (s *Service) ValidateOrder(orderData *model.Order) []*ValidationError {
	var validationErrors []*ValidationError

	// Проверка обязательного поля OrderUID
	if orderData.OrderUID == "" {
		validationErrors = append(validationErrors, &ValidationError{
			Field:   "OrderUID",
			Message: "Поле OrderUID обязательно для заполнения",
		})
		s.logger.Println("Ошибка валидации: поле OrderUID обязательно для заполнения")
	} else if !isValidUUID(orderData.OrderUID) {
		// Проверка формата UUID
		validationErrors = append(validationErrors, &ValidationError{
			Field:   "OrderUID",
			Message: "Поле OrderUID должно быть действительным UUID v4",
		})
		s.logger.Println("Ошибка валидации: поле OrderUID должно быть действительным UUID v4")
	}

	// Добавьте другие проверки по необходимости
	// Например, проверка обязательных полей для Payment и Item

	return validationErrors
}

// isValidUUID проверяет, является ли строка валидным UUID.
func isValidUUID(u string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$")
	return r.MatchString(u)
}

// ValidateAndDeserializeOrder валидирует и десериализует JSON в структуру Order.
func (s *Service) ValidateAndDeserializeOrder(data []byte) (*model.Order, []*ValidationError) {
	var order model.Order
	if err := json.Unmarshal(data, &order); err != nil {
		s.logger.Println("Ошибка десериализации заказа:", err)
		return nil, []*ValidationError{{Message: "Ошибка десериализации JSON"}}
	}

	validationErrors := s.ValidateOrder(&order)
	if len(validationErrors) > 0 {
		return nil, validationErrors
	}

	return &order, nil
}
