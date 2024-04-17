package validator

import (
	"encoding/json"
	"fmt"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// ValidationError представляет ошибку валидации для поля.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Service предоставляет методы для валидации моделей.
type Service struct {
	validate *validator.Validate
	logger   *logrus.Logger
}

// NewService создает новый экземпляр Service.
func _(logger *logrus.Logger) *Service {
	v := validator.New()
	// Здесь можно зарегистрировать кастомные валидационные функции или теги
	return &Service{
		validate: v,
		logger:   logger,
	}
}

// ValidateOrder валидирует заказ, используя теги в структуре Order.
func (s *Service) ValidateOrder(orderData *model.Order) []*ValidationError {
	var validationErrors []*ValidationError

	err := s.validate.Struct(orderData)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validationError := &ValidationError{
				Field:   err.Field(),
				Message: s.customErrorMessage(err),
			}
			validationErrors = append(validationErrors, validationError)
		}
	}

	return validationErrors
}

// customErrorMessage возвращает кастомное сообщение об ошибке на основе типа валидации.
func (s *Service) customErrorMessage(err validator.FieldError) string {
	// Здесь можно добавить логику для разных типов ошибок
	// Например, использовать switch err.Tag() для разных кейсов
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("Поле %s обязательно для заполнения", err.Field())
	case "e164":
		return fmt.Sprintf("Поле %s должно быть в формате E.164", err.Field())
	case "email":
		return fmt.Sprintf("Поле %s должно быть действительным адресом электронной почты", err.Field())
	case "uuid4":
		return fmt.Sprintf("Поле %s должно быть действительным UUID v4", err.Field())
	case "gt":
		return fmt.Sprintf("Поле %s должно быть больше указанного значения", err.Field())
	case "gte":
		return fmt.Sprintf("Поле %s должно быть больше или равно указанному значению", err.Field())
	default:
		return fmt.Sprintf("Неверное значение поля %s", err.Field())
	}
}

// ValidateAndDeserializeOrder валидирует и десериализует JSON в структуру Order.
func (s *Service) ValidateAndDeserializeOrder(data []byte) (*model.Order, []*ValidationError) {
	var order model.Order
	if err := json.Unmarshal(data, &order); err != nil {
		s.logger.WithError(err).Error("Ошибка десериализации заказа")
		return nil, []*ValidationError{{Message: "Ошибка десериализации JSON"}}
	}

	validationErrors := s.ValidateOrder(&order)
	if len(validationErrors) > 0 {
		return nil, validationErrors
	}

	return &order, nil
}
