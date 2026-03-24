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
	UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (int64, error)
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
// If the product is already in the cart, the quantity is added (UPSERT).
func (h *CartServiceHandler) AddItem(
	ctx context.Context,
	req *connect.Request[ecv1.AddItemRequest],
) (*connect.Response[ecv1.AddItemResponse], error) {
	if req.Msg.CustomerId <= 0 || req.Msg.ProductId <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}
	if req.Msg.Quantity < 1 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}

	cart, err := h.q.GetCartByCustomerID(ctx, req.Msg.CustomerId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			cart, err = h.q.CreateCart(ctx, req.Msg.CustomerId)
			if err != nil {
				slog.ErrorContext(ctx, "database error", "error", err, "method", "AddItem.CreateCart")
				return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
			}
		} else {
			slog.ErrorContext(ctx, "database error", "error", err, "method", "AddItem.GetCartByCustomerID")
			return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
		}
	}

	_, err = h.productClient.GetProduct(ctx, connect.NewRequest(&productv1.GetProductRequest{
		Id: req.Msg.ProductId,
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("product not found"))
		}
		slog.ErrorContext(ctx, "product service error", "error", err, "method", "AddItem.GetProduct")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	err = h.q.UpsertCartItem(ctx, db.UpsertCartItemParams{
		CartID:    cart.ID,
		ProductID: req.Msg.ProductId,
		Quantity:  req.Msg.Quantity,
	})
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "AddItem.UpsertCartItem")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	items, err := h.q.ListCartItems(ctx, cart.ID)
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "AddItem.ListCartItems")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	return connect.NewResponse(&ecv1.AddItemResponse{
		Cart: dbCartToProto(cart, items),
	}), nil
}

// RemoveItem removes a cart item from the customer's cart.
func (h *CartServiceHandler) RemoveItem(
	ctx context.Context,
	req *connect.Request[ecv1.RemoveItemRequest],
) (*connect.Response[ecv1.RemoveItemResponse], error) {
	if req.Msg.CustomerId <= 0 || req.Msg.ItemId <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}

	cart, err := h.q.GetCartByCustomerID(ctx, req.Msg.CustomerId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("cart not found"))
		}
		slog.ErrorContext(ctx, "database error", "error", err, "method", "RemoveItem.GetCartByCustomerID")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	_, err = h.q.GetCartItem(ctx, db.GetCartItemParams{
		ID:     req.Msg.ItemId,
		CartID: cart.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("item not found"))
		}
		slog.ErrorContext(ctx, "database error", "error", err, "method", "RemoveItem.GetCartItem")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	err = h.q.RemoveCartItem(ctx, db.RemoveCartItemParams{
		ID:     req.Msg.ItemId,
		CartID: cart.ID,
	})
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "RemoveItem.RemoveCartItem")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	items, err := h.q.ListCartItems(ctx, cart.ID)
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "RemoveItem.ListCartItems")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	return connect.NewResponse(&ecv1.RemoveItemResponse{
		Cart: dbCartToProto(cart, items),
	}), nil
}

// UpdateQuantity updates the quantity of a cart item.
// The quantity is replaced (not added) with the new value.
func (h *CartServiceHandler) UpdateQuantity(
	ctx context.Context,
	req *connect.Request[ecv1.UpdateQuantityRequest],
) (*connect.Response[ecv1.UpdateQuantityResponse], error) {
	if req.Msg.CustomerId <= 0 || req.Msg.ItemId <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}
	if req.Msg.Quantity < 1 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid argument"))
	}

	cart, err := h.q.GetCartByCustomerID(ctx, req.Msg.CustomerId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("cart not found"))
		}
		slog.ErrorContext(ctx, "database error", "error", err, "method", "UpdateQuantity.GetCartByCustomerID")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	_, err = h.q.GetCartItem(ctx, db.GetCartItemParams{
		ID:     req.Msg.ItemId,
		CartID: cart.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("item not found"))
		}
		slog.ErrorContext(ctx, "database error", "error", err, "method", "UpdateQuantity.GetCartItem")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	rows, err := h.q.UpdateCartItemQuantity(ctx, db.UpdateCartItemQuantityParams{
		Quantity: req.Msg.Quantity,
		ID:       req.Msg.ItemId,
		CartID:   cart.ID,
	})
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "UpdateQuantity.UpdateCartItemQuantity")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}
	if rows == 0 {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("item not found"))
	}

	items, err := h.q.ListCartItems(ctx, cart.ID)
	if err != nil {
		slog.ErrorContext(ctx, "database error", "error", err, "method", "UpdateQuantity.ListCartItems")
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	return connect.NewResponse(&ecv1.UpdateQuantityResponse{
		Cart: dbCartToProto(cart, items),
	}), nil
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
