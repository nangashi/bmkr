# Phase 4: 各サービスの最小起動とヘルスチェック — タスクリスト

## Task 1: 商品管理サービスのエントリポイント実装

### やること

商品管理サービスの `main.go` を作成し、Echo サーバー起動・DB 接続・Connect RPC ハンドラマウント・ヘルスチェックを実装する。最初に実装するサービスとして、他の Go サービスの参考パターンとする。

### 方針

- `services/product-mgmt/main.go` にエントリポイントを作成する
- Echo v4 を HTTP フレームワークとして使用する
- `pgx/v5` の `pgxpool` で DB コネクションプールを初期化する
  - 接続先: `postgres://postgres:postgres@localhost:5432/product?sslmode=disable`
  - 環境変数 `DATABASE_URL` でオーバーライド可能にする
- 生成コードの `productv1connect.NewProductServiceHandler` で Connect RPC ハンドラをマウントする
  - `echo.WrapHandler` で Echo に統合する
  - ハンドラは `productv1connect.UnimplementedProductServiceHandler` をそのまま使用する（スタブ）
- `/health` エンドポイントを実装する（DB ping を含む）
- ポートは `:8081`（環境変数 `PORT` でオーバーライド可能）
- 依存パッケージを `go.mod` に追加する:
  - `github.com/labstack/echo/v4`
  - `connectrpc.com/connect`（gen/go 経由で間接的に入るが、直接参照も必要）
  - `golang.org/x/net`（HTTP/2 h2c サポート用）

### スコープ

- `services/product-mgmt/main.go` の作成
- Echo サーバー + DB 接続 + Connect RPC マウント + ヘルスチェック
- `go.mod` への依存追加と `go mod tidy`

### スコープ外

- Connect RPC ハンドラのビジネスロジック実装（Phase 5）
- 管理画面のルーティング
- 認証ミドルウェア

### 受け入れ条件

- サービスが起動できる（Docker Compose で DB 起動済みの前提）:
  ```
  cd services/product-mgmt && go run main.go
  ```
  でサーバーが `:8081` で起動する
- ヘルスチェックが応答する:
  ```
  curl http://localhost:8081/health
  ```
  が 200 を返す
- Connect RPC エンドポイントにリクエストすると `unimplemented` エラーが返る:
  ```
  curl -X POST http://localhost:8081/product.v1.ProductService/GetProduct \
    -H 'Content-Type: application/json' \
    -d '{"id":1}'
  ```
  が Connect の `unimplemented` エラーレスポンスを返す
- `go build ./...` がエラーなく完了する

---

## Task 2: 顧客管理サービスのエントリポイント実装

### やること

顧客管理サービスの `main.go` を作成する。Task 1（商品管理）と同じパターンを適用する。

### 方針

- `services/customer-mgmt/main.go` にエントリポイントを作成する
- Task 1 と同じ構成（Echo + pgxpool + Connect RPC + ヘルスチェック）
  - 接続先: `postgres://postgres:postgres@localhost:5432/customer?sslmode=disable`
  - ポート: `:8082`
- 生成コードの `customerv1connect.NewCustomerServiceHandler` でハンドラをマウント
  - `customerv1connect.UnimplementedCustomerServiceHandler` を使用
- 依存パッケージの追加（echo, connect, x/net）

### スコープ

- `services/customer-mgmt/main.go` の作成
- `go.mod` への依存追加と `go mod tidy`

### スコープ外

- Connect RPC ハンドラのビジネスロジック実装（Phase 6）
- 管理画面のルーティング

### 受け入れ条件

- サービスが起動できる:
  ```
  cd services/customer-mgmt && go run main.go
  ```
  でサーバーが `:8082` で起動する
- ヘルスチェックが応答する:
  ```
  curl http://localhost:8082/health
  ```
  が 200 を返す
- Connect RPC エンドポイントにリクエストすると `unimplemented` エラーが返る:
  ```
  curl -X POST http://localhost:8082/customer.v1.CustomerService/GetCustomer \
    -H 'Content-Type: application/json' \
    -d '{"id":1}'
  ```
- `go build ./...` がエラーなく完了する

---

## Task 3: ECサイトサービスのエントリポイント実装

### やること

ECサイトサービスの `main.go` を作成する。Task 1（商品管理）と同じパターンを適用する。

### 方針

- `services/ec-site/main.go` にエントリポイントを作成する
- Task 1 と同じ構成（Echo + pgxpool + Connect RPC + ヘルスチェック）
  - 接続先: `postgres://postgres:postgres@localhost:5432/ecsite?sslmode=disable`
  - ポート: `:8080`
- 生成コードの `ecv1connect.NewCartServiceHandler` でハンドラをマウント
  - `ecv1connect.UnimplementedCartServiceHandler` を使用
  - `order.connect.go` の `OrderService` はスタブのまま（RPC 定義なしのため `NewOrderServiceHandler` は存在しない）
- 依存パッケージの追加（echo, connect, x/net）

### スコープ

- `services/ec-site/main.go` の作成
- `go.mod` への依存追加と `go mod tidy`

### スコープ外

- Connect RPC ハンドラのビジネスロジック実装（Phase 7）
- サービス間通信の Connect RPC クライアント

### 受け入れ条件

- サービスが起動できる:
  ```
  cd services/ec-site && go run main.go
  ```
  でサーバーが `:8080` で起動する
- ヘルスチェックが応答する:
  ```
  curl http://localhost:8080/health
  ```
  が 200 を返す
- Connect RPC エンドポイントにリクエストすると `unimplemented` エラーが返る:
  ```
  curl -X POST http://localhost:8080/ec.v1.CartService/GetCart \
    -H 'Content-Type: application/json' \
    -d '{"customer_id":1}'
  ```
- `go build ./...` がエラーなく完了する

---

## Task 4: BFF のエントリポイント実装

### やること

BFF（Fastify）の `src/index.ts` を作成し、サーバー起動とヘルスチェックエンドポイントを実装する。

### 方針

- `services/bff/src/index.ts` にエントリポイントを作成する
- Fastify を HTTP フレームワークとして使用する
- `/health` エンドポイントを実装する（DB 接続はないのでサーバー生存確認のみ）
- ポートは `3000`（環境変数 `PORT` でオーバーライド可能）
- `package.json` に以下の依存を追加する:
  - `fastify`
  - `@connectrpc/connect-node`（Connect RPC Node.js トランスポート、Phase 8 で使用するが先にインストール）
- `package.json` に `scripts.dev` を追加する（`npx tsx src/index.ts` 等）
- `tsconfig.json` を必要に応じて調整する

### スコープ

- `services/bff/src/index.ts` の作成
- `package.json` への依存追加と scripts 設定
- Fastify サーバー起動 + ヘルスチェック

### スコープ外

- Connect RPC クライアントによるバックエンド呼び出し（Phase 8）
- 認証ミドルウェア
- CORS / Cookie 設定

### 受け入れ条件

- サービスが起動できる:
  ```
  cd services/bff && pnpm dev
  ```
  でサーバーが `:3000` で起動する
- ヘルスチェックが応答する:
  ```
  curl http://localhost:3000/health
  ```
  が 200 を返す
- `pnpm exec tsc --noEmit` がエラーなく完了する

---

## Task 5: Justfile dev タスクの追加

### やること

Justfile に `just dev` タスクを追加し、4 サービスを並列起動できるようにする。

### 方針

- `just dev` で 4 サービスを並列起動する
- バックグラウンドプロセスとして各サービスを起動し、Ctrl+C で全停止できるようにする
- 方法の候補:
  - シェルのバックグラウンドプロセス + `trap` で一括停止
  - `&` でバックグラウンド起動、`wait` で待機
- 各サービスの起動コマンド:
  - 商品管理: `cd services/product-mgmt && go run main.go`
  - 顧客管理: `cd services/customer-mgmt && go run main.go`
  - ECサイト: `cd services/ec-site && go run main.go`
  - BFF: `cd services/bff && pnpm dev`
- Docker Compose（DB）は `just db-up` で別途起動する前提（`just dev` には含めない）

### スコープ

- Justfile への `dev` タスク追加

### スコープ外

- Docker Compose の自動起動（`just db-up` で手動起動）
- ホットリロードの設定

### 受け入れ条件

- `just --list` で `dev` タスクが表示される
- `just dev` で 4 サービスが同時に起動する
- Ctrl+C で全サービスが停止する
- 起動後、各サービスのヘルスチェックが応答する:
  ```
  curl http://localhost:3000/health
  curl http://localhost:8080/health
  curl http://localhost:8081/health
  curl http://localhost:8082/health
  ```
  全て 200 を返す

---

## Task 6: 最終検証 — Phase 4 受け入れ条件の確認

### やること

Phase 4 全体の受け入れ条件を一括で検証し、全てのゴールが達成されていることを確認する。

### 方針

- phase.md に記載された Phase 4 の受け入れ条件を全て実行する
- 前提: `just db-up && just db-migrate` で DB が起動・マイグレーション済みであること
- 失敗する項目があれば該当タスクに戻って修正する
- 全検証が通れば Phase 4 完了とする

### スコープ

- Phase 4 受け入れ条件の一括検証

### スコープ外

- Phase 5 以降の作業

### 受け入れ条件

- `just dev` で 4 サービスが同時に起動する
- 各サービスのヘルスチェックが応答する:
  ```
  curl http://localhost:3000/health
  curl http://localhost:8080/health
  curl http://localhost:8081/health
  curl http://localhost:8082/health
  ```
  全て 200 を返す
- BFF から商品管理の Connect RPC エンドポイントを叩くと `unimplemented` エラーが返る:
  ```
  curl -X POST http://localhost:8081/product.v1.ProductService/GetProduct \
    -H 'Content-Type: application/json' \
    -d '{"id":1}'
  ```
- 各 Go サービスが DB に接続できている（ヘルスチェックで DB ping が成功）
- 各 Go サービスの `go build ./...` がエラーなく完了する:
  ```
  cd services/product-mgmt && go build ./... && echo "OK"
  cd services/customer-mgmt && go build ./... && echo "OK"
  cd services/ec-site && go build ./... && echo "OK"
  ```
- BFF の TypeScript 型チェックがエラーなく完了する:
  ```
  cd services/bff && pnpm exec tsc --noEmit && echo "OK"
  ```

---

## タスク依存関係

```
Task 1: 商品管理エントリポイント ─┐
Task 2: 顧客管理エントリポイント ─┼─► Task 5: Justfile dev タスク
Task 3: ECサイトエントリポイント ─┤           │
Task 4: BFF エントリポイント ─────┘           │
                                              ▼
                                    Task 6: 最終検証
```

- Task 1 を最初に実装し、Go サービスの参考パターンとする
- Task 2, 3 は Task 1 のパターンを適用するため、Task 1 の後に実施が望ましい
- Task 4 は TypeScript で独立しているため、Task 1 と並行実施可能
- Task 5 は全サービスのエントリポイントが揃った後に実施
- Task 6 は全タスク完了後の最終検証
