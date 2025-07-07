package usecase

import (
	"context"
	"testing"
	"time"

	"gophermart/internal/domain"
	"gophermart/internal/usecase/mocks"
	"gophermart/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

func TestUserUseCase(t *testing.T) {
	// Создаем JWT менеджер для тестов
	jwtManager := jwt.NewManager([]byte("test_secret"), 24*time.Hour)

	tests := []struct {
		name          string
		operation     string
		credentials   *domain.Credentials
		mockBehavior  func(*mocks.MockStorage)
		expectedError error
		wantToken     bool
	}{
		{
			name:      "Успешная регистрация",
			operation: "register",
			credentials: &domain.Credentials{
				Login:    "newuser",
				Password: "password123",
			},
			mockBehavior: func(s *mocks.MockStorage) {
				s.CreateUserFunc = func(ctx context.Context, login, passwordHash string) error {
					return nil
				}
				s.GetUserByLoginFunc = func(ctx context.Context, login string) (*domain.User, error) {
					return &domain.User{
						ID:        1,
						Login:     "newuser",
						CreatedAt: time.Now(),
					}, nil
				}
			},
			expectedError: nil,
			wantToken:     true,
		},
		{
			name:      "Регистрация существующего пользователя",
			operation: "register",
			credentials: &domain.Credentials{
				Login:    "existinguser",
				Password: "password123",
			},
			mockBehavior: func(s *mocks.MockStorage) {
				s.CreateUserFunc = func(ctx context.Context, login, passwordHash string) error {
					return domain.ErrUserExists
				}
			},
			expectedError: domain.ErrUserExists,
			wantToken:     false,
		},
		{
			name:      "Успешная авторизация",
			operation: "login",
			credentials: &domain.Credentials{
				Login:    "existinguser",
				Password: "password123",
			},
			mockBehavior: func(s *mocks.MockStorage) {
				s.GetUserByLoginFunc = func(ctx context.Context, login string) (*domain.User, error) {
					hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
					return &domain.User{
						ID:           1,
						Login:        "existinguser",
						PasswordHash: string(hashedPassword),
						CreatedAt:    time.Now(),
					}, nil
				}
			},
			expectedError: nil,
			wantToken:     true,
		},
		{
			name:      "Авторизация с неверным паролем",
			operation: "login",
			credentials: &domain.Credentials{
				Login:    "existinguser",
				Password: "wrongpassword",
			},
			mockBehavior: func(s *mocks.MockStorage) {
				s.GetUserByLoginFunc = func(ctx context.Context, login string) (*domain.User, error) {
					hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
					return &domain.User{
						ID:           1,
						Login:        "existinguser",
						PasswordHash: string(hashedPassword),
						CreatedAt:    time.Now(),
					}, nil
				}
			},
			expectedError: domain.ErrInvalidCredentials,
			wantToken:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка мока хранилища
			mockStorage := &mocks.MockStorage{}
			tt.mockBehavior(mockStorage)

			// Создание usecase
			uc := NewUserUseCase(mockStorage, jwtManager)

			var token string
			var err error

			// Выполнение операции
			ctx := context.Background()
			switch tt.operation {
			case "register":
				token, err = uc.Register(ctx, tt.credentials)
			case "login":
				token, err = uc.Login(ctx, tt.credentials)
			}

			// Проверка ошибки
			if tt.expectedError != err {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			// Проверка токена
			if tt.wantToken {
				if token == "" {
					t.Error("Expected token, got empty string")
				}
				// Проверяем, что токен валидный
				userID, err := jwtManager.ParseToken(token)
				if err != nil {
					t.Errorf("Invalid token: %v", err)
				}
				if userID != 1 {
					t.Errorf("Expected user ID 1, got %d", userID)
				}
			} else if token != "" {
				t.Error("Expected empty token, got non-empty string")
			}
		})
	}
}
