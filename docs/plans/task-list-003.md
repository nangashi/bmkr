# Phase 3: Protobuf スキーマ最小定義とコード生成 — タスクリスト

## Task 1: ProductService の Protobuf スキーマ定義

### やること

`proto/product/v1/product.proto` に疎通確認用の最小メッセージ型と RPC メソッドを定義する。

### 方針

- 既存のスタブ `.proto` ファイルにメッセージ型と RPC を追加する
- DB スキーマ（`products` テーブル）に対応する `Product` メッセージを定義する
- 疎通確認に必要な 2 RPC のリクエスト/レスポンスメッセージを定義する:
  - `CreateProductRequest` / `CreateProductResponse`
  - `GetProductRequest` / `GetProductResponse`
- `google.protobuf.Timestamp` を使用して日時フィールドを表現する
- ID は `int64` とする（DB の BIGSERIAL に対応）
- 価格は `int64`（円単位、小数なし）

### スコープ

- `product.proto` へのメッセージ型と 2 RPC の追加

### スコープ外

- ListProducts, UpdateProduct, DeleteProduct, AllocateStock の定義（積み残し）
- コード生成の実行

### 受け入れ条件

- `buf build` がエラーなく完了する:
  ```
  buf build
  echo $?
  ```
  が `0` を返す
- `Product` メッセージが定義されている:
  ```
  grep "message Product " proto/product/v1/product.proto
  ```
- 2 つの RPC メソッドが定義されている:
  ```
  grep -c "rpc " proto/product/v1/product.proto
  ```
  が `2` を返す
- リクエスト/レスポンスメッセージが 2 組（4 個）+ Product = 5 メッセージ:
  ```
  grep -c "^message " proto/product/v1/product.proto
  ```
  が `5` を返す

---

## Task 2: CustomerService の Protobuf スキーマ定義

### やること

`proto/customer/v1/customer.proto` に疎通確認用の最小メッセージ型と RPC メソッドを定義する。

### 方針

- DB スキーマ（`customers` テーブル）に対応する `Customer` メッセージを定義する
- 以下の 2 RPC を定義する:
  - `CreateCustomer` — 顧客の新規作成
  - `GetCustomer` — 顧客の取得（ID 指定）
- `password_hash` はレスポンスに含めない（セキュリティ考慮）
- リクエストにはパスワード（平文）を受け取り、サービス側でハッシュ化する想定

### スコープ

- `customer.proto` へのメッセージ型と 2 RPC の追加

### スコープ外

- ListCustomers, GetCustomerByEmail の定義（積み残し）
- コード生成の実行

### 受け入れ条件

- `buf lint` がエラーなく完了する:
  ```
  buf lint
  echo $?
  ```
  が `0` を返す
- `Customer` メッセージが定義されている:
  ```
  grep "message Customer " proto/customer/v1/customer.proto
  ```
- 2 つの RPC メソッドが定義されている:
  ```
  grep -c "rpc " proto/customer/v1/customer.proto
  ```
  が `2` を返す

---

## Task 3: CartService の Protobuf スキーマ定義

### やること

`proto/ec/v1/cart.proto` に疎通確認用の最小メッセージ型と RPC メソッドを定義する。

### 方針

- DB スキーマ（`carts`, `cart_items` テーブル）に対応する `Cart`、`CartItem` メッセージを定義する
- 以下の 1 RPC を定義する:
  - `GetCart` — カート内容の取得（customer_id 指定）
- `customer_id` でカートを識別する（1 顧客 1 カート）

### スコープ

- `cart.proto` へのメッセージ型と 1 RPC の追加

### スコープ外

- AddItem, RemoveItem, UpdateQuantity の定義（積み残し）
- `order.proto` への RPC 定義（積み残し、スタブのまま維持）
- コード生成の実行

### 受け入れ条件

- `buf lint` がエラーなく完了する:
  ```
  buf lint
  echo $?
  ```
  が `0` を返す
- `Cart` と `CartItem` メッセージが定義されている:
  ```
  grep "message Cart " proto/ec/v1/cart.proto
  grep "message CartItem " proto/ec/v1/cart.proto
  ```
- 1 つの RPC メソッドが定義されている:
  ```
  grep -c "rpc " proto/ec/v1/cart.proto
  ```
  が `1` を返す

---

## Task 4: 全 Protobuf スキーマの buf lint 一括検証

### やること

全 `.proto` ファイルに対して `buf lint` を実行し、STANDARD ルールへの適合を確認・修正する。

### 方針

- `buf lint` を実行し、エラーがあれば修正する
- STANDARD ルールセットの主な要件:
  - サービス名は `Service` サフィックス
  - RPC リクエスト/レスポンスは `<RPC名>Request` / `<RPC名>Response` 命名
  - フィールドは `lower_snake_case`
  - enum 値は `UPPER_SNAKE_CASE` でプレフィックス付き
- `buf breaking` は初回スキーマのため実行しない

### スコープ

- 全 `.proto` ファイルの lint 適合確認
- lint エラーの修正

### スコープ外

- コード生成の実行
- `.proto` ファイルへの新規 RPC / メッセージの追加

### 受け入れ条件

- `buf lint` がエラーなく完了する:
  ```
  buf lint
  echo $?
  ```
  が `0` を返す
- RPC を持つ `.proto` ファイルが空でないことを確認:
  ```
  grep -c "rpc " proto/product/v1/product.proto proto/customer/v1/customer.proto proto/ec/v1/cart.proto
  ```
  product.proto: 2、customer.proto: 2、cart.proto: 1

---

## Task 5: buf.gen.yaml の更新（生成先ディレクトリの調整）

### やること

`buf.gen.yaml` の出力先を整理し、TypeScript コードを BFF パッケージ配下に、Go コードをルートの `gen/go/` に生成する構成にする。

### 方針

- TypeScript の生成先を `gen/ts` から `services/bff/gen` に変更する（BFF が唯一の TypeScript 消費者）
- Go の生成先は `gen/go` のまま維持する（複数 Go サービスが共有するため）
- `clean` オプションを有効にして、生成前に出力ディレクトリをクリーンする
- `.gitignore` の `gen/` ルールにより生成コードは Git 管理外とする

### スコープ

- `buf.gen.yaml` の更新

### スコープ外

- `buf generate` の実行（Task 7 で実施）
- Go モジュールの設定

### 受け入れ条件

- `buf.gen.yaml` の TypeScript 出力先が `services/bff/gen` になっている:
  ```
  grep "services/bff/gen" buf.gen.yaml
  ```
- Go の出力先が `gen/go` のまま維持されている:
  ```
  grep "gen/go" buf.gen.yaml
  ```
- `buf.gen.yaml` が有効な構文である:
  ```
  buf build
  echo $?
  ```

---

## Task 6: Go 生成コード用モジュールの作成

### やること

`gen/go/` に独立した Go モジュールを作成し、生成コードが他の Go サービスから参照可能な状態にする。

### 方針

- `gen/go/go.mod` を作成する（モジュールパス: `github.com/nangashi/bmkr/gen/go`）
- proto ファイルの `go_package` オプションが `github.com/nangashi/bmkr/gen/go/...` となっているため、このモジュールパスに合わせる
- 依存パッケージ（`google.golang.org/protobuf`、`connectrpc.com/connect`）は `buf generate` 後に `go mod tidy` で追加する
- `.gitignore` の `gen/` ルールに該当するため、このファイルは生成コードと同様に Git 管理外となる点に注意する

### スコープ

- `gen/go/go.mod` の作成

### スコープ外

- `buf generate` の実行
- Go サービスからの参照設定（Task 8 で実施）

### 受け入れ条件

- `gen/go/go.mod` が存在する:
  ```
  cat gen/go/go.mod
  ```
  でモジュールパス `github.com/nangashi/bmkr/gen/go` が確認できる
- Go バージョンが `.mise.toml` と一致している:
  ```
  grep "go " gen/go/go.mod
  ```

---

## Task 7: buf generate の実行（Go / TypeScript コード生成）

### やること

`buf generate` を実行し、全 `.proto` ファイルから Go / TypeScript のコードを生成する。

### 方針

- `buf generate` を実行する
- 生成後、Go モジュール（`gen/go/`）で `go mod tidy` を実行して依存関係を解決する
- 生成されるファイルの確認:
  - `gen/go/product/v1/` — ProductService の Go コード（pb + connect）
  - `gen/go/customer/v1/` — CustomerService の Go コード（pb + connect）
  - `gen/go/ec/v1/` — CartService の Go コード（pb + connect）、OrderService は pb のみ（RPC なし）
  - `services/bff/gen/product/v1/` — ProductService の TypeScript コード
  - `services/bff/gen/customer/v1/` — CustomerService の TypeScript コード
  - `services/bff/gen/ec/v1/` — CartService の TypeScript コード

### スコープ

- `buf generate` の実行
- `gen/go/go.mod` の依存関係解決（`go mod tidy`）

### スコープ外

- 各 Go サービスの `go.mod` への依存追加（Task 8 で実施）
- BFF の TypeScript 依存パッケージ追加（Task 9 で実施）

### 受け入れ条件

- `buf generate` がエラーなく完了する:
  ```
  buf generate
  echo $?
  ```
  が `0` を返す
- Go コードが生成されている:
  ```
  ls gen/go/product/v1/*.go
  ls gen/go/customer/v1/*.go
  ls gen/go/ec/v1/*.go
  ```
- TypeScript コードが生成されている:
  ```
  ls services/bff/gen/product/v1/*.ts
  ls services/bff/gen/customer/v1/*.ts
  ls services/bff/gen/ec/v1/*.ts
  ```
- Go の生成コードモジュールの依存解決ができる:
  ```
  cd gen/go && go mod tidy && echo $?
  ```
  が `0` を返す

---

## Task 8: Go サービスの依存関係設定（生成コード参照）

### やること

各 Go サービスの `go.mod` に生成コードモジュールへの参照と必要な依存パッケージを追加する。

### 方針

- 各 Go サービスの `go.mod` に以下を追加する:
  - `require github.com/nangashi/bmkr/gen/go`（生成コードモジュール）
  - `replace github.com/nangashi/bmkr/gen/go => ../../gen/go`（ローカル参照）
- `go mod tidy` で間接依存を解決する
- 3 サービス（ec-site, product-mgmt, customer-mgmt）全てに同じ設定を行う

### スコープ

- 3 サービスの `go.mod` への依存追加と `replace` ディレクティブの設定
- `go mod tidy` の実行

### スコープ外

- Go ソースコードの作成（Phase 4 以降）
- `go build` の実行（Go ソースコードがないため）

### 受け入れ条件

- 各サービスの `go.mod` に `replace` ディレクティブが設定されている:
  ```
  grep "replace.*gen/go" services/product-mgmt/go.mod
  grep "replace.*gen/go" services/customer-mgmt/go.mod
  grep "replace.*gen/go" services/ec-site/go.mod
  ```
- 各サービスで `go mod tidy` がエラーなく完了する:
  ```
  cd services/product-mgmt && go mod tidy && echo "OK"
  cd services/customer-mgmt && go mod tidy && echo "OK"
  cd services/ec-site && go mod tidy && echo "OK"
  ```

---

## Task 9: BFF の TypeScript 依存パッケージ追加

### やること

BFF パッケージに Connect RPC / Protobuf の TypeScript ランタイム依存を追加し、生成コードが型チェックを通過する状態にする。

### 方針

- `services/bff/` に以下のパッケージを追加する:
  - `@connectrpc/connect` — Connect RPC クライアントランタイム
  - `@bufbuild/protobuf` — Protobuf ランタイム
- `pnpm add` で依存追加する
- `tsconfig.json` を必要に応じて更新する（生成コードディレクトリの include 追加等）

### スコープ

- BFF への npm パッケージ追加
- `tsconfig.json` の調整

### スコープ外

- Fastify 等のアプリケーション依存（Phase 4 以降）
- TypeScript ソースコードの作成

### 受け入れ条件

- 依存パッケージがインストールされている:
  ```
  pnpm ls --filter @bmkr/bff @connectrpc/connect @bufbuild/protobuf
  ```
  で両パッケージが表示される
- 生成された TypeScript コードが型チェックを通過する:
  ```
  cd services/bff && pnpm exec tsc --noEmit
  echo $?
  ```
  が `0` を返す

---

## Task 10: 生成 Go コードのビルド検証

### やること

生成された Go コード（`gen/go/`）が正常にビルドできることを確認する。

### 方針

- `gen/go/` ディレクトリで `go build ./...` を実行する
- ビルドエラーがあれば、proto 定義または buf.gen.yaml を修正する
- 各 Go サービスからの import が解決できることを確認する

### スコープ

- `gen/go/` の Go ビルド確認
- ビルドエラーの修正

### スコープ外

- 各サービスの `main.go` 作成（Phase 4）
- テストの実行

### 受け入れ条件

- 生成コードモジュールのビルドが成功する:
  ```
  cd gen/go && go build ./...
  echo $?
  ```
  が `0` を返す
- `go vet` もエラーなく通過する:
  ```
  cd gen/go && go vet ./...
  echo $?
  ```
  が `0` を返す

---

## Task 11: 商品管理の sqlc クエリ定義

### やること

商品管理サービスの sqlc クエリファイルを作成し、疎通確認に必要な最小クエリを定義する。

### 方針

- `services/product-mgmt/db/queries/products.sql` にクエリを定義する
- sqlc のアノテーション（`-- name: <名前> :<タイプ>`）を使用する
- 以下の 2 クエリを定義する:
  - `GetProduct` (`:one`) — ID 指定で商品を取得
  - `CreateProduct` (`:one`) — 商品の新規作成（RETURNING で作成結果を返す）

### スコープ

- `services/product-mgmt/db/queries/products.sql` の作成

### スコープ外

- ListProducts, UpdateProduct, DeleteProduct, AllocateStock のクエリ（積み残し）
- `admin_users` テーブルのクエリ
- `sqlc generate` の実行（Task 14 で一括実施）

### 受け入れ条件

- クエリファイルが存在する:
  ```
  ls services/product-mgmt/db/queries/products.sql
  ```
- sqlc のアノテーションが正しく記載されている:
  ```
  grep -c "^-- name:" services/product-mgmt/db/queries/products.sql
  ```
  が `2` を返す
- sqlc のコンパイルが通る:
  ```
  cd services/product-mgmt && sqlc compile
  echo $?
  ```
  が `0` を返す

---

## Task 12: 顧客管理の sqlc クエリ定義

### やること

顧客管理サービスの sqlc クエリファイルを作成し、疎通確認に必要な最小クエリを定義する。

### 方針

- `services/customer-mgmt/db/queries/customers.sql` にクエリを定義する
- 以下の 2 クエリを定義する:
  - `GetCustomer` (`:one`) — ID 指定で顧客を取得
  - `CreateCustomer` (`:one`) — 顧客の新規作成（RETURNING で作成結果を返す）

### スコープ

- `services/customer-mgmt/db/queries/customers.sql` の作成

### スコープ外

- ListCustomers, GetCustomerByEmail のクエリ（積み残し）
- `admin_users` テーブルのクエリ
- `sqlc generate` の実行

### 受け入れ条件

- クエリファイルが存在する:
  ```
  ls services/customer-mgmt/db/queries/customers.sql
  ```
- sqlc のアノテーションが正しく記載されている:
  ```
  grep -c "^-- name:" services/customer-mgmt/db/queries/customers.sql
  ```
  が `2` を返す
- sqlc のコンパイルが通る:
  ```
  cd services/customer-mgmt && sqlc compile
  echo $?
  ```
  が `0` を返す

---

## Task 13: ECサイトの sqlc クエリ定義（カート最小）

### やること

ECサイトサービスのカート関連 sqlc クエリファイルを作成し、GetCart の疎通確認に必要な最小クエリを定義する。

### 方針

- `services/ec-site/db/queries/carts.sql` にカート関連クエリを定義する
- 以下の 3 クエリを定義する:
  - `GetCartByCustomerID` (`:one`) — 顧客 ID でカートを取得
  - `CreateCart` (`:one`) — カートの新規作成（GetCart 内で自動作成用）
  - `ListCartItems` (`:many`) — カート内の商品一覧取得

### スコープ

- `services/ec-site/db/queries/carts.sql` の作成

### スコープ外

- AddCartItem, RemoveCartItem, UpdateCartItemQuantity, ClearCartItems のクエリ（積み残し）
- 注文関連クエリ（積み残し）
- `sqlc generate` の実行

### 受け入れ条件

- クエリファイルが存在する:
  ```
  ls services/ec-site/db/queries/carts.sql
  ```
- sqlc のアノテーションが正しく記載されている:
  ```
  grep -c "^-- name:" services/ec-site/db/queries/carts.sql
  ```
  が `3` を返す
- sqlc のコンパイルが通る:
  ```
  cd services/ec-site && sqlc compile
  echo $?
  ```
  が `0` を返す

---

## Task 14: sqlc generate の実行（全サービス一括）

### やること

全 Go サービスで `sqlc generate` を実行し、型安全なクエリ関数の Go コードを生成する。

### 方針

- 3 サービスそれぞれで `sqlc generate` を実行する
- 生成先は各サービスの `db/generated/` ディレクトリ（`sqlc.yaml` で設定済み）
- 生成後、各サービスで `go mod tidy` を実行して依存関係を解決する
- 生成されるファイル:
  - `db.go` — DB インターフェース
  - `models.go` — テーブルに対応する Go 構造体
  - `*.sql.go` — クエリごとの関数
  - `querier.go` — Querier インターフェース（`emit_interface: true`）

### スコープ

- 3 サービスの `sqlc generate` 実行
- 生成コードの確認
- `go mod tidy` による依存解決

### スコープ外

- 生成コードを使ったアプリケーションの実装

### 受け入れ条件

- 全サービスで `sqlc generate` が成功する:
  ```
  cd services/product-mgmt && sqlc generate && echo "OK"
  cd services/customer-mgmt && sqlc generate && echo "OK"
  cd services/ec-site && sqlc generate && echo "OK"
  ```
- 各サービスに生成コードが存在する:
  ```
  ls services/product-mgmt/db/generated/
  ls services/customer-mgmt/db/generated/
  ls services/ec-site/db/generated/
  ```
  各ディレクトリに `db.go`、`models.go`、`querier.go` が含まれる
- 生成コードを含めて Go ビルドが通る:
  ```
  cd services/product-mgmt && go mod tidy && go build ./...
  cd services/customer-mgmt && go mod tidy && go build ./...
  cd services/ec-site && go mod tidy && go build ./...
  ```

---

## Task 15: 生成コードの .gitignore 管理方針の確定

### やること

生成コード（buf / sqlc）の Git 管理方針を確定し、`.gitignore` を調整する。

### 方針

- 生成コードは Git 管理外とする（現状の `.gitignore` の `gen/` ルールを維持）
- 理由: `.proto` ファイルと `db/queries/*.sql` がソースオブトゥルースであり、生成コードは常に再生成可能
- sqlc の生成先 `db/generated/` も `.gitignore` に追加する
- `services/bff/gen/` も `.gitignore` に追加する（TypeScript 生成コード）
- Justfile の `generate` タスクに sqlc generate を追加し、`just generate` で全コード生成を一括実行可能にする

### スコープ

- `.gitignore` の更新（sqlc 生成コードと BFF 生成コードの追加）
- Justfile の `generate` タスクの更新

### スコープ外

- CI/CD での自動生成設定

### 受け入れ条件

- `.gitignore` に sqlc 生成コードのルールが含まれている:
  ```
  grep "db/generated" .gitignore
  ```
- `.gitignore` に BFF 生成コードのルールが含まれている:
  ```
  grep "services/bff/gen" .gitignore
  ```
- `just generate` で buf generate と sqlc generate が一括実行される:
  ```
  just generate
  echo $?
  ```
  が `0` を返す
- 生成後、`git status` で生成コードが追跡対象外であることを確認:
  ```
  git status gen/ services/bff/gen/ services/*/db/generated/
  ```
  で untracked / modified ファイルが表示されない

---

## Task 16: 最終検証 — Phase 3 受け入れ条件の確認

### やること

Phase 3 全体の受け入れ条件を一括で検証し、全てのゴールが達成されていることを確認する。

### 方針

- phase.md に記載された Phase 3 の受け入れ条件を全て実行する
- 失敗する項目があれば該当タスクに戻って修正する
- 全検証が通れば Phase 3 完了とする

### スコープ

- Phase 3 受け入れ条件の一括検証

### スコープ外

- Phase 4 以降の作業

### 受け入れ条件

- `buf lint` がエラーなく通過する:
  ```
  buf lint
  echo $?
  ```
  が `0` を返す
- `buf generate` が成功する:
  ```
  buf generate
  echo $?
  ```
  が `0` を返す
- Go の生成コード（Connect RPC）が存在する:
  ```
  ls gen/go/product/v1/*connect*.go
  ls gen/go/customer/v1/*connect*.go
  ls gen/go/ec/v1/*connect*.go
  ```
- TypeScript の生成コード（Connect RPC）が存在する:
  ```
  ls services/bff/gen/product/v1/*connect*.ts
  ls services/bff/gen/customer/v1/*connect*.ts
  ls services/bff/gen/ec/v1/*connect*.ts
  ```
- `sqlc generate` が全サービスで成功する:
  ```
  cd services/product-mgmt && sqlc generate && echo "OK"
  cd services/customer-mgmt && sqlc generate && echo "OK"
  cd services/ec-site && sqlc generate && echo "OK"
  ```
- 生成された Go コードが `go build ./...` でビルドできる:
  ```
  cd gen/go && go build ./... && echo "OK"
  cd services/product-mgmt && go build ./... && echo "OK"
  cd services/customer-mgmt && go build ./... && echo "OK"
  cd services/ec-site && go build ./... && echo "OK"
  ```
- 生成された TypeScript コードが型チェックを通過する:
  ```
  cd services/bff && pnpm exec tsc --noEmit && echo "OK"
  ```
- `just generate` で全コード生成が一括実行できる:
  ```
  just generate
  echo $?
  ```
  が `0` を返す

---

## タスク依存関係

```
Task 1: ProductService スキーマ定義 ─┐
Task 2: CustomerService スキーマ定義 ─┼─► Task 4: buf lint 一括検証
Task 3: CartService スキーマ定義 ─────┘
                                           │
                                           └─► Task 5: buf.gen.yaml 更新
                                                 │
                                                 └─► Task 6: Go 生成コード用モジュール作成
                                                       │
                                                       └─► Task 7: buf generate 実行
                                                             │
                                                             ├─► Task 8: Go サービス依存関係設定
                                                             │     │
                                                             │     └─► Task 10: Go ビルド検証
                                                             │
                                                             └─► Task 9: BFF TypeScript 依存追加

Task 11: 商品管理 sqlc クエリ定義 ─┐
Task 12: 顧客管理 sqlc クエリ定義 ─┼─► Task 14: sqlc generate 実行
Task 13: ECサイト sqlc クエリ定義 ─┘

Task 10 + Task 9 + Task 14 ─► Task 15: .gitignore 管理方針確定
                                         │
                                         └─► Task 16: 最終検証
```

- Task 1, 2, 3 は proto 定義タスクで互いに独立（並行実施可能）
- Task 4 は全 proto 定義完了後の一括検証
- Task 5 → 6 → 7 はコード生成パイプラインの順序依存
- Task 11, 12, 13 は sqlc クエリ定義で互いに独立（並行実施可能、Task 7 とも独立）
- Task 14 は全クエリ定義完了後に一括生成
- Task 15 は全生成完了後に .gitignore を確定
- Task 16 は全タスク完了後の最終検証
