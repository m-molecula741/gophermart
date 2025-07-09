package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Config содержит все настройки приложения
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWT                  JWTConfig
}

// JWTConfig содержит настройки JWT
type JWTConfig struct {
	SigningKey []byte
	TokenTTL   time.Duration
}

// NewConfig создает новый экземпляр конфигурации
func NewConfig() (*Config, error) {
	var cfg Config

	// Создаем новый FlagSet
	flags := flag.NewFlagSet("config", flag.ContinueOnError)

	// Чтение флагов командной строки
	flags.StringVar(&cfg.RunAddress, "a", "", "address and port to run server")
	flags.StringVar(&cfg.DatabaseURI, "d", "", "database connection string")
	flags.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system address")

	// Парсим флаги только если не в тестовом режиме
	if len(os.Args) > 0 {
		flags.Parse(os.Args[1:])
	}

	// Чтение переменных окружения, если флаги не установлены
	if cfg.RunAddress == "" {
		cfg.RunAddress = os.Getenv("RUN_ADDRESS")
	}
	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = os.Getenv("DATABASE_URI")
	}
	if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	}

	// Настройки JWT по умолчанию
	cfg.JWT = JWTConfig{
		SigningKey: []byte("your-secret-key"),
		TokenTTL:   24 * time.Hour,
	}

	// Валидация конфигурации
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validate проверяет корректность конфигурации
func (c *Config) validate() error {
	if c.RunAddress == "" {
		return fmt.Errorf("server address is required (use -a flag or RUN_ADDRESS env)")
	}
	if c.DatabaseURI == "" {
		return fmt.Errorf("database URI is required (use -d flag or DATABASE_URI env)")
	}
	if c.AccrualSystemAddress == "" {
		return fmt.Errorf("accrual system address is required (use -r flag or ACCRUAL_SYSTEM_ADDRESS env)")
	}
	return nil
}
