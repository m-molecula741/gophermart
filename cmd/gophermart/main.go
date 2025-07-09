package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gophermart/internal/accrual"
	"gophermart/internal/app"
	"gophermart/internal/config"
	"gophermart/internal/handler"
	"gophermart/internal/logger"
	"gophermart/internal/storage"
	"gophermart/internal/usecase"
	"gophermart/pkg/jwt"

	"go.uber.org/zap"
)

func main() {
	// Инициализируем логгер
	if err := logger.Initialize("info"); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting GopherMart service")

	// Загружаем конфигурацию
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Error("Failed to load config", zap.Error(err))
		os.Exit(1)
	}
	logger.Info("Config loaded successfully",
		zap.String("run_address", cfg.RunAddress),
		zap.String("database_uri", cfg.DatabaseURI),
		zap.String("accrual_address", cfg.AccrualSystemAddress))

	// Инициализируем хранилище
	store, err := storage.NewPostgresRepository(context.Background(), cfg.DatabaseURI)
	if err != nil {
		logger.Error("Failed to initialize storage", zap.Error(err))
		os.Exit(1)
	}
	defer store.Close()
	logger.Info("Storage initialized successfully")

	// Инициализируем JWT manager
	jwtManager := jwt.NewManager(cfg.JWT.SigningKey, cfg.JWT.TokenTTL)
	logger.Info("JWT manager initialized",
		zap.Duration("token_ttl", cfg.JWT.TokenTTL))

	// Инициализируем сервис начислений
	accrualService := accrual.NewService(cfg.AccrualSystemAddress)
	logger.Info("Accrual service initialized",
		zap.String("address", cfg.AccrualSystemAddress))

	// Инициализируем usecase и обработчики
	userUseCase := usecase.NewUserUseCase(store, jwtManager)
	orderUseCase := usecase.NewOrderUseCase(store, accrualService)
	balanceUseCase := usecase.NewBalanceUseCase(store)

	authHandler := handler.NewAuthHandler(userUseCase)
	orderHandler := handler.NewOrderHandler(orderUseCase)
	balanceHandler := handler.NewBalanceHandler(balanceUseCase)

	h := handler.NewHandler(authHandler, orderHandler, balanceHandler)
	router := handler.NewRouter(h)
	logger.Info("Handlers initialized successfully")

	// Создаем сервер
	srv := app.NewServer(cfg.RunAddress, router)

	// Запускаем сервер в отдельной горутине
	go func() {
		logger.Info("Starting HTTP server", zap.String("address", cfg.RunAddress))
		if err := srv.Run(); err != nil {
			logger.Error("Server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	logger.Info("Shutting down server...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем обработку заказов
	orderUseCase.Shutdown(ctx)

	// Останавливаем HTTP сервер
	if err := srv.Stop(ctx); err != nil {
		logger.Error("Failed to stop server", zap.Error(err))
	}

	// Закрываем соединение с БД
	if err := store.Close(); err != nil {
		logger.Error("Failed to close database connection", zap.Error(err))
	}

	logger.Info("Server stopped")
}
