package validator

import (
	"regexp"
	"strings"
)

// UserData представляет данные пользователя, которые необходимо валидировать.
type UserData struct {
	Email    string
	Username string
	Password string
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

	if len(validationErrors) == 0 {
		return nil
	}

	return validationErrors
}

// isValidEmail проверяет, является ли строка валидным email адресом.
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 3 && len(email) > 254 {
		return false
	}
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
}
