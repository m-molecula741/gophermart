package domain

// Credentials представляет данные для аутентификации
type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
