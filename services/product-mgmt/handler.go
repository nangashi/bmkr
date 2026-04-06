package main

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	productv1 "github.com/nangashi/bmkr/gen/go/product/v1"
	"github.com/nangashi/bmkr/lib/go/pgutil"
	db "github.com/nangashi/bmkr/services/product-mgmt/db/generated"
)

// productStore は ProductServiceHandler が必要とする DB 操作を定義する。
// *db.Queries がこのインターフェースを満たす。
// テスト時にモック実装へ差し替え可能にする。
type productStore interface {
	GetProduct(ctx context.Context, id int64) (db.Product, error)
	CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error)
	ListProducts(ctx context.Context) ([]db.Product, error)
	GetProductsByIDs(ctx context.Context, ids []int64) ([]db.Product, error)
}

// コンパイル時に *db.Queries が productStore を満たすことを保証する。
var _ productStore = (*db.Queries)(nil)

type ProductServiceHandler struct {
	store productStore
	pool  *pgxpool.Pool // AllocateStock のトランザクション処理に使用
}

func (h *ProductServiceHandler) CreateProduct(
	ctx context.Context,
	req *connect.Request[productv1.CreateProductRequest],
) (*connect.Response[productv1.CreateProductResponse], error) {
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}
	if req.Msg.Price < 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}
	if req.Msg.StockQuantity < 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}

	product, err := h.store.CreateProduct(ctx, db.CreateProductParams{
		Name:          req.Msg.Name,
		Description:   req.Msg.Description,
		Price:         req.Msg.Price,
		StockQuantity: int32(req.Msg.StockQuantity),
	})
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "CreateProduct")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	return connect.NewResponse(&productv1.CreateProductResponse{
		Product: dbProductToProto(product),
	}), nil
}

func (h *ProductServiceHandler) GetProduct(
	ctx context.Context,
	req *connect.Request[productv1.GetProductRequest],
) (*connect.Response[productv1.GetProductResponse], error) {
	if req.Msg.Id <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}
	product, err := h.store.GetProduct(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("product not found"))
		}
		slog.ErrorContext(ctx, "database error", "error", err, "method", "GetProduct")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	return connect.NewResponse(&productv1.GetProductResponse{
		Product: dbProductToProto(product),
	}), nil
}

// ListProducts handles the ListProducts RPC.
//
// 動作:
//   - store.ListProducts で全商品を ID 昇順で取得する（ページネーションなし）
//   - 各 db.Product を dbProductToProto で protobuf メッセージに変換する
//   - 商品が0件の場合、空の products スライスを持つレスポンスを返す（エラーにしない）
//
// エラー:
//   - DB エラー時は connect.CodeInternal を返す（エラー詳細はログに出力し、クライアントには汎用メッセージのみ返す）
func (h *ProductServiceHandler) ListProducts(
	ctx context.Context,
	req *connect.Request[productv1.ListProductsRequest],
) (*connect.Response[productv1.ListProductsResponse], error) {
	_ = req

	products, err := h.store.ListProducts(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "ListProducts")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	protoProducts := make([]*productv1.Product, 0, len(products))
	for _, product := range products {
		protoProducts = append(protoProducts, dbProductToProto(product))
	}

	return connect.NewResponse(&productv1.ListProductsResponse{
		Products: protoProducts,
	}), nil
}

// BatchGetProducts handles the BatchGetProducts RPC.
//
// 動作:
//   - ids が空、100 件超、または非正値を含む場合は INVALID_ARGUMENT を返す
//   - ids に重複がある場合は検索前に除去し、各 product は高々 1 回返す
//   - store.GetProductsByIDs で該当商品を一括取得する
//   - 存在しない ID は無視し、見つかった商品のみ返す（部分取得）
//   - 各 db.Product を dbProductToProto で protobuf メッセージに変換する
//
// エラー:
//   - DB エラー時は connect.CodeInternal を返す
func (h *ProductServiceHandler) BatchGetProducts(
	ctx context.Context,
	req *connect.Request[productv1.BatchGetProductsRequest],
) (*connect.Response[productv1.BatchGetProductsResponse], error) {
	ids := req.Msg.GetIds()
	if len(ids) == 0 || len(ids) > 100 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("ids must contain between 1 and 100 elements"))
	}

	uniqueIDs := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("ids must be positive"))
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	products, err := h.store.GetProductsByIDs(ctx, uniqueIDs)
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "BatchGetProducts")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	protoProducts := make([]*productv1.Product, 0, len(products))
	for _, product := range products {
		protoProducts = append(protoProducts, dbProductToProto(product))
	}

	return connect.NewResponse(&productv1.BatchGetProductsResponse{
		Products: protoProducts,
	}), nil
}

// AllocateStock handles the AllocateStock RPC.
//
// wip: 動作:
// wip:   - items のバリデーション（1件以上、各 product_id > 0、各 quantity >= 1）
// wip:   - pool.BeginTx でトランザクションを開始する
// wip:   - 各 item に対して db.AllocateStock（UPDATE stock_quantity - quantity WHERE id = $1 AND stock_quantity >= quantity）を実行
// wip:   - AllocateStock の返り値が 0 の場合、在庫不足として tx.Rollback し RESOURCE_EXHAUSTED を返す
// wip:   - 全件成功した場合 tx.Commit する
// wip:   - DB エラー時は tx.Rollback し INTERNAL を返す
// wip:
// wip: エラー:
// wip:   - items が空 → INVALID_ARGUMENT
// wip:   - product_id <= 0 or quantity < 1 → INVALID_ARGUMENT
// wip:   - 在庫不足（AllocateStock rows == 0）→ RESOURCE_EXHAUSTED、全件ロールバック
// wip:   - product_id が存在しない → stock_quantity >= quantity が常に偽となり rows == 0 → RESOURCE_EXHAUSTED として扱う
// wip:   - DB エラー → INTERNAL（ログ出力）
func (h *ProductServiceHandler) AllocateStock(
	ctx context.Context,
	req *connect.Request[productv1.AllocateStockRequest],
) (*connect.Response[productv1.AllocateStockResponse], error) {
	panic("not implemented")
}

func dbProductToProto(p db.Product) *productv1.Product {
	return &productv1.Product{
		Id:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		StockQuantity: int64(p.StockQuantity),
		CreatedAt:     pgutil.PgTimestampToProto(p.CreatedAt),
		UpdatedAt:     pgutil.PgTimestampToProto(p.UpdatedAt),
	}
}
