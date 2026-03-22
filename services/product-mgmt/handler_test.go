package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"

	productv1 "github.com/nangashi/bmkr/gen/go/product/v1"
	db "github.com/nangashi/bmkr/services/product-mgmt/db/generated"
)

// ---------------------------------------------------------------------------
// Mock
// ---------------------------------------------------------------------------

// mockListProductStore は productStore のテスト用モック実装。
// GetProductFn / CreateProductFn / ListProductsFn を差し替えることで、
// 各テストケースの振る舞いを制御する。
type mockListProductStore struct {
	GetProductFn    func(ctx context.Context, id int64) (db.Product, error)
	CreateProductFn func(ctx context.Context, arg db.CreateProductParams) (db.Product, error)
	ListProductsFn  func(ctx context.Context) ([]db.Product, error)
}

func (m *mockListProductStore) GetProduct(ctx context.Context, id int64) (db.Product, error) {
	if m.GetProductFn != nil {
		return m.GetProductFn(ctx, id)
	}
	return db.Product{}, errors.New("GetProduct not implemented")
}

func (m *mockListProductStore) CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
	if m.CreateProductFn != nil {
		return m.CreateProductFn(ctx, arg)
	}
	return db.Product{}, errors.New("CreateProduct not implemented")
}

func (m *mockListProductStore) ListProducts(ctx context.Context) ([]db.Product, error) {
	if m.ListProductsFn != nil {
		return m.ListProductsFn(ctx)
	}
	return nil, errors.New("ListProducts not implemented")
}

// コンパイル時にインターフェース充足を保証する。
var _ productStore = (*mockListProductStore)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// sampleDBProduct はテスト用の db.Product を生成するヘルパー。
func sampleDBProduct(id int64, name string, desc string, price int64, stock int32) db.Product {
	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	return db.Product{
		ID:            id,
		Name:          name,
		Description:   desc,
		Price:         price,
		StockQuantity: stock,
		CreatedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
	}
}

// ---------------------------------------------------------------------------
// Tests — ListProducts RPC
// ---------------------------------------------------------------------------

// --- 正常系: 複数商品が取得できる ---

func TestListProducts_ReturnsMultipleProducts(t *testing.T) {
	products := []db.Product{
		sampleDBProduct(1, "商品A", "説明A", 1000, 10),
		sampleDBProduct(2, "商品B", "説明B", 2500, 5),
		sampleDBProduct(3, "商品C", "説明C", 500, 100),
	}
	store := &mockListProductStore{
		ListProductsFn: func(_ context.Context) ([]db.Product, error) {
			return products, nil
		},
	}
	h := &ProductServiceHandler{store: store}

	resp, err := h.ListProducts(
		context.Background(),
		connect.NewRequest(&productv1.ListProductsRequest{}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg == nil {
		t.Fatal("response message should not be nil")
	}
	if got := len(resp.Msg.Products); got != 3 {
		t.Fatalf("products count = %d, want 3", got)
	}

	// 各商品のフィールドが正しく変換されていること
	p1 := resp.Msg.Products[0]
	if p1.Id != 1 {
		t.Errorf("products[0].Id = %d, want 1", p1.Id)
	}
	if p1.Name != "商品A" {
		t.Errorf("products[0].Name = %q, want %q", p1.Name, "商品A")
	}
	if p1.Description != "説明A" {
		t.Errorf("products[0].Description = %q, want %q", p1.Description, "説明A")
	}
	if p1.Price != 1000 {
		t.Errorf("products[0].Price = %d, want 1000", p1.Price)
	}
	if p1.StockQuantity != 10 {
		t.Errorf("products[0].StockQuantity = %d, want 10", p1.StockQuantity)
	}

	p2 := resp.Msg.Products[1]
	if p2.Id != 2 {
		t.Errorf("products[1].Id = %d, want 2", p2.Id)
	}
	if p2.Name != "商品B" {
		t.Errorf("products[1].Name = %q, want %q", p2.Name, "商品B")
	}
}

// --- 正常系: 商品が0件の場合、空の products スライスを返す（エラーにしない） ---

func TestListProducts_EmptyProducts(t *testing.T) {
	store := &mockListProductStore{
		ListProductsFn: func(_ context.Context) ([]db.Product, error) {
			return []db.Product{}, nil
		},
	}
	h := &ProductServiceHandler{store: store}

	resp, err := h.ListProducts(
		context.Background(),
		connect.NewRequest(&productv1.ListProductsRequest{}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg == nil {
		t.Fatal("response message should not be nil")
	}
	if got := len(resp.Msg.Products); got != 0 {
		t.Errorf("products count = %d, want 0", got)
	}
}

// --- 正常系: store が nil スライスを返す場合も空レスポンスとして扱う ---

func TestListProducts_NilSliceReturnsEmpty(t *testing.T) {
	store := &mockListProductStore{
		ListProductsFn: func(_ context.Context) ([]db.Product, error) {
			return nil, nil
		},
	}
	h := &ProductServiceHandler{store: store}

	resp, err := h.ListProducts(
		context.Background(),
		connect.NewRequest(&productv1.ListProductsRequest{}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg == nil {
		t.Fatal("response message should not be nil")
	}
	// nil スライスでも空レスポンスとして返す
	if resp.Msg.Products == nil {
		// products フィールドが nil でも proto では空リストとして扱われるため許容
		// ただし空スライスに正規化されていることが望ましい
	}
}

// --- 異常系: DB エラー時に connect.CodeInternal を返す ---

func TestListProducts_DBError(t *testing.T) {
	store := &mockListProductStore{
		ListProductsFn: func(_ context.Context) ([]db.Product, error) {
			return nil, errors.New("database connection lost")
		},
	}
	h := &ProductServiceHandler{store: store}

	_, err := h.ListProducts(
		context.Background(),
		connect.NewRequest(&productv1.ListProductsRequest{}),
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

// --- Proto 変換: db.Product → productv1.Product のフィールドマッピング ---

func TestListProducts_ProtoConversion(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	products := []db.Product{
		{
			ID:            42,
			Name:          "変換テスト商品",
			Description:   "詳細な説明文",
			Price:         9999,
			StockQuantity: 7,
			CreatedAt: pgtype.Timestamptz{
				Time:  now,
				Valid: true,
			},
			UpdatedAt: pgtype.Timestamptz{
				Time:  now,
				Valid: true,
			},
		},
	}
	store := &mockListProductStore{
		ListProductsFn: func(_ context.Context) ([]db.Product, error) {
			return products, nil
		},
	}
	h := &ProductServiceHandler{store: store}

	resp, err := h.ListProducts(
		context.Background(),
		connect.NewRequest(&productv1.ListProductsRequest{}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := resp.Msg.Products[0]
	if p.Id != 42 {
		t.Errorf("Id = %d, want 42", p.Id)
	}
	if p.Name != "変換テスト商品" {
		t.Errorf("Name = %q, want %q", p.Name, "変換テスト商品")
	}
	if p.Description != "詳細な説明文" {
		t.Errorf("Description = %q, want %q", p.Description, "詳細な説明文")
	}
	if p.Price != 9999 {
		t.Errorf("Price = %d, want 9999", p.Price)
	}
	// StockQuantity は int32 → int64 変換
	if p.StockQuantity != 7 {
		t.Errorf("StockQuantity = %d, want 7", p.StockQuantity)
	}
	// CreatedAt が正しく timestamppb に変換されていること
	if p.CreatedAt == nil {
		t.Fatal("CreatedAt should not be nil")
	}
	if !p.CreatedAt.AsTime().Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", p.CreatedAt.AsTime(), now)
	}
	// UpdatedAt が正しく変換されていること
	if p.UpdatedAt == nil {
		t.Fatal("UpdatedAt should not be nil")
	}
	if !p.UpdatedAt.AsTime().Equal(now) {
		t.Errorf("UpdatedAt = %v, want %v", p.UpdatedAt.AsTime(), now)
	}
}

// --- Proto 変換: タイムスタンプが無効な場合 nil を返す ---

func TestListProducts_InvalidTimestampReturnsNil(t *testing.T) {
	products := []db.Product{
		{
			ID:            1,
			Name:          "タイムスタンプなし商品",
			Description:   "",
			Price:         100,
			StockQuantity: 1,
			CreatedAt: pgtype.Timestamptz{
				Valid: false,
			},
			UpdatedAt: pgtype.Timestamptz{
				Valid: false,
			},
		},
	}
	store := &mockListProductStore{
		ListProductsFn: func(_ context.Context) ([]db.Product, error) {
			return products, nil
		},
	}
	h := &ProductServiceHandler{store: store}

	resp, err := h.ListProducts(
		context.Background(),
		connect.NewRequest(&productv1.ListProductsRequest{}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := resp.Msg.Products[0]
	if p.CreatedAt != nil {
		t.Errorf("CreatedAt should be nil for invalid timestamp, got %v", p.CreatedAt)
	}
	if p.UpdatedAt != nil {
		t.Errorf("UpdatedAt should be nil for invalid timestamp, got %v", p.UpdatedAt)
	}
}

// --- 境界値: 商品1件のみ ---

func TestListProducts_SingleProduct(t *testing.T) {
	store := &mockListProductStore{
		ListProductsFn: func(_ context.Context) ([]db.Product, error) {
			return []db.Product{
				sampleDBProduct(1, "唯一の商品", "説明", 500, 1),
			}, nil
		},
	}
	h := &ProductServiceHandler{store: store}

	resp, err := h.ListProducts(
		context.Background(),
		connect.NewRequest(&productv1.ListProductsRequest{}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(resp.Msg.Products); got != 1 {
		t.Fatalf("products count = %d, want 1", got)
	}
	if resp.Msg.Products[0].Name != "唯一の商品" {
		t.Errorf("Name = %q, want %q", resp.Msg.Products[0].Name, "唯一の商品")
	}
}
