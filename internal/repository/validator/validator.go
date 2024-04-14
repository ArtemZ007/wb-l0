package validator

import (
	"regexp"
	"strings"
	"time"
)

// Validator предоставляет интерфейс для валидации данных.
type Validator interface {
	Validate() []string
}

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

// Validate для UserData.
func (u *UserData) Validate() []string {
	var errors []string

	if !isValidEmail(u.Email) {
		errors = append(errors, "Email is not valid")
	}
	if len(u.Username) < 3 {
		errors = append(errors, "Username must be at least 3 characters long")
	}
	if len(u.Password) < 6 {
		errors = append(errors, "Password must be at least 6 characters long")
	}

	return errors
}

// Validate для OrderData.
func (o *OrderData) Validate() []string {
	var errors []string

	if o.OrderUID == "" {
		errors = append(errors, "OrderUID is required")
	}
	if !isValidPhone(o.DeliveryPhone) {
		errors = append(errors, "Delivery phone is not valid")
	}
	if o.PaymentAmount <= 0 {
		errors = append(errors, "Payment amount must be greater than 0")
	}
	if !isValidDate(o.DateCreated) {
		errors = append(errors, "Date created is not valid")
	}
	for _, item := range o.Items {
		errors = append(errors, item.Validate()...)
	}

	return errors
}

// Validate для ItemData.
func (i *ItemData) Validate() []string {
	var errors []string

	if i.ChrtID <= 0 {
		errors = append(errors, "Item ChrtID must be greater than 0")
	}
	if i.Price <= 0 {
		errors = append(errors, "Item price must be greater than 0")
	}
	if i.Name == "" {
		errors = append(errors, "Item name is required")
	}

	return errors
}

// Вспомогательные функции валидации.
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(strings.TrimSpace(email))
}

func isValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+\d{1,3}\d{3,}$`)
	return phoneRegex.MatchString(strings.TrimSpace(phone))
}

func isValidDate(dateStr string) bool {
	_, err := time.Parse(time.RFC3339, dateStr)
	return err == nil
}
