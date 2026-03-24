package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	ecv1 "github.com/nangashi/bmkr/gen/go/ec/v1"
	"github.com/nangashi/bmkr/gen/go/ec/v1/ecv1connect"
)

// ---------------------------------------------------------------------------
// Test handler: interceptor テスト用の最小限の CartService 実装
// ---------------------------------------------------------------------------

type testCartServiceHandler struct {
	ecv1connect.UnimplementedCartServiceHandler
	getCartFn func(ctx context.Context, req *connect.Request[ecv1.GetCartRequest]) (*connect.Response[ecv1.GetCartResponse], error)
}

func (h *testCartServiceHandler) GetCart(ctx context.Context, req *connect.Request[ecv1.GetCartRequest]) (*connect.Response[ecv1.GetCartResponse], error) {
	return h.getCartFn(ctx, req)
}

// ---------------------------------------------------------------------------
// Tests: newLoggingInterceptor の canonical log line 出力
// ---------------------------------------------------------------------------

func TestNewLoggingInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		requestID  string
		getCartFn  func(ctx context.Context, req *connect.Request[ecv1.GetCartRequest]) (*connect.Response[ecv1.GetCartResponse], error)
		wantStatus string
		wantErr    bool
	}{
		{
			name:      "success: status=ok, canonical log line with all fields",
			requestID: "req-abc-123",
			getCartFn: func(_ context.Context, _ *connect.Request[ecv1.GetCartRequest]) (*connect.Response[ecv1.GetCartResponse], error) {
				return connect.NewResponse(&ecv1.GetCartResponse{}), nil
			},
			wantStatus: "ok",
			wantErr:    false,
		},
		{
			name:      "error: status=error when handler returns connect error",
			requestID: "req-def-456",
			getCartFn: func(_ context.Context, _ *connect.Request[ecv1.GetCartRequest]) (*connect.Response[ecv1.GetCartResponse], error) {
				return nil, connect.NewError(connect.CodeInternal, errors.New("db error"))
			},
			wantStatus: "error",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records := captureSlog(t)

			_, handler := ecv1connect.NewCartServiceHandler(
				&testCartServiceHandler{getCartFn: tt.getCartFn},
				connect.WithInterceptors(newLoggingInterceptor()),
			)

			server := httptest.NewServer(handler)
			t.Cleanup(server.Close)

			client := ecv1connect.NewCartServiceClient(http.DefaultClient, server.URL)
			req := connect.NewRequest(&ecv1.GetCartRequest{CustomerId: 1})
			req.Header().Set("X-Request-Id", tt.requestID)

			_, err := client.GetCart(context.Background(), req)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// canonical log line が1行だけ出力されること
			if len(*records) != 1 {
				t.Fatalf("expected 1 log record, got %d", len(*records))
			}
			rec := (*records)[0]

			if rec.Level != slog.LevelInfo {
				t.Errorf("level = %v, want INFO", rec.Level)
			}
			if rec.Message != "request completed" {
				t.Errorf("message = %q, want %q", rec.Message, "request completed")
			}

			// method = RPC メソッド名（Spec().Procedure）
			wantMethod := "/ec.v1.CartService/GetCart"
			if rec.Attrs["method"] != wantMethod {
				t.Errorf("method = %v, want %q", rec.Attrs["method"], wantMethod)
			}

			// status = "ok" or "error"
			if rec.Attrs["status"] != tt.wantStatus {
				t.Errorf("status = %v, want %q", rec.Attrs["status"], tt.wantStatus)
			}

			// request_id = リクエストヘッダの X-Request-Id
			if rec.Attrs["request_id"] != tt.requestID {
				t.Errorf("request_id = %v, want %q", rec.Attrs["request_id"], tt.requestID)
			}

			// duration_ms が非負の int64
			dms, ok := rec.Attrs["duration_ms"]
			if !ok {
				t.Error("missing duration_ms attribute")
			} else if d, ok := dms.(int64); !ok || d < 0 {
				t.Errorf("duration_ms = %v (%T), want non-negative int64", dms, dms)
			}
		})
	}
}
