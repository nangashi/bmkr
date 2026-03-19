# Phase 8: BFF の最小実装 — タスクリスト

## Task 1: BFF から商品管理サービスへの Connect RPC プロキシ実装

### やること

Fastify サーバーに Connect RPC クライアントを組み込み、BFF 経由で商品管理サービスの GetProduct を呼び出せるエンドポイントを実装する。

### 方針

- `services/bff/src/index.ts` に Connect RPC クライアントの初期化とプロキシエンドポイントを追加する
- `@connectrpc/connect-node` の `createConnectTransport` で商品管理サービスへのトランスポートを作成する
  - 接続先: `http://localhost:8081`（環境変数 `PRODUCT_SERVICE_URL` でオーバーライド可能）
- `@connectrpc/connect` の `createClient` で `ProductService` クライアントを生成する
  - 生成コードの `ProductService`（`gen/product/v1/product_connect.ts`）を使用
- Fastify のルートとして Connect RPC プロキシを実装する
  - 方法の候補:
    - A: Fastify ルートで JSON リクエストを受け取り、Connect RPC クライアントで商品管理を呼び出してレスポンスを返す REST スタイルのプロキシ
    - B: `@connectrpc/connect-node` の `connectNodeAdapter` を使って Connect RPC ハンドラとして直接マウントする
  - 方法 B を優先する（phase.md の受け入れ条件が Connect RPC 形式のリクエストを期待しているため）
- `connectNodeAdapter` を使用して Fastify に Connect RPC ハンドラをマウントする
  - BFF 側の `ProductService` ハンドラ内で、商品管理サービスの Connect RPC クライアントを呼び出す

### スコープ

- `services/bff/src/index.ts` の修正（Connect RPC クライアント + プロキシエンドポイント）
- BFF → 商品管理への 1 エンドポイント（GetProduct）のプロキシ

### スコープ外

- エンドユーザー認証（@fastify/jwt、HttpOnly Cookie）
- BFF → バックエンド間認証（OAuth 2.0 Client Credentials）
- API 集約ロジック（複数サービスのデータ統合）
- CORS / Cookie 設定
- CreateProduct のプロキシ

### 受け入れ条件

- `pnpm exec tsc --noEmit` がエラーなく完了する
- curl で BFF 経由で商品情報が返る:
  ```
  curl http://localhost:3000/product.v1.ProductService/GetProduct \
    -H 'Content-Type: application/json' \
    -d '{"id":1}'
  ```
  が商品情報を返す（事前に商品管理サービスで商品が作成済みであること）

---

## Task 2: 最終検証 — Phase 8 受け入れ条件の確認

### やること

Phase 8 全体の受け入れ条件を一括で検証し、全てのゴールが達成されていることを確認する。

### 方針

- phase.md に記載された Phase 8 の受け入れ条件を全て実行する
- 前提: `just db-up && just db-migrate` で DB が起動・マイグレーション済みであること
- 前提: 商品管理サービス（`:8081`）と BFF（`:3000`）が起動していること
- 前提: 商品管理サービスに商品が 1 件以上作成済みであること
- 失敗する項目があれば Task 1 に戻って修正する
- 全検証が通れば Phase 8 完了とする

### スコープ

- Phase 8 受け入れ条件の一括検証

### スコープ外

- Phase 9 以降の作業

### 受け入れ条件

- `pnpm exec tsc --noEmit` がエラーなく完了する:
  ```
  cd services/bff && pnpm exec tsc --noEmit && echo "OK"
  ```
- curl で BFF 経由で商品情報が返る:
  ```
  curl http://localhost:3000/product.v1.ProductService/GetProduct \
    -H 'Content-Type: application/json' \
    -d '{"id":1}'
  ```
  が商品情報の JSON を返す
- ヘルスチェックが引き続き応答する:
  ```
  curl http://localhost:3000/health
  ```
  が 200 を返す

---

## タスク依存関係

```
Task 1: Connect RPC プロキシ実装
  │
  ▼
Task 2: 最終検証
```

- Task 1 は Phase 5（商品管理サービスの実装）の完了が前提
- Task 2 は Task 1 完了後の最終検証
