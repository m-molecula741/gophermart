package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gophermart/internal/domain"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
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
		`INSERT INTO orders (number, user_id, status, uploaded_at) 
		 VALUES ($1, $2, $3, $4)`,
		number, userID, domain.StatusNew, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("error creating order: %w", err)
	}
	return nil
}

// GetOrderByNumber находит заказ по номеру
func (r *PostgresRepository) GetOrderByNumber(ctx context.Context, number string) (*domain.Order, error) {
	var order domain.Order
	var processedAt sql.NullTime

	err := r.pool.QueryRow(ctx,
		`SELECT number, user_id, status, accrual, uploaded_at, processed_at 
		 FROM orders 
		 WHERE number = $1`,
		number,
	).Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
		&processedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("error getting order by number: %w", err)
	}

	if processedAt.Valid {
		order.ProcessedAt = &processedAt.Time
	}

	return &order, nil
}

// GetUserOrders возвращает все заказы пользователя
func (r *PostgresRepository) GetUserOrders(ctx context.Context, userID int64) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT number, user_id, status, accrual, uploaded_at, processed_at 
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
		var processedAt sql.NullTime

		err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
			&processedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning order: %w", err)
		}

		if processedAt.Valid {
			order.ProcessedAt = &processedAt.Time
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatusAndBalance атомарно обновляет статус заказа и баланс пользователя
func (r *PostgresRepository) UpdateOrderStatusAndBalance(ctx context.Context, number string, status domain.OrderStatus, accrual float64, userID int64) error {
	// Получаем соединение из пула
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error acquiring connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Обновляем статус заказа
	_, err = tx.Exec(ctx,
		`UPDATE orders 
         SET status = $1, accrual = $2, processed_at = $3 
         WHERE number = $4`,
		status, accrual, time.Now(), number,
	)
	if err != nil {
		return fmt.Errorf("error updating order status: %w", err)
	}

	// Если статус PROCESSED и есть начисление, обновляем баланс
	if status == domain.StatusProcessed && accrual > 0 {
		// Проверяем существование записи в таблице balances
		var exists bool
		err := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM balances WHERE user_id = $1)`,
			userID,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking balance existence: %w", err)
		}

		if exists {
			// Обновляем существующий баланс
			_, err = tx.Exec(ctx,
				`UPDATE balances 
                 SET current = current + $1 
                 WHERE user_id = $2`,
				accrual, userID,
			)
		} else {
			// Создаем новую запись баланса
			_, err = tx.Exec(ctx,
				`INSERT INTO balances (user_id, current, withdrawn) 
                 VALUES ($1, $2, 0)`,
				userID, accrual,
			)
		}
		if err != nil {
			return fmt.Errorf("error updating balance: %w", err)
		}
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// GetBalance возвращает баланс пользователя
func (r *PostgresRepository) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	var balance domain.Balance

	// Получаем текущий баланс из таблицы balances
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(current, 0), COALESCE(withdrawn, 0)
		 FROM balances 
		 WHERE user_id = $1`,
		userID,
	).Scan(&balance.Current, &balance.Withdrawn)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Если записи нет, возвращаем нулевой баланс
			return &domain.Balance{Current: 0, Withdrawn: 0}, nil
		}
		return nil, fmt.Errorf("error getting balance: %w", err)
	}

	return &balance, nil
}

// CreateWithdrawal создает новое списание
func (r *PostgresRepository) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	// Получаем соединение из пула
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error acquiring connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Создаем запись о списании
	_, err = tx.Exec(ctx,
		`INSERT INTO withdrawals (user_id, order_number, sum) 
		 VALUES ($1, $2, $3)`,
		userID, orderNumber, sum,
	)
	if err != nil {
		return fmt.Errorf("error creating withdrawal: %w", err)
	}

	// Обновляем баланс пользователя
	result, err := tx.Exec(ctx,
		`UPDATE balances 
		 SET current = current - $1,
		     withdrawn = withdrawn + $1
		 WHERE user_id = $2 AND current >= $1`,
		sum, userID,
	)
	if err != nil {
		return fmt.Errorf("error updating balance: %w", err)
	}

	// Проверяем, что обновление баланса произошло
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrInsufficientFunds
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
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
