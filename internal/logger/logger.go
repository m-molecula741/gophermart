package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Initialize инициализирует логгер
func Initialize(level string) error {
	// Настраиваем конфигурацию
	config := zap.NewProductionConfig()

	// Устанавливаем уровень логирования
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Настраиваем формат времени
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Создаем логгер
	var err error
	log, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

// Info логирует информационное сообщение
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Error логирует ошибку
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

// Debug логирует отладочное сообщение
func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

// Warn логирует предупреждение
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Sync сбрасывает буферы логгера
func Sync() error {
	return log.Sync()
}
