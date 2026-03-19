# Phase 6: 顧客管理サービスの最小実装 — タスクリスト

## Task 1: CustomerService ハンドラの実装（CreateCustomer, GetCustomer）

### やること

`UnimplementedCustomerServiceHandler` を置き換える独自の `CustomerServiceHandler` を実装し、sqlc 生成コードを使って DB 操作を行う。

### 方針

- `services/customer-mgmt/handler.go` に `CustomerServiceHandler` 構造体を作成する
  - sqlc 生成の `db.Queries` をフィールドに持つ
  - `customerv1connect.CustomerServiceHandler` インターフェースを満たす
- `CreateCustomer` の実装:
  - リクエストから `db.CreateCustomerParams` を組み立てる
  - `password` フィールドはそのまま `password_hash` に格納する（Phase 6 ではハッシュ化なし、疎通確認優先）
  - `Queries.CreateCustomer` で DB に INSERT する
  - DB から返された `db.Customer` を proto の `customerv1.Customer` に変換してレスポンスを返す
  - proto の `Customer` メッセージには `password` / `password_hash` フィールドがないため、レスポンスにパスワードは含まれない
- `GetCustomer` の実装:
  - リクエストの `id` を使って `Queries.GetCustomer` で DB から SELECT する
  - レコードが見つからない場合は `connect.CodeNotFound` エラーを返す
  - DB から返された `db.Customer` を proto の `customerv1.Customer` に変換してレスポンスを返す
- DB モデル → Proto メッセージの変換ヘルパー関数を作成する
  - `db.Customer` → `customerv1.Customer` の変換
  - `pgtype.Timestamptz` → `timestamppb.Timestamp` の変換
- `main.go` を修正する:
  - `db.New(pool)` で `Queries` を初期化
  - `UnimplementedCustomerServiceHandler` の代わりに新しいハンドラを渡す
- Phase 5（商品管理）の `handler.go` と同じパターンを適用する

### スコープ

- `services/customer-mgmt/handler.go` の作成（ハンドラ + 変換ヘルパー）
- `services/customer-mgmt/main.go` の修正（ハンドラの差し替え）

### スコープ外

- パスワードのハッシュ化（bcrypt 等）
- ListCustomers, GetCustomerByEmail
- 管理画面 UI（templ + HTMX）
- 管理画面の認証（echo-jwt）

### 受け入れ条件

- `go build ./...` がエラーなく完了する
- Connect RPC 経由で顧客の作成ができる:
  ```
  buf curl --data '{"name":"テスト顧客","email":"test@example.com","password":"password"}' \
    http://localhost:8082/customer.v1.CustomerService/CreateCustomer
  ```
  が作成された顧客情報を返す（password は含まれない）
- Connect RPC 経由で顧客の取得ができる:
  ```
  buf curl --data '{"id":1}' \
    http://localhost:8082/customer.v1.CustomerService/GetCustomer
  ```
  が顧客情報を返す
- 存在しない ID で GetCustomer を呼ぶと `not_found` エラーが返る

---

## Task 2: 最終検証 — Phase 6 受け入れ条件の確認

### やること

Phase 6 全体の受け入れ条件を一括で検証し、全てのゴールが達成されていることを確認する。

### 方針

- phase.md に記載された Phase 6 の受け入れ条件を全て実行する
- 前提: `just db-up && just db-migrate` で DB が起動・マイグレーション済みであること
- 前提: `cd services/customer-mgmt && go run main.go` でサービスが起動していること
- 失敗する項目があれば Task 1 に戻って修正する
- 全検証が通れば Phase 6 完了とする

### スコープ

- Phase 6 受け入れ条件の一括検証

### スコープ外

- Phase 7 以降の作業

### 受け入れ条件

- `go build ./...` がエラーなく完了する:
  ```
  cd services/customer-mgmt && go build ./... && echo "OK"
  ```
- Connect RPC 経由で顧客の作成ができる:
  ```
  buf curl --data '{"name":"テスト顧客","email":"test@example.com","password":"password"}' \
    http://localhost:8082/customer.v1.CustomerService/CreateCustomer
  ```
  が `CreateCustomerResponse` を返し、`customer.id` が割り振られている
- Connect RPC 経由で顧客の取得ができる（上記で作成した ID を使用）:
  ```
  buf curl --data '{"id":1}' \
    http://localhost:8082/customer.v1.CustomerService/GetCustomer
  ```
  が `GetCustomerResponse` を返し、作成した顧客の情報が含まれている
- ヘルスチェックが引き続き応答する:
  ```
  curl http://localhost:8082/health
  ```
  が 200 を返す

---

## タスク依存関係

```
Task 1: CustomerService ハンドラ実装
  │
  ▼
Task 2: 最終検証
```

- Task 1 で実装が完了した後、Task 2 で受け入れ条件を検証する
- Phase 5（商品管理）と Phase 6（顧客管理）は並行実施可能
