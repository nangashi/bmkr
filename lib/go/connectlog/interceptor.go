// Package connectlog provides a Connect unary interceptor for canonical
// request logging per ADR-0016.
package connectlog

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

// NewLoggingInterceptor returns a Connect unary interceptor that outputs
// a canonical log line for every unary RPC call.
//
// Log fields (slog, InfoContext):
//   - "method":      req.Spec().Procedure  (e.g. "/ec.v1.CartService/GetCart")
//   - "status":      "ok" | "error"
//   - "duration_ms": elapsed milliseconds
//   - "request_id":  X-Request-Id header value (uses string literal, not echo constant)
//
// The function is stateless and safe for concurrent use.
// Note: does NOT depend on the Echo package — uses "X-Request-Id" string literal.
func NewLoggingInterceptor() connect.UnaryInterceptorFunc {
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
				"request_id", req.Header().Get("X-Request-Id"),
			)
			return resp, err
		}
	}
}
