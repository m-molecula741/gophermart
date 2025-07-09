package usecase

import (
	"context"
	"testing"
	"time"

	"gophermart/internal/domain"
	"gophermart/internal/usecase/mocks"
)

func TestBalanceUseCase(t *testing.T) {
	tests := []struct {
		name         string
		userID       int64
		mockBehavior func(*mocks.MockStorage)
		wantBalance  *domain.Balance
		wantErr      bool
	}{
		{
			name:   "Успешное получение баланса",
			userID: 1,
			mockBehavior: func(s *mocks.MockStorage) {
				s.GetBalanceFunc = func(ctx context.Context, userID int64) (*domain.Balance, error) {
					return &domain.Balance{
						Current:   1000,
						Withdrawn: 500,
					}, nil
				}
			},
			wantBalance: &domain.Balance{
				Current:   1000,
				Withdrawn: 500,
			},
			wantErr: false,
		},
		{
			name:   "Пользователь не найден",
			userID: 999,
			mockBehavior: func(s *mocks.MockStorage) {
				s.GetBalanceFunc = func(ctx context.Context, userID int64) (*domain.Balance, error) {
					return nil, domain.ErrUserNotFound
				}
			},
			wantBalance: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &mocks.MockStorage{}
			tt.mockBehavior(mockStorage)

			uc := NewBalanceUseCase(mockStorage)

			balance, err := uc.GetBalance(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && balance.Current != tt.wantBalance.Current {
				t.Errorf("GetBalance() current = %v, want %v", balance.Current, tt.wantBalance.Current)
			}
			if !tt.wantErr && balance.Withdrawn != tt.wantBalance.Withdrawn {
				t.Errorf("GetBalance() withdrawn = %v, want %v", balance.Withdrawn, tt.wantBalance.Withdrawn)
			}
		})
	}
}

func TestBalanceUseCase_Withdraw(t *testing.T) {
	tests := []struct {
		name         string
		userID       int64
		request      domain.WithdrawalRequest
		mockBehavior func(*mocks.MockStorage)
		wantErr      bool
	}{
		{
			name:   "Успешное списание",
			userID: 1,
			request: domain.WithdrawalRequest{
				Order: "12345678903",
				Sum:   100,
			},
			mockBehavior: func(s *mocks.MockStorage) {
				s.GetBalanceFunc = func(ctx context.Context, userID int64) (*domain.Balance, error) {
					return &domain.Balance{
						Current:   1000,
						Withdrawn: 500,
					}, nil
				}
				s.CreateWithdrawalFunc = func(ctx context.Context, userID int64, orderNumber string, sum float64) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:   "Недостаточно средств",
			userID: 1,
			request: domain.WithdrawalRequest{
				Order: "12345678903",
				Sum:   2000,
			},
			mockBehavior: func(s *mocks.MockStorage) {
				s.GetBalanceFunc = func(ctx context.Context, userID int64) (*domain.Balance, error) {
					return &domain.Balance{
						Current:   1000,
						Withdrawn: 500,
					}, nil
				}
			},
			wantErr: true,
		},
		{
			name:   "Неверный номер заказа",
			userID: 1,
			request: domain.WithdrawalRequest{
				Order: "invalid",
				Sum:   100,
			},
			mockBehavior: func(s *mocks.MockStorage) {
				s.GetBalanceFunc = func(ctx context.Context, userID int64) (*domain.Balance, error) {
					return &domain.Balance{
						Current:   1000,
						Withdrawn: 500,
					}, nil
				}
			},
			wantErr: true,
		},
		{
			name:   "Отрицательная сумма",
			userID: 1,
			request: domain.WithdrawalRequest{
				Order: "12345678903",
				Sum:   -100,
			},
			mockBehavior: func(s *mocks.MockStorage) {
				s.GetBalanceFunc = func(ctx context.Context, userID int64) (*domain.Balance, error) {
					return &domain.Balance{
						Current:   1000,
						Withdrawn: 500,
					}, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &mocks.MockStorage{}
			tt.mockBehavior(mockStorage)

			uc := NewBalanceUseCase(mockStorage)

			err := uc.Withdraw(context.Background(), tt.userID, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Withdraw() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBalanceUseCase_GetWithdrawals(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		userID        int64
		mockBehavior  func(*mocks.MockStorage)
		wantWithdraws []domain.Withdrawal
		wantErr       bool
	}{
		{
			name:   "Успешное получение списаний",
			userID: 1,
			mockBehavior: func(s *mocks.MockStorage) {
				s.GetUserWithdrawalsFunc = func(ctx context.Context, userID int64) ([]domain.Withdrawal, error) {
					return []domain.Withdrawal{
						{
							OrderNumber: "12345678903",
							Sum:         100,
							ProcessedAt: now,
						},
					}, nil
				}
			},
			wantWithdraws: []domain.Withdrawal{
				{
					OrderNumber: "12345678903",
					Sum:         100,
					ProcessedAt: now,
				},
			},
			wantErr: false,
		},
		{
			name:   "Нет списаний",
			userID: 1,
			mockBehavior: func(s *mocks.MockStorage) {
				s.GetUserWithdrawalsFunc = func(ctx context.Context, userID int64) ([]domain.Withdrawal, error) {
					return nil, nil
				}
			},
			wantWithdraws: nil,
			wantErr:       false,
		},
		{
			name:   "Ошибка при получении списаний",
			userID: 1,
			mockBehavior: func(s *mocks.MockStorage) {
				s.GetUserWithdrawalsFunc = func(ctx context.Context, userID int64) ([]domain.Withdrawal, error) {
					return nil, domain.ErrUserNotFound
				}
			},
			wantWithdraws: nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &mocks.MockStorage{}
			tt.mockBehavior(mockStorage)

			uc := NewBalanceUseCase(mockStorage)

			withdrawals, err := uc.GetWithdrawals(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWithdrawals() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(withdrawals) != len(tt.wantWithdraws) {
					t.Errorf("GetWithdrawals() got %v withdrawals, want %v", len(withdrawals), len(tt.wantWithdraws))
					return
				}
				if len(withdrawals) > 0 {
					if withdrawals[0].OrderNumber != tt.wantWithdraws[0].OrderNumber {
						t.Errorf("GetWithdrawals() order = %v, want %v", withdrawals[0].OrderNumber, tt.wantWithdraws[0].OrderNumber)
					}
					if withdrawals[0].Sum != tt.wantWithdraws[0].Sum {
						t.Errorf("GetWithdrawals() sum = %v, want %v", withdrawals[0].Sum, tt.wantWithdraws[0].Sum)
					}
				}
			}
		})
	}
}
