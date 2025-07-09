package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gophermart/internal/domain"
	"gophermart/internal/handler/mocks"
	"gophermart/internal/logger"
)

func init() {
	if err := logger.Initialize("error"); err != nil {
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
			name:          "Неверный формат JSON при регистрации",
			endpoint:      "/api/user/register",
			method:        http.MethodPost,
			requestBody:   "invalid json",
			mockBehavior:  func(m *mocks.MockUserUseCase) {},
			expectedCode:  http.StatusBadRequest,
			expectedToken: "",
		},
		{
			name:          "Неверный формат JSON при логине",
			endpoint:      "/api/user/login",
			method:        http.MethodPost,
			requestBody:   "invalid json",
			mockBehavior:  func(m *mocks.MockUserUseCase) {},
			expectedCode:  http.StatusBadRequest,
			expectedToken: "",
		},
		{
			name:     "Внутренняя ошибка сервера при регистрации",
			endpoint: "/api/user/register",
			method:   http.MethodPost,
			requestBody: domain.Credentials{
				Login:    "testuser",
				Password: "testpass",
			},
			mockBehavior: func(m *mocks.MockUserUseCase) {
				m.RegisterFunc = func(ctx context.Context, creds *domain.Credentials) error {
					return errors.New("internal error")
				}
			},
			expectedCode:  http.StatusInternalServerError,
			expectedToken: "",
		},
		{
			name:     "Внутренняя ошибка сервера при логине",
			endpoint: "/api/user/login",
			method:   http.MethodPost,
			requestBody: domain.Credentials{
				Login:    "testuser",
				Password: "testpass",
			},
			mockBehavior: func(m *mocks.MockUserUseCase) {
				m.LoginFunc = func(ctx context.Context, creds *domain.Credentials) (string, error) {
					return "", errors.New("internal error")
				}
			},
			expectedCode:  http.StatusInternalServerError,
			expectedToken: "",
		},
		{
			name:     "Пустые учетные данные при регистрации",
			endpoint: "/api/user/register",
			method:   http.MethodPost,
			requestBody: domain.Credentials{
				Login:    "",
				Password: "",
			},
			mockBehavior: func(m *mocks.MockUserUseCase) {
				m.RegisterFunc = func(ctx context.Context, creds *domain.Credentials) error {
					return domain.ErrInvalidCredentials
				}
			},
			expectedCode:  http.StatusBadRequest,
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &mocks.MockUserUseCase{}
			tt.mockBehavior(mockUseCase)
			mockUseCase.ValidateTokenFunc = func(ctx context.Context, token string) (int64, error) {
				return 1, nil
			}

			handler := NewAuthHandler(mockUseCase)

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

			switch tt.endpoint {
			case "/api/user/register":
				handler.Register(w, req)
			case "/api/user/login":
				handler.Login(w, req)
			}

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

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		setupAuth    func(*http.Request)
		mockBehavior func(*mocks.MockUserUseCase)
		expectedCode int
	}{
		{
			name: "Успешная авторизация",
			setupAuth: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer valid.token.123")
			},
			mockBehavior: func(m *mocks.MockUserUseCase) {
				m.ValidateTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 1, nil
				}
			},
			expectedCode: http.StatusOK,
		},
		{
			name:      "Отсутствует заголовок Authorization",
			setupAuth: func(r *http.Request) {},
			mockBehavior: func(m *mocks.MockUserUseCase) {
				m.ValidateTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, domain.ErrInvalidToken
				}
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Неверный формат токена",
			setupAuth: func(r *http.Request) {
				r.Header.Set("Authorization", "InvalidFormat")
			},
			mockBehavior: func(m *mocks.MockUserUseCase) {
				m.ValidateTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, domain.ErrInvalidToken
				}
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Недействительный токен",
			setupAuth: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer invalid.token")
			},
			mockBehavior: func(m *mocks.MockUserUseCase) {
				m.ValidateTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, domain.ErrInvalidToken
				}
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &mocks.MockUserUseCase{}
			tt.mockBehavior(mockUseCase)

			handler := NewAuthHandler(mockUseCase)

			// Создаем тестовый обработчик, который будет вызван после middleware
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем, что ID пользователя добавлен в контекст
				userID, ok := r.Context().Value(userIDKey).(int64)
				if !ok && tt.expectedCode == http.StatusOK {
					t.Error("User ID not found in context")
				}
				if ok && userID != 1 && tt.expectedCode == http.StatusOK {
					t.Errorf("Expected user ID 1, got %d", userID)
				}
				w.WriteHeader(http.StatusOK)
			})

			// Создаем middleware
			middleware := handler.AuthMiddleware(nextHandler)

			// Создаем тестовый запрос
			req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
			tt.setupAuth(req)
			w := httptest.NewRecorder()

			// Выполняем запрос через middleware
			middleware.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}
