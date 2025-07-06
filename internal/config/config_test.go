package config

import (
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	// Очистка переменных окружения после тестов
	defer func() {
		os.Unsetenv("RUN_ADDRESS")
		os.Unsetenv("DATABASE_URI")
		os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")
	}()

	tests := []struct {
		name      string
		envVars   map[string]string
		wantError bool
	}{
		{
			name: "all env vars set",
			envVars: map[string]string{
				"RUN_ADDRESS":            "localhost:8080",
				"DATABASE_URI":           "postgres://localhost:5432/db",
				"ACCRUAL_SYSTEM_ADDRESS": "http://localhost:8081",
			},
			wantError: false,
		},
		{
			name: "missing run address",
			envVars: map[string]string{
				"DATABASE_URI":           "postgres://localhost:5432/db",
				"ACCRUAL_SYSTEM_ADDRESS": "http://localhost:8081",
			},
			wantError: true,
		},
		{
			name: "missing database uri",
			envVars: map[string]string{
				"RUN_ADDRESS":            "localhost:8080",
				"ACCRUAL_SYSTEM_ADDRESS": "http://localhost:8081",
			},
			wantError: true,
		},
		{
			name: "missing accrual address",
			envVars: map[string]string{
				"RUN_ADDRESS":  "localhost:8080",
				"DATABASE_URI": "postgres://localhost:5432/db",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Установка переменных окружения для теста
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Запуск теста
			cfg, err := NewConfig()

			// Проверка результатов
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if err == nil {
				// Проверка значений только если нет ошибки
				if v, ok := tt.envVars["RUN_ADDRESS"]; ok && cfg.RunAddress != v {
					t.Errorf("expected RunAddress %s, got %s", v, cfg.RunAddress)
				}
				if v, ok := tt.envVars["DATABASE_URI"]; ok && cfg.DatabaseURI != v {
					t.Errorf("expected DatabaseURI %s, got %s", v, cfg.DatabaseURI)
				}
				if v, ok := tt.envVars["ACCRUAL_SYSTEM_ADDRESS"]; ok && cfg.AccrualSystemAddress != v {
					t.Errorf("expected AccrualSystemAddress %s, got %s", v, cfg.AccrualSystemAddress)
				}
			}

			// Очистка переменных окружения после каждого теста
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}
