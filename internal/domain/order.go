package domain

import "time"

// OrderStatus представляет статус обработки заказа
type OrderStatus string

const (
	// StatusNew - заказ загружен в систему, но не попал в обработку
	StatusNew OrderStatus = "NEW"
	// StatusProcessing - вознаграждение за заказ рассчитывается
	StatusProcessing OrderStatus = "PROCESSING"
	// StatusInvalid - система расчёта вознаграждений отказала в расчёте
	StatusInvalid OrderStatus = "INVALID"
	// StatusProcessed - данные по заказу проверены и информация о расчёте успешно получена
	StatusProcessed OrderStatus = "PROCESSED"
)

// Order представляет информацию о заказе
type Order struct {
	UserID      int64       `json:"-"`
	Number      string      `json:"number"`
	Status      OrderStatus `json:"status"`
	Accrual     float64     `json:"accrual,omitempty"`
	UploadedAt  time.Time   `json:"uploaded_at"`
	ProcessedAt *time.Time  `json:"-"`
}
