package mocks

import (
	"context"
	"gophermart/internal/domain"
)

// MockUserUseCase мок для UserUseCase
type MockUserUseCase struct {
	RegisterFunc func(ctx context.Context, creds *domain.Credentials) (string, error)
	LoginFunc    func(ctx context.Context, creds *domain.Credentials) (string, error)
}

func (m *MockUserUseCase) Register(ctx context.Context, creds *domain.Credentials) (string, error) {
	return m.RegisterFunc(ctx, creds)
}

func (m *MockUserUseCase) Login(ctx context.Context, creds *domain.Credentials) (string, error) {
	return m.LoginFunc(ctx, creds)
}
