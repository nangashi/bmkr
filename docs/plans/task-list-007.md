# Phase 7: ECサイトバックエンドの最小実装 — タスクリスト

## Task 1: CartService ハンドラの実装（GetCart）

### やること

`UnimplementedCartServiceHandler` を置き換える独自の `CartServiceHandler` を実装し、sqlc 生成コードを使って DB 操作を行う。カートが存在しない場合は自動作成する。

### 方針

- `services/ec-site/handler.go` に `CartServiceHandler` 構造体を作成する
  - sqlc 生成の `db.Queries` をフィールドに持つ
  - `ecv1connect.CartServiceHandler` インターフェースを満たす
- `GetCart` の実装:
  - リクエストの `customer_id` を使って `Queries.GetCartByCustomerID` で DB からカートを検索する
  - カートが存在しない場合は `Queries.CreateCart` で空カートを自動作成する
  - `Queries.ListCartItems` でカート内のアイテム一覧を取得する
  - DB から返された `db.Cart` + `[]db.CartItem` を proto の `ecv1.Cart` に変換してレスポンスを返す
- DB モデル → Proto メッセージの変換ヘルパー関数を作成する
  - `db.Cart` + `[]db.CartItem` → `ecv1.Cart` の変換
  - `db.CartItem` → `ecv1.CartItem` の変換
  - `pgtype.Timestamptz` → `timestamppb.Timestamp` の変換
  - 型の違いに注意: `CartItem.Quantity` が sqlc では `int32`、proto でも `int32`
- `main.go` を修正する:
  - `db.New(pool)` で `Queries` を初期化
  - `UnimplementedCartServiceHandler` の代わりに新しいハンドラを渡す

### スコープ

- `services/ec-site/handler.go` の作成（ハンドラ + 変換ヘルパー）
- `services/ec-site/main.go` の修正（ハンドラの差し替え）

### スコープ外

- AddItem, RemoveItem, UpdateQuantity
- OrderService の実装
- バリデーション（customer_id の存在チェック等）

### 受け入れ条件

- `go build ./...` がエラーなく完了する
- Connect RPC 経由でカート取得ができる（カートが自動作成される）:
  ```
  buf curl --data '{"customer_id":1}' \
    http://localhost:8080/ec.v1.CartService/GetCart
  ```
  が空のカート情報を返す（items は空配列）
- 同じ customer_id で再度呼ぶと同じカートが返る（冪等性）

---

## Task 2: ECサイト → 商品管理への Connect RPC クライアント疎通確認

### やること

ECサイトから商品管理サービスへの Connect RPC クライアント呼び出しを実装し、サービス間通信が動作することを確認する。

### 方針

- `services/ec-site/main.go`（または `handler.go`）に商品管理サービスへの Connect RPC クライアントを初期化する
  - `productv1connect.NewProductServiceClient` を使用
  - 接続先: `http://localhost:8081`（環境変数 `PRODUCT_SERVICE_URL` でオーバーライド可能）
- `GetCart` ハンドラ内で商品管理サービスへの疎通確認を行う
  - カート内にアイテムがある場合、各 `product_id` に対して `GetProduct` を呼び出す（ログ出力で疎通確認）
  - カートが空の場合でも、サービス間接続の初期化が成功していることを確認する
- `go.mod` に商品管理の生成コード依存が必要な場合は追加する（`gen/go` の replace ディレクティブ経由で既に利用可能な想定）

### スコープ

- Connect RPC クライアントの初期化と `GetCart` ハンドラからの呼び出し
- サービス間通信の疎通確認（ログ出力）

### スコープ外

- サービス間認証（OAuth 2.0 Client Credentials）
- 商品情報のレスポンスへの統合（カート内アイテムに商品名・価格を付与する等）
- エラーハンドリングの本格実装（商品管理サービスが停止中の場合の処理等）

### 受け入れ条件

- `go build ./...` がエラーなく完了する
- ECサイトと商品管理の両サービスを起動した状態で、GetCart を呼び出すとサービス間通信のログが確認できる
- 商品管理サービスが起動していない場合でも、GetCart 自体はエラーにならない（サービス間通信の失敗はログ出力のみ）

---

## Task 3: 最終検証 — Phase 7 受け入れ条件の確認

### やること

Phase 7 全体の受け入れ条件を一括で検証し、全てのゴールが達成されていることを確認する。

### 方針

- phase.md に記載された Phase 7 の受け入れ条件を全て実行する
- 前提: `just db-up && just db-migrate` で DB が起動・マイグレーション済みであること
- 前提: 商品管理サービス（`:8081`）と ECサイトサービス（`:8080`）の両方が起動していること
- 失敗する項目があれば該当タスクに戻って修正する
- 全検証が通れば Phase 7 完了とする

### スコープ

- Phase 7 受け入れ条件の一括検証

### スコープ外

- Phase 8 以降の作業

### 受け入れ条件

- `go build ./...` がエラーなく完了する:
  ```
  cd services/ec-site && go build ./... && echo "OK"
  ```
- buf curl でカート取得ができる:
  ```
  buf curl --data '{"customer_id":1}' \
    http://localhost:8080/ec.v1.CartService/GetCart
  ```
  がカート情報を返す
- ECサイト → 商品管理への Connect RPC 呼び出しが成功する（ログで確認）
- ヘルスチェックが引き続き応答する:
  ```
  curl http://localhost:8080/health
  ```
  が 200 を返す

---

## タスク依存関係

```
Task 1: CartService ハンドラ実装（GetCart）
  │
  ▼
Task 2: サービス間 Connect RPC クライアント疎通
  │
  ▼
Task 3: 最終検証
```

- Task 1 で GetCart の基本実装を完了した後、Task 2 でサービス間通信を追加する
- Task 2 は Phase 5（商品管理サービス）の完了が前提
- Task 3 は全タスク完了後の最終検証
