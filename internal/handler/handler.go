package handler

// Handler содержит все HTTP обработчики
type Handler struct {
	auth  *AuthHandler
	order *OrderHandler
}

// NewHandler создает новый экземпляр Handler
func NewHandler(auth *AuthHandler, order *OrderHandler) *Handler {
	return &Handler{
		auth:  auth,
		order: order,
	}
}
