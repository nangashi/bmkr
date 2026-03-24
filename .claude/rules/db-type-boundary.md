---
globs: services/**/handler*.go
---

# DB型のレイヤー境界

## アンチパターン1: DB型の直接露出

sqlc 生成型（`db.Product` 等）をテンプレートやハンドラのレスポンス層に直接渡す。DB スキーマ変更がプレゼンテーション層に波及し、変更の影響範囲が予測困難になる。

## アンチパターン2: nullable フィールドの Valid チェック漏れ

pgx/pgtype の nullable フィールド（`pgtype.Timestamptz`, `pgtype.Text` 等）の `.Valid` を確認せずに `.Time` や `.String` にアクセスする。ゼロ値が返り、サイレントに不正な値が使われる。

## 正しいパターン

レイヤー間にはプレゼンテーション用の型を定義し、ハンドラで変換する。変換時に nullable フィールドは必ず Valid チェックを行う。テンプレートに渡すデータは表示用にフォーマットする。

## 具体例

```go
// Bad: sqlc型を直接templに渡す
func (h *AdminHandler) HandleProductList(c echo.Context) error {
    products, _ := h.queries.ListProducts(ctx, params)
    return Render(c, ProductListPage(products)) // []db.Product を直接渡している
}

// Good: プレゼンテーション型に変換 + nullable の Valid チェック
type ProductItem struct {
    ID        int64
    Name      string
    Price     string // "1,000" のようにフォーマット済み
    CreatedAt string // "2006-01-02 15:04" のようにフォーマット済み
}

func (h *AdminHandler) HandleProductList(c echo.Context) error {
    rows, _ := h.queries.ListProducts(ctx, params)
    items := make([]ProductItem, len(rows))
    for i, r := range rows {
        createdAt := ""
        if r.CreatedAt.Valid {
            createdAt = r.CreatedAt.Time.Format("2006-01-02 15:04")
        }
        items[i] = ProductItem{ID: r.ID, Name: r.Name, Price: formatPrice(r.Price), CreatedAt: createdAt}
    }
    return Render(c, ProductListPage(items))
}
```

nullable フィールドのフォーマットをヘルパー関数に切り出してもよい:
```go
func formatTimestamptz(ts pgtype.Timestamptz, layout string) string {
    if ts.Valid {
        return ts.Time.Format(layout)
    }
    return ""
}
```

## 参照

- ADR-0004 (sqlc 採用)
