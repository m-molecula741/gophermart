package mocks

import (
	"context"
	"gophermart/internal/domain"
)

// MockUserUseCase мок для UserUseCase
type MockUserUseCase struct {
	RegisterFunc      func(ctx context.Context, creds *domain.Credentials) error
	LoginFunc         func(ctx context.Context, creds *domain.Credentials) (string, error)
	ValidateTokenFunc func(ctx context.Context, token string) (int64, error)
}

func (m *MockUserUseCase) Register(ctx context.Context, creds *domain.Credentials) error {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, creds)
	}
	return nil
}

func (m *MockUserUseCase) Login(ctx context.Context, creds *domain.Credentials) (string, error) {
	return m.LoginFunc(ctx, creds)
}

func (m *MockUserUseCase) ValidateToken(ctx context.Context, token string) (int64, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(ctx, token)
	}
	return 0, nil
}
