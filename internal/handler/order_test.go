package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gophermart/internal/domain"
	"gophermart/internal/handler/mocks"
	"gophermart/internal/logger"
)

func init() {
	if err := logger.Initialize("error"); err != nil {
		panic(err)
	}
}

func TestOrderHandler(t *testing.T) {
	// Тесты для загрузки заказа
	uploadTests := []struct {
		name         string
		orderNumber  string
		mockBehavior func(*mocks.MockOrderUseCase)
		expectedCode int
		withAuth     bool
		userID       int64
	}{
		{
			name:        "Успешная загрузка заказа",
			orderNumber: "12345678903",
			mockBehavior: func(m *mocks.MockOrderUseCase) {
				m.UploadOrderFunc = func(ctx context.Context, userID int64, orderNumber string) error {
					return nil
				}
				m.ShutdownFunc = func(ctx context.Context) {}
			},
			expectedCode: http.StatusAccepted,
			withAuth:     true,
			userID:       1,
		},
		{
			name:        "Неверный формат номера заказа",
			orderNumber: "invalid",
			mockBehavior: func(m *mocks.MockOrderUseCase) {
				m.UploadOrderFunc = func(ctx context.Context, userID int64, orderNumber string) error {
					return domain.ErrInvalidOrderNumber
				}
				m.ShutdownFunc = func(ctx context.Context) {}
			},
			expectedCode: http.StatusUnprocessableEntity,
			withAuth:     true,
			userID:       1,
		},
		{
			name:        "Заказ уже существует у текущего пользователя",
			orderNumber: "12345678903",
			mockBehavior: func(m *mocks.MockOrderUseCase) {
				m.UploadOrderFunc = func(ctx context.Context, userID int64, orderNumber string) error {
					return domain.ErrOrderBelongsToUser
				}
				m.ShutdownFunc = func(ctx context.Context) {}
			},
			expectedCode: http.StatusOK,
			withAuth:     true,
			userID:       1,
		},
		{
			name:        "Заказ принадлежит другому пользователю",
			orderNumber: "12345678903",
			mockBehavior: func(m *mocks.MockOrderUseCase) {
				m.UploadOrderFunc = func(ctx context.Context, userID int64, orderNumber string) error {
					return domain.ErrOrderBelongsToAnotherUser
				}
				m.ShutdownFunc = func(ctx context.Context) {}
			},
			expectedCode: http.StatusConflict,
			withAuth:     true,
			userID:       1,
		},
		{
			name:        "Внутренняя ошибка сервера при загрузке заказа",
			orderNumber: "12345678903",
			mockBehavior: func(m *mocks.MockOrderUseCase) {
				m.UploadOrderFunc = func(ctx context.Context, userID int64, orderNumber string) error {
					return errors.New("internal error")
				}
				m.ShutdownFunc = func(ctx context.Context) {}
			},
			expectedCode: http.StatusInternalServerError,
			withAuth:     true,
			userID:       1,
		},
		{
			name:        "Пустой номер заказа",
			orderNumber: "",
			mockBehavior: func(m *mocks.MockOrderUseCase) {
				m.ShutdownFunc = func(ctx context.Context) {}
			},
			expectedCode: http.StatusBadRequest,
			withAuth:     true,
			userID:       1,
		},
		{
			name:        "Без авторизации",
			orderNumber: "12345678903",
			mockBehavior: func(m *mocks.MockOrderUseCase) {
				m.ShutdownFunc = func(ctx context.Context) {}
			},
			expectedCode: http.StatusUnauthorized,
			withAuth:     false,
		},
	}

	for _, tt := range uploadTests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &mocks.MockOrderUseCase{}
			tt.mockBehavior(mockUseCase)

			handler := NewOrderHandler(mockUseCase)

			req := httptest.NewRequest(http.MethodPost, "/api/user/orders",
				bytes.NewBufferString(tt.orderNumber))
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

			if tt.withAuth {
				ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			handler.UploadOrder(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}

	// Тесты для получения списка заказов
	getTests := []struct {
		name         string
		mockBehavior func(*mocks.MockOrderUseCase)
		expectedCode int
		expectedBody []domain.Order
		withAuth     bool
		userID       int64
	}{
		{
			name: "Успешное получение заказов",
			mockBehavior: func(m *mocks.MockOrderUseCase) {
				m.GetUserOrdersFunc = func(ctx context.Context, userID int64) ([]domain.Order, error) {
					return []domain.Order{
						{
							Number:     "12345678903",
							Status:     domain.StatusProcessed,
							Accrual:    500,
							UploadedAt: time.Now(),
						},
					}, nil
				}
				m.ShutdownFunc = func(ctx context.Context) {}
			},
			expectedCode: http.StatusOK,
			expectedBody: []domain.Order{
				{
					Number:     "12345678903",
					Status:     domain.StatusProcessed,
					Accrual:    500,
					UploadedAt: time.Now(),
				},
			},
			withAuth: true,
			userID:   1,
		},
		{
			name: "Нет заказов",
			mockBehavior: func(m *mocks.MockOrderUseCase) {
				m.GetUserOrdersFunc = func(ctx context.Context, userID int64) ([]domain.Order, error) {
					return nil, nil
				}
				m.ShutdownFunc = func(ctx context.Context) {}
			},
			expectedCode: http.StatusNoContent,
			expectedBody: nil,
			withAuth:     true,
			userID:       1,
		},
		{
			name: "Внутренняя ошибка при получении заказов",
			mockBehavior: func(m *mocks.MockOrderUseCase) {
				m.GetUserOrdersFunc = func(ctx context.Context, userID int64) ([]domain.Order, error) {
					return nil, errors.New("internal error")
				}
				m.ShutdownFunc = func(ctx context.Context) {}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: nil,
			withAuth:     true,
			userID:       1,
		},
		{
			name: "Без авторизации",
			mockBehavior: func(m *mocks.MockOrderUseCase) {
				m.ShutdownFunc = func(ctx context.Context) {}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: nil,
			withAuth:     false,
		},
	}

	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &mocks.MockOrderUseCase{}
			tt.mockBehavior(mockUseCase)

			handler := NewOrderHandler(mockUseCase)

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			w := httptest.NewRecorder()

			if tt.withAuth {
				ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			handler.GetOrders(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			if tt.expectedBody != nil {
				var response []domain.Order
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}

				if len(response) != len(tt.expectedBody) {
					t.Errorf("Expected %d orders, got %d", len(tt.expectedBody), len(response))
				}

				if len(response) > 0 {
					if response[0].Number != tt.expectedBody[0].Number {
						t.Errorf("Expected order number %s, got %s",
							tt.expectedBody[0].Number, response[0].Number)
					}
					if response[0].Status != tt.expectedBody[0].Status {
						t.Errorf("Expected order status %s, got %s",
							tt.expectedBody[0].Status, response[0].Status)
					}
					if response[0].Accrual != tt.expectedBody[0].Accrual {
						t.Errorf("Expected order accrual %f, got %f",
							tt.expectedBody[0].Accrual, response[0].Accrual)
					}
				}
			}
		})
	}
}
