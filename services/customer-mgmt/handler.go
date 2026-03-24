package main

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"

	customerv1 "github.com/nangashi/bmkr/gen/go/customer/v1"
	db "github.com/nangashi/bmkr/services/customer-mgmt/db/generated"
)

// customerStore は CustomerServiceHandler が必要とする DB 操作を定義する。
// *db.Queries がこのインターフェースを満たす。
type customerStore interface {
	CreateCustomer(ctx context.Context, arg db.CreateCustomerParams) (db.Customer, error)
	GetCustomer(ctx context.Context, id int64) (db.Customer, error)
}

var _ customerStore = (*db.Queries)(nil)

type CustomerServiceHandler struct {
	store customerStore
}

func (h *CustomerServiceHandler) CreateCustomer(
	ctx context.Context,
	req *connect.Request[customerv1.CreateCustomerRequest],
) (*connect.Response[customerv1.CreateCustomerResponse], error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Msg.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	customer, err := h.store.CreateCustomer(ctx, db.CreateCustomerParams{
		Name:         req.Msg.Name,
		Email:        req.Msg.Email,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&customerv1.CreateCustomerResponse{
		Customer: dbCustomerToProto(customer),
	}), nil
}

func (h *CustomerServiceHandler) GetCustomer(
	ctx context.Context,
	req *connect.Request[customerv1.GetCustomerRequest],
) (*connect.Response[customerv1.GetCustomerResponse], error) {
	customer, err := h.store.GetCustomer(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("customer not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&customerv1.GetCustomerResponse{
		Customer: dbCustomerToProto(customer),
	}), nil
}

func dbCustomerToProto(c db.Customer) *customerv1.Customer {
	return &customerv1.Customer{
		Id:        c.ID,
		Name:      c.Name,
		Email:     c.Email,
		CreatedAt: pgTimestampToProto(c.CreatedAt),
		UpdatedAt: pgTimestampToProto(c.UpdatedAt),
	}
}

func pgTimestampToProto(ts pgtype.Timestamptz) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}
	return timestamppb.New(ts.Time)
}
