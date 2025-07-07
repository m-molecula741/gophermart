package handler

import (
	"context"

	"gophermart/internal/domain"
)

// UserUseCase определяет методы для работы с пользователями
type UserUseCase interface {
	Register(ctx context.Context, creds *domain.Credentials) (string, error)
	Login(ctx context.Context, creds *domain.Credentials) (string, error)
}

// OrderUseCase определяет интерфейс для бизнес-логики работы с заказами
type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID int64, orderNumber string) error
	GetUserOrders(ctx context.Context, userID int64) ([]domain.Order, error)
	ProcessOrder(ctx context.Context, orderNumber string) error
}

// BalanceUseCase определяет интерфейс для бизнес-логики работы с балансом
type BalanceUseCase interface {
	GetBalance(ctx context.Context, userID int64) (*domain.Balance, error)
	Withdraw(ctx context.Context, userID int64, withdrawal domain.WithdrawalRequest) error
	GetWithdrawals(ctx context.Context, userID int64) ([]domain.Withdrawal, error)
}

// AuthMiddleware определяет интерфейс для middleware аутентификации
type AuthMiddleware interface {
	GetUserID(token string) (int64, error)
}
 