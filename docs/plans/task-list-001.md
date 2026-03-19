# Phase 1: 開発環境とモノレポ基盤 — タスクリスト

## Task 1: mise 設定ファイルの作成

### やること

`.mise.toml` を作成し、プロジェクトで使用する全ツールのバージョンを固定する。

### 方針

- プロジェクトルートに `.mise.toml` を配置する
- Node.js / Go / pnpm / buf / sqlc / goose / just / lefthook の 8 ツールをバージョン指定する
- バージョンは各ツールの最新安定版を採用する
- `mise install` 一発で全ツールが揃う状態にする

### スコープ

- `.mise.toml` の作成とツールバージョンの記載

### スコープ外

- 各ツールの設定ファイル（Justfile、lefthook.yml 等）の作成
- pnpm workspace やパッケージの初期化

### 受け入れ条件

- `mise install` がエラーなく完了する
- 以下のコマンドが全てバージョンを返す:
  ```
  mise exec -- node --version
  mise exec -- go version
  mise exec -- pnpm --version
  mise exec -- buf --version
  mise exec -- sqlc version
  mise exec -- goose --version
  mise exec -- just --version
  mise exec -- lefthook version
  ```

---

## Task 2: .gitignore の整備

### やること

モノレポ全体に必要な `.gitignore` ルールを整備する。

### 方針

- 既存の `.gitignore`（`.output/` のみ）に追記する
- Node.js 関連（`node_modules/`）、Go 関連（バイナリ）、コード生成出力（`gen/`）、IDE、OS ファイルを対象とする
- 各サービス個別の `.gitignore` は作らず、ルートで一括管理する

### スコープ

- ルート `.gitignore` の更新

### スコープ外

- サービス個別の `.gitignore`
- `.gitattributes` などの他の git 設定ファイル

### 受け入れ条件

- `cat .gitignore` で以下のカテゴリのルールが含まれている:
  - `.output/`
  - `node_modules/`
  - Go バイナリ（`/services/*/tmp/`）
  - コード生成出力（`gen/`）
  - IDE ファイル（`.idea/`、`.vscode/` 等）
  - OS ファイル（`.DS_Store` 等）
- `git status` で `node_modules/` や `gen/` が追跡対象外であること（後続タスクで生成されるため、この時点では確認不要）

---

## Task 3: pnpm ワークスペースの初期化

### やること

モノレポのルートに `package.json` と `pnpm-workspace.yaml` を作成し、pnpm ワークスペースを構成する。

### 方針

- ルート `package.json` は `private: true` とし、ワークスペースルートとして扱う
- `pnpm-workspace.yaml` で `services/bff` と `services/ec-site/frontend`（将来の React SPA）をワークスペースとして定義する
- Go サービスは pnpm ワークスペースに含めない（Go Modules で管理）
- ルートの devDependencies にはこの時点では何も追加しない

### スコープ

- ルート `package.json` の作成
- `pnpm-workspace.yaml` の作成
- `pnpm install` の実行確認

### スコープ外

- 各ワークスペースパッケージの `package.json` 作成（後続タスクで実施）
- 依存パッケージのインストール

### 受け入れ条件

- `pnpm install` がエラーなく完了する（ワークスペースが空でも OK）
- `cat pnpm-workspace.yaml` でワークスペースパッケージのパスが定義されている
- `cat package.json` で `"private": true` が設定されている

---

## Task 4: BFF パッケージの初期化

### やること

`services/bff/` に TypeScript パッケージとして `package.json` を作成し、pnpm ワークスペースに認識させる。

### 方針

- `services/bff/package.json` を作成する
- パッケージ名は `@bmkr/bff` とする
- この時点では依存パッケージを追加しない（Fastify 等は後続フェーズで追加）
- `tsconfig.json` のひな型を配置する

### スコープ

- `services/bff/package.json` の作成
- `services/bff/tsconfig.json` の作成
- pnpm ワークスペースでの認識確認

### スコープ外

- Fastify や Connect RPC 等の依存パッケージのインストール
- TypeScript ソースコードの作成
- ビルド設定

### 受け入れ条件

- `pnpm ls --filter @bmkr/bff` がパッケージ情報を表示する
- `cat services/bff/package.json` でパッケージ名 `@bmkr/bff` が確認できる
- `cat services/bff/tsconfig.json` が存在する

---

## Task 5: ECサイト Go モジュールの初期化

### やること

`services/ec-site/` に Go モジュールを作成する。

### 方針

- モジュールパスは `github.com/nangashi/bmkr/services/ec-site` とする
- Go のバージョンは mise で固定したバージョンに合わせる
- `main.go` は作成しない（Phase 4 で実施）
- `.keep` ファイルを配置してディレクトリが git に追跡されるようにする

### スコープ

- `services/ec-site/go.mod` の作成

### スコープ外

- Go ソースコードの作成
- 依存パッケージの追加（Echo, Connect RPC 等）

### 受け入れ条件

- `cat services/ec-site/go.mod` でモジュールパス `github.com/nangashi/bmkr/services/ec-site` が確認できる
- `cd services/ec-site && go mod tidy` がエラーなく完了する

---

## Task 6: 商品管理 Go モジュールの初期化

### やること

`services/product-mgmt/` に Go モジュールを作成する。

### 方針

- モジュールパスは `github.com/nangashi/bmkr/services/product-mgmt` とする
- Go のバージョンは mise で固定したバージョンに合わせる
- `.keep` ファイルを配置してディレクトリが git に追跡されるようにする

### スコープ

- `services/product-mgmt/go.mod` の作成

### スコープ外

- Go ソースコードの作成
- 依存パッケージの追加

### 受け入れ条件

- `cat services/product-mgmt/go.mod` でモジュールパス `github.com/nangashi/bmkr/services/product-mgmt` が確認できる
- `cd services/product-mgmt && go mod tidy` がエラーなく完了する

---

## Task 7: 顧客管理 Go モジュールの初期化

### やること

`services/customer-mgmt/` に Go モジュールを作成する。

### 方針

- モジュールパスは `github.com/nangashi/bmkr/services/customer-mgmt` とする
- Go のバージョンは mise で固定したバージョンに合わせる
- `.keep` ファイルを配置してディレクトリが git に追跡されるようにする

### スコープ

- `services/customer-mgmt/go.mod` の作成

### スコープ外

- Go ソースコードの作成
- 依存パッケージの追加

### 受け入れ条件

- `cat services/customer-mgmt/go.mod` でモジュールパス `github.com/nangashi/bmkr/services/customer-mgmt` が確認できる
- `cd services/customer-mgmt && go mod tidy` がエラーなく完了する

---

## Task 8: Protobuf ディレクトリとスタブスキーマの作成

### やること

`proto/` ディレクトリに Protobuf のパッケージ構成を作成し、各サービスの `.proto` ファイルのスタブを配置する。

### 方針

- `proto/` 配下にサービスごとのディレクトリを作成する（`proto/product/v1/`、`proto/customer/v1/`、`proto/ec/v1/`）
- 各 `.proto` ファイルには `syntax`、`package`、`option go_package` のみを記載する（RPC 定義は Phase 3）
- バージョニングは `v1` パッケージとする

### スコープ

- `proto/` ディレクトリ構成の作成
- 各サービスのスタブ `.proto` ファイル作成（空の service 定義のみ）

### スコープ外

- RPC メソッドやメッセージ型の定義（Phase 3）
- コード生成の実行

### 受け入れ条件

- 以下のファイルが存在する:
  ```
  ls proto/product/v1/product.proto
  ls proto/customer/v1/customer.proto
  ls proto/ec/v1/cart.proto
  ls proto/ec/v1/order.proto
  ```
- 各 `.proto` ファイルに `syntax = "proto3";` と `package` 宣言が含まれている:
  ```
  head -3 proto/product/v1/product.proto
  ```

---

## Task 9: buf CLI の設定

### やること

`buf.yaml` と `buf.gen.yaml` を作成し、Protobuf のリント・コード生成パイプラインを構成する。

### 方針

- `buf.yaml` はプロジェクトルートに配置し、`proto/` をソースディレクトリとして指定する
- `buf.gen.yaml` で Connect RPC の Go / TypeScript プラグインを設定する
  - Go: `protoc-gen-go` + `protoc-gen-connect-go`
  - TypeScript: `@connectrpc/protoc-gen-connect-es` + `@bufbuild/protoc-gen-es`
- 生成先は各サービスの `gen/` ディレクトリとする
- `buf lint` のルールは `DEFAULT` を採用する

### スコープ

- `buf.yaml` の作成
- `buf.gen.yaml` の作成
- `buf build` と `buf lint` の実行確認

### スコープ外

- `buf generate` の実行（プラグインの npm / go install は Phase 3 で本格的に行う）
- BSR (Buf Schema Registry) への公開

### 受け入れ条件

- `buf build` がエラーなく完了する
- `buf lint` がエラーなく完了する（スタブ .proto に対して）
- `cat buf.yaml` で lint ルールとモジュール設定が確認できる
- `cat buf.gen.yaml` で Go / TypeScript の生成プラグインが設定されている

---

## Task 10: Justfile の導入

### やること

プロジェクトルートに `Justfile` を作成し、開発で頻繁に使うコマンドをタスクとしてまとめる。

### 方針

- Phase 1 時点で必要なタスクのみ定義する（後続フェーズで随時追加）
- `just setup`: `mise install` + `pnpm install` を実行
- `just lint`: `buf lint` を実行（将来的に他の lint も追加）
- `just generate`: `buf generate` を実行（Phase 3 で本格化するが枠だけ用意）
- タスク名は英語、コメントは日本語で記載する

### スコープ

- `Justfile` の作成
- `setup`、`lint`、`generate` タスクの定義

### スコープ外

- `dev`（サーバー起動）、`db-migrate`、`test` 等のタスク（後続フェーズで追加）

### 受け入れ条件

- `just --list` でタスク一覧が表示される
- `just setup` が `mise install` と `pnpm install` を順に実行して正常完了する
- `just lint` が `buf lint` を実行して正常完了する

---

## Task 11: lefthook の導入

### やること

lefthook を設定し、コミット時に自動でリントが実行されるようにする。

### 方針

- `lefthook.yml` をプロジェクトルートに配置する
- `pre-commit` フックで `buf lint` を実行する
- Phase 1 時点ではリント対象が少ないため、最小限のフック設定とする
- 後続フェーズで TypeScript / Go の lint を追加する

### スコープ

- `lefthook.yml` の作成
- `pre-commit` フックの設定
- `lefthook install` の実行

### スコープ外

- TypeScript の ESLint / Biome 設定
- Go の golangci-lint 設定
- `pre-push`、`commit-msg` 等の他のフック

### 受け入れ条件

- `lefthook install` が成功する
- `ls .git/hooks/pre-commit` でフックファイルが存在する
- `lefthook run pre-commit` が `buf lint` を実行して正常完了する

---

## タスク依存関係

```
Task 1: mise 設定
  │
  ├─► Task 2: .gitignore
  │
  ├─► Task 3: pnpm ワークスペース初期化
  │     │
  │     └─► Task 4: BFF パッケージ初期化
  │
  ├─► Task 5: ECサイト Go モジュール
  ├─► Task 6: 商品管理 Go モジュール
  ├─► Task 7: 顧客管理 Go モジュール
  │
  └─► Task 8: Protobuf スタブスキーマ
        │
        └─► Task 9: buf CLI 設定
              │
              ├─► Task 10: Justfile
              │
              └─► Task 11: lefthook
```

- Task 1 が全タスクの前提（ツールが必要）
- Task 3 → 4 は pnpm ワークスペースの依存
- Task 8 → 9 は .proto ファイルが buf 設定の前提
- Task 5, 6, 7 は互いに独立
- Task 10, 11 は他のタスク完了後が望ましい（lint 対象が揃うため）
