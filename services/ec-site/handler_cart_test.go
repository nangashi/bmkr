package main

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	ecv1 "github.com/nangashi/bmkr/gen/go/ec/v1"
	productv1 "github.com/nangashi/bmkr/gen/go/product/v1"
	db "github.com/nangashi/bmkr/services/ec-site/db/generated"
)

// ---------------------------------------------------------------------------
// Mock: cartQuerier
// ---------------------------------------------------------------------------

type mockCartQuerier struct {
	getCartByCustomerIDFn    func(ctx context.Context, customerID int64) (db.Cart, error)
	createCartFn             func(ctx context.Context, customerID int64) (db.Cart, error)
	listCartItemsFn          func(ctx context.Context, cartID int64) ([]db.CartItem, error)
	upsertCartItemFn         func(ctx context.Context, arg db.UpsertCartItemParams) error
	removeCartItemFn         func(ctx context.Context, arg db.RemoveCartItemParams) error
	updateCartItemQuantityFn func(ctx context.Context, arg db.UpdateCartItemQuantityParams) (int64, error)
	getCartItemFn            func(ctx context.Context, arg db.GetCartItemParams) (db.CartItem, error)
}

func (m *mockCartQuerier) GetCartByCustomerID(ctx context.Context, customerID int64) (db.Cart, error) {
	if m.getCartByCustomerIDFn != nil {
		return m.getCartByCustomerIDFn(ctx, customerID)
	}
	panic("GetCartByCustomerID not expected")
}

func (m *mockCartQuerier) CreateCart(ctx context.Context, customerID int64) (db.Cart, error) {
	if m.createCartFn != nil {
		return m.createCartFn(ctx, customerID)
	}
	panic("CreateCart not expected")
}

func (m *mockCartQuerier) ListCartItems(ctx context.Context, cartID int64) ([]db.CartItem, error) {
	if m.listCartItemsFn != nil {
		return m.listCartItemsFn(ctx, cartID)
	}
	panic("ListCartItems not expected")
}

func (m *mockCartQuerier) UpsertCartItem(ctx context.Context, arg db.UpsertCartItemParams) error {
	if m.upsertCartItemFn != nil {
		return m.upsertCartItemFn(ctx, arg)
	}
	panic("UpsertCartItem not expected")
}

func (m *mockCartQuerier) RemoveCartItem(ctx context.Context, arg db.RemoveCartItemParams) error {
	if m.removeCartItemFn != nil {
		return m.removeCartItemFn(ctx, arg)
	}
	panic("RemoveCartItem not expected")
}

func (m *mockCartQuerier) UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (int64, error) {
	if m.updateCartItemQuantityFn != nil {
		return m.updateCartItemQuantityFn(ctx, arg)
	}
	panic("UpdateCartItemQuantity not expected")
}

func (m *mockCartQuerier) GetCartItem(ctx context.Context, arg db.GetCartItemParams) (db.CartItem, error) {
	if m.getCartItemFn != nil {
		return m.getCartItemFn(ctx, arg)
	}
	panic("GetCartItem not expected")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

var errDB = errors.New("database error")

func successQuerier() *mockCartQuerier {
	cart := testCart()
	items := testCartItems()
	return &mockCartQuerier{
		getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
			return cart, nil
		},
		listCartItemsFn: func(_ context.Context, _ int64) ([]db.CartItem, error) {
			return items, nil
		},
	}
}

func productFoundClient() *mockProductServiceClient {
	return &mockProductServiceClient{
		GetProductFn: func(_ context.Context, _ *connect.Request[productv1.GetProductRequest]) (*connect.Response[productv1.GetProductResponse], error) {
			return connect.NewResponse(&productv1.GetProductResponse{}), nil
		},
	}
}

// ---------------------------------------------------------------------------
// Tests: GetCart — バリデーション
// ---------------------------------------------------------------------------

func TestGetCart_Validation(t *testing.T) {
	tests := []struct {
		name       string
		customerID int64
	}{
		{
			name:       "customer_id が 0",
			customerID: 0,
		},
		{
			name:       "customer_id が負数",
			customerID: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &CartServiceHandler{
				q:             &mockCartQuerier{},
				productClient: &mockProductServiceClient{},
			}

			_, err := handler.GetCart(
				context.Background(),
				connect.NewRequest(&ecv1.GetCartRequest{CustomerId: tt.customerID}),
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
// Tests: AddItem
// ---------------------------------------------------------------------------

func TestAddItem(t *testing.T) {
	tests := []struct {
		name     string
		req      *ecv1.AddItemRequest
		querier  func() *mockCartQuerier
		client   func() *mockProductServiceClient
		wantCode connect.Code
		wantErr  bool
	}{
		{
			name: "success with existing cart",
			req:  &ecv1.AddItemRequest{CustomerId: 100, ProductId: 200, Quantity: 3},
			querier: func() *mockCartQuerier {
				q := successQuerier()
				q.upsertCartItemFn = func(_ context.Context, _ db.UpsertCartItemParams) error {
					return nil
				}
				return q
			},
			client:  productFoundClient,
			wantErr: false,
		},
		{
			name: "success with new cart creation",
			req:  &ecv1.AddItemRequest{CustomerId: 100, ProductId: 200, Quantity: 1},
			querier: func() *mockCartQuerier {
				cart := testCart()
				items := testCartItems()
				return &mockCartQuerier{
					getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
						return db.Cart{}, pgx.ErrNoRows
					},
					createCartFn: func(_ context.Context, _ int64) (db.Cart, error) {
						return cart, nil
					},
					upsertCartItemFn: func(_ context.Context, _ db.UpsertCartItemParams) error {
						return nil
					},
					listCartItemsFn: func(_ context.Context, _ int64) ([]db.CartItem, error) {
						return items, nil
					},
				}
			},
			client:  productFoundClient,
			wantErr: false,
		},
		{
			name:     "invalid argument: quantity < 1 (zero)",
			req:      &ecv1.AddItemRequest{CustomerId: 100, ProductId: 200, Quantity: 0},
			querier:  func() *mockCartQuerier { return &mockCartQuerier{} },
			client:   func() *mockProductServiceClient { return &mockProductServiceClient{} },
			wantCode: connect.CodeInvalidArgument,
			wantErr:  true,
		},
		{
			name:     "invalid argument: quantity < 1 (negative)",
			req:      &ecv1.AddItemRequest{CustomerId: 100, ProductId: 200, Quantity: -5},
			querier:  func() *mockCartQuerier { return &mockCartQuerier{} },
			client:   func() *mockProductServiceClient { return &mockProductServiceClient{} },
			wantCode: connect.CodeInvalidArgument,
			wantErr:  true,
		},
		{
			name:     "invalid argument: customer_id <= 0",
			req:      &ecv1.AddItemRequest{CustomerId: 0, ProductId: 200, Quantity: 1},
			querier:  func() *mockCartQuerier { return &mockCartQuerier{} },
			client:   func() *mockProductServiceClient { return &mockProductServiceClient{} },
			wantCode: connect.CodeInvalidArgument,
			wantErr:  true,
		},
		{
			name:     "invalid argument: product_id <= 0",
			req:      &ecv1.AddItemRequest{CustomerId: 100, ProductId: 0, Quantity: 1},
			querier:  func() *mockCartQuerier { return &mockCartQuerier{} },
			client:   func() *mockProductServiceClient { return &mockProductServiceClient{} },
			wantCode: connect.CodeInvalidArgument,
			wantErr:  true,
		},
		{
			name: "not found: product does not exist",
			req:  &ecv1.AddItemRequest{CustomerId: 100, ProductId: 999, Quantity: 1},
			querier: func() *mockCartQuerier {
				return successQuerier()
			},
			client: func() *mockProductServiceClient {
				return &mockProductServiceClient{
					GetProductFn: func(_ context.Context, _ *connect.Request[productv1.GetProductRequest]) (*connect.Response[productv1.GetProductResponse], error) {
						return nil, connect.NewError(connect.CodeNotFound, errors.New("product not found"))
					},
				}
			},
			wantCode: connect.CodeNotFound,
			wantErr:  true,
		},
		{
			name: "internal: GetCartByCustomerID DB error",
			req:  &ecv1.AddItemRequest{CustomerId: 100, ProductId: 200, Quantity: 1},
			querier: func() *mockCartQuerier {
				return &mockCartQuerier{
					getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
						return db.Cart{}, errDB
					},
				}
			},
			client:   productFoundClient,
			wantCode: connect.CodeInternal,
			wantErr:  true,
		},
		{
			name: "internal: UpsertCartItem DB error",
			req:  &ecv1.AddItemRequest{CustomerId: 100, ProductId: 200, Quantity: 1},
			querier: func() *mockCartQuerier {
				q := successQuerier()
				q.upsertCartItemFn = func(_ context.Context, _ db.UpsertCartItemParams) error {
					return errDB
				}
				return q
			},
			client:   productFoundClient,
			wantCode: connect.CodeInternal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &CartServiceHandler{
				q:             tt.querier(),
				productClient: tt.client(),
			}

			resp, err := handler.AddItem(context.Background(), connect.NewRequest(tt.req))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if got := connect.CodeOf(err); got != tt.wantCode {
					t.Errorf("error code = %v, want %v", got, tt.wantCode)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp == nil || resp.Msg.Cart == nil {
					t.Fatal("expected non-nil response with cart")
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: RemoveItem
// ---------------------------------------------------------------------------

func TestRemoveItem(t *testing.T) {
	tests := []struct {
		name     string
		req      *ecv1.RemoveItemRequest
		querier  func() *mockCartQuerier
		wantCode connect.Code
		wantErr  bool
	}{
		{
			name: "success: remove item",
			req:  &ecv1.RemoveItemRequest{CustomerId: 100, ItemId: 10},
			querier: func() *mockCartQuerier {
				cart := testCart()
				return &mockCartQuerier{
					getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
						return cart, nil
					},
					getCartItemFn: func(_ context.Context, _ db.GetCartItemParams) (db.CartItem, error) {
						return db.CartItem{ID: 10, CartID: 1, ProductID: 200, Quantity: 2, CreatedAt: pgtype.Timestamptz{Valid: true}}, nil
					},
					removeCartItemFn: func(_ context.Context, _ db.RemoveCartItemParams) error {
						return nil
					},
					listCartItemsFn: func(_ context.Context, _ int64) ([]db.CartItem, error) {
						return []db.CartItem{}, nil // empty after removal
					},
				}
			},
			wantErr: false,
		},
		{
			name: "invalid argument: customer_id <= 0",
			req:  &ecv1.RemoveItemRequest{CustomerId: 0, ItemId: 10},
			querier: func() *mockCartQuerier {
				return &mockCartQuerier{}
			},
			wantCode: connect.CodeInvalidArgument,
			wantErr:  true,
		},
		{
			name: "invalid argument: item_id <= 0",
			req:  &ecv1.RemoveItemRequest{CustomerId: 100, ItemId: 0},
			querier: func() *mockCartQuerier {
				return &mockCartQuerier{}
			},
			wantCode: connect.CodeInvalidArgument,
			wantErr:  true,
		},
		{
			name: "not found: cart does not exist",
			req:  &ecv1.RemoveItemRequest{CustomerId: 100, ItemId: 10},
			querier: func() *mockCartQuerier {
				return &mockCartQuerier{
					getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
						return db.Cart{}, pgx.ErrNoRows
					},
				}
			},
			wantCode: connect.CodeNotFound,
			wantErr:  true,
		},
		{
			name: "not found: item does not exist in cart",
			req:  &ecv1.RemoveItemRequest{CustomerId: 100, ItemId: 999},
			querier: func() *mockCartQuerier {
				cart := testCart()
				return &mockCartQuerier{
					getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
						return cart, nil
					},
					getCartItemFn: func(_ context.Context, _ db.GetCartItemParams) (db.CartItem, error) {
						return db.CartItem{}, pgx.ErrNoRows
					},
				}
			},
			wantCode: connect.CodeNotFound,
			wantErr:  true,
		},
		{
			name: "internal: RemoveCartItem DB error",
			req:  &ecv1.RemoveItemRequest{CustomerId: 100, ItemId: 10},
			querier: func() *mockCartQuerier {
				cart := testCart()
				return &mockCartQuerier{
					getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
						return cart, nil
					},
					getCartItemFn: func(_ context.Context, _ db.GetCartItemParams) (db.CartItem, error) {
						return db.CartItem{ID: 10, CartID: 1, ProductID: 200, Quantity: 2, CreatedAt: pgtype.Timestamptz{Valid: true}}, nil
					},
					removeCartItemFn: func(_ context.Context, _ db.RemoveCartItemParams) error {
						return errDB
					},
				}
			},
			wantCode: connect.CodeInternal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &CartServiceHandler{
				q:             tt.querier(),
				productClient: &mockProductServiceClient{},
			}

			resp, err := handler.RemoveItem(context.Background(), connect.NewRequest(tt.req))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if got := connect.CodeOf(err); got != tt.wantCode {
					t.Errorf("error code = %v, want %v", got, tt.wantCode)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp == nil || resp.Msg.Cart == nil {
					t.Fatal("expected non-nil response with cart")
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: UpdateQuantity
// ---------------------------------------------------------------------------

func TestUpdateQuantity(t *testing.T) {
	tests := []struct {
		name     string
		req      *ecv1.UpdateQuantityRequest
		querier  func() *mockCartQuerier
		wantCode connect.Code
		wantErr  bool
	}{
		{
			name: "success: update quantity",
			req:  &ecv1.UpdateQuantityRequest{CustomerId: 100, ItemId: 10, Quantity: 5},
			querier: func() *mockCartQuerier {
				cart := testCart()
				items := testCartItems()
				return &mockCartQuerier{
					getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
						return cart, nil
					},
					getCartItemFn: func(_ context.Context, _ db.GetCartItemParams) (db.CartItem, error) {
						return db.CartItem{ID: 10, CartID: 1, ProductID: 200, Quantity: 2, CreatedAt: pgtype.Timestamptz{Valid: true}}, nil
					},
					updateCartItemQuantityFn: func(_ context.Context, _ db.UpdateCartItemQuantityParams) (int64, error) {
						return 1, nil
					},
					listCartItemsFn: func(_ context.Context, _ int64) ([]db.CartItem, error) {
						return items, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "invalid argument: quantity < 1 (zero)",
			req:  &ecv1.UpdateQuantityRequest{CustomerId: 100, ItemId: 10, Quantity: 0},
			querier: func() *mockCartQuerier {
				return &mockCartQuerier{}
			},
			wantCode: connect.CodeInvalidArgument,
			wantErr:  true,
		},
		{
			name: "invalid argument: quantity < 1 (negative)",
			req:  &ecv1.UpdateQuantityRequest{CustomerId: 100, ItemId: 10, Quantity: -3},
			querier: func() *mockCartQuerier {
				return &mockCartQuerier{}
			},
			wantCode: connect.CodeInvalidArgument,
			wantErr:  true,
		},
		{
			name: "invalid argument: customer_id <= 0",
			req:  &ecv1.UpdateQuantityRequest{CustomerId: 0, ItemId: 10, Quantity: 5},
			querier: func() *mockCartQuerier {
				return &mockCartQuerier{}
			},
			wantCode: connect.CodeInvalidArgument,
			wantErr:  true,
		},
		{
			name: "not found: cart does not exist",
			req:  &ecv1.UpdateQuantityRequest{CustomerId: 100, ItemId: 10, Quantity: 5},
			querier: func() *mockCartQuerier {
				return &mockCartQuerier{
					getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
						return db.Cart{}, pgx.ErrNoRows
					},
				}
			},
			wantCode: connect.CodeNotFound,
			wantErr:  true,
		},
		{
			name: "not found: item does not exist in cart",
			req:  &ecv1.UpdateQuantityRequest{CustomerId: 100, ItemId: 999, Quantity: 5},
			querier: func() *mockCartQuerier {
				cart := testCart()
				return &mockCartQuerier{
					getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
						return cart, nil
					},
					getCartItemFn: func(_ context.Context, _ db.GetCartItemParams) (db.CartItem, error) {
						return db.CartItem{}, pgx.ErrNoRows
					},
				}
			},
			wantCode: connect.CodeNotFound,
			wantErr:  true,
		},
		{
			name: "internal: UpdateCartItemQuantity DB error",
			req:  &ecv1.UpdateQuantityRequest{CustomerId: 100, ItemId: 10, Quantity: 5},
			querier: func() *mockCartQuerier {
				cart := testCart()
				return &mockCartQuerier{
					getCartByCustomerIDFn: func(_ context.Context, _ int64) (db.Cart, error) {
						return cart, nil
					},
					getCartItemFn: func(_ context.Context, _ db.GetCartItemParams) (db.CartItem, error) {
						return db.CartItem{ID: 10, CartID: 1, ProductID: 200, Quantity: 2, CreatedAt: pgtype.Timestamptz{Valid: true}}, nil
					},
					updateCartItemQuantityFn: func(_ context.Context, _ db.UpdateCartItemQuantityParams) (int64, error) {
						return 0, errDB
					},
				}
			},
			wantCode: connect.CodeInternal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &CartServiceHandler{
				q:             tt.querier(),
				productClient: &mockProductServiceClient{},
			}

			resp, err := handler.UpdateQuantity(context.Background(), connect.NewRequest(tt.req))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if got := connect.CodeOf(err); got != tt.wantCode {
					t.Errorf("error code = %v, want %v", got, tt.wantCode)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp == nil || resp.Msg.Cart == nil {
					t.Fatal("expected non-nil response with cart")
				}
			}
		})
	}
}
