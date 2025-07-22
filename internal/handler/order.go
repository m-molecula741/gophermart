package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"gophermart/internal/domain"
	"gophermart/internal/logger"

	"go.uber.org/zap"
)

// OrderHandler обрабатывает запросы для работы с заказами
type OrderHandler struct {
	orderUseCase OrderUseCase
}

// NewOrderHandler создает новый экземпляр OrderHandler
func NewOrderHandler(orderUseCase OrderUseCase) *OrderHandler {
	return &OrderHandler{
		orderUseCase: orderUseCase,
	}
}

// UploadOrder обрабатывает загрузку нового заказа
func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		logger.Error("Failed to get user ID from context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Читаем номер заказа из тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	orderNumber := string(body)

	// Проверяем, что номер заказа не пустой
	if orderNumber == "" {
		logger.Error("Empty order number")
		http.Error(w, "order number cannot be empty", http.StatusBadRequest)
		return
	}

	// Загружаем заказ
	err = h.orderUseCase.UploadOrder(r.Context(), userID, orderNumber)
	if err != nil {
		logger.Error("Failed to upload order", zap.Error(err))
		switch err {
		case domain.ErrInvalidOrderNumber:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case domain.ErrOrderBelongsToUser:
			w.WriteHeader(http.StatusOK)
		case domain.ErrOrderBelongsToAnotherUser:
			w.WriteHeader(http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetOrders возвращает список заказов пользователя
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		logger.Error("Failed to get user ID from context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем список заказов
	orders, err := h.orderUseCase.GetUserOrders(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to get user orders", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Если заказов нет, возвращаем 204
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Сериализуем заказы в JSON
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		logger.Error("Failed to encode orders", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
