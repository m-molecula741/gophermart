package domain

import "time"

// User представляет модель пользователя в системе
type User struct {
	ID           int64     `json:"-"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthResponse представляет ответ при успешной аутентификации
type AuthResponse struct {
	Token string `json:"token"`
}

// AuthError представляет ошибку аутентификации
type AuthError struct {
	Error string `json:"error"`
}
 