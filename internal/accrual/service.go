package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gophermart/internal/domain"
	"gophermart/internal/logger"

	"go.uber.org/zap"
)

type Service struct {
	baseURL    string
	httpClient *http.Client
}

func NewService(baseURL string) *Service {
	return &Service{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type accrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func (s *Service) GetOrderAccrual(ctx context.Context, orderNumber string) (*domain.Order, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp accrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		order := &domain.Order{
			Number:  accrualResp.Order,
			Status:  domain.OrderStatus(accrualResp.Status),
			Accrual: accrualResp.Accrual,
		}

		logger.Info("Got accrual response",
			zap.String("order", orderNumber),
			zap.String("status", string(order.Status)),
			zap.Float64("accrual", order.Accrual))

		return order, nil

	case http.StatusTooManyRequests:
		// Получаем время ожидания из заголовка
		retryAfterStr := resp.Header.Get("Retry-After")
		retryAfterSec, err := strconv.Atoi(retryAfterStr)
		if err != nil {
			logger.Error("Invalid Retry-After header",
				zap.String("retry_after", retryAfterStr),
				zap.Error(err))
			retryAfterSec = 60 // Используем значение по умолчанию
		}

		retryAfter := time.Duration(retryAfterSec) * time.Second
		logger.Warn("Too many requests to accrual service",
			zap.String("order", orderNumber),
			zap.Duration("retry_after", retryAfter))

		return nil, domain.NewTooManyRequestsError(retryAfter)

	case http.StatusNoContent:
		logger.Info("Order not registered in accrual system", zap.String("order", orderNumber))
		return nil, nil

	default:
		logger.Error("Unexpected response from accrual service",
			zap.String("order", orderNumber),
			zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
