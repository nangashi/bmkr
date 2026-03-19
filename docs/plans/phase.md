# 開発フェーズ計画

## Phase 1: 開発環境とモノレポ基盤

### やること

開発ツールチェーンとモノレポの骨格を構築し、空のサービスディレクトリが正しく認識される状態にする。

### やることの詳細リスト

- mise による Node.js / Go / pnpm / buf / sqlc / goose のバージョン管理設定
- pnpm workspace の初期化（`pnpm-workspace.yaml`、ルート `package.json`）
- `services/bff/`、`services/ec-site/`、`services/product-mgmt/`、`services/customer-mgmt/` のディレクトリ作成
- BFF の `package.json` 作成（依存は空でよい）
- 各 Go サービスの `go.mod` 初期化
- `proto/` ディレクトリと `buf.yaml` / `buf.gen.yaml` の設定
- ルート `.gitignore` の更新（`node_modules/`、Go バイナリ、`gen/` 等）
- lefthook の導入（初期フックは lint のみ、対象がまだないのでスタブ）
- Just（Justfile）の導入（`just setup` で mise install + pnpm install を実行）

### スコープ

- ツールのバージョン固定と再現可能なインストール
- モノレポのディレクトリ構成と workspace 認識
- buf CLI による Protobuf ツールチェーンの設定
- lefthook / Just の設定ファイル配置

### スコープ外

- 各サービスのアプリケーションコード（main.go / index.ts 等）
- Docker / Docker Compose
- Protobuf のスキーマ定義（`.proto` ファイルの中身）
- CI/CD

### 受け入れ条件

- `mise install` が成功し、`mise current` で Node.js / Go / pnpm / buf / sqlc / goose のバージョンが表示される
- `pnpm install` がエラーなく完了する
- `pnpm ls --filter ./services/bff` が BFF パッケージを認識する
- 各 Go サービスディレクトリで `go mod tidy` がエラーなく完了する（依存がないので no-op）
- `buf build` が `proto/` ディレクトリに対して成功する（空スキーマまたはスタブ .proto で）
- `just setup` で上記が一括実行できる
- `lefthook install` が成功し `.git/hooks/` にフックが配置される

---

## Phase 2: Docker Compose とデータベース基盤

### やること

PostgreSQL 3 インスタンスと Ory Hydra を Docker Compose で起動し、goose マイグレーションの仕組みを整える。

### やることの詳細リスト

- `docker-compose.yml` の作成（PostgreSQL x3: ec-site-db, product-db, customer-db）
- Ory Hydra コンテナの追加（in-memory モードまたは専用 PostgreSQL）
- 各 Go サービスに goose マイグレーションディレクトリ作成（`services/*/db/migrations/`）
- 初期マイグレーション SQL の作成（各サービスのテーブル定義）
  - product-mgmt: `products` テーブル
  - customer-mgmt: `customers` テーブル
  - ec-site: `carts`、`cart_items`、`orders`、`order_items` テーブル
- sqlc 設定ファイル（`sqlc.yaml`）の作成と goose マイグレーションディレクトリの参照設定
- Ory Hydra の OAuth 2.0 クライアント登録スクリプト（BFF 用、ECサイト用）
- Just に `just db-up` / `just db-migrate` / `just db-reset` タスク追加

### スコープ

- Docker Compose による全インフラの起動
- goose によるマイグレーション適用・ロールバック
- sqlc の設定と goose マイグレーションとの連携確認
- Ory Hydra の Client Credentials Grant 設定

### スコープ外

- アプリケーションサーバーの Docker 化（ローカル直接実行を前提）
- sqlc によるクエリコード生成（Phase 3 で実施）
- シードデータの投入

### 受け入れ条件

- `docker compose up -d` で PostgreSQL x3 と Ory Hydra が起動する
- 各 DB に `psql` で接続できる
- `just db-migrate` で全サービスのマイグレーションが適用され、テーブルが作成される
- `goose -dir services/product-mgmt/db/migrations status` でマイグレーション状態が確認できる
- Ory Hydra の Client Credentials でアクセストークンが取得できる:
  ```
  curl -X POST http://localhost:4444/oauth2/token \
    -d grant_type=client_credentials \
    -d client_id=<client_id> \
    -d client_secret=<client_secret>
  ```
  が JWT を返す

---

## Phase 3: Protobuf スキーマ最小定義とコード生成

### やること

疎通確認に必要な最小限の Protobuf スキーマと sqlc クエリを定義し、buf / sqlc のコード生成パイプラインが動作することを検証する。

### やることの詳細リスト

- 商品管理 API の `.proto` 定義（ProductService: CreateProduct, GetProduct の 2 RPC）
- 顧客管理 API の `.proto` 定義（CustomerService: CreateCustomer, GetCustomer の 2 RPC）
- ECサイト API の `.proto` 定義（CartService: GetCart の 1 RPC）
- `buf.gen.yaml` の更新（connect-es / connect-go プラグイン設定）
- `buf generate` によるコード生成と出力先の整理
- 疎通確認に必要な最小 sqlc クエリ定義と `sqlc generate` による Go コード生成
  - product-mgmt: CreateProduct, GetProduct
  - customer-mgmt: CreateCustomer, GetCustomer
  - ec-site: GetCartByCustomerID, CreateCart, ListCartItems
- 生成コードの `.gitignore` 管理方針の決定

### スコープ

- 疎通確認に必要な最小限の Protobuf スキーマ定義（5 RPC）
- buf による TypeScript / Go コード生成パイプライン
- 疎通確認に必要な最小限の sqlc クエリと Go コード生成
- `buf lint` によるスキーマの品質チェック

### スコープ外

- 全 RPC メソッドの定義（積み残しで対応）
- OrderService の RPC 定義（`.proto` スタブは残すが RPC は定義しない）
- 全 sqlc クエリの定義
- 認証・認可ロジック
- BFF の API 集約ロジック

### 受け入れ条件

- `buf lint` がエラーなく通過する
- `buf generate` が成功し、以下が生成される:
  - `services/bff/gen/` 配下に TypeScript の Connect RPC クライアントコード
  - `gen/go/` 配下に Go の Connect RPC サーバーコード
- `sqlc generate` が成功し、各 Go サービスに型安全なクエリ関数が生成される
- 生成された Go コードが `go build ./...` でコンパイルできる
- 生成された TypeScript コードが `pnpm tsc --noEmit` で型チェックを通過する

---

## Phase 4: 各サービスの最小起動とヘルスチェック

### やること

各サービスのエントリポイントを実装し、サーバー起動と Connect RPC ハンドラのマウントまでを行う。

### やることの詳細リスト

- BFF: Fastify サーバー起動、Connect RPC サーバー/クライアントのセットアップ、ヘルスチェックエンドポイント
- ECサイト: Echo サーバー起動、Connect RPC ハンドラマウント（`echo.WrapHandler`）、DB 接続、ヘルスチェック
- 商品管理: Echo サーバー起動、Connect RPC ハンドラマウント、DB 接続、ヘルスチェック
- 顧客管理: Echo サーバー起動、Connect RPC ハンドラマウント、DB 接続、ヘルスチェック
- 各サービスの Connect RPC ハンドラにスタブ実装（`unimplemented` を返す）
- Just に `just dev` タスク追加（全サービスの並列起動）

### スコープ

- 各サービスの main 関数 / エントリポイント
- Connect RPC ハンドラのマウント（スタブ実装）
- DB コネクションプールの初期化
- ヘルスチェックエンドポイント

### スコープ外

- ビジネスロジックの実装
- 認証・認可ミドルウェア
- フロントエンド UI
- 管理画面 UI

### 受け入れ条件

- `just dev` で 4 サービスが同時に起動する
- 各サービスのヘルスチェックが応答する:
  - `curl http://localhost:3000/health` (BFF) → 200
  - `curl http://localhost:8080/health` (ECサイト) → 200
  - `curl http://localhost:8081/health` (商品管理) → 200
  - `curl http://localhost:8082/health` (顧客管理) → 200
- BFF から商品管理の Connect RPC エンドポイントを curl で叩くと `unimplemented` エラーが返る（疎通確認）
- 各 Go サービスが DB に接続できている（ヘルスチェックに DB ping を含む）

---

## Phase 5: 商品管理サービスの最小実装

### やること

商品管理サービスの CreateProduct / GetProduct を実装し、Connect RPC 経由で DB 操作ができることを検証する。

### やることの詳細リスト

- Connect RPC ハンドラの実装（CreateProduct, GetProduct の 2 RPC）
- sqlc 生成コードを使った DB 操作

### スコープ

- CreateProduct, GetProduct の 2 RPC 実装

### スコープ外

- UpdateProduct, DeleteProduct, ListProducts, AllocateStock
- 管理画面 UI（templ + HTMX）
- 管理画面の認証（echo-jwt）
- シードデータ

### 受け入れ条件

- Connect RPC 経由で商品の作成・取得ができる（buf curl で確認）:
  ```
  buf curl --data '{"name":"テスト商品","price":1000,"stock_quantity":10}' \
    http://localhost:8081/product.v1.ProductService/CreateProduct
  ```
  ```
  buf curl --data '{"id":1}' \
    http://localhost:8081/product.v1.ProductService/GetProduct
  ```

---

## Phase 6: 顧客管理サービスの最小実装

### やること

顧客管理サービスの CreateCustomer / GetCustomer を実装し、商品管理と同様のパターンで疎通を確認する。

### やることの詳細リスト

- Connect RPC ハンドラの実装（CreateCustomer, GetCustomer の 2 RPC）
- sqlc 生成コードを使った DB 操作

### スコープ

- CreateCustomer, GetCustomer の 2 RPC 実装

### スコープ外

- ListCustomers
- 管理画面 UI（templ + HTMX）
- 管理画面の認証（echo-jwt）

### 受け入れ条件

- Connect RPC 経由で顧客の作成・取得ができる（buf curl で確認）:
  ```
  buf curl --data '{"name":"テスト顧客","email":"test@example.com","password":"password"}' \
    http://localhost:8082/customer.v1.CustomerService/CreateCustomer
  ```
  ```
  buf curl --data '{"id":1}' \
    http://localhost:8082/customer.v1.CustomerService/GetCustomer
  ```

---

## Phase 7: ECサイトバックエンドの最小実装

### やること

ECサイトの GetCart を実装し、ECサイト → 商品管理サービスへの Connect RPC 呼び出しによるサービス間通信の疎通を確認する。

### やることの詳細リスト

- Connect RPC ハンドラの実装（GetCart の 1 RPC）
- GetCart 内での自動カート作成（存在しない場合は空カートを作成して返す）
- ECサイト → 商品管理への Connect RPC クライアント呼び出し（認証なし、GetProduct で疎通確認）

### スコープ

- GetCart の 1 RPC 実装
- サービス間の Connect RPC 通信（認証なし）

### スコープ外

- AddItem, RemoveItem, UpdateQuantity
- OrderService の実装（PlaceOrder, ListOrders）
- サービス間認証（OAuth 2.0 Client Credentials）

### 受け入れ条件

- buf curl でカート取得ができる:
  ```
  buf curl --data '{"customer_id":1}' \
    http://localhost:8080/ec.v1.CartService/GetCart
  ```
- ECサイト → 商品管理への Connect RPC 呼び出しが成功する（ログで確認）

---

## Phase 8: BFF の最小実装

### やること

BFF（Fastify）からバックエンドサービスへの Connect RPC プロキシを最小実装し、BFF → バックエンド間の疎通を確認する。

### やることの詳細リスト

- Fastify サーバーの最小実装
- Connect RPC クライアントでバックエンドの商品管理サービスへ転送（1 エンドポイント）

### スコープ

- BFF → 商品管理への 1 エンドポイントのプロキシ疎通

### スコープ外

- エンドユーザー認証（@fastify/jwt、HttpOnly Cookie、リフレッシュトークン）
- BFF → バックエンド間の認証（OAuth 2.0 Client Credentials）
- API 集約ロジック（複数サービスのデータ統合）
- CORS / Cookie 設定

### 受け入れ条件

- curl で BFF 経由で商品情報が返る:
  ```
  curl http://localhost:3000/product.v1.ProductService/GetProduct \
    -H 'Content-Type: application/json' \
    -d '{"id":1}'
  ```

---

## Phase 9: フロントエンドの最小実装

### やること

React で最小ページを作成し、BFF 経由でバックエンドのデータを取得・表示する疎通確認を行う。

### やることの詳細リスト

- Vite + React + TypeScript プロジェクトのセットアップ
- @connectrpc/connect-web による BFF への接続設定
- 商品情報を表示する最小ページ 1 つ

### スコープ

- BFF 経由でバックエンドデータを表示する 1 ページ

### スコープ外

- ログイン / ログアウト画面
- 商品一覧・詳細ページ
- カートページ
- 注文確定・履歴ページ
- 認証状態管理

### 受け入れ条件

- `http://localhost:5173` でフロントエンドが表示される
- BFF 経由で商品情報が画面に表示される

---

## Phase 10: 結合確認

### やること

フロントエンド → BFF → バックエンド → DB の全レイヤーの疎通を一括確認する。

### やることの詳細リスト

- 全サービスの一括起動確認（`just dev`）
- フロントエンド → DB までの疎通 E2E 確認

### スコープ

- 全サービス一括起動
- 端から端までの疎通確認

### スコープ外

- シードデータの整備
- lefthook フックの本設定
- Just タスクの整理
- CLAUDE.md の更新
- E2E ビジネスシナリオ

### 受け入れ条件

- `just setup && docker compose up -d && just db-migrate && just dev` で全サービスが起動する
- 以下の疎通シナリオが手動で完走する:
  1. buf curl で商品管理に商品を作成
  2. フロントエンドで BFF 経由で商品情報が表示される
- `just generate` で Protobuf / sqlc のコード生成が一括実行できる

---

## フェーズ間の依存関係

```
Phase 1  開発環境とモノレポ基盤
  │
  ▼
Phase 2  Docker Compose とデータベース基盤
  │
  ▼
Phase 3  Protobuf / sqlc 最小定義とコード生成
  │
  ▼
Phase 4  各サービスの最小起動
  │
  ├──────────────────┐
  ▼                  ▼
Phase 5            Phase 6
商品管理(2 RPC)     顧客管理(2 RPC)
  │                  │
  └─────┬────────────┘
        ▼
      Phase 7  ECサイト(1 RPC + サービス間通信)
        │
        ▼
      Phase 8  BFF(1エンドポイントプロキシ)
        │
        ▼
      Phase 9  フロントエンド(1ページ)
        │
        ▼
      Phase 10  結合確認
```

Phase 5（商品管理）と Phase 6（顧客管理）は並行実施可能。

---

## 積み残し（実装の提案）

以下は初期実装（疎通確認）から除外した機能。各フェーズの疎通確認完了後、順次追加で対応する。

### Proto / sqlc 追加定義

- ProductService: ListProducts, UpdateProduct, DeleteProduct, AllocateStock
- CustomerService: ListCustomers, GetCustomerByEmail
- CartService: AddItem, RemoveItem, UpdateQuantity
- OrderService: PlaceOrder, ListOrders（`.proto` の RPC 定義含む）
- 上記に対応する sqlc クエリの追加

### 商品管理サービス

- 残りの RPC 実装（List, Update, Delete, AllocateStock）
- 管理画面 UI（templ + HTMX）— 商品一覧、登録・編集フォーム
- 管理画面の認証（echo-jwt）— ログイン画面、JWT 発行、Cookie 格納
- 管理者ユーザーのシードデータ
- 商品のシードデータ

### 顧客管理サービス

- 残りの RPC 実装（ListCustomers）
- 管理画面 UI（templ + HTMX）— 顧客一覧、顧客詳細
- 管理画面の認証（echo-jwt）
- 注文履歴の表示（ECサイト API との連携）

### ECサイトバックエンド

- カート操作の全 RPC 実装（AddItem, RemoveItem, UpdateQuantity）
- 注文の全 RPC 実装（PlaceOrder, ListOrders）
- 顧客管理への Connect RPC クライアント呼び出し（顧客情報参照）
- サービス間認証（OAuth 2.0 Client Credentials、Ory Hydra 連携）
- Connect RPC インターセプタによるトークン付与・検証

### BFF

- エンドユーザー認証（@fastify/jwt、HttpOnly Cookie、リフレッシュトークン）
- BFF → バックエンド間認証（OAuth 2.0 Client Credentials）
- API 集約ロジック（商品一覧 + 在庫情報の統合等）
- CORS / Cookie 設定
- Connect RPC サーバーハンドラ（フロント向け全 API）

### フロントエンド

- ログイン / ログアウト画面
- 商品一覧・詳細ページ
- カートページ
- 注文確定・履歴ページ
- 認証状態管理（Cookie ベース）
- 商品画像プレースホルダー

### 結合・品質整備

- E2E ビジネスシナリオ（商品登録 → カート → 注文 → 在庫減少 → 注文履歴）
- シードデータの整備（管理者ユーザー、テスト商品、テスト顧客）
- lefthook フックの本設定（lint、型チェック、buf lint）
- Just タスクの整理（setup, dev, build, lint, test, db-migrate, db-reset, generate）
- CLAUDE.md の更新（開発手順、コマンド一覧）
