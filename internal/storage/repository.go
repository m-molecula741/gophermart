package storage

import (
	"context"
	"fmt"

	"gophermart/internal/domain"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
)

// PostgresRepository реализует интерфейс Storage для PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(ctx context.Context, dsn string) (*PostgresRepository, error) {
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &PostgresRepository{
		pool: pool,
	}, nil
}

// Ping проверяет соединение с базой данных
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// Close закрывает соединение с базой данных
func (r *PostgresRepository) Close() error {
	if r.pool != nil {
		r.pool.Close()
	}
	return nil
}

// CreateUser создает нового пользователя
func (r *PostgresRepository) CreateUser(ctx context.Context, login, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (login, password_hash) 
		 VALUES ($1, $2)`,
		login, passwordHash,
	)
	if err != nil {
		// Проверяем, является ли ошибка нарушением уникального ограничения
		if pqErr, ok := err.(*pgconn.PgError); ok {
			if pqErr.Code == "23505" { // unique_violation
				return domain.ErrUserExists
			}
		}
		return fmt.Errorf("error creating user: %w", err)
	}
	return nil
}

// GetUserByLogin находит пользователя по логину
func (r *PostgresRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	var user domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, login, password_hash, created_at 
		 FROM users 
		 WHERE login = $1`,
		login,
	).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("error getting user by login: %w", err)
	}
	return &user, nil
}

// CreateOrder создает новый заказ
func (r *PostgresRepository) CreateOrder(ctx context.Context, userID int64, number string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO orders (number, user_id, status) 
		 VALUES ($1, $2, $3)`,
		number, userID, domain.OrderStatusNew,
	)
	if err != nil {
		return fmt.Errorf("error creating order: %w", err)
	}
	return nil
}

// GetOrderByNumber находит заказ по номеру
func (r *PostgresRepository) GetOrderByNumber(ctx context.Context, number string) (*domain.Order, error) {
	var order domain.Order
	err := r.pool.QueryRow(ctx,
		`SELECT number, user_id, status, accrual, uploaded_at 
		 FROM orders 
		 WHERE number = $1`,
		number,
	).Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)

	if err != nil {
		return nil, fmt.Errorf("error getting order by number: %w", err)
	}
	return &order, nil
}

// GetUserOrders возвращает все заказы пользователя
func (r *PostgresRepository) GetUserOrders(ctx context.Context, userID int64) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT number, user_id, status, accrual, uploaded_at 
		 FROM orders 
		 WHERE user_id = $1 
		 ORDER BY uploaded_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting user orders: %w", err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		err := rows.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning order: %w", err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// UpdateOrderStatus обновляет статус заказа
func (r *PostgresRepository) UpdateOrderStatus(ctx context.Context, number string, status domain.OrderStatus, accrual float64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders 
		 SET status = $1, accrual = $2 
		 WHERE number = $3`,
		status, accrual, number,
	)
	if err != nil {
		return fmt.Errorf("error updating order status: %w", err)
	}
	return nil
}

// GetBalance возвращает баланс пользователя
func (r *PostgresRepository) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	var balance domain.Balance
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(accrual), 0) as current,
		        COALESCE(SUM(CASE WHEN status = $1 THEN accrual ELSE 0 END), 0) as withdrawn
		 FROM orders 
		 WHERE user_id = $2`,
		domain.OrderStatusProcessed, userID,
	).Scan(&balance.Current, &balance.Withdrawn)

	if err != nil {
		return nil, fmt.Errorf("error getting balance: %w", err)
	}
	return &balance, nil
}

// CreateWithdrawal создает новое списание
func (r *PostgresRepository) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO withdrawals (user_id, order_number, sum) 
		 VALUES ($1, $2, $3)`,
		userID, orderNumber, sum,
	)
	if err != nil {
		return fmt.Errorf("error creating withdrawal: %w", err)
	}
	return nil
}

// GetUserWithdrawals возвращает все списания пользователя
func (r *PostgresRepository) GetUserWithdrawals(ctx context.Context, userID int64) ([]domain.Withdrawal, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT order_number, sum, processed_at 
		 FROM withdrawals 
		 WHERE user_id = $1 
		 ORDER BY processed_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting user withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []domain.Withdrawal
	for rows.Next() {
		var w domain.Withdrawal
		err := rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, w)
	}
	return withdrawals, nil
}
