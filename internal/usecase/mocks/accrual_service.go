package mocks

import (
	"context"
	"gophermart/internal/domain"
)

// MockAccrualService мок для AccrualService
type MockAccrualService struct {
	GetOrderAccrualFunc func(ctx context.Context, orderNumber string) (*domain.Order, error)
}

// GetOrderAccrual реализация метода интерфейса
func (m *MockAccrualService) GetOrderAccrual(ctx context.Context, orderNumber string) (*domain.Order, error) {
	if m.GetOrderAccrualFunc != nil {
		return m.GetOrderAccrualFunc(ctx, orderNumber)
	}
	return nil, nil
}
