package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/munchies/platform/backend/internal/config"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/platform/sms"
	"github.com/munchies/platform/backend/internal/server"
	redisclient "github.com/munchies/platform/backend/internal/platform/redis"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	setupLogger(cfg.Server.Environment)

	log.Info().
		Str("environment", string(cfg.Server.Environment)).
		Int("port", cfg.Server.Port).
		Msg("starting munchies platform API")

	ctx := context.Background()

	// Connect to PostgreSQL
	pool, err := connectDB(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	queries := sqlc.New(pool)

	// Connect to Redis (optional — app starts without it in dev)
	var redisClient *redisclient.Client
	if cfg.Redis.URL != "" {
		redisClient, err = redisclient.New(cfg.Redis.URL)
		if err != nil {
			log.Warn().Err(err).Msg("failed to connect to redis — running without caching")
		} else {
			defer redisClient.Close()
		}
	}

	// SMS sender
	var smsSender sms.Sender
	if cfg.Services.SMSAPIKey != "" && cfg.Services.SMSBaseURL != "" {
		smsSender = sms.NewSSLWireless(cfg.Services.SMSAPIKey, cfg.Services.SMSBaseURL)
	} else {
		log.Warn().Msg("SMS credentials not configured — using noop sender")
		smsSender = &sms.NoopSender{}
	}

	deps := server.Deps{
		Queries: queries,
		Pool:    pool,
		Redis:   redisClient,
		SMS:     smsSender,
	}

	srv := server.New(cfg, deps)

	// Start background jobs (reconciliation, etc.)
	jobCtx, jobCancel := context.WithCancel(ctx)
	defer jobCancel()
	srv.StartBackgroundJobs(jobCtx)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      srv.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", httpServer.Addr).Msg("HTTP server listening")
		serverErr <- httpServer.ListenAndServe()
	}()

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
		os.Exit(1)
	}

	log.Info().Msg("server stopped gracefully")
}

func connectDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not configured")
	}
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("parse db url: %w", err)
	}
	poolCfg.MaxConns = int32(cfg.Database.MaxOpenConns)
	poolCfg.MaxConnLifetime = cfg.Database.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return pool, nil
}

func setupLogger(env config.Environment) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if env.IsDevelopment() {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		log.Logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Caller().
			Str("service", "munchies-api").
			Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

