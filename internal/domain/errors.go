package domain

import (
	"errors"
	"time"
)

var (
	// ErrInvalidCredentials возвращается при неверных учетных данных
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserExists возвращается при попытке зарегистрировать существующего пользователя
	ErrUserExists = errors.New("user already exists")

	// ErrOrderExists возвращается при попытке добавить существующий заказ
	ErrOrderExists = errors.New("order already exists")

	// ErrInvalidOrderNumber возвращается при неверном номере заказа
	ErrInvalidOrderNumber = errors.New("invalid order number")

	// ErrInsufficientFunds возвращается при недостаточном балансе
	ErrInsufficientFunds = errors.New("insufficient funds")

	// ErrOrderNotFound возвращается, когда заказ не найден
	ErrOrderNotFound = errors.New("order not found")

	// ErrUserNotFound пользователь не найден
	ErrUserNotFound = errors.New("user not found")
)

// TooManyRequestsError ошибка превышения лимита запросов
type TooManyRequestsError struct {
	RetryAfter time.Duration
}

func (e *TooManyRequestsError) Error() string {
	return "too many requests"
}

// NewTooManyRequestsError создает новую ошибку превышения лимита запросов
func NewTooManyRequestsError(retryAfter time.Duration) *TooManyRequestsError {
	return &TooManyRequestsError{
		RetryAfter: retryAfter,
	}
}
