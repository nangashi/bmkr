package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/nangashi/bmkr/gen/go/product/v1/productv1connect"
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

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	queries := db.New(pool)

	e := echo.New()

	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		RequestIDHandler: func(c echo.Context, requestID string) {
			c.Request().Header.Set(echo.HeaderXRequestID, requestID)
		},
	}))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogMethod:    true,
		LogLatency:   true,
		LogRequestID: true,
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), ".")
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			status := "ok"
			if v.Status >= 400 {
				status = "error"
			}
			slog.InfoContext(c.Request().Context(), "request completed",
				"method", v.Method+" "+v.URI,
				"status", status,
				"duration_ms", v.Latency.Milliseconds(),
				"request_id", v.RequestID,
			)
			return nil
		},
	}))

	e.GET("/health", func(c echo.Context) error {
		if err := pool.Ping(c.Request().Context()); err != nil {
			slog.ErrorContext(c.Request().Context(), "health check failed", "error", err)
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	path, handler := productv1connect.NewProductServiceHandler(
		&ProductServiceHandler{store: queries},
		connect.WithInterceptors(newLoggingInterceptor()),
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("product-mgmt service starting", "port", port)
		serveErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", "error", err)
			pool.Close()
			os.Exit(1)
		}
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("shutting down server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown timed out, forcing close", "error", err)
		_ = server.Close()
		pool.Close()
		os.Exit(1)
	}
	pool.Close()
	slog.Info("server stopped")
}

// newLoggingInterceptor returns a Connect unary interceptor that outputs
// a canonical log line with method, status, duration_ms, and request_id.
func newLoggingInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			resp, err := next(ctx, req)
			duration := time.Since(start)

			status := "ok"
			if err != nil {
				status = "error"
			}

			slog.InfoContext(ctx, "request completed",
				"method", req.Spec().Procedure,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"request_id", req.Header().Get(echo.HeaderXRequestID),
			)
			return resp, err
		}
	}
}
