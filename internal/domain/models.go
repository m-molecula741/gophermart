package domain

import "time"

// User представляет пользователя системы
type User struct {
	ID           int64     `json:"-" db:"id"`
	Login        string    `json:"login" db:"login"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"-" db:"created_at"`
}

// Order представляет заказ в системе
type Order struct {
	Number     string    `json:"number" db:"number"`
	UserID     int64     `json:"-" db:"user_id"`
	Status     string    `json:"status" db:"status"`
	Accrual    float64   `json:"accrual,omitempty" db:"accrual"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

// Withdrawal представляет списание баллов
type Withdrawal struct {
	OrderNumber string    `json:"order" db:"order_number"`
	Sum         float64   `json:"sum" db:"sum"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}

// Balance представляет баланс пользователя
type Balance struct {
	Current   float64 `json:"current" db:"current"`
	Withdrawn float64 `json:"withdrawn" db:"withdrawn"`
}

// OrderStatus представляет возможные статусы заказа
type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

// WithdrawalRequest представляет запрос на списание баллов
type WithdrawalRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// Credentials представляет данные для аутентификации
type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
