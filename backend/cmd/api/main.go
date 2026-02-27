package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/munchies/platform/backend/internal/config"
	"github.com/munchies/platform/backend/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	setupLogger(cfg.Server.Environment)

	log.Info().
		Str("environment", string(cfg.Server.Environment)).
		Int("port", cfg.Server.Port).
		Msg("starting munchies platform API")

	// Create and configure the HTTP server
	srv := server.New(cfg)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      srv.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", httpServer.Addr).Msg("HTTP server listening")
		serverErr <- httpServer.ListenAndServe()
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed to start")
		}
	case sig := <-quit:
		log.Info().Str("signal", sig.String()).Msg("shutting down server")
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
		os.Exit(1)
	}

	log.Info().Msg("server stopped gracefully")
}

func setupLogger(env config.Environment) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if env.IsDevelopment() {
		// Pretty console output for local/dev
		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// Structured JSON output for staging/production
		log.Logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Caller().
			Str("service", "munchies-api").
			Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
