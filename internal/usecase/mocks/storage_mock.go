package mocks

import (
	"context"
	"gophermart/internal/domain"
)

// MockStorage мок для Storage
type MockStorage struct {
	CreateUserFunc     func(ctx context.Context, login, passwordHash string) error
	GetUserByLoginFunc func(ctx context.Context, login string) (*domain.User, error)
}

func (m *MockStorage) CreateUser(ctx context.Context, login, passwordHash string) error {
	return m.CreateUserFunc(ctx, login, passwordHash)
}

func (m *MockStorage) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	return m.GetUserByLoginFunc(ctx, login)
}

// Реализация остальных методов интерфейса Storage с пустыми заглушками
func (m *MockStorage) Ping(ctx context.Context) error                                     { return nil }
func (m *MockStorage) Close() error                                                       { return nil }
func (m *MockStorage) CreateOrder(ctx context.Context, userID int64, number string) error { return nil }
func (m *MockStorage) GetOrderByNumber(ctx context.Context, number string) (*domain.Order, error) {
	return nil, nil
}
func (m *MockStorage) GetUserOrders(ctx context.Context, userID int64) ([]domain.Order, error) {
	return nil, nil
}
func (m *MockStorage) UpdateOrderStatus(ctx context.Context, number string, status domain.OrderStatus, accrual float64) error {
	return nil
}
func (m *MockStorage) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	return nil, nil
}
func (m *MockStorage) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	return nil
}
func (m *MockStorage) GetUserWithdrawals(ctx context.Context, userID int64) ([]domain.Withdrawal, error) {
	return nil, nil
}
