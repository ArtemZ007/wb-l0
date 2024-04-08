package validator

import (
	"regexp"
	"strings"
	"time"
)

// UserData представляет данные пользователя, которые необходимо валидировать.
type UserData struct {
	Email    string
	Username string
	Password string
}

// OrderData представляет данные заказа, которые необходимо валидировать.
type OrderData struct {
	OrderUID      string
	TrackNumber   string
	DeliveryName  string
	DeliveryPhone string
	DeliveryZip   string
	DeliveryCity  string
	DeliveryEmail string
	PaymentAmount int
	PaymentDt     int64
	Items         []ItemData
	DateCreated   string
}

// ItemData представляет данные товара в заказе.
type ItemData struct {
	ChrtID int
	Price  int
	Name   string
	Brand  string
}

// Validate проверяет, что данные пользователя корректны.
// Возвращает слайс строк с описанием ошибок валидации или nil, если данные валидны.
func (u *UserData) Validate() []string {
	var validationErrors []string

	if !isValidEmail(u.Email) {
		validationErrors = append(validationErrors, "Email is not valid")
	}

	if len(u.Username) < 3 {
		validationErrors = append(validationErrors, "Username must be at least 3 characters long")
	}

	if len(u.Password) < 6 {
		validationErrors = append(validationErrors, "Password must be at least 6 characters long")
	}

	return validationErrors
}

// Validate проверяет, что данные заказа корректны.
// Возвращает слайс строк с описанием ошибок валидации или nil, если данные валидны.
func (o *OrderData) Validate() []string {
	var validationErrors []string

	if o.OrderUID == "" {
		validationErrors = append(validationErrors, "OrderUID is required")
	}

	if !isValidPhone(o.DeliveryPhone) {
		validationErrors = append(validationErrors, "Delivery phone is not valid")
	}

	if o.PaymentAmount <= 0 {
		validationErrors = append(validationErrors, "Payment amount must be greater than 0")
	}

	if !isValidDate(o.DateCreated) {
		validationErrors = append(validationErrors, "Date created is not valid")
	}

	for _, item := range o.Items {
		if item.ChrtID <= 0 {
			validationErrors = append(validationErrors, "Item ChrtID must be greater than 0")
		}
		if item.Price <= 0 {
			validationErrors = append(validationErrors, "Item price must be greater than 0")
		}
		if item.Name == "" {
			validationErrors = append(validationErrors, "Item name is required")
		}
	}

	return validationErrors
}

// isValidEmail проверяет, является ли строка валидным email адресом.
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
}

// isValidPhone проверяет, является ли строка валидным номером телефона.
func isValidPhone(phone string) bool {
	phone = strings.TrimSpace(phone)
	phoneRegex := regexp.MustCompile(`^\+\d{1,3}\d{3,}$`)
	return phoneRegex.MatchString(phone)
}

// isValidDate проверяет, является ли строка валидным форматом даты.
func isValidDate(dateStr string) bool {
	_, err := time.Parse(time.RFC3339, dateStr)
	return err == nil
}
