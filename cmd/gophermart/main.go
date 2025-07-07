package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gophermart/internal/app"
	"gophermart/internal/config"
	"gophermart/internal/handler"
	"gophermart/internal/storage"
	"gophermart/internal/usecase"
	"gophermart/pkg/jwt"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Инициализируем хранилище
	store, err := storage.NewPostgresRepository(context.Background(), cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Инициализируем JWT manager
	jwtManager := jwt.NewManager(cfg.JWT.SigningKey, cfg.JWT.TokenTTL)

	// Инициализируем usecase и обработчики
	userUseCase := usecase.NewUserUseCase(store, jwtManager)
	authHandler := handler.NewAuthHandler(userUseCase)
	h := handler.NewHandler(authHandler)
	router := handler.NewRouter(h)

	// Создаем сервер
	srv := app.NewServer(cfg.RunAddress, router)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := srv.Run(); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Printf("failed to stop server: %v", err)
	}

	log.Println("server stopped")
}
