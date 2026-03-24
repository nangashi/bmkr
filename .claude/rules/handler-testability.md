---
globs: services/**/handler*.go
---

# Handler の DB 依存注入

## アンチパターン

ハンドラが `*db.Queries` を直接フィールドに持つ。テストで DB をモックできず、DB エラーパスのテストが不可能になる。

## 正しいパターン

ハンドラが必要とする DB 操作だけを持つインターフェースを定義し、ハンドラはそのインターフェースに依存させる。テストではモック実装を注入する。

## 具体例

```go
// Bad: 具体型への直接依存
type AdminHandler struct {
    queries *db.Queries
}

// Good: インターフェース経由の依存注入
type adminQuerier interface {
    ListProducts(ctx context.Context, arg db.ListProductsParams) ([]db.Product, error)
    CountProducts(ctx context.Context) (int64, error)
}

type AdminHandler struct {
    q adminQuerier // *db.Queries がこのインターフェースを満たす
}
```

テストでは:
```go
type mockQuerier struct {
    listProductsFunc func(ctx context.Context, arg db.ListProductsParams) ([]db.Product, error)
    // ...
}
```
