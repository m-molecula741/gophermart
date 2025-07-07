package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter создает и настраивает роутер
func NewRouter(h *Handler) chi.Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Post("/api/user/register", h.auth.Register)
	r.Post("/api/user/login", h.auth.Login)

	// Protected routes будут добавлены позже
	// r.Group(func(r chi.Router) {
	//     r.Use(h.auth.AuthMiddleware)
	//     // Protected endpoints
	// })

	return r
}
 