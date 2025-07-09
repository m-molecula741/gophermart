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
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Public routes
	r.Post("/api/user/register", h.auth.Register)
	r.Post("/api/user/login", h.auth.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(h.auth.AuthMiddleware)
		r.Post("/api/user/orders", h.order.UploadOrder)
		r.Get("/api/user/orders", h.order.GetOrders)
	})

	return r
}
