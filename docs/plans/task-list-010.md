# Phase 10: 結合確認 — タスクリスト

## Task 1: 全サービスの一括起動確認

### やること

`just dev` で全サービス（BFF、商品管理、顧客管理、ECサイト、フロントエンド）が同時に起動できることを確認する。

### 方針

- `just setup && docker compose up -d && just db-migrate && just dev` を順に実行する
- 各サービスのヘルスチェックが応答することを確認する:
  - BFF: `curl http://localhost:3000/health`
  - ECサイト: `curl http://localhost:8080/health`
  - 商品管理: `curl http://localhost:8081/health`
  - 顧客管理: `curl http://localhost:8082/health`
  - フロントエンド: `http://localhost:5173` がブラウザで表示される
- 起動に失敗するサービスがあれば原因を調査し修正する
- Justfile の `dev` タスクにフロントエンドが含まれていることを確認する（Phase 9 で追加済みのはず）

### スコープ

- 全サービスの一括起動確認
- 起動時の問題の修正

### スコープ外

- 新機能の実装
- E2E ビジネスシナリオ

### 受け入れ条件

- `just setup && docker compose up -d && just db-migrate && just dev` で全サービスが起動する
- 全サービスのヘルスチェックが 200 を返す
- フロントエンドが `http://localhost:5173` で表示される

---

## Task 2: 疎通シナリオの実行（商品作成 → フロントエンド表示）

### やること

フロントエンド → BFF → バックエンド → DB の全レイヤーの疎通を一括確認する。

### 方針

- 以下の手順を順に実行する:
  1. buf curl で商品管理サービスに商品を作成する:
     ```
     buf curl --data '{"name":"テスト商品","description":"結合確認用","price":1500,"stock_quantity":20}' \
       http://localhost:8081/product.v1.ProductService/CreateProduct
     ```
  2. buf curl で商品が取得できることを確認する:
     ```
     buf curl --data '{"id":1}' \
       http://localhost:8081/product.v1.ProductService/GetProduct
     ```
  3. BFF 経由で商品が取得できることを確認する:
     ```
     curl http://localhost:3000/product.v1.ProductService/GetProduct \
       -H 'Content-Type: application/json' \
       -d '{"id":1}'
     ```
  4. フロントエンド（`http://localhost:5173`）で商品情報が表示されることを確認する
- 失敗するステップがあれば原因を調査し修正する

### スコープ

- 商品作成 → BFF 経由取得 → フロントエンド表示の疎通確認

### スコープ外

- カート操作の E2E シナリオ
- 注文の E2E シナリオ
- 顧客作成の E2E シナリオ

### 受け入れ条件

- buf curl で商品管理に商品を作成できる
- BFF 経由で商品情報が取得できる
- フロントエンドで BFF 経由の商品情報が表示される

---

## Task 3: コード生成の一括実行確認

### やること

`just generate` で Protobuf / sqlc のコード生成が一括実行できることを確認する。

### 方針

- `just generate` を実行する
- 以下が成功することを確認する:
  - `buf generate` による Protobuf コード生成（Go + TypeScript）
  - `sqlc generate` による Go クエリコード生成（3 サービス分）
- 生成後、各サービスのビルド / 型チェックが通ることを確認する:
  - `cd services/product-mgmt && go build ./...`
  - `cd services/customer-mgmt && go build ./...`
  - `cd services/ec-site && go build ./...`
  - `cd services/bff && pnpm exec tsc --noEmit`
- 問題があれば Justfile の `generate` タスクを修正する

### スコープ

- `just generate` の動作確認
- 生成後のビルド / 型チェック確認

### スコープ外

- 新しい `.proto` ファイルや sqlc クエリの追加
- lefthook フックの本設定

### 受け入れ条件

- `just generate` がエラーなく完了する
- 生成後、全サービスの `go build ./...` と `pnpm exec tsc --noEmit` が成功する

---

## Task 4: 最終検証 — Phase 10 受け入れ条件の確認

### やること

Phase 10 全体の受け入れ条件を一括で検証し、全てのゴールが達成されていることを確認する。全フェーズの結合確認完了をもって初期実装（疎通確認）の完了とする。

### 方針

- phase.md に記載された Phase 10 の受け入れ条件を全て実行する
- 失敗する項目があれば該当タスクに戻って修正する
- 全検証が通れば Phase 10（および全フェーズ）完了とする

### スコープ

- Phase 10 受け入れ条件の一括検証

### スコープ外

- 積み残し機能の実装

### 受け入れ条件

- `just setup && docker compose up -d && just db-migrate && just dev` で全サービスが起動する
- 以下の疎通シナリオが手動で完走する:
  1. buf curl で商品管理に商品を作成
  2. フロントエンドで BFF 経由で商品情報が表示される
- `just generate` で Protobuf / sqlc のコード生成が一括実行できる

---

## タスク依存関係

```
Task 1: 全サービス一括起動確認 ─┐
                                ├─► Task 4: 最終検証
Task 2: 疎通シナリオ実行 ───────┤
                                │
Task 3: コード生成一括確認 ─────┘
```

- Task 1 は全フェーズ（Phase 1〜9）の完了が前提
- Task 2 は Task 1 の完了が前提（全サービス起動後に疎通確認）
- Task 3 は Task 1 と並行実施可能
- Task 4 は全タスク完了後の最終検証
