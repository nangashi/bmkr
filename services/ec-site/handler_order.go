package main

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	ecv1 "github.com/nangashi/bmkr/gen/go/ec/v1"
	productv1 "github.com/nangashi/bmkr/gen/go/product/v1"
	"github.com/nangashi/bmkr/gen/go/product/v1/productv1connect"
	"github.com/nangashi/bmkr/lib/go/pgutil"
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
	pool          *pgxpool.Pool
}

// PlaceOrder handles the PlaceOrder RPC.
func (h *OrderServiceHandler) PlaceOrder(
	ctx context.Context,
	req *connect.Request[ecv1.PlaceOrderRequest],
) (*connect.Response[ecv1.PlaceOrderResponse], error) {
	if req.Msg.CustomerId <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}
	customerID := req.Msg.CustomerId

	// カートを取得
	cart, err := h.q.GetCartByCustomerID(ctx, customerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("cart not found"))
		}
		slog.ErrorContext(ctx, "database error", "error", err, "method", "PlaceOrder.GetCartByCustomerID")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	// カートアイテムを取得（空チェック）
	cartItems, err := h.q.ListCartItems(ctx, cart.ID)
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "PlaceOrder.ListCartItems")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}
	if len(cartItems) == 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("cart is empty"))
	}

	// BatchGetProducts で商品情報をスナップショット取得
	productIDs := make([]int64, len(cartItems))
	for i, item := range cartItems {
		productIDs[i] = item.ProductID
	}

	batchResp, err := h.productClient.BatchGetProducts(ctx, connect.NewRequest(&productv1.BatchGetProductsRequest{
		Ids: productIDs,
	}))
	if err != nil {
		slog.ErrorContext(ctx, "product service error", "error", err, "method", "PlaceOrder.BatchGetProducts")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}
	if len(batchResp.Msg.Products) != len(cartItems) {
		slog.ErrorContext(ctx, "product count mismatch", "expected", len(cartItems), "got", len(batchResp.Msg.Products), "method", "PlaceOrder.BatchGetProducts")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	// 商品情報をマップに変換
	productMap := make(map[int64]*productv1.Product, len(batchResp.Msg.Products))
	for _, p := range batchResp.Msg.Products {
		productMap[p.Id] = p
	}

	// AllocateStock で在庫を確保
	allocItems := make([]*productv1.StockAllocationItem, len(cartItems))
	for i, item := range cartItems {
		allocItems[i] = &productv1.StockAllocationItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	_, err = h.productClient.AllocateStock(ctx, connect.NewRequest(&productv1.AllocateStockRequest{
		Items: allocItems,
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeResourceExhausted {
			return nil, connect.NewError(connect.CodeResourceExhausted, errors.New("insufficient stock"))
		}
		slog.ErrorContext(ctx, "product service error", "error", err, "method", "PlaceOrder.AllocateStock")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	// トランザクション: CreateOrder → CreateOrderItem × len(items) → ClearCartItems
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "PlaceOrder.Begin")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txq := db.New(tx)

	// トランザクション内でカートアイテムを再取得（TOCTOU 防止）
	txCartItems, err := txq.ListCartItems(ctx, cart.ID)
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "PlaceOrder.ListCartItems.tx")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	// total_amount を計算
	var totalAmount int64
	for _, item := range txCartItems {
		p, ok := productMap[item.ProductID]
		if !ok {
			slog.ErrorContext(ctx, "product not found in map", "product_id", item.ProductID, "method", "PlaceOrder")
			return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
		}
		totalAmount += p.Price * int64(item.Quantity)
	}

	order, err := txq.CreateOrder(ctx, db.CreateOrderParams{
		CustomerID:  customerID,
		TotalAmount: totalAmount,
		Status:      "placed",
	})
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "PlaceOrder.CreateOrder")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	orderItems := make([]*ecv1.OrderItem, len(txCartItems))
	for i, item := range txCartItems {
		p := productMap[item.ProductID]
		if err := txq.CreateOrderItem(ctx, db.CreateOrderItemParams{
			OrderID:     order.ID,
			ProductID:   item.ProductID,
			ProductName: p.Name,
			Price:       p.Price,
			Quantity:    item.Quantity,
		}); err != nil {
			slog.ErrorContext(ctx, "database error", "error", err, "method", "PlaceOrder.CreateOrderItem")
			return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
		}
		orderItems[i] = &ecv1.OrderItem{
			ProductId:   item.ProductID,
			ProductName: p.Name,
			Price:       p.Price,
			Quantity:    item.Quantity,
		}
	}

	if err := txq.ClearCartItems(ctx, cart.ID); err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "PlaceOrder.ClearCartItems")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	if err := tx.Commit(ctx); err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "PlaceOrder.Commit")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	return connect.NewResponse(&ecv1.PlaceOrderResponse{
		Order: &ecv1.Order{
			Id:          order.ID,
			CustomerId:  order.CustomerID,
			TotalAmount: order.TotalAmount,
			Status:      order.Status,
			CreatedAt:   pgutil.PgTimestampToProto(order.CreatedAt),
			Items:       orderItems,
		},
	}), nil
}

// ListOrders handles the ListOrders RPC.
func (h *OrderServiceHandler) ListOrders(
	ctx context.Context,
	req *connect.Request[ecv1.ListOrdersRequest],
) (*connect.Response[ecv1.ListOrdersResponse], error) {
	if req.Msg.CustomerId <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}
	customerID := req.Msg.CustomerId

	orders, err := h.q.ListOrdersByCustomerID(ctx, customerID)
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "ListOrders.ListOrdersByCustomerID")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	protoOrders := make([]*ecv1.Order, len(orders))
	for i, o := range orders {
		protoOrders[i] = dbOrderToProto(o, nil)
	}

	return connect.NewResponse(&ecv1.ListOrdersResponse{
		Orders: protoOrders,
	}), nil
}

// GetOrder handles the GetOrder RPC.
func (h *OrderServiceHandler) GetOrder(
	ctx context.Context,
	req *connect.Request[ecv1.GetOrderRequest],
) (*connect.Response[ecv1.GetOrderResponse], error) {
	if req.Msg.CustomerId <= 0 || req.Msg.OrderId <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}

	order, err := h.q.GetOrderByIDAndCustomerID(ctx, db.GetOrderByIDAndCustomerIDParams{
		ID:         req.Msg.OrderId,
		CustomerID: req.Msg.CustomerId,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("order not found"))
		}
		slog.ErrorContext(ctx, "database error", "error", err, "method", "GetOrder.GetOrderByIDAndCustomerID")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	orderItems, err := h.q.ListOrderItemsByOrderID(ctx, order.ID)
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "GetOrder.ListOrderItemsByOrderID")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	return connect.NewResponse(&ecv1.GetOrderResponse{
		Order: dbOrderToProto(order, orderItems),
	}), nil
}

func dbOrderToProto(o db.Order, items []db.OrderItem) *ecv1.Order {
	protoItems := make([]*ecv1.OrderItem, len(items))
	for i, item := range items {
		protoItems[i] = &ecv1.OrderItem{
			Id:          item.ID,
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		}
	}
	return &ecv1.Order{
		Id:          o.ID,
		CustomerId:  o.CustomerID,
		TotalAmount: o.TotalAmount,
		Status:      o.Status,
		CreatedAt:   pgutil.PgTimestampToProto(o.CreatedAt),
		Items:       protoItems,
	}
}
