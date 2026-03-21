---
id: pgtype-nullable
category: correctness
scope: [backend, admin-ui]
severity: high
detectable_by_linter: false
---

# pgtype nullable フィールドの Valid チェック

## アンチパターン

pgx/pgtype の nullable フィールド（`pgtype.Timestamptz`, `pgtype.Text` 等）の `.Valid` を確認せずに `.Time` や `.String` にアクセスする。ゼロ値が返り、サイレントに不正な値が使われる。

## 正しいパターン

nullable フィールドを使う前に必ず `.Valid` をチェックし、無効時のフォールバック値を明示する。

## 具体例

```go
// Bad: Valid チェックなし
item.CreatedAt = product.CreatedAt.Time.Format("2006-01-02 15:04")

// Good: Valid チェックあり
if product.CreatedAt.Valid {
    item.CreatedAt = product.CreatedAt.Time.Format("2006-01-02 15:04")
} else {
    item.CreatedAt = ""
}

// Good: ヘルパー関数に切り出す
func formatTimestamptz(ts pgtype.Timestamptz, layout string) string {
    if ts.Valid {
        return ts.Time.Format(layout)
    }
    return ""
}
```

## 背景

ベンチマークで Opus が `CreatedAt.Valid` チェックを漏らすバグが確認されている。Codex は nullable フィールドの扱いに強い傾向があるが、モデルに依存せず明示的に注意すべきパターン。
