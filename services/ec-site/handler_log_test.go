package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	ecv1 "github.com/nangashi/bmkr/gen/go/ec/v1"
	productv1 "github.com/nangashi/bmkr/gen/go/product/v1"
	"github.com/nangashi/bmkr/gen/go/product/v1/productv1connect"
	db "github.com/nangashi/bmkr/services/ec-site/db/generated"
)

// ---------------------------------------------------------------------------
// Mock: DBTX (pgx の低レベルインターフェース)
// ---------------------------------------------------------------------------

// mockDBTX は db.DBTX のテスト用モック。
// QueryRow / Query の呼び出しを制御し、sqlc 生成コードを通じて
// CartServiceHandler に間接的にテストデータを注入する。
type mockDBTX struct {
	QueryRowFn func(ctx context.Context, sql string, args ...interface{}) pgx.Row
	QueryFn    func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

func (m *mockDBTX) Exec(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag(""), nil
}

func (m *mockDBTX) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return m.QueryFn(ctx, sql, args...)
}

func (m *mockDBTX) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return m.QueryRowFn(ctx, sql, args...)
}

// ---------------------------------------------------------------------------
// Mock: pgx.Row (QueryRow の戻り値)
// ---------------------------------------------------------------------------

// mockRow は pgx.Row のテスト用実装。Scan に返す値を制御する。
type mockRow struct {
	scanFn func(dest ...interface{}) error
}

func (r *mockRow) Scan(dest ...interface{}) error {
	return r.scanFn(dest...)
}

// ---------------------------------------------------------------------------
// Mock: pgx.Rows (Query の戻り値)
// ---------------------------------------------------------------------------

// mockRows は pgx.Rows のテスト用実装。
// items に設定した行データを順番に返す。
type mockRows struct {
	items   []mockRowData
	current int
}

type mockRowData struct {
	scanFn func(dest ...interface{}) error
}

func (r *mockRows) Next() bool {
	if r.current < len(r.items) {
		r.current++
		return true
	}
	return false
}

func (r *mockRows) Scan(dest ...interface{}) error {
	return r.items[r.current-1].scanFn(dest...)
}

func (r *mockRows) Close()                                       {}
func (r *mockRows) Err() error                                   { return nil }
func (r *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.NewCommandTag("") }
func (r *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *mockRows) Values() ([]interface{}, error)               { return nil, nil }
func (r *mockRows) RawValues() [][]byte                          { return nil }
func (r *mockRows) Conn() *pgx.Conn                              { return nil }

// ---------------------------------------------------------------------------
// Mock: ProductServiceClient
// ---------------------------------------------------------------------------

// mockProductServiceClient は productv1connect.ProductServiceClient のテスト用モック。
type mockProductServiceClient struct {
	productv1connect.ProductServiceClient
	GetProductFn       func(ctx context.Context, req *connect.Request[productv1.GetProductRequest]) (*connect.Response[productv1.GetProductResponse], error)
	BatchGetProductsFn func(ctx context.Context, req *connect.Request[productv1.BatchGetProductsRequest]) (*connect.Response[productv1.BatchGetProductsResponse], error)
}

func (m *mockProductServiceClient) GetProduct(ctx context.Context, req *connect.Request[productv1.GetProductRequest]) (*connect.Response[productv1.GetProductResponse], error) {
	return m.GetProductFn(ctx, req)
}

func (m *mockProductServiceClient) BatchGetProducts(ctx context.Context, req *connect.Request[productv1.BatchGetProductsRequest]) (*connect.Response[productv1.BatchGetProductsResponse], error) {
	if m.BatchGetProductsFn != nil {
		return m.BatchGetProductsFn(ctx, req)
	}
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
}

func (m *mockProductServiceClient) CreateProduct(_ context.Context, _ *connect.Request[productv1.CreateProductRequest]) (*connect.Response[productv1.CreateProductResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))
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
// Helpers: DBTX fixture builders
// ---------------------------------------------------------------------------

// cartRow は GetCartByCustomerID / CreateCart 用の mockRow を返す。
// Cart の 4 フィールド (ID, CustomerID, CreatedAt, UpdatedAt) を Scan で設定する。
func cartRow(cart db.Cart) *mockRow {
	return &mockRow{
		scanFn: func(dest ...interface{}) error {
			if len(dest) != 4 {
				return fmt.Errorf("expected 4 scan targets, got %d", len(dest))
			}
			*dest[0].(*int64) = cart.ID
			*dest[1].(*int64) = cart.CustomerID
			*dest[2].(*pgtype.Timestamptz) = cart.CreatedAt
			*dest[3].(*pgtype.Timestamptz) = cart.UpdatedAt
			return nil
		},
	}
}

// noCartRow は GetCartByCustomerID でカートが見つからない場合の mockRow を返す。
func noCartRow() *mockRow {
	return &mockRow{
		scanFn: func(_ ...interface{}) error {
			return pgx.ErrNoRows
		},
	}
}

// cartItemRows は ListCartItems 用の mockRows を返す。
func cartItemRows(items []db.CartItem) *mockRows {
	rows := &mockRows{}
	for _, item := range items {
		item := item // capture
		rows.items = append(rows.items, mockRowData{
			scanFn: func(dest ...interface{}) error {
				if len(dest) != 5 {
					return fmt.Errorf("expected 5 scan targets, got %d", len(dest))
				}
				*dest[0].(*int64) = item.ID
				*dest[1].(*int64) = item.CartID
				*dest[2].(*int64) = item.ProductID
				*dest[3].(*int32) = item.Quantity
				*dest[4].(*pgtype.Timestamptz) = item.CreatedAt
				return nil
			},
		})
	}
	return rows
}

// ---------------------------------------------------------------------------
// Helpers: fixture data
// ---------------------------------------------------------------------------

func testCart() db.Cart {
	return db.Cart{
		ID:         1,
		CustomerID: 100,
		CreatedAt:  pgtype.Timestamptz{Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Valid: true},
	}
}

func testCartItems() []db.CartItem {
	return []db.CartItem{
		{ID: 10, CartID: 1, ProductID: 200, Quantity: 2, CreatedAt: pgtype.Timestamptz{Valid: true}},
		{ID: 11, CartID: 1, ProductID: 300, Quantity: 1, CreatedAt: pgtype.Timestamptz{Valid: true}},
	}
}

// ---------------------------------------------------------------------------
// newTestHandler は DBTX モック経由で CartServiceHandler を構築するヘルパー。
// getCartByCustomerID / createCart は QueryRow、listCartItems は Query に対応する。
// ---------------------------------------------------------------------------

type dbtxScenario struct {
	// QueryRow の呼び出し回数に応じて戻す Row を制御する。
	// getCartByCustomerID → 1回目, createCart → 2回目 (新規作成フローの場合)
	queryRowCalls []*mockRow
	queryRowIdx   int

	// Query (ListCartItems) の結果
	listCartItemsRows *mockRows
	listCartItemsErr  error
}

func (s *dbtxScenario) newDBTX() *mockDBTX {
	return &mockDBTX{
		QueryRowFn: func(_ context.Context, _ string, _ ...interface{}) pgx.Row {
			idx := s.queryRowIdx
			s.queryRowIdx++
			if idx < len(s.queryRowCalls) {
				return s.queryRowCalls[idx]
			}
			return noCartRow()
		},
		QueryFn: func(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
			if s.listCartItemsErr != nil {
				return nil, s.listCartItemsErr
			}
			return s.listCartItemsRows, nil
		},
	}
}

// ---------------------------------------------------------------------------
// Tests: 商品取得失敗時に slog.WarnContext がログ出力される
// ---------------------------------------------------------------------------

func TestGetCart_ProductFetchFailure_LogsWarn(t *testing.T) {
	records := captureSlog(t)

	items := testCartItems() // 2 items: ProductID=200, ProductID=300
	scenario := &dbtxScenario{
		queryRowCalls:     []*mockRow{cartRow(testCart())},
		listCartItemsRows: cartItemRows(items),
	}
	queries := db.New(scenario.newDBTX())

	productErr := errors.New("service unavailable")
	client := &mockProductServiceClient{
		BatchGetProductsFn: func(_ context.Context, _ *connect.Request[productv1.BatchGetProductsRequest]) (*connect.Response[productv1.BatchGetProductsResponse], error) {
			return nil, productErr
		},
	}

	handler := &CartServiceHandler{
		q:             queries,
		productClient: client,
	}

	resp, err := handler.GetCart(context.Background(), connect.NewRequest(&ecv1.GetCartRequest{
		CustomerId: 100,
	}))
	// 商品取得失敗はエラーにならず、レスポンスは返る
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	// slog.WarnContext が呼ばれていること
	if len(*records) == 0 {
		t.Fatal("expected at least one log record, got none")
	}

	// BatchGetProducts 失敗で WARN ログが1回出ること
	warnCount := 0
	for _, rec := range *records {
		if rec.Level == slog.LevelWarn && rec.Message == "failed to batch get products" {
			warnCount++

			// error フィールドが含まれること
			if _, ok := rec.Attrs["error"]; !ok {
				t.Error("WARN log should contain 'error' attribute")
			}
		}
	}

	if warnCount != 1 {
		t.Errorf("expected 1 WARN log for 'failed to batch get products', got %d", warnCount)
	}
}

// ---------------------------------------------------------------------------
// Tests: 商品取得失敗時のフィールド値の検証
// ---------------------------------------------------------------------------

func TestGetCart_ProductFetchFailure_LogFieldValues(t *testing.T) {
	records := captureSlog(t)

	singleItem := []db.CartItem{
		{ID: 10, CartID: 1, ProductID: 42, Quantity: 1, CreatedAt: pgtype.Timestamptz{Valid: true}},
	}
	scenario := &dbtxScenario{
		queryRowCalls:     []*mockRow{cartRow(testCart())},
		listCartItemsRows: cartItemRows(singleItem),
	}
	queries := db.New(scenario.newDBTX())

	productErr := errors.New("connection refused")
	client := &mockProductServiceClient{
		BatchGetProductsFn: func(_ context.Context, _ *connect.Request[productv1.BatchGetProductsRequest]) (*connect.Response[productv1.BatchGetProductsResponse], error) {
			return nil, productErr
		},
	}

	handler := &CartServiceHandler{
		q:             queries,
		productClient: client,
	}

	_, _ = handler.GetCart(context.Background(), connect.NewRequest(&ecv1.GetCartRequest{
		CustomerId: 100,
	}))

	if len(*records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(*records))
	}

	rec := (*records)[0]

	// error の値が productErr であること
	if errVal, ok := rec.Attrs["error"]; !ok {
		t.Error("missing 'error' attribute")
	} else if errVal != productErr {
		t.Errorf("error = %v, want %v", errVal, productErr)
	}
}

// ---------------------------------------------------------------------------
// Tests: 正常系ではログが出力されない（カート新規作成 + 商品取得成功）
// ---------------------------------------------------------------------------

func TestGetCart_NewCartNormalFlow_NoLogs(t *testing.T) {
	records := captureSlog(t)

	items := testCartItems()
	scenario := &dbtxScenario{
		// 1回目: GetCartByCustomerID → ErrNoRows、2回目: CreateCart → 成功
		queryRowCalls:     []*mockRow{noCartRow(), cartRow(testCart())},
		listCartItemsRows: cartItemRows(items),
	}
	queries := db.New(scenario.newDBTX())

	client := &mockProductServiceClient{
		BatchGetProductsFn: func(_ context.Context, _ *connect.Request[productv1.BatchGetProductsRequest]) (*connect.Response[productv1.BatchGetProductsResponse], error) {
			return connect.NewResponse(&productv1.BatchGetProductsResponse{}), nil
		},
	}

	handler := &CartServiceHandler{
		q:             queries,
		productClient: client,
	}

	resp, err := handler.GetCart(context.Background(), connect.NewRequest(&ecv1.GetCartRequest{
		CustomerId: 100,
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	// カート作成成功・商品取得成功いずれでもログが出力されないこと
	if len(*records) != 0 {
		t.Errorf("expected no log records in normal flow, got %d:", len(*records))
		for i, rec := range *records {
			t.Errorf("  [%d] level=%s msg=%q attrs=%v", i, rec.Level, rec.Message, rec.Attrs)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests: 既存カートが見つかった場合も正常系ログなし
// ---------------------------------------------------------------------------

func TestGetCart_ExistingCart_NoLogs(t *testing.T) {
	records := captureSlog(t)

	items := testCartItems()
	scenario := &dbtxScenario{
		queryRowCalls:     []*mockRow{cartRow(testCart())},
		listCartItemsRows: cartItemRows(items),
	}
	queries := db.New(scenario.newDBTX())

	client := &mockProductServiceClient{
		BatchGetProductsFn: func(_ context.Context, _ *connect.Request[productv1.BatchGetProductsRequest]) (*connect.Response[productv1.BatchGetProductsResponse], error) {
			return connect.NewResponse(&productv1.BatchGetProductsResponse{}), nil
		},
	}

	handler := &CartServiceHandler{
		q:             queries,
		productClient: client,
	}

	_, err := handler.GetCart(context.Background(), connect.NewRequest(&ecv1.GetCartRequest{
		CustomerId: 100,
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 正常系ではログが出力されないこと
	if len(*records) != 0 {
		t.Errorf("expected no log records for existing cart flow, got %d:", len(*records))
		for i, rec := range *records {
			t.Errorf("  [%d] level=%s msg=%q attrs=%v", i, rec.Level, rec.Message, rec.Attrs)
		}
	}
}
