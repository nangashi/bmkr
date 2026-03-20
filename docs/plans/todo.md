# TODO: 積み残し機能一覧

Phase 1〜10（疎通確認）は完了済み。以下は機能単位で整理した積み残しタスク。

## 実装済みの現状

| レイヤー | 実装済み |
|---|---|
| product-mgmt | CreateProduct, GetProduct（proto / sqlc / handler） |
| customer-mgmt | CreateCustomer, GetCustomer（proto / sqlc / handler） |
| ec-site backend | GetCart（proto / sqlc / handler）、product-mgmt への Connect RPC 呼び出し（ログのみ） |
| BFF | GetProduct プロキシ 1 エンドポイントのみ |
| frontend | 商品 ID 指定の単品取得ページ 1 つのみ（ルーティングなし） |
| infra | PostgreSQL x1（DB 3 つ）、Ory Hydra（in-memory）、lefthook（fmt-check / gitleaks / lint） |

---

## TODO 1: 商品一覧・詳細ページ（エンドユーザー向け）

エンドユーザーが商品を閲覧できる機能。現状は ID 指定の単品取得のみ。

- [ ] proto: `ProductService.ListProducts` RPC 追加
- [ ] sqlc: `ListProducts` クエリ追加
- [ ] product-mgmt: `ListProducts` ハンドラ実装
- [ ] BFF: `ListProducts` / `GetProduct` プロキシエンドポイント追加
- [ ] frontend: React Router 導入
- [ ] frontend: 商品一覧ページ（ListProducts 呼び出し）
- [ ] frontend: 商品詳細ページ（GetProduct 呼び出し）
- [ ] 商品シードデータ作成

**依存:** なし（最初に着手可能）

---

## TODO 2: カート機能

エンドユーザーがカートに商品を追加・編集・確認できる機能。

- [ ] proto: `CartService.AddItem` RPC 追加
- [ ] proto: `CartService.RemoveItem` RPC 追加
- [ ] proto: `CartService.UpdateQuantity` RPC 追加
- [ ] sqlc: `AddCartItem` / `RemoveCartItem` / `UpdateCartItemQuantity` / `GetCartItem` クエリ追加
- [ ] ec-site: `AddItem` / `RemoveItem` / `UpdateQuantity` ハンドラ実装
- [ ] BFF: CartService プロキシエンドポイント追加（GetCart / AddItem / RemoveItem / UpdateQuantity）
- [ ] frontend: カートページ（カート内容表示、数量変更、削除）
- [ ] frontend: 商品詳細ページに「カートに追加」ボタン

**依存:** TODO 1（商品一覧・詳細ページ）

---

## TODO 3: 注文機能

エンドユーザーがカートから注文を確定し、注文履歴を確認できる機能。

- [ ] proto: `OrderService.PlaceOrder` RPC 追加（メッセージ定義含む）
- [ ] proto: `OrderService.ListOrders` RPC 追加（メッセージ定義含む）
- [ ] proto: `ProductService.AllocateStock` RPC 追加（在庫引き当て）
- [ ] sqlc: `CreateOrder` / `CreateOrderItem` / `ListOrdersByCustomerID` / `GetOrder` クエリ追加
- [ ] sqlc: `AllocateStock`（product-mgmt 側）クエリ追加
- [ ] product-mgmt: `AllocateStock` ハンドラ実装
- [ ] ec-site: `PlaceOrder` ハンドラ実装（product-mgmt AllocateStock 呼び出し含む）
- [ ] ec-site: `ListOrders` ハンドラ実装
- [ ] ec-site → customer-mgmt: Connect RPC クライアント追加（顧客情報参照）
- [ ] BFF: OrderService プロキシエンドポイント追加（PlaceOrder / ListOrders）
- [ ] frontend: 注文確定ページ（カート → 注文）
- [ ] frontend: 注文履歴ページ

**依存:** TODO 2（カート機能）

---

## TODO 4: エンドユーザー認証

ログイン・ログアウトと認証状態の管理。

- [ ] proto: `CustomerService.GetCustomerByEmail` RPC 追加
- [ ] sqlc: `GetCustomerByEmail` クエリ追加
- [ ] customer-mgmt: `GetCustomerByEmail` ハンドラ実装
- [ ] docker-compose: Valkey コンテナ追加（ADR-0014）
- [ ] BFF: `@fastify/jwt` + `@fastify/cookie` 導入
- [ ] BFF: ログイン API（customer-mgmt 呼び出し → JWT 発行 → HttpOnly Cookie）
- [ ] BFF: リフレッシュトークン発行・ローテーション・Valkey ブラックリスト
- [ ] BFF: ログアウト API（Cookie 削除 + リフレッシュトークン無効化）
- [ ] BFF: 認証ミドルウェア（JWT 検証、保護エンドポイントへの適用）
- [ ] BFF: CORS 設定（`@fastify/cors`）
- [ ] frontend: ログインページ
- [ ] frontend: ログアウト機能
- [ ] frontend: 認証状態管理（Cookie ベース）
- [ ] frontend: 未認証時のリダイレクト
- [ ] テスト顧客シードデータ

**依存:** なし（TODO 1 と並行着手可能だが、TODO 2・3 の前に完了が望ましい）

---

## TODO 5: サービス間認証

バックエンドサービス間の OAuth 2.0 Client Credentials 認証。

- [ ] ec-site: Connect RPC インターセプタ — Client Credentials トークン取得・キャッシュ（Ory Hydra）
- [ ] ec-site: 送信リクエストに `Authorization` ヘッダ付与
- [ ] product-mgmt: Connect RPC インターセプタ — サービストークン検証
- [ ] customer-mgmt: Connect RPC インターセプタ — サービストークン検証
- [ ] BFF → バックエンド: Client Credentials トークン取得・付与
- [ ] ADR-0013 準拠: デュアルヘッダ（`Authorization` = サービストークン、`X-User-Token` = ユーザー JWT）の実装

**依存:** TODO 4（エンドユーザー認証）— ユーザー JWT が存在してデュアルヘッダが成立する

---

## TODO 6: 商品管理画面

管理者が商品を CRUD 操作する管理画面。

- [ ] proto: `ProductService.UpdateProduct` RPC 追加
- [ ] proto: `ProductService.DeleteProduct` RPC 追加
- [ ] sqlc: `UpdateProduct` / `DeleteProduct` クエリ追加
- [ ] product-mgmt: `UpdateProduct` / `DeleteProduct` ハンドラ実装
- [ ] product-mgmt: 管理者認証（echo-jwt / golang-jwt）— ログイン画面、JWT 発行、Cookie 格納
- [ ] product-mgmt: 管理画面 UI（templ + HTMX）— 商品一覧
- [ ] product-mgmt: 管理画面 UI（templ + HTMX）— 商品登録フォーム
- [ ] product-mgmt: 管理画面 UI（templ + HTMX）— 商品編集フォーム
- [ ] product-mgmt: 管理画面 UI（templ + HTMX）— 商品削除
- [ ] product-mgmt: 管理者ユーザーシードデータ（admin_users テーブルは定義済み）

**依存:** TODO 1（ListProducts が共通基盤）

---

## TODO 7: 顧客管理画面

管理者が顧客情報を閲覧する管理画面。

- [ ] proto: `CustomerService.ListCustomers` RPC 追加
- [ ] sqlc: `ListCustomers` クエリ追加
- [ ] customer-mgmt: `ListCustomers` ハンドラ実装
- [ ] customer-mgmt: 管理者認証（echo-jwt / golang-jwt）— ログイン画面、JWT 発行、Cookie 格納
- [ ] customer-mgmt: 管理画面 UI（templ + HTMX）— 顧客一覧
- [ ] customer-mgmt: 管理画面 UI（templ + HTMX）— 顧客詳細
- [ ] customer-mgmt: 管理画面 UI — 注文履歴表示（ec-site API 連携）
- [ ] customer-mgmt: 管理者ユーザーシードデータ（admin_users テーブルは定義済み）

**依存:** TODO 3（注文履歴表示に注文機能が必要）、TODO 6 と管理者認証の実装パターンを共有

---

## TODO 8: 結合・品質整備

- [ ] E2E ビジネスシナリオ（商品登録 → カート → 注文 → 在庫減少 → 注文履歴）
- [ ] BFF: API 集約ロジック（商品一覧 + 在庫情報統合等）
- [ ] CLAUDE.md 更新（開発手順、コマンド一覧）

**依存:** TODO 1〜7 の主要機能が揃った後

---

## 推奨実装順序

```
TODO 1  商品一覧・詳細           TODO 4  エンドユーザー認証
  |                                |
  v                                v
TODO 2  カート機能              TODO 5  サービス間認証
  |                                |
  v                                |
TODO 3  注文機能  <----------------+
  |
  +---> TODO 6  商品管理画面
  |
  +---> TODO 7  顧客管理画面
  |
  v
TODO 8  結合・品質整備
```

- TODO 1 と TODO 4 は並行着手可能
- TODO 6 は TODO 1 完了後いつでも着手可能（TODO 3 を待たなくてよい）
- TODO 7 は注文履歴表示があるため TODO 3 の後が望ましい
