package mocks

import (
	"context"
	"gophermart/internal/domain"
)

// MockOrderUseCase мок для OrderUseCase
type MockOrderUseCase struct {
	UploadOrderFunc    func(ctx context.Context, userID int64, orderNumber string) error
	GetUserOrdersFunc  func(ctx context.Context, userID int64) ([]domain.Order, error)
	GetBalanceFunc     func(ctx context.Context, userID int64) (*domain.Balance, error)
	WithdrawFunc       func(ctx context.Context, userID int64, orderNumber string, sum float64) error
	GetWithdrawalsFunc func(ctx context.Context, userID int64) ([]domain.Withdrawal, error)
	ShutdownFunc       func(ctx context.Context)
}

func (m *MockOrderUseCase) UploadOrder(ctx context.Context, userID int64, orderNumber string) error {
	if m.UploadOrderFunc != nil {
		return m.UploadOrderFunc(ctx, userID, orderNumber)
	}
	return nil
}

func (m *MockOrderUseCase) GetUserOrders(ctx context.Context, userID int64) ([]domain.Order, error) {
	if m.GetUserOrdersFunc != nil {
		return m.GetUserOrdersFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockOrderUseCase) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockOrderUseCase) Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	if m.WithdrawFunc != nil {
		return m.WithdrawFunc(ctx, userID, orderNumber, sum)
	}
	return nil
}

func (m *MockOrderUseCase) GetWithdrawals(ctx context.Context, userID int64) ([]domain.Withdrawal, error) {
	if m.GetWithdrawalsFunc != nil {
		return m.GetWithdrawalsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockOrderUseCase) Shutdown(ctx context.Context) {
	if m.ShutdownFunc != nil {
		m.ShutdownFunc(ctx)
	}
}
