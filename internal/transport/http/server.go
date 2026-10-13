package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	mw "github.com/AndroDeMohawk/MovieVerse/internal/transport/http/middleware"
	"github.com/AndroDeMohawk/MovieVerse/internal/transport/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	httpServer *http.Server
	log        *slog.Logger
}

func NewServer(port string, rdb *redis.Client, wsHub *ws.Hub, log *slog.Logger) *Server {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/ws", func(r chi.Router) {
		r.Use(mw.RateLimiter(rdb, 10, time.Minute))
		r.Get("/movies/{id}", func(w http.ResponseWriter, r *http.Request) {
			ws.ServeWS(wsHub, w, r)
		})
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: srv,
		log:        log,
	}
}

func (s *Server) Run() error {
	const op = "http.Server.Run"

	s.log.Info("starting HTTP/WebSocket server", slog.String("addr", s.httpServer.Addr))

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	const op = "http.Server.Stop"

	s.log.Info("stopping HTTP server", slog.String("addr", s.httpServer.Addr))

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
