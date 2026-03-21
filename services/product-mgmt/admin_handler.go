package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	db "github.com/nangashi/bmkr/services/product-mgmt/db/generated"
	"github.com/nangashi/bmkr/services/product-mgmt/templates"
)

// AdminProductStore は管理画面の商品一覧で必要なデータアクセスのインターフェース。
// *db.Queries をハンドラに直接持たせず、このインターフェース経由で注入する。
// テスト時にモック実装へ差し替え可能にする。
//
// 動作:
//   - CountProducts: products テーブルの全行数を返す。エラー時は 0 と error を返す
//   - ListProductsPaginated: LIMIT/OFFSET で商品を取得する。
//     結果が0件の場合は空スライスと nil を返す
type AdminProductStore interface {
	CountProducts(ctx context.Context) (int64, error)
	ListProductsPaginated(ctx context.Context, arg db.ListProductsPaginatedParams) ([]db.Product, error)
}

// コンパイル時に *db.Queries が AdminProductStore を満たすことを保証する。
var _ AdminProductStore = (*db.Queries)(nil)

// AdminHandler は管理画面用の Echo ハンドラ群をまとめる構造体。
// store フィールドに AdminProductStore を注入する。
type AdminHandler struct {
	store AdminProductStore
}

// NewAdminHandler は AdminHandler を生成する。
// store が nil の場合の挙動は呼び出し側で保証する。
func NewAdminHandler(store AdminProductStore) *AdminHandler {
	return &AdminHandler{store: store}
}

// HandleProductList は GET /admin/products に対応するハンドラ。
// クエリパラメータ "page" でページ番号を受け取り、商品一覧を返す。
//
// 動作:
//   - page パラメータが未指定または不正な場合、ページ1として扱う
//   - page が 0 以下の場合、ページ1にクランプする
//   - page が総ページ数を超える場合、最終ページにクランプする
//   - 1ページあたり defaultPerPage（20）件を表示する
//   - CountProducts で総件数を取得し、総ページ数を算出する
//   - ListProductsPaginated で該当ページの商品を取得する
//   - db.Product を ProductItem に変換し、CreatedAt は "2006-01-02 15:04" 形式にする
//   - ProductListData を構築する
//   - HX-Request ヘッダの有無で分岐:
//     HTMX リクエスト（HX-Request: true）→ ProductTable パーシャルのみレンダリング
//     通常リクエスト → ProductListPage フルページをレンダリング
//   - Content-Type: "text/html; charset=utf-8" を設定する
//   - 商品が0件の場合、空テーブルと「商品がありません」メッセージを表示する
//   - DB エラー時は HTTP 500 を返す（echo.NewHTTPError(500, err) でエラー情報を保持）
func (h *AdminHandler) HandleProductList(c echo.Context) error {
	page := parseProductListPage(c.QueryParam("page"))
	products, totalCount, totalPages, currentPage, err := h.fetchProducts(c.Request().Context(), page)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	data := ProductListData{
		Products:    products,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		PerPage:     defaultPerPage,
	}

	pageData := toProductListPageData(data)
	component := templates.ProductListPage(pageData)
	if c.Request().Header.Get("HX-Request") == "true" {
		component = templates.ProductTableSection(pageData)
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/html; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *AdminHandler) fetchProducts(ctx context.Context, requestedPage int) ([]ProductItem, int64, int, int, error) {
	currentPage := requestedPage
	totalCount, err := h.store.CountProducts(ctx)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	totalPages := calcTotalPages(totalCount, defaultPerPage)
	if currentPage > totalPages {
		currentPage = totalPages
	}

	products, err := h.store.ListProductsPaginated(ctx, db.ListProductsPaginatedParams{
		Limit:  int32(defaultPerPage),
		Offset: offsetForPage(currentPage),
	})
	if err != nil {
		return nil, 0, 0, 0, err
	}

	return toProductItems(products), totalCount, totalPages, currentPage, nil
}

func parseProductListPage(raw string) int {
	page, err := strconv.Atoi(raw)
	if err != nil || page <= 0 {
		return 1
	}
	return page
}

func calcTotalPages(totalCount int64, perPage int) int {
	if totalCount <= 0 {
		return 1
	}
	return int((totalCount + int64(perPage) - 1) / int64(perPage))
}

func offsetForPage(page int) int32 {
	return int32((page - 1) * defaultPerPage)
}

func toProductItems(products []db.Product) []ProductItem {
	if len(products) == 0 {
		return []ProductItem{}
	}

	items := make([]ProductItem, 0, len(products))
	for _, product := range products {
		items = append(items, ProductItem{
			ID:            product.ID,
			Name:          product.Name,
			Price:         product.Price,
			StockQuantity: product.StockQuantity,
			CreatedAt:     formatProductCreatedAt(product.CreatedAt),
		})
	}
	return items
}

func formatProductCreatedAt(ts pgtype.Timestamptz) string {
	if ts.Valid {
		return ts.Time.Format("2006-01-02 15:04")
	}
	return ""
}

func toProductListPageData(data ProductListData) templates.ProductListPageData {
	products := make([]templates.ProductItemData, 0, len(data.Products))
	for _, product := range data.Products {
		products = append(products, templates.ProductItemData{
			ID:            fmt.Sprintf("%d", product.ID),
			Name:          product.Name,
			Price:         formatIntWithCommas(product.Price),
			StockQuantity: formatIntWithCommas(int64(product.StockQuantity)),
			CreatedAt:     product.CreatedAt,
		})
	}

	return templates.ProductListPageData{
		Products:    products,
		CurrentPage: data.CurrentPage,
		TotalPages:  data.TotalPages,
		TotalCount:  data.TotalCount,
	}
}

func formatIntWithCommas(n int64) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}

	digits := strconv.FormatInt(n, 10)
	if len(digits) <= 3 {
		return sign + digits
	}

	out := make([]byte, 0, len(digits)+(len(digits)-1)/3)
	lead := len(digits) % 3
	if lead == 0 {
		lead = 3
	}
	out = append(out, digits[:lead]...)
	for i := lead; i < len(digits); i += 3 {
		out = append(out, ',')
		out = append(out, digits[i:i+3]...)
	}
	return sign + string(out)
}

var _ templ.Component = templates.ProductListPage(templates.ProductListPageData{})
