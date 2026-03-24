package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	ecv1 "github.com/nangashi/bmkr/gen/go/ec/v1"
	productv1 "github.com/nangashi/bmkr/gen/go/product/v1"
	"github.com/nangashi/bmkr/gen/go/product/v1/productv1connect"
	db "github.com/nangashi/bmkr/services/ec-site/db/generated"
)

// cartQuerier は CartServiceHandler が必要とする DB 操作のインターフェース。
// *db.Queries がこのインターフェースを満たす。
// テストではモック実装を注入する（handler-testability.md 準拠）。
type cartQuerier interface {
	GetCartByCustomerID(ctx context.Context, customerID int64) (db.Cart, error)
	CreateCart(ctx context.Context, customerID int64) (db.Cart, error)
	ListCartItems(ctx context.Context, cartID int64) ([]db.CartItem, error)
	UpsertCartItem(ctx context.Context, arg db.UpsertCartItemParams) error
	RemoveCartItem(ctx context.Context, arg db.RemoveCartItemParams) error
	UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) error
	GetCartItem(ctx context.Context, arg db.GetCartItemParams) (db.CartItem, error)
}

type CartServiceHandler struct {
	q             cartQuerier
	productClient productv1connect.ProductServiceClient
}

func (h *CartServiceHandler) GetCart(
	ctx context.Context,
	req *connect.Request[ecv1.GetCartRequest],
) (*connect.Response[ecv1.GetCartResponse], error) {
	customerID := req.Msg.CustomerId

	// Get or create cart
	cart, err := h.q.GetCartByCustomerID(ctx, customerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			cart, err = h.q.CreateCart(ctx, customerID)
			if err != nil {
				slog.ErrorContext(ctx, "database error", "error", err, "method", "GetCart.CreateCart")
				return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
			}
		} else {
			slog.ErrorContext(ctx, "database error", "error", err, "method", "GetCart.GetCartByCustomerID")
			return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
		}
	}

	// List cart items
	items, err := h.q.ListCartItems(ctx, cart.ID)
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "GetCart.ListCartItems")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	// Fetch product info in batch (log only)
	if len(items) > 0 {
		productIDs := make([]int64, len(items))
		for i, item := range items {
			productIDs[i] = item.ProductID
		}

		_, err := h.productClient.BatchGetProducts(ctx, connect.NewRequest(&productv1.BatchGetProductsRequest{
			Ids: productIDs,
		}))
		if err != nil {
			slog.WarnContext(ctx, "failed to batch get products", "error", err)
		}
	}

	return connect.NewResponse(&ecv1.GetCartResponse{
		Cart: dbCartToProto(cart, items),
	}), nil
}

// AddItem adds a product to the customer's cart.
//
// wip: 正常系フロー
//   - req.Msg.CustomerId <= 0 または req.Msg.ProductId <= 0 の場合、INVALID_ARGUMENT を返す
//   - req.Msg.Quantity が 1 未満の場合、INVALID_ARGUMENT を返す
//   - customer_id でカートを取得する（GetCartByCustomerID）
//   - カートが存在しない場合（pgx.ErrNoRows）、新規作成する（CreateCart）
//   - GetCartByCustomerID / CreateCart の DB エラーは INTERNAL を返す
//   - product_id の存在を ProductService.GetProduct で確認する
//   - GetProduct が NOT_FOUND を返した場合、NOT_FOUND を返す
//   - GetProduct がその他のエラーを返した場合、INTERNAL を返す
//   - UpsertCartItem で cart_items にアイテムを挿入/更新する（同一 product_id は quantity 加算）
//   - UpsertCartItem の DB エラーは INTERNAL を返す
//   - ListCartItems でカート内容を取得し、dbCartToProto で変換してレスポンスを返す
//   - ListCartItems の DB エラーは INTERNAL を返す
//
// wip: エッジケース
//   - quantity == 0 → INVALID_ARGUMENT
//   - quantity が負の値 → INVALID_ARGUMENT
//   - 同一 product_id を2回 AddItem → 2回目は quantity が加算される（UPSERT）
//   - product_id が存在しない → NOT_FOUND（ProductService で確認）
//   - カートが初めて作成される場合 → CreateCart 後に UpsertCartItem
func (h *CartServiceHandler) AddItem(
	ctx context.Context,
	req *connect.Request[ecv1.AddItemRequest],
) (*connect.Response[ecv1.AddItemResponse], error) {
	panic("not implemented")
}

// RemoveItem removes a cart item from the customer's cart.
//
// wip: 正常系フロー
//   - req.Msg.CustomerId <= 0 または req.Msg.ItemId <= 0 の場合、INVALID_ARGUMENT を返す
//   - customer_id でカートを取得する（GetCartByCustomerID）
//   - カートが存在しない場合（pgx.ErrNoRows）、NOT_FOUND を返す
//   - GetCartByCustomerID の DB エラーは INTERNAL を返す
//   - GetCartItem で item_id がカート内に存在するか確認する
//   - GetCartItem が pgx.ErrNoRows を返した場合、NOT_FOUND を返す
//   - GetCartItem の DB エラーは INTERNAL を返す
//   - RemoveCartItem でアイテムを削除する
//   - RemoveCartItem の DB エラーは INTERNAL を返す
//   - ListCartItems でカート内容を取得し、dbCartToProto で変換してレスポンスを返す
//   - ListCartItems の DB エラーは INTERNAL を返す
//
// wip: エッジケース
//   - 存在しない item_id → NOT_FOUND
//   - 存在しない customer_id（カートなし）→ NOT_FOUND
//   - 削除後にカートが空になる場合 → 空のアイテムリストを含むカートを返す（カート自体は削除しない）
func (h *CartServiceHandler) RemoveItem(
	ctx context.Context,
	req *connect.Request[ecv1.RemoveItemRequest],
) (*connect.Response[ecv1.RemoveItemResponse], error) {
	panic("not implemented")
}

// UpdateQuantity updates the quantity of a cart item.
//
// wip: 正常系フロー
//   - req.Msg.CustomerId <= 0 または req.Msg.ItemId <= 0 の場合、INVALID_ARGUMENT を返す
//   - req.Msg.Quantity が 1 未満の場合、INVALID_ARGUMENT を返す
//   - customer_id でカートを取得する（GetCartByCustomerID）
//   - カートが存在しない場合（pgx.ErrNoRows）、NOT_FOUND を返す
//   - GetCartByCustomerID の DB エラーは INTERNAL を返す
//   - GetCartItem で item_id がカート内に存在するか確認する
//   - GetCartItem が pgx.ErrNoRows を返した場合、NOT_FOUND を返す
//   - GetCartItem の DB エラーは INTERNAL を返す
//   - UpdateCartItemQuantity で quantity を置換する（加算ではなく置換）
//   - UpdateCartItemQuantity の DB エラーは INTERNAL を返す
//   - ListCartItems でカート内容を取得し、dbCartToProto で変換してレスポンスを返す
//   - ListCartItems の DB エラーは INTERNAL を返す
//
// wip: エッジケース
//   - quantity == 0 → INVALID_ARGUMENT
//   - quantity が負の値 → INVALID_ARGUMENT
//   - 存在しない item_id → NOT_FOUND
//   - 存在しない customer_id（カートなし）→ NOT_FOUND
func (h *CartServiceHandler) UpdateQuantity(
	ctx context.Context,
	req *connect.Request[ecv1.UpdateQuantityRequest],
) (*connect.Response[ecv1.UpdateQuantityResponse], error) {
	panic("not implemented")
}

func dbCartToProto(c db.Cart, items []db.CartItem) *ecv1.Cart {
	protoItems := make([]*ecv1.CartItem, len(items))
	for i, item := range items {
		protoItems[i] = dbCartItemToProto(item)
	}
	return &ecv1.Cart{
		Id:         c.ID,
		CustomerId: c.CustomerID,
		Items:      protoItems,
		CreatedAt:  pgTimestampToProto(c.CreatedAt),
		UpdatedAt:  pgTimestampToProto(c.UpdatedAt),
	}
}

func dbCartItemToProto(item db.CartItem) *ecv1.CartItem {
	return &ecv1.CartItem{
		Id:        item.ID,
		ProductId: item.ProductID,
		Quantity:  item.Quantity,
		CreatedAt: pgTimestampToProto(item.CreatedAt),
	}
}

func pgTimestampToProto(ts pgtype.Timestamptz) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}
	return timestamppb.New(ts.Time)
}

func newProductServiceClient(baseURL string) productv1connect.ProductServiceClient {
	httpClient := &http.Client{Timeout: 5 * time.Second}
	return productv1connect.NewProductServiceClient(httpClient, baseURL)
}
