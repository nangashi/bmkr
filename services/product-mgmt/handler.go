package main

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	productv1 "github.com/nangashi/bmkr/gen/go/product/v1"
	db "github.com/nangashi/bmkr/services/product-mgmt/db/generated"
)

type ProductServiceHandler struct {
	queries *db.Queries
}

func (h *ProductServiceHandler) CreateProduct(
	ctx context.Context,
	req *connect.Request[productv1.CreateProductRequest],
) (*connect.Response[productv1.CreateProductResponse], error) {
	product, err := h.queries.CreateProduct(ctx, db.CreateProductParams{
		Name:          req.Msg.Name,
		Description:   req.Msg.Description,
		Price:         req.Msg.Price,
		StockQuantity: int32(req.Msg.StockQuantity),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&productv1.CreateProductResponse{
		Product: dbProductToProto(product),
	}), nil
}

func (h *ProductServiceHandler) GetProduct(
	ctx context.Context,
	req *connect.Request[productv1.GetProductRequest],
) (*connect.Response[productv1.GetProductResponse], error) {
	product, err := h.queries.GetProduct(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("product not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&productv1.GetProductResponse{
		Product: dbProductToProto(product),
	}), nil
}

func dbProductToProto(p db.Product) *productv1.Product {
	return &productv1.Product{
		Id:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		StockQuantity: int64(p.StockQuantity),
		CreatedAt:     pgTimestampToProto(p.CreatedAt),
		UpdatedAt:     pgTimestampToProto(p.UpdatedAt),
	}
}

func pgTimestampToProto(ts pgtype.Timestamptz) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}
	return timestamppb.New(ts.Time)
}
