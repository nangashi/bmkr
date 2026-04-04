package connectlog_test

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
	"github.com/nangashi/bmkr/lib/go/connectlog"
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
// Helpers: slog capture
// ---------------------------------------------------------------------------

// logRecord はキャプチャしたログレコードを保持する。
type logRecord struct {
	Level   slog.Level
	Message string
	Attrs   map[string]any
}

// captureSlog は slog.SetDefault でカスタムハンドラを設定し、
// ログ出力をキャプチャするヘルパー。テスト終了後に元のデフォルトに復元する。
func captureSlog(t *testing.T) *[]logRecord {
	t.Helper()
	original := slog.Default()
	t.Cleanup(func() { slog.SetDefault(original) })

	var records []logRecord
	handler := &captureHandler{records: &records}
	slog.SetDefault(slog.New(handler))
	return &records
}

// captureHandler は slog.Handler のテスト用実装。
type captureHandler struct {
	records *[]logRecord
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	rec := logRecord{
		Level:   r.Level,
		Message: r.Message,
		Attrs:   make(map[string]any),
	}
	r.Attrs(func(a slog.Attr) bool {
		rec.Attrs[a.Key] = a.Value.Any()
		return true
	})
	*h.records = append(*h.records, rec)
	return nil
}

func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

// ---------------------------------------------------------------------------
// Tests: connectlog.NewLoggingInterceptor の canonical log line 出力
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
				connect.WithInterceptors(connectlog.NewLoggingInterceptor()),
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

			// request_id = リクエストヘッダの X-Request-Id (string literal, not echo constant)
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
