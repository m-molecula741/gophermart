package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"gophermart/internal/domain"
	"gophermart/internal/logger"
	"gophermart/internal/usecase"

	"go.uber.org/zap"
)

type contextKey string

const userIDKey contextKey = "user_id"

// AuthHandler обрабатывает запросы аутентификации
type AuthHandler struct {
	userUseCase usecase.UserUseCase
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(userUseCase usecase.UserUseCase) *AuthHandler {
	return &AuthHandler{
		userUseCase: userUseCase,
	}
}

// AuthMiddleware проверяет JWT токен и добавляет ID пользователя в контекст
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Error("Missing Authorization header")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Проверяем формат токена
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Error("Invalid Authorization header format")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Проверяем токен
		userID, err := h.userUseCase.ValidateToken(r.Context(), parts[1])
		if err != nil {
			logger.Error("Invalid token", zap.Error(err))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Добавляем ID пользователя в контекст
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Register обрабатывает регистрацию пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds domain.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		logger.Error("Failed to decode registration request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Info("Processing registration request", zap.String("login", creds.Login))

	if err := h.userUseCase.Register(r.Context(), &creds); err != nil {
		switch err {
		case domain.ErrUserExists:
			logger.Warn("Registration failed: user already exists", zap.String("login", creds.Login))
			http.Error(w, "user already exists", http.StatusConflict)
		default:
			logger.Error("Registration failed", zap.Error(err), zap.String("login", creds.Login))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	logger.Info("User registered successfully", zap.String("login", creds.Login))
	w.WriteHeader(http.StatusOK)
}

// Login обрабатывает вход пользователя
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds domain.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		logger.Error("Failed to decode login request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Info("Processing login request", zap.String("login", creds.Login))

	token, err := h.userUseCase.Login(r.Context(), &creds)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			logger.Warn("Login failed: invalid credentials", zap.String("login", creds.Login))
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			logger.Error("Login failed: internal error", zap.Error(err), zap.String("login", creds.Login))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	logger.Info("User logged in successfully", zap.String("login", creds.Login))

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	// Устанавливаем токен в заголовок Authorization
	w.Header().Set("Authorization", "Bearer "+token)

	// Отправляем успешный ответ с данными пользователя
	response := map[string]string{
		"login": creds.Login,
		"token": token,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
