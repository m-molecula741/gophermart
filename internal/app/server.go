package app

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	readTimeout     = 5 * time.Second
	writeTimeout    = 5 * time.Second
	shutdownTimeout = 3 * time.Second
)

type Server struct {
	server *http.Server
}

func NewServer(address string, handler http.Handler) *Server {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Монтируем обработчики
	r.Mount("/", handler)

	s := &http.Server{
		Addr:         address,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return &Server{
		server: s,
	}
}

// Run запускает сервер
func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

// Stop останавливает сервер
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
