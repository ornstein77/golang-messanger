// Command app is the entry point for the messenger HTTP backend.
//
//	@title			Messenger API
//	@version		1.0
//	@description	A small messenger backend. Register with a nickname, password and
//	@description	phone number, log in to obtain a bearer token, then browse users
//	@description	and exchange direct messages.
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//
//	@license.name	MIT
//
//	@host			localhost:8080
//	@BasePath		/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and the session token.
package main

import (
	"context"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	_ "messenger/docs"

	"messenger/internal/core/config"
	"messenger/internal/core/logger"
	"messenger/internal/core/repository/postgres/pool"
	"messenger/internal/core/transport/http/middleware"
	"messenger/internal/core/transport/http/server"

	authrepo "messenger/internal/features/auth/repository"
	authservice "messenger/internal/features/auth/service"
	authtransport "messenger/internal/features/auth/transport"

	messagesrepo "messenger/internal/features/messages/repository"
	messagesservice "messenger/internal/features/messages/service"
	messagestransport "messenger/internal/features/messages/transport"

	usersrepo "messenger/internal/features/users/repository"
	usersservice "messenger/internal/features/users/service"
	userstransport "messenger/internal/features/users/transport"
)

func main() {
	if err := run(); err != nil {
		// The logger may not be initialized yet, so fall back to stderr.
		panic(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log, err := logger.New(cfg.Log)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	log.Info("starting messenger", zap.String("env", cfg.Env))

	// Root context cancelled on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := pool.New(ctx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer dbPool.Close()
	log.Info("connected to postgres", zap.String("host", cfg.Postgres.Host))

	// Repositories.
	authRepository := authrepo.New(dbPool)
	usersRepository := usersrepo.New(dbPool)
	messagesRepository := messagesrepo.New(dbPool)

	// Services.
	authSvc := authservice.New(authRepository, cfg.Session)
	usersSvc := usersservice.New(usersRepository)
	messagesSvc := messagesservice.New(messagesRepository)

	// Auth middleware shared by protected features.
	authMW := middleware.Auth(authSvc, log)

	// Feature transports.
	authHandler := authtransport.New(authSvc, authMW, log)
	usersHandler := userstransport.New(usersSvc, authMW, log)
	messagesHandler := messagestransport.New(messagesSvc, authMW, log)

	srv := server.New(cfg.HTTP, log, authHandler, usersHandler, messagesHandler)

	// Start the server and wait for either a fatal error or a shutdown signal.
	serverErr := make(chan error, 1)
	go func() { serverErr <- srv.Start() }()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
		return err
	}

	log.Info("server stopped cleanly")
	return nil
}
