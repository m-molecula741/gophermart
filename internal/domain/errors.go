package domain

import "errors"

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
)
