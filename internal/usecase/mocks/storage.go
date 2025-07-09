package mocks

import (
	"context"
	"gophermart/internal/domain"
)

// MockStorage мок для Storage
type MockStorage struct {
	// Пользователи
	CreateUserFunc     func(ctx context.Context, login, passwordHash string) error
	GetUserByLoginFunc func(ctx context.Context, login string) (*domain.User, error)

	// Заказы
	CreateOrderFunc                 func(ctx context.Context, userID int64, number string) error
	GetUserOrdersFunc               func(ctx context.Context, userID int64) ([]domain.Order, error)
	GetOrderByNumberFunc            func(ctx context.Context, number string) (*domain.Order, error)
	UpdateOrderStatusAndBalanceFunc func(ctx context.Context, number string, status domain.OrderStatus, accrual float64, userID int64) error

	// Баланс и списания
	GetBalanceFunc         func(ctx context.Context, userID int64) (*domain.Balance, error)
	CreateWithdrawalFunc   func(ctx context.Context, userID int64, orderNumber string, sum float64) error
	GetUserWithdrawalsFunc func(ctx context.Context, userID int64) ([]domain.Withdrawal, error)

	// Служебные методы
	PingFunc  func(ctx context.Context) error
	CloseFunc func() error
}

// Пользователи
func (m *MockStorage) CreateUser(ctx context.Context, login, passwordHash string) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, login, passwordHash)
	}
	return nil
}

func (m *MockStorage) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	if m.GetUserByLoginFunc != nil {
		return m.GetUserByLoginFunc(ctx, login)
	}
	return nil, nil
}

// Заказы
func (m *MockStorage) CreateOrder(ctx context.Context, userID int64, number string) error {
	if m.CreateOrderFunc != nil {
		return m.CreateOrderFunc(ctx, userID, number)
	}
	return nil
}

func (m *MockStorage) GetUserOrders(ctx context.Context, userID int64) ([]domain.Order, error) {
	if m.GetUserOrdersFunc != nil {
		return m.GetUserOrdersFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockStorage) GetOrderByNumber(ctx context.Context, number string) (*domain.Order, error) {
	if m.GetOrderByNumberFunc != nil {
		return m.GetOrderByNumberFunc(ctx, number)
	}
	return nil, nil
}

func (m *MockStorage) UpdateOrderStatusAndBalance(ctx context.Context, number string, status domain.OrderStatus, accrual float64, userID int64) error {
	if m.UpdateOrderStatusAndBalanceFunc != nil {
		return m.UpdateOrderStatusAndBalanceFunc(ctx, number, status, accrual, userID)
	}
	return nil
}

// Баланс и списания
func (m *MockStorage) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockStorage) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	if m.CreateWithdrawalFunc != nil {
		return m.CreateWithdrawalFunc(ctx, userID, orderNumber, sum)
	}
	return nil
}

func (m *MockStorage) GetUserWithdrawals(ctx context.Context, userID int64) ([]domain.Withdrawal, error) {
	if m.GetUserWithdrawalsFunc != nil {
		return m.GetUserWithdrawalsFunc(ctx, userID)
	}
	return nil, nil
}

// Служебные методы
func (m *MockStorage) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil
}

func (m *MockStorage) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
