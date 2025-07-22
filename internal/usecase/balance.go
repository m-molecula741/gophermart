package usecase

import (
	"context"
	"strconv"

	"gophermart/internal/domain"
	"gophermart/internal/logger"

	"go.uber.org/zap"
)

type balanceUseCase struct {
	storage Storage
}

// NewBalanceUseCase создает новый экземпляр BalanceUseCase
func NewBalanceUseCase(storage Storage) *balanceUseCase {
	return &balanceUseCase{
		storage: storage,
	}
}

// GetBalance возвращает текущий баланс пользователя
func (uc *balanceUseCase) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	balance, err := uc.storage.GetBalance(ctx, userID)
	if err != nil {
		logger.Error("Failed to get balance",
			zap.Error(err),
			zap.Int64("user_id", userID))
		return nil, err
	}

	logger.Info("Retrieved user balance",
		zap.Int64("user_id", userID),
		zap.Float64("current", balance.Current),
		zap.Float64("withdrawn", balance.Withdrawn))
	return balance, nil
}

// Withdraw списывает баллы с баланса пользователя
func (uc *balanceUseCase) Withdraw(ctx context.Context, userID int64, withdrawal domain.WithdrawalRequest) error {
	// Проверяем, что сумма положительная
	if withdrawal.Sum <= 0 {
		logger.Error("Invalid withdrawal amount",
			zap.Float64("sum", withdrawal.Sum))
		return domain.ErrInvalidAmount
	}

	// Проверяем, что номер заказа состоит только из цифр
	if _, err := strconv.ParseInt(withdrawal.Order, 10, 64); err != nil {
		logger.Error("Invalid order number format",
			zap.String("number", withdrawal.Order))
		return domain.ErrInvalidOrderNumber
	}

	// Проверяем номер по алгоритму Луна
	if !validateLuhn(withdrawal.Order) {
		logger.Error("Order number failed Luhn validation",
			zap.String("number", withdrawal.Order))
		return domain.ErrInvalidOrderNumber
	}

	// Получаем текущий баланс
	balance, err := uc.storage.GetBalance(ctx, userID)
	if err != nil {
		logger.Error("Failed to get balance for withdrawal",
			zap.Error(err),
			zap.Int64("user_id", userID))
		return err
	}

	// Проверяем достаточность средств
	if balance.Current < withdrawal.Sum {
		logger.Warn("Insufficient funds for withdrawal",
			zap.Int64("user_id", userID),
			zap.Float64("balance", balance.Current),
			zap.Float64("requested", withdrawal.Sum))
		return domain.ErrInsufficientFunds
	}

	// Создаем запись о списании
	if err := uc.storage.CreateWithdrawal(ctx, userID, withdrawal.Order, withdrawal.Sum); err != nil {
		logger.Error("Failed to create withdrawal",
			zap.Error(err),
			zap.Int64("user_id", userID),
			zap.String("order", withdrawal.Order),
			zap.Float64("sum", withdrawal.Sum))
		return err
	}

	logger.Info("Withdrawal created successfully",
		zap.Int64("user_id", userID),
		zap.String("order", withdrawal.Order),
		zap.Float64("sum", withdrawal.Sum))
	return nil
}

// GetWithdrawals возвращает историю списаний пользователя
func (uc *balanceUseCase) GetWithdrawals(ctx context.Context, userID int64) ([]domain.Withdrawal, error) {
	withdrawals, err := uc.storage.GetUserWithdrawals(ctx, userID)
	if err != nil {
		logger.Error("Failed to get withdrawals",
			zap.Error(err),
			zap.Int64("user_id", userID))
		return nil, err
	}

	if len(withdrawals) == 0 {
		logger.Info("No withdrawals found for user",
			zap.Int64("user_id", userID))
		return nil, nil
	}

	logger.Info("Retrieved user withdrawals",
		zap.Int64("user_id", userID),
		zap.Int("count", len(withdrawals)))
	return withdrawals, nil
}
