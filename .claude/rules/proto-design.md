---
globs: proto/**/*.proto
---

# Proto RPC の設計規約

## 1. 拒否ケースと委譲を明示する

### アンチパターン

RPC の proto コメントに正常系の動作だけ書き、「何を拒否するか」「何をあえてやらないか」を書かない。LLM は正常系に偏ったコードを生成しやすく、エラーハンドリングが抜ける。また「ここでやるべきか別の RPC でやるべきか」の判断が曖昧になり、責務が重複・欠落する。

### 正しいパターン

`Rejects:` （この RPC が明確にエラーを返すケース）と `Defers:` （意識的にやらないと決めたこと）を書く。

```protobuf
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

### 書き方のルール

1. Rejects には Connect/gRPC のステータスコード（`NOT_FOUND`, `INVALID_ARGUMENT`, `ALREADY_EXISTS` 等）を使う
2. Defers には「やらない理由」または「どこでやるか」を併記する
3. 境界ケース（同一商品の重複追加等）は Merges や Notes として正常系側に書く

## 2. 事前条件・事後条件を書く

### アンチパターン

RPC の proto コメントに「何をするか」だけ書き、呼び出し側が満たすべき前提や、完了後に保証される状態を書かない。LLM がハンドラを生成するとき、バリデーションや副作用の実装が抜ける原因になる。

### 正しいパターン

`Pre:` （呼び出し前に満たすべき条件）と `Post:` （呼び出し後に保証される状態変化）を明示する。

```protobuf
// PlaceOrder creates an order from the customer's cart.
//
// Pre:
//   - cart must have at least one item
//   - all items must have sufficient stock (checked via ProductService.AllocateStock)
//
// Post:
//   - order is created with status PLACED
//   - stock is decremented for each item
//   - cart is cleared
//
// Rejects:
//   FAILED_PRECONDITION — cart is empty
//   RESOURCE_EXHAUSTED  — one or more items have insufficient stock
//
// Defers:
//   決済処理は行わない（将来の PaymentService に委譲予定）
rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
```

### なぜ必要か

- **Pre** を書くことで、ハンドラ冒頭のバリデーションが漏れなく実装される
- **Post** を書くことで、副作用（在庫減算、カートクリア等）の実装忘れを防ぐ
- **Rejects** を書くことで、ハンドラのエラー分岐が網羅的に実装される
- **Defers** を書くことで「ここでやらない」が設計判断として記録され、不要なチェックの混入を防ぐ
- テスト設計時に Rejects + Pre がそのまま異常系テストケースの一覧になる
