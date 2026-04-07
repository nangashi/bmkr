package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/nangashi/bmkr/gen/go/product/v1/productv1connect"
	"github.com/nangashi/bmkr/lib/go/connectlog"
	"github.com/nangashi/bmkr/lib/go/echomw"
	"github.com/nangashi/bmkr/lib/go/shutdown"
	db "github.com/nangashi/bmkr/services/product-mgmt/db/generated"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/product?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		slog.Error("failed to parse database config", "error", err)
		os.Exit(1)
	}
	// MaxConns: 5 — admin/management service with low request frequency;
	//   3 services total 30 connections << PostgreSQL default max_connections (100)
	config.MaxConns = 5
	// MinConns: 2 — keep warm connections to reduce first-request latency
	config.MinConns = 2
	// ConnectTimeout: 5s — Docker Compose internal communication is typically milliseconds;
	//   5s detects DB startup delays early while staying well within WriteTimeout (30s)
	config.ConnConfig.ConnectTimeout = 5 * time.Second
	// MaxConnLifetime: 1h — pgx v5 default, made explicit for clarity
	config.MaxConnLifetime = 1 * time.Hour
	// MaxConnLifetimeJitter: 5min — prevents thundering herd when all 3 services recycle connections simultaneously
	config.MaxConnLifetimeJitter = 5 * time.Minute
	// MaxConnIdleTime: 30min — pgx v5 default, appropriate for development with intermittent requests
	config.MaxConnIdleTime = 30 * time.Minute
	// HealthCheckPeriod: 1min — pgx v5 default, detects DB restart recovery within 1 minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	queries := db.New(pool)

	e := echo.New()

	echomw.SetupMiddleware(e)

	e.GET("/health", func(c echo.Context) error {
		if err := pool.Ping(c.Request().Context()); err != nil {
			slog.ErrorContext(c.Request().Context(), "health check failed", "error", err)
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	path, handler := productv1connect.NewProductServiceHandler(
		&ProductServiceHandler{store: queries},
		connect.WithInterceptors(connectlog.NewLoggingInterceptor()),
	)
	e.Any(path+"*", echo.WrapHandler(handler))

	// 管理画面ルーティング
	adminHandler := NewAdminHandler(queries)
	admin := e.Group("/admin")
	admin.GET("/products", adminHandler.HandleProductList)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      h2c.NewHandler(e, &http2.Server{}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("product-mgmt service starting", "port", port)
		serveErr <- server.ListenAndServe()
	}()
	shutdown.GracefulShutdown(server, serveErr, pool.Close)
}
