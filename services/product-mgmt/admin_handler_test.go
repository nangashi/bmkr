package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	db "github.com/nangashi/bmkr/services/product-mgmt/db/generated"
)

// ---------------------------------------------------------------------------
// Mock
// ---------------------------------------------------------------------------

// mockProductStore は AdminProductStore のテスト用モック実装。
// CountFn / ListFn を差し替えることで、各テストケースの振る舞いを制御する。
type mockProductStore struct {
	CountFn func(ctx context.Context) (int64, error)
	ListFn  func(ctx context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error)
}

func (m *mockProductStore) CountProducts(ctx context.Context) (int64, error) {
	return m.CountFn(ctx)
}

func (m *mockProductStore) ListProductsPaginated(ctx context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
	return m.ListFn(ctx, arg)
}

// コンパイル時にインターフェース充足を保証する。
var _ AdminProductStore = (*mockProductStore)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newContext は echo.Context を生成するヘルパー。
// target は "/admin/products?page=2" のようなパス+クエリ文字列。
// htmx が true のとき HX-Request ヘッダを付与する。
func newContext(target string, htmx bool) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if htmx {
		req.Header.Set("HX-Request", "true")
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// sampleProduct は CreatedAt に固定タイムスタンプを持つテスト用 db.Product を返す。
func sampleProduct(id int64, name string, price int64, stock int32) db.Product {
	ts := pgtype.Timestamptz{
		Time:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		Valid: true,
	}
	return db.Product{
		ID:            id,
		Name:          name,
		Price:         price,
		StockQuantity: stock,
		CreatedAt:     ts,
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// --- page パラメータのデフォルト値 ---

func TestHandleProductList_PageDefault(t *testing.T) {
	// page パラメータ未指定 → ページ1として扱われる（offset=0 で呼ばれる）
	var capturedArg db.ListProductsPaginatedParams
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) { return 1, nil },
		ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
			capturedArg = arg
			return []db.Product{sampleProduct(1, "商品A", 1000, 10)}, nil
		},
	}
	h := NewAdminHandler(store)
	c, rec := newContext("/admin/products", false)

	err := h.HandleProductList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	// offset=0 はページ1を意味する
	if capturedArg.Offset != 0 {
		t.Errorf("offset = %d, want 0 (page 1)", capturedArg.Offset)
	}
	if capturedArg.Limit != int32(defaultPerPage) {
		t.Errorf("limit = %d, want %d", capturedArg.Limit, defaultPerPage)
	}
}

// --- page 不正値のフォールバック ---

func TestHandleProductList_PageInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"zero", "/admin/products?page=0"},
		{"negative", "/admin/products?page=-5"},
		{"non-numeric", "/admin/products?page=abc"},
		{"empty", "/admin/products?page="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedArg db.ListProductsPaginatedParams
			store := &mockProductStore{
				CountFn: func(_ context.Context) (int64, error) { return 5, nil },
				ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
					capturedArg = arg
					return []db.Product{sampleProduct(1, "商品A", 1000, 10)}, nil
				},
			}
			h := NewAdminHandler(store)
			c, rec := newContext(tt.query, false)

			err := h.HandleProductList(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			// すべてページ1にフォールバックされるため offset=0
			if capturedArg.Offset != 0 {
				t.Errorf("offset = %d, want 0 (page 1 fallback)", capturedArg.Offset)
			}
		})
	}
}

// --- 正常系: 商品一覧が表示される ---

func TestHandleProductList_NormalRendering(t *testing.T) {
	products := []db.Product{
		sampleProduct(1, "テスト商品", 1500, 5),
		sampleProduct(2, "サンプル品", 3000, 20),
	}
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) { return 2, nil },
		ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
			return products, nil
		},
	}
	h := NewAdminHandler(store)
	c, rec := newContext("/admin/products", false)

	err := h.HandleProductList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := rec.Body.String()

	// 商品名が HTML に含まれること
	if !strings.Contains(body, "テスト商品") {
		t.Error("response body should contain product name 'テスト商品'")
	}
	if !strings.Contains(body, "サンプル品") {
		t.Error("response body should contain product name 'サンプル品'")
	}
	// CreatedAt が "2006-01-02 15:04" 形式で表示されること
	if !strings.Contains(body, "2025-01-15 10:30") {
		t.Error("response body should contain formatted date '2025-01-15 10:30'")
	}
	// Content-Type
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

// --- ページネーション: 総ページ数の計算 ---

func TestHandleProductList_TotalPagesCalculation(t *testing.T) {
	tests := []struct {
		name       string
		count      int64
		wantInBody string // "currentPage / totalPages" 形式の部分文字列
	}{
		{"exact_fit", 40, "1 / 2"},      // 40件 / 20件 = 2ページ → 1 / 2
		{"with_remainder", 41, "1 / 3"}, // 41件 / 20件 = 3ページ（切り上げ）→ 1 / 3
		{"one_page", 20, ""},            // ちょうど1ページ → ページネーション表示なし
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockProductStore{
				CountFn: func(_ context.Context) (int64, error) { return tt.count, nil },
				ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
					// 少なくとも1件返す
					return []db.Product{sampleProduct(1, "A", 100, 1)}, nil
				},
			}
			h := NewAdminHandler(store)
			c, rec := newContext("/admin/products", false)

			err := h.HandleProductList(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			body := rec.Body.String()
			if tt.wantInBody != "" && !strings.Contains(body, tt.wantInBody) {
				t.Errorf("body should contain %q for %d items", tt.wantInBody, tt.count)
			}
		})
	}
}

// --- ページネーション: page が総ページ数超過時のクランプと再フェッチ ---

func TestHandleProductList_PageClampToLastPage(t *testing.T) {
	// 総件数25件 → 2ページ。page=100 を指定 → 2にクランプされる
	var capturedArg db.ListProductsPaginatedParams
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) { return 25, nil },
		ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
			capturedArg = arg
			return []db.Product{sampleProduct(1, "末尾商品", 500, 3)}, nil
		},
	}
	h := NewAdminHandler(store)
	c, rec := newContext("/admin/products?page=100", false)

	err := h.HandleProductList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := rec.Body.String()

	// 最終ページ（page=2）にクランプされたので offset = (2-1)*20 = 20
	expectedOffset := int32((2 - 1) * defaultPerPage)
	if capturedArg.Offset != expectedOffset {
		t.Errorf("offset = %d, want %d (clamped to last page)", capturedArg.Offset, expectedOffset)
	}
	// ページ表示が 2 / 2 であること
	if !strings.Contains(body, "2 / 2") {
		t.Error("body should contain '2 / 2' for clamped page")
	}
}

// --- HX-Request ヘッダ分岐: 通常リクエスト → フルページ ---

func TestHandleProductList_FullPage(t *testing.T) {
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) { return 1, nil },
		ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
			return []db.Product{sampleProduct(1, "商品A", 1000, 10)}, nil
		},
	}
	h := NewAdminHandler(store)
	c, rec := newContext("/admin/products", false) // HTMX なし

	err := h.HandleProductList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := rec.Body.String()

	// フルページ → <!doctype html> を含む（templ は小文字の doctype を出力する）
	if !strings.Contains(body, "<!doctype html>") {
		t.Error("full page response should contain '<!doctype html>'")
	}
	// Layout の <header> を含む
	if !strings.Contains(body, "<header>") {
		t.Error("full page response should contain '<header>'")
	}
}

// --- HX-Request ヘッダ分岐: HTMX → ProductTableSection パーシャル ---

func TestHandleProductList_HTMXPartial(t *testing.T) {
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) { return 1, nil },
		ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
			return []db.Product{sampleProduct(1, "商品A", 1000, 10)}, nil
		},
	}
	h := NewAdminHandler(store)
	c, rec := newContext("/admin/products", true) // HTMX リクエスト

	err := h.HandleProductList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := rec.Body.String()

	// パーシャル → <!doctype html> を含まない
	if strings.Contains(body, "<!doctype html>") {
		t.Error("HTMX partial should NOT contain '<!doctype html>'")
	}
	// ProductTableSection の div を含む
	if !strings.Contains(body, `id="product-table-section"`) {
		t.Error("HTMX partial should contain 'id=\"product-table-section\"'")
	}
	// テーブルを含む
	if !strings.Contains(body, "<table>") {
		t.Error("HTMX partial should contain '<table>'")
	}
}

// --- DB エラー時: HTTP 500 ---

func TestHandleProductList_DBErrorOnCount(t *testing.T) {
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) {
			return 0, errors.New("db connection lost")
		},
		ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
			t.Fatal("ListFn should not be called when CountProducts fails")
			return nil, nil
		},
	}
	h := NewAdminHandler(store)
	c, _ := newContext("/admin/products", false)

	err := h.HandleProductList(c)
	// echo.NewHTTPError(500, ...) が返されることを期待
	if err == nil {
		t.Fatal("expected error for DB failure, got nil")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusInternalServerError {
		t.Errorf("HTTP error code = %d, want %d", he.Code, http.StatusInternalServerError)
	}
}

func TestHandleProductList_DBErrorOnList(t *testing.T) {
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) { return 10, nil },
		ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
			return nil, errors.New("query timeout")
		},
	}
	h := NewAdminHandler(store)
	c, _ := newContext("/admin/products", false)

	err := h.HandleProductList(c)
	if err == nil {
		t.Fatal("expected error for DB failure on list, got nil")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusInternalServerError {
		t.Errorf("HTTP error code = %d, want %d", he.Code, http.StatusInternalServerError)
	}
}

// --- 商品0件: 空テーブル ---

func TestHandleProductList_EmptyProducts(t *testing.T) {
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) { return 0, nil },
		ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
			return []db.Product{}, nil
		},
	}
	h := NewAdminHandler(store)
	c, rec := newContext("/admin/products", false)

	err := h.HandleProductList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := rec.Body.String()

	// 「商品がありません」のメッセージが含まれること
	if !strings.Contains(body, "商品がありません") {
		t.Error("empty product list should contain '商品がありません'")
	}
	// ページネーションが表示されないこと（nav 要素がない）
	if strings.Contains(body, `aria-label="ページネーション"`) {
		t.Error("empty product list should NOT show pagination")
	}
}

// ---------------------------------------------------------------------------
// Edge cases (Phase 3b: Codex 提案から採用)
// ---------------------------------------------------------------------------

// --- 0件で page 超過が来た場合のクランプ ---

func TestHandleProductList_EmptyWithLargePage(t *testing.T) {
	var capturedOffset int32
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) { return 0, nil },
		ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
			capturedOffset = arg.Offset
			return []db.Product{}, nil
		},
	}
	h := NewAdminHandler(store)
	c, rec := newContext("/admin/products?page=999", false)

	err := h.HandleProductList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedOffset != 0 {
		t.Errorf("expected offset=0 after clamp, got %d", capturedOffset)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "商品がありません") {
		t.Error("should show empty message")
	}
	if strings.Contains(body, `aria-label="ページネーション"`) {
		t.Error("should NOT show pagination for 0 products")
	}
}

// --- 最終ページちょうどを指定した場合の off-by-one ---

func TestHandleProductList_ExactLastPage(t *testing.T) {
	var capturedOffset int32
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) { return 21, nil },
		ListFn: func(_ context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error) {
			capturedOffset = arg.Offset
			return []db.Product{sampleProduct(21, "Last", 100, 1)}, nil
		},
	}
	h := NewAdminHandler(store)
	c, rec := newContext("/admin/products?page=2", false)

	err := h.HandleProductList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedOffset != int32(defaultPerPage) {
		t.Errorf("expected offset=%d for page 2, got %d", defaultPerPage, capturedOffset)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "2 / 2") {
		t.Error("should show '2 / 2' for last page")
	}
	// 「次へ」は disabled
	if !strings.Contains(body, `aria-disabled="true">次へ`) {
		t.Error("'次へ' should be disabled on last page")
	}
}

// --- HX-Request ヘッダが "true" 以外の値の場合 ---

func TestHandleProductList_HXRequestNonTrue(t *testing.T) {
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) { return 1, nil },
		ListFn: func(_ context.Context, _ db.ListProductsPaginatedParams) ([]db.Product, error) {
			return []db.Product{sampleProduct(1, "Test", 100, 1)}, nil
		},
	}
	h := NewAdminHandler(store)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/products", nil)
	req.Header.Set("HX-Request", "false")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleProductList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := rec.Body.String()
	// フルページとして描画されること
	if !strings.Contains(body, "<!doctype html>") {
		t.Error("HX-Request: false should render full page with <!doctype html>")
	}
}

// --- CountProducts > 0 だが ListProductsPaginated が空を返す（競合状態） ---

func TestHandleProductList_CountPositiveButListEmpty(t *testing.T) {
	store := &mockProductStore{
		CountFn: func(_ context.Context) (int64, error) { return 5, nil },
		ListFn: func(_ context.Context, _ db.ListProductsPaginatedParams) ([]db.Product, error) {
			return nil, nil // 競合状態: 件数取得後にデータが消えた
		},
	}
	h := NewAdminHandler(store)
	c, rec := newContext("/admin/products", false)

	err := h.HandleProductList(c)
	if err != nil {
		t.Fatalf("unexpected error (should not panic): %v", err)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "商品がありません") {
		t.Error("should show empty message when list returns nil despite count > 0")
	}
}
