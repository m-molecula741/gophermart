package handler

// Handler содержит все HTTP обработчики
type Handler struct {
	auth *AuthHandler
}

// NewHandler создает новый экземпляр Handler
func NewHandler(auth *AuthHandler) *Handler {
	return &Handler{
		auth: auth,
	}
}
