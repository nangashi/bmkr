---
id: go-logging
category: observability
scope: [backend]
severity: high
detectable_by_linter: false
---

# ログ設計ガイド

本プロジェクトのGoサービス（ec-site, product-mgmt, customer-mgmt）におけるログ出力の方針。

## ログレベルの判断基準

| レベル | 用途 | 例 |
|--------|------|-----|
| **ERROR** | 処理が完了できない障害。対応が必要 | DB 接続失敗、必須の外部サービス応答不能 |
| **WARN** | 処理は継続できるが想定外の状態。放置すると問題化しうる | リトライで回復した外部 API 呼び出し、非推奨パラメータの使用 |
| **INFO** | 正常系の重要なイベント。本番環境で常時出力 | サービス起動、canonical log line |
| **DEBUG** | 開発・調査時のみ必要な詳細情報。本番では無効化 | リクエスト/レスポンスのペイロード詳細、中間計算結果 |

### 判断フローチャート

1. 処理は続行できるか？ → No → **ERROR**
2. リトライや代替手段で回復したか？ → Yes → **WARN**
3. 運用者が常に知るべき事象か？ → Yes → **INFO**
4. 上記いずれでもない → **DEBUG**

### よくある判断例

- 外部サービス呼び出し失敗（リトライなし）→ そのリクエストが失敗するなら **ERROR**、スキップして続行するなら **WARN**
- カート取得時に商品情報の取得に失敗 → 処理は `continue` で続行しているので **WARN**
- DB 接続失敗（起動時）→ サービス起動不能なので **ERROR**（`log.Fatalf` 相当）

## Canonical Log Line

リクエストごとに1回、処理完了時にまとめて出力する。ハンドラ内で `log.Printf` を散発的に呼ばない。

### アンチパターン

```go
func (h *Handler) GetCart(ctx context.Context, req *connect.Request[ecv1.GetCartRequest]) (...) {
    log.Printf("GetCart called for customer_id=%d", req.Msg.CustomerId)  // (1) 開始ログ
    // ...
    log.Printf("created new cart (id=%d)", cart.ID)                      // (2) 中間ログ
    // ...
    log.Printf("fetched product: id=%d name=%s", p.Id, p.Name)          // (3) 中間ログ
    // ...
    log.Printf("GetCart completed")                                      // (4) 終了ログ
}
```

問題点:
- 1リクエストで複数行のログが出力され、ログ量が膨張する
- 各行が独立しており、横断的な検索・フィルタリングが困難
- ログの追加・削除がハンドラのビジネスロジックに混在する

### 正しいパターン

Echo ミドルウェアまたはインターセプタで、処理完了時に1回だけ出力する。

```go
// ミドルウェアが出力する canonical log line のイメージ
slog.InfoContext(ctx, "request completed",
    "method",      "ec.v1.CartService/GetCart",
    "status",      "ok",
    "duration_ms",  42,
    "customer_id",  12345,
    "request_id",  "req-abc-123",
)
```

ハンドラ内でログを出力するのは、WARN 以上の異常系に限定する。

```go
// ハンドラ内で出力してよいケース: 処理続行するが想定外の状態
if err != nil {
    slog.WarnContext(ctx, "product fetch failed, skipping item",
        "product_id", item.ProductID,
        "error",      err,
    )
    continue
}
```

## 構造化ログのフィールド規約

### 必須フィールド（canonical log line）

| フィールド名 | 型 | 説明 |
|-------------|-----|------|
| `method` | string | RPC メソッド名（`ec.v1.CartService/GetCart`） |
| `status` | string | `ok` / `error` |
| `duration_ms` | int | 処理時間（ミリ秒） |
| `request_id` | string | リクエスト識別子（サービス間で伝播） |

### 推奨フィールド（必要に応じて追加）

| フィールド名 | 型 | 説明 |
|-------------|-----|------|
| `customer_id` | int | 対象顧客の識別子 |
| `error` | string | エラーメッセージ（異常時のみ） |
| `trace_id` | string | 分散トレース ID（OpenTelemetry 導入後） |

### フィールド名の原則

- スネークケースで統一する（`requestId` ではなく `request_id`）
- OpenTelemetry Semantic Conventions に将来合わせやすい命名を選ぶ
- 同じ概念に複数の名前を使わない（`user_id` と `customer_id` を混在させない）

## ログに出してはいけない情報

| 種別 | 具体例 | 理由 |
|------|--------|------|
| 認証情報 | パスワード、トークン、API キー | 漏洩時の被害が甚大 |
| 個人情報 | メールアドレス、氏名、住所 | プライバシー保護 |
| 決済情報 | クレジットカード番号 | PCI DSS 違反 |
| リクエスト/レスポンスの生データ | JSON ボディ全体 | 上記を含む可能性がある |

識別子（`customer_id`, `cart_id` 等）は出力してよい。これらは単体では個人を特定できず、ログ検索に必要。

## 環境別の設定方針

| 環境 | 最低ログレベル | フォーマット |
|------|--------------|-------------|
| 開発（ローカル） | DEBUG | テキスト（人間が読みやすい） |
| 本番 | INFO | JSON（ツールで解析しやすい） |

## 監視との連携を意識した設計

ログメッセージの文字列に依存した監視条件を作らない。

```go
// Bad: メッセージ文字列を変更すると監視が壊れる
slog.Error("payment gateway timeout")

// Good: 専用フィールドで分類する
slog.ErrorContext(ctx, "external service timeout",
    "error_category", "payment_gateway",
    "error_type",     "timeout",
)
```

`error_category` や `error_type` のようなフィールドは将来の監視条件として使える。現時点で監視基盤がなくても、フィールドを入れておくことで後からの導入が容易になる。
