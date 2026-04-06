package main

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"

	ecv1 "github.com/nangashi/bmkr/gen/go/ec/v1"
	productv1 "github.com/nangashi/bmkr/gen/go/product/v1"
	db "github.com/nangashi/bmkr/services/ec-site/db/generated"
)

// ---------------------------------------------------------------------------
// Mock: orderQuerier
// ---------------------------------------------------------------------------

type mockOrderQuerier struct {
	getCartByCustomerIDFn       func(ctx context.Context, customerID int64) (db.Cart, error)
	listCartItemsFn             func(ctx context.Context, cartID int64) ([]db.CartItem, error)
	clearCartItemsFn            func(ctx context.Context, cartID int64) error
	createOrderFn               func(ctx context.Context, arg db.CreateOrderParams) (db.Order, error)
	createOrderItemFn           func(ctx context.Context, arg db.CreateOrderItemParams) error
	listOrdersByCustomerIDFn    func(ctx context.Context, customerID int64) ([]db.Order, error)
	getOrderByIDAndCustomerIDFn func(ctx context.Context, arg db.GetOrderByIDAndCustomerIDParams) (db.Order, error)
	listOrderItemsByOrderIDFn   func(ctx context.Context, orderID int64) ([]db.OrderItem, error)
}

func (m *mockOrderQuerier) GetCartByCustomerID(ctx context.Context, customerID int64) (db.Cart, error) {
	if m.getCartByCustomerIDFn != nil {
		return m.getCartByCustomerIDFn(ctx, customerID)
	}
	panic("GetCartByCustomerID not expected")
}

func (m *mockOrderQuerier) ListCartItems(ctx context.Context, cartID int64) ([]db.CartItem, error) {
	if m.listCartItemsFn != nil {
		return m.listCartItemsFn(ctx, cartID)
	}
	panic("ListCartItems not expected")
}

func (m *mockOrderQuerier) ClearCartItems(ctx context.Context, cartID int64) error {
	if m.clearCartItemsFn != nil {
		return m.clearCartItemsFn(ctx, cartID)
	}
	panic("ClearCartItems not expected")
}

func (m *mockOrderQuerier) CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
	if m.createOrderFn != nil {
		return m.createOrderFn(ctx, arg)
	}
	panic("CreateOrder not expected")
}

func (m *mockOrderQuerier) CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) error {
	if m.createOrderItemFn != nil {
		return m.createOrderItemFn(ctx, arg)
	}
	panic("CreateOrderItem not expected")
}

func (m *mockOrderQuerier) ListOrdersByCustomerID(ctx context.Context, customerID int64) ([]db.Order, error) {
	if m.listOrdersByCustomerIDFn != nil {
		return m.listOrdersByCustomerIDFn(ctx, customerID)
	}
	panic("ListOrdersByCustomerID not expected")
}

func (m *mockOrderQuerier) GetOrderByIDAndCustomerID(ctx context.Context, arg db.GetOrderByIDAndCustomerIDParams) (db.Order, error) {
	if m.getOrderByIDAndCustomerIDFn != nil {
		return m.getOrderByIDAndCustomerIDFn(ctx, arg)
	}
	panic("GetOrderByIDAndCustomerID not expected")
}

func (m *mockOrderQuerier) ListOrderItemsByOrderID(ctx context.Context, orderID int64) ([]db.OrderItem, error) {
	if m.listOrderItemsByOrderIDFn != nil {
		return m.listOrderItemsByOrderIDFn(ctx, orderID)
	}
	panic("ListOrderItemsByOrderID not expected")
}

// コンパイル時に mockOrderQuerier が orderQuerier を満たすことを保証する。
var _ orderQuerier = (*mockOrderQuerier)(nil)

// ---------------------------------------------------------------------------
// Tests: PlaceOrder
// ---------------------------------------------------------------------------

func TestPlaceOrder_Validation(t *testing.T) {
	tests := []struct {
		name       string
		customerID int64
		wantCode   connect.Code
	}{
		{
			name:       "customer_id が 0",
			customerID: 0,
			wantCode:   connect.CodeInvalidArgument,
		},
		{
			name:       "customer_id が負数",
			customerID: -1,
			wantCode:   connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &OrderServiceHandler{
				q:             &mockOrderQuerier{},
				productClient: &mockProductServiceClient{},
			}

			_, err := handler.PlaceOrder(
				context.Background(),
				connect.NewRequest(&ecv1.PlaceOrderRequest{CustomerId: tt.customerID}),
			)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if got := connect.CodeOf(err); got != tt.wantCode {
				t.Errorf("error code = %v, want %v", got, tt.wantCode)
			}
		})
	}
}

func TestPlaceOrder_CartNotFound(t *testing.T) {
	handler := &OrderServiceHandler{
		q: &mockOrderQuerier{
			getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
				return db.Cart{}, pgx.ErrNoRows
			},
		},
		productClient: &mockProductServiceClient{},
	}

	_, err := handler.PlaceOrder(
		context.Background(),
		connect.NewRequest(&ecv1.PlaceOrderRequest{CustomerId: 100}),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Errorf("error code = %v, want %v", got, connect.CodeFailedPrecondition)
	}
}

func TestPlaceOrder_CartEmpty(t *testing.T) {
	handler := &OrderServiceHandler{
		q: &mockOrderQuerier{
			getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
				return db.Cart{ID: 1, CustomerID: 100}, nil
			},
			listCartItemsFn: func(_ context.Context, _ int64) ([]db.CartItem, error) {
				return []db.CartItem{}, nil
			},
		},
		productClient: &mockProductServiceClient{},
	}

	_, err := handler.PlaceOrder(
		context.Background(),
		connect.NewRequest(&ecv1.PlaceOrderRequest{CustomerId: 100}),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := connect.CodeOf(err); got != connect.CodeFailedPrecondition {
		t.Errorf("error code = %v, want %v", got, connect.CodeFailedPrecondition)
	}
}

func TestPlaceOrder_AllocateStockResourceExhausted(t *testing.T) {
	handler := &OrderServiceHandler{
		q: &mockOrderQuerier{
			getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
				return db.Cart{ID: 1, CustomerID: 100}, nil
			},
			listCartItemsFn: func(_ context.Context, _ int64) ([]db.CartItem, error) {
				return []db.CartItem{
					{ID: 10, CartID: 1, ProductID: 200, Quantity: 2},
				}, nil
			},
		},
		productClient: &mockProductServiceClient{
			BatchGetProductsFn: func(_ context.Context, _ *connect.Request[productv1.BatchGetProductsRequest]) (*connect.Response[productv1.BatchGetProductsResponse], error) {
				return connect.NewResponse(&productv1.BatchGetProductsResponse{
					Products: []*productv1.Product{
						{Id: 200, Name: "商品A", Price: 1000},
					},
				}), nil
			},
			AllocateStockFn: func(_ context.Context, _ *connect.Request[productv1.AllocateStockRequest]) (*connect.Response[productv1.AllocateStockResponse], error) {
				return nil, connect.NewError(connect.CodeResourceExhausted, errors.New("insufficient stock"))
			},
		},
	}

	_, err := handler.PlaceOrder(
		context.Background(),
		connect.NewRequest(&ecv1.PlaceOrderRequest{CustomerId: 100}),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := connect.CodeOf(err); got != connect.CodeResourceExhausted {
		t.Errorf("error code = %v, want %v", got, connect.CodeResourceExhausted)
	}
}

func TestPlaceOrder_AllocateStockOtherError_ReturnsInternal(t *testing.T) {
	handler := &OrderServiceHandler{
		q: &mockOrderQuerier{
			getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
				return db.Cart{ID: 1, CustomerID: 100}, nil
			},
			listCartItemsFn: func(_ context.Context, _ int64) ([]db.CartItem, error) {
				return []db.CartItem{
					{ID: 10, CartID: 1, ProductID: 200, Quantity: 2},
				}, nil
			},
		},
		productClient: &mockProductServiceClient{
			BatchGetProductsFn: func(_ context.Context, _ *connect.Request[productv1.BatchGetProductsRequest]) (*connect.Response[productv1.BatchGetProductsResponse], error) {
				return connect.NewResponse(&productv1.BatchGetProductsResponse{
					Products: []*productv1.Product{
						{Id: 200, Name: "商品A", Price: 1000},
					},
				}), nil
			},
			AllocateStockFn: func(_ context.Context, _ *connect.Request[productv1.AllocateStockRequest]) (*connect.Response[productv1.AllocateStockResponse], error) {
				return nil, connect.NewError(connect.CodeInternal, errors.New("product service unavailable"))
			},
		},
	}

	_, err := handler.PlaceOrder(
		context.Background(),
		connect.NewRequest(&ecv1.PlaceOrderRequest{CustomerId: 100}),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := connect.CodeOf(err); got != connect.CodeInternal {
		t.Errorf("error code = %v, want %v", got, connect.CodeInternal)
	}
}

// ---------------------------------------------------------------------------
// Tests: ListOrders
// ---------------------------------------------------------------------------

func TestListOrders_Validation(t *testing.T) {
	tests := []struct {
		name       string
		customerID int64
	}{
		{name: "customer_id が 0", customerID: 0},
		{name: "customer_id が負数", customerID: -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &OrderServiceHandler{
				q:             &mockOrderQuerier{},
				productClient: &mockProductServiceClient{},
			}

			_, err := handler.ListOrders(
				context.Background(),
				connect.NewRequest(&ecv1.ListOrdersRequest{CustomerId: tt.customerID}),
			)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
				t.Errorf("error code = %v, want %v", got, connect.CodeInvalidArgument)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: GetOrder
// ---------------------------------------------------------------------------

func TestGetOrder_Validation(t *testing.T) {
	tests := []struct {
		name       string
		customerID int64
		orderID    int64
	}{
		{name: "customer_id が 0", customerID: 0, orderID: 1},
		{name: "customer_id が負数", customerID: -1, orderID: 1},
		{name: "order_id が 0", customerID: 100, orderID: 0},
		{name: "order_id が負数", customerID: 100, orderID: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &OrderServiceHandler{
				q:             &mockOrderQuerier{},
				productClient: &mockProductServiceClient{},
			}

			_, err := handler.GetOrder(
				context.Background(),
				connect.NewRequest(&ecv1.GetOrderRequest{
					CustomerId: tt.customerID,
					OrderId:    tt.orderID,
				}),
			)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
				t.Errorf("error code = %v, want %v", got, connect.CodeInvalidArgument)
			}
		})
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	tests := []struct {
		name string
		dbFn func(ctx context.Context, arg db.GetOrderByIDAndCustomerIDParams) (db.Order, error)
	}{
		{
			name: "order_id が存在しない（ErrNoRows）",
			dbFn: func(_ context.Context, _ db.GetOrderByIDAndCustomerIDParams) (db.Order, error) {
				return db.Order{}, pgx.ErrNoRows
			},
		},
		{
			name: "別顧客の注文（customer_id 不一致）も ErrNoRows で NOT_FOUND",
			dbFn: func(_ context.Context, _ db.GetOrderByIDAndCustomerIDParams) (db.Order, error) {
				// GetOrderByIDAndCustomerID は AND 条件なので、別顧客の注文も ErrNoRows になる
				return db.Order{}, pgx.ErrNoRows
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &OrderServiceHandler{
				q: &mockOrderQuerier{
					getOrderByIDAndCustomerIDFn: tt.dbFn,
				},
				productClient: &mockProductServiceClient{},
			}

			_, err := handler.GetOrder(
				context.Background(),
				connect.NewRequest(&ecv1.GetOrderRequest{
					CustomerId: 100,
					OrderId:    999,
				}),
			)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if got := connect.CodeOf(err); got != connect.CodeNotFound {
				t.Errorf("error code = %v, want %v", got, connect.CodeNotFound)
			}
		})
	}
}
