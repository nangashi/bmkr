# Phase 2: Docker Compose とデータベース基盤 — タスクリスト

## Task 1: Docker Compose — PostgreSQL x3 の構成

### やること

`docker-compose.yml` を作成し、3 つの PostgreSQL インスタンス（ec-site-db, product-db, customer-db）を定義する。

### 方針

- プロジェクトルートに `docker-compose.yml` を配置する
- Database-per-Service パターンに従い、各サービスに独立した PostgreSQL コンテナを割り当てる
- ポートはホスト側で衝突しないようにずらす（5432 / 5433 / 5434）
- データ永続化のため named volume を使用する
- ヘルスチェック設定を入れ、DB が ready になるまで待機できるようにする
- 環境変数（DB名、ユーザー、パスワード）は `.env` ファイルではなく `docker-compose.yml` に直書きする（ローカル専用のため）

### スコープ

- `docker-compose.yml` の作成（PostgreSQL x3 のみ）
- 各 DB の接続情報定義

### スコープ外

- Ory Hydra コンテナ（Task 6 で追加）
- アプリケーションサーバーのコンテナ化
- `.env` ファイルによる環境変数の外出し

### 受け入れ条件

- `docker compose up -d` で 3 つの PostgreSQL コンテナが起動する:
  ```
  docker compose ps
  ```
  で ec-site-db, product-db, customer-db が `running (healthy)` と表示される
- 各 DB に接続できる:
  ```
  docker compose exec product-db psql -U postgres -d product -c "SELECT 1"
  docker compose exec customer-db psql -U postgres -d customer -c "SELECT 1"
  docker compose exec ec-site-db psql -U postgres -d ecsite -c "SELECT 1"
  ```
- `docker compose down` で全コンテナが停止する

---

## Task 2: 商品管理 — goose マイグレーションと初期テーブル定義

### やること

商品管理サービスの goose マイグレーションディレクトリを作成し、初期テーブル定義の SQL を配置する。

### 方針

- マイグレーションディレクトリは `services/product-mgmt/db/migrations/` とする
- goose のタイムスタンプ形式を使用する（sqlc との辞書順不整合を回避）
- `-- +goose Up` / `-- +goose Down` マーカーを使用する
- 初期マイグレーションで以下のテーブルを作成する:
  - `products`: 商品情報（id, name, description, price, stock_quantity, created_at, updated_at）
  - `admin_users`: 管理画面の認証用（id, email, password_hash, created_at）
- 主キーは UUID ではなく BIGSERIAL を使用する（学習用途でシンプルさを優先）

### スコープ

- `services/product-mgmt/db/migrations/` ディレクトリの作成
- 初期マイグレーション SQL ファイルの作成（`products` テーブル、`admin_users` テーブル）

### スコープ外

- マイグレーションの実行（Task 8 の Justfile タスクで一括実行）
- sqlc 設定（Task 5 で実施）
- シードデータの投入

### 受け入れ条件

- マイグレーションファイルが存在する:
  ```
  ls services/product-mgmt/db/migrations/*.sql
  ```
- マイグレーションファイルに `-- +goose Up` と `-- +goose Down` マーカーが含まれている:
  ```
  grep "+goose" services/product-mgmt/db/migrations/*.sql
  ```
- Docker Compose 起動中に goose でマイグレーションを適用できる:
  ```
  GOOSE_DRIVER=postgres \
  GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5433/product?sslmode=disable" \
  goose -dir services/product-mgmt/db/migrations up
  ```
- テーブルが作成されていることを確認:
  ```
  docker compose exec product-db psql -U postgres -d product \
    -c "\dt"
  ```
  で `products` と `admin_users` が表示される

---

## Task 3: 顧客管理 — goose マイグレーションと初期テーブル定義

### やること

顧客管理サービスの goose マイグレーションディレクトリを作成し、初期テーブル定義の SQL を配置する。

### 方針

- マイグレーションディレクトリは `services/customer-mgmt/db/migrations/` とする
- 初期マイグレーションで以下のテーブルを作成する:
  - `customers`: 顧客情報（id, name, email, password_hash, created_at, updated_at）
  - `admin_users`: 管理画面の認証用（id, email, password_hash, created_at）
- `customers.email` にはユニーク制約を付与する
- 方式は Task 2（商品管理）と統一する

### スコープ

- `services/customer-mgmt/db/migrations/` ディレクトリの作成
- 初期マイグレーション SQL ファイルの作成

### スコープ外

- マイグレーションの実行
- sqlc 設定
- シードデータの投入

### 受け入れ条件

- マイグレーションファイルが存在する:
  ```
  ls services/customer-mgmt/db/migrations/*.sql
  ```
- Docker Compose 起動中に goose でマイグレーションを適用できる:
  ```
  GOOSE_DRIVER=postgres \
  GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5434/customer?sslmode=disable" \
  goose -dir services/customer-mgmt/db/migrations up
  ```
- テーブルが作成されていることを確認:
  ```
  docker compose exec customer-db psql -U postgres -d customer \
    -c "\dt"
  ```
  で `customers` と `admin_users` が表示される

---

## Task 4: ECサイト — goose マイグレーションと初期テーブル定義

### やること

ECサイトサービスの goose マイグレーションディレクトリを作成し、初期テーブル定義の SQL を配置する。

### 方針

- マイグレーションディレクトリは `services/ec-site/db/migrations/` とする
- 初期マイグレーションで以下のテーブルを作成する:
  - `carts`: カート（id, customer_id, created_at, updated_at）
  - `cart_items`: カート内商品（id, cart_id, product_id, quantity, created_at）
  - `orders`: 注文（id, customer_id, total_amount, status, created_at）
  - `order_items`: 注文明細（id, order_id, product_id, product_name, price, quantity）
- `customer_id` / `product_id` は外部キー制約を付けない（Database-per-Service のため、参照先は別 DB）
- `orders.status` は文字列型とする（例: `pending`, `confirmed`, `cancelled`）

### スコープ

- `services/ec-site/db/migrations/` ディレクトリの作成
- 初期マイグレーション SQL ファイルの作成

### スコープ外

- マイグレーションの実行
- sqlc 設定
- シードデータの投入

### 受け入れ条件

- マイグレーションファイルが存在する:
  ```
  ls services/ec-site/db/migrations/*.sql
  ```
- Docker Compose 起動中に goose でマイグレーションを適用できる:
  ```
  GOOSE_DRIVER=postgres \
  GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/ecsite?sslmode=disable" \
  goose -dir services/ec-site/db/migrations up
  ```
- テーブルが作成されていることを確認:
  ```
  docker compose exec ec-site-db psql -U postgres -d ecsite \
    -c "\dt"
  ```
  で `carts`、`cart_items`、`orders`、`order_items` が表示される

---

## Task 5: 各サービスの sqlc 設定ファイルの作成

### やること

3 つの Go サービスそれぞれに `sqlc.yaml` を作成し、goose マイグレーションディレクトリをスキーマソースとして参照する。

### 方針

- 各サービスの `sqlc.yaml` は `services/<service>/sqlc.yaml` に配置する
- `schema` には goose マイグレーションディレクトリ（`db/migrations/`）を指定する
- `queries` ディレクトリ（`db/queries/`）を指定する（Phase 3 でクエリファイルを追加）
- `engine` は `postgresql`、`gen.go.out` で Go コードの出力先を指定する
- `emit_interface: true` を設定し、テスト用の Querier interface を生成する
- pgx/v5 ドライバを使用する設定にする

### スコープ

- 3 サービス分の `sqlc.yaml` 作成
- クエリディレクトリ（`db/queries/`）の作成（中身は空、`.keep` ファイルのみ）

### スコープ外

- SQL クエリファイルの作成（Phase 3）
- `sqlc generate` の実行（クエリがないため生成するコードがない）

### 受け入れ条件

- 各サービスに sqlc 設定ファイルが存在する:
  ```
  cat services/product-mgmt/sqlc.yaml
  cat services/customer-mgmt/sqlc.yaml
  cat services/ec-site/sqlc.yaml
  ```
- 各 sqlc.yaml が goose マイグレーションディレクトリを `schema` として参照している
- 各サービスに `db/queries/` ディレクトリが存在する:
  ```
  ls services/product-mgmt/db/queries/
  ls services/customer-mgmt/db/queries/
  ls services/ec-site/db/queries/
  ```
- `sqlc compile` が各サービスで成功する（スキーマの解析が通ること）:
  ```
  cd services/product-mgmt && sqlc compile
  cd services/customer-mgmt && sqlc compile
  cd services/ec-site && sqlc compile
  ```

---

## Task 6: Docker Compose — Ory Hydra の追加

### やること

`docker-compose.yml` に Ory Hydra コンテナを追加し、OAuth 2.0 認可サーバーをローカルで起動できるようにする。

### 方針

- Ory Hydra は in-memory モード（`DSN=memory`）で起動する（DB 不要で軽量）
- Public API（トークンエンドポイント）: ポート 4444
- Admin API（クライアント管理）: ポート 4445
- JWT モードを有効にし、アクセストークンを JWT 形式で発行する
- `--dev` フラグを使用して HTTPS 要件を無効化する（ローカル開発用）
- ヘルスチェック設定を入れる

### スコープ

- `docker-compose.yml` への Ory Hydra サービス定義の追加
- Hydra の起動確認

### スコープ外

- OAuth 2.0 クライアントの登録（Task 7 で実施）
- アプリケーションコードからのトークン取得
- Login/Consent App（Client Credentials Grant では不要）

### 受け入れ条件

- `docker compose up -d` で Hydra が起動する:
  ```
  docker compose ps
  ```
  で hydra が `running (healthy)` と表示される
- Public API が応答する:
  ```
  curl http://localhost:4444/.well-known/openid-configuration
  ```
  が JSON を返す
- Admin API が応答する:
  ```
  curl http://localhost:4445/admin/clients
  ```
  が空の配列 `[]` を返す

---

## Task 7: Ory Hydra — OAuth 2.0 クライアント登録スクリプトの作成

### やること

BFF 用・ECサイト用の OAuth 2.0 クライアントを Hydra に登録するスクリプトを作成する。

### 方針

- `scripts/hydra-setup.sh` にクライアント登録スクリプトを配置する
- Hydra Admin API を使ってクライアントを登録する（`POST /admin/clients`）
- 以下の 2 クライアントを登録する:
  - `bff-client`: BFF → バックエンド API 呼び出し用
  - `ec-site-client`: ECサイト → 商品管理 / 顧客管理 呼び出し用
- `grant_types` は `client_credentials` のみ
- `token_endpoint_auth_method` は `client_secret_post`
- スクリプトは冪等にする（既にクライアントが存在する場合はスキップまたは上書き）
- `scope` は `service` とする（将来的にサービスごとのスコープ分割も可能）

### スコープ

- クライアント登録スクリプトの作成
- スクリプトの実行によるクライアント登録
- Client Credentials Grant でのトークン取得確認

### スコープ外

- アプリケーションコードへの OAuth 2.0 クライアント設定の組み込み
- スコープベースの認可制御の詳細設計

### 受け入れ条件

- スクリプトが存在する:
  ```
  ls scripts/hydra-setup.sh
  ```
- Docker Compose 起動中にスクリプトを実行するとクライアントが登録される:
  ```
  bash scripts/hydra-setup.sh
  ```
- 登録されたクライアントが確認できる:
  ```
  curl http://localhost:4445/admin/clients
  ```
  で `bff-client` と `ec-site-client` が表示される
- Client Credentials Grant でトークンが取得できる:
  ```
  curl -s -X POST http://localhost:4444/oauth2/token \
    -d grant_type=client_credentials \
    -d client_id=bff-client \
    -d client_secret=bff-secret \
    -d scope=service
  ```
  が `access_token` を含む JSON を返す
- 取得したトークンが JWT 形式である（`.` で 3 パートに分かれている）

---

## Task 8: Justfile — DB 関連タスクの追加

### やること

Justfile に Docker Compose の起動・停止と goose マイグレーションを実行するタスクを追加する。

### 方針

- Phase 1 で作成した Justfile に追記する
- タスク名は `db-up` / `db-down` / `db-migrate` / `db-rollback` / `db-reset` / `hydra-setup` とする
- `db-migrate` は 3 サービスのマイグレーションを順に実行する
- 各タスクの接続情報はタスク内に直書きする（docker-compose.yml と一致させる）
- `db-reset` は volume 削除 → 再起動 → マイグレーション再適用を一括で行う

### スコープ

- Justfile への DB 関連タスクの追加
- `hydra-setup` タスク（スクリプト実行のラッパー）

### スコープ外

- `dev`（アプリサーバー起動）タスク（Phase 4 で追加）
- シードデータ投入タスク

### 受け入れ条件

- `just --list` で新しいタスクが表示される:
  ```
  just --list
  ```
  に `db-up`、`db-down`、`db-migrate`、`db-rollback`、`db-reset`、`hydra-setup` が含まれる
- `just db-up` で Docker Compose が起動する
- `just db-migrate` で 3 サービスのマイグレーションが順に適用される:
  ```
  just db-migrate
  ```
  の出力に各サービスの `OK` が含まれる
- `just hydra-setup` で Ory Hydra のクライアント登録が実行される
- `just db-reset` で DB が初期化されマイグレーションが再適用される:
  ```
  just db-reset
  ```
  の後、全テーブルが再作成されている

---

## タスク依存関係

```
Task 1: Docker Compose — PostgreSQL x3
  │
  ├─► Task 2: 商品管理 goose マイグレーション
  ├─► Task 3: 顧客管理 goose マイグレーション
  ├─► Task 4: ECサイト goose マイグレーション
  │     │
  │     └──► Task 5: 各サービスの sqlc 設定
  │
  └─► Task 6: Docker Compose — Ory Hydra
        │
        └─► Task 7: Ory Hydra クライアント登録スクリプト
              │
              └─► Task 8: Justfile DB タスク追加
```

- Task 1 が全タスクの前提（PostgreSQL が必要）
- Task 2, 3, 4 は互いに独立（各サービスの DB は独立）
- Task 5 は Task 2, 3, 4 に依存（マイグレーション SQL をスキーマとして参照）
- Task 6 は Task 1 に依存（docker-compose.yml への追記）
- Task 7 は Task 6 に依存（Hydra が起動している必要）
- Task 8 は全タスク完了後に追加（全操作を Justfile にまとめる）
