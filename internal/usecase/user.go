package usecase

import (
	"context"

	"gophermart/internal/domain"
	"gophermart/internal/logger"
	"gophermart/pkg/jwt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// userUseCase реализует бизнес-логику для работы с пользователями
type userUseCase struct {
	storage Storage
	jwt     jwt.TokenManager
}

// NewUserUseCase создает новый экземпляр userUseCase
func NewUserUseCase(storage Storage, jwt jwt.TokenManager) *userUseCase {
	return &userUseCase{
		storage: storage,
		jwt:     jwt,
	}
}

// Register регистрирует нового пользователя
func (uc *userUseCase) Register(ctx context.Context, creds *domain.Credentials) error {
	logger.Debug("Hashing password for new user", zap.String("login", creds.Login))

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return err
	}

	// Создаем пользователя
	logger.Debug("Creating new user in storage", zap.String("login", creds.Login))
	if err := uc.storage.CreateUser(ctx, creds.Login, string(hashedPassword)); err != nil {
		logger.Error("Failed to create user", zap.Error(err), zap.String("login", creds.Login))
		return err
	}

	logger.Info("User registered successfully", zap.String("login", creds.Login))
	return nil
}

// Login аутентифицирует пользователя
func (uc *userUseCase) Login(ctx context.Context, creds *domain.Credentials) (string, error) {
	// Получаем пользователя из БД
	logger.Debug("Getting user from storage", zap.String("login", creds.Login))
	user, err := uc.storage.GetUserByLogin(ctx, creds.Login)
	if err != nil {
		logger.Warn("User not found", zap.Error(err), zap.String("login", creds.Login))
		return "", domain.ErrInvalidCredentials
	}

	// Проверяем пароль
	logger.Debug("Comparing passwords", zap.String("login", creds.Login))
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		logger.Warn("Invalid password", zap.String("login", creds.Login))
		return "", domain.ErrInvalidCredentials
	}

	// Генерируем JWT токен
	logger.Debug("Generating JWT token", zap.Int64("user_id", user.ID))
	token, err := uc.jwt.GenerateToken(user.ID)
	if err != nil {
		logger.Error("Failed to generate token", zap.Error(err), zap.Int64("user_id", user.ID))
		return "", err
	}

	logger.Info("User logged in successfully", zap.String("login", creds.Login), zap.Int64("user_id", user.ID))
	return token, nil
}

// ValidateToken проверяет токен и возвращает ID пользователя
func (uc *userUseCase) ValidateToken(ctx context.Context, token string) (int64, error) {
	userID, err := uc.jwt.ValidateToken(token)
	if err != nil {
		logger.Error("Failed to validate token", zap.Error(err))
		return 0, err
	}
	return userID, nil
}
