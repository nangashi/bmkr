# Phase 5: 商品管理サービスの最小実装 — タスクリスト

## Task 1: ProductService ハンドラの実装（CreateProduct, GetProduct）

### やること

`UnimplementedProductServiceHandler` を置き換える独自の `ProductServiceHandler` を実装し、sqlc 生成コードを使って DB 操作を行う。

### 方針

- `services/product-mgmt/handler.go` に `ProductServiceHandler` 構造体を作成する
  - sqlc 生成の `db.Queries` をフィールドに持つ
  - `productv1connect.ProductServiceHandler` インターフェースを満たす
- `CreateProduct` の実装:
  - リクエストから `db.CreateProductParams` を組み立てる
  - `Queries.CreateProduct` で DB に INSERT する
  - DB から返された `db.Product` を proto の `productv1.Product` に変換してレスポンスを返す
- `GetProduct` の実装:
  - リクエストの `id` を使って `Queries.GetProduct` で DB から SELECT する
  - レコードが見つからない場合は `connect.CodeNotFound` エラーを返す
  - DB から返された `db.Product` を proto の `productv1.Product` に変換してレスポンスを返す
- DB モデル → Proto メッセージの変換ヘルパー関数を作成する
  - `db.Product` → `productv1.Product` の変換
  - `pgtype.Timestamptz` → `timestamppb.Timestamp` の変換
  - 型の違いに注意: `StockQuantity` が sqlc では `int32`、proto では `int64`
- `main.go` を修正する:
  - `db.New(pool)` で `Queries` を初期化
  - `UnimplementedProductServiceHandler` の代わりに新しいハンドラを渡す

### スコープ

- `services/product-mgmt/handler.go` の作成（ハンドラ + 変換ヘルパー）
- `services/product-mgmt/main.go` の修正（ハンドラの差し替え）

### スコープ外

- バリデーション（名前の必須チェック、価格の正値チェック等）
- UpdateProduct, DeleteProduct, ListProducts, AllocateStock
- 管理画面 UI
- 認証ミドルウェア

### 受け入れ条件

- `go build ./...` がエラーなく完了する
- Connect RPC 経由で商品の作成ができる:
  ```
  buf curl --data '{"name":"テスト商品","price":1000,"stock_quantity":10}' \
    http://localhost:8081/product.v1.ProductService/CreateProduct
  ```
  が作成された商品情報を返す
- Connect RPC 経由で商品の取得ができる:
  ```
  buf curl --data '{"id":1}' \
    http://localhost:8081/product.v1.ProductService/GetProduct
  ```
  が商品情報を返す
- 存在しない ID で GetProduct を呼ぶと `not_found` エラーが返る

---

## Task 2: 最終検証 — Phase 5 受け入れ条件の確認

### やること

Phase 5 全体の受け入れ条件を一括で検証し、全てのゴールが達成されていることを確認する。

### 方針

- phase.md に記載された Phase 5 の受け入れ条件を全て実行する
- 前提: `just db-up && just db-migrate` で DB が起動・マイグレーション済みであること
- 前提: `cd services/product-mgmt && go run main.go` でサービスが起動していること
- 失敗する項目があれば Task 1 に戻って修正する
- 全検証が通れば Phase 5 完了とする

### スコープ

- Phase 5 受け入れ条件の一括検証

### スコープ外

- Phase 6 以降の作業

### 受け入れ条件

- `go build ./...` がエラーなく完了する:
  ```
  cd services/product-mgmt && go build ./... && echo "OK"
  ```
- Connect RPC 経由で商品の作成ができる:
  ```
  buf curl --data '{"name":"テスト商品","price":1000,"stock_quantity":10}' \
    http://localhost:8081/product.v1.ProductService/CreateProduct
  ```
  が `CreateProductResponse` を返し、`product.id` が割り振られている
- Connect RPC 経由で商品の取得ができる（上記で作成した ID を使用）:
  ```
  buf curl --data '{"id":1}' \
    http://localhost:8081/product.v1.ProductService/GetProduct
  ```
  が `GetProductResponse` を返し、作成した商品の情報が含まれている
- ヘルスチェックが引き続き応答する:
  ```
  curl http://localhost:8081/health
  ```
  が 200 を返す

---

## タスク依存関係

```
Task 1: ProductService ハンドラ実装
  │
  ▼
Task 2: 最終検証
```

- Task 1 で実装が完了した後、Task 2 で受け入れ条件を検証する
