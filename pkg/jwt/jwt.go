package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// Manager управляет JWT токенами
type Manager struct {
	signingKey []byte
	tokenTTL   time.Duration
}

// NewManager создает новый Manager
func NewManager(signingKey []byte, tokenTTL time.Duration) *Manager {
	return &Manager{
		signingKey: signingKey,
		tokenTTL:   tokenTTL,
	}
}

// Claims представляет данные JWT токена
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.StandardClaims
}

// GenerateToken создает новый JWT токен
func (m *Manager) GenerateToken(userID int64) (string, error) {
	claims := Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(m.tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.signingKey)
}

// ParseToken проверяет и парсит JWT токен
func (m *Manager) ParseToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.signingKey, nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}

	return claims.UserID, nil
}
