package domain

import "time"

// OrderStatus представляет статус заказа
type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "NEW"
	OrderStatusInvalid   OrderStatus = "INVALID"
	OrderStatusAccepted  OrderStatus = "ACCEPTED"
	OrderStatusProcessed OrderStatus = "PROCESSED"
)

// Order представляет заказ в системе
type Order struct {
	Number     string      `json:"number"`
	UserID     int64       `json:"-"`
	Status     OrderStatus `json:"status"`
	Accrual    float64     `json:"accrual,omitempty"`
	UploadedAt time.Time   `json:"uploaded_at"`
}

// Balance представляет баланс пользователя
type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// Withdrawal представляет операцию списания баллов
type Withdrawal struct {
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// WithdrawalRequest представляет запрос на списание баллов
type WithdrawalRequest struct {
	OrderNumber string  `json:"order"`
	Sum         float64 `json:"sum"`
}

// Credentials представляет данные для аутентификации
type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
