// Package echomw provides the canonical Echo middleware stack shared
// across all bmkr Go services.
package echomw

import (
	"log/slog"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupMiddleware registers the canonical middleware stack onto e.
//
// Middleware registered (in order):
//  1. RequestID — generates a UUID and bridges it to the X-Request-Id
//     request header so the Connect interceptor can read it via
//     req.Header().Get(echo.HeaderXRequestID).
//  2. RequestLogger — logs canonical log lines for non-RPC requests.
//     Fields: method="<METHOD> <URI>", status="ok"|"error",
//     duration_ms, request_id.
//     Skipper: paths containing "." are skipped (those are Connect RPC
//     routes logged by the Connect interceptor instead, per ADR-0016).
//
// SetupMiddleware does not return an error; Echo middleware registration
// is always successful.
func SetupMiddleware(e *echo.Echo) {
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
}
