# Phase 9: フロントエンドの最小実装 — タスクリスト

## Task 1: Vite + React + TypeScript プロジェクトのセットアップ

### やること

`services/ec-site/frontend` に Vite + React + TypeScript プロジェクトを作成し、開発サーバーが起動する状態にする。

### 方針

- `services/ec-site/frontend` ディレクトリに Vite プロジェクトを初期化する
  - `pnpm create vite` または手動で `package.json` / `vite.config.ts` を作成
  - テンプレート: `react-ts`
- `package.json` に以下の依存を追加する:
  - `react`, `react-dom`
  - `@bufbuild/protobuf` — Protobuf ランタイム
  - `@connectrpc/connect` — Connect RPC コア
  - `@connectrpc/connect-web` — Connect RPC ブラウザ向けトランスポート
- `tsconfig.json` を適切に設定する
- `vite.config.ts` に BFF へのプロキシ設定を追加する:
  - `/product.v1.ProductService/*` → `http://localhost:3000` へプロキシ
  - 開発時の CORS 問題を回避する
- `pnpm-workspace.yaml` は `services/ec-site/frontend` が既に登録済み
- Justfile の `dev` タスクにフロントエンドの起動を追加する

### スコープ

- Vite + React + TypeScript プロジェクトの初期化
- 依存パッケージのインストール
- vite.config.ts のプロキシ設定
- Justfile の `dev` タスク更新

### スコープ外

- Connect RPC クライアントによるデータ取得
- 商品情報の表示

### 受け入れ条件

- `http://localhost:5173` でフロントエンドの初期ページが表示される
- `cd services/ec-site/frontend && pnpm exec tsc --noEmit` がエラーなく完了する
- `pnpm install` がワークスペース全体でエラーなく完了する

---

## Task 2: 商品情報を表示する最小ページの実装

### やること

BFF 経由で商品管理サービスから商品情報を取得し、画面に表示する最小ページを実装する。

### 方針

- Connect RPC クライアントのセットアップ:
  - `@connectrpc/connect-web` の `createConnectTransport` でトランスポートを作成
  - ベース URL は Vite のプロキシ経由で `/` を指定（開発時）
  - `createClient` で `ProductService` クライアントを生成
- BFF 用の生成コードを参照する:
  - `services/bff/gen/product/v1/product_connect.ts` と `product_pb.ts` を利用する
  - 方法の候補:
    - A: BFF の `gen/` を直接参照する（ワークスペースの `@bmkr/bff` パッケージから import）
    - B: フロントエンド用に別途 `buf generate` の出力先を設定する
    - C: `gen/` をフロントエンドの `src/` 配下にコピーまたはシンボリックリンクする
  - 方法 A を優先する（モノレポのワークスペース機能を活用）
- 最小ページの実装:
  - `App.tsx` に商品情報取得・表示のロジックを実装
  - 商品 ID をハードコードまたは入力フィールドで指定
  - 取得した商品の名前・価格・在庫数を表示
  - スタイリングは最小限（CSS フレームワークなし）

### スコープ

- Connect RPC クライアントのセットアップ
- 商品情報を表示する 1 ページ

### スコープ外

- ログイン / ログアウト画面
- 商品一覧ページ
- カートページ
- 注文確定・履歴ページ
- 認証状態管理
- CSS フレームワーク / UIライブラリ

### 受け入れ条件

- `http://localhost:5173` で商品情報が画面に表示される
  - 前提: 商品管理サービスに商品が 1 件以上作成済み、BFF が起動中
- BFF 経由で商品管理サービスの GetProduct が呼び出され、レスポンスが画面に描画される
- `pnpm exec tsc --noEmit` がエラーなく完了する

---

## Task 3: 最終検証 — Phase 9 受け入れ条件の確認

### やること

Phase 9 全体の受け入れ条件を一括で検証し、全てのゴールが達成されていることを確認する。

### 方針

- phase.md に記載された Phase 9 の受け入れ条件を全て実行する
- 前提: `just db-up && just db-migrate` で DB が起動・マイグレーション済みであること
- 前提: 全バックエンドサービスと BFF が起動していること
- 前提: 商品管理サービスに商品が 1 件以上作成済みであること
- 失敗する項目があれば該当タスクに戻って修正する
- 全検証が通れば Phase 9 完了とする

### スコープ

- Phase 9 受け入れ条件の一括検証

### スコープ外

- Phase 10 以降の作業

### 受け入れ条件

- `http://localhost:5173` でフロントエンドが表示される
- BFF 経由で商品情報が画面に表示される

---

## タスク依存関係

```
Task 1: Vite + React プロジェクトセットアップ
  │
  ▼
Task 2: 商品情報表示ページ実装
  │
  ▼
Task 3: 最終検証
```

- Task 1 でプロジェクトの骨格を作成した後、Task 2 でデータ取得・表示を実装する
- Task 2 は Phase 8（BFF の実装）の完了が前提
- Task 3 は全タスク完了後の最終検証
