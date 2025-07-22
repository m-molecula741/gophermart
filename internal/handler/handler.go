package handler

// Handler содержит все HTTP обработчики
type Handler struct {
	auth    *AuthHandler
	order   *OrderHandler
	balance *BalanceHandler
}

// NewHandler создает новый экземпляр Handler
func NewHandler(auth *AuthHandler, order *OrderHandler, balance *BalanceHandler) *Handler {
	return &Handler{
		auth:    auth,
		order:   order,
		balance: balance,
	}
}
