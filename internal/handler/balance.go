package handler

import (
	"encoding/json"
	"net/http"

	"gophermart/internal/domain"
	"gophermart/internal/logger"

	"go.uber.org/zap"
)

// BalanceHandler обрабатывает запросы для работы с балансом
type BalanceHandler struct {
	balanceUseCase BalanceUseCase
}

// NewBalanceHandler создает новый экземпляр BalanceHandler
func NewBalanceHandler(balanceUseCase BalanceUseCase) *BalanceHandler {
	return &BalanceHandler{
		balanceUseCase: balanceUseCase,
	}
}

// GetBalance возвращает текущий баланс пользователя
func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		logger.Error("Failed to get user ID from context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceUseCase.GetBalance(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to get balance", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		logger.Error("Failed to encode balance", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

// Withdraw обрабатывает запрос на списание баллов
func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		logger.Error("Failed to get user ID from context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var withdrawal domain.WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&withdrawal); err != nil {
		logger.Error("Failed to decode withdrawal request", zap.Error(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err := h.balanceUseCase.Withdraw(r.Context(), userID, withdrawal)
	if err != nil {
		switch err {
		case domain.ErrInvalidOrderNumber:
			http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		case domain.ErrInsufficientFunds:
			http.Error(w, "insufficient funds", http.StatusPaymentRequired)
		default:
			logger.Error("Failed to process withdrawal", zap.Error(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals возвращает историю списаний
func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		logger.Error("Failed to get user ID from context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.balanceUseCase.GetWithdrawals(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to get withdrawals", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		logger.Error("Failed to encode withdrawals", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
