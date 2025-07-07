package handler

import (
	"encoding/json"
	"net/http"

	"gophermart/internal/domain"
	"gophermart/internal/usecase"
)

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

// Register обрабатывает регистрацию пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds domain.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.userUseCase.Register(r.Context(), &creds)
	if err != nil {
		switch err {
		case domain.ErrUserExists:
			http.Error(w, "user already exists", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
}

// Login обрабатывает вход пользователя
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds domain.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.userUseCase.Login(r.Context(), &creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
}
