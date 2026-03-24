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
	defer pool.Close()

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
		Addr:    ":" + port,
		Handler: h2c.NewHandler(e, &http2.Server{}),
	}

	slog.Info("product-mgmt service starting", "port", port)
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
