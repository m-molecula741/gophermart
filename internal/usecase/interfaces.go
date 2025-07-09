package usecase

import (
	"context"

	"gophermart/internal/domain"
)

// Storage определяет интерфейс для работы с хранилищем данных
type Storage interface {
	// Пользователи (для регистрации и аутентификации)
	CreateUser(ctx context.Context, login, passwordHash string) error
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)

	// Заказы
	CreateOrder(ctx context.Context, userID int64, number string) error
	GetUserOrders(ctx context.Context, userID int64) ([]domain.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, number string, status domain.OrderStatus, accrual float64) error

	// Баланс и списания
	GetBalance(ctx context.Context, userID int64) (*domain.Balance, error)
	CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error
	GetUserWithdrawals(ctx context.Context, userID int64) ([]domain.Withdrawal, error)

	// Служебные методы
	Ping(ctx context.Context) error
	Close() error
}

// AccrualService определяет интерфейс для работы с внешней системой начисления баллов
type AccrualService interface {
	GetOrderAccrual(ctx context.Context, orderNumber string) (*domain.Order, error)
}

// TokenManager определяет интерфейс для работы с JWT токенами
type TokenManager interface {
	CreateToken(userID int64) (string, error)
	ParseToken(token string) (int64, error)
}

// PasswordManager определяет интерфейс для работы с паролями
type PasswordManager interface {
	HashPassword(password string) (string, error)
	ComparePasswords(hashedPassword, password string) error
}

// OrderValidator определяет интерфейс для валидации номеров заказов
type OrderValidator interface {
	ValidateOrderNumber(number string) error
}

// UserUseCase определяет методы для работы с пользователями
type UserUseCase interface {
	// Register регистрирует нового пользователя
	Register(ctx context.Context, creds *domain.Credentials) (string, error)
	// Login аутентифицирует пользователя
	Login(ctx context.Context, creds *domain.Credentials) (string, error)
	// ValidateToken проверяет токен и возвращает ID пользователя
	ValidateToken(ctx context.Context, token string) (int64, error)
}
