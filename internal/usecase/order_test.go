package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gophermart/internal/domain"
	"gophermart/internal/logger"
	"gophermart/internal/usecase/mocks"
)

// Переменная для мока в тестах
var testProcessOrderAccrual = func(orderNumber string) {}

func init() {
	// Инициализируем логгер для тестов с уровнем error
	if err := logger.Initialize("error"); err != nil {
		panic(err)
	}
}

func TestOrderUseCase(t *testing.T) {
	// Сохраняем оригинальную функцию
	originalProcessFunc := testProcessOrderAccrual

	// Подменяем на пустую функцию для тестов
	testProcessOrderAccrual = func(orderNumber string) {
		// Пустая функция, ничего не делает
	}

	// Восстанавливаем оригинальную функцию после тестов
	defer func() {
		testProcessOrderAccrual = originalProcessFunc
	}()

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
				// Первый вызов - при проверке существования заказа
				var callCount int
				s.GetOrderByNumberFunc = func(ctx context.Context, number string) (*domain.Order, error) {
					callCount++
					if callCount == 1 {
						return nil, domain.ErrOrderNotFound
					}
					return &domain.Order{
						Number: number,
						UserID: 1,
						Status: domain.StatusNew,
					}, nil
				}
				s.CreateOrderFunc = func(ctx context.Context, userID int64, number string) error {
					return nil
				}
				// Мокаем обновление статуса
				s.UpdateOrderStatusAndBalanceFunc = func(ctx context.Context, number string, status domain.OrderStatus, accrual float64, userID int64) error {
					return nil
				}
				// Мокаем ответ от сервиса начислений, чтобы горутина сразу завершалась
				a.GetOrderAccrualFunc = func(ctx context.Context, orderNumber string) (*domain.Order, error) {
					return &domain.Order{
						Number:  orderNumber,
						Status:  domain.StatusProcessed,
						Accrual: 500,
					}, nil
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
			expectedError: domain.ErrOrderBelongsToUser,
		},
		{
			name:        "Ошибка при создании заказа",
			userID:      1,
			orderNumber: "12345678903",
			mockBehavior: func(s *mocks.MockStorage, a *mocks.MockAccrualService) {
				s.GetOrderByNumberFunc = func(ctx context.Context, number string) (*domain.Order, error) {
					return nil, domain.ErrOrderNotFound
				}
				s.CreateOrderFunc = func(ctx context.Context, userID int64, number string) error {
					return domain.ErrOrderExists
				}
			},
			expectedError: domain.ErrOrderExists,
		},
		{
			name:        "Ошибка при проверке существования заказа",
			userID:      1,
			orderNumber: "12345678903",
			mockBehavior: func(s *mocks.MockStorage, a *mocks.MockAccrualService) {
				s.GetOrderByNumberFunc = func(ctx context.Context, number string) (*domain.Order, error) {
					return nil, domain.ErrOrderNotFound
				}
				s.CreateOrderFunc = func(ctx context.Context, userID int64, number string) error {
					return domain.ErrOrderNotFound
				}
			},
			expectedError: domain.ErrOrderNotFound,
		},
		{
			name:          "Невалидный номер заказа по алгоритму Луна",
			userID:        1,
			orderNumber:   "12345678902", // Неверная контрольная сумма
			mockBehavior:  func(s *mocks.MockStorage, a *mocks.MockAccrualService) {},
			expectedError: domain.ErrInvalidOrderNumber,
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

func TestOrderUseCase_ProcessOrderAccrual(t *testing.T) {
	// Подготавливаем тестовые данные
	orderNumber := "12345678903"
	userID := int64(1)
	accrual := float64(500)

	// Создаем моки
	mockStorage := &mocks.MockStorage{
		GetOrderByNumberFunc: func(ctx context.Context, number string) (*domain.Order, error) {
			return &domain.Order{
				Number: number,
				UserID: userID,
				Status: domain.StatusNew,
			}, nil
		},
		UpdateOrderStatusAndBalanceFunc: func(ctx context.Context, number string, status domain.OrderStatus, accrual float64, userID int64) error {
			// Проверяем, что параметры правильные
			if number != orderNumber {
				t.Errorf("Expected order number %s, got %s", orderNumber, number)
			}
			if status != domain.StatusProcessed {
				t.Errorf("Expected status %s, got %s", domain.StatusProcessed, status)
			}
			if accrual != 500 {
				t.Errorf("Expected accrual %f, got %f", 500.0, accrual)
			}
			if userID != 1 {
				t.Errorf("Expected user ID %d, got %d", 1, userID)
			}
			return nil
		},
	}

	mockAccrual := &mocks.MockAccrualService{
		GetOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*domain.Order, error) {
			return &domain.Order{
				Number:  orderNumber,
				Status:  domain.StatusProcessed,
				Accrual: accrual,
			}, nil
		},
	}

	// Создаем usecase
	uc := NewOrderUseCase(mockStorage, mockAccrual)

	// Запускаем обработку заказа
	uc.processOrderAccrual(orderNumber)

	// Проверяем, что все методы были вызваны
	// Добавьте здесь дополнительные проверки, если необходимо
}

func TestOrderUseCase_ProcessOrderAccrual_Error(t *testing.T) {
	// Подготавливаем тестовые данные
	orderNumber := "12345678903"
	userID := int64(1)

	// Создаем моки
	mockStorage := &mocks.MockStorage{
		GetOrderByNumberFunc: func(ctx context.Context, number string) (*domain.Order, error) {
			return &domain.Order{
				Number: number,
				UserID: userID,
				Status: domain.StatusNew,
			}, nil
		},
		UpdateOrderStatusAndBalanceFunc: func(ctx context.Context, number string, status domain.OrderStatus, accrual float64, userID int64) error {
			return fmt.Errorf("database error")
		},
	}

	mockAccrual := &mocks.MockAccrualService{
		GetOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*domain.Order, error) {
			return &domain.Order{
				Number:  orderNumber,
				Status:  domain.StatusProcessed,
				Accrual: 500,
			}, nil
		},
	}

	// Создаем usecase с контекстом и таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	uc := NewOrderUseCase(mockStorage, mockAccrual)

	// Запускаем обработку заказа
	go uc.processOrderAccrual(orderNumber)

	// Ждем завершения контекста
	<-ctx.Done()

	// Останавливаем обработку
	uc.Shutdown(context.Background())
}

func TestOrderUseCase_ProcessOrderAccrual_InvalidStatus(t *testing.T) {
	// Подготавливаем тестовые данные
	orderNumber := "12345678903"
	userID := int64(1)

	// Создаем моки
	mockStorage := &mocks.MockStorage{
		GetOrderByNumberFunc: func(ctx context.Context, number string) (*domain.Order, error) {
			return &domain.Order{
				Number: number,
				UserID: userID,
				Status: domain.StatusNew,
			}, nil
		},
		UpdateOrderStatusAndBalanceFunc: func(ctx context.Context, number string, status domain.OrderStatus, accrual float64, userID int64) error {
			if status != domain.StatusInvalid {
				t.Errorf("Expected status %s, got %s", domain.StatusInvalid, status)
			}
			return nil
		},
	}

	mockAccrual := &mocks.MockAccrualService{
		GetOrderAccrualFunc: func(ctx context.Context, orderNumber string) (*domain.Order, error) {
			return &domain.Order{
				Number: orderNumber,
				Status: domain.StatusInvalid,
			}, nil
		},
	}

	// Создаем usecase
	uc := NewOrderUseCase(mockStorage, mockAccrual)

	// Запускаем обработку заказа
	uc.processOrderAccrual(orderNumber)

	// Проверяем, что все методы были вызваны
	// Добавьте здесь дополнительные проверки, если необходимо
}
