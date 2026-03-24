---
id: proto-reject-defer
category: design
scope: [all]
severity: high
detectable_by_linter: false
---

# Proto RPC の拒否ケースと委譲を明示する

## アンチパターン

RPC の proto コメントに正常系の動作だけ書き、「何を拒否するか」「何をあえてやらないか」を書かない。LLM は正常系に偏ったコードを生成しやすく、エラーハンドリングが抜ける。また「ここでやるべきか別の RPC でやるべきか」の判断が曖昧になり、責務が重複・欠落する。

```protobuf
// Bad: 正常系だけ
// AddItem adds a product to the customer's cart.
rpc AddItem(AddItemRequest) returns (AddItemResponse);
```

## 正しいパターン

`Rejects:` （この RPC が明確にエラーを返すケース）と `Defers:` （意識的にやらないと決めたこと）を書く。

```protobuf
// Good: 拒否と委譲が明示されている
// AddItem adds a product to the customer's cart.
//
// Rejects:
//   NOT_FOUND        — product_id が存在しない
//   INVALID_ARGUMENT — quantity が 1 未満
//
// Merges:
//   同一 product_id が既にカートにある場合は quantity を加算する
//
// Defers:
//   在庫上限チェックは行わない（PlaceOrder 時に AllocateStock で実施）
rpc AddItem(AddItemRequest) returns (AddItemResponse);
```

## なぜ必要か

- **Rejects** を書くことで、ハンドラのエラー分岐が網羅的に実装される
- **Defers** を書くことで「ここでやらない」が設計判断として記録され、不要なチェックの混入を防ぐ
- テスト設計時に Rejects がそのまま異常系テストケースの一覧になる

## 書き方のルール

1. Rejects には Connect/gRPC のステータスコード（`NOT_FOUND`, `INVALID_ARGUMENT`, `ALREADY_EXISTS` 等）を使う
2. Defers には「やらない理由」または「どこでやるか」を併記する
3. 境界ケース（同一商品の重複追加等）は Merges や Notes として正常系側に書く

## 具体例

```protobuf
// PlaceOrder creates an order from the customer's cart.
//
// Rejects:
//   FAILED_PRECONDITION — cart is empty
//   RESOURCE_EXHAUSTED  — one or more items have insufficient stock
//                         (detected via ProductService.AllocateStock)
//
// Defers:
//   決済処理は行わない（将来の PaymentService に委譲予定）
//   配送先の検証は行わない
rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
```

```protobuf
// DeleteProduct removes a product from the catalog.
//
// Rejects:
//   NOT_FOUND           — product_id が存在しない
//   FAILED_PRECONDITION — product がアクティブな注文に含まれている
//
// Defers:
//   カート内の該当商品の自動削除は行わない（カート表示時に存在チェック）
rpc DeleteProduct(DeleteProductRequest) returns (DeleteProductResponse);
```
