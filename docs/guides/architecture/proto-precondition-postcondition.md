---
id: proto-precondition-postcondition
category: design
scope: [all]
severity: high
detectable_by_linter: false
---

# Proto RPC に事前条件・事後条件を書く

## アンチパターン

RPC の proto コメントに「何をするか」だけ書き、呼び出し側が満たすべき前提や、完了後に保証される状態を書かない。LLM がハンドラを生成するとき、バリデーションや副作用の実装が抜ける原因になる。

```protobuf
// Bad: 動作の概要だけ
// PlaceOrder creates an order from the cart.
rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
```

## 正しいパターン

`Pre:` （呼び出し前に満たすべき条件）と `Post:` （呼び出し後に保証される状態変化）を明示する。

```protobuf
// Good: 契約が明示されている
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
rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
```

## なぜ必要か

- Pre を書くことで、ハンドラ冒頭のバリデーションが漏れなく実装される
- Post を書くことで、副作用（在庫減算、カートクリア等）の実装忘れを防ぐ
- 呼び出し側も Pre を見て「何を事前に保証すべきか」を判断できる

## 具体例

```protobuf
// GetCart returns the cart for the given customer.
//
// Pre:
//   - customer_id must be a valid, existing customer
//
// Post:
//   - returns the cart with all items; if no cart exists, returns an empty cart
rpc GetCart(GetCartRequest) returns (GetCartResponse);
```

```protobuf
// CreateProduct registers a new product.
//
// Pre:
//   - name must be non-empty
//   - price must be > 0
//
// Post:
//   - product is persisted with a new unique ID
//   - created_at and updated_at are set to the current time
rpc CreateProduct(CreateProductRequest) returns (CreateProductResponse);
```
