package handler

import (
	"net/http"

	"gophermart/internal/usecase"
)

type Handler struct {
	storage usecase.Storage
}

func NewHandler(storage usecase.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// Ping проверяет доступность сервера и соединение с БД
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.storage.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
