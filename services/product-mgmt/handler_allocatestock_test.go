package main

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"

	productv1 "github.com/nangashi/bmkr/gen/go/product/v1"
)

// ---------------------------------------------------------------------------
// Tests — AllocateStock RPC: バリデーション
// ---------------------------------------------------------------------------
// トランザクション内部（在庫デクリメント・DB エラー）は pool を直接使用するため
// モック困難。バリデーション分岐のみ UT でカバーし、トランザクション処理は統合テストで担保する。

func TestAllocateStock_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  *productv1.AllocateStockRequest
	}{
		{
			name: "items が空",
			req: &productv1.AllocateStockRequest{
				Items: []*productv1.StockAllocationItem{},
			},
		},
		{
			name: "items が nil",
			req: &productv1.AllocateStockRequest{
				Items: nil,
			},
		},
		{
			name: "product_id が 0",
			req: &productv1.AllocateStockRequest{
				Items: []*productv1.StockAllocationItem{
					{ProductId: 0, Quantity: 1},
				},
			},
		},
		{
			name: "product_id が負数",
			req: &productv1.AllocateStockRequest{
				Items: []*productv1.StockAllocationItem{
					{ProductId: -1, Quantity: 1},
				},
			},
		},
		{
			name: "quantity が 0",
			req: &productv1.AllocateStockRequest{
				Items: []*productv1.StockAllocationItem{
					{ProductId: 1, Quantity: 0},
				},
			},
		},
		{
			name: "quantity が負数",
			req: &productv1.AllocateStockRequest{
				Items: []*productv1.StockAllocationItem{
					{ProductId: 1, Quantity: -1},
				},
			},
		},
		{
			name: "複数アイテムのうち1件が product_id <= 0",
			req: &productv1.AllocateStockRequest{
				Items: []*productv1.StockAllocationItem{
					{ProductId: 1, Quantity: 2},
					{ProductId: 0, Quantity: 1},
				},
			},
		},
		{
			name: "複数アイテムのうち1件が quantity < 1",
			req: &productv1.AllocateStockRequest{
				Items: []*productv1.StockAllocationItem{
					{ProductId: 1, Quantity: 2},
					{ProductId: 2, Quantity: 0},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// pool は nil でもバリデーションエラーで早期リターンするため問題ない
			h := &ProductServiceHandler{
				store: &mockListProductStore{},
				pool:  nil,
			}

			_, err := h.AllocateStock(
				context.Background(),
				connect.NewRequest(tt.req),
			)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var connectErr *connect.Error
			if !errors.As(err, &connectErr) {
				t.Fatalf("expected *connect.Error, got %T: %v", err, err)
			}
			if connectErr.Code() != connect.CodeInvalidArgument {
				t.Errorf("error code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
			}
		})
	}
}
