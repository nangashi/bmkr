// Package echomw provides the canonical Echo middleware stack shared
// across all bmkr Go services.
package echomw

import "github.com/labstack/echo/v4"

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
// wip: use middleware.RequestIDWithConfig with RequestIDHandler that calls
// c.Request().Header.Set(echo.HeaderXRequestID, requestID).
// Use middleware.RequestLoggerWithConfig with LogStatus, LogURI, LogMethod,
// LogLatency, LogRequestID all true. In LogValuesFunc derive status
// "ok"/"error" from v.Status>=400 and call slog.InfoContext with the
// canonical fields. Set Skipper to return strings.Contains(c.Path(), ".").
func SetupMiddleware(e *echo.Echo) {
	panic("not implemented")
}
