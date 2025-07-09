package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gophermart/internal/domain"
	"gophermart/internal/handler/mocks"
	"gophermart/internal/logger"
)

func init() {
	// Инициализируем логгер для тестов
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
}

func TestAuthHandler(t *testing.T) {
	tests := []struct {
		name          string
		endpoint      string
		method        string
		requestBody   interface{}
		mockBehavior  func(*mocks.MockUserUseCase)
		expectedCode  int
		expectedToken string
	}{
		{
			name:     "Успешная регистрация",
			endpoint: "/api/user/register",
			method:   http.MethodPost,
			requestBody: domain.Credentials{
				Login:    "testuser",
				Password: "testpass",
			},
			mockBehavior: func(m *mocks.MockUserUseCase) {
				m.RegisterFunc = func(ctx context.Context, creds *domain.Credentials) error {
					return nil
				}
			},
			expectedCode:  http.StatusOK,
			expectedToken: "",
		},
		{
			name:     "Регистрация существующего пользователя",
			endpoint: "/api/user/register",
			method:   http.MethodPost,
			requestBody: domain.Credentials{
				Login:    "existinguser",
				Password: "testpass",
			},
			mockBehavior: func(m *mocks.MockUserUseCase) {
				m.RegisterFunc = func(ctx context.Context, creds *domain.Credentials) error {
					return domain.ErrUserExists
				}
			},
			expectedCode:  http.StatusConflict,
			expectedToken: "",
		},
		{
			name:     "Успешная авторизация",
			endpoint: "/api/user/login",
			method:   http.MethodPost,
			requestBody: domain.Credentials{
				Login:    "testuser",
				Password: "testpass",
			},
			mockBehavior: func(m *mocks.MockUserUseCase) {
				m.LoginFunc = func(ctx context.Context, creds *domain.Credentials) (string, error) {
					return "test.token.123", nil
				}
			},
			expectedCode:  http.StatusOK,
			expectedToken: "Bearer test.token.123",
		},
		{
			name:     "Неверные учетные данные",
			endpoint: "/api/user/login",
			method:   http.MethodPost,
			requestBody: domain.Credentials{
				Login:    "testuser",
				Password: "wrongpass",
			},
			mockBehavior: func(m *mocks.MockUserUseCase) {
				m.LoginFunc = func(ctx context.Context, creds *domain.Credentials) (string, error) {
					return "", domain.ErrInvalidCredentials
				}
			},
			expectedCode:  http.StatusUnauthorized,
			expectedToken: "",
		},
		{
			name:          "Неверный формат JSON",
			endpoint:      "/api/user/register",
			method:        http.MethodPost,
			requestBody:   "invalid json",
			mockBehavior:  func(m *mocks.MockUserUseCase) {},
			expectedCode:  http.StatusBadRequest,
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка мока
			mockUseCase := &mocks.MockUserUseCase{}
			tt.mockBehavior(mockUseCase)
			mockUseCase.ValidateTokenFunc = func(ctx context.Context, token string) (int64, error) {
				return 1, nil
			}

			// Создание хендлера
			handler := NewAuthHandler(mockUseCase)

			// Создание запроса
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatal(err)
				}
			}

			req := httptest.NewRequest(tt.method, tt.endpoint, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Выполнение запроса
			switch tt.endpoint {
			case "/api/user/register":
				handler.Register(w, req)
			case "/api/user/login":
				handler.Login(w, req)
			}

			// Проверка результатов
			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			if tt.expectedToken != "" {
				token := w.Header().Get("Authorization")
				if token != tt.expectedToken {
					t.Errorf("Expected token %s, got %s", tt.expectedToken, token)
				}
			}
		})
	}
}
