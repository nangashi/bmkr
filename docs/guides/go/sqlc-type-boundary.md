---
id: sqlc-type-boundary
category: architecture
scope: [backend, admin-ui]
severity: high
detectable_by_linter: false
---

# DB型のレイヤー境界

## アンチパターン

sqlc 生成型（`db.Product` 等）をテンプレートやハンドラのレスポンス層に直接渡す。DB スキーマ変更がプレゼンテーション層に波及し、変更の影響範囲が予測困難になる。

## 正しいパターン

レイヤー間にはプレゼンテーション用の型を定義し、ハンドラで変換する。

## 具体例

```go
// Bad: sqlc型を直接templに渡す
func (h *AdminHandler) HandleProductList(c echo.Context) error {
    products, _ := h.queries.ListProducts(ctx, params)
    return Render(c, ProductListPage(products)) // []db.Product を直接渡している
}

// Good: プレゼンテーション型に変換
type ProductItem struct {
    ID    int64
    Name  string
    Price int64
}

func (h *AdminHandler) HandleProductList(c echo.Context) error {
    rows, _ := h.queries.ListProducts(ctx, params)
    items := make([]ProductItem, len(rows))
    for i, r := range rows {
        items[i] = ProductItem{ID: r.ID, Name: r.Name, Price: r.Price}
    }
    return Render(c, ProductListPage(items))
}
```

## テンプレートに渡すデータは表示用にフォーマットする

プレゼンテーション型を定義するだけでなく、フィールドの型自体を表示用の `string` にする。テンプレート側でフォーマット処理を持たせない。

```go
// Bad: テンプレートが time.Time や int64 のフォーマットを担う
type ProductItem struct {
    Price     int64
    CreatedAt time.Time
}

// Good: ハンドラで string に変換済み
type ProductItem struct {
    Price     string // "1,000" のようにフォーマット済み
    CreatedAt string // "2006-01-02 15:04" のようにフォーマット済み
}
```

これにより、テンプレートは純粋な表示のみに集中でき、フォーマットロジックのテストもハンドラ側で完結する。

## 参照

- ADR-0004 (sqlc 採用)
- `docs/guides/go/pgtype-nullable.md` — nullable フィールドのフォーマット時は Valid チェックが必要
