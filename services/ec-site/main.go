package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
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
		// 動作: DB接続失敗時に slog.Error でエラーをログ出力し、os.Exit(1) でプロセスを終了する
		// フィールド規約: スネークケース（"error"）
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

	e.GET("/health", func(c echo.Context) error {
		if err := pool.Ping(c.Request().Context()); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	cartPath, cartHandler := ecv1connect.NewCartServiceHandler(&CartServiceHandler{queries: queries, productClient: productClient})
	e.Any(cartPath+"*", echo.WrapHandler(cartHandler))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: h2c.NewHandler(e, &http2.Server{}),
	}

	// 動作: サーバー起動時に slog.Info でサービス名とポート番号をログ出力する
	// フィールド規約: スネークケース（"port"）
	slog.Info("ec-site service starting", "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// 動作: サーバー起動失敗時に slog.Error でエラーをログ出力し、os.Exit(1) でプロセスを終了する
		// フィールド規約: スネークケース（"error"）
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
