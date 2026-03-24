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

	e.GET("/health", func(c echo.Context) error {
		if err := pool.Ping(c.Request().Context()); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	path, handler := productv1connect.NewProductServiceHandler(&ProductServiceHandler{queries: queries})
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
