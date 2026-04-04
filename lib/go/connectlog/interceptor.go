// Package connectlog provides a Connect unary interceptor for canonical
// request logging per ADR-0016.
package connectlog

import "connectrpc.com/connect"

// NewLoggingInterceptor returns a Connect unary interceptor that outputs
// a canonical log line for every unary RPC call.
//
// wip: wrap next(ctx, req) with timing, derive status "ok"/"error" from err,
// then call slog.InfoContext with fields: method=req.Spec().Procedure,
// status, duration_ms, request_id=req.Header().Get("X-Request-Id").
// Pass through resp and err unchanged to the caller.
// Note: use string literal "X-Request-Id" instead of echo.HeaderXRequestID
// to avoid coupling this package to the Echo framework.
func NewLoggingInterceptor() connect.UnaryInterceptorFunc {
	panic("not implemented")
}
