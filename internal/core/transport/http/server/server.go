// Package server assembles the HTTP router and manages the http.Server lifecycle.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	"messenger/internal/core/config"
	"messenger/internal/core/transport/http/middleware"
)

// Registrar is implemented by feature transports to attach their routes.
type Registrar interface {
	Register(mux *http.ServeMux)
}

// Server wraps an *http.Server with graceful start/stop helpers.
type Server struct {
	httpServer *http.Server
	log        *zap.Logger
}

// New builds the router (health, swagger and all feature routes), wraps it with
// the global middleware stack and returns a ready-to-start Server.
func New(cfg config.HTTP, log *zap.Logger, registrars ...Registrar) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("/swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	for _, reg := range registrars {
		reg.Register(mux)
	}

	handler := middleware.Chain(mux,
		middleware.Recover(log),
		middleware.RequestLogger(log),
		middleware.CORS(),
	)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Addr(),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
		log: log,
	}
}

// Start begins serving and blocks until the server stops. A clean shutdown
// (via Shutdown) is not reported as an error.
func (s *Server) Start() error {
	s.log.Info("http server listening", zap.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully drains in-flight requests.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// healthHandler is a simple liveness probe.
//
//	@Summary		Liveness probe
//	@Description	Returns ok when the service is up.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
