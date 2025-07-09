package usecase

import (
	"context"
	"testing"
	"time"

	"gophermart/internal/domain"
	"gophermart/internal/logger"
	"gophermart/internal/usecase/mocks"
)

func init() {
	// Инициализируем логгер для тестов
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
}

func TestOrderUseCase(t *testing.T) {
	// Тесты для загрузки заказа
	uploadTests := []struct {
		name          string
		userID        int64
		orderNumber   string
		mockBehavior  func(*mocks.MockStorage, *mocks.MockAccrualService)
		expectedError error
	}{
		{
			name:        "Успешная загрузка заказа",
			userID:      1,
			orderNumber: "12345678903", // Валидный номер по алгоритму Луна
			mockBehavior: func(s *mocks.MockStorage, a *mocks.MockAccrualService) {
				s.GetOrderByNumberFunc = func(ctx context.Context, number string) (*domain.Order, error) {
					return nil, domain.ErrOrderNotFound
				}
				s.CreateOrderFunc = func(ctx context.Context, userID int64, number string) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:          "Неверный формат номера заказа",
			userID:        1,
			orderNumber:   "invalid",
			mockBehavior:  func(s *mocks.MockStorage, a *mocks.MockAccrualService) {},
			expectedError: domain.ErrInvalidOrderNumber,
		},
		{
			name:        "Заказ уже существует у другого пользователя",
			userID:      1,
			orderNumber: "12345678903",
			mockBehavior: func(s *mocks.MockStorage, a *mocks.MockAccrualService) {
				s.GetOrderByNumberFunc = func(ctx context.Context, number string) (*domain.Order, error) {
					return &domain.Order{
						UserID: 2,
						Number: "12345678903",
					}, nil
				}
			},
			expectedError: domain.ErrOrderExists,
		},
		{
			name:        "Заказ уже существует у текущего пользователя",
			userID:      1,
			orderNumber: "12345678903",
			mockBehavior: func(s *mocks.MockStorage, a *mocks.MockAccrualService) {
				s.GetOrderByNumberFunc = func(ctx context.Context, number string) (*domain.Order, error) {
					return &domain.Order{
						UserID: 1,
						Number: "12345678903",
					}, nil
				}
			},
			expectedError: nil,
		},
	}

	for _, tt := range uploadTests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка моков
			mockStorage := &mocks.MockStorage{}
			mockAccrual := &mocks.MockAccrualService{}
			tt.mockBehavior(mockStorage, mockAccrual)

			// Создание usecase
			uc := NewOrderUseCase(mockStorage, mockAccrual)

			// Выполнение операции
			err := uc.UploadOrder(context.Background(), tt.userID, tt.orderNumber)

			// Проверка результатов
			if err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}
		})
	}

	// Тесты для получения списка заказов
	getTests := []struct {
		name           string
		userID         int64
		mockBehavior   func(*mocks.MockStorage, *mocks.MockAccrualService)
		expectedOrders []domain.Order
		expectedError  error
	}{
		{
			name:   "Успешное получение заказов",
			userID: 1,
			mockBehavior: func(s *mocks.MockStorage, a *mocks.MockAccrualService) {
				s.GetUserOrdersFunc = func(ctx context.Context, userID int64) ([]domain.Order, error) {
					return []domain.Order{
						{
							UserID:     1,
							Number:     "12345678903",
							Status:     domain.StatusProcessed,
							Accrual:    500,
							UploadedAt: time.Now(),
						},
					}, nil
				}
			},
			expectedOrders: []domain.Order{
				{
					UserID:     1,
					Number:     "12345678903",
					Status:     domain.StatusProcessed,
					Accrual:    500,
					UploadedAt: time.Now(),
				},
			},
			expectedError: nil,
		},
		{
			name:   "Нет заказов",
			userID: 1,
			mockBehavior: func(s *mocks.MockStorage, a *mocks.MockAccrualService) {
				s.GetUserOrdersFunc = func(ctx context.Context, userID int64) ([]domain.Order, error) {
					return nil, nil
				}
			},
			expectedOrders: nil,
			expectedError:  nil,
		},
	}

	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка моков
			mockStorage := &mocks.MockStorage{}
			mockAccrual := &mocks.MockAccrualService{}
			tt.mockBehavior(mockStorage, mockAccrual)

			// Создание usecase
			uc := NewOrderUseCase(mockStorage, mockAccrual)

			// Выполнение операции
			orders, err := uc.GetUserOrders(context.Background(), tt.userID)

			// Проверка результатов
			if err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			if tt.expectedOrders != nil {
				if len(orders) != len(tt.expectedOrders) {
					t.Errorf("Expected %d orders, got %d", len(tt.expectedOrders), len(orders))
				}

				if len(orders) > 0 {
					if orders[0].Number != tt.expectedOrders[0].Number {
						t.Errorf("Expected order number %s, got %s",
							tt.expectedOrders[0].Number, orders[0].Number)
					}
					if orders[0].Status != tt.expectedOrders[0].Status {
						t.Errorf("Expected order status %s, got %s",
							tt.expectedOrders[0].Status, orders[0].Status)
					}
					if orders[0].Accrual != tt.expectedOrders[0].Accrual {
						t.Errorf("Expected order accrual %f, got %f",
							tt.expectedOrders[0].Accrual, orders[0].Accrual)
					}
				}
			}
		})
	}
}
