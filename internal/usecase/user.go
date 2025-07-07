package usecase

import (
	"context"

	"gophermart/internal/domain"
	"gophermart/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

// userUseCase реализует бизнес-логику для работы с пользователями
type userUseCase struct {
	storage Storage
	jwt     *jwt.Manager
}

// NewUserUseCase создает новый экземпляр userUseCase
func NewUserUseCase(storage Storage, jwt *jwt.Manager) *userUseCase {
	return &userUseCase{
		storage: storage,
		jwt:     jwt,
	}
}

// Register регистрирует нового пользователя
func (uc *userUseCase) Register(ctx context.Context, creds *domain.Credentials) (string, error) {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Создаем пользователя
	if err := uc.storage.CreateUser(ctx, creds.Login, string(hashedPassword)); err != nil {
		return "", err
	}

	// Получаем созданного пользователя для ID
	user, err := uc.storage.GetUserByLogin(ctx, creds.Login)
	if err != nil {
		return "", err
	}

	// Генерируем JWT токен
	token, err := uc.jwt.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Login аутентифицирует пользователя
func (uc *userUseCase) Login(ctx context.Context, creds *domain.Credentials) (string, error) {
	// Получаем пользователя из БД
	user, err := uc.storage.GetUserByLogin(ctx, creds.Login)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	// Генерируем JWT токен
	token, err := uc.jwt.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}
