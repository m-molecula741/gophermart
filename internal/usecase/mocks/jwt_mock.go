package mocks

import (
	"gophermart/pkg/jwt"
)

// MockJWTManager мок для JWT менеджера
type MockJWTManager struct {
	GenerateTokenFunc func(userID int64) (string, error)
	ValidateTokenFunc func(token string) (int64, error)
}

var _ jwt.TokenManager = (*MockJWTManager)(nil)

// GenerateToken генерирует токен
func (m *MockJWTManager) GenerateToken(userID int64) (string, error) {
	return m.GenerateTokenFunc(userID)
}

// ValidateToken проверяет токен
func (m *MockJWTManager) ValidateToken(token string) (int64, error) {
	return m.ValidateTokenFunc(token)
}
