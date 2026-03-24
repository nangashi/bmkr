package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

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

	// RequestID MW: generate/propagate X-Request-Id
	// RequestIDHandler bridges the ID to the request header so that
	// the Connect interceptor can read it via req.Header().
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		RequestIDHandler: func(c echo.Context, requestID string) {
			c.Request().Header.Set(echo.HeaderXRequestID, requestID)
		},
	}))

	// Request logger for non-RPC endpoints (health, etc.)
	// Skipper: paths containing "." are RPC routes (e.g. /ec.v1.CartService/*)
	// and are logged by the Connect interceptor instead.
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogMethod:    true,
		LogLatency:   true,
		LogError:     true,
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

// wip: newLoggingInterceptor returns a Connect unary interceptor for canonical log line output.
// wip:
// wip: 動作:
// wip:   - リクエスト開始時刻を time.Now() で記録する
// wip:   - next ハンドラを呼び出す
// wip:   - time.Since(start) で duration を計算する
// wip:   - err == nil なら status = "ok"、err != nil なら status = "error"
// wip:   - req.Header().Get("X-Request-Id") で request_id を取得する
// wip:   - req.Spec().Procedure で RPC メソッド名を取得する
// wip:   - slog.InfoContext で canonical log line を1行出力する
// wip:     フィールド: method, status, duration_ms, request_id
func newLoggingInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			panic("not implemented")
		}
	}
}
