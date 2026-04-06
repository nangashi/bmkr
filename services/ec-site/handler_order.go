package main

import (
	"context"

	"connectrpc.com/connect"

	ecv1 "github.com/nangashi/bmkr/gen/go/ec/v1"
	"github.com/nangashi/bmkr/gen/go/product/v1/productv1connect"
	db "github.com/nangashi/bmkr/services/ec-site/db/generated"
)

// orderQuerier は OrderServiceHandler が必要とする DB 操作のインターフェース。
// *db.Queries がこのインターフェースを満たす。
// テストではモック実装を注入する（handler-testability.md 準拠）。
type orderQuerier interface {
	// Cart 操作
	GetCartByCustomerID(ctx context.Context, customerID int64) (db.Cart, error)
	ListCartItems(ctx context.Context, cartID int64) ([]db.CartItem, error)
	ClearCartItems(ctx context.Context, cartID int64) error

	// Order 操作
	CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error)
	CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) error
	ListOrdersByCustomerID(ctx context.Context, customerID int64) ([]db.Order, error)
	GetOrderByIDAndCustomerID(ctx context.Context, arg db.GetOrderByIDAndCustomerIDParams) (db.Order, error)
	ListOrderItemsByOrderID(ctx context.Context, orderID int64) ([]db.OrderItem, error)
}

// コンパイル時に *db.Queries が orderQuerier を満たすことを保証する。
var _ orderQuerier = (*db.Queries)(nil)

// OrderServiceHandler は OrderService RPC を実装する。
// PlaceOrder 時に ProductService の AllocateStock を呼び出して在庫を確保する。
type OrderServiceHandler struct {
	q             orderQuerier
	productClient productv1connect.ProductServiceClient
}

// PlaceOrder handles the PlaceOrder RPC.
//
// wip: 動作:
// wip:   - customer_id のバリデーション（> 0）
// wip:   - GetCartByCustomerID でカートを取得。存在しない場合 FAILED_PRECONDITION
// wip:   - ListCartItems でカートアイテムを取得。0件の場合 FAILED_PRECONDITION
// wip:   - ProductService.BatchGetProducts で全商品の price, name を取得（スナップショット用）
// wip:     取得件数がカートアイテム数と一致しない場合 INTERNAL（商品削除等の異常系）
// wip:   - ProductService.AllocateStock で在庫を確保。
// wip:     RESOURCE_EXHAUSTED が返った場合はそのまま RESOURCE_EXHAUSTED を返す
// wip:   - トランザクション: CreateOrder → CreateOrderItem × len(items) → ClearCartItems
// wip:   - レスポンスは Order（items 込み）を返す
// wip:
// wip: エラー:
// wip:   - customer_id <= 0 → INVALID_ARGUMENT
// wip:   - カートなし / 空カート → FAILED_PRECONDITION
// wip:   - BatchGetProducts 取得件数不一致 → INTERNAL（ログ出力）
// wip:   - AllocateStock RESOURCE_EXHAUSTED → RESOURCE_EXHAUSTED をそのまま伝播
// wip:   - ProductService その他エラー → INTERNAL（ログ出力）
// wip:   - DB エラー → INTERNAL（ログ出力）
func (h *OrderServiceHandler) PlaceOrder(
	ctx context.Context,
	req *connect.Request[ecv1.PlaceOrderRequest],
) (*connect.Response[ecv1.PlaceOrderResponse], error) {
	panic("not implemented")
}

// ListOrders handles the ListOrders RPC.
//
// wip: 動作:
// wip:   - customer_id のバリデーション（> 0）
// wip:   - ListOrdersByCustomerID で注文一覧を created_at DESC で取得
// wip:   - 注文が0件の場合、空の orders スライスを持つレスポンスを返す（エラーにしない）
// wip:   - 各注文は items なしのサマリー形式で返す（ListOrderItemsByOrderID は呼ばない）
// wip:
// wip: エラー:
// wip:   - customer_id <= 0 → INVALID_ARGUMENT
// wip:   - DB エラー → INTERNAL（ログ出力）
func (h *OrderServiceHandler) ListOrders(
	ctx context.Context,
	req *connect.Request[ecv1.ListOrdersRequest],
) (*connect.Response[ecv1.ListOrdersResponse], error) {
	panic("not implemented")
}

// GetOrder handles the GetOrder RPC.
//
// wip: 動作:
// wip:   - customer_id, order_id のバリデーション（両方 > 0）
// wip:   - GetOrderByIDAndCustomerID で注文を取得。
// wip:     ErrNoRows の場合（order_id 不存在 or 別顧客の注文）NOT_FOUND を返す
// wip:   - ListOrderItemsByOrderID でアイテムを取得
// wip:   - レスポンスは Order（items 込み）を返す
// wip:
// wip: エラー:
// wip:   - customer_id <= 0 or order_id <= 0 → INVALID_ARGUMENT
// wip:   - ErrNoRows → NOT_FOUND（customer_id が一致しない注文も NOT_FOUND として返す）
// wip:   - DB エラー → INTERNAL（ログ出力）
func (h *OrderServiceHandler) GetOrder(
	ctx context.Context,
	req *connect.Request[ecv1.GetOrderRequest],
) (*connect.Response[ecv1.GetOrderResponse], error) {
	panic("not implemented")
}
