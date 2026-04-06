package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	customerv1 "github.com/nangashi/bmkr/gen/go/customer/v1"
	db "github.com/nangashi/bmkr/services/customer-mgmt/db/generated"
)

// ---------------------------------------------------------------------------
// Mock
// ---------------------------------------------------------------------------

type mockCustomerStore struct {
	CreateCustomerFn func(ctx context.Context, arg db.CreateCustomerParams) (db.Customer, error)
	GetCustomerFn    func(ctx context.Context, id int64) (db.Customer, error)
}

func (m *mockCustomerStore) CreateCustomer(ctx context.Context, arg db.CreateCustomerParams) (db.Customer, error) {
	if m.CreateCustomerFn != nil {
		return m.CreateCustomerFn(ctx, arg)
	}
	return db.Customer{}, errors.New("CreateCustomer not implemented")
}

func (m *mockCustomerStore) GetCustomer(ctx context.Context, id int64) (db.Customer, error) {
	if m.GetCustomerFn != nil {
		return m.GetCustomerFn(ctx, id)
	}
	return db.Customer{}, errors.New("GetCustomer not implemented")
}

var _ customerStore = (*mockCustomerStore)(nil)

// ---------------------------------------------------------------------------
// Tests — CreateCustomer: bcrypt ハッシュ化
// ---------------------------------------------------------------------------

func TestCreateCustomer_PasswordIsHashed(t *testing.T) {
	password := "my-secret-password"
	var captured db.CreateCustomerParams

	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)
	store := &mockCustomerStore{
		CreateCustomerFn: func(_ context.Context, arg db.CreateCustomerParams) (db.Customer, error) {
			captured = arg
			return db.Customer{
				ID:           1,
				Name:         arg.Name,
				Email:        arg.Email,
				PasswordHash: arg.PasswordHash,
				CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
			}, nil
		},
	}
	h := &CustomerServiceHandler{store: store}

	resp, err := h.CreateCustomer(
		context.Background(),
		connect.NewRequest(&customerv1.CreateCustomerRequest{
			Name:     "テストユーザー",
			Email:    "test@example.com",
			Password: password,
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg == nil {
		t.Fatal("response message should not be nil")
	}

	// DB に渡された値が平文パスワードではないこと
	if captured.PasswordHash == password {
		t.Error("PasswordHash must not be the plain password")
	}

	// bcrypt ハッシュ文字列であること（$2a$ プレフィックス）
	if !strings.HasPrefix(captured.PasswordHash, "$2a$") && !strings.HasPrefix(captured.PasswordHash, "$2b$") {
		t.Errorf("PasswordHash should start with $2a$ or $2b$, got %q", captured.PasswordHash)
	}

	// bcrypt.CompareHashAndPassword で検証可能であること
	if err := bcrypt.CompareHashAndPassword([]byte(captured.PasswordHash), []byte(password)); err != nil {
		t.Errorf("bcrypt.CompareHashAndPassword failed: %v", err)
	}
}

func TestCreateCustomer_HashErrorReturnsInvalidArgument(t *testing.T) {
	// bcrypt は 72 バイト超のパスワードでエラーを返す
	longPassword := strings.Repeat("a", 73)

	store := &mockCustomerStore{
		CreateCustomerFn: func(_ context.Context, arg db.CreateCustomerParams) (db.Customer, error) {
			t.Fatal("store.CreateCustomer should not be called on hash error")
			return db.Customer{}, nil
		},
	}
	h := &CustomerServiceHandler{store: store}

	_, err := h.CreateCustomer(
		context.Background(),
		connect.NewRequest(&customerv1.CreateCustomerRequest{
			Name:     "テストユーザー",
			Email:    "test@example.com",
			Password: longPassword,
		}),
	)
	if err == nil {
		t.Fatal("expected error for password too long, got nil")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected *connect.Error, got %T: %v", err, err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
	}
}

func TestCreateCustomer_DBErrorReturnsInternal(t *testing.T) {
	store := &mockCustomerStore{
		CreateCustomerFn: func(_ context.Context, _ db.CreateCustomerParams) (db.Customer, error) {
			return db.Customer{}, errors.New("database connection lost")
		},
	}
	h := &CustomerServiceHandler{store: store}

	_, err := h.CreateCustomer(
		context.Background(),
		connect.NewRequest(&customerv1.CreateCustomerRequest{
			Name:     "テストユーザー",
			Email:    "test@example.com",
			Password: "valid-password",
		}),
	)
	if err == nil {
		t.Fatal("expected error for DB failure, got nil")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected *connect.Error, got %T: %v", err, err)
	}
	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeInternal)
	}
}

// ---------------------------------------------------------------------------
// Tests — CreateCustomer: 入力バリデーション
// ---------------------------------------------------------------------------

func TestCreateCustomer_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  *customerv1.CreateCustomerRequest
	}{
		{
			name: "name が空文字",
			req: &customerv1.CreateCustomerRequest{
				Name:     "",
				Email:    "test@example.com",
				Password: "valid-password",
			},
		},
		{
			name: "email が空文字",
			req: &customerv1.CreateCustomerRequest{
				Name:     "テストユーザー",
				Email:    "",
				Password: "valid-password",
			},
		},
		{
			name: "email が不正フォーマット（@なし）",
			req: &customerv1.CreateCustomerRequest{
				Name:     "テストユーザー",
				Email:    "invalid-email",
				Password: "valid-password",
			},
		},
		{
			name: "email が不正フォーマット（ドメインなし）",
			req: &customerv1.CreateCustomerRequest{
				Name:     "テストユーザー",
				Email:    "user@",
				Password: "valid-password",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockCustomerStore{
				CreateCustomerFn: func(_ context.Context, _ db.CreateCustomerParams) (db.Customer, error) {
					t.Fatal("store.CreateCustomer should not be called on validation error")
					return db.Customer{}, nil
				},
			}
			h := &CustomerServiceHandler{store: store}

			_, err := h.CreateCustomer(
				context.Background(),
				connect.NewRequest(tt.req),
			)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var connectErr *connect.Error
			if !errors.As(err, &connectErr) {
				t.Fatalf("expected *connect.Error, got %T: %v", err, err)
			}
			if connectErr.Code() != connect.CodeInvalidArgument {
				t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests — CreateCustomer: bcrypt ErrPasswordTooLong のメッセージ検証
// ---------------------------------------------------------------------------

func TestCreateCustomer_HashErrorTooLong_ReturnsGenericMessage(t *testing.T) {
	// bcrypt は 72 バイト超のパスワードで ErrPasswordTooLong を返す
	longPassword := strings.Repeat("a", 73)

	store := &mockCustomerStore{
		CreateCustomerFn: func(_ context.Context, _ db.CreateCustomerParams) (db.Customer, error) {
			t.Fatal("store.CreateCustomer should not be called on hash error")
			return db.Customer{}, nil
		},
	}
	h := &CustomerServiceHandler{store: store}

	_, err := h.CreateCustomer(
		context.Background(),
		connect.NewRequest(&customerv1.CreateCustomerRequest{
			Name:     "テストユーザー",
			Email:    "test@example.com",
			Password: longPassword,
		}),
	)
	if err == nil {
		t.Fatal("expected error for password too long, got nil")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected *connect.Error, got %T: %v", err, err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
	}
	// エラーメッセージに bcrypt の内部エラー文字列が漏れていないこと
	if strings.Contains(connectErr.Message(), bcrypt.ErrPasswordTooLong.Error()) {
		t.Errorf("error message should not expose bcrypt internal error, got %q", connectErr.Message())
	}
}

// ---------------------------------------------------------------------------
// Tests — GetCustomer RPC: バリデーション
// ---------------------------------------------------------------------------

func TestGetCustomer_Validation(t *testing.T) {
	tests := []struct {
		name string
		id   int64
	}{
		{name: "id が 0", id: 0},
		{name: "id が負数", id: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockCustomerStore{
				GetCustomerFn: func(_ context.Context, _ int64) (db.Customer, error) {
					t.Fatal("store.GetCustomer should not be called on validation error")
					return db.Customer{}, nil
				},
			}
			h := &CustomerServiceHandler{store: store}

			_, err := h.GetCustomer(
				context.Background(),
				connect.NewRequest(&customerv1.GetCustomerRequest{Id: tt.id}),
			)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var connectErr *connect.Error
			if !errors.As(err, &connectErr) {
				t.Fatalf("expected *connect.Error, got %T: %v", err, err)
			}
			if connectErr.Code() != connect.CodeInvalidArgument {
				t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
			}
		})
	}
}
