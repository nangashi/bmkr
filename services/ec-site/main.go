package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/nangashi/bmkr/gen/go/ec/v1/ecv1connect"
	db "github.com/nangashi/bmkr/services/ec-site/db/generated"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/ecsite?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	queries := db.New(pool)

	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
	if productServiceURL == "" {
		productServiceURL = "http://localhost:8081"
	}
	productClient := newProductServiceClient(productServiceURL)

	e := echo.New()

	// RequestIDHandler bridges the ID to the request header so that
	// the Connect interceptor can read it via req.Header().
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		RequestIDHandler: func(c echo.Context, requestID string) {
			c.Request().Header.Set(echo.HeaderXRequestID, requestID)
		},
	}))

	// Skipper: paths containing "." are RPC routes (e.g. /ec.v1.CartService/*)
	// and are logged by the Connect interceptor instead.
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
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	cartPath, cartHandler := ecv1connect.NewCartServiceHandler(
		&CartServiceHandler{queries: queries, productClient: productClient},
		connect.WithInterceptors(newLoggingInterceptor()),
	)
	e.Any(cartPath+"*", echo.WrapHandler(cartHandler))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: h2c.NewHandler(e, &http2.Server{}),
	}

	slog.Info("ec-site service starting", "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
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
