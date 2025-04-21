package api

import (
	"context"
	"net/http"
	"time"

	v1 "task-runner-service/internal/api/v1"

	"github.com/go-chi/chi/v5"
)

type HTTPConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Server struct {
	server *http.Server
}

func NewServer(cfg *HTTPConfig, handler *v1.Handler) *Server {
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	srv := &http.Server{
		Addr:         cfg.Host + ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &Server{
		server: srv,
	}
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
