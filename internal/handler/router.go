package handler

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter(h *Handler) chi.Router {
	router := chi.NewRouter()

	router.Get("/ping", h.Ping)

	return router
}
