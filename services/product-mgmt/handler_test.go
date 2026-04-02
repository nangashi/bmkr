package main

import (
	"context"
	"errors"
	"strings"
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
	GetProductFn       func(ctx context.Context, id int64) (db.Product, error)
	CreateProductFn    func(ctx context.Context, arg db.CreateProductParams) (db.Product, error)
	ListProductsFn     func(ctx context.Context) ([]db.Product, error)
	GetProductsByIDsFn func(ctx context.Context, ids []int64) ([]db.Product, error)
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

func (m *mockListProductStore) GetProductsByIDs(ctx context.Context, ids []int64) ([]db.Product, error) {
	if m.GetProductsByIDsFn != nil {
		return m.GetProductsByIDsFn(ctx, ids)
	}
	return nil, errors.New("GetProductsByIDs not implemented")
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
	// nil スライスでも空レスポンスとして返す（nil / 空スライスどちらも許容）
	if got := len(resp.Msg.Products); got != 0 {
		t.Errorf("products count = %d, want 0", got)
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

// ---------------------------------------------------------------------------
// Tests — BatchGetProducts RPC
// ---------------------------------------------------------------------------

func TestBatchGetProducts(t *testing.T) {
	tests := []struct {
		name string
		// request
		ids []int64
		// store behavior
		storeFn func(ctx context.Context, ids []int64) ([]db.Product, error)
		// expectations
		wantCode      connect.Code
		wantErr       bool
		wantCount     int
		wantIDs       []int64 // expected product IDs in response (order-independent)
		checkStoreIDs []int64 // IDs that should be passed to store (after dedup)
	}{
		{
			name: "正常系: 複数IDで複数商品が返る",
			ids:  []int64{1, 2, 3},
			storeFn: func(_ context.Context, ids []int64) ([]db.Product, error) {
				return []db.Product{
					sampleDBProduct(1, "商品A", "説明A", 1000, 10),
					sampleDBProduct(2, "商品B", "説明B", 2500, 5),
					sampleDBProduct(3, "商品C", "説明C", 500, 100),
				}, nil
			},
			wantCount: 3,
			wantIDs:   []int64{1, 2, 3},
		},
		{
			name: "正常系: 単一IDで1商品が返る",
			ids:  []int64{42},
			storeFn: func(_ context.Context, ids []int64) ([]db.Product, error) {
				return []db.Product{
					sampleDBProduct(42, "単一商品", "説明", 999, 1),
				}, nil
			},
			wantCount: 1,
			wantIDs:   []int64{42},
		},
		{
			name: "正常系: 存在しないIDは無視（部分取得）",
			ids:  []int64{1, 999},
			storeFn: func(_ context.Context, ids []int64) ([]db.Product, error) {
				// ID=999 は存在しないので ID=1 のみ返す
				return []db.Product{
					sampleDBProduct(1, "商品A", "説明A", 1000, 10),
				}, nil
			},
			wantCount: 1,
			wantIDs:   []int64{1},
		},
		{
			name: "正常系: 全IDが存在しない場合は空レスポンス",
			ids:  []int64{998, 999},
			storeFn: func(_ context.Context, ids []int64) ([]db.Product, error) {
				return []db.Product{}, nil
			},
			wantCount: 0,
		},
		{
			name: "正常系: 重複IDは除去される",
			ids:  []int64{1, 2, 1, 2, 1},
			storeFn: func(_ context.Context, ids []int64) ([]db.Product, error) {
				return []db.Product{
					sampleDBProduct(1, "商品A", "説明A", 1000, 10),
					sampleDBProduct(2, "商品B", "説明B", 2500, 5),
				}, nil
			},
			checkStoreIDs: []int64{1, 2},
			wantCount:     2,
			wantIDs:       []int64{1, 2},
		},
		{
			name:     "異常系: 空のIDs",
			ids:      []int64{},
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name:     "異常系: nilのIDs（空と同等）",
			ids:      nil,
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name:     "異常系: 101件超のIDs",
			ids:      make101IDs(),
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name:     "異常系: 非正値（0）を含む",
			ids:      []int64{1, 0, 3},
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name:     "異常系: 非正値（負数）を含む",
			ids:      []int64{1, -5, 3},
			wantErr:  true,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name: "異常系: DBエラー時はCodeInternal",
			ids:  []int64{1},
			storeFn: func(_ context.Context, ids []int64) ([]db.Product, error) {
				return nil, errors.New("database connection lost")
			},
			wantErr:  true,
			wantCode: connect.CodeInternal,
		},
		{
			name: "境界値: ちょうど100件のIDs",
			ids:  make100IDs(),
			storeFn: func(_ context.Context, ids []int64) ([]db.Product, error) {
				return []db.Product{sampleDBProduct(1, "商品1", "説明", 100, 1)}, nil
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedIDs []int64
			store := &mockListProductStore{}
			if tt.storeFn != nil {
				store.GetProductsByIDsFn = func(ctx context.Context, ids []int64) ([]db.Product, error) {
					capturedIDs = ids
					return tt.storeFn(ctx, ids)
				}
			}
			h := &ProductServiceHandler{store: store}

			resp, err := h.BatchGetProducts(
				context.Background(),
				connect.NewRequest(&productv1.BatchGetProductsRequest{Ids: tt.ids}),
			)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var connectErr *connect.Error
				if !errors.As(err, &connectErr) {
					t.Fatalf("expected *connect.Error, got %T: %v", err, err)
				}
				if connectErr.Code() != tt.wantCode {
					t.Errorf("error code = %v, want %v", connectErr.Code(), tt.wantCode)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Msg == nil {
				t.Fatal("response message should not be nil")
			}
			if got := len(resp.Msg.Products); got != tt.wantCount {
				t.Fatalf("products count = %d, want %d", got, tt.wantCount)
			}

			// wantIDs が指定されている場合、レスポンスに含まれるIDを検証（順序不問）
			if tt.wantIDs != nil {
				gotIDs := make(map[int64]bool)
				for _, p := range resp.Msg.Products {
					gotIDs[p.Id] = true
				}
				for _, wantID := range tt.wantIDs {
					if !gotIDs[wantID] {
						t.Errorf("expected product ID %d in response, but not found", wantID)
					}
				}
			}

			// checkStoreIDs が指定されている場合、storeに渡されたIDを検証（重複除去確認）
			if tt.checkStoreIDs != nil {
				wantSet := make(map[int64]bool)
				for _, id := range tt.checkStoreIDs {
					wantSet[id] = true
				}
				gotSet := make(map[int64]bool)
				for _, id := range capturedIDs {
					gotSet[id] = true
				}
				if len(gotSet) != len(wantSet) {
					t.Errorf("store received %d unique IDs, want %d", len(gotSet), len(wantSet))
				}
				for id := range wantSet {
					if !gotSet[id] {
						t.Errorf("expected store to receive ID %d, but not found", id)
					}
				}
			}
		})
	}
}

// make100IDs は境界値テスト用に 1..100 の ID スライスを返す。
func make100IDs() []int64 {
	ids := make([]int64, 100)
	for i := range ids {
		ids[i] = int64(i + 1)
	}
	return ids
}

// make101IDs は上限超過テスト用に 1..101 の ID スライスを返す。
func make101IDs() []int64 {
	ids := make([]int64, 101)
	for i := range ids {
		ids[i] = int64(i + 1)
	}
	return ids
}

// ---------------------------------------------------------------------------
// Tests — CreateProduct RPC: バリデーション
// ---------------------------------------------------------------------------

func TestCreateProduct_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  *productv1.CreateProductRequest
	}{
		{
			name: "name が空文字",
			req: &productv1.CreateProductRequest{
				Name:          "",
				Description:   "説明",
				Price:         1000,
				StockQuantity: 10,
			},
		},
		{
			name: "price が負数",
			req: &productv1.CreateProductRequest{
				Name:          "商品A",
				Description:   "説明",
				Price:         -1,
				StockQuantity: 10,
			},
		},
		{
			name: "stock_quantity が負数",
			req: &productv1.CreateProductRequest{
				Name:          "商品A",
				Description:   "説明",
				Price:         1000,
				StockQuantity: -1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockListProductStore{
				CreateProductFn: func(_ context.Context, _ db.CreateProductParams) (db.Product, error) {
					t.Fatal("store.CreateProduct should not be called on validation error")
					return db.Product{}, nil
				},
			}
			h := &ProductServiceHandler{store: store}

			_, err := h.CreateProduct(
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
// Tests — BatchGetProducts RPC: DB エラーメッセージの隠蔽
// ---------------------------------------------------------------------------

func TestBatchGetProducts_DBError_HidesRawMessage(t *testing.T) {
	rawDBError := "pq: connection refused to host 10.0.0.1:5432"
	store := &mockListProductStore{
		GetProductsByIDsFn: func(_ context.Context, _ []int64) ([]db.Product, error) {
			return nil, errors.New(rawDBError)
		},
	}
	h := &ProductServiceHandler{store: store}

	_, err := h.BatchGetProducts(
		context.Background(),
		connect.NewRequest(&productv1.BatchGetProductsRequest{Ids: []int64{1}}),
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
	// エラーメッセージに生の DB エラーが含まれていないこと
	if strings.Contains(connectErr.Message(), rawDBError) {
		t.Errorf("error message should not contain raw DB error, got %q", connectErr.Message())
	}
}
