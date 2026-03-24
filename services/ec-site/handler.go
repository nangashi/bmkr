package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	ecv1 "github.com/nangashi/bmkr/gen/go/ec/v1"
	productv1 "github.com/nangashi/bmkr/gen/go/product/v1"
	"github.com/nangashi/bmkr/gen/go/product/v1/productv1connect"
	db "github.com/nangashi/bmkr/services/ec-site/db/generated"
)

type CartServiceHandler struct {
	queries       *db.Queries
	productClient productv1connect.ProductServiceClient
}

func (h *CartServiceHandler) GetCart(
	ctx context.Context,
	req *connect.Request[ecv1.GetCartRequest],
) (*connect.Response[ecv1.GetCartResponse], error) {
	customerID := req.Msg.CustomerId

	// Get or create cart
	cart, err := h.queries.GetCartByCustomerID(ctx, customerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			cart, err = h.queries.CreateCart(ctx, customerID)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
		} else {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// List cart items
	items, err := h.queries.ListCartItems(ctx, cart.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Inter-service communication: fetch product info for each item (log only)
	for _, item := range items {
		_, err := h.productClient.GetProduct(ctx, connect.NewRequest(&productv1.GetProductRequest{
			Id: item.ProductID,
		}))
		if err != nil {
			slog.WarnContext(ctx, "failed to get product", "product_id", item.ProductID, "error", err)
			continue
		}
	}

	return connect.NewResponse(&ecv1.GetCartResponse{
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
	return productv1connect.NewProductServiceClient(http.DefaultClient, baseURL)
}
