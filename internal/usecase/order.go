package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"gophermart/internal/domain"
	"gophermart/internal/logger"

	"go.uber.org/zap"
)

type orderUseCase struct {
	storage Storage
	accrual AccrualService
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewOrderUseCase создает новый экземпляр OrderUseCase
func NewOrderUseCase(storage Storage, accrual AccrualService) *orderUseCase {
	ctx, cancel := context.WithCancel(context.Background())
	return &orderUseCase{
		storage: storage,
		accrual: accrual,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Shutdown gracefully останавливает все фоновые процессы
func (uc *orderUseCase) Shutdown(ctx context.Context) {
	uc.cancel()

	// Ждем завершения контекста или таймаута
	select {
	case <-uc.ctx.Done():
		logger.Info("Order processing gracefully stopped")
	case <-ctx.Done():
		logger.Warn("Order processing shutdown timeout")
	}
}

// validateLuhn проверяет номер заказа по алгоритму Луна
func validateLuhn(number string) bool {
	sum := 0
	isSecond := false

	// Проходим по цифрам справа налево
	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')

		if isSecond {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += d
		isSecond = !isSecond
	}

	return sum%10 == 0
}

// processOrderAccrual обрабатывает начисление баллов за заказ
func (uc *orderUseCase) processOrderAccrual(orderNumber string) {
	logger.Info("Starting accrual processing",
		zap.String("order", orderNumber))

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	retryAfter := time.Duration(0)

	for {
		select {
		case <-uc.ctx.Done():
			logger.Info("Context cancelled, stopping accrual processing",
				zap.String("order", orderNumber))
			return
		case <-ticker.C:
			// Если есть задержка для повторного запроса, проверяем её
			if retryAfter > 0 {
				logger.Info("Waiting for retry delay to expire",
					zap.Duration("retry_after", retryAfter),
					zap.String("order", orderNumber))
				if time.Now().Before(time.Now().Add(-retryAfter)) {
					continue
				}
				retryAfter = 0
			}

			// Получаем информацию о начислении
			logger.Info("Requesting accrual info",
				zap.String("order", orderNumber))
			order, err := uc.accrual.GetOrderAccrual(uc.ctx, orderNumber)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}

				// Проверяем, является ли ошибка TooManyRequests
				var tooManyRequestsErr *domain.TooManyRequestsError
				if errors.As(err, &tooManyRequestsErr) {
					retryAfter = tooManyRequestsErr.RetryAfter
					logger.Warn("Too many requests to accrual service",
						zap.Duration("retry_after", retryAfter),
						zap.String("order", orderNumber))
					continue
				}

				logger.Error("Failed to get order accrual",
					zap.Error(err),
					zap.String("order", orderNumber))
				continue
			}

			if order == nil {
				logger.Info("Order not found in accrual system, continuing to retry",
					zap.String("order", orderNumber))
				continue
			}

			logger.Info("Received accrual response",
				zap.String("order", orderNumber),
				zap.String("status", string(order.Status)),
				zap.Float64("accrual", order.Accrual))

			// Получаем существующий заказ для определения userID
			existingOrder, err := uc.storage.GetOrderByNumber(uc.ctx, orderNumber)
			if err != nil {
				logger.Error("Failed to get existing order",
					zap.Error(err),
					zap.String("order", orderNumber))
				continue
			}

			// Атомарно обновляем статус заказа и баланс
			if err := uc.storage.UpdateOrderStatusAndBalance(uc.ctx, orderNumber, order.Status, order.Accrual, existingOrder.UserID); err != nil {
				logger.Error("Failed to update order status and balance",
					zap.Error(err),
					zap.String("order", orderNumber))
				continue
			}

			logger.Info("Updated order status and balance in database",
				zap.String("order", orderNumber),
				zap.String("status", string(order.Status)),
				zap.Float64("accrual", order.Accrual))

			// Если статус окончательный, завершаем обработку
			if order.Status == domain.StatusProcessed || order.Status == domain.StatusInvalid {
				logger.Info("Order processing completed",
					zap.String("order", orderNumber),
					zap.String("status", string(order.Status)),
					zap.Float64("accrual", order.Accrual))
				return
			}
		}
	}
}

// UploadOrder загружает новый номер заказа
func (uc *orderUseCase) UploadOrder(ctx context.Context, userID int64, orderNumber string) error {
	// Проверяем, что номер заказа состоит только из цифр
	if _, err := strconv.ParseInt(orderNumber, 10, 64); err != nil {
		logger.Error("Invalid order number format", zap.String("number", orderNumber))
		return domain.ErrInvalidOrderNumber
	}

	// Проверяем номер по алгоритму Луна
	if !validateLuhn(orderNumber) {
		logger.Error("Order number failed Luhn validation", zap.String("number", orderNumber))
		return domain.ErrInvalidOrderNumber
	}

	// Проверяем существование заказа
	existingOrder, err := uc.storage.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			// Если заказ не найден, создаем новый
			if err := uc.storage.CreateOrder(ctx, userID, orderNumber); err != nil {
				logger.Error("Failed to create order",
					zap.Error(err),
					zap.String("number", orderNumber),
					zap.Int64("user_id", userID))
				return err
			}

			// Запускаем обработку начисления в фоновом режиме
			go uc.processOrderAccrual(orderNumber)

			logger.Info("Order uploaded successfully",
				zap.String("number", orderNumber),
				zap.Int64("user_id", userID))
			return nil
		}
		// Если это любая другая ошибка
		logger.Error("Failed to check order existence", zap.Error(err))
		return err
	}

	// Если заказ существует, проверяем принадлежность
	if existingOrder.UserID == userID {
		logger.Info("Order already uploaded by current user",
			zap.String("number", orderNumber),
			zap.Int64("user_id", userID))
		return domain.ErrOrderBelongsToUser
	}

	// Если заказ принадлежит другому пользователю
	logger.Warn("Order belongs to another user",
		zap.String("number", orderNumber),
		zap.Int64("user_id", userID),
		zap.Int64("owner_id", existingOrder.UserID))
	return domain.ErrOrderExists
}

// GetUserOrders возвращает список заказов пользователя
func (uc *orderUseCase) GetUserOrders(ctx context.Context, userID int64) ([]domain.Order, error) {
	orders, err := uc.storage.GetUserOrders(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user orders",
			zap.Error(err),
			zap.Int64("user_id", userID))
		return nil, err
	}

	if len(orders) == 0 {
		logger.Info("No orders found for user", zap.Int64("user_id", userID))
		return nil, nil
	}

	logger.Info("Retrieved user orders",
		zap.Int64("user_id", userID),
		zap.Int("count", len(orders)))
	return orders, nil
}
