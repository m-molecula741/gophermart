package mocks

import (
	"context"
	"gophermart/internal/domain"
)

// MockOrderUseCase мок для OrderUseCase
type MockOrderUseCase struct {
	UploadOrderFunc   func(ctx context.Context, userID int64, orderNumber string) error
	GetUserOrdersFunc func(ctx context.Context, userID int64) ([]domain.Order, error)
	ShutdownFunc      func(ctx context.Context)
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

func (m *MockOrderUseCase) Shutdown(ctx context.Context) {
	if m.ShutdownFunc != nil {
		m.ShutdownFunc(ctx)
	}
}
