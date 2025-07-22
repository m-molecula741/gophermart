package usecase

import (
	"context"
	"testing"
	"time"

	"gophermart/internal/domain"
	"gophermart/internal/logger"
	"gophermart/internal/usecase/mocks"
	"gophermart/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

func init() {
	// Инициализируем логгер для тестов с уровнем error
	if err := logger.Initialize("error"); err != nil {
		panic(err)
	}
}

func TestUserUseCase(t *testing.T) {
	// Создаем JWT менеджер для тестов
	jwtManager := jwt.NewManager([]byte("test_secret"), 24*time.Hour)

	tests := []struct {
		name          string
		operation     string
		credentials   *domain.Credentials
		mockBehavior  func(*mocks.MockStorage)
		expectedError error
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
			},
			expectedError: nil,
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка мока хранилища
			mockStorage := &mocks.MockStorage{}
			tt.mockBehavior(mockStorage)

			// Создание usecase
			uc := NewUserUseCase(mockStorage, jwtManager)

			var err error

			// Выполнение операции
			ctx := context.Background()
			switch tt.operation {
			case "register":
				err = uc.Register(ctx, tt.credentials)
			case "login":
				var token string
				token, err = uc.Login(ctx, tt.credentials)
				if err == nil {
					if tt.expectedError != nil {
						t.Errorf("Expected error %v, got nil", tt.expectedError)
					} else if token == "" {
						t.Error("Expected non-empty token for successful login")
					}
				}
			}

			// Проверка ошибки
			if tt.expectedError != err {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestUserUseCase_ValidateToken(t *testing.T) {
	tests := []struct {
		name         string
		token        string
		mockBehavior func(*mocks.MockStorage, *mocks.MockJWTManager)
		wantUserID   int64
		wantErr      bool
	}{
		{
			name:  "Успешная валидация токена",
			token: "valid_token",
			mockBehavior: func(s *mocks.MockStorage, j *mocks.MockJWTManager) {
				j.ValidateTokenFunc = func(token string) (int64, error) {
					return 1, nil
				}
			},
			wantUserID: 1,
			wantErr:    false,
		},
		{
			name:  "Недействительный токен",
			token: "invalid_token",
			mockBehavior: func(s *mocks.MockStorage, j *mocks.MockJWTManager) {
				j.ValidateTokenFunc = func(token string) (int64, error) {
					return 0, domain.ErrInvalidToken
				}
			},
			wantUserID: 0,
			wantErr:    true,
		},
		{
			name:  "Пустой токен",
			token: "",
			mockBehavior: func(s *mocks.MockStorage, j *mocks.MockJWTManager) {
				j.ValidateTokenFunc = func(token string) (int64, error) {
					return 0, domain.ErrInvalidToken
				}
			},
			wantUserID: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &mocks.MockStorage{}
			mockJWT := &mocks.MockJWTManager{}
			tt.mockBehavior(mockStorage, mockJWT)

			uc := NewUserUseCase(mockStorage, mockJWT)

			userID, err := uc.ValidateToken(context.Background(), tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if userID != tt.wantUserID {
				t.Errorf("ValidateToken() userID = %v, want %v", userID, tt.wantUserID)
			}
		})
	}
}
